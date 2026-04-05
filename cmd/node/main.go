package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dominium/internal/api"
	"dominium/internal/miner"
	"dominium/internal/network"
	"dominium/internal/transaction"
)

// Node representa um participante ativo na rede Dominium.
// Ele encapsula o estado atual das contas (posse de NFTs), a Mempool (transações pendentes),
// a Blockchain (histórico de blocos validados) e a camada de rede P2P (Kafka).
type Node struct {
	id         string                    // Identificador único do nó na rede
	state      *transaction.AccountState // Estado atual validado (quem possui qual NFT)
	mempool    *transaction.Mempool      // Fila de transações aguardando mineração
	blockchain *miner.Blockchain         // Estrutura de dados em árvore contendo os blocos validados
	p2p        *network.P2P              // Barramento de eventos para comunicação P2P (Gossip)
	difficulty int32                     // Nível de dificuldade para o Proof of Work (Nbits zeros)
}

// NewNode inicializa uma nova instância de um nó da blockchain.
// Configura o estado inicial, inicializa a cadeia de blocos em memória e conecta-se ao Kafka.
func NewNode(nodeID string, kafkaBroker string) *Node {
	state := transaction.NewAccountState()
	return &Node{
		id:         nodeID,
		state:      state,
		mempool:    transaction.NewMempool(state),
		blockchain: miner.NewBlockchain(),
		p2p:        network.NewP2P([]string{kafkaBroker}, nodeID),
		difficulty: 8, // Dificuldade ajustada para 8 bits para rodar de forma fluída nos testes
	}
}

// Start inicia todos os serviços assíncronos do Nó: Consumo da rede P2P, Mineração e API HTTP.
// Retorna erro caso algum dos serviços não consiga ser inicializado.
func (n *Node) Start(ctx context.Context, mine bool, apiAddr string) error {
	log.Printf("Iniciando Node [%s]...", n.id)

	// 1. Iniciar Consumidor de Transações via P2P (Kafka)
	// Escuta transações geradas por outros nós e tenta adicioná-las à Mempool local.
	n.p2p.SubscribeTransactions(ctx, func(tx *transaction.Transaction) {
		if err := n.mempool.Add(tx); err != nil {
			log.Printf("[P2P Tx] Rejeitada %s: %v", tx.GetID(), err)
			return
		}
		log.Printf("[P2P Tx] Transação recebida e validada na mempool: %s", tx.GetID())
	})

	// 2. Iniciar Consumidor de Blocos via P2P (Kafka)
	// Escuta blocos minerados por outros nós e os adiciona à cadeia, resolvendo forks se necessário.
	n.p2p.SubscribeBlocks(ctx, func(block *miner.Block) {
		if err := n.blockchain.AddBlock(*block); err != nil {
			// Se o bloco já existir, for órfão ou inválido, é ignorado silenciosamente.
			return
		}

		log.Printf("[P2P Block] Novo bloco inserido na cadeia! Hash: %x", block.Hash)

		// Remove da Mempool as transações que acabaram de ser confirmadas neste bloco
		// e aplica-as definitivamente no estado (AccountState).
		var txIDs []string
		for _, tx := range block.Transactions {
			txIDs = append(txIDs, tx.ID)
			tx.Execute(n.state)
		}
		n.mempool.Remove(txIDs)
	})

	// 3. Iniciar Mineração local (se a flag -mine estiver ativada)
	if mine {
		go n.miningLoop(ctx)
	}

	// 4. Iniciar Servidor API REST
	// Permite a injeção externa de novas transações na rede.
	apiServer := api.NewServer(n.blockchain, n.mempool, n.p2p, n.state) // Passamos o AccountState para a API para que ela possa criar transações válidas (ex: Mint)
    go apiServer.Start(apiAddr)

	return nil
}

// miningLoop é a rotina em background responsável por agregar transações da Mempool
// e executar o algoritmo de Proof of Work (PoW) para criar novos blocos.
func (n *Node) miningLoop(ctx context.Context) {
	log.Println("[Minerador] Iniciando loop de mineração...")

	// Inicialização: Se a cadeia estiver completamente vazia, o minerador cria o Bloco Gênesis.
	if len(n.blockchain.GetCanonicalChain()) == 0 {
		genesisTx, _ := transaction.NewTransaction("GENESIS", "GENESIS", "0", transaction.TypeMintNFT)
		genesisBlock := miner.NewBlock([]byte{}, []transaction.Transaction{*genesisTx}, n.difficulty)
		miner.Mine(genesisBlock)
		
		n.blockchain.AddBlock(*genesisBlock)
		log.Printf("[Minerador] Bloco Gênesis criado! Hash: %x", genesisBlock.Hash)
		
		// Propaga o bloco Gênesis para que outros nós iniciem a partir da mesma raiz
		n.p2p.PublishBlock(ctx, genesisBlock)
	}

	// Loop contínuo de mineração (tenta criar um bloco a cada 5 segundos)
	for {
		select {
		case <-ctx.Done():
			return // Sai do loop graciosamente se o contexto for cancelado (ex: Ctrl+C)
		case <-time.After(5 * time.Second):
			// Puxa um máximo de 10 transações validadas da Mempool
			pendingTxs := n.mempool.GetPending(10)
			if len(pendingTxs) == 0 {
				continue // Poupa CPU: Não minera blocos vazios
			}

			// Converte os dados genéricos da Mempool para o tipo concreto Transaction
			var txs []transaction.Transaction
			var txIDs []string
			for _, pending := range pendingTxs {
				tx := pending.(*transaction.Transaction)
				txs = append(txs, *tx)
				txIDs = append(txIDs, tx.ID)
			}

			// Prepara o Header do novo bloco apontando para o Tip (Ponta) atual da cadeia
			prevHash := n.blockchain.GetLatestHash()
			newBlock := miner.NewBlock(prevHash, txs, n.difficulty)

			log.Printf("[Minerador] Iniciando PoW para bloco com %d transações...", len(txs))
			
			// Executa a mineração de fato (processo intensivo de CPU)
			miner.Mine(newBlock)

			// Tenta inserir o bloco recém-minerado na própria cadeia
			if err := n.blockchain.AddBlock(*newBlock); err != nil {
				log.Printf("[Minerador] Falha ao adicionar bloco (possível colisão/fork de rede): %v", err)
				continue
			}

			log.Printf("[Minerador] Sucesso! Bloco %x minerado (Nonce: %d). Propagando...", newBlock.Hash, newBlock.Nonce)

			// Atualiza o estado local consolidando as transações e as remove da Mempool
			for _, tx := range txs {
				tx.Execute(n.state)
			}
			n.mempool.Remove(txIDs)

			// Propaga o novo bloco validado para todos os outros nós via Kafka
			_ = n.p2p.PublishBlock(ctx, newBlock)
		}
	}
}

// Stop trata do encerramento seguro e gracioso (Graceful Shutdown) do Nó,
// garantindo que não haja corrupção de dados ao desligar o processo.
func (n *Node) Stop(ctx context.Context) error {
	log.Println("Encerrando Node Dominium...")
	// TODO (Futuro): Adicionar lógica para dar flush no banco de dados local (ex: LevelDB)
	// TODO (Futuro): Adicionar encerramento forçado das streams de leitura do Kafka.
	_ = ctx
	return nil
}

// main é o ponto de entrada da aplicação.
// Lê os argumentos da linha de comandos e orquestra o ciclo de vida do Nó.
func main() {
	var (
		nodeID  = flag.String("id", "node-1", "ID unico do nó na rede P2P")
		apiAddr = flag.String("api", ":8080", "Endereço e porta para a API REST local")
		p2pAddr = flag.String("p2p", "localhost:9092", "Endereço do cluster Kafka para rede P2P")
		dataDir = flag.String("data", "./data", "Diretório local para persistência de dados")
		mine    = flag.Bool("mine", false, "Habilita o loop de mineração de blocos local")
	)
	flag.Parse()

	_ = dataDir // Será utilizado na próxima fase quando implementarmos persistência em disco.

	// Instancia e configura o Nó com os parâmetros passados pelo usuário
	n := NewNode(*nodeID, *p2pAddr)

	// Gerenciamento de contexto para interceptar sinais do Sistema Operacional (ex: Ctrl+C)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Inicia os motores do Nó (passando a porta da API também)
	if err := n.Start(ctx, *mine, *apiAddr); err != nil {
		log.Fatalf("Falha crítica ao iniciar nó: %v", err)
	}

	// Trava a execução da main goroutine até receber o sinal de interrupção (Ctrl+C)
	<-ctx.Done()

	// Dá um prazo máximo de 5 segundos para que as goroutines terminem seus processos em andamento
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Encerra a aplicação
	if err := n.Stop(shutdownCtx); err != nil {
		log.Printf("Erro durante encerramento: %v", err)
	}
}
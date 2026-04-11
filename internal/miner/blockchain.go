package miner

import (
	"bytes"
	"encoding/hex"
	"errors"
	"log"
	"sync"
)

// BlockNode embrulha um bloco com metadados para facilitar a navegação em árvore (para forks).
type BlockNode struct {
	Block  Block
	Parent *BlockNode
	Height uint64
}

// Blockchain agora suporta forks mantendo todos os blocos num mapa e rastreando a ponta (tip) com maior altura.
type Blockchain struct {
	mu        sync.RWMutex
	blocks    map[string]*BlockNode // Mapeia o Hash (em hex) para o nó correspondente
	bestChain *BlockNode            // Aponta para o bloco no topo da corrente mais longa
}

// NewBlockchain inicializa a cadeia estruturada em árvore.
func NewBlockchain() *Blockchain {
	return &Blockchain{
		blocks: make(map[string]*BlockNode),
	}
}

// GetLatestHash retorna o hash do bloco na ponta da corrente mais longa.
func (bc *Blockchain) GetLatestHash() []byte {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if bc.bestChain == nil {
		return []byte{}
	}
	return bc.bestChain.Block.Hash
}

// IsEmpty retorna true quando a cadeia ainda nao possui nenhum bloco.
func (bc *Blockchain) IsEmpty() bool {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.bestChain == nil
}

// AddBlock tenta inserir um novo bloco na estrutura de árvore.
func (bc *Blockchain) AddBlock(newBlock Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	blockHashHex := hex.EncodeToString(newBlock.Hash)
	if _, exists := bc.blocks[blockHashHex]; exists {
		return errors.New("bloco ja existe na blockchain")
	}

	newNode := &BlockNode{
		Block: newBlock,
	}

	// Lógica para o bloco Gênesis (não tem bloco anterior)
	if len(newBlock.HashOfPrevious) == 0 {
		if bc.bestChain != nil {
			return errors.New("bloco genesis ja existe")
		}
		newNode.Height = 0
		bc.blocks[blockHashHex] = newNode
		bc.bestChain = newNode
		return nil
	}

	// Procura o bloco pai no mapa para verificar se a ligação é válida
	parentHashHex := hex.EncodeToString(newBlock.HashOfPrevious)
	parentNode, exists := bc.blocks[parentHashHex]
	if !exists {
		// Num cenário real P2P, poderiamos pedir este bloco à rede
		return errors.New("bloco anterior (pai) nao encontrado - bloco orfao")
	}

	// Valida as regras de consenso do bloco em relação ao seu pai específico
	if !ValidateBlock(newBlock, &parentNode.Block) {
		return errors.New("bloco invalido: falha na validacao de consenso")
	}

	// Configura a relação com o pai e incrementa a altura
	newNode.Parent = parentNode
	newNode.Height = parentNode.Height + 1

	// Insere no armazenamento
	bc.blocks[blockHashHex] = newNode

	// Aplica a REGRA DA CORRENTE MAIS LONGA (Nakamoto Consensus)
	// Se a altura deste novo bloco for maior que a do bestChain atual, ele torna-se o novo tip oficial
	if newNode.Height > bc.bestChain.Height {
		oldHeight := bc.bestChain.Height
		bc.bestChain = newNode
		// Log de reorganização se houve mudança
		if oldHeight != newNode.Height-1 {
			log.Printf("[Consenso] Fork detectado! Altura Local: %d, Altura Recebida: %d. Mudando de cadeia.", oldHeight, newNode.Height)
		}
	}

	return nil
}

// GetCanonicalChain retorna a lista linear de blocos que formam a corrente "oficial" atual.
// Navega de trás para a frente a partir do bestChain e inverte a ordem (Gênesis -> Tip).
func (bc *Blockchain) GetCanonicalChain() []Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if bc.bestChain == nil {
		return []Block{}
	}

	var chain []Block
	current := bc.bestChain

	for current != nil {
		chain = append(chain, current.Block)
		current = current.Parent
	}

	// Inverte o slice
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}

	return chain
}

// GetChainLength retorna o comprimento da corrente canônica atual.
func (bc *Blockchain) GetChainLength() uint64 {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if bc.bestChain == nil {
		return 0
	}
	return bc.bestChain.Height + 1
}

// GetBlockByHash retorna um bloco pelo seu hash, se existir.
func (bc *Blockchain) GetBlockByHash(hash []byte) (*Block, bool) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	hashHex := hex.EncodeToString(hash)
	node, exists := bc.blocks[hashHex]
	if !exists {
		return nil, false
	}
	return &node.Block, true
}

// FindCommonAncestor encontra o ancestral comum entre duas cadeias.
func (bc *Blockchain) FindCommonAncestor(hash1, hash2 []byte) ([]byte, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return bc.findCommonAncestorNoLock(hash1, hash2)
}

// findCommonAncestorNoLock encontra o ancestral comum assumindo que o lock ja foi adquirido.
func (bc *Blockchain) findCommonAncestorNoLock(hash1, hash2 []byte) ([]byte, error) {
	hash1Hex := hex.EncodeToString(hash1)
	hash2Hex := hex.EncodeToString(hash2)

	node1, exists1 := bc.blocks[hash1Hex]
	node2, exists2 := bc.blocks[hash2Hex]

	if !exists1 || !exists2 {
		return nil, errors.New("um dos blocos nao existe")
	}

	// Caminha para cima nas duas cadeias até encontrar um ancestral comum
	ancestors1 := make(map[string]bool)
	current := node1
	for current != nil {
		ancestors1[hex.EncodeToString(current.Block.Hash)] = true
		current = current.Parent
	}

	current = node2
	for current != nil {
		if ancestors1[hex.EncodeToString(current.Block.Hash)] {
			return current.Block.Hash, nil
		}
		current = current.Parent
	}

	return nil, errors.New("ancestral comum nao encontrado")
}

// Reorganize reorganiza a blockchain para uma nova corrente mais longa.
func (bc *Blockchain) Reorganize(newTipHash []byte) ([]Block, []Block, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	newTipHex := hex.EncodeToString(newTipHash)
	newTipNode, exists := bc.blocks[newTipHex]
	if !exists {
		return nil, nil, errors.New("novo tip nao encontrado")
	}

	if bc.bestChain == nil {
		bc.bestChain = newTipNode
		return nil, []Block{newTipNode.Block}, nil
	}

	// Encontra o ancestral comum
	commonAncestorHash, err := bc.findCommonAncestorNoLock(bc.bestChain.Block.Hash, newTipHash)
	if err != nil {
		return nil, nil, err
	}

	// Se o ancestral comum é o bestChain atual, não há reorganização necessária
	if bytes.Equal(commonAncestorHash, bc.bestChain.Block.Hash) {
		return nil, nil, nil
	}

	// Coleta blocos a serem desconectados (da corrente antiga)
	var blocksToDisconnect []Block
	current := bc.bestChain
	for current != nil {
		currentHash := hex.EncodeToString(current.Block.Hash)
		commonHashHex := hex.EncodeToString(commonAncestorHash)
		if currentHash == commonHashHex {
			break
		}
		blocksToDisconnect = append(blocksToDisconnect, current.Block)
		current = current.Parent
	}

	// Coleta blocos a serem conectados (da nova corrente)
	var blocksToConnect []Block
	current = newTipNode
	for current != nil {
		currentHash := hex.EncodeToString(current.Block.Hash)
		commonHashHex := hex.EncodeToString(commonAncestorHash)
		if currentHash == commonHashHex {
			break
		}
		blocksToConnect = append([]Block{current.Block}, blocksToConnect...) // prepend
		current = current.Parent
	}

	// Executa a reorganização
	bc.bestChain = newTipNode

	return blocksToDisconnect, blocksToConnect, nil
}

func (bc *Blockchain) GetAllBlocks() map[string]BlockNode {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	snapshot := make(map[string]BlockNode, len(bc.blocks))
	for hashHex, node := range bc.blocks {
		if node == nil {
			continue
		}
		snapshot[hashHex] = BlockNode{
			Block:  node.Block,
			Height: node.Height,
		}
	}

	return snapshot
}
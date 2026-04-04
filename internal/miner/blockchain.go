package miner

import (
	"encoding/hex"
	"errors"
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
		bc.bestChain = newNode
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
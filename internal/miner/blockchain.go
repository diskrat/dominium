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
	mu              sync.RWMutex
	blocks          map[string]*BlockNode          // Mapeia o Hash (em hex) para o nó correspondente
	bestChain       *BlockNode                     // Aponta para o bloco no topo da corrente mais longa
	orphans         map[string]Block               // Blocos sem pai conhecido, por hash
	orphansByParent map[string]map[string]struct{} // parentHashHex -> set(childHashHex)
	discarded       map[string]Block               // Blocos descartados por consenso (fork perdedor)
}

// AddBlockResult descreve o efeito de uma adicao de bloco.
type AddBlockResult struct {
	Added      bool
	Orphan     bool
	TipUpdated bool
	NewTipHash []byte
}

var (
	ErrDuplicateBlock = errors.New("bloco ja existe na blockchain")
	ErrOrphanBlock    = errors.New("bloco anterior (pai) nao encontrado - bloco orfao")
)

// NewBlockchain inicializa a cadeia estruturada em árvore.
func NewBlockchain() *Blockchain {
	return &Blockchain{
		blocks:          make(map[string]*BlockNode),
		orphans:         make(map[string]Block),
		orphansByParent: make(map[string]map[string]struct{}),
		discarded:       make(map[string]Block),
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
func (bc *Blockchain) AddBlock(newBlock Block) (AddBlockResult, error) {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	result := AddBlockResult{}
	oldBest := bc.bestChain

	blockHashHex := hex.EncodeToString(newBlock.Hash)
	if _, exists := bc.blocks[blockHashHex]; exists {
		return result, ErrDuplicateBlock
	}
	if _, exists := bc.orphans[blockHashHex]; exists {
		return result, ErrDuplicateBlock
	}

	newNode := &BlockNode{
		Block: newBlock,
	}

	// Lógica para o bloco Gênesis (não tem bloco anterior)
	if len(newBlock.HashOfPrevious) == 0 {
		if bc.bestChain != nil {
			return result, errors.New("bloco genesis ja existe")
		}
		newNode.Height = 0
		bc.blocks[blockHashHex] = newNode
		bc.bestChain = newNode
		result.Added = true
		result.TipUpdated = true
		result.NewTipHash = newNode.Block.Hash
		bc.attachOrphans(newNode)
		return result, nil
	}

	// Procura o bloco pai no mapa para verificar se a ligação é válida
	parentHashHex := hex.EncodeToString(newBlock.HashOfPrevious)
	parentNode, exists := bc.blocks[parentHashHex]
	if !exists {
		bc.addOrphan(newBlock)
		result.Orphan = true
		return result, ErrOrphanBlock
	}

	// Valida as regras de consenso do bloco em relação ao seu pai específico
	if !ValidateBlock(newBlock, &parentNode.Block) {
		return result, errors.New("bloco invalido: falha na validacao de consenso")
	}

	// Configura a relação com o pai e incrementa a altura
	newNode.Parent = parentNode
	newNode.Height = parentNode.Height + 1

	// Insere no armazenamento
	bc.blocks[blockHashHex] = newNode
	result.Added = true
	bc.updateBestChain(newNode, parentNode)
	bc.attachOrphans(newNode)
	bc.reselectBestChain()
	bc.pruneStaleForks()

	if bc.bestChain != oldBest {
		result.TipUpdated = true
		result.NewTipHash = bc.bestChain.Block.Hash
	}

	return result, nil
}

func (bc *Blockchain) addOrphan(block Block) {
	blockHashHex := hex.EncodeToString(block.Hash)
	bc.orphans[blockHashHex] = block

	parentHex := hex.EncodeToString(block.HashOfPrevious)
	children, exists := bc.orphansByParent[parentHex]
	if !exists {
		children = make(map[string]struct{})
		bc.orphansByParent[parentHex] = children
	}
	children[blockHashHex] = struct{}{}
}

func (bc *Blockchain) popOrphansByParent(parentHashHex string) []Block {
	children, exists := bc.orphansByParent[parentHashHex]
	if !exists {
		return nil
	}
	delete(bc.orphansByParent, parentHashHex)

	var blocks []Block
	for childHashHex := range children {
		if orphan, ok := bc.orphans[childHashHex]; ok {
			blocks = append(blocks, orphan)
			delete(bc.orphans, childHashHex)
		}
	}
	return blocks
}

func (bc *Blockchain) attachOrphans(parent *BlockNode) {
	queue := []*BlockNode{parent}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		parentHashHex := hex.EncodeToString(current.Block.Hash)
		orphans := bc.popOrphansByParent(parentHashHex)
		for _, orphan := range orphans {
			childHashHex := hex.EncodeToString(orphan.Hash)
			if _, exists := bc.blocks[childHashHex]; exists {
				continue
			}
			if !ValidateBlock(orphan, &current.Block) {
				continue
			}

			childNode := &BlockNode{
				Block:  orphan,
				Parent: current,
				Height: current.Height + 1,
			}
			bc.blocks[childHashHex] = childNode
			bc.updateBestChain(childNode, current)
			queue = append(queue, childNode)
		}
	}
}

func (bc *Blockchain) updateBestChain(candidate, parent *BlockNode) {
	if bc.bestChain == nil {
		bc.bestChain = candidate
		return
	}
	if parent == bc.bestChain {
		bc.bestChain = candidate
		return
	}
	if candidate.Height >= bc.bestChain.Height+2 {
		oldHeight := bc.bestChain.Height
		bc.bestChain = candidate
		log.Printf("[Consenso] Fork detectado! Altura Local: %d, Altura Recebida: %d. Mudando de cadeia.", oldHeight, candidate.Height)
	}
}

func (bc *Blockchain) reselectBestChain() {
	if bc.bestChain == nil {
		return
	}

	best := bc.bestChain
	parentSet := make(map[string]bool)
	for _, node := range bc.blocks {
		if node != nil && node.Parent != nil {
			parentSet[hex.EncodeToString(node.Parent.Block.Hash)] = true
		}
	}

	for hashHex, node := range bc.blocks {
		if node == nil {
			continue
		}
		if parentSet[hashHex] {
			continue
		}
		if node.Height >= best.Height+2 {
			best = node
		}
	}

	if best != bc.bestChain {
		oldHeight := bc.bestChain.Height
		bc.bestChain = best
		log.Printf("[Consenso] Fork detectado! Altura Local: %d, Altura Recebida: %d. Mudando de cadeia.", oldHeight, best.Height)
	}
}

func (bc *Blockchain) pruneStaleForks() {
	if bc.bestChain == nil {
		return
	}

	canonical := make(map[string]bool)
	current := bc.bestChain
	for current != nil {
		canonical[hex.EncodeToString(current.Block.Hash)] = true
		current = current.Parent
	}

	parentSet := make(map[string]bool)
	for _, node := range bc.blocks {
		if node != nil && node.Parent != nil {
			parentSet[hex.EncodeToString(node.Parent.Block.Hash)] = true
		}
	}

	for hashHex, node := range bc.blocks {
		if node == nil {
			continue
		}
		if canonical[hashHex] {
			continue
		}
		if parentSet[hashHex] {
			continue
		}
		if bc.bestChain.Height < node.Height+2 {
			continue
		}
		bc.discardBranch(node, canonical)
	}
}

func (bc *Blockchain) discardBranch(tip *BlockNode, canonical map[string]bool) {
	current := tip
	for current != nil {
		hashHex := hex.EncodeToString(current.Block.Hash)
		if canonical[hashHex] {
			return
		}
		bc.discarded[hashHex] = current.Block
		delete(bc.blocks, hashHex)
		current = current.Parent
	}
}

// GetOrphanCount retorna quantos blocos estao aguardando o pai.
func (bc *Blockchain) GetOrphanCount() int {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return len(bc.orphans)
}

// GetOrphansSnapshot retorna uma copia dos blocos orfaos atuais.
func (bc *Blockchain) GetOrphansSnapshot() map[string]Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	snapshot := make(map[string]Block, len(bc.orphans))
	for hashHex, block := range bc.orphans {
		snapshot[hashHex] = block
	}
	return snapshot
}

// GetDiscardedSnapshot retorna uma copia dos blocos descartados por consenso.
func (bc *Blockchain) GetDiscardedSnapshot() map[string]Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	snapshot := make(map[string]Block, len(bc.discarded))
	for hashHex, block := range bc.discarded {
		snapshot[hashHex] = block
	}
	return snapshot
}

// PopDiscardedSnapshot retorna e limpa os blocos descartados.
func (bc *Blockchain) PopDiscardedSnapshot() map[string]Block {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	snapshot := make(map[string]Block, len(bc.discarded))
	for hashHex, block := range bc.discarded {
		snapshot[hashHex] = block
	}
	bc.discarded = make(map[string]Block)
	return snapshot
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

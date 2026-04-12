package miner

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"sync"

	"dominium/internal/transaction"
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
}

// AddBlockResult descreve o efeito de uma adicao de bloco.
type AddBlockResult struct {
	Added      bool
	Orphan     bool
	TipUpdated bool
	NewTipHash []byte
	Discarded  []Block
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
	}
}

func (bc *Blockchain) chainToNode(node *BlockNode) []Block {
	var chain []Block
	current := node
	for current != nil {
		chain = append(chain, current.Block)
		current = current.Parent
	}
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain
}

func ValidateTransactionsAgainstState(state *transaction.AccountState, block Block, adminPubKey string) error {
	if state == nil {
		return errors.New("estado nulo")
	}

	for _, tx := range block.Transactions {
		if err := state.EnsureAccount(tx.PublKey); err != nil {
			return err
		}
		if err := state.EnsureAccount(tx.Recipient); err != nil {
			return err
		}

		if err := tx.Validate(state); err != nil {
			return fmt.Errorf("transacao invalida no bloco %x: %w", block.Hash, err)
		}
		if err := tx.ValidateConsensusRules(state, adminPubKey); err != nil {
			return fmt.Errorf("transacao viola regras de consenso no bloco %x: %w", block.Hash, err)
		}
		if err := tx.Execute(state); err != nil {
			return fmt.Errorf("falha ao executar transacao no bloco %x: %w", block.Hash, err)
		}
	}
	return nil
}

func (bc *Blockchain) applyBlockTransactions(state *transaction.AccountState, block Block, adminPubKey string) error {
	return ValidateTransactionsAgainstState(state, block, adminPubKey)
}

func (bc *Blockchain) validateBlockTransactions(parent *BlockNode, block Block, adminPubKey string) error {
	state := transaction.NewAccountState()
	if parent != nil {
		previousChain := bc.chainToNode(parent)
		for _, previousBlock := range previousChain {
			if err := bc.applyBlockTransactions(state, previousBlock, adminPubKey); err != nil {
				return fmt.Errorf("estado invalido antes do bloco: %w", err)
			}
		}
	}
	return bc.applyBlockTransactions(state, block, adminPubKey)
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
func (bc *Blockchain) AddBlock(newBlock Block, adminPubKey string) (AddBlockResult, error) {
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
		bc.attachOrphans(newNode, adminPubKey)
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

	if err := bc.validateBlockTransactions(parentNode, newBlock, adminPubKey); err != nil {
		return result, fmt.Errorf("bloco invalido: %w", err)
	}

	// Configura a relação com o pai e incrementa a altura
	newNode.Parent = parentNode
	newNode.Height = parentNode.Height + 1

	// Insere no armazenamento
	bc.blocks[blockHashHex] = newNode
	result.Added = true
	bc.updateBestChain(newNode, parentNode)
	bc.attachOrphans(newNode, adminPubKey)
	bc.reselectBestChain()
	if bc.bestChain != oldBest {
		// pruneStaleForks now returns the list of discarded blocks (deleted)
		discarded := bc.pruneStaleForks()
		result.Discarded = discarded
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

func (bc *Blockchain) attachOrphans(parent *BlockNode, adminPubKey string) {
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
			if err := bc.validateBlockTransactions(current, orphan, adminPubKey); err != nil {
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

func (bc *Blockchain) pruneStaleForks() []Block {
	if bc.bestChain == nil {
		return nil
	}

	// Build canonical set (hashes that are in the current best chain)
	canonical := make(map[string]bool)
	current := bc.bestChain
	for current != nil {
		canonical[hex.EncodeToString(current.Block.Hash)] = true
		current = current.Parent
	}

	// Build a set of parent hashes to detect internal nodes
	parentSet := make(map[string]bool)
	for _, node := range bc.blocks {
		if node != nil && node.Parent != nil {
			parentSet[hex.EncodeToString(node.Parent.Block.Hash)] = true
		}
	}

	var removed []Block
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
		// discard the branch and collect removed blocks
		discarded := bc.discardBranch(node, canonical)
		if len(discarded) > 0 {
			removed = append(removed, discarded...)
		}
	}

	return removed
}

func (bc *Blockchain) discardBranch(tip *BlockNode, canonical map[string]bool) []Block {
	var removed []Block
	current := tip
	for current != nil {
		hashHex := hex.EncodeToString(current.Block.Hash)
		if canonical[hashHex] {
			break
		}
		removed = append(removed, current.Block)
		delete(bc.blocks, hashHex)
		current = current.Parent
	}
	return removed
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

func BuildStateFromChain(chain []Block) (*transaction.AccountState, map[string]bool, error) {
	state := transaction.NewAccountState()
	chainTxs := make(map[string]bool)

	for _, block := range chain {
		for _, tx := range block.Transactions {
			chainTxs[tx.ID] = true
			if err := state.EnsureAccount(tx.PublKey); err != nil {
				return nil, nil, err
			}
			if err := state.EnsureAccount(tx.Recipient); err != nil {
				return nil, nil, err
			}
			if err := tx.Execute(state); err != nil {
				return nil, nil, err
			}
		}
	}

	return state, chainTxs, nil
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

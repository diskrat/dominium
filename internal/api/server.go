package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"dominium/internal/miner"
	"dominium/internal/network"
	"dominium/internal/transaction"
)

// Server expõe a API HTTP para interagirmos com o nó
type Server struct {
	blockchain *miner.Blockchain
	mempool    *transaction.Mempool
	p2p        *network.P2P
	generator  *transaction.Generator
	state      *transaction.AccountState
}

// NewServer inicializa a API.
func NewServer(bc *miner.Blockchain, mp *transaction.Mempool, p2p *network.P2P, state *transaction.AccountState) *Server {
	// 2. Usamos uma seed fixa (ex: 42) para que TODOS os nós gerem a mesma chave de Admin
	gen := transaction.NewGenerator(42)
	admin, _ := gen.CreateIdentity("Dominium-Admin")
	gen.SetAdmin(admin)

	// 3. CRIAÇÃO DA CONTA DE ADMIN PARA TESTES:
	_, err := state.CreateAccount(admin.PublicKey)
	if err != nil && err.Error() != "conta ja existe" {
		log.Printf("[API] Aviso ao criar conta admin: %v", err)
	}

	return &Server{
		blockchain: bc,
		mempool:    mp,
		p2p:        p2p,
		generator:  gen,
		state:      state,
	}
}

func (s *Server) Start(addr string) {
	mux := http.NewServeMux()

	// Endpoints
	mux.HandleFunc("/blocks", s.getBlocks)
	mux.HandleFunc("/mint", s.mintTestNFT) // Rota de conveniência para testes

	log.Printf("[API] Servidor HTTP a escutar em %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("[API] Erro no servidor HTTP: %v", err)
	}
}

// getBlocks retorna toda a corrente canônica da blockchain
func (s *Server) getBlocks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	chain := s.blockchain.GetCanonicalChain()
	json.NewEncoder(w).Encode(chain)
}

// mintTestNFT gera uma transação de Mint, adiciona à mempool local e publica no Kafka
func (s *Server) mintTestNFT(w http.ResponseWriter, r *http.Request) {
	// 1. Gera uma identidade aleatória para receber o NFT
	randomUser, _ := s.generator.CreateIdentity(fmt.Sprintf("User-%s", s.generator.RandomNFTID()[:4]))

	// Injetamos o destinatário aleatório no estado da blockchain para passar na validação rigorosa
	_, _ = s.state.CreateAccount(randomUser.PublicKey)

	// 2. Cria a transação assinada usando o seu Generator
	tx, err := s.generator.MintTx(randomUser, "")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 3. Adiciona à Mempool local
	if err := s.mempool.Add(tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 4. Propaga a transação para o resto da rede P2P (Kafka)
	_ = s.p2p.PublishTransaction(context.Background(), tx)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Transacao gerada e propagada com sucesso!",
		"tx_id":   tx.GetID(),
		"nft_id":  tx.NFTID,
		"owner":   randomUser.Name,
	})
}
// main.go
package main

import (
	"fmt"
	"time"

	"github.com/diskrat/dominium/internal/core"
)

func main() {
	fmt.Println("=== Iniciando Sistema Dominium (Registro de Terras RN) ===")

	//INSTANCIANDO os componentes principais
	mempool := core.NewMempool()
	dificuldade := 4 // O hash final precisará começar com "0000"
	minerador := core.NovoMinerador(dificuldade)

	//envio de matrIculas de imOveis pelos cartOrios (EXEMPLO)
	fmt.Println("\n[1] Recebendo transações...")
	core.NewAndPostTransaction(mempool, "MAT-001", "Fazenda em Macaíba - 50 Hectares", 500)
	core.NewAndPostTransaction(mempool, "MAT-002", "Lote em Ponta Negra - Beira Mar", 850)
	core.NewAndPostTransaction(mempool, "MAT-003", "Sítio em Mossoró - Área de Extração", 300)

	// carrega a comunicacao sem quebrar
	time.Sleep(1 * time.Second)

	//montar o bloco com o que tem na mempool
	fmt.Println("\n[2] Montando o Bloco...")
	// Como é o primeiro bloco (genesis), o Hash anterior é "0000...000"
	hashAnterior := "0000000000000000000000000000000000000000000000000000000000000000"
	bloco := mempool.MontarProximoBloco(hashAnterior, dificuldade)

	fmt.Printf("Bloco montado com %d transações. Merkle Root: %s\n", 
		bloco.Header.TransactionCounter, bloco.Header.MerkleRootHash)

	// entregamos o bloco para o minerador trabalhar (prova de trabai)
	fmt.Println("\n[3] Iniciando Mineração (Aguarde...)")
	inicio := time.Now()  //iniciando motores
	
	hashFinal := minerador.Minar(bloco) //decolar
	
	tempoGasto := time.Since(inicio) //tempo gasto para minerar

	fmt.Println("\n==================================================")
	fmt.Println("BLOCO MINERADO: SUÇESSUUU!")
	fmt.Printf("Hash Final: %s\n", hashFinal)
	fmt.Printf("Nonce encontrado: %d\n", bloco.Header.Nonce)
	fmt.Printf("Tempo de mineração: %v\n", tempoGasto)
	fmt.Println("==================================================")

	//impamos a mempool, pois as terras agora estão registradas
	var idsConfirmados []string
	for _, tx := range bloco.Body.Transactions {
		idsConfirmados = append(idsConfirmados, tx.TXid)
	}
	mempool.LimparConfirmadas(idsConfirmados)
	
	fmt.Println("\n[4] Mempool limpa. Aguardando novas transações...")
}
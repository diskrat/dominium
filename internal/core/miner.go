// internal\core\miner.go
package core

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

type Miner struct {
	Dificuldade int
}

func NovoMinerador(dif int) *Miner {
	return &Miner{Dificuldade: dif}
}

func (m *Miner) Minar(bloco *Block) string {
	target := strings.Repeat("0", m.Dificuldade) // Ex: "0000" se dif for 4

	for {
		// geramos o hash do cabeçalho atual
		hash := m.CalcularHash(bloco.Header)

		// verificamos se o hash atende à dificuldade
		if strings.HasPrefix(hash, target) {
			bloco.Header.Hash = hash // Salvamos o hash vencedor
			return hash
		}

		// sSe n deu certo, incrementamos o nonce e tentamos de novo
		bloco.Header.Nonce++
	}
}

//  calcular o hash transforma os dados do header em uma string SHA-256 (prova de trabalhjo)
func (m *Miner) CalcularHash(h Header) string {
    dados := fmt.Sprintf("%s%s%d%d%d", 
        h.HashOfPrevious, h.MerkleRootHash, h.Timestamp, h.Nbits, h.Nonce)
    
    hash := sha256.Sum256([]byte(dados))
    return fmt.Sprintf("%x", hash) //retorno em hexadecimal
}

func (b *Body) GerarMerkleRoot() string {
    var hashes []string

    // primeiro, tiramos o hash individual de cada transaçao
    for _, tx := range b.Transactions {
        hashTx := sha256.Sum256([]byte(tx.TXid + tx.Data))
        hashes = append(hashes, fmt.Sprintf("%x", hashTx))
    }

    // se não houver 10 transações, retornamos um hash vazio para as transacoes faltantes
    if len(hashes) == 0 {
        return ""
    }

    //subindo a arvore ate sobrar apenas um hash (a root)
    for len(hashes) > 1 {
        // Se o numero de hashes for impar, duplicamos o ultimo para parear
        if len(hashes)%2 != 0 {
            hashes = append(hashes, hashes[len(hashes)-1])
        }

        //gpt me acuda
        var nivelSuperior []string
        for i := 0; i < len(hashes); i += 2 {
            combinado := hashes[i] + hashes[i+1]
            hashPai := sha256.Sum256([]byte(combinado))
            nivelSuperior = append(nivelSuperior, fmt.Sprintf("%x", hashPai))
        }
        hashes = nivelSuperior
    }

    return hashes[0] //hash do merkleroot
}
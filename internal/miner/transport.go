package miner

// BlockSink publica blocos minerados ou recebidos para um transporte de rede.
type BlockSink interface {
	Publish(block *Block) error
}

// BlockSource fornece blocos para consumo pelos nós da rede.
type BlockSource interface {
	Subscribe(handler func(*Block) error) error
}

// BlockTransport combina a capacidade de publicar e receber blocos.
type BlockTransport interface {
	BlockSink
	BlockSource
}

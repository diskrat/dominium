package transaction

// TransactionSink publica transacoes sem acoplar o core a um transporte.
type TransactionSink interface {
	Publish(tx *Transaction) error
}

// TransactionSource fornece transacoes para consumo pelos nos.
type TransactionSource interface {
	Subscribe(handler func(*Transaction) error) error
}

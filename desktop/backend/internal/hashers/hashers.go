package hashers

import (
	"crypto/sha256"
	"hash"
	"pennyapp/backend/model"
)

type TransactionHasher interface {
	Hash(t model.Transaction) hash.Hash
}

type TransactionHashV1 struct{}

func (h TransactionHashV1) String() string {
	return "v1"
}

func (h TransactionHashV1) Hash(t model.Transaction) hash.Hash {
	s := sha256.New()
	s.Write([]byte(t.Date))
	s.Write([]byte(t.Memo))
	return s
}

func NewTransactionHashV1() TransactionHashV1 {
	return TransactionHashV1{}
}

type CreditCardHasher interface {
	Hash(accountID string) hash.Hash
}

// CreditCardHasher is not versioned. There most likely won't be a need to
// change the hash algo for credit cards.
type CreditCardHash struct{}

func NewCreditCardHash() CreditCardHash {
	return CreditCardHash{}
}

func (c CreditCardHash) Hash(accountID string) hash.Hash {
	s := sha256.New()
	s.Write([]byte(accountID))
	return s
}

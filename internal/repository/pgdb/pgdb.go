package pgdb

import (
	"database/sql"
)

type Repository struct {
	*UserRepository
	*CoinTransferRepository
	*PurchaseRepository
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		UserRepository:         NewUserRepository(db),
		CoinTransferRepository: NewCoinTransferRepository(db),
		PurchaseRepository:     NewPurchaseRepository(db),
	}
}

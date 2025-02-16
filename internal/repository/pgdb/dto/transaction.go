package dto

type TransactionDTO struct {
	FromUser string `db:"from_user"`
	ToUser   string `db:"to_user"`
	Amount   int    `db:"amount"`
}

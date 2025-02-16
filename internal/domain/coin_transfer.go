package domain

type CoinTransfer struct {
	FromUserID      string
	ToUserID        string
	Amount          int
	TransactionType string
}

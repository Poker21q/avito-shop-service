package dto

type CoinInfoDTO struct {
	UserID      string `db:"user_id"`
	CoinBalance int    `db:"coin_balance"`
}

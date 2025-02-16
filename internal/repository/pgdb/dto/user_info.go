package dto

type UserInfoDTO struct {
	CoinInfo     CoinInfoDTO        `db:"coin_info"`
	Inventory    []UserInventoryDTO `db:"inventory"`
	Transactions []TransactionDTO   `db:"transactions"`
}

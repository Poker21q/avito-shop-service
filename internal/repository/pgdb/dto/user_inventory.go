package dto

type UserInventoryDTO struct {
	MerchName string `db:"name"`
	Quantity  int    `db:"quantity"`
}

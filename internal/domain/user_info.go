package domain

type UserInfo struct {
	CoinBalance         int
	Inventory           []UserInventory
	CoinHistoryReceived []CoinTransfer
	CoinHistorySent     []CoinTransfer
}

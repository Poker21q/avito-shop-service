package domain

type User struct {
	ID           string
	Name         string
	PasswordHash string
	CoinBalance  int
}

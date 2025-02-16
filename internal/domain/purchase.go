package domain

import (
	"time"
)

type Purchase struct {
	ID           string
	UserID       string
	TotalPrice   int
	PurchaseDate time.Time
	Items        []PurchaseItem
}

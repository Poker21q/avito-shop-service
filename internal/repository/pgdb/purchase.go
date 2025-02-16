package pgdb

import (
	"context"
	"database/sql"
	"errors"
	"github.com/google/uuid"
	"merch/internal/domain"
)

type PurchaseRepository struct {
	db *sql.DB
}

func NewPurchaseRepository(db *sql.DB) *PurchaseRepository {
	return &PurchaseRepository{db: db}
}

func (r *PurchaseRepository) Buy(ctx context.Context, userID, merchName string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return errors.Join(domain.ErrInternalServerError, err)
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	merchID, price, coinBalance, err := r.fetchMerchandiseAndBalance(ctx, tx, userID, merchName)
	if err != nil {
		return err
	}

	if coinBalance < price {
		return domain.ErrInsufficientFunds
	}

	purchaseID := uuid.New()

	if err = r.executePurchaseTransaction(ctx, tx, userID, price, purchaseID, merchID); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return errors.Join(domain.ErrInternalServerError, err)
	}

	return nil
}

func (r *PurchaseRepository) fetchMerchandiseAndBalance(ctx context.Context, tx *sql.Tx, userID, merchName string) (int, int, int, error) {
	var merchID, price, coinBalance int
	err := tx.QueryRowContext(ctx, `
		SELECT m.merch_id, m.price, u.coin_balance 
		FROM merch m
		JOIN users u ON u.user_id = $1
		WHERE m.name = $2`, userID, merchName).Scan(&merchID, &price, &coinBalance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, 0, 0, domain.ErrNotFound
		}
		return 0, 0, 0, errors.Join(domain.ErrInternalServerError, err)
	}
	return merchID, price, coinBalance, nil
}

func (r *PurchaseRepository) executePurchaseTransaction(ctx context.Context, tx *sql.Tx, userID string, price int, purchaseID uuid.UUID, merchID int) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE users 
		SET coin_balance = coin_balance - $1 
		WHERE user_id = $2
	`, price, userID)
	if err != nil {
		return errors.Join(domain.ErrInternalServerError, err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO purchases (purchase_id, user_id, total_price) 
		VALUES ($1, $2, $3)
	`, purchaseID, userID, price)
	if err != nil {
		return errors.Join(domain.ErrInternalServerError, err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO purchase_items (purchase_id, merch_id, quantity, price_at_purchase) 
		VALUES ($1, $2, $3, $4)
	`, purchaseID, merchID, 1, price)
	if err != nil {
		return errors.Join(domain.ErrInternalServerError, err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO user_inventory (user_id, merch_id, quantity)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, merch_id) 
		DO UPDATE SET quantity = user_inventory.quantity + $3
	`, userID, merchID, 1)
	if err != nil {
		return errors.Join(domain.ErrInternalServerError, err)
	}

	return nil
}

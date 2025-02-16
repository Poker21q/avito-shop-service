package pgdb

import (
	"context"
	"database/sql"
	"errors"
	"merch/internal/domain"
)

type CoinTransferRepository struct {
	db *sql.DB
}

func NewCoinTransferRepository(db *sql.DB) *CoinTransferRepository {
	return &CoinTransferRepository{db: db}
}

func (r *CoinTransferRepository) SendCoins(ctx context.Context, fromUserID string, toUserName string, amount int) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return errors.Join(domain.ErrInternalServerError, err)
	}
	defer func(tx *sql.Tx) {
		_ = tx.Rollback()
	}(tx)

	fromBalance, err := r.fetchBalance(ctx, tx, fromUserID)
	if err != nil {
		return err
	}

	if fromBalance < amount {
		return domain.ErrInsufficientFunds
	}

	toUserID, err := r.fetchUserIDByName(ctx, tx, toUserName)
	if err != nil {
		return err
	}

	if err = r.executeCoinTransfer(ctx, tx, fromUserID, toUserID, amount); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return errors.Join(domain.ErrInternalServerError, err)
	}

	return nil
}

func (r *CoinTransferRepository) fetchBalance(ctx context.Context, tx *sql.Tx, userID string) (int, error) {
	var balance int
	err := tx.QueryRowContext(ctx, `
		SELECT coin_balance
		FROM users
		WHERE user_id = $1`, userID).Scan(&balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrNotFound
		}
		return 0, errors.Join(domain.ErrInternalServerError, err)
	}
	return balance, nil
}

func (r *CoinTransferRepository) executeCoinTransfer(ctx context.Context, tx *sql.Tx, fromUserID, toUserID string, amount int) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE users SET coin_balance = coin_balance - $1 WHERE user_id = $2;
	`, amount, fromUserID)
	if err != nil {
		return errors.Join(domain.ErrInternalServerError, err)
	}

	_, err = tx.ExecContext(ctx, `
		UPDATE users SET coin_balance = coin_balance + $1 WHERE user_id = $2;
	`, amount, toUserID)
	if err != nil {
		return errors.Join(domain.ErrInternalServerError, err)
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO coin_transfers (from_user_id, to_user_id, amount)
		VALUES ($1, $2, $3);
	`, fromUserID, toUserID, amount)
	if err != nil {
		return errors.Join(domain.ErrInternalServerError, err)
	}

	return nil
}

func (r *CoinTransferRepository) fetchUserIDByName(ctx context.Context, tx *sql.Tx, userName string) (string, error) {
	var userID string
	err := tx.QueryRowContext(ctx, `
		SELECT user_id
		FROM users
		WHERE name = $1
	`, userName).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrNotFound
		}
		return "", errors.Join(domain.ErrInternalServerError, err)
	}
	return userID, nil
}

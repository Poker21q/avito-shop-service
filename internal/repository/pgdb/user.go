package pgdb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"golang.org/x/sync/singleflight" // inspired by https://balun.courses/courses/concurrency/patterns?#topic1
	"merch/internal/domain"
	"merch/internal/repository/pgdb/dto"
)

type UserRepository struct {
	db    *sql.DB
	group *singleflight.Group
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db:    db,
		group: &singleflight.Group{},
	}
}

func (r *UserRepository) IsUserExists(ctx context.Context, username string) (bool, error) {
	result, err, _ := r.group.Do("IsUserExists:"+username, func() (interface{}, error) {
		const query = `SELECT user_id FROM users WHERE name = $1`
		var userID string
		err := r.db.QueryRowContext(ctx, query, username).Scan(&userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return false, nil
			}
			return false, err
		}
		return true, nil
	})

	if err != nil {
		return false, err
	}
	return result.(bool), nil
}

func (r *UserRepository) CreateUser(ctx context.Context, username, passwordHash string) (string, error) {
	const query = `INSERT INTO users (name, password_hash) VALUES ($1, $2) RETURNING user_id`
	var userID string
	err := r.db.QueryRowContext(ctx, query, username, passwordHash).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("CreateUser failed: %w", errors.Join(domain.ErrInternalServerError, err))
	}
	return userID, nil
}

func (r *UserRepository) Auth(ctx context.Context, username, passwordHash string) (string, error) {
	result, err, _ := r.group.Do("Auth:"+username, func() (interface{}, error) {
		const query = `SELECT user_id FROM users WHERE name = $1 AND password_hash = $2`
		var userID string
		err := r.db.QueryRowContext(ctx, query, username, passwordHash).Scan(&userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return "", fmt.Errorf("Auth failed for username %s: %w", username, errors.Join(domain.ErrInvalidCredentials, err))
			}
			return "", fmt.Errorf("Auth failed for username %s: %w", username, errors.Join(domain.ErrInternalServerError, err))
		}
		return userID, nil
	})

	if err != nil {
		return "", err
	}
	return result.(string), nil
}

func (r *UserRepository) GetUserInfo(ctx context.Context, userID string) (*domain.UserInfo, error) {
	result, err, _ := r.group.Do("GetUserInfo:"+userID, func() (interface{}, error) {
		dbTx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		if err != nil {
			return nil, fmt.Errorf("GetUserInfo: BeginTx failed for userID %s: %w", userID, errors.Join(domain.ErrInternalServerError, err))
		}
		defer dbTx.Rollback()

		username, err := r.getUsernameByID(dbTx, userID, ctx)
		if err != nil {
			return nil, fmt.Errorf("GetUserInfo: getUsernameByID failed for userID %s: %w", userID, errors.Join(domain.ErrInvalidCredentials, err))
		}

		coinInfo, err := r.getCoinInfo(dbTx, userID, ctx)
		if err != nil {
			return nil, fmt.Errorf("GetUserInfo: getCoinInfo failed for userID %s: %w", userID, errors.Join(domain.ErrInternalServerError, err))
		}

		userInventory, err := r.getUserInventory(dbTx, userID, ctx)
		if err != nil {
			return nil, fmt.Errorf("GetUserInfo: getUserInventory failed for userID %s: %w", userID, err)
		}

		transactions, err := r.getUserTransactions(dbTx, userID, ctx)
		if err != nil {
			return nil, fmt.Errorf("GetUserInfo: getUserTransactions failed for userID %s: %w", userID, err)
		}

		if err := dbTx.Commit(); err != nil {
			return nil, fmt.Errorf("GetUserInfo: Commit failed for userID %s: %w", userID, errors.Join(domain.ErrInternalServerError, fmt.Errorf("transaction commit failed: %w", err)))
		}

		return mapUserInfoToDomain(coinInfo, userInventory, transactions, username), nil
	})

	if err != nil {
		return nil, err
	}
	return result.(*domain.UserInfo), nil
}

func (r *UserRepository) getUsernameByID(dbTx *sql.Tx, userID string, ctx context.Context) (string, error) {
	var username string
	const query = `SELECT name FROM users WHERE user_id = $1`
	err := dbTx.QueryRowContext(ctx, query, userID).Scan(&username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", domain.ErrNotFound
		}
		return "", errors.Join(domain.ErrInternalServerError, err)
	}
	return username, nil
}

func (r *UserRepository) getCoinInfo(tx *sql.Tx, userID string, ctx context.Context) (*dto.CoinInfoDTO, error) {
	const query = `SELECT user_id, coin_balance FROM users WHERE user_id = $1`
	var coinInfo dto.CoinInfoDTO
	if err := tx.QueryRowContext(ctx, query, userID).Scan(&coinInfo.UserID, &coinInfo.CoinBalance); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, errors.Join(domain.ErrInternalServerError, err)
	}
	return &coinInfo, nil
}

func (r *UserRepository) getUserInventory(tx *sql.Tx, userID string, ctx context.Context) ([]dto.UserInventoryDTO, error) {
	const query = `
		SELECT m.name, ui.quantity
		FROM user_inventory ui
		JOIN merch m ON ui.merch_id = m.merch_id
		WHERE ui.user_id = $1`
	rows, err := tx.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Join(domain.ErrInternalServerError, err)
	}
	defer rows.Close()

	var inventory []dto.UserInventoryDTO
	for rows.Next() {
		var item dto.UserInventoryDTO
		if err := rows.Scan(&item.MerchName, &item.Quantity); err != nil {
			return nil, errors.Join(domain.ErrInternalServerError, err)
		}
		inventory = append(inventory, item)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Join(domain.ErrInternalServerError, err)
	}

	return inventory, nil
}

func (r *UserRepository) getUserTransactions(tx *sql.Tx, userID string, ctx context.Context) ([]dto.TransactionDTO, error) {
	const query = `
		SELECT 
			u_from.name AS from_user, 
			u_to.name AS to_user,
			ct.amount
		FROM coin_transfers ct
		JOIN users u_from ON ct.from_user_id = u_from.user_id
		JOIN users u_to ON ct.to_user_id = u_to.user_id
		WHERE ct.from_user_id = $1 OR ct.to_user_id = $1`
	rows, err := tx.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, errors.Join(domain.ErrInternalServerError, err)
	}
	defer rows.Close()

	var transactions []dto.TransactionDTO
	for rows.Next() {
		var transaction dto.TransactionDTO
		if err := rows.Scan(&transaction.FromUser, &transaction.ToUser, &transaction.Amount); err != nil {
			return nil, errors.Join(domain.ErrInternalServerError, err)
		}
		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Join(domain.ErrInternalServerError, err)
	}

	return transactions, nil
}

func mapUserInfoToDomain(coinInfo *dto.CoinInfoDTO, inventory []dto.UserInventoryDTO, transactions []dto.TransactionDTO, username string) *domain.UserInfo {
	sentTransfers, receivedTransfers := mapTransactionsToDomain(transactions, username)

	return &domain.UserInfo{
		CoinBalance:         coinInfo.CoinBalance,
		Inventory:           mapInventoryToDomain(inventory),
		CoinHistoryReceived: receivedTransfers,
		CoinHistorySent:     sentTransfers,
	}
}

func mapInventoryToDomain(dto []dto.UserInventoryDTO) []domain.UserInventory {
	var inventory []domain.UserInventory
	for _, item := range dto {
		inventory = append(inventory, domain.UserInventory{
			MerchName: item.MerchName,
			Quantity:  item.Quantity,
		})
	}
	return inventory
}

func mapTransactionsToDomain(dto []dto.TransactionDTO, username string) ([]domain.CoinTransfer, []domain.CoinTransfer) {
	var sentTransfers []domain.CoinTransfer
	var receivedTransfers []domain.CoinTransfer

	for _, transaction := range dto {
		if transaction.FromUser == username {
			sentTransfers = append(sentTransfers, domain.CoinTransfer{
				FromUserID:      transaction.FromUser,
				ToUserID:        transaction.ToUser,
				Amount:          transaction.Amount,
				TransactionType: "sent",
			})
		} else if transaction.ToUser == username {
			receivedTransfers = append(receivedTransfers, domain.CoinTransfer{
				FromUserID:      transaction.FromUser,
				ToUserID:        transaction.ToUser,
				Amount:          transaction.Amount,
				TransactionType: "received",
			})
		}
	}

	return sentTransfers, receivedTransfers
}

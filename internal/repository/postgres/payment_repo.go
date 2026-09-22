package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"wallet-observer/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository struct {
	pool *pgxpool.Pool
}

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{pool: pool}
}

func (r *PaymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	query := `
		INSERT INTO payments (
			id, order_id, customer_ref, amount, currency, network,
			pay_address, pay_url, status, tx_hash, provider_name,
			created_at, updated_at, expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`
	_, err := r.pool.Exec(ctx, query,
		p.ID, p.OrderID, p.CustomerRef, p.Amount, string(p.Currency), p.Network,
		p.PayAddress, p.PayURL, string(p.Status), p.TxHash, p.ProviderName,
		p.CreatedAt, p.UpdatedAt, p.ExpiresAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert payment: %w", err)
	}
	return nil
}

func (r *PaymentRepository) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, customer_ref, amount, currency, network,
		       pay_address, pay_url, status, tx_hash, provider_name,
		       created_at, updated_at, expires_at
		FROM payments
		WHERE id = $1
	`
	return r.scanPayment(r.pool.QueryRow(ctx, query, id))
}

func (r *PaymentRepository) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	query := `
		SELECT id, order_id, customer_ref, amount, currency, network,
		       pay_address, pay_url, status, tx_hash, provider_name,
		       created_at, updated_at, expires_at
		FROM payments
		WHERE order_id = $1
	`
	return r.scanPayment(r.pool.QueryRow(ctx, query, orderID))
}

func (r *PaymentRepository) Update(ctx context.Context, p *domain.Payment) error {
	p.UpdatedAt = time.Now()
	query := `
		UPDATE payments
		SET status = $1, tx_hash = $2, updated_at = $3
		WHERE id = $4
	`
	tag, err := r.pool.Exec(ctx, query, string(p.Status), p.TxHash, p.UpdatedAt, p.ID)
	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PaymentRepository) List(ctx context.Context, limit, offset int) ([]*domain.Payment, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, order_id, customer_ref, amount, currency, network,
		       pay_address, pay_url, status, tx_hash, provider_name,
		       created_at, updated_at, expires_at
		FROM payments
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list payments: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		var p domain.Payment
		var currencyStr, statusStr string
		var customerRef, payURL, txHash *string

		err := rows.Scan(
			&p.ID, &p.OrderID, &customerRef, &p.Amount, &currencyStr, &p.Network,
			&p.PayAddress, &payURL, &statusStr, &txHash, &p.ProviderName,
			&p.CreatedAt, &p.UpdatedAt, &p.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment row: %w", err)
		}

		p.Currency = domain.Currency(currencyStr)
		p.Status = domain.PaymentStatus(statusStr)
		if customerRef != nil {
			p.CustomerRef = *customerRef
		}
		if payURL != nil {
			p.PayURL = *payURL
		}
		if txHash != nil {
			p.TxHash = *txHash
		}
		payments = append(payments, &p)
	}

	return payments, nil
}

func (r *PaymentRepository) scanPayment(row pgx.Row) (*domain.Payment, error) {
	var p domain.Payment
	var currencyStr, statusStr string
	var customerRef, payURL, txHash *string

	err := row.Scan(
		&p.ID, &p.OrderID, &customerRef, &p.Amount, &currencyStr, &p.Network,
		&p.PayAddress, &payURL, &statusStr, &txHash, &p.ProviderName,
		&p.CreatedAt, &p.UpdatedAt, &p.ExpiresAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("query error: %w", err)
	}

	p.Currency = domain.Currency(currencyStr)
	p.Status = domain.PaymentStatus(statusStr)
	if customerRef != nil {
		p.CustomerRef = *customerRef
	}
	if payURL != nil {
		p.PayURL = *payURL
	}
	if txHash != nil {
		p.TxHash = *txHash
	}

	return &p, nil
}

package domain

import "context"

type PaymentRepository interface {
	Create(ctx context.Context, payment *Payment) error
	GetByID(ctx context.Context, id string) (*Payment, error)
	GetByOrderID(ctx context.Context, orderID string) (*Payment, error)
	Update(ctx context.Context, payment *Payment) error
	List(ctx context.Context, limit, offset int) ([]*Payment, error)
}

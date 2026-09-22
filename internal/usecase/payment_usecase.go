package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"time"

	"wallet-observer/internal/domain"
)

type PaymentUseCase struct {
	repo      domain.PaymentRepository
	providers map[string]domain.PaymentProvider
}

func NewPaymentUseCase(repo domain.PaymentRepository, providers map[string]domain.PaymentProvider) *PaymentUseCase {
	return &PaymentUseCase{
		repo:      repo,
		providers: providers,
	}
}

func (u *PaymentUseCase) CreatePayment(ctx context.Context, input domain.CreatePaymentInput) (*domain.Payment, error) {
	if input.Amount <= 0 {
		return nil, domain.ErrInvalidAmount
	}
	if input.OrderID == "" {
		return nil, fmt.Errorf("order_id is required")
	}

	// Default provider selection based on currency/network or input
	providerName := input.Provider
	if providerName == "" {
		if input.Currency == domain.CurrencyTON || input.Network == "ton" {
			providerName = "ton"
		} else {
			providerName = "ton" // default direct blockchain provider
		}
	}

	provider, exists := u.providers[providerName]
	if !exists {
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}

	// Generate invoice from provider
	invResp, err := provider.CreateInvoice(ctx, domain.CreateInvoiceRequest{
		OrderID:     input.OrderID,
		Amount:      input.Amount,
		Currency:    input.Currency,
		Network:     input.Network,
		Description: fmt.Sprintf("Payment for order %s", input.OrderID),
	})
	if err != nil {
		slog.Error("Failed to create provider invoice", "provider", providerName, "err", err)
		return nil, domain.ErrProviderFailed
	}

	now := time.Now()
	paymentID := generateID("pay")

	payment := &domain.Payment{
		ID:           paymentID,
		OrderID:      input.OrderID,
		CustomerRef:  input.CustomerRef,
		Amount:       input.Amount,
		Currency:     input.Currency,
		Network:      input.Network,
		PayAddress:   invResp.PayAddress,
		PayURL:       invResp.PayURL,
		Status:       domain.StatusPending,
		ProviderName: provider.Name(),
		CreatedAt:    now,
		UpdatedAt:    now,
		ExpiresAt:    now.Add(30 * time.Minute),
	}

	if err := u.repo.Create(ctx, payment); err != nil {
		return nil, fmt.Errorf("failed to save payment: %w", err)
	}

	slog.Info("Payment initialized", "id", payment.ID, "order_id", payment.OrderID, "provider", payment.ProviderName)
	return payment, nil
}

func (u *PaymentUseCase) GetPayment(ctx context.Context, id string) (*domain.Payment, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *PaymentUseCase) GetPaymentByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	return u.repo.GetByOrderID(ctx, orderID)
}

func (u *PaymentUseCase) ListPayments(ctx context.Context, limit, offset int) ([]*domain.Payment, error) {
	return u.repo.List(ctx, limit, offset)
}

func (u *PaymentUseCase) UpdateStatus(ctx context.Context, id string, status domain.PaymentStatus, txHash string) (*domain.Payment, error) {
	payment, err := u.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	payment.Status = status
	if txHash != "" {
		payment.TxHash = txHash
	}

	if err := u.repo.Update(ctx, payment); err != nil {
		return nil, err
	}

	slog.Info("Payment status updated", "id", payment.ID, "status", status, "tx_hash", txHash)
	return payment, nil
}

func generateID(prefix string) string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("%s_%s", prefix, hex.EncodeToString(bytes))
}

package usecase_test

import (
	"context"
	"testing"
	"time"

	"wallet-observer/internal/domain"
	"wallet-observer/internal/usecase"
)

// MockPaymentRepo implements domain.PaymentRepository in-memory for testing
type MockPaymentRepo struct {
	payments map[string]*domain.Payment
}

func NewMockPaymentRepo() *MockPaymentRepo {
	return &MockPaymentRepo{payments: make(map[string]*domain.Payment)}
}

func (m *MockPaymentRepo) Create(ctx context.Context, p *domain.Payment) error {
	m.payments[p.ID] = p
	return nil
}

func (m *MockPaymentRepo) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	p, ok := m.payments[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return p, nil
}

func (m *MockPaymentRepo) GetByOrderID(ctx context.Context, orderID string) (*domain.Payment, error) {
	for _, p := range m.payments {
		if p.OrderID == orderID {
			return p, nil
		}
	}
	return nil, domain.ErrNotFound
}

func (m *MockPaymentRepo) Update(ctx context.Context, p *domain.Payment) error {
	if _, ok := m.payments[p.ID]; !ok {
		return domain.ErrNotFound
	}
	m.payments[p.ID] = p
	return nil
}

func (m *MockPaymentRepo) List(ctx context.Context, limit, offset int) ([]*domain.Payment, error) {
	var list []*domain.Payment
	for _, p := range m.payments {
		list = append(list, p)
	}
	return list, nil
}

// MockProvider implements domain.PaymentProvider for testing
type MockProvider struct {
	name string
}

func (m *MockProvider) Name() string { return m.name }
func (m *MockProvider) CreateInvoice(ctx context.Context, req domain.CreateInvoiceRequest) (*domain.InvoiceResponse, error) {
	return &domain.InvoiceResponse{
		ProviderPaymentID: "mock_123",
		PayAddress:        "test_address",
		PayURL:            "https://test.pay/123",
	}, nil
}
func (m *MockProvider) VerifyPayment(ctx context.Context, id string) (*domain.PaymentStatusResponse, error) {
	return &domain.PaymentStatusResponse{
		Status: domain.StatusCompleted,
		TxHash: "0xhash123",
	}, nil
}

func TestCreatePayment(t *testing.T) {
	repo := NewMockPaymentRepo()
	mockProvider := &MockProvider{name: "ton"}
	providers := map[string]domain.PaymentProvider{
		"ton": mockProvider,
	}

	uc := usecase.NewPaymentUseCase(repo, providers)

	payment, err := uc.CreatePayment(context.Background(), domain.CreatePaymentInput{
		OrderID:  "ord_1001",
		Amount:   2.5,
		Currency: domain.CurrencyTON,
		Network:  "ton",
		Provider: "ton",
	})

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if payment.OrderID != "ord_1001" {
		t.Errorf("expected orderID ord_1001, got %s", payment.OrderID)
	}

	if payment.Status != domain.StatusPending {
		t.Errorf("expected status PENDING, got %s", payment.Status)
	}

	// Verify retrieval by ID
	fetched, err := uc.GetPayment(context.Background(), payment.ID)
	if err != nil {
		t.Fatalf("expected to get payment by ID, got: %v", err)
	}
	if fetched.ID != payment.ID {
		t.Errorf("expected ID %s, got %s", payment.ID, fetched.ID)
	}
}

func TestPaymentExpiration(t *testing.T) {
	p := &domain.Payment{
		Status:    domain.StatusPending,
		ExpiresAt: time.Now().Add(-1 * time.Minute),
	}
	if !p.IsExpired() {
		t.Errorf("expected payment to be expired")
	}
}

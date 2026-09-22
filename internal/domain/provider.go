package domain

import "context"

type CreateInvoiceRequest struct {
	OrderID     string
	Amount      float64
	Currency    Currency
	Network     string
	Description string
}

type InvoiceResponse struct {
	ProviderPaymentID string
	PayAddress        string
	PayURL            string
}

type PaymentStatusResponse struct {
	Status PaymentStatus
	TxHash string
}

type PaymentProvider interface {
	Name() string
	CreateInvoice(ctx context.Context, req CreateInvoiceRequest) (*InvoiceResponse, error)
	VerifyPayment(ctx context.Context, providerPaymentID string) (*PaymentStatusResponse, error)
}

package domain

import (
	"time"
)

type PaymentStatus string

const (
	StatusPending   PaymentStatus = "PENDING"
	StatusCompleted PaymentStatus = "COMPLETED"
	StatusFailed    PaymentStatus = "FAILED"
	StatusExpired   PaymentStatus = "EXPIRED"
)

type Currency string

const (
	CurrencyTON  Currency = "TON"
	CurrencyUSDT Currency = "USDT"
	CurrencyBTC  Currency = "BTC"
	CurrencyETH  Currency = "ETH"
)

type Payment struct {
	ID           string        `json:"id"`
	OrderID      string        `json:"order_id"`
	CustomerRef  string        `json:"customer_ref,omitempty"`
	Amount       float64       `json:"amount"`
	Currency     Currency      `json:"currency"`
	Network      string        `json:"network"`
	PayAddress   string        `json:"pay_address"`
	PayURL       string        `json:"pay_url,omitempty"`
	Status       PaymentStatus `json:"status"`
	TxHash       string        `json:"tx_hash,omitempty"`
	ProviderName string        `json:"provider_name"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	ExpiresAt    time.Time     `json:"expires_at"`
}

type CreatePaymentInput struct {
	OrderID     string   `json:"order_id"`
	CustomerRef string   `json:"customer_ref,omitempty"`
	Amount      float64  `json:"amount"`
	Currency    Currency `json:"currency"`
	Network     string   `json:"network"`
	Provider    string   `json:"provider,omitempty"`
}

func (p *Payment) IsExpired() bool {
	return time.Now().After(p.ExpiresAt) && p.Status == StatusPending
}

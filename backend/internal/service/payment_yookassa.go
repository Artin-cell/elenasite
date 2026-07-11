package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"elena-backend/internal/config"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const yookassaBaseURL = "https://api.yookassa.ru/v3"

type YooKassaProvider struct {
	shopID    string
	secretKey string
	http      *http.Client
}

func NewYooKassaProvider(cfg config.YooKassaConfig) *YooKassaProvider {
	return &YooKassaProvider{
		shopID:    cfg.ShopID,
		secretKey: cfg.SecretKey,
		http: &http.Client{
			Timeout: 15 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
			},
		},
	}
}

type yookassaCreatePaymentRequest struct {
	Amount       yookassaAmount        `json:"amount"`
	Capture      bool                  `json:"capture"` 
	Confirmation yookassaConfirmation  `json:"confirmation"`
	Description  string                `json:"description,omitempty"`
	Metadata     map[string]string     `json:"metadata,omitempty"`
	Receipt      *yookassaReceipt      `json:"receipt,omitempty"`
}

type yookassaAmount struct {
	Value    string `json:"value"`    
	Currency string `json:"currency"` 
}

type yookassaConfirmation struct {
	Type      string `json:"type"` 
	ReturnURL string `json:"return_url"`
}

type yookassaReceipt struct {
	Customer yookassaCustomer `json:"customer"`
	Items    []yookassaItem   `json:"items"`
}

type yookassaCustomer struct {
	Email string `json:"email"`
}

type yookassaItem struct {
	Description    string          `json:"description"`
	Quantity       string          `json:"quantity"`
	Amount         yookassaAmount  `json:"amount"`
	VatCode        int             `json:"vat_code"`
	PaymentMode    string          `json:"payment_mode"`
	PaymentSubject string          `json:"payment_subject"`
}

type yookassaConfirmationResponse struct {
	Type            string `json:"type"`
	ConfirmationURL string `json:"confirmation_url"`
}

type yookassaPaymentResponse struct {
	ID           string                         `json:"id"`
	Status       string                         `json:"status"`
	Paid         bool                           `json:"paid"`
	Amount       yookassaAmount                 `json:"amount"`
	Confirmation *yookassaConfirmationResponse  `json:"confirmation"`
	Metadata     map[string]string              `json:"metadata"`
}

func (p *YooKassaProvider) CreatePayment(ctx context.Context, req CreatePaymentReq) (*CreatePaymentResp, error) {
	body := yookassaCreatePaymentRequest{
		Amount: yookassaAmount{
			Value:    kopeksToRubles(req.AmountKopeks),
			Currency: "RUB",
		},
		Capture: true,
		Confirmation: yookassaConfirmation{
			Type:      "redirect",
			ReturnURL: req.ReturnURL,
		},
		Description: req.Description,
		Metadata:    req.Metadata,
	}

	if req.ClientEmail != "" {
		body.Receipt = &yookassaReceipt{
			Customer: yookassaCustomer{Email: req.ClientEmail},
			Items: []yookassaItem{
				{
					Description:    req.Description,
					Quantity:       "1",
					Amount:         body.Amount,
					VatCode:        1,
					PaymentMode:    "full_payment",
					PaymentSubject: "service",
				},
			},
		}
	}

	var resp yookassaPaymentResponse
	if err := p.doRequest(ctx, http.MethodPost, "/payments", body, &resp, true); err != nil {
		return nil, fmt.Errorf("yookassa create payment: %w", err)
	}

	var confirmURL string
	if resp.Confirmation != nil {
		confirmURL = resp.Confirmation.ConfirmationURL
	}

	return &CreatePaymentResp{
		PaymentID:       resp.ID,
		ConfirmationURL: confirmURL,
		Status:          resp.Status,
	}, nil
}


func (p *YooKassaProvider) GetPayment(ctx context.Context, paymentID string) (*PaymentInfo, error) {
	var raw struct {
		ID       string            `json:"id"`
		Status   string            `json:"status"`
		Paid     bool              `json:"paid"`
		Amount   yookassaAmount    `json:"amount"`
		Metadata map[string]string `json:"metadata"`
	}
	if err := p.doRequest(ctx, http.MethodGet, "/payments/"+paymentID, nil, &raw, false); err != nil {
		return nil, fmt.Errorf("yookassa get payment: %w", err)
	}

	amountKopeks, _ := rublesToKopeks(raw.Amount.Value)

	return &PaymentInfo{
		PaymentID:    raw.ID,
		Status:       raw.Status,
		Paid:         raw.Paid,
		AmountKopeks: amountKopeks,
		Metadata:     raw.Metadata,
	}, nil
}

type yookassaRefundRequest struct {
	PaymentID string          `json:"payment_id"`
	Amount    yookassaAmount  `json:"amount"`
}

func (p *YooKassaProvider) Refund(ctx context.Context, paymentID string, amountKopeks int64) (*RefundInfo, error) {
	body := yookassaRefundRequest{
		PaymentID: paymentID,
		Amount: yookassaAmount{
			Value:    kopeksToRubles(amountKopeks),
			Currency: "RUB",
		},
	}

	var resp struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := p.doRequest(ctx, http.MethodPost, "/refunds", body, &resp, true); err != nil {
		return nil, fmt.Errorf("yookassa refund: %w", err)
	}

	return &RefundInfo{RefundID: resp.ID, Status: resp.Status}, nil
}

func (p *YooKassaProvider) doRequest(ctx context.Context, method, path string, body any, out any, idempotent bool) error {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, yookassaBaseURL+path, reader)
	if err != nil {
		return err
	}

	req.SetBasicAuth(p.shopID, p.secretKey)
	req.Header.Set("Content-Type", "application/json")
	if idempotent {
		req.Header.Set("Idempotence-Key", uuid.NewString())
	}

	resp, err := p.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("yookassa http %d: %s", resp.StatusCode, string(respBody))
	}

	if out != nil {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

func kopeksToRubles(kopeks int64) string {
	rub := kopeks / 100
	kop := kopeks % 100
	return fmt.Sprintf("%d.%02d", rub, kop)
}

func rublesToKopeks(value string) (int64, error) {
	var rub, kop int64
	_, err := fmt.Sscanf(value, "%d.%d", &rub, &kop)
	if err != nil {
		return 0, err
	}
	return rub*100 + kop, nil
}

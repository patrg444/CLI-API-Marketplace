package store

import (
	"database/sql"
	"time"
)

// Invoice represents an invoice in the system
type Invoice struct {
	ID              string    `json:"id"`
	ConsumerID      string    `json:"consumer_id"`
	StripeInvoiceID string    `json:"stripe_invoice_id,omitempty"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	PeriodStart     time.Time `json:"period_start"`
	PeriodEnd       time.Time `json:"period_end"`
	PDFURL          string    `json:"pdf_url,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

// InvoiceWithDetails includes additional information
type InvoiceWithDetails struct {
	Invoice
	ConsumerEmail string           `json:"consumer_email"`
	LineItems     []InvoiceLineItem `json:"line_items"`
}

// InvoiceLineItem represents a line item on an invoice
type InvoiceLineItem struct {
	Description string  `json:"description"`
	Quantity    int64   `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Amount      float64 `json:"amount"`
}

// InvoiceStore handles invoice data operations
type InvoiceStore struct {
	db *sql.DB
}

// NewInvoiceStore creates a new invoice store
func NewInvoiceStore(db *sql.DB) *InvoiceStore {
	return &InvoiceStore{db: db}
}

// Create creates a new invoice
func (s *InvoiceStore) Create(invoice *Invoice) error {
	query := `
		INSERT INTO invoices (
			consumer_id, stripe_invoice_id, amount, currency, 
			status, period_start, period_end, pdf_url
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`
	
	err := s.db.QueryRow(
		query,
		invoice.ConsumerID,
		invoice.StripeInvoiceID,
		invoice.Amount,
		invoice.Currency,
		invoice.Status,
		invoice.PeriodStart,
		invoice.PeriodEnd,
		invoice.PDFURL,
	).Scan(&invoice.ID, &invoice.CreatedAt)
	
	return err
}

// GetByID retrieves an invoice by ID
func (s *InvoiceStore) GetByID(id string) (*Invoice, error) {
	query := `
		SELECT 
			id, consumer_id, stripe_invoice_id, amount, currency,
			status, period_start, period_end, pdf_url, created_at
		FROM invoices
		WHERE id = $1
	`
	
	invoice := &Invoice{}
	err := s.db.QueryRow(query, id).Scan(
		&invoice.ID,
		&invoice.ConsumerID,
		&invoice.StripeInvoiceID,
		&invoice.Amount,
		&invoice.Currency,
		&invoice.Status,
		&invoice.PeriodStart,
		&invoice.PeriodEnd,
		&invoice.PDFURL,
		&invoice.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	return invoice, err
}

// GetByStripeID retrieves an invoice by Stripe invoice ID
func (s *InvoiceStore) GetByStripeID(stripeID string) (*Invoice, error) {
	query := `
		SELECT 
			id, consumer_id, stripe_invoice_id, amount, currency,
			status, period_start, period_end, pdf_url, created_at
		FROM invoices
		WHERE stripe_invoice_id = $1
	`
	
	invoice := &Invoice{}
	err := s.db.QueryRow(query, stripeID).Scan(
		&invoice.ID,
		&invoice.ConsumerID,
		&invoice.StripeInvoiceID,
		&invoice.Amount,
		&invoice.Currency,
		&invoice.Status,
		&invoice.PeriodStart,
		&invoice.PeriodEnd,
		&invoice.PDFURL,
		&invoice.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	return invoice, err
}

// ListByConsumer lists all invoices for a consumer
func (s *InvoiceStore) ListByConsumer(consumerID string, limit, offset int) ([]*Invoice, error) {
	query := `
		SELECT 
			id, consumer_id, stripe_invoice_id, amount, currency,
			status, period_start, period_end, pdf_url, created_at
		FROM invoices
		WHERE consumer_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := s.db.Query(query, consumerID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var invoices []*Invoice
	for rows.Next() {
		invoice := &Invoice{}
		err := rows.Scan(
			&invoice.ID,
			&invoice.ConsumerID,
			&invoice.StripeInvoiceID,
			&invoice.Amount,
			&invoice.Currency,
			&invoice.Status,
			&invoice.PeriodStart,
			&invoice.PeriodEnd,
			&invoice.PDFURL,
			&invoice.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}
	
	return invoices, rows.Err()
}

// Update updates an invoice
func (s *InvoiceStore) Update(invoice *Invoice) error {
	query := `
		UPDATE invoices
		SET 
			stripe_invoice_id = $2,
			amount = $3,
			currency = $4,
			status = $5,
			period_start = $6,
			period_end = $7,
			pdf_url = $8
		WHERE id = $1
	`
	
	_, err := s.db.Exec(
		query,
		invoice.ID,
		invoice.StripeInvoiceID,
		invoice.Amount,
		invoice.Currency,
		invoice.Status,
		invoice.PeriodStart,
		invoice.PeriodEnd,
		invoice.PDFURL,
	)
	
	return err
}

// UpdateStatus updates the status of an invoice
func (s *InvoiceStore) UpdateStatus(id, status string) error {
	query := `
		UPDATE invoices
		SET status = $2
		WHERE id = $1
	`
	
	_, err := s.db.Exec(query, id, status)
	return err
}

// GetTotalRevenue gets the total revenue for a period
func (s *InvoiceStore) GetTotalRevenue(start, end time.Time) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM invoices
		WHERE status = 'paid' 
			AND created_at >= $1 
			AND created_at <= $2
	`
	
	var total float64
	err := s.db.QueryRow(query, start, end).Scan(&total)
	return total, err
}

// GetRevenueByAPI gets revenue grouped by API for a period
func (s *InvoiceStore) GetRevenueByAPI(start, end time.Time) (map[string]float64, error) {
	query := `
		SELECT 
			s.api_id,
			COALESCE(SUM(i.amount), 0) as revenue
		FROM invoices i
		JOIN subscriptions s ON i.consumer_id = s.consumer_id
		WHERE i.status = 'paid' 
			AND i.created_at >= $1 
			AND i.created_at <= $2
			AND s.status = 'active'
		GROUP BY s.api_id
	`
	
	rows, err := s.db.Query(query, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	revenue := make(map[string]float64)
	for rows.Next() {
		var apiID string
		var amount float64
		if err := rows.Scan(&apiID, &amount); err != nil {
			return nil, err
		}
		revenue[apiID] = amount
	}
	
	return revenue, rows.Err()
}

// GetUnpaidInvoices gets all unpaid invoices
func (s *InvoiceStore) GetUnpaidInvoices() ([]*Invoice, error) {
	query := `
		SELECT 
			id, consumer_id, stripe_invoice_id, amount, currency,
			status, period_start, period_end, pdf_url, created_at
		FROM invoices
		WHERE status IN ('draft', 'open', 'past_due')
		ORDER BY created_at ASC
	`
	
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var invoices []*Invoice
	for rows.Next() {
		invoice := &Invoice{}
		err := rows.Scan(
			&invoice.ID,
			&invoice.ConsumerID,
			&invoice.StripeInvoiceID,
			&invoice.Amount,
			&invoice.Currency,
			&invoice.Status,
			&invoice.PeriodStart,
			&invoice.PeriodEnd,
			&invoice.PDFURL,
			&invoice.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}
	
	return invoices, rows.Err()
}

// GetMonthlyRevenue gets revenue statistics by month
func (s *InvoiceStore) GetMonthlyRevenue(months int) ([]map[string]interface{}, error) {
	query := `
		SELECT 
			DATE_TRUNC('month', created_at) as month,
			COUNT(*) as invoice_count,
			SUM(amount) as total_revenue
		FROM invoices
		WHERE status = 'paid'
			AND created_at >= NOW() - INTERVAL '%d months'
		GROUP BY month
		ORDER BY month DESC
	`
	
	rows, err := s.db.Query(query, months)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var results []map[string]interface{}
	for rows.Next() {
		var month time.Time
		var count int
		var revenue float64
		
		if err := rows.Scan(&month, &count, &revenue); err != nil {
			return nil, err
		}
		
		results = append(results, map[string]interface{}{
			"month":         month,
			"invoice_count": count,
			"revenue":       revenue,
		})
	}
	
	return results, rows.Err()
}

// Delete deletes an invoice (use with caution)
func (s *InvoiceStore) Delete(id string) error {
	query := `DELETE FROM invoices WHERE id = $1`
	_, err := s.db.Exec(query, id)
	return err
}

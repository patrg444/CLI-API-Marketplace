package stripe

import (
	"fmt"

	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/checkout/session"
	"github.com/stripe/stripe-go/v76/customer"
	"github.com/stripe/stripe-go/v76/invoice"
	"github.com/stripe/stripe-go/v76/paymentmethod"
	"github.com/stripe/stripe-go/v76/price"
	"github.com/stripe/stripe-go/v76/product"
	"github.com/stripe/stripe-go/v76/subscription"
	"github.com/stripe/stripe-go/v76/usagerecord"
)

// Client wraps the Stripe API client
type Client struct {
	apiKey string
}

// NewClient creates a new Stripe client
func NewClient(apiKey string) *Client {
	stripe.Key = apiKey
	return &Client{apiKey: apiKey}
}

// CreateCustomer creates a new Stripe customer
func (c *Client) CreateCustomer(email, name, cognitoUserID string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		Email: stripe.String(email),
		Name:  stripe.String(name),
		Metadata: map[string]string{
			"cognito_user_id": cognitoUserID,
		},
	}
	return customer.New(params)
}

// GetCustomer retrieves a Stripe customer
func (c *Client) GetCustomer(customerID string) (*stripe.Customer, error) {
	return customer.Get(customerID, nil)
}

// CreatePaymentMethod attaches a payment method to a customer
func (c *Client) AttachPaymentMethod(paymentMethodID, customerID string) (*stripe.PaymentMethod, error) {
	params := &stripe.PaymentMethodAttachParams{
		Customer: stripe.String(customerID),
	}
	return paymentmethod.Attach(paymentMethodID, params)
}

// SetDefaultPaymentMethod sets the default payment method for a customer
func (c *Client) SetDefaultPaymentMethod(customerID, paymentMethodID string) (*stripe.Customer, error) {
	params := &stripe.CustomerParams{
		InvoiceSettings: &stripe.CustomerInvoiceSettingsParams{
			DefaultPaymentMethod: stripe.String(paymentMethodID),
		},
	}
	return customer.Update(customerID, params)
}

// ListPaymentMethods lists all payment methods for a customer
func (c *Client) ListPaymentMethods(customerID string) ([]*stripe.PaymentMethod, error) {
	params := &stripe.PaymentMethodListParams{
		Customer: stripe.String(customerID),
		Type:     stripe.String(string(stripe.PaymentMethodTypeCard)),
	}
	
	var methods []*stripe.PaymentMethod
	iter := paymentmethod.List(params)
	for iter.Next() {
		methods = append(methods, iter.PaymentMethod())
	}
	
	return methods, iter.Err()
}

// DetachPaymentMethod detaches a payment method from a customer
func (c *Client) DetachPaymentMethod(paymentMethodID string) (*stripe.PaymentMethod, error) {
	return paymentmethod.Detach(paymentMethodID, nil)
}

// CreateProduct creates a new product in Stripe
func (c *Client) CreateProduct(apiID, apiName, description string) (*stripe.Product, error) {
	params := &stripe.ProductParams{
		Name:        stripe.String(apiName),
		Description: stripe.String(description),
		Metadata: map[string]string{
			"api_id": apiID,
		},
	}
	return product.New(params)
}

// CreatePrice creates a new price for a product
func (c *Client) CreatePrice(productID string, unitAmount int64, currency string, recurring bool, interval string) (*stripe.Price, error) {
	params := &stripe.PriceParams{
		Product:    stripe.String(productID),
		UnitAmount: stripe.Int64(unitAmount),
		Currency:   stripe.String(currency),
	}
	
	if recurring {
		params.Recurring = &stripe.PriceRecurringParams{
			Interval: stripe.String(interval),
		}
	}
	
	return price.New(params)
}

// CreateMeteredPrice creates a metered usage price
func (c *Client) CreateMeteredPrice(productID string, currency string, interval string, tierMode bool) (*stripe.Price, error) {
	params := &stripe.PriceParams{
		Product:  stripe.String(productID),
		Currency: stripe.String(currency),
		Recurring: &stripe.PriceRecurringParams{
			Interval:  stripe.String(interval),
			UsageType: stripe.String(string(stripe.PriceRecurringUsageTypeMetered)),
		},
	}
	
	if tierMode {
		params.TiersMode = stripe.String(string(stripe.PriceTiersModeGraduated))
		// Tiers would be added here based on pricing plan
	}
	
	return price.New(params)
}

// CreateSubscription creates a new subscription
func (c *Client) CreateSubscription(customerID string, priceID string, metadata map[string]string) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionParams{
		Customer: stripe.String(customerID),
		Items: []*stripe.SubscriptionItemsParams{
			{
				Price: stripe.String(priceID),
			},
		},
		Metadata: metadata,
	}
	
	return subscription.New(params)
}

// GetSubscription retrieves a subscription
func (c *Client) GetSubscription(subscriptionID string) (*stripe.Subscription, error) {
	return subscription.Get(subscriptionID, nil)
}

// CancelSubscription cancels a subscription
func (c *Client) CancelSubscription(subscriptionID string, immediately bool) (*stripe.Subscription, error) {
	params := &stripe.SubscriptionCancelParams{}
	if immediately {
		params.InvoiceNow = stripe.Bool(true)
		params.Prorate = stripe.Bool(true)
	}
	return subscription.Cancel(subscriptionID, params)
}

// UpdateSubscription updates a subscription (for upgrades/downgrades)
func (c *Client) UpdateSubscription(subscriptionID string, newPriceID string) (*stripe.Subscription, error) {
	// First, get the subscription to find the item to update
	sub, err := c.GetSubscription(subscriptionID)
	if err != nil {
		return nil, err
	}
	
	if len(sub.Items.Data) == 0 {
		return nil, fmt.Errorf("subscription has no items")
	}
	
	// Update the first item (assuming single product subscriptions)
	params := &stripe.SubscriptionParams{
		Items: []*stripe.SubscriptionItemsParams{
			{
				ID:    stripe.String(sub.Items.Data[0].ID),
				Price: stripe.String(newPriceID),
			},
		},
		ProrationBehavior: stripe.String(string(stripe.SubscriptionProrationBehaviorCreateProrations)),
	}
	
	return subscription.Update(subscriptionID, params)
}

// RecordUsage records metered usage for a subscription
func (c *Client) RecordUsage(subscriptionItemID string, quantity int64, timestamp int64) (*stripe.UsageRecord, error) {
	params := &stripe.UsageRecordParams{
		Quantity:  stripe.Int64(quantity),
		Timestamp: stripe.Int64(timestamp),
	}
	
	return usagerecord.New(subscriptionItemID, params)
}

// CreateInvoice creates an invoice for a customer
func (c *Client) CreateInvoice(customerID string, subscriptionID string) (*stripe.Invoice, error) {
	params := &stripe.InvoiceParams{
		Customer:     stripe.String(customerID),
		Subscription: stripe.String(subscriptionID),
	}
	
	return invoice.New(params)
}

// FinalizeInvoice finalizes an invoice
func (c *Client) FinalizeInvoice(invoiceID string) (*stripe.Invoice, error) {
	return invoice.FinalizeInvoice(invoiceID, nil)
}

// GetInvoice retrieves an invoice
func (c *Client) GetInvoice(invoiceID string) (*stripe.Invoice, error) {
	return invoice.Get(invoiceID, nil)
}

// ListInvoices lists invoices for a customer
func (c *Client) ListInvoices(customerID string, limit int64) ([]*stripe.Invoice, error) {
	params := &stripe.InvoiceListParams{
		Customer: stripe.String(customerID),
	}
	params.Limit = stripe.Int64(limit)
	
	var invoices []*stripe.Invoice
	iter := invoice.List(params)
	for iter.Next() {
		invoices = append(invoices, iter.Invoice())
	}
	
	return invoices, iter.Err()
}

// CreateCheckoutSession creates a Stripe Checkout session
func (c *Client) CreateCheckoutSession(customerID, priceID, successURL, cancelURL string, metadata map[string]string) (*stripe.CheckoutSession, error) {
	params := &stripe.CheckoutSessionParams{
		Customer: stripe.String(customerID),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		SuccessURL: stripe.String(successURL),
		CancelURL:  stripe.String(cancelURL),
		Metadata:   metadata,
	}
	
	return session.New(params)
}

// GetCheckoutSession retrieves a checkout session
func (c *Client) GetCheckoutSession(sessionID string) (*stripe.CheckoutSession, error) {
	return session.Get(sessionID, nil)
}

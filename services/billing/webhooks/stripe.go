package webhooks

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/api-platform/billing-service/store"
	"github.com/stripe/stripe-go/v76"
	"github.com/stripe/stripe-go/v76/webhook"
)

// StripeWebhookHandler handles Stripe webhook events
type StripeWebhookHandler struct {
	endpointSecret    string
	billingStore      *store.BillingStore
	consumerStore     *store.ConsumerStore
	subscriptionStore *store.SubscriptionStore
	invoiceStore      *store.InvoiceStore
}

// NewStripeWebhookHandler creates a new Stripe webhook handler
func NewStripeWebhookHandler(
	endpointSecret string,
	billingStore *store.BillingStore,
	consumerStore *store.ConsumerStore,
	subscriptionStore *store.SubscriptionStore,
	invoiceStore *store.InvoiceStore,
) *StripeWebhookHandler {
	return &StripeWebhookHandler{
		endpointSecret:    endpointSecret,
		billingStore:      billingStore,
		consumerStore:     consumerStore,
		subscriptionStore: subscriptionStore,
		invoiceStore:      invoiceStore,
	}
}

// HandleStripeWebhook processes incoming Stripe webhook events
func (h *StripeWebhookHandler) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	const MaxBodyBytes = int64(65536)
	r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
	
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	
	// Verify webhook signature
	event, err := webhook.ConstructEvent(payload, r.Header.Get("Stripe-Signature"), h.endpointSecret)
	if err != nil {
		log.Printf("Error verifying webhook signature: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	// Handle the event
	switch event.Type {
	case "customer.created":
		err = h.handleCustomerCreated(event)
	case "customer.subscription.created":
		err = h.handleSubscriptionCreated(event)
	case "customer.subscription.updated":
		err = h.handleSubscriptionUpdated(event)
	case "customer.subscription.deleted":
		err = h.handleSubscriptionDeleted(event)
	case "customer.subscription.trial_will_end":
		err = h.handleSubscriptionTrialWillEnd(event)
	case "invoice.paid":
		err = h.handleInvoicePaid(event)
	case "invoice.payment_failed":
		err = h.handleInvoicePaymentFailed(event)
	case "invoice.created":
		err = h.handleInvoiceCreated(event)
	case "invoice.finalized":
		err = h.handleInvoiceFinalized(event)
	case "checkout.session.completed":
		err = h.handleCheckoutSessionCompleted(event)
	case "payment_intent.succeeded":
		err = h.handlePaymentIntentSucceeded(event)
	case "payment_intent.payment_failed":
		err = h.handlePaymentIntentFailed(event)
	default:
		log.Printf("Unhandled event type: %s", event.Type)
	}
	
	if err != nil {
		log.Printf("Error handling webhook event %s: %v", event.Type, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

// handleCustomerCreated handles customer.created events
func (h *StripeWebhookHandler) handleCustomerCreated(event stripe.Event) error {
	var customer stripe.Customer
	if err := json.Unmarshal(event.Data.Raw, &customer); err != nil {
		return fmt.Errorf("error parsing customer: %v", err)
	}
	
	log.Printf("Customer created: %s", customer.ID)
	// Customer creation is typically handled in the API endpoint
	// This is just for logging/monitoring
	
	return nil
}

// handleSubscriptionCreated handles customer.subscription.created events
func (h *StripeWebhookHandler) handleSubscriptionCreated(event stripe.Event) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return fmt.Errorf("error parsing subscription: %v", err)
	}
	
	log.Printf("Subscription created: %s", subscription.ID)
	
	// Update subscription status in database
	sub, err := h.subscriptionStore.GetByStripeID(subscription.ID)
	if err != nil {
		return fmt.Errorf("error getting subscription: %v", err)
	}
	
	if sub != nil {
		sub.Status = string(subscription.Status)
		if err := h.subscriptionStore.Update(sub); err != nil {
			return fmt.Errorf("error updating subscription: %v", err)
		}
	}
	
	return nil
}

// handleSubscriptionUpdated handles customer.subscription.updated events
func (h *StripeWebhookHandler) handleSubscriptionUpdated(event stripe.Event) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return fmt.Errorf("error parsing subscription: %v", err)
	}
	
	log.Printf("Subscription updated: %s, status: %s", subscription.ID, subscription.Status)
	
	// Update subscription in database
	sub, err := h.subscriptionStore.GetByStripeID(subscription.ID)
	if err != nil {
		return fmt.Errorf("error getting subscription: %v", err)
	}
	
	if sub != nil {
		sub.Status = string(subscription.Status)
		
		// Handle cancellation
		if subscription.CanceledAt > 0 {
			cancelTime := time.Unix(subscription.CanceledAt, 0)
			sub.CancelledAt = &cancelTime
		}
		
		// Update expiry for canceled subscriptions
		if subscription.Status == stripe.SubscriptionStatusCanceled && subscription.CurrentPeriodEnd > 0 {
			expiryTime := time.Unix(subscription.CurrentPeriodEnd, 0)
			sub.ExpiresAt = &expiryTime
		}
		
		if err := h.subscriptionStore.Update(sub); err != nil {
			return fmt.Errorf("error updating subscription: %v", err)
		}
		
		// If subscription is canceled, we might need to revoke API access
		if subscription.Status == stripe.SubscriptionStatusCanceled {
			// TODO: Call API key service to revoke/deactivate keys
			log.Printf("Subscription canceled, should revoke API access for subscription: %s", sub.ID)
		}
	}
	
	return nil
}

// handleSubscriptionDeleted handles customer.subscription.deleted events
func (h *StripeWebhookHandler) handleSubscriptionDeleted(event stripe.Event) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return fmt.Errorf("error parsing subscription: %v", err)
	}
	
	log.Printf("Subscription deleted: %s", subscription.ID)
	
	// Update subscription status
	sub, err := h.subscriptionStore.GetByStripeID(subscription.ID)
	if err != nil {
		return fmt.Errorf("error getting subscription: %v", err)
	}
	
	if sub != nil {
		now := time.Now()
		if err := h.subscriptionStore.Cancel(sub.ID, now); err != nil {
			return fmt.Errorf("error canceling subscription: %v", err)
		}
		
		// TODO: Revoke API access
		log.Printf("Subscription deleted, should revoke API access for subscription: %s", sub.ID)
	}
	
	return nil
}

// handleSubscriptionTrialWillEnd handles customer.subscription.trial_will_end events
func (h *StripeWebhookHandler) handleSubscriptionTrialWillEnd(event stripe.Event) error {
	var subscription stripe.Subscription
	if err := json.Unmarshal(event.Data.Raw, &subscription); err != nil {
		return fmt.Errorf("error parsing subscription: %v", err)
	}
	
	log.Printf("Subscription trial will end: %s", subscription.ID)
	
	// TODO: Send notification email to customer
	// This is typically handled by a notification service
	
	return nil
}

// handleInvoicePaid handles invoice.paid events
func (h *StripeWebhookHandler) handleInvoicePaid(event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("error parsing invoice: %v", err)
	}
	
	log.Printf("Invoice paid: %s, amount: %d", invoice.ID, invoice.AmountPaid)
	
	// Get or create invoice record
	inv, err := h.invoiceStore.GetByStripeID(invoice.ID)
	if err != nil {
		return fmt.Errorf("error getting invoice: %v", err)
	}
	
	// Get consumer by Stripe customer ID
	consumer, err := h.consumerStore.GetByStripeCustomerID(invoice.Customer.ID)
	if err != nil || consumer == nil {
		return fmt.Errorf("error getting consumer: %v", err)
	}
	
	if inv == nil {
		// Create new invoice record
		inv = &store.Invoice{
			ConsumerID:      consumer.ID,
			StripeInvoiceID: invoice.ID,
			Amount:          float64(invoice.AmountPaid) / 100, // Convert cents to dollars
			Currency:        string(invoice.Currency),
			Status:          "paid",
			PeriodStart:     time.Unix(invoice.PeriodStart, 0),
			PeriodEnd:       time.Unix(invoice.PeriodEnd, 0),
			PDFURL:          invoice.InvoicePDF,
		}
		
		if err := h.invoiceStore.Create(inv); err != nil {
			return fmt.Errorf("error creating invoice: %v", err)
		}
	} else {
		// Update existing invoice
		inv.Status = "paid"
		inv.Amount = float64(invoice.AmountPaid) / 100
		inv.PDFURL = invoice.InvoicePDF
		
		if err := h.invoiceStore.Update(inv); err != nil {
			return fmt.Errorf("error updating invoice: %v", err)
		}
	}
	
	// Update subscription status if needed
	if invoice.Subscription != nil {
		sub, err := h.subscriptionStore.GetByStripeID(invoice.Subscription.ID)
		if err == nil && sub != nil && sub.Status != "active" {
			sub.Status = "active"
			if err := h.subscriptionStore.Update(sub); err != nil {
				log.Printf("Error updating subscription status: %v", err)
			}
		}
	}
	
	return nil
}

// handleInvoicePaymentFailed handles invoice.payment_failed events
func (h *StripeWebhookHandler) handleInvoicePaymentFailed(event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("error parsing invoice: %v", err)
	}
	
	log.Printf("Invoice payment failed: %s", invoice.ID)
	
	// Update invoice status
	inv, err := h.invoiceStore.GetByStripeID(invoice.ID)
	if err != nil {
		return fmt.Errorf("error getting invoice: %v", err)
	}
	
	if inv != nil {
		inv.Status = "payment_failed"
		if err := h.invoiceStore.Update(inv); err != nil {
			return fmt.Errorf("error updating invoice: %v", err)
		}
	}
	
	// Update subscription status
	if invoice.Subscription != nil {
		sub, err := h.subscriptionStore.GetByStripeID(invoice.Subscription.ID)
		if err == nil && sub != nil {
			sub.Status = "past_due"
			if err := h.subscriptionStore.Update(sub); err != nil {
				log.Printf("Error updating subscription status: %v", err)
			}
			
			// TODO: Consider limiting API access for past_due subscriptions
			log.Printf("Subscription past due, consider limiting API access: %s", sub.ID)
		}
	}
	
	// TODO: Send payment failure notification
	
	return nil
}

// handleInvoiceCreated handles invoice.created events
func (h *StripeWebhookHandler) handleInvoiceCreated(event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("error parsing invoice: %v", err)
	}
	
	log.Printf("Invoice created: %s", invoice.ID)
	
	// We typically don't store draft invoices
	// This event is mainly for logging
	
	return nil
}

// handleInvoiceFinalized handles invoice.finalized events
func (h *StripeWebhookHandler) handleInvoiceFinalized(event stripe.Event) error {
	var invoice stripe.Invoice
	if err := json.Unmarshal(event.Data.Raw, &invoice); err != nil {
		return fmt.Errorf("error parsing invoice: %v", err)
	}
	
	log.Printf("Invoice finalized: %s, amount: %d", invoice.ID, invoice.Total)
	
	// Get consumer by Stripe customer ID
	consumer, err := h.consumerStore.GetByStripeCustomerID(invoice.Customer.ID)
	if err != nil || consumer == nil {
		return fmt.Errorf("error getting consumer: %v", err)
	}
	
	// Create invoice record when finalized
	inv := &store.Invoice{
		ConsumerID:      consumer.ID,
		StripeInvoiceID: invoice.ID,
		Amount:          float64(invoice.Total) / 100, // Convert cents to dollars
		Currency:        string(invoice.Currency),
		Status:          string(invoice.Status),
		PeriodStart:     time.Unix(invoice.PeriodStart, 0),
		PeriodEnd:       time.Unix(invoice.PeriodEnd, 0),
		PDFURL:          invoice.InvoicePDF,
	}
	
	if err := h.invoiceStore.Create(inv); err != nil {
		// If invoice already exists, update it
		existing, _ := h.invoiceStore.GetByStripeID(invoice.ID)
		if existing != nil {
			existing.Amount = float64(invoice.Total) / 100
			existing.Status = string(invoice.Status)
			existing.PDFURL = invoice.InvoicePDF
			return h.invoiceStore.Update(existing)
		}
		return fmt.Errorf("error creating invoice: %v", err)
	}
	
	return nil
}

// handleCheckoutSessionCompleted handles checkout.session.completed events
func (h *StripeWebhookHandler) handleCheckoutSessionCompleted(event stripe.Event) error {
	var session stripe.CheckoutSession
	if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
		return fmt.Errorf("error parsing checkout session: %v", err)
	}
	
	log.Printf("Checkout session completed: %s", session.ID)
	
	// The subscription should already be created via API
	// This is mainly for verification and logging
	
	return nil
}

// handlePaymentIntentSucceeded handles payment_intent.succeeded events
func (h *StripeWebhookHandler) handlePaymentIntentSucceeded(event stripe.Event) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		return fmt.Errorf("error parsing payment intent: %v", err)
	}
	
	log.Printf("Payment intent succeeded: %s, amount: %d", paymentIntent.ID, paymentIntent.Amount)
	
	// Payment intents are typically handled through invoice events
	// This is mainly for logging
	
	return nil
}

// handlePaymentIntentFailed handles payment_intent.payment_failed events
func (h *StripeWebhookHandler) handlePaymentIntentFailed(event stripe.Event) error {
	var paymentIntent stripe.PaymentIntent
	if err := json.Unmarshal(event.Data.Raw, &paymentIntent); err != nil {
		return fmt.Errorf("error parsing payment intent: %v", err)
	}
	
	log.Printf("Payment intent failed: %s", paymentIntent.ID)
	
	// Payment failures are typically handled through invoice events
	// This is mainly for logging
	
	return nil
}

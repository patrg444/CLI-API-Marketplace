package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/api-platform/billing-service/middleware"
	"github.com/api-platform/billing-service/store"
	"github.com/api-platform/billing-service/stripe"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
)

// BillingHandler handles billing-related HTTP requests
type BillingHandler struct {
	billingStore      *store.BillingStore
	consumerStore     *store.ConsumerStore
	subscriptionStore *store.SubscriptionStore
	invoiceStore      *store.InvoiceStore
	stripeClient      *stripe.Client
	redis             *redis.Client
	apiKeyServiceURL  string
}

// NewBillingHandler creates a new billing handler
func NewBillingHandler(
	billingStore *store.BillingStore,
	consumerStore *store.ConsumerStore,
	subscriptionStore *store.SubscriptionStore,
	invoiceStore *store.InvoiceStore,
	stripeClient *stripe.Client,
	redisClient *redis.Client,
) *BillingHandler {
	return &BillingHandler{
		billingStore:      billingStore,
		consumerStore:     consumerStore,
		subscriptionStore: subscriptionStore,
		invoiceStore:      invoiceStore,
		stripeClient:      stripeClient,
		redis:             redisClient,
		apiKeyServiceURL:  "http://apikey-service:8080", // Configure this
	}
}

// RegisterConsumer creates or retrieves a consumer account
func (h *BillingHandler) RegisterConsumer(w http.ResponseWriter, r *http.Request) {
	userContext, err := middleware.GetUserContext(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Check if consumer already exists
	existingConsumer, err := h.consumerStore.GetByCognitoID(userContext.CognitoUserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error checking consumer")
		return
	}

	if existingConsumer != nil {
		respondWithJSON(w, http.StatusOK, existingConsumer)
		return
	}

	// Parse request body
	var req struct {
		CompanyName string `json:"company_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Create Stripe customer
	stripeCustomer, err := h.stripeClient.CreateCustomer(
		userContext.Email,
		req.CompanyName,
		userContext.CognitoUserID,
	)
	if err != nil {
		log.Printf("Error creating Stripe customer: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error creating payment account")
		return
	}

	// Create consumer in database
	consumer := &store.Consumer{
		CognitoUserID:    userContext.CognitoUserID,
		Email:            userContext.Email,
		StripeCustomerID: stripeCustomer.ID,
		CompanyName:      req.CompanyName,
	}

	if err := h.consumerStore.Create(consumer); err != nil {
		log.Printf("Error creating consumer: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error creating consumer")
		return
	}

	respondWithJSON(w, http.StatusCreated, consumer)
}

// GetConsumer retrieves consumer information
func (h *BillingHandler) GetConsumer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	consumerID := vars["consumerId"]

	consumer, err := h.consumerStore.GetByID(consumerID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving consumer")
		return
	}

	if consumer == nil {
		respondWithError(w, http.StatusNotFound, "Consumer not found")
		return
	}

	respondWithJSON(w, http.StatusOK, consumer)
}

// CreateSubscription creates a new subscription
func (h *BillingHandler) CreateSubscription(w http.ResponseWriter, r *http.Request) {
	userContext, err := middleware.GetUserContext(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get consumer
	consumer, err := h.consumerStore.GetByCognitoID(userContext.CognitoUserID)
	if err != nil || consumer == nil {
		respondWithError(w, http.StatusNotFound, "Consumer not found")
		return
	}

	// Parse request
	var req struct {
		PricingPlanID  string `json:"pricing_plan_id"`
		PaymentMethod  string `json:"payment_method_id,omitempty"`
		SuccessURL     string `json:"success_url"`
		CancelURL      string `json:"cancel_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get pricing plan details
	planData, err := h.billingStore.PricingPlan.GetPricingPlanWithAPI(req.PricingPlanID)
	if err != nil || planData == nil {
		respondWithError(w, http.StatusNotFound, "Pricing plan not found")
		return
	}

	plan := planData["plan"].(*store.PricingPlan)
	apiName := planData["api_name"].(string)

	// Check if consumer already has an active subscription to this API
	hasSubscription, err := h.subscriptionStore.CheckExistingSubscription(consumer.ID, plan.APIID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error checking subscriptions")
		return
	}
	if hasSubscription {
		respondWithError(w, http.StatusConflict, "Already subscribed to this API")
		return
	}

	// Create or retrieve Stripe product and price
	var stripePriceID string
	if plan.StripePriceID != "" {
		stripePriceID = plan.StripePriceID
	} else {
		// Create Stripe product if not exists
		stripeProduct, err := h.stripeClient.CreateProduct(
			plan.APIID,
			apiName,
			fmt.Sprintf("%s - %s plan", apiName, plan.Name),
		)
		if err != nil {
			log.Printf("Error creating Stripe product: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Error setting up payment")
			return
		}

		// Create Stripe price based on plan type
		var stripePrice interface{}
		switch plan.Type {
		case "subscription":
			price := int64(plan.MonthlyPrice * 100) // Convert to cents
			stripePrice, err = h.stripeClient.CreatePrice(
				stripeProduct.ID,
				price,
				"usd",
				true,
				"month",
			)
		case "pay_per_use":
			stripePrice, err = h.stripeClient.CreateMeteredPrice(
				stripeProduct.ID,
				"usd",
				"month",
				false,
			)
		default:
			respondWithError(w, http.StatusBadRequest, "Unsupported plan type")
			return
		}

		if err != nil {
			log.Printf("Error creating Stripe price: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Error setting up pricing")
			return
		}

		// Update plan with Stripe price ID
		stripePriceID = stripePrice.(interface{ ID string }).ID
		h.billingStore.PricingPlan.UpdateStripePriceID(plan.ID, stripePriceID)
	}

	// Generate API key for the subscription
	apiKey, err := h.generateAPIKey(consumer.ID, plan.APIID)
	if err != nil {
		log.Printf("Error generating API key: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error generating API access")
		return
	}

	// Create checkout session or subscription directly
	var response interface{}
	if req.PaymentMethod != "" {
		// Direct subscription creation with existing payment method
		stripeSubscription, err := h.stripeClient.CreateSubscription(
			consumer.StripeCustomerID,
			stripePriceID,
			map[string]string{
				"consumer_id":     consumer.ID,
				"api_id":          plan.APIID,
				"pricing_plan_id": plan.ID,
				"api_key_id":      apiKey["id"].(string),
			},
		)
		if err != nil {
			log.Printf("Error creating Stripe subscription: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Error creating subscription")
			return
		}

		// Create subscription in database
		subscription := &store.Subscription{
			ConsumerID:           consumer.ID,
			APIID:                plan.APIID,
			PricingPlanID:        plan.ID,
			APIKeyID:             apiKey["id"].(string),
			StripeSubscriptionID: stripeSubscription.ID,
			Status:               string(stripeSubscription.Status),
		}

		if err := h.subscriptionStore.Create(subscription); err != nil {
			log.Printf("Error creating subscription record: %v", err)
			// TODO: Cancel Stripe subscription
			respondWithError(w, http.StatusInternalServerError, "Error creating subscription")
			return
		}

		response = map[string]interface{}{
			"subscription": subscription,
			"api_key":      apiKey["key"],
		}
	} else {
		// Create Stripe Checkout session
		checkoutSession, err := h.stripeClient.CreateCheckoutSession(
			consumer.StripeCustomerID,
			stripePriceID,
			req.SuccessURL,
			req.CancelURL,
			map[string]string{
				"consumer_id":     consumer.ID,
				"api_id":          plan.APIID,
				"pricing_plan_id": plan.ID,
				"api_key_id":      apiKey["id"].(string),
			},
		)
		if err != nil {
			log.Printf("Error creating checkout session: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Error creating payment session")
			return
		}

		// Create pending subscription
		subscription := &store.Subscription{
			ConsumerID:           consumer.ID,
			APIID:                plan.APIID,
			PricingPlanID:        plan.ID,
			APIKeyID:             apiKey["id"].(string),
			StripeSubscriptionID: "", // Will be updated via webhook
			Status:               "pending",
		}

		if err := h.subscriptionStore.Create(subscription); err != nil {
			log.Printf("Error creating subscription record: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Error creating subscription")
			return
		}

		response = map[string]interface{}{
			"checkout_url": checkoutSession.URL,
			"session_id":   checkoutSession.ID,
		}
	}

	respondWithJSON(w, http.StatusCreated, response)
}

// ListSubscriptions lists all subscriptions for the current user
func (h *BillingHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	userContext, err := middleware.GetUserContext(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get consumer
	consumer, err := h.consumerStore.GetByCognitoID(userContext.CognitoUserID)
	if err != nil || consumer == nil {
		respondWithError(w, http.StatusNotFound, "Consumer not found")
		return
	}

	// Get subscriptions
	subscriptions, err := h.subscriptionStore.ListByConsumer(consumer.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving subscriptions")
		return
	}

	respondWithJSON(w, http.StatusOK, subscriptions)
}

// GetSubscription retrieves a specific subscription
func (h *BillingHandler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subscriptionID := vars["subscriptionId"]

	subscription, err := h.subscriptionStore.GetWithDetails(subscriptionID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving subscription")
		return
	}

	if subscription == nil {
		respondWithError(w, http.StatusNotFound, "Subscription not found")
		return
	}

	respondWithJSON(w, http.StatusOK, subscription)
}

// CancelSubscription cancels a subscription
func (h *BillingHandler) CancelSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subscriptionID := vars["subscriptionId"]

	subscription, err := h.subscriptionStore.GetByID(subscriptionID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving subscription")
		return
	}

	if subscription == nil {
		respondWithError(w, http.StatusNotFound, "Subscription not found")
		return
	}

	// Cancel in Stripe
	if subscription.StripeSubscriptionID != "" {
		_, err = h.stripeClient.CancelSubscription(subscription.StripeSubscriptionID, false)
		if err != nil {
			log.Printf("Error canceling Stripe subscription: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Error canceling subscription")
			return
		}
	}

	// Update local record
	now := time.Now()
	if err := h.subscriptionStore.Cancel(subscriptionID, now); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating subscription")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Subscription canceled successfully",
	})
}

// UpgradeSubscription upgrades/downgrades a subscription
func (h *BillingHandler) UpgradeSubscription(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subscriptionID := vars["subscriptionId"]

	var req struct {
		NewPricingPlanID string `json:"new_pricing_plan_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Get existing subscription
	subscription, err := h.subscriptionStore.GetByID(subscriptionID)
	if err != nil || subscription == nil {
		respondWithError(w, http.StatusNotFound, "Subscription not found")
		return
	}

	// Get new pricing plan
	newPlan, err := h.billingStore.PricingPlan.GetByID(req.NewPricingPlanID)
	if err != nil || newPlan == nil {
		respondWithError(w, http.StatusNotFound, "Pricing plan not found")
		return
	}

	// Ensure same API
	if newPlan.APIID != subscription.APIID {
		respondWithError(w, http.StatusBadRequest, "Cannot change API")
		return
	}

	// Update Stripe subscription
	if subscription.StripeSubscriptionID != "" && newPlan.StripePriceID != "" {
		_, err = h.stripeClient.UpdateSubscription(subscription.StripeSubscriptionID, newPlan.StripePriceID)
		if err != nil {
			log.Printf("Error updating Stripe subscription: %v", err)
			respondWithError(w, http.StatusInternalServerError, "Error updating subscription")
			return
		}
	}

	// Update local record
	subscription.PricingPlanID = req.NewPricingPlanID
	if err := h.subscriptionStore.Update(subscription); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error updating subscription")
		return
	}

	respondWithJSON(w, http.StatusOK, subscription)
}

// GetSubscriptionUsage retrieves usage data for a subscription
func (h *BillingHandler) GetSubscriptionUsage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	subscriptionID := vars["subscriptionId"]

	// Get date range from query params
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var start, end time.Time
	if startStr != "" {
		start, _ = time.Parse("2006-01-02", startStr)
	} else {
		start = time.Now().AddDate(0, -1, 0) // Default to last month
	}

	if endStr != "" {
		end, _ = time.Parse("2006-01-02", endStr)
	} else {
		end = time.Now()
	}

	// TODO: Fetch usage from metering service
	// For now, return mock data
	usage := map[string]interface{}{
		"subscription_id": subscriptionID,
		"period_start":    start,
		"period_end":      end,
		"total_calls":     1000,
		"total_cost":      25.00,
	}

	respondWithJSON(w, http.StatusOK, usage)
}

// AddPaymentMethod adds a new payment method
func (h *BillingHandler) AddPaymentMethod(w http.ResponseWriter, r *http.Request) {
	userContext, err := middleware.GetUserContext(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get consumer
	consumer, err := h.consumerStore.GetByCognitoID(userContext.CognitoUserID)
	if err != nil || consumer == nil {
		respondWithError(w, http.StatusNotFound, "Consumer not found")
		return
	}

	var req struct {
		PaymentMethodID string `json:"payment_method_id"`
		SetAsDefault    bool   `json:"set_as_default"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Attach payment method to customer
	paymentMethod, err := h.stripeClient.AttachPaymentMethod(req.PaymentMethodID, consumer.StripeCustomerID)
	if err != nil {
		log.Printf("Error attaching payment method: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error adding payment method")
		return
	}

	// Set as default if requested
	if req.SetAsDefault {
		_, err = h.stripeClient.SetDefaultPaymentMethod(consumer.StripeCustomerID, req.PaymentMethodID)
		if err != nil {
			log.Printf("Error setting default payment method: %v", err)
		}
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"id":   paymentMethod.ID,
		"type": paymentMethod.Type,
		"card": paymentMethod.Card,
	})
}

// ListPaymentMethods lists all payment methods
func (h *BillingHandler) ListPaymentMethods(w http.ResponseWriter, r *http.Request) {
	userContext, err := middleware.GetUserContext(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get consumer
	consumer, err := h.consumerStore.GetByCognitoID(userContext.CognitoUserID)
	if err != nil || consumer == nil {
		respondWithError(w, http.StatusNotFound, "Consumer not found")
		return
	}

	// List payment methods from Stripe
	methods, err := h.stripeClient.ListPaymentMethods(consumer.StripeCustomerID)
	if err != nil {
		log.Printf("Error listing payment methods: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error retrieving payment methods")
		return
	}

	respondWithJSON(w, http.StatusOK, methods)
}

// RemovePaymentMethod removes a payment method
func (h *BillingHandler) RemovePaymentMethod(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	paymentMethodID := vars["paymentMethodId"]

	// Detach payment method
	_, err := h.stripeClient.DetachPaymentMethod(paymentMethodID)
	if err != nil {
		log.Printf("Error detaching payment method: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error removing payment method")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Payment method removed successfully",
	})
}

// SetDefaultPaymentMethod sets a default payment method
func (h *BillingHandler) SetDefaultPaymentMethod(w http.ResponseWriter, r *http.Request) {
	userContext, err := middleware.GetUserContext(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	vars := mux.Vars(r)
	paymentMethodID := vars["paymentMethodId"]

	// Get consumer
	consumer, err := h.consumerStore.GetByCognitoID(userContext.CognitoUserID)
	if err != nil || consumer == nil {
		respondWithError(w, http.StatusNotFound, "Consumer not found")
		return
	}

	// Set default payment method
	_, err = h.stripeClient.SetDefaultPaymentMethod(consumer.StripeCustomerID, paymentMethodID)
	if err != nil {
		log.Printf("Error setting default payment method: %v", err)
		respondWithError(w, http.StatusInternalServerError, "Error updating payment method")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Default payment method updated",
	})
}

// ListInvoices lists all invoices for the current user
func (h *BillingHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	userContext, err := middleware.GetUserContext(r)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get consumer
	consumer, err := h.consumerStore.GetByCognitoID(userContext.CognitoUserID)
	if err != nil || consumer == nil {
		respondWithError(w, http.StatusNotFound, "Consumer not found")
		return
	}

	// Get pagination params
	limit := 20
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		limit, _ = strconv.Atoi(l)
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		offset, _ = strconv.Atoi(o)
	}

	// Get invoices
	invoices, err := h.invoiceStore.ListByConsumer(consumer.ID, limit, offset)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving invoices")
		return
	}

	respondWithJSON(w, http.StatusOK, invoices)
}

// GetInvoice retrieves a specific invoice
func (h *BillingHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceID := vars["invoiceId"]

	invoice, err := h.invoiceStore.GetByID(invoiceID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error retrieving invoice")
		return
	}

	if invoice == nil {
		respondWithError(w, http.StatusNotFound, "Invoice not found")
		return
	}

	respondWithJSON(w, http.StatusOK, invoice)
}

// DownloadInvoice redirects to invoice PDF
func (h *BillingHandler) DownloadInvoice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceID := vars["invoiceId"]

	invoice, err := h.invoiceStore.GetByID(invoiceID)
	if err != nil || invoice == nil {
		respondWithError(w, http.StatusNotFound, "Invoice not found")
		return
	}

	if invoice.PDFURL == "" {
		respondWithError(w, http.StatusNotFound, "Invoice PDF not available")
		return
	}

	// Redirect to Stripe invoice PDF
	http.Redirect(w, r, invoice.PDFURL, http.StatusTemporaryRedirect)
}

// GetAPIUsageSummary gets usage summary for an API (for creators)
func (h *BillingHandler) GetAPIUsageSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	apiID := vars["apiId"]

	// Get date range
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var start, end time.Time
	if startStr != "" {
		start, _ = time.Parse("2006-01-02", startStr)
	} else {
		start = time.Now().AddDate(0, -1, 0)
	}

	if endStr != "" {
		end, _ = time.Parse("2006-01-02", endStr)
	} else {
		end = time.Now()
	}

	// TODO: Aggregate usage data from metering service
	// For now, return mock data
	usage := map[string]interface{}{
		"api_id":          apiID,
		"period_start":    start,
		"period_end":      end,
		"total_calls":     50000,
		"unique_consumers": 25,
		"total_revenue":   1250.00,
	}

	respondWithJSON(w, http.StatusOK, usage)
}

// GetAPIEarnings gets earnings data for an API (for creators)
func (h *BillingHandler) GetAPIEarnings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	apiID := vars["apiId"]

	// Get date range
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var start, end time.Time
	if startStr != "" {
		start, _ = time.Parse("2006-01-02", startStr)
	} else {
		start = time.Now().AddDate(0, -1, 0)
	}

	if endStr != "" {
		end, _ = time.Parse("2006-01-02", endStr)
	} else {
		end = time.Now()
	}

	// Get revenue data
	revenue, err := h.invoiceStore.GetRevenueByAPI(start, end)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error calculating earnings")
		return
	}

	apiRevenue := revenue[apiID]
	platformCommission := apiRevenue * 0.20 // 20% commission
	netEarnings := apiRevenue - platformCommission

	earnings := map[string]interface{}{
		"api_id":              apiID,
		"period_start":        start,
		"period_end":          end,
		"gross_revenue":       apiRevenue,
		"platform_commission": platformCommission,
		"net_earnings":        netEarnings,
		"payout_status":       "pending", // TODO: Get from payout service
	}

	respondWithJSON(w, http.StatusOK, earnings)
}

// generateAPIKey calls the API key service to generate a new key
func (h *BillingHandler) generateAPIKey(consumerID, apiID string) (map[string]interface{}, error) {
	// TODO: Call API key service
	// For now, return mock data
	return map[string]interface{}{
		"id":  "key_" + time.Now().Format("20060102150405"),
		"key": "sk_test_" + generateRandomString(32),
	}, nil
}

// Helper functions
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Error marshaling JSON"}`))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

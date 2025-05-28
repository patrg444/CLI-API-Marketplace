package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/yourusername/api-direct/services/payout/middleware"
	"github.com/yourusername/api-direct/services/payout/store"
	"github.com/yourusername/api-direct/services/payout/stripe"
)

// Handlers holds all HTTP handlers
type Handlers struct {
	stripeClient  *stripe.Client
	payoutStore   *store.PayoutStore
	accountStore  *store.AccountStore
	earningsStore *store.EarningsStore
}

// NewHandlers creates a new handlers instance
func NewHandlers(stripeClient *stripe.Client, payoutStore *store.PayoutStore, accountStore *store.AccountStore, earningsStore *store.EarningsStore) *Handlers {
	return &Handlers{
		stripeClient:  stripeClient,
		payoutStore:   payoutStore,
		accountStore:  accountStore,
		earningsStore: earningsStore,
	}
}

// StartOnboarding starts the Stripe Connect onboarding process
func (h *Handlers) StartOnboarding(w http.ResponseWriter, r *http.Request) {
	creatorID := middleware.GetCreatorID(r)
	if creatorID == "" {
		respondWithError(w, http.StatusUnauthorized, "Creator ID not found")
		return
	}

	// Check if account already exists
	existingAccount, err := h.accountStore.GetAccountByCreatorID(creatorID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to check existing account")
		return
	}

	var accountID string
	if existingAccount == nil {
		// Create new Stripe Connect account
		claims := middleware.GetUserClaims(r)
		stripeAccount, err := h.stripeClient.CreateConnectedAccount(claims.Email, creatorID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to create Stripe account")
			return
		}

		// Save to database
		_, err = h.accountStore.CreateAccount(creatorID, stripeAccount.ID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to save account")
			return
		}

		accountID = stripeAccount.ID
	} else {
		accountID = existingAccount.StripeAccountID
	}

	// Create onboarding link
	returnURL := fmt.Sprintf("%s/api/v1/accounts/onboard/callback?creator_id=%s", r.Host, creatorID)
	refreshURL := fmt.Sprintf("%s/api/v1/accounts/onboard?refresh=true", r.Host)

	accountLink, err := h.stripeClient.CreateAccountLink(accountID, returnURL, refreshURL)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create onboarding link")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"onboarding_url": accountLink.URL,
	})
}

// OnboardingCallback handles the return from Stripe onboarding
func (h *Handlers) OnboardingCallback(w http.ResponseWriter, r *http.Request) {
	creatorID := r.URL.Query().Get("creator_id")
	if creatorID == "" {
		respondWithError(w, http.StatusBadRequest, "Missing creator ID")
		return
	}

	// Get account details from database
	account, err := h.accountStore.GetAccountByCreatorID(creatorID)
	if err != nil || account == nil {
		respondWithError(w, http.StatusNotFound, "Account not found")
		return
	}

	// Check account status from Stripe
	stripeStatus, err := h.stripeClient.GetAccountStatus(account.StripeAccountID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get account status")
		return
	}

	// Update account status
	status := "onboarding"
	if stripeStatus.ChargesEnabled && stripeStatus.PayoutsEnabled {
		status = "active"
		h.accountStore.CompleteOnboarding(creatorID)
	}

	err = h.accountStore.UpdateAccountStatus(
		account.StripeAccountID,
		status,
		stripeStatus.DetailsSubmitted,
		stripeStatus.ChargesEnabled,
		stripeStatus.PayoutsEnabled,
	)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update account status")
		return
	}

	// Redirect to creator portal
	http.Redirect(w, r, "/creator-portal/payouts?onboarding=complete", http.StatusFound)
}

// GetAccountStatus returns the current status of a creator's payment account
func (h *Handlers) GetAccountStatus(w http.ResponseWriter, r *http.Request) {
	creatorID := middleware.GetCreatorID(r)
	if creatorID == "" {
		respondWithError(w, http.StatusUnauthorized, "Creator ID not found")
		return
	}

	account, err := h.accountStore.GetAccountByCreatorID(creatorID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get account")
		return
	}

	if account == nil {
		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"has_account": false,
			"status":      "not_created",
		})
		return
	}

	// Get latest status from Stripe
	stripeStatus, err := h.stripeClient.GetAccountStatus(account.StripeAccountID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get Stripe status")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"has_account":        true,
		"status":            account.AccountStatus,
		"details_submitted": stripeStatus.DetailsSubmitted,
		"charges_enabled":   stripeStatus.ChargesEnabled,
		"payouts_enabled":   stripeStatus.PayoutsEnabled,
		"requirements":      stripeStatus.Requirements,
		"created_at":        account.CreatedAt,
	})
}

// GetDashboardLink creates a Stripe dashboard link for the creator
func (h *Handlers) GetDashboardLink(w http.ResponseWriter, r *http.Request) {
	creatorID := middleware.GetCreatorID(r)
	if creatorID == "" {
		respondWithError(w, http.StatusUnauthorized, "Creator ID not found")
		return
	}

	account, err := h.accountStore.GetAccountByCreatorID(creatorID)
	if err != nil || account == nil {
		respondWithError(w, http.StatusNotFound, "Payment account not found")
		return
	}

	loginLink, err := h.stripeClient.CreateLoginLink(account.StripeAccountID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to create dashboard link")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"dashboard_url": loginLink.URL,
	})
}

// GetEarnings returns earnings summary for a creator
func (h *Handlers) GetEarnings(w http.ResponseWriter, r *http.Request) {
	creatorID := middleware.GetCreatorID(r)
	if creatorID == "" {
		respondWithError(w, http.StatusUnauthorized, "Creator ID not found")
		return
	}

	// Get earnings summary
	summary, err := h.earningsStore.GetEarningsSummary(creatorID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get earnings summary")
		return
	}

	// Get detailed earnings by API
	earnings, err := h.earningsStore.GetCreatorEarnings(creatorID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get earnings details")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"summary":  summary,
		"earnings": earnings,
	})
}

// GetAPIEarnings returns earnings for a specific API
func (h *Handlers) GetAPIEarnings(w http.ResponseWriter, r *http.Request) {
	creatorID := middleware.GetCreatorID(r)
	if creatorID == "" {
		respondWithError(w, http.StatusUnauthorized, "Creator ID not found")
		return
	}

	vars := mux.Vars(r)
	apiID := vars["apiId"]

	earnings, err := h.earningsStore.GetAPIEarnings(creatorID, apiID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get API earnings")
		return
	}

	if earnings == nil {
		respondWithError(w, http.StatusNotFound, "No earnings found for this API")
		return
	}

	respondWithJSON(w, http.StatusOK, earnings)
}

// ListPayouts returns a list of payouts for a creator
func (h *Handlers) ListPayouts(w http.ResponseWriter, r *http.Request) {
	creatorID := middleware.GetCreatorID(r)
	if creatorID == "" {
		respondWithError(w, http.StatusUnauthorized, "Creator ID not found")
		return
	}

	// Parse pagination
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	// Get payouts
	payouts, total, err := h.payoutStore.GetPayoutsByCreator(creatorID, limit, offset)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get payouts")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"payouts":     payouts,
		"total":       total,
		"page":        page,
		"total_pages": (total + limit - 1) / limit,
	})
}

// GetPayoutDetails returns details of a specific payout
func (h *Handlers) GetPayoutDetails(w http.ResponseWriter, r *http.Request) {
	creatorID := middleware.GetCreatorID(r)
	if creatorID == "" {
		respondWithError(w, http.StatusUnauthorized, "Creator ID not found")
		return
	}

	vars := mux.Vars(r)
	payoutID := vars["payoutId"]

	// Get payout
	payout, err := h.payoutStore.GetPayout(payoutID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get payout")
		return
	}

	if payout == nil || payout.CreatorID != creatorID {
		respondWithError(w, http.StatusNotFound, "Payout not found")
		return
	}

	// Get line items
	lineItems, err := h.payoutStore.GetPayoutLineItems(payoutID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get payout details")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"payout":     payout,
		"line_items": lineItems,
	})
}

// GetUpcomingPayout returns information about the upcoming payout
func (h *Handlers) GetUpcomingPayout(w http.ResponseWriter, r *http.Request) {
	creatorID := middleware.GetCreatorID(r)
	if creatorID == "" {
		respondWithError(w, http.StatusUnauthorized, "Creator ID not found")
		return
	}

	// Get payout summary
	summary, err := h.payoutStore.GetPayoutSummary(creatorID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get payout summary")
		return
	}

	// Calculate next payout date
	now := time.Now()
	nextPayoutDate := time.Date(now.Year(), now.Month()+1, 1, 0, 0, 0, 0, now.Location())
	
	// Get upcoming amount
	periodStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)
	
	upcomingAmount, err := h.payoutStore.GetUpcomingPayoutAmount(creatorID, periodStart, periodEnd)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to calculate upcoming payout")
		return
	}

	summary.NextPayoutAmount = upcomingAmount
	summary.NextPayoutDate = nextPayoutDate.Format("2006-01-02")

	respondWithJSON(w, http.StatusOK, summary)
}

// GetPlatformRevenue returns platform revenue analytics (admin only)
func (h *Handlers) GetPlatformRevenue(w http.ResponseWriter, r *http.Request) {
	// Parse date range
	monthStr := r.URL.Query().Get("month")
	if monthStr == "" {
		monthStr = time.Now().Format("2006-01")
	}

	month, err := time.Parse("2006-01", monthStr)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid month format")
		return
	}

	revenue, err := h.earningsStore.GetPlatformRevenue(month)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get platform revenue")
		return
	}

	if revenue == nil {
		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"month":                 month.Format("2006-01"),
			"total_gross_revenue":   0,
			"total_platform_fees":   0,
			"total_creator_payouts": 0,
		})
		return
	}

	respondWithJSON(w, http.StatusOK, revenue)
}

// GetPlatformAnalytics returns platform-wide analytics (admin only)
func (h *Handlers) GetPlatformAnalytics(w http.ResponseWriter, r *http.Request) {
	// Get revenue history
	months := 12
	if m := r.URL.Query().Get("months"); m != "" {
		months, _ = strconv.Atoi(m)
	}

	revenueHistory, err := h.earningsStore.GetPlatformRevenueHistory(months)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get revenue history")
		return
	}

	// Get account stats
	accountStats, err := h.accountStore.GetAccountStats()
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get account stats")
		return
	}

	// Get payout stats
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Second)

	payoutStats, err := h.payoutStore.GetPayoutStats(startOfMonth, endOfMonth)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to get payout stats")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]interface{}{
		"revenue_history": revenueHistory,
		"account_stats":   accountStats,
		"payout_stats":    payoutStats,
	})
}

// HandleStripeWebhook processes Stripe webhooks
func (h *Handlers) HandleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	// Read body
	payload := make([]byte, r.ContentLength)
	_, err := r.Body.Read(payload)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to read request body")
		return
	}

	// Verify signature
	event, err := h.stripeClient.ValidateWebhookSignature(payload, r.Header.Get("Stripe-Signature"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid webhook signature")
		return
	}

	// Handle event
	if err := h.stripeClient.HandleWebhookEvent(event); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to handle webhook")
		return
	}

	w.WriteHeader(http.StatusOK)
}

// Helper functions
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

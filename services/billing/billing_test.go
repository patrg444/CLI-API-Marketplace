package billing

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/webhook"
)

// Mock Stripe client
type MockStripeClient struct {
	mock.Mock
}

func (m *MockStripeClient) CreateCustomer(params *stripe.CustomerParams) (*stripe.Customer, error) {
	args := m.Called(params)
	return args.Get(0).(*stripe.Customer), args.Error(1)
}

func (m *MockStripeClient) CreateSubscription(params *stripe.SubscriptionParams) (*stripe.Subscription, error) {
	args := m.Called(params)
	return args.Get(0).(*stripe.Subscription), args.Error(1)
}

// Mock database
type MockDB struct {
	mock.Mock
}

func (m *MockDB) SaveSubscription(ctx context.Context, sub *Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockDB) GetSubscription(ctx context.Context, id string) (*Subscription, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*Subscription), args.Error(1)
}

func TestCreateSubscription(t *testing.T) {
	ctx := context.Background()
	mockStripe := new(MockStripeClient)
	mockDB := new(MockDB)
	
	service := &BillingService{
		stripe: mockStripe,
		db:     mockDB,
	}

	t.Run("successful subscription creation", func(t *testing.T) {
		// Setup
		req := &CreateSubscriptionRequest{
			UserID:    "user-123",
			PlanID:    "plan-pro",
			Email:     "user@example.com",
			PaymentMethodID: "pm_test123",
		}

		mockCustomer := &stripe.Customer{
			ID:    "cus_test123",
			Email: req.Email,
		}

		mockSubscription := &stripe.Subscription{
			ID:     "sub_test123",
			Status: stripe.SubscriptionStatusActive,
			Items: &stripe.SubscriptionItemList{
				Data: []*stripe.SubscriptionItem{
					{
						Price: &stripe.Price{
							ID:       "price_test123",
							Currency: "usd",
							UnitAmount: 9900, // $99.00
						},
					},
				},
			},
			CurrentPeriodEnd: time.Now().Add(30 * 24 * time.Hour).Unix(),
		}

		// Expectations
		mockStripe.On("CreateCustomer", mock.MatchedBy(func(params *stripe.CustomerParams) bool {
			return *params.Email == req.Email
		})).Return(mockCustomer, nil)

		mockStripe.On("CreateSubscription", mock.MatchedBy(func(params *stripe.SubscriptionParams) bool {
			return *params.Customer == mockCustomer.ID
		})).Return(mockSubscription, nil)

		mockDB.On("SaveSubscription", ctx, mock.MatchedBy(func(sub *Subscription) bool {
			return sub.UserID == req.UserID && sub.StripeSubscriptionID == mockSubscription.ID
		})).Return(nil)

		// Execute
		result, err := service.CreateSubscription(ctx, req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, mockSubscription.ID, result.StripeSubscriptionID)
		assert.Equal(t, req.UserID, result.UserID)
		assert.Equal(t, "active", result.Status)

		mockStripe.AssertExpectations(t)
		mockDB.AssertExpectations(t)
	})

	t.Run("handles Stripe customer creation failure", func(t *testing.T) {
		req := &CreateSubscriptionRequest{
			UserID: "user-123",
			PlanID: "plan-pro",
			Email:  "invalid@example.com",
		}

		mockStripe.On("CreateCustomer", mock.Anything).Return((*stripe.Customer)(nil), 
			&stripe.Error{Code: stripe.ErrorCodeCardDeclined})

		_, err := service.CreateSubscription(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "card_declined")
	})

	t.Run("validates required fields", func(t *testing.T) {
		testCases := []struct {
			name string
			req  *CreateSubscriptionRequest
			expectedError string
		}{
			{
				name: "missing user ID",
				req: &CreateSubscriptionRequest{
					PlanID: "plan-pro",
					Email:  "user@example.com",
				},
				expectedError: "user ID is required",
			},
			{
				name: "missing plan ID",
				req: &CreateSubscriptionRequest{
					UserID: "user-123",
					Email:  "user@example.com",
				},
				expectedError: "plan ID is required",
			},
			{
				name: "invalid email",
				req: &CreateSubscriptionRequest{
					UserID: "user-123",
					PlanID: "plan-pro",
					Email:  "not-an-email",
				},
				expectedError: "invalid email format",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := service.CreateSubscription(ctx, tc.req)
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			})
		}
	})
}

func TestCancelSubscription(t *testing.T) {
	ctx := context.Background()
	mockStripe := new(MockStripeClient)
	mockDB := new(MockDB)
	
	service := &BillingService{
		stripe: mockStripe,
		db:     mockDB,
	}

	t.Run("successful cancellation", func(t *testing.T) {
		subID := "sub_test123"
		userID := "user-123"

		existingSub := &Subscription{
			ID:                   "internal-sub-123",
			UserID:               userID,
			StripeSubscriptionID: subID,
			Status:               "active",
		}

		mockDB.On("GetSubscription", ctx, subID).Return(existingSub, nil)
		mockDB.On("SaveSubscription", ctx, mock.MatchedBy(func(sub *Subscription) bool {
			return sub.Status == "cancelled"
		})).Return(nil)

		err := service.CancelSubscription(ctx, userID, subID)

		assert.NoError(t, err)
		mockDB.AssertExpectations(t)
	})

	t.Run("prevents cancellation by wrong user", func(t *testing.T) {
		subID := "sub_test123"
		wrongUserID := "user-456"

		existingSub := &Subscription{
			UserID:               "user-123",
			StripeSubscriptionID: subID,
			Status:               "active",
		}

		mockDB.On("GetSubscription", ctx, subID).Return(existingSub, nil)

		err := service.CancelSubscription(ctx, wrongUserID, subID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})
}

func TestHandleWebhook(t *testing.T) {
	mockDB := new(MockDB)
	service := &BillingService{
		db:             mockDB,
		webhookSecret:  "whsec_test123",
	}

	t.Run("handles invoice payment succeeded", func(t *testing.T) {
		// Create a mock webhook event
		event := &stripe.Event{
			Type: "invoice.payment_succeeded",
			Data: &stripe.EventData{
				Object: map[string]interface{}{
					"subscription": "sub_test123",
					"amount_paid":  9900,
					"currency":     "usd",
				},
			},
		}

		mockDB.On("SavePayment", mock.Anything, mock.MatchedBy(func(payment *Payment) bool {
			return payment.Amount == 9900 && payment.Status == "succeeded"
		})).Return(nil)

		err := service.HandleWebhook(context.Background(), event)

		assert.NoError(t, err)
		mockDB.AssertExpectations(t)
	})

	t.Run("handles subscription updated", func(t *testing.T) {
		event := &stripe.Event{
			Type: "customer.subscription.updated",
			Data: &stripe.EventData{
				Object: map[string]interface{}{
					"id":     "sub_test123",
					"status": "active",
					"current_period_end": time.Now().Add(30 * 24 * time.Hour).Unix(),
				},
			},
		}

		existingSub := &Subscription{
			StripeSubscriptionID: "sub_test123",
			Status:              "trialing",
		}

		mockDB.On("GetSubscription", mock.Anything, "sub_test123").Return(existingSub, nil)
		mockDB.On("SaveSubscription", mock.Anything, mock.MatchedBy(func(sub *Subscription) bool {
			return sub.Status == "active"
		})).Return(nil)

		err := service.HandleWebhook(context.Background(), event)

		assert.NoError(t, err)
		mockDB.AssertExpectations(t)
	})

	t.Run("handles payment method failed", func(t *testing.T) {
		event := &stripe.Event{
			Type: "invoice.payment_failed",
			Data: &stripe.EventData{
				Object: map[string]interface{}{
					"subscription": "sub_test123",
					"attempt_count": 3,
				},
			},
		}

		mockDB.On("UpdateSubscriptionStatus", mock.Anything, "sub_test123", "past_due").Return(nil)
		// Should notify user about payment failure
		mockDB.On("CreateNotification", mock.Anything, mock.Anything).Return(nil)

		err := service.HandleWebhook(context.Background(), event)

		assert.NoError(t, err)
		mockDB.AssertExpectations(t)
	})
}

func TestCalculateUsageCharges(t *testing.T) {
	service := &BillingService{}

	t.Run("calculates usage-based charges correctly", func(t *testing.T) {
		usage := &UsageRecord{
			SubscriptionID: "sub_test123",
			APIKey:         "key_test123",
			Period:         "2023-06",
			TotalCalls:     15000,
			FreeQuota:      1000,
			PricePerCall:   0.001, // $0.001 per call
		}

		charges := service.CalculateUsageCharges(usage)

		// Should charge for 14000 calls (15000 - 1000 free)
		expectedCharges := 14.00
		assert.Equal(t, expectedCharges, charges.TotalAmount)
		assert.Equal(t, 14000, charges.BillableCalls)
	})

	t.Run("handles usage within free quota", func(t *testing.T) {
		usage := &UsageRecord{
			TotalCalls:   500,
			FreeQuota:    1000,
			PricePerCall: 0.001,
		}

		charges := service.CalculateUsageCharges(usage)

		assert.Equal(t, 0.0, charges.TotalAmount)
		assert.Equal(t, 0, charges.BillableCalls)
	})

	t.Run("applies volume discounts", func(t *testing.T) {
		usage := &UsageRecord{
			TotalCalls:   1000000, // 1 million calls
			FreeQuota:    1000,
			PricePerCall: 0.001,
			VolumeDiscounts: []VolumeDiscount{
				{MinCalls: 10000, MaxCalls: 100000, DiscountPercent: 10},
				{MinCalls: 100000, MaxCalls: 1000000, DiscountPercent: 20},
			},
		}

		charges := service.CalculateUsageCharges(usage)

		// Complex calculation with tiered discounts
		// First 1000: free
		// Next 9000: $0.001 * 9000 = $9
		// Next 90000: $0.001 * 90000 * 0.9 = $81
		// Next 899000: $0.001 * 899000 * 0.8 = $719.20
		// Total: $809.20

		assert.InDelta(t, 809.20, charges.TotalAmount, 0.01)
	})
}

func TestRefundHandling(t *testing.T) {
	ctx := context.Background()
	mockStripe := new(MockStripeClient)
	mockDB := new(MockDB)
	
	service := &BillingService{
		stripe: mockStripe,
		db:     mockDB,
	}

	t.Run("processes refund successfully", func(t *testing.T) {
		req := &RefundRequest{
			PaymentID: "pay_test123",
			Amount:    5000, // $50.00
			Reason:    "customer_request",
		}

		payment := &Payment{
			ID:              "internal-pay-123",
			StripePaymentID: "pi_test123",
			Amount:          9900,
			Status:          "succeeded",
		}

		mockDB.On("GetPayment", ctx, req.PaymentID).Return(payment, nil)
		
		// Verify refund amount doesn't exceed payment
		assert.Less(t, req.Amount, payment.Amount)

		mockDB.On("SaveRefund", ctx, mock.MatchedBy(func(refund *Refund) bool {
			return refund.Amount == req.Amount && refund.Status == "pending"
		})).Return(nil)

		result, err := service.ProcessRefund(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.Amount, result.Amount)
	})

	t.Run("prevents refund exceeding payment amount", func(t *testing.T) {
		req := &RefundRequest{
			PaymentID: "pay_test123",
			Amount:    10000, // $100.00
		}

		payment := &Payment{
			Amount: 5000, // $50.00
			Status: "succeeded",
		}

		mockDB.On("GetPayment", ctx, req.PaymentID).Return(payment, nil)

		_, err := service.ProcessRefund(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "exceeds payment amount")
	})

	t.Run("prevents duplicate refunds", func(t *testing.T) {
		req := &RefundRequest{
			PaymentID: "pay_test123",
			Amount:    5000,
		}

		payment := &Payment{
			Amount:   5000,
			Status:   "refunded",
			Refunded: true,
		}

		mockDB.On("GetPayment", ctx, req.PaymentID).Return(payment, nil)

		_, err := service.ProcessRefund(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already refunded")
	})
}

func TestInvoiceGeneration(t *testing.T) {
	ctx := context.Background()
	mockDB := new(MockDB)
	
	service := &BillingService{
		db: mockDB,
	}

	t.Run("generates monthly invoice", func(t *testing.T) {
		subID := "sub_test123"
		period := "2023-06"

		subscription := &Subscription{
			ID:     subID,
			UserID: "user-123",
			PlanID: "plan-pro",
			Status: "active",
		}

		usage := &UsageRecord{
			SubscriptionID: subID,
			Period:        period,
			TotalCalls:    50000,
			FreeQuota:     10000,
			PricePerCall:  0.001,
		}

		mockDB.On("GetSubscription", ctx, subID).Return(subscription, nil)
		mockDB.On("GetUsageForPeriod", ctx, subID, period).Return(usage, nil)

		invoice, err := service.GenerateInvoice(ctx, subID, period)

		assert.NoError(t, err)
		assert.NotNil(t, invoice)
		assert.Equal(t, 40.0, invoice.UsageCharges) // 40000 * 0.001
		assert.Equal(t, 99.0, invoice.SubscriptionCharges)
		assert.Equal(t, 139.0, invoice.TotalAmount)
	})
}

// Test helper functions
func TestValidation(t *testing.T) {
	t.Run("validates email format", func(t *testing.T) {
		validEmails := []string{
			"user@example.com",
			"test.user+tag@example.co.uk",
			"name@subdomain.example.com",
		}

		invalidEmails := []string{
			"notanemail",
			"@example.com",
			"user@",
			"user..name@example.com",
		}

		for _, email := range validEmails {
			assert.True(t, isValidEmail(email), "Expected %s to be valid", email)
		}

		for _, email := range invalidEmails {
			assert.False(t, isValidEmail(email), "Expected %s to be invalid", email)
		}
	})

	t.Run("validates plan IDs", func(t *testing.T) {
		validPlans := []string{"plan-free", "plan-starter", "plan-pro", "plan-enterprise"}
		
		for _, plan := range validPlans {
			assert.True(t, isValidPlan(plan), "Expected %s to be valid", plan)
		}

		assert.False(t, isValidPlan("invalid-plan"))
		assert.False(t, isValidPlan(""))
	})
}
"""
Stripe payment integration for API-Direct
Handles subscriptions, usage-based billing, and payouts
"""

import os
import logging
from datetime import datetime, timedelta
from typing import Dict, Any, Optional, List
import asyncpg
from fastapi import HTTPException
import stripe
from decimal import Decimal

logger = logging.getLogger(__name__)

# Initialize Stripe
stripe.api_key = os.getenv("STRIPE_SECRET_KEY", "sk_test_...")
STRIPE_WEBHOOK_SECRET = os.getenv("STRIPE_WEBHOOK_SECRET", "whsec_...")
PLATFORM_FEE_PERCENT = float(os.getenv("PLATFORM_FEE_PERCENT", "0.20"))  # 20% platform fee


class PaymentManager:
    """Manages Stripe payments, subscriptions, and payouts"""
    
    def __init__(self, db_pool: asyncpg.Pool):
        self.db_pool = db_pool
        
        # Subscription products
        self.products = {
            'free': {
                'name': 'Free Plan',
                'price': 0,
                'features': {
                    'api_limit': 3,
                    'request_limit': 10000,
                    'deployment_type': ['hosted'],
                    'support': 'community'
                }
            },
            'starter': {
                'name': 'Starter Plan',
                'price': 29,
                'stripe_price_id': os.getenv('STRIPE_STARTER_PRICE_ID', 'price_starter'),
                'features': {
                    'api_limit': 10,
                    'request_limit': 100000,
                    'deployment_type': ['hosted'],
                    'support': 'email',
                    'analytics': 'basic'
                }
            },
            'pro': {
                'name': 'Pro Plan',
                'price': 99,
                'stripe_price_id': os.getenv('STRIPE_PRO_PRICE_ID', 'price_pro'),
                'features': {
                    'api_limit': 50,
                    'request_limit': 1000000,
                    'deployment_type': ['hosted', 'byoa'],
                    'support': 'priority',
                    'analytics': 'advanced',
                    'custom_domain': True
                }
            },
            'enterprise': {
                'name': 'Enterprise Plan',
                'price': 'custom',
                'features': {
                    'api_limit': 'unlimited',
                    'request_limit': 'unlimited',
                    'deployment_type': ['hosted', 'byoa', 'on-premise'],
                    'support': 'dedicated',
                    'analytics': 'enterprise',
                    'custom_domain': True,
                    'sla': True
                }
            }
        }
    
    async def create_customer(self, user_id: str, email: str, name: str) -> str:
        """Create Stripe customer for user"""
        try:
            # Create Stripe customer
            customer = stripe.Customer.create(
                email=email,
                name=name,
                metadata={'user_id': user_id}
            )
            
            # Store customer ID
            async with self.db_pool.acquire() as conn:
                await conn.execute("""
                    UPDATE users 
                    SET stripe_customer_id = $1
                    WHERE id = $2
                """, customer.id, user_id)
            
            logger.info(f"Created Stripe customer {customer.id} for user {user_id}")
            return customer.id
            
        except stripe.error.StripeError as e:
            logger.error(f"Stripe error creating customer: {e}")
            raise HTTPException(status_code=400, detail=str(e))
    
    async def create_subscription(
        self,
        user_id: str,
        plan: str,
        payment_method_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """Create or update user subscription"""
        if plan not in self.products or plan == 'free':
            raise HTTPException(status_code=400, detail="Invalid subscription plan")
        
        async with self.db_pool.acquire() as conn:
            # Get user info
            user = await conn.fetchrow("""
                SELECT email, name, stripe_customer_id
                FROM users WHERE id = $1
            """, user_id)
            
            if not user:
                raise HTTPException(status_code=404, detail="User not found")
            
            # Create customer if needed
            customer_id = user['stripe_customer_id']
            if not customer_id:
                customer_id = await self.create_customer(
                    user_id, user['email'], user['name']
                )
            
            # Attach payment method if provided
            if payment_method_id:
                stripe.PaymentMethod.attach(
                    payment_method_id,
                    customer=customer_id
                )
                
                # Set as default payment method
                stripe.Customer.modify(
                    customer_id,
                    invoice_settings={'default_payment_method': payment_method_id}
                )
            
            # Check for existing subscription
            existing = await conn.fetchrow("""
                SELECT stripe_subscription_id 
                FROM subscriptions 
                WHERE user_id = $1 AND status = 'active'
            """, user_id)
            
            if existing and existing['stripe_subscription_id']:
                # Update existing subscription
                subscription = stripe.Subscription.modify(
                    existing['stripe_subscription_id'],
                    items=[{
                        'price': self.products[plan]['stripe_price_id']
                    }]
                )
            else:
                # Create new subscription
                subscription = stripe.Subscription.create(
                    customer=customer_id,
                    items=[{
                        'price': self.products[plan]['stripe_price_id']
                    }],
                    metadata={'user_id': str(user_id)}
                )
            
            # Update database
            await conn.execute("""
                INSERT INTO subscriptions (
                    user_id, plan, status, stripe_subscription_id,
                    stripe_customer_id, amount, currency,
                    current_period_start, current_period_end
                ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
                ON CONFLICT (user_id) WHERE status = 'active'
                DO UPDATE SET
                    plan = EXCLUDED.plan,
                    stripe_subscription_id = EXCLUDED.stripe_subscription_id,
                    amount = EXCLUDED.amount,
                    current_period_start = EXCLUDED.current_period_start,
                    current_period_end = EXCLUDED.current_period_end,
                    updated_at = NOW()
            """, 
                user_id, plan, subscription.status,
                subscription.id, customer_id,
                Decimal(str(self.products[plan]['price'])), 'usd',
                datetime.fromtimestamp(subscription.current_period_start),
                datetime.fromtimestamp(subscription.current_period_end)
            )
            
            # Update user premium status
            await conn.execute("""
                UPDATE users 
                SET is_premium = $1
                WHERE id = $2
            """, plan in ['pro', 'enterprise'], user_id)
            
            logger.info(f"Created subscription {subscription.id} for user {user_id}")
            
            return {
                'subscription_id': subscription.id,
                'status': subscription.status,
                'plan': plan,
                'current_period_end': subscription.current_period_end
            }
    
    async def cancel_subscription(self, user_id: str) -> Dict[str, Any]:
        """Cancel user subscription"""
        async with self.db_pool.acquire() as conn:
            subscription = await conn.fetchrow("""
                SELECT stripe_subscription_id
                FROM subscriptions
                WHERE user_id = $1 AND status = 'active'
            """, user_id)
            
            if not subscription or not subscription['stripe_subscription_id']:
                raise HTTPException(status_code=404, detail="No active subscription found")
            
            # Cancel at period end
            stripe_sub = stripe.Subscription.modify(
                subscription['stripe_subscription_id'],
                cancel_at_period_end=True
            )
            
            # Update database
            await conn.execute("""
                UPDATE subscriptions
                SET status = 'cancelled',
                    updated_at = NOW()
                WHERE user_id = $1
            """, user_id)
            
            return {
                'message': 'Subscription will be cancelled at period end',
                'cancel_at': stripe_sub.cancel_at
            }
    
    async def record_api_usage(
        self,
        api_id: str,
        user_id: str,
        consumer_id: str,
        amount: Decimal
    ) -> str:
        """Record API usage for billing"""
        async with self.db_pool.acquire() as conn:
            # Create billing event
            billing_id = await conn.fetchval("""
                INSERT INTO billing_events (
                    user_id, api_id, event_type, amount,
                    description, metadata, status
                ) VALUES ($1, $2, $3, $4, $5, $6, $7)
                RETURNING id
            """,
                consumer_id, api_id, 'api_usage', amount,
                f'API usage charge',
                {
                    'api_owner_id': str(user_id),
                    'timestamp': datetime.utcnow().isoformat()
                },
                'pending'
            )
            
            # Calculate platform fee and net amount for API owner
            platform_fee = amount * Decimal(str(PLATFORM_FEE_PERCENT))
            net_amount = amount - platform_fee
            
            # Create revenue record for API owner
            await conn.execute("""
                INSERT INTO billing_events (
                    user_id, api_id, event_type, amount,
                    platform_fee, net_amount, description,
                    metadata, status
                ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
            """,
                user_id, api_id, 'api_revenue', amount,
                platform_fee, net_amount,
                f'API revenue (after {PLATFORM_FEE_PERCENT*100}% platform fee)',
                {'consumer_id': str(consumer_id)},
                'pending'
            )
            
            return str(billing_id)
    
    async def process_payout(self, user_id: str) -> Dict[str, Any]:
        """Process payout for API creator"""
        async with self.db_pool.acquire() as conn:
            # Get payout settings
            payout_settings = await conn.fetchrow("""
                SELECT stripe_account_id, minimum_amount
                FROM payout_settings
                WHERE user_id = $1 AND account_verified = TRUE
            """, user_id)
            
            if not payout_settings:
                raise HTTPException(
                    status_code=400,
                    detail="Please complete payout setup first"
                )
            
            # Calculate pending payout amount
            pending = await conn.fetchrow("""
                SELECT 
                    SUM(net_amount) as total,
                    COUNT(*) as transaction_count
                FROM billing_events
                WHERE user_id = $1 
                AND event_type = 'api_revenue'
                AND status = 'pending'
            """, user_id)
            
            if not pending['total'] or pending['total'] < payout_settings['minimum_amount']:
                return {
                    'message': 'Insufficient balance for payout',
                    'current_balance': float(pending['total'] or 0),
                    'minimum_required': float(payout_settings['minimum_amount'])
                }
            
            try:
                # Create Stripe transfer
                transfer = stripe.Transfer.create(
                    amount=int(pending['total'] * 100),  # Convert to cents
                    currency='usd',
                    destination=payout_settings['stripe_account_id'],
                    metadata={
                        'user_id': str(user_id),
                        'transaction_count': str(pending['transaction_count'])
                    }
                )
                
                # Create payout record
                payout_id = await conn.fetchval("""
                    INSERT INTO billing_events (
                        user_id, event_type, amount, status,
                        stripe_payout_id, description
                    ) VALUES ($1, $2, $3, $4, $5, $6)
                    RETURNING id
                """,
                    user_id, 'payout', pending['total'], 'completed',
                    transfer.id, f"Payout for {pending['transaction_count']} transactions"
                )
                
                # Mark revenue events as processed
                await conn.execute("""
                    UPDATE billing_events
                    SET status = 'processed',
                        metadata = metadata || '{"payout_id": $1}'
                    WHERE user_id = $2
                    AND event_type = 'api_revenue'
                    AND status = 'pending'
                """, str(payout_id), user_id)
                
                # Send notification
                from main import websocket_manager
                await websocket_manager.notify_payout_completed(
                    user_id=str(user_id),
                    amount=float(pending['total']),
                    payout_id=str(payout_id)
                )
                
                return {
                    'payout_id': str(payout_id),
                    'amount': float(pending['total']),
                    'transaction_count': pending['transaction_count'],
                    'status': 'completed',
                    'stripe_transfer_id': transfer.id
                }
                
            except stripe.error.StripeError as e:
                logger.error(f"Stripe payout error: {e}")
                raise HTTPException(status_code=400, detail=str(e))
    
    async def setup_connect_account(
        self,
        user_id: str,
        country: str = 'US',
        business_type: str = 'individual'
    ) -> Dict[str, Any]:
        """Setup Stripe Connect account for payouts"""
        try:
            # Create Connect account
            account = stripe.Account.create(
                type='express',
                country=country,
                business_type=business_type,
                capabilities={
                    'transfers': {'requested': True}
                },
                metadata={'user_id': str(user_id)}
            )
            
            # Generate account link for onboarding
            account_link = stripe.AccountLink.create(
                account=account.id,
                refresh_url=f"{os.getenv('FRONTEND_URL')}/settings/payouts",
                return_url=f"{os.getenv('FRONTEND_URL')}/settings/payouts/success",
                type='account_onboarding'
            )
            
            # Store account ID
            async with self.db_pool.acquire() as conn:
                await conn.execute("""
                    INSERT INTO payout_settings (
                        user_id, stripe_account_id, account_verified
                    ) VALUES ($1, $2, $3)
                    ON CONFLICT (user_id) DO UPDATE SET
                        stripe_account_id = EXCLUDED.stripe_account_id,
                        updated_at = NOW()
                """, user_id, account.id, False)
            
            return {
                'account_id': account.id,
                'onboarding_url': account_link.url
            }
            
        except stripe.error.StripeError as e:
            logger.error(f"Stripe Connect error: {e}")
            raise HTTPException(status_code=400, detail=str(e))
    
    async def handle_webhook(self, payload: bytes, signature: str) -> Dict[str, Any]:
        """Handle Stripe webhook events"""
        try:
            # Verify webhook signature
            event = stripe.Webhook.construct_event(
                payload, signature, STRIPE_WEBHOOK_SECRET
            )
            
            # Handle different event types
            if event['type'] == 'customer.subscription.updated':
                await self._handle_subscription_updated(event['data']['object'])
            
            elif event['type'] == 'customer.subscription.deleted':
                await self._handle_subscription_deleted(event['data']['object'])
            
            elif event['type'] == 'invoice.payment_succeeded':
                await self._handle_payment_succeeded(event['data']['object'])
            
            elif event['type'] == 'account.updated':
                await self._handle_connect_account_updated(event['data']['object'])
            
            logger.info(f"Processed webhook event: {event['type']}")
            return {'status': 'success'}
            
        except Exception as e:
            logger.error(f"Webhook error: {e}")
            raise HTTPException(status_code=400, detail=str(e))
    
    async def _handle_subscription_updated(self, subscription: Dict[str, Any]):
        """Handle subscription update webhook"""
        user_id = subscription['metadata'].get('user_id')
        if not user_id:
            return
        
        async with self.db_pool.acquire() as conn:
            await conn.execute("""
                UPDATE subscriptions
                SET status = $1,
                    current_period_start = $2,
                    current_period_end = $3,
                    updated_at = NOW()
                WHERE stripe_subscription_id = $4
            """,
                subscription['status'],
                datetime.fromtimestamp(subscription['current_period_start']),
                datetime.fromtimestamp(subscription['current_period_end']),
                subscription['id']
            )
    
    async def _handle_subscription_deleted(self, subscription: Dict[str, Any]):
        """Handle subscription cancellation webhook"""
        async with self.db_pool.acquire() as conn:
            # Update subscription status
            await conn.execute("""
                UPDATE subscriptions
                SET status = 'cancelled',
                    updated_at = NOW()
                WHERE stripe_subscription_id = $1
            """, subscription['id'])
            
            # Update user premium status
            user_id = subscription['metadata'].get('user_id')
            if user_id:
                await conn.execute("""
                    UPDATE users
                    SET is_premium = FALSE
                    WHERE id = $1
                """, user_id)
    
    async def _handle_payment_succeeded(self, invoice: Dict[str, Any]):
        """Handle successful payment webhook"""
        # Record payment in billing events
        subscription_id = invoice.get('subscription')
        if not subscription_id:
            return
        
        async with self.db_pool.acquire() as conn:
            # Get user from subscription
            user = await conn.fetchrow("""
                SELECT user_id
                FROM subscriptions
                WHERE stripe_subscription_id = $1
            """, subscription_id)
            
            if user:
                await conn.execute("""
                    INSERT INTO billing_events (
                        user_id, event_type, amount, status,
                        stripe_charge_id, description
                    ) VALUES ($1, $2, $3, $4, $5, $6)
                """,
                    user['user_id'], 'subscription_payment',
                    Decimal(str(invoice['amount_paid'] / 100)),
                    'completed', invoice['charge'],
                    f"Subscription payment for {invoice['period_start']} to {invoice['period_end']}"
                )
    
    async def _handle_connect_account_updated(self, account: Dict[str, Any]):
        """Handle Connect account update webhook"""
        user_id = account['metadata'].get('user_id')
        if not user_id:
            return
        
        # Check if account is fully verified
        verified = (
            account.get('charges_enabled', False) and
            account.get('payouts_enabled', False)
        )
        
        async with self.db_pool.acquire() as conn:
            await conn.execute("""
                UPDATE payout_settings
                SET account_verified = $1,
                    updated_at = NOW()
                WHERE stripe_account_id = $2
            """, verified, account['id'])
    
    async def get_billing_overview(self, user_id: str) -> Dict[str, Any]:
        """Get user's billing overview"""
        async with self.db_pool.acquire() as conn:
            # Get subscription info
            subscription = await conn.fetchrow("""
                SELECT plan, status, current_period_end
                FROM subscriptions
                WHERE user_id = $1 AND status IN ('active', 'trialing')
                ORDER BY created_at DESC
                LIMIT 1
            """, user_id)
            
            # Get revenue stats
            revenue = await conn.fetchrow("""
                SELECT 
                    COUNT(*) as transaction_count,
                    SUM(amount) as gross_revenue,
                    SUM(net_amount) as net_revenue,
                    SUM(CASE WHEN status = 'pending' THEN net_amount ELSE 0 END) as pending_payout
                FROM billing_events
                WHERE user_id = $1 AND event_type = 'api_revenue'
                AND created_at >= NOW() - INTERVAL '30 days'
            """, user_id)
            
            # Get usage stats
            usage = await conn.fetchrow("""
                SELECT 
                    COUNT(*) as api_calls,
                    SUM(amount_charged) as usage_charges
                FROM api_calls
                WHERE api_id IN (SELECT id FROM apis WHERE user_id = $1)
                AND created_at >= NOW() - INTERVAL '30 days'
            """, user_id)
            
            return {
                'subscription': {
                    'plan': subscription['plan'] if subscription else 'free',
                    'status': subscription['status'] if subscription else 'none',
                    'renews_at': subscription['current_period_end'].isoformat() if subscription else None
                },
                'revenue': {
                    'gross': float(revenue['gross_revenue'] or 0),
                    'net': float(revenue['net_revenue'] or 0),
                    'pending_payout': float(revenue['pending_payout'] or 0),
                    'transaction_count': revenue['transaction_count'] or 0
                },
                'usage': {
                    'api_calls': usage['api_calls'] or 0,
                    'charges': float(usage['usage_charges'] or 0)
                },
                'features': self.products[subscription['plan'] if subscription else 'free']['features']
            }
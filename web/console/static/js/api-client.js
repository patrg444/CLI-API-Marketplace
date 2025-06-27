/**
 * API-Direct Frontend API Client
 * Handles all communication with the backend API
 */

class APIClient {
    constructor(baseURL = 'http://localhost:8000') {
        this.baseURL = baseURL;
        this.token = localStorage.getItem('api_token');
    }

    /**
     * Get authentication headers
     */
    getHeaders() {
        const headers = {
            'Content-Type': 'application/json',
        };
        
        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }
        
        return headers;
    }

    /**
     * Make authenticated API request
     */
    async request(endpoint, options = {}) {
        const url = `${this.baseURL}${endpoint}`;
        const config = {
            headers: this.getHeaders(),
            ...options
        };

        try {
            const response = await fetch(url, config);
            
            if (!response.ok) {
                if (response.status === 401) {
                    // Token expired, redirect to login
                    window.location.href = '/login';
                    return;
                }
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            
            return await response.json();
        } catch (error) {
            console.error('API Request failed:', error);
            throw error;
        }
    }

    /**
     * GET request
     */
    async get(endpoint) {
        return this.request(endpoint, { method: 'GET' });
    }

    /**
     * POST request
     */
    async post(endpoint, data) {
        return this.request(endpoint, {
            method: 'POST',
            body: JSON.stringify(data)
        });
    }

    /**
     * PUT request
     */
    async put(endpoint, data) {
        return this.request(endpoint, {
            method: 'PUT',
            body: JSON.stringify(data)
        });
    }

    /**
     * DELETE request
     */
    async delete(endpoint) {
        return this.request(endpoint, { method: 'DELETE' });
    }

    // ==================== DASHBOARD APIs ====================

    /**
     * Get dashboard overview data
     */
    async getDashboardOverview() {
        return this.get('/api/dashboard/overview');
    }

    // ==================== API MANAGEMENT ====================

    /**
     * Get all APIs for the user
     */
    async getAPIs() {
        return this.get('/api/apis');
    }

    /**
     * Get detailed information about a specific API
     */
    async getAPI(apiId) {
        return this.get(`/api/apis/${apiId}`);
    }

    /**
     * Create new API deployment
     */
    async createAPI(apiData) {
        return this.post('/api/apis', apiData);
    }

    /**
     * Update API configuration
     */
    async updateAPI(apiId, apiData) {
        return this.put(`/api/apis/${apiId}`, apiData);
    }

    /**
     * Delete API
     */
    async deleteAPI(apiId) {
        return this.delete(`/api/apis/${apiId}`);
    }

    /**
     * Get API logs
     */
    async getAPILogs(apiId, limit = 100) {
        return this.get(`/api/apis/${apiId}/logs?limit=${limit}`);
    }

    // ==================== ANALYTICS ====================

    /**
     * Get traffic analytics
     */
    async getTrafficAnalytics(period = '7d', apiId = null) {
        let endpoint = `/api/analytics/traffic?period=${period}`;
        if (apiId) {
            endpoint += `&api_id=${apiId}`;
        }
        return this.get(endpoint);
    }

    /**
     * Get analytics traffic data (for analytics page)
     */
    async getAnalyticsTraffic(period = '7d') {
        return this.get(`/api/analytics/traffic?period=${period}`);
    }

    /**
     * Get analytics performance data (for analytics page)
     */
    async getAnalyticsPerformance(period = '7d') {
        return this.get(`/api/analytics/performance?period=${period}`);
    }

    /**
     * Get performance analytics
     */
    async getPerformanceAnalytics() {
        return this.get('/api/analytics/performance');
    }

    /**
     * Get geographic analytics
     */
    async getGeographicAnalytics() {
        return this.get('/api/analytics/geography');
    }

    // ==================== BILLING & REVENUE ====================

    /**
     * Get billing overview
     */
    async getBillingOverview() {
        return this.get('/api/billing/overview');
    }

    /**
     * Get transaction history
     */
    async getTransactions(limit = 50, offset = 0) {
        return this.get(`/api/billing/transactions?limit=${limit}&offset=${offset}`);
    }

    /**
     * Get payout history
     */
    async getPayouts() {
        return this.get('/api/billing/payouts');
    }

    /**
     * Request instant payout
     */
    async requestPayout(amount) {
        return this.post('/api/billing/payout-request', { amount });
    }

    /**
     * Get revenue data over time (for earnings chart)
     */
    async getRevenueData(period = '7d') {
        return this.get(`/api/billing/revenue?period=${period}`);
    }

    /**
     * Get earnings overview and billing data
     */
    async getEarnings() {
        return this.get('/api/billing/earnings');
    }

    // ==================== MARKETPLACE ====================

    /**
     * Get marketplace listings
     */
    async getMarketplaceListings() {
        return this.get('/api/marketplace/listings');
    }

    /**
     * Publish API to marketplace
     */
    async publishToMarketplace(apiId, listingData) {
        return this.post('/api/marketplace/publish', {
            api_id: apiId,
            ...listingData
        });
    }

    /**
     * Publish API to marketplace (alias)
     */
    async publishAPI(apiId, listingData) {
        return this.publishToMarketplace(apiId, listingData);
    }

    /**
     * Subscribe to an API in the marketplace
     */
    async subscribeToAPI(apiId, plan = 'basic') {
        return this.post('/api/marketplace/subscribe', {
            api_id: apiId,
            plan: plan
        });
    }

    /**
     * Get reviews for an API
     */
    async getAPIReviews(apiId) {
        return this.get(`/api/marketplace/reviews?api_id=${apiId}`);
    }

    /**
     * Get featured APIs for marketplace
     */
    async getFeaturedAPIs() {
        return this.get('/api/marketplace/featured');
    }

    /**
     * Get specific marketplace API details
     */
    async getMarketplaceAPI(apiId) {
        return this.get(`/api/marketplace/api/${apiId}`);
    }

    /**
     * Get user's marketplace listings
     */
    async getMyMarketplaceListings() {
        return this.get('/api/marketplace/my-listings');
    }

    /**
     * Update marketplace listing
     */
    async updateMarketplaceListing(apiId, listingData) {
        return this.put(`/api/marketplace/listings/${apiId}`, listingData);
    }

    /**
     * Delete marketplace listing
     */
    async deleteMarketplaceListing(apiId) {
        return this.delete(`/api/marketplace/listings/${apiId}`);
    }

    // ==================== REAL-TIME ====================

    /**
     * Get real-time status
     */
    async getRealTimeStatus() {
        return this.get('/api/realtime/status');
    }

    /**
     * Set up WebSocket connection for real-time updates
     */
    setupWebSocket(onMessage, onError = null) {
        const wsURL = this.baseURL.replace('http', 'ws') + '/ws';
        const ws = new WebSocket(wsURL);
        
        ws.onopen = () => {
            console.log('WebSocket connected');
            // Send authentication
            ws.send(JSON.stringify({
                type: 'auth',
                token: this.token
            }));
        };
        
        ws.onmessage = (event) => {
            try {
                const data = JSON.parse(event.data);
                onMessage(data);
            } catch (error) {
                console.error('Failed to parse WebSocket message:', error);
            }
        };
        
        ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            if (onError) onError(error);
        };
        
        ws.onclose = () => {
            console.log('WebSocket disconnected');
            // Attempt to reconnect after 5 seconds
            setTimeout(() => {
                this.setupWebSocket(onMessage, onError);
            }, 5000);
        };
        
        return ws;
    }
}

// ==================== UTILITY FUNCTIONS ====================

/**
 * Format numbers for display
 */
function formatNumber(num) {
    if (num >= 1000000) {
        return (num / 1000000).toFixed(1) + 'M';
    } else if (num >= 1000) {
        return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
}

/**
 * Format currency for display
 */
function formatCurrency(amount) {
    return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD'
    }).format(amount);
}

/**
 * Format relative time
 */
function formatRelativeTime(dateString) {
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now - date;
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    const diffMinutes = Math.floor(diffMs / (1000 * 60));

    if (diffDays > 0) {
        return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
    } else if (diffHours > 0) {
        return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
    } else if (diffMinutes > 0) {
        return `${diffMinutes} minute${diffMinutes > 1 ? 's' : ''} ago`;
    } else {
        return 'Just now';
    }
}

/**
 * Show notification to user
 */
function showNotification(message, type = 'info', duration = 3000) {
    const notification = document.createElement('div');
    notification.className = `fixed top-4 right-4 px-6 py-3 rounded-lg shadow-lg z-50 ${
        type === 'success' ? 'bg-green-500 text-white' : 
        type === 'error' ? 'bg-red-500 text-white' : 
        type === 'warning' ? 'bg-yellow-500 text-white' :
        'bg-blue-500 text-white'
    }`;
    notification.textContent = message;
    
    document.body.appendChild(notification);
    
    // Animate in
    notification.style.transform = 'translateX(100%)';
    requestAnimationFrame(() => {
        notification.style.transition = 'transform 0.3s ease';
        notification.style.transform = 'translateX(0)';
    });
    
    // Remove after duration
    setTimeout(() => {
        notification.style.transform = 'translateX(100%)';
        setTimeout(() => {
            if (notification.parentNode) {
                notification.parentNode.removeChild(notification);
            }
        }, 300);
    }, duration);
}

/**
 * Handle API errors gracefully
 */
function handleAPIError(error, context = '') {
    console.error(`API Error ${context}:`, error);
    
    let message = 'An unexpected error occurred';
    if (error.message) {
        message = error.message;
    }
    
    showNotification(message, 'error');
}

// Create global API client instance
const apiClient = new APIClient();

// Export for use in other scripts
if (typeof module !== 'undefined' && module.exports) {
    module.exports = { APIClient, apiClient, formatNumber, formatCurrency, formatRelativeTime, showNotification, handleAPIError };
}
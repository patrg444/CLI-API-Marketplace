// Dashboard JavaScript - Connects to Backend API
const API_BASE_URL = 'http://localhost:8000';

// Check authentication
function checkAuth() {
    const token = localStorage.getItem('api_token') || sessionStorage.getItem('api_token');
    if (!token) {
        window.location.href = '/login.html';
        return null;
    }
    return token;
}

// API request helper
async function apiRequest(endpoint, options = {}) {
    const token = checkAuth();
    if (!token) return;
    
    try {
        const response = await fetch(`${API_BASE_URL}${endpoint}`, {
            ...options,
            headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json',
                ...options.headers
            }
        });
        
        if (response.status === 401) {
            localStorage.removeItem('api_token');
            sessionStorage.removeItem('api_token');
            window.location.href = '/login.html';
            return;
        }
        
        if (!response.ok) {
            throw new Error(`HTTP error! status: ${response.status}`);
        }
        
        return await response.json();
    } catch (error) {
        console.error('API request failed:', error);
        throw error;
    }
}

// Load dashboard data
async function loadDashboardData() {
    try {
        // Load user info
        const userStr = localStorage.getItem('user');
        if (userStr) {
            const user = JSON.parse(userStr);
            document.querySelector('.user-name').textContent = user.name || user.email;
            document.querySelector('.user-avatar').textContent = (user.name || user.email).charAt(0).toUpperCase();
        }
        
        // Load dashboard stats
        const stats = await apiRequest('/dashboard/stats');
        updateDashboardStats(stats);
        
        // Load analytics
        const analytics = await apiRequest('/analytics/overview?period=7d');
        updateAnalyticsChart(analytics);
        
        // Load recent APIs
        const apisData = await apiRequest('/apis');
        updateRecentAPIs(apisData.apis);
        
        // Load billing summary
        const billing = await apiRequest('/billing/summary');
        updateBillingSummary(billing);
        
        // Set up WebSocket for real-time updates
        setupWebSocket();
        
    } catch (error) {
        console.error('Failed to load dashboard data:', error);
        showNotification('Failed to load dashboard data', 'error');
    }
}

// Update dashboard stats
function updateDashboardStats(stats) {
    // Total APIs
    document.querySelector('[data-stat="total-apis"]').textContent = stats.total_apis || 0;
    
    // Active Subscriptions
    document.querySelector('[data-stat="active-subscriptions"]').textContent = stats.active_subscriptions || 0;
    
    // Monthly Revenue
    document.querySelector('[data-stat="monthly-revenue"]').textContent = `$${(stats.monthly_revenue || 0).toFixed(2)}`;
    
    // API Calls Today
    document.querySelector('[data-stat="api-calls"]').textContent = (stats.api_calls_today || 0).toLocaleString();
    
    // Growth percentage
    const growthElement = document.querySelector('[data-stat="growth"]');
    if (growthElement && stats.growth_percentage !== undefined) {
        growthElement.textContent = `${stats.growth_percentage > 0 ? '+' : ''}${stats.growth_percentage}%`;
        growthElement.className = stats.growth_percentage > 0 ? 'text-green-600' : 'text-red-600';
    }
    
    // Popular APIs
    if (stats.popular_apis) {
        updatePopularAPIs(stats.popular_apis);
    }
}

// Update analytics chart
function updateAnalyticsChart(analytics) {
    // This would integrate with a charting library like Chart.js
    console.log('Analytics data:', analytics);
    
    // For now, update summary stats
    if (document.querySelector('[data-analytics="total-calls"]')) {
        document.querySelector('[data-analytics="total-calls"]').textContent = analytics.total_calls.toLocaleString();
    }
    if (document.querySelector('[data-analytics="unique-users"]')) {
        document.querySelector('[data-analytics="unique-users"]').textContent = analytics.unique_users;
    }
    if (document.querySelector('[data-analytics="avg-latency"]')) {
        document.querySelector('[data-analytics="avg-latency"]').textContent = `${analytics.average_latency}ms`;
    }
    if (document.querySelector('[data-analytics="error-rate"]')) {
        document.querySelector('[data-analytics="error-rate"]').textContent = `${(analytics.error_rate * 100).toFixed(2)}%`;
    }
}

// Update recent APIs list
function updateRecentAPIs(apis) {
    const container = document.querySelector('[data-section="recent-apis"]');
    if (!container) return;
    
    container.innerHTML = apis.map(api => `
        <div class="bg-white p-4 rounded-lg border border-gray-200 hover:shadow-md transition-shadow">
            <div class="flex justify-between items-start mb-2">
                <h4 class="font-semibold text-gray-900">${api.name}</h4>
                <span class="px-2 py-1 text-xs font-medium rounded-full ${
                    api.status === 'active' ? 'bg-green-100 text-green-800' : 'bg-gray-100 text-gray-800'
                }">
                    ${api.status}
                </span>
            </div>
            <p class="text-sm text-gray-600 mb-3">${api.description}</p>
            <div class="flex justify-between text-sm">
                <span class="text-gray-500">
                    <i class="fas fa-chart-line mr-1"></i>
                    ${api.monthly_calls.toLocaleString()} calls
                </span>
                <span class="text-gray-500">
                    <i class="fas fa-users mr-1"></i>
                    ${api.subscribers} subscribers
                </span>
            </div>
        </div>
    `).join('');
}

// Update billing summary
function updateBillingSummary(billing) {
    if (document.querySelector('[data-billing="balance"]')) {
        document.querySelector('[data-billing="balance"]').textContent = `$${billing.current_balance.toFixed(2)}`;
    }
    if (document.querySelector('[data-billing="pending"]')) {
        document.querySelector('[data-billing="pending"]').textContent = `$${billing.pending_payout.toFixed(2)}`;
    }
    if (document.querySelector('[data-billing="next-payout"]')) {
        document.querySelector('[data-billing="next-payout"]').textContent = new Date(billing.next_payout_date).toLocaleDateString();
    }
}

// Update popular APIs
function updatePopularAPIs(popularAPIs) {
    const container = document.querySelector('[data-section="popular-apis"]');
    if (!container) return;
    
    container.innerHTML = popularAPIs.map(api => `
        <div class="flex justify-between items-center py-2">
            <span class="text-sm font-medium text-gray-900">${api.name}</span>
            <span class="text-sm text-gray-500">${api.calls.toLocaleString()} calls</span>
        </div>
    `).join('');
}

// Set up WebSocket for real-time updates
function setupWebSocket() {
    const wsUrl = API_BASE_URL.replace('http', 'ws') + '/ws';
    const ws = new WebSocket(wsUrl);
    
    ws.onopen = () => {
        console.log('WebSocket connected');
    };
    
    ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        
        if (data.type === 'stats_update') {
            // Update real-time stats
            if (data.data.api_calls !== undefined) {
                document.querySelector('[data-stat="api-calls"]').textContent = data.data.api_calls.toLocaleString();
            }
            if (data.data.active_users !== undefined) {
                document.querySelector('[data-stat="active-users"]').textContent = data.data.active_users;
            }
        }
    };
    
    ws.onerror = (error) => {
        console.error('WebSocket error:', error);
    };
    
    ws.onclose = () => {
        console.log('WebSocket disconnected');
        // Reconnect after 5 seconds
        setTimeout(setupWebSocket, 5000);
    };
}

// Show notification
function showNotification(message, type = 'info') {
    // Create notification element if it doesn't exist
    let notification = document.getElementById('notification');
    if (!notification) {
        notification = document.createElement('div');
        notification.id = 'notification';
        notification.className = 'fixed top-4 right-4 px-6 py-3 rounded-lg shadow-lg z-50 hidden transition-all duration-300';
        document.body.appendChild(notification);
    }
    
    // Set message and style
    notification.textContent = message;
    notification.className = `fixed top-4 right-4 px-6 py-3 rounded-lg shadow-lg z-50 ${
        type === 'success' ? 'bg-green-500 text-white' : 
        type === 'error' ? 'bg-red-500 text-white' : 
        'bg-blue-500 text-white'
    }`;
    
    // Show notification
    notification.classList.remove('hidden');
    
    // Hide after 4 seconds
    setTimeout(() => {
        notification.classList.add('hidden');
    }, 4000);
}

// Logout function
function logout() {
    localStorage.removeItem('api_token');
    sessionStorage.removeItem('api_token');
    localStorage.removeItem('user');
    window.location.href = '/login.html';
}

// Initialize dashboard when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    // Check authentication first
    const token = checkAuth();
    if (!token) return;
    
    // Load dashboard data
    loadDashboardData();
    
    // Set up logout button
    const logoutBtn = document.querySelector('[data-action="logout"]');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', logout);
    }
    
    // Refresh data every 30 seconds
    setInterval(() => {
        loadDashboardData();
    }, 30000);
});
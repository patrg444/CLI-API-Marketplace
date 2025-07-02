// Dashboard functionality with real API integration
document.addEventListener('DOMContentLoaded', async () => {
  // Load user info
  try {
    const user = await apiClient.getMe();
    document.querySelector('.user-name').textContent = user.name || user.email;
    document.querySelector('.user-email').textContent = user.email;
  } catch (error) {
    console.error('Failed to load user info:', error);
  }

  // Load dashboard metrics
  loadDashboardMetrics();
  loadRecentDeployments();
  
  // Set up real-time updates via WebSocket
  setupWebSocketUpdates();
});

async function loadDashboardMetrics() {
  try {
    // Get dashboard overview
    const stats = await apiClient.getDashboardStats();
    
    // Update revenue metric
    const revenueCard = document.querySelector('[data-metric="revenue"]');
    if (revenueCard && stats.metrics) {
      revenueCard.querySelector('.text-3xl').textContent = `$${stats.metrics.total_revenue_30d.toFixed(2)}`;
      revenueCard.querySelector('.text-sm.text-green-600').textContent = 'Last 30 days';
    }
    
    // Update API calls metric
    const callsCard = document.querySelector('[data-metric="calls"]');
    if (callsCard && stats.metrics) {
      callsCard.querySelector('.text-3xl').textContent = stats.metrics.total_calls_30d.toLocaleString();
      callsCard.querySelector('.text-sm.text-blue-600').textContent = 'Last 30 days';
    }
    
    // Update active deployments
    const deploymentsCard = document.querySelector('[data-metric="deployments"]');
    if (deploymentsCard && stats.metrics) {
      deploymentsCard.querySelector('.text-3xl').textContent = stats.metrics.active_deployments;
      deploymentsCard.querySelector('.text-sm.text-gray-500').textContent = 
        `${stats.metrics.hosted_deployments} hosted, ${stats.metrics.byoa_deployments} BYOA`;
    }
    
    // Update revenue chart if exists
    if (stats.revenue_trend && window.Chart) {
      updateRevenueChart(stats.revenue_trend);
    }
    
  } catch (error) {
    console.error('Failed to load dashboard metrics:', error);
    showError('Failed to load dashboard metrics');
  }
}

async function loadRecentDeployments() {
  try {
    const deployments = await apiClient.getMyAPIs();
    const container = document.getElementById('recent-deployments');
    
    if (!container) return;
    
    if (deployments.length === 0) {
      container.innerHTML = `
        <div class="text-center py-8 text-gray-500">
          <i class="fas fa-server text-4xl mb-4"></i>
          <p>No deployments yet</p>
          <a href="/deploy.html" class="text-indigo-600 hover:text-indigo-700 mt-2 inline-block">
            Deploy your first API →
          </a>
        </div>
      `;
      return;
    }
    
    // Show recent 5 deployments
    container.innerHTML = deployments.slice(0, 5).map(api => `
      <div class="flex items-center justify-between p-4 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors">
        <div class="flex items-center">
          <div class="w-3 h-3 ${getStatusClass(api.status)} rounded-full mr-4"></div>
          <div>
            <div class="font-medium text-gray-900">${api.name}</div>
            <div class="text-sm text-gray-600">
              ${api.deployment_type === 'hosted' ? 'Hosted' : 'BYOA'} • 
              ${api.endpoint_url || 'Setting up...'}
            </div>
          </div>
        </div>
        <div class="text-right">
          <div class="text-sm font-medium text-gray-900">
            ${api.calls_today || 0} calls today
          </div>
          <div class="text-xs text-gray-500">
            ${new Date(api.updated_at).toLocaleDateString()}
          </div>
        </div>
      </div>
    `).join('');
    
  } catch (error) {
    console.error('Failed to load deployments:', error);
  }
}

function getStatusClass(status) {
  switch (status) {
    case 'running':
      return 'bg-green-500';
    case 'deploying':
      return 'bg-yellow-500 animate-pulse';
    case 'error':
      return 'bg-red-500';
    default:
      return 'bg-gray-500';
  }
}

function setupWebSocketUpdates() {
  const ws = apiClient.connectWebSocket();
  
  if (!ws) return;
  
  ws.onmessage = (event) => {
    const data = JSON.parse(event.data);
    
    switch (data.type) {
      case 'deployment_update':
        handleDeploymentUpdate(data);
        break;
      case 'metrics_update':
        updateMetrics(data);
        break;
      case 'api_call':
        handleAPICall(data);
        break;
    }
  };
}

function handleDeploymentUpdate(data) {
  // Reload deployments when status changes
  if (data.status === 'running' || data.status === 'error') {
    loadRecentDeployments();
  }
  
  // Show notification
  showNotification(`API ${data.api_name} is now ${data.status}`, 
    data.status === 'running' ? 'success' : 'error');
}

function updateMetrics(data) {
  // Update specific metrics in real-time
  if (data.metric === 'revenue') {
    const revenueCard = document.querySelector('[data-metric="revenue"] .text-3xl');
    if (revenueCard) {
      revenueCard.textContent = `$${data.value.toFixed(2)}`;
    }
  } else if (data.metric === 'calls') {
    const callsCard = document.querySelector('[data-metric="calls"] .text-3xl');
    if (callsCard) {
      const current = parseInt(callsCard.textContent.replace(/,/g, '')) || 0;
      callsCard.textContent = (current + 1).toLocaleString();
    }
  }
}

function handleAPICall(data) {
  // Animate the calls metric when an API is called
  const callsCard = document.querySelector('[data-metric="calls"]');
  if (callsCard) {
    callsCard.classList.add('animate-pulse');
    setTimeout(() => callsCard.classList.remove('animate-pulse'), 1000);
  }
}

function updateRevenueChart(trendData) {
  const ctx = document.getElementById('revenueChart');
  if (!ctx) return;
  
  new Chart(ctx, {
    type: 'line',
    data: {
      labels: trendData.map(d => new Date(d.date).toLocaleDateString()),
      datasets: [{
        label: 'Revenue',
        data: trendData.map(d => d.revenue),
        borderColor: 'rgb(79, 70, 229)',
        backgroundColor: 'rgba(79, 70, 229, 0.1)',
        tension: 0.4
      }]
    },
    options: {
      responsive: true,
      plugins: {
        legend: {
          display: false
        }
      },
      scales: {
        y: {
          beginAtZero: true,
          ticks: {
            callback: value => `$${value}`
          }
        }
      }
    }
  });
}

function showNotification(message, type = 'info') {
  const notification = document.createElement('div');
  notification.className = `fixed top-4 right-4 p-4 rounded-lg shadow-lg ${
    type === 'success' ? 'bg-green-500' : 
    type === 'error' ? 'bg-red-500' : 
    'bg-blue-500'
  } text-white z-50`;
  notification.textContent = message;
  
  document.body.appendChild(notification);
  
  setTimeout(() => {
    notification.classList.add('opacity-0', 'transition-opacity');
    setTimeout(() => notification.remove(), 300);
  }, 3000);
}

function showError(message) {
  const alerts = document.querySelectorAll('.text-sm.text-green-600, .text-sm.text-blue-600');
  alerts.forEach(alert => {
    alert.textContent = 'Failed to load';
    alert.className = 'text-sm text-red-600';
  });
}

// Quick action handlers
document.addEventListener('click', (e) => {
  if (e.target.closest('.deploy-new-api')) {
    window.location.href = '/deploy.html';
  }
});

// Auto-refresh every 30 seconds
setInterval(() => {
  loadDashboardMetrics();
  loadRecentDeployments();
}, 30000);
// API Configuration page functionality
const urlParams = new URLSearchParams(window.location.search);
const apiId = urlParams.get('api_id');

let currentConfig = {};
let originalConfig = {};

// Initialize
document.addEventListener('DOMContentLoaded', async () => {
    if (!apiId) {
        window.location.href = '/apis';
        return;
    }
    
    // Load configuration
    await loadConfiguration();
    
    // Set up form submission
    document.getElementById('configForm').addEventListener('submit', handleSubmit);
});

// Load configuration
async function loadConfiguration() {
    try {
        // Get API details
        const apis = await apiClient.getMyAPIs();
        const api = apis.find(a => a.id === apiId);
        
        if (!api) {
            throw new Error('API not found');
        }
        
        // Set API name
        document.getElementById('apiName').textContent = api.name;
        
        // Get detailed configuration
        const config = api.config || {};
        currentConfig = { ...api, ...config };
        originalConfig = JSON.parse(JSON.stringify(currentConfig));
        
        // Populate form
        populateForm();
        
    } catch (error) {
        console.error('Error loading configuration:', error);
        showNotification('Failed to load configuration', 'error');
    }
}

// Populate form with configuration
function populateForm() {
    // General tab
    document.getElementById('apiNameInput').value = currentConfig.name || '';
    document.getElementById('description').value = currentConfig.description || '';
    document.getElementById('version').value = currentConfig.version || '1.0.0';
    
    // Set status radio
    const statusRadio = document.querySelector(`input[name="status"][value="${currentConfig.status || 'active'}"]`);
    if (statusRadio) statusRadio.checked = true;
    
    // Environment variables
    const envVars = currentConfig.env_vars || {};
    populateEnvVars(envVars);
    
    // Runtime settings
    document.getElementById('runtime').value = currentConfig.runtime || 'python:3.9';
    document.getElementById('memory').value = currentConfig.memory || '512';
    document.getElementById('timeout').value = currentConfig.timeout || '30';
    document.getElementById('maxConcurrency').value = currentConfig.max_concurrency || '100';
    document.getElementById('minInstances').value = currentConfig.min_instances || '0';
    document.getElementById('maxInstances').value = currentConfig.max_instances || '10';
    
    // Networking
    document.getElementById('endpoint').value = currentConfig.endpoint || `https://api.api-direct.io/v1/${apiId}`;
    document.getElementById('corsOrigins').value = (currentConfig.cors_origins || []).join('\n');
    document.getElementById('corsCredentials').checked = currentConfig.cors_credentials || false;
    document.getElementById('rateLimit').value = currentConfig.rate_limit || '60';
    document.getElementById('rateLimitBurst').value = currentConfig.rate_limit_burst || '100';
    
    // Monitoring
    document.getElementById('logLevel').value = currentConfig.log_level || 'info';
    document.getElementById('healthCheckEnabled').checked = currentConfig.health_check_enabled !== false;
    document.getElementById('healthCheckPath').value = currentConfig.health_check_path || '/health';
    document.getElementById('errorRateThreshold').value = currentConfig.error_rate_threshold || '5';
    document.getElementById('latencyThreshold').value = currentConfig.latency_threshold || '1000';
}

// Populate environment variables
function populateEnvVars(envVars) {
    const container = document.getElementById('envVarsContainer');
    container.innerHTML = '';
    
    Object.entries(envVars).forEach(([key, value]) => {
        addEnvVarRow(key, value);
    });
    
    // Add empty row if no vars
    if (Object.keys(envVars).length === 0) {
        addEnvVarRow('', '');
    }
}

// Add environment variable row
function addEnvVarRow(key = '', value = '') {
    const container = document.getElementById('envVarsContainer');
    const row = document.createElement('div');
    row.className = 'flex gap-3 env-var-row';
    row.innerHTML = `
        <input type="text" placeholder="KEY" value="${escapeHtml(key)}" 
               class="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white env-key">
        <input type="text" placeholder="VALUE" value="${escapeHtml(value)}" 
               class="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white env-value">
        <button type="button" onclick="removeEnvVar(this)" class="p-2 text-red-600 hover:text-red-700">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
            </svg>
        </button>
    `;
    container.appendChild(row);
}

// Add environment variable
function addEnvVar() {
    addEnvVarRow();
}

// Remove environment variable
function removeEnvVar(button) {
    button.closest('.env-var-row').remove();
}

// Get environment variables from form
function getEnvVars() {
    const envVars = {};
    document.querySelectorAll('.env-var-row').forEach(row => {
        const key = row.querySelector('.env-key').value.trim();
        const value = row.querySelector('.env-value').value.trim();
        if (key) {
            envVars[key] = value;
        }
    });
    return envVars;
}

// Switch tab
function switchTab(tabName) {
    // Update tab buttons
    document.querySelectorAll('.tab-btn').forEach(btn => {
        if (btn.dataset.tab === tabName) {
            btn.classList.add('active');
        } else {
            btn.classList.remove('active');
        }
    });
    
    // Update tab content
    document.querySelectorAll('.tab-content').forEach(content => {
        if (content.id === `${tabName}-tab`) {
            content.classList.remove('hidden');
        } else {
            content.classList.add('hidden');
        }
    });
}

// Handle form submission
async function handleSubmit(e) {
    e.preventDefault();
    
    const formData = new FormData(e.target);
    
    // Build configuration object
    const config = {
        name: formData.get('name'),
        description: formData.get('description'),
        status: formData.get('status'),
        env_vars: getEnvVars(),
        memory: parseInt(formData.get('memory')),
        timeout: parseInt(formData.get('timeout')),
        max_concurrency: parseInt(formData.get('maxConcurrency')),
        min_instances: parseInt(formData.get('minInstances')),
        max_instances: parseInt(formData.get('maxInstances')),
        cors_origins: formData.get('corsOrigins').split('\n').filter(o => o.trim()),
        cors_credentials: formData.get('corsCredentials') === 'on',
        rate_limit: parseInt(formData.get('rateLimit')),
        rate_limit_burst: parseInt(formData.get('rateLimitBurst')),
        log_level: formData.get('logLevel'),
        health_check_enabled: formData.get('healthCheckEnabled') === 'on',
        health_check_path: formData.get('healthCheckPath'),
        error_rate_threshold: parseFloat(formData.get('errorRateThreshold')),
        latency_threshold: parseInt(formData.get('latencyThreshold'))
    };
    
    try {
        // Show loading
        const submitBtn = e.target.querySelector('button[type="submit"]');
        const originalText = submitBtn.textContent;
        submitBtn.disabled = true;
        submitBtn.textContent = 'Saving...';
        
        // Update configuration
        await apiClient.updateAPIConfig(apiId, config);
        
        // Update originals
        originalConfig = JSON.parse(JSON.stringify(config));
        
        showNotification('Configuration saved successfully', 'success');
        
        // Check if restart needed
        if (hasEnvVarChanges(config.env_vars)) {
            const restart = confirm('Environment variables have changed. Would you like to restart the API now?');
            if (restart) {
                await restartAPI();
            }
        }
        
    } catch (error) {
        console.error('Error saving configuration:', error);
        showNotification('Failed to save configuration', 'error');
    } finally {
        // Reset button
        const submitBtn = e.target.querySelector('button[type="submit"]');
        submitBtn.disabled = false;
        submitBtn.textContent = 'Save Changes';
    }
}

// Check if environment variables changed
function hasEnvVarChanges(newEnvVars) {
    const oldEnvVars = originalConfig.env_vars || {};
    
    // Check if keys are different
    const oldKeys = Object.keys(oldEnvVars).sort();
    const newKeys = Object.keys(newEnvVars).sort();
    
    if (oldKeys.join(',') !== newKeys.join(',')) {
        return true;
    }
    
    // Check if values are different
    for (const key of oldKeys) {
        if (oldEnvVars[key] !== newEnvVars[key]) {
            return true;
        }
    }
    
    return false;
}

// Restart API
async function restartAPI() {
    try {
        await apiClient.restartAPI(apiId);
        showNotification('API restart initiated', 'success');
    } catch (error) {
        console.error('Error restarting API:', error);
        showNotification('Failed to restart API', 'error');
    }
}

// Reset form
function resetForm() {
    if (confirm('Are you sure you want to reset all changes?')) {
        populateForm();
        showNotification('Form reset to original values', 'info');
    }
}

// Copy endpoint
function copyEndpoint() {
    const endpoint = document.getElementById('endpoint');
    endpoint.select();
    document.execCommand('copy');
    showNotification('Endpoint copied to clipboard', 'success');
}

// Show notification
function showNotification(message, type = 'info') {
    // Create notification element
    const notification = document.createElement('div');
    notification.className = `fixed top-4 right-4 max-w-sm w-full bg-white shadow-lg rounded-lg pointer-events-auto ring-1 ring-black ring-opacity-5 overflow-hidden z-50`;
    
    const bgColor = type === 'success' ? 'bg-green-50' : type === 'error' ? 'bg-red-50' : 'bg-blue-50';
    const textColor = type === 'success' ? 'text-green-800' : type === 'error' ? 'text-red-800' : 'text-blue-800';
    const iconColor = type === 'success' ? 'text-green-400' : type === 'error' ? 'text-red-400' : 'text-blue-400';
    
    notification.innerHTML = `
        <div class="p-4">
            <div class="flex items-start">
                <div class="flex-shrink-0">
                    <svg class="h-6 w-6 ${iconColor}" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                        ${type === 'success' ? 
                            '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />' :
                            type === 'error' ?
                            '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />' :
                            '<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />'
                        }
                    </svg>
                </div>
                <div class="ml-3 w-0 flex-1 pt-0.5">
                    <p class="text-sm font-medium ${textColor}">${message}</p>
                </div>
            </div>
        </div>
    `;
    
    document.body.appendChild(notification);
    
    // Auto remove after 5 seconds
    setTimeout(() => {
        notification.remove();
    }, 5000);
}

// Escape HTML
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}
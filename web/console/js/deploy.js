// Deploy page functionality
let currentDeploymentId = null;
let deploymentCheckInterval = null;

// Initialize page
document.addEventListener('DOMContentLoaded', function() {
    // Handle deployment type selection
    document.querySelectorAll('input[name="deploymentType"]').forEach(radio => {
        radio.addEventListener('change', handleDeploymentTypeChange);
    });
    
    // Handle code source selection
    document.querySelectorAll('input[name="codeSource"]').forEach(radio => {
        radio.addEventListener('change', handleCodeSourceChange);
    });
    
    // Handle file upload
    const codeFile = document.getElementById('codeFile');
    codeFile.addEventListener('change', handleFileUpload);
    
    // Drag and drop
    const uploadArea = document.getElementById('codeUpload').querySelector('.border-dashed');
    uploadArea.addEventListener('dragover', handleDragOver);
    uploadArea.addEventListener('drop', handleDrop);
    
    // Form submission
    document.getElementById('deploymentForm').addEventListener('submit', handleDeploy);
    
    // Update deployment type UI
    updateDeploymentTypeUI();
});

// Handle deployment type change
function handleDeploymentTypeChange(e) {
    const byoaConfig = document.getElementById('byoaConfig');
    const isbyoa = e.target.value === 'byoa';
    
    if (isbyoa) {
        byoaConfig.classList.remove('hidden');
        document.getElementById('customEndpoint').required = true;
    } else {
        byoaConfig.classList.add('hidden');
        document.getElementById('customEndpoint').required = false;
    }
    
    updateDeploymentTypeUI();
}

// Update deployment type UI
function updateDeploymentTypeUI() {
    document.querySelectorAll('.deployment-type-option').forEach(option => {
        const radio = option.querySelector('input[type="radio"]');
        const check = option.querySelector('.deployment-check');
        
        if (radio.checked) {
            option.classList.add('border-blue-500', 'bg-blue-50', 'dark:bg-blue-900/20');
            check.classList.remove('hidden');
        } else {
            option.classList.remove('border-blue-500', 'bg-blue-50', 'dark:bg-blue-900/20');
            check.classList.add('hidden');
        }
    });
}

// Handle code source change
function handleCodeSourceChange(e) {
    const codeUpload = document.getElementById('codeUpload');
    const gitUrl = document.getElementById('gitUrl');
    const codeEditor = document.getElementById('codeEditor');
    
    // Hide all
    codeUpload.classList.add('hidden');
    gitUrl.classList.add('hidden');
    codeEditor.classList.add('hidden');
    
    // Show selected
    switch(e.target.value) {
        case 'upload':
            codeUpload.classList.remove('hidden');
            break;
        case 'git':
            gitUrl.classList.remove('hidden');
            break;
        case 'editor':
            codeEditor.classList.remove('hidden');
            break;
    }
}

// Handle file upload
function handleFileUpload(e) {
    const file = e.target.files[0];
    if (file) {
        const uploadArea = document.getElementById('codeUpload').querySelector('.border-dashed');
        uploadArea.innerHTML = `
            <svg class="mx-auto h-12 w-12 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <p class="mt-2 text-sm text-gray-900 dark:text-white font-medium">${file.name}</p>
            <p class="text-xs text-gray-500 dark:text-gray-400">${(file.size / 1024 / 1024).toFixed(2)} MB</p>
            <button type="button" onclick="resetFileUpload()" class="mt-2 text-sm text-blue-600 hover:text-blue-500">
                Change file
            </button>
        `;
    }
}

// Reset file upload
function resetFileUpload() {
    document.getElementById('codeFile').value = '';
    const uploadArea = document.getElementById('codeUpload').querySelector('.border-dashed');
    uploadArea.innerHTML = `
        <svg class="mx-auto h-12 w-12 text-gray-400" stroke="currentColor" fill="none" viewBox="0 0 48 48">
            <path d="M28 8H12a4 4 0 00-4 4v20m32-12v8m0 0v8a4 4 0 01-4 4H12a4 4 0 01-4-4v-4m32-4l-3.172-3.172a4 4 0 00-5.656 0L28 28M8 32l9.172-9.172a4 4 0 015.656 0L28 28m0 0l4 4m4-24h8m-4-4v8m-12 4h.02" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
        <p class="mt-2 text-sm text-gray-600 dark:text-gray-400">
            <button type="button" onclick="document.getElementById('codeFile').click()" class="font-medium text-blue-600 hover:text-blue-500">
                Upload a ZIP file
            </button>
            or drag and drop
        </p>
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">ZIP up to 50MB</p>
    `;
}

// Handle drag over
function handleDragOver(e) {
    e.preventDefault();
    e.currentTarget.classList.add('border-blue-500', 'bg-blue-50');
}

// Handle drop
function handleDrop(e) {
    e.preventDefault();
    e.currentTarget.classList.remove('border-blue-500', 'bg-blue-50');
    
    const files = e.dataTransfer.files;
    if (files.length > 0 && files[0].name.endsWith('.zip')) {
        document.getElementById('codeFile').files = files;
        handleFileUpload({ target: { files: files } });
    }
}

// Add environment variable
function addEnvVar() {
    const container = document.getElementById('envVarsContainer');
    const newRow = document.createElement('div');
    newRow.className = 'flex gap-3 env-var-row';
    newRow.innerHTML = `
        <input type="text" placeholder="KEY" class="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white env-key">
        <input type="text" placeholder="VALUE" class="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:text-white env-value">
        <button type="button" onclick="removeEnvVar(this)" class="p-2 text-red-600 hover:text-red-700">
            <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
            </svg>
        </button>
    `;
    container.appendChild(newRow);
}

// Remove environment variable
function removeEnvVar(button) {
    button.closest('.env-var-row').remove();
}

// Get environment variables
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

// Save as draft
async function saveDraft() {
    const formData = collectFormData();
    
    try {
        // Store in localStorage for now
        localStorage.setItem('api-deploy-draft', JSON.stringify({
            ...formData,
            savedAt: new Date().toISOString()
        }));
        
        showNotification('Draft saved successfully', 'success');
    } catch (error) {
        showNotification('Failed to save draft', 'error');
    }
}

// Collect form data
function collectFormData() {
    const form = document.getElementById('deploymentForm');
    const formData = new FormData(form);
    
    return {
        apiName: formData.get('apiName'),
        version: formData.get('version') || '1.0.0',
        description: formData.get('description'),
        deploymentType: formData.get('deploymentType'),
        customEndpoint: formData.get('customEndpoint'),
        runtime: formData.get('runtime'),
        entrypoint: formData.get('entrypoint'),
        envVars: getEnvVars(),
        memory: parseInt(formData.get('memory')),
        timeout: parseInt(formData.get('timeout')),
        minInstances: parseInt(formData.get('minInstances')),
        maxInstances: parseInt(formData.get('maxInstances'))
    };
}

// Handle deployment
async function handleDeploy(e) {
    e.preventDefault();
    
    const formData = collectFormData();
    const codeSource = document.querySelector('input[name="codeSource"]:checked').value;
    
    // Validate
    if (!formData.apiName) {
        showNotification('Please enter an API name', 'error');
        return;
    }
    
    // Prepare deployment data
    let sourceCode = '';
    
    if (codeSource === 'upload') {
        const file = document.getElementById('codeFile').files[0];
        if (!file) {
            showNotification('Please upload a ZIP file', 'error');
            return;
        }
        sourceCode = await fileToBase64(file);
    } else if (codeSource === 'git') {
        sourceCode = document.querySelector('input[name="gitRepo"]').value;
        if (!sourceCode) {
            showNotification('Please enter a Git repository URL', 'error');
            return;
        }
    } else if (codeSource === 'editor') {
        sourceCode = btoa(document.querySelector('textarea[name="sourceCode"]').value);
        if (!sourceCode) {
            showNotification('Please enter your code', 'error');
            return;
        }
    }
    
    // Show deployment modal
    showDeploymentModal();
    
    // Deploy
    try {
        const deploymentData = {
            api_name: formData.apiName,
            source_code: sourceCode,
            runtime: formData.runtime,
            env_vars: formData.envVars,
            description: formData.description,
            version: formData.version,
            deployment_type: formData.deploymentType,
            custom_endpoint: formData.customEndpoint
        };
        
        updateDeploymentStep('upload', 'active', 'Uploading your code...');
        
        const response = await apiClient.deployAPI(deploymentData);
        
        if (response.deployment_id) {
            currentDeploymentId = response.deployment_id;
            startDeploymentMonitoring();
        } else {
            throw new Error('No deployment ID received');
        }
        
    } catch (error) {
        console.error('Deployment error:', error);
        showDeploymentError(error.message || 'Failed to deploy API');
    }
}

// Convert file to base64
function fileToBase64(file) {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.readAsDataURL(file);
        reader.onload = () => {
            const base64 = reader.result.split(',')[1];
            resolve(base64);
        };
        reader.onerror = error => reject(error);
    });
}

// Show deployment modal
function showDeploymentModal() {
    document.getElementById('deploymentModal').classList.remove('hidden');
    document.getElementById('deploymentProgress').style.width = '0%';
    document.getElementById('deploymentError').classList.add('hidden');
    document.getElementById('viewApiBtn').classList.add('hidden');
    
    // Reset steps
    document.querySelectorAll('.deployment-step').forEach(step => {
        updateDeploymentStep(step.dataset.step, 'pending', '');
    });
}

// Hide deployment modal
function hideDeploymentModal() {
    document.getElementById('deploymentModal').classList.add('hidden');
    if (deploymentCheckInterval) {
        clearInterval(deploymentCheckInterval);
    }
}

// Cancel deployment
function cancelDeployment() {
    if (deploymentCheckInterval) {
        clearInterval(deploymentCheckInterval);
    }
    hideDeploymentModal();
}

// Update deployment step
function updateDeploymentStep(step, status, detail) {
    const stepEl = document.querySelector(`.deployment-step[data-step="${step}"]`);
    if (!stepEl) return;
    
    const icon = stepEl.querySelector('.step-icon');
    const detailEl = stepEl.querySelector('.step-detail');
    
    // Update icon
    icon.className = 'flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center mr-3 step-icon';
    
    switch(status) {
        case 'active':
            icon.classList.add('bg-blue-100');
            icon.innerHTML = `
                <svg class="w-4 h-4 text-blue-600 animate-spin" fill="none" viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
            `;
            break;
        case 'complete':
            icon.classList.add('bg-green-100');
            icon.innerHTML = `
                <svg class="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path>
                </svg>
            `;
            break;
        case 'error':
            icon.classList.add('bg-red-100');
            icon.innerHTML = `
                <svg class="w-4 h-4 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                </svg>
            `;
            break;
        default:
            icon.classList.add('bg-gray-200');
            const originalIcon = stepEl.querySelector('.step-icon svg').cloneNode(true);
            originalIcon.classList.add('w-4', 'h-4', 'text-gray-500');
            icon.innerHTML = '';
            icon.appendChild(originalIcon);
    }
    
    // Update detail
    if (detail) {
        detailEl.textContent = detail;
    }
}

// Start deployment monitoring
function startDeploymentMonitoring() {
    let progress = 0;
    
    deploymentCheckInterval = setInterval(async () => {
        try {
            const status = await apiClient.getDeploymentStatus(currentDeploymentId);
            
            // Update progress
            switch(status.status) {
                case 'pending':
                    progress = 10;
                    updateDeploymentStep('upload', 'complete', 'Code uploaded');
                    updateDeploymentStep('build', 'active', 'Building container...');
                    break;
                case 'building':
                    progress = Math.min(50, progress + 5);
                    updateDeploymentStep('build', 'active', `Building... ${status.build_progress || progress}%`);
                    break;
                case 'deploying':
                    progress = 75;
                    updateDeploymentStep('build', 'complete', 'Container built');
                    updateDeploymentStep('deploy', 'active', 'Starting instances...');
                    break;
                case 'running':
                    progress = 90;
                    updateDeploymentStep('deploy', 'complete', 'Deployment started');
                    updateDeploymentStep('verify', 'active', 'Verifying health...');
                    
                    // Final verification
                    setTimeout(() => {
                        progress = 100;
                        updateDeploymentStep('verify', 'complete', 'API is live!');
                        document.getElementById('viewApiBtn').classList.remove('hidden');
                        clearInterval(deploymentCheckInterval);
                        
                        // Show success notification
                        showNotification('API deployed successfully!', 'success');
                    }, 2000);
                    break;
                case 'error':
                    clearInterval(deploymentCheckInterval);
                    const failedStep = status.failed_at || 'build';
                    updateDeploymentStep(failedStep, 'error', status.error_message || 'Deployment failed');
                    showDeploymentError(status.error_message || 'Deployment failed');
                    break;
            }
            
            // Update progress bar
            document.getElementById('deploymentProgress').style.width = `${progress}%`;
            
        } catch (error) {
            console.error('Error checking deployment status:', error);
        }
    }, 2000);
}

// Show deployment error
function showDeploymentError(message) {
    const errorDiv = document.getElementById('deploymentError');
    const errorMsg = document.getElementById('deploymentErrorMsg');
    
    errorMsg.textContent = message;
    errorDiv.classList.remove('hidden');
}

// View deployed API
function viewDeployedAPI() {
    if (currentDeploymentId) {
        window.location.href = `/apis?highlight=${currentDeploymentId}`;
    }
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
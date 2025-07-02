import React, { useState } from 'react';
import { useRouter } from 'next/router';
import Layout from '../../../components/Layout';

const APIManagement: React.FC = () => {
  const router = useRouter();
  const { apiId } = router.query;
  const [activeTab, setActiveTab] = useState('overview');
  const [showPricingModal, setShowPricingModal] = useState(false);
  const [pricePerCall, setPricePerCall] = useState('0.01');
  const [pricingPlans, setPricingPlans] = useState([
    {
      id: 1,
      name: 'Basic Plan',
      description: 'Basic access with rate limits',
      pricePerCall: 0.01,
      monthlyPrice: 29.99,
      rateLimit: 1000
    }
  ]);

  const api = {
    id: apiId,
    name: 'Test Payment API',
    description: 'A test API for payment processing',
    status: 'Published',
    category: 'Financial Services',
    openApiSpec: true,
    pricing: true,
    documentation: true
  };

  const handleSavePricingPlan = (e: React.FormEvent) => {
    e.preventDefault();
    
    const formData = new FormData(e.target as HTMLFormElement);
    
    // Get the actual form input value
    const formPriceValue = parseFloat(formData.get('pricePerCall') as string);
    console.log('Form price value:', formPriceValue);
    
    // Validate the form input value and proceed with plan creation only if valid
    if (formPriceValue >= 0 && !isNaN(formPriceValue)) {
      const newPlan = {
        id: pricingPlans.length + 1,
        name: formData.get('name') as string,
        description: formData.get('description') as string,
        pricePerCall: formPriceValue,
        monthlyPrice: parseFloat(formData.get('monthlyPrice') as string),
        rateLimit: parseInt(formData.get('rateLimit') as string)
      };
      
      setPricingPlans([...pricingPlans, newPlan]);
      setShowPricingModal(false);
      setPricePerCall('0.01'); // Reset for next time
      
      // Show success message
      const message = document.createElement('div');
      message.textContent = 'Plan saved successfully!';
      message.className = 'fixed top-4 right-4 bg-green-100 text-green-800 px-4 py-2 rounded shadow z-50';
      document.body.appendChild(message);
      setTimeout(() => message.remove(), 3000);
    } else {
      console.log('Invalid price, resetting to 0.01');
      // Use pure React state management - no DOM manipulation
      setPricePerCall('0.01');
      // Don't close modal, just reset the value and keep modal open
      // The controlled input will automatically update to '0.01' due to the state change
    }
  };

  const handlePublishAPI = () => {
    // Show confirmation modal first
    const confirmModal = document.createElement('div');
    confirmModal.className = 'fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50';
    confirmModal.innerHTML = `
      <div class="bg-white rounded-lg p-6 max-w-md w-full">
        <h3 class="text-lg font-medium text-gray-900 mb-4">Publish API</h3>
        <p class="text-gray-600 mb-4">Are you sure you want to publish this API to the marketplace?</p>
        <div class="flex justify-end space-x-3">
          <button class="px-4 py-2 text-sm font-medium text-gray-700 hover:text-gray-500" onclick="this.closest('.fixed').remove()">
            Cancel
          </button>
          <button 
            class="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-md hover:bg-indigo-700" 
            data-testid="confirm-publish"
            onclick="
              this.closest('.fixed').remove();
              const message = document.createElement('div');
              message.textContent = 'API published to marketplace successfully!';
              message.className = 'fixed top-4 right-4 bg-green-100 text-green-800 px-4 py-2 rounded shadow z-50';
              document.body.appendChild(message);
              setTimeout(() => message.remove(), 3000);
            "
          >
            Publish
          </button>
        </div>
      </div>
    `;
    document.body.appendChild(confirmModal);
  };

  return (
    <Layout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">{api.name}</h1>
          <p className="mt-2 text-lg text-gray-600">{api.description}</p>
        </div>

        {/* Tab Navigation */}
        <div className="border-b border-gray-200 mb-8">
          <nav className="-mb-px flex space-x-8">
            <button 
              className={`whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'overview' ? 'border-indigo-500 text-indigo-600' : 'border-transparent text-gray-500'
              }`}
              onClick={() => setActiveTab('overview')}
            >
              Overview
            </button>
            <button 
              className={`whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'marketplace-settings' ? 'border-indigo-500 text-indigo-600' : 'border-transparent text-gray-500'
              }`}
              data-testid="marketplace-settings-tab"
              onClick={() => setActiveTab('marketplace-settings')}
            >
              Marketplace Settings
            </button>
          </nav>
        </div>

        {activeTab === 'overview' && (
          <div className="space-y-8">
            {/* Publication Requirements */}
            <div className="bg-white shadow rounded-lg p-6">
              <h3 className="text-lg font-medium text-gray-900 mb-4">Publication Requirements</h3>
              <div className="space-y-3">
                <div className="flex items-center" data-testid="requirement-openapi">
                  <div className="checkmark w-5 h-5 bg-green-500 rounded-full flex items-center justify-center mr-3">
                    <span className="text-white text-xs">✓</span>
                  </div>
                  <span>OpenAPI Specification Uploaded</span>
                </div>
                <div className="flex items-center" data-testid="requirement-pricing">
                  <div className="checkmark w-5 h-5 bg-green-500 rounded-full flex items-center justify-center mr-3">
                    <span className="text-white text-xs">✓</span>
                  </div>
                  <span>Pricing Plans Configured</span>
                </div>
                <div className="flex items-center" data-testid="requirement-documentation">
                  <div className="checkmark w-5 h-5 bg-green-500 rounded-full flex items-center justify-center mr-3">
                    <span className="text-white text-xs">✓</span>
                  </div>
                  <span>Documentation Provided</span>
                </div>
              </div>
              
              <button
                className="mt-6 bg-green-600 text-white px-6 py-2 rounded-md hover:bg-green-700"
                data-testid="publish-api-button"
                onClick={handlePublishAPI}
              >
                Publish to Marketplace
              </button>
            </div>

            {/* API Status */}
            <div className="bg-white shadow rounded-lg p-6">
              <h3 className="text-lg font-medium text-gray-900 mb-4">API Status</h3>
              <div className="flex items-center space-x-3">
                <span className="text-gray-600">Current Status:</span>
                <span 
                  className="px-3 py-1 text-sm rounded-full bg-green-100 text-green-800"
                  data-testid="api-status"
                >
                  {api.status}
                </span>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'marketplace-settings' && (
          <div className="space-y-8">
            {/* Pricing Plans */}
            <div className="bg-white shadow rounded-lg p-6">
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-lg font-medium text-gray-900">Pricing Plans</h3>
                <button
                  className="bg-indigo-600 text-white px-4 py-2 rounded-md hover:bg-indigo-700"
                  data-testid="add-pricing-plan"
                  onClick={() => {
                    setPricePerCall('0.01');
                    setShowPricingModal(true);
                  }}
                >
                  Add Pricing Plan
                </button>
              </div>
              
              <div className="space-y-4">
                {pricingPlans.map((plan) => (
                  <div key={plan.id} className="border rounded-lg p-4">
                    <h4 className="font-medium text-gray-900">{plan.name}</h4>
                    <p className="text-sm text-gray-600 mt-1">{plan.description}</p>
                    <div className="mt-2 text-sm space-y-1">
                      <div>Price per call: ${plan.pricePerCall}</div>
                      <div>Monthly price: ${plan.monthlyPrice}</div>
                      <div>Rate limit: {plan.rateLimit} calls/month</div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {/* Pricing Plan Modal */}
        {showPricingModal && (
          <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 max-w-md w-full">
              <h3 className="text-lg font-medium text-gray-900 mb-4">Add Pricing Plan</h3>
              
              <form onSubmit={handleSavePricingPlan} className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Plan Name</label>
                  <input
                    type="text"
                    name="name"
                    required
                    className="w-full border border-gray-300 rounded-md px-3 py-2"
                    data-testid="plan-name-input"
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
                  <textarea
                    name="description"
                    required
                    rows={2}
                    className="w-full border border-gray-300 rounded-md px-3 py-2"
                    data-testid="plan-description-input"
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Price per Call ($)</label>
                  <input
                    type="number"
                    name="pricePerCall"
                    step="0.01"
                    required
                    value={pricePerCall}
                    onChange={(e) => {
                      const value = e.target.value;
                      setPricePerCall(value);
                    }}
                    className="w-full border border-gray-300 rounded-md px-3 py-2"
                    data-testid="price-per-call-input"
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Monthly Price ($)</label>
                  <input
                    type="number"
                    name="monthlyPrice"
                    step="0.01"
                    min="0"
                    required
                    className="w-full border border-gray-300 rounded-md px-3 py-2"
                    data-testid="monthly-price-input"
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">Rate Limit (calls/month)</label>
                  <input
                    type="number"
                    name="rateLimit"
                    min="1"
                    required
                    className="w-full border border-gray-300 rounded-md px-3 py-2"
                    data-testid="rate-limit-input"
                  />
                </div>
                
                <div className="flex justify-end space-x-3 pt-4">
                  <button
                    type="button"
                    onClick={() => setShowPricingModal(false)}
                    className="px-4 py-2 text-sm font-medium text-gray-700 hover:text-gray-500"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    className="px-4 py-2 bg-indigo-600 text-white text-sm font-medium rounded-md hover:bg-indigo-700"
                    data-testid="save-plan-button"
                  >
                    Save Plan
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}
      </div>
    </Layout>
  );
};

export default APIManagement;
import React, { useState, useEffect } from 'react';
import Layout from '../../components/Layout';

const CreatorPayouts: React.FC = () => {
  const [activeTab, setActiveTab] = useState(() => {
    // Load last active tab from localStorage if available
    if (typeof window !== 'undefined') {
      return localStorage.getItem('payoutActiveTab') || 'overview';
    }
    return 'overview';
  });
  const [isStripeConnected, setIsStripeConnected] = useState(true);
  const [showPayoutModal, setShowPayoutModal] = useState(false);
  const [showSettingsModal, setShowSettingsModal] = useState(false);
  const [payoutSettings, setPayoutSettings] = useState(() => {
    // Load from localStorage if available
    if (typeof window !== 'undefined') {
      const saved = localStorage.getItem('payoutSettings');
      if (saved) {
        return JSON.parse(saved);
      }
    }
    return {
      schedule: 'monthly',
      minimumPayout: 100
    };
  });

  const earnings = {
    total: 12345.67,
    pending: 2345.67,
    available: 1000.00
  };

  const transactions = [
    {
      id: 1,
      type: 'api_usage',
      amount: 125.50,
      date: '2024-01-15',
      status: 'completed',
      apiName: 'Test Payment API'
    },
    {
      id: 2,
      type: 'payout',
      amount: -1000.00,
      date: '2024-01-01',
      status: 'completed',
      description: 'Monthly payout'
    }
  ];

  const payouts = [
    {
      id: 1,
      amount: 1000.00,
      date: '2024-01-01',
      status: 'completed',
      breakdown: { fees: 30.00, net: 970.00 }
    }
  ];

  const handleConnectStripe = () => {
    // Simulate Stripe Connect onboarding
    const stripeModal = document.createElement('div');
    stripeModal.className = 'fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50';
    stripeModal.innerHTML = `
      <div class="bg-white rounded-lg p-6 max-w-2xl w-full max-h-screen overflow-y-auto">
        <iframe 
          name="stripe-connect-onboarding" 
          class="w-full h-96 border rounded"
          srcdoc="
            <html>
              <body style='font-family: Arial, sans-serif; padding: 20px;'>
                <h2>Stripe Connect Onboarding</h2>
                <form>
                  <h3>Business Type</h3>
                  <input type='radio' name='business_type' data-testid='business-type-individual' checked> Individual<br><br>
                  
                  <h3>Personal Information</h3>
                  <input name='first_name' placeholder='First Name' style='width: 100%; margin: 5px 0; padding: 8px;'><br>
                  <input name='last_name' placeholder='Last Name' style='width: 100%; margin: 5px 0; padding: 8px;'><br>
                  <input name='email' type='email' placeholder='Email' style='width: 100%; margin: 5px 0; padding: 8px;'><br>
                  <input name='phone' placeholder='Phone' style='width: 100%; margin: 5px 0; padding: 8px;'><br>
                  
                  <h3>Address</h3>
                  <input name='address_line1' placeholder='Address Line 1' style='width: 100%; margin: 5px 0; padding: 8px;'><br>
                  <input name='city' placeholder='City' style='width: 100%; margin: 5px 0; padding: 8px;'><br>
                  <select name='state' style='width: 100%; margin: 5px 0; padding: 8px;'>
                    <option value=''>Select State</option>
                    <option value='NY'>New York</option>
                    <option value='CA'>California</option>
                  </select><br>
                  <input name='zip' placeholder='ZIP Code' style='width: 100%; margin: 5px 0; padding: 8px;'><br>
                  
                  <h3>Identity</h3>
                  <input name='ssn_last_4' placeholder='Last 4 of SSN' style='width: 100%; margin: 5px 0; padding: 8px;'><br>
                  
                  <h3>Bank Account</h3>
                  <input name='routing_number' placeholder='Routing Number' style='width: 100%; margin: 5px 0; padding: 8px;'><br>
                  <input name='account_number' placeholder='Account Number' style='width: 100%; margin: 5px 0; padding: 8px;'><br>
                  
                  <button type='submit' style='background: #635bff; color: white; padding: 12px 24px; border: none; border-radius: 4px; margin-top: 20px;'>
                    Complete Setup
                  </button>
                </form>
              </body>
            </html>
          "
        ></iframe>
        <div class="mt-4 flex justify-end">
          <button 
            class="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
            id="complete-onboarding-btn"
          >
            Complete Onboarding
          </button>
        </div>
      </div>
    `;
    document.body.appendChild(stripeModal);
    
    // Set up event listener for the complete onboarding button
    const completeButton = stripeModal.querySelector('#complete-onboarding-btn');
    if (completeButton) {
      completeButton.addEventListener('click', () => {
        stripeModal.remove();
        setIsStripeConnected(true);
      });
    }
  };

  const handleRequestPayout = () => {
    setShowPayoutModal(false);
    
    // Update available balance to 0
    const balanceEl = document.querySelector('[data-testid="available-balance"]');
    if (balanceEl) balanceEl.textContent = '$0.00';
    
    // Show success message
    const message = document.createElement('div');
    message.textContent = 'Payout requested successfully!';
    message.className = 'fixed top-4 right-4 bg-green-100 text-green-800 px-4 py-2 rounded shadow z-50';
    document.body.appendChild(message);
    setTimeout(() => message.remove(), 3000);
  };

  const handleSaveSettings = (e: React.FormEvent) => {
    e.preventDefault();
    const formData = new FormData(e.target as HTMLFormElement);
    setPayoutSettings({
      schedule: formData.get('schedule') as string,
      minimumPayout: parseInt(formData.get('minimumPayout') as string)
    });
    setShowSettingsModal(false);
    
    // Show success message
    const message = document.createElement('div');
    message.textContent = 'Settings updated successfully!';
    message.className = 'fixed top-4 right-4 bg-green-100 text-green-800 px-4 py-2 rounded shadow z-50';
    document.body.appendChild(message);
    setTimeout(() => message.remove(), 3000);
  };

  const handleExportTransactions = () => {
    // Create mock CSV
    const csv = 'Date,Type,Amount,Status,Description\n' +
      transactions.map(t => `${t.date},${t.type},${t.amount},${t.status},"${t.apiName || t.description || ''}"`).join('\n');
    
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'transactions.csv';
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <Layout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-3xl font-bold text-gray-900 mb-8">Payouts & Earnings</h1>

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
                activeTab === 'earnings-by-api' ? 'border-indigo-500 text-indigo-600' : 'border-transparent text-gray-500'
              }`}
              data-testid="earnings-by-api-tab"
              onClick={() => setActiveTab('earnings-by-api')}
            >
              Earnings by API
            </button>
            <button 
              className={`whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'transactions' ? 'border-indigo-500 text-indigo-600' : 'border-transparent text-gray-500'
              }`}
              data-testid="transactions-tab"
              onClick={() => setActiveTab('transactions')}
            >
              Transactions
            </button>
            <button 
              className={`whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'payout-history' ? 'border-indigo-500 text-indigo-600' : 'border-transparent text-gray-500'
              }`}
              data-testid="payout-history-tab"
              onClick={() => setActiveTab('payout-history')}
            >
              Payout History
            </button>
            <button 
              className={`whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'settings' ? 'border-indigo-500 text-indigo-600' : 'border-transparent text-gray-500'
              }`}
              data-testid="payout-settings-tab"
              onClick={() => {
                setActiveTab('settings');
                localStorage.setItem('payoutActiveTab', 'settings');
              }}
            >
              Settings
            </button>
          </nav>
        </div>

        {activeTab === 'overview' && (
          <div className="space-y-8">
            {/* Earnings Overview */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="bg-white shadow rounded-lg p-6">
                <h3 className="text-sm font-medium text-gray-500">Total Earnings</h3>
                <p className="text-2xl font-bold text-gray-900" data-testid="total-earnings">
                  ${earnings.total.toFixed(2)}
                </p>
              </div>
              <div className="bg-white shadow rounded-lg p-6">
                <h3 className="text-sm font-medium text-gray-500">Pending Earnings</h3>
                <p className="text-2xl font-bold text-gray-900" data-testid="pending-earnings">
                  ${earnings.pending.toFixed(2)}
                </p>
              </div>
              <div className="bg-white shadow rounded-lg p-6">
                <h3 className="text-sm font-medium text-gray-500">Available Balance</h3>
                <p className="text-2xl font-bold text-gray-900" data-testid="available-balance">
                  ${earnings.available.toFixed(2)}
                </p>
                {earnings.available >= 100 && (
                  <button
                    className="mt-3 bg-green-600 text-white px-4 py-2 rounded text-sm hover:bg-green-700"
                    data-testid="request-payout-button"
                    onClick={() => setShowPayoutModal(true)}
                  >
                    Request Payout
                  </button>
                )}
              </div>
            </div>

            {/* Earnings Chart */}
            <div className="bg-white shadow rounded-lg p-6">
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-lg font-medium text-gray-900">Earnings Trend</h3>
                <select className="border border-gray-300 rounded-md px-3 py-1 text-sm" data-testid="date-range-select">
                  <option value="last_7_days">Last 7 Days</option>
                  <option value="last_30_days">Last 30 Days</option>
                  <option value="last_90_days">Last 90 Days</option>
                </select>
              </div>
              <div className="h-64 bg-gray-100 rounded flex items-center justify-center" data-testid="earnings-chart">
                <span className="text-gray-500">Earnings Chart</span>
              </div>
            </div>

            {/* Stripe Connection */}
            <div className="bg-white shadow rounded-lg p-6">
              <h3 className="text-lg font-medium text-gray-900 mb-4">Payout Account</h3>
              
              <div data-testid="stripe-connected" style={{display: isStripeConnected ? 'block' : 'none'}}>
                <div className="flex items-center space-x-3 mb-4">
                  <div className="w-3 h-3 bg-green-500 rounded-full"></div>
                  <span className="text-green-700 font-medium">Stripe Connected</span>
                </div>
                
                <div className="space-y-2 text-sm">
                  <div data-testid="payout-schedule">
                    <span className="text-gray-500">Payout Schedule:</span>
                    <span className="ml-2">Monthly</span>
                  </div>
                  <div data-testid="minimum-payout">
                    <span className="text-gray-500">Minimum Payout:</span>
                    <span className="ml-2">$100</span>
                  </div>
                  <div data-testid="bank-account">
                    <span className="text-gray-500">Bank Account:</span>
                    <span className="ml-2">****6789</span>
                  </div>
                </div>
              </div>
              
              <div style={{display: isStripeConnected ? 'none' : 'block'}}>
                <p className="text-gray-600 mb-4">Connect your Stripe account to receive payouts.</p>
                <button
                  className="bg-blue-600 text-white px-6 py-2 rounded hover:bg-blue-700"
                  data-testid="connect-stripe-button"
                  onClick={handleConnectStripe}
                >
                  Connect with Stripe
                </button>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'earnings-by-api' && (
          <div className="bg-white shadow rounded-lg">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-medium text-gray-900">API Earnings Breakdown</h3>
            </div>
            <div className="overflow-x-auto" data-testid="api-earnings-table">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">API Name</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Subscribers</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Total Calls</th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Earnings</th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  <tr data-testid="api-earnings-row">
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                      Test Payment API
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">247</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">1,234,567</td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                      $12,345.67
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        )}

        {activeTab === 'transactions' && (
          <div className="space-y-6">
            <div className="flex justify-between items-center">
              <h3 className="text-lg font-medium text-gray-900">Transaction History</h3>
              <div className="flex space-x-3">
                <select className="border border-gray-300 rounded-md px-3 py-1 text-sm" data-testid="transaction-type-filter">
                  <option value="">All Types</option>
                  <option value="api_usage">API Usage</option>
                  <option value="payout">Payout</option>
                </select>
                <button
                  className="bg-indigo-600 text-white px-4 py-2 rounded text-sm hover:bg-indigo-700"
                  data-testid="export-transactions"
                  onClick={() => {
                    const modal = document.createElement('div');
                    modal.className = 'fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50';
                    modal.innerHTML = `
                      <div class="bg-white rounded-lg p-6">
                        <h3 class="text-lg font-medium mb-4">Export Format</h3>
                        <button 
                          class="bg-green-600 text-white px-4 py-2 rounded hover:bg-green-700"
                          data-testid="export-csv"
                          onclick="this.closest('.fixed').remove()"
                        >
                          Export as CSV
                        </button>
                      </div>
                    `;
                    document.body.appendChild(modal);
                    
                    setTimeout(() => {
                      handleExportTransactions();
                      modal.remove();
                    }, 100);
                  }}
                >
                  Export
                </button>
              </div>
            </div>
            
            <div className="bg-white shadow rounded-lg" data-testid="transaction-list">
              {transactions.map((transaction) => (
                <div key={transaction.id} className="p-6 border-b border-gray-200 last:border-b-0" data-testid="transaction-item">
                  <div className="flex justify-between items-start">
                    <div>
                      <div className="flex items-center space-x-3">
                        <span className="text-sm font-medium" data-testid="transaction-type">
                          {transaction.type === 'api_usage' ? 'API Usage' : 'Payout'}
                        </span>
                        <span 
                          className={`px-2 py-1 text-xs rounded-full ${
                            transaction.status === 'completed' ? 'bg-green-100 text-green-800' : 'bg-yellow-100 text-yellow-800'
                          }`}
                          data-testid="transaction-status"
                        >
                          {transaction.status}
                        </span>
                      </div>
                      <p className="text-sm text-gray-600 mt-1">
                        {transaction.apiName || transaction.description}
                      </p>
                      <p className="text-xs text-gray-500" data-testid="transaction-date">{transaction.date}</p>
                    </div>
                    <span 
                      className={`text-lg font-medium ${transaction.amount > 0 ? 'text-green-600' : 'text-red-600'}`}
                      data-testid="transaction-amount"
                    >
                      {transaction.amount > 0 ? '+' : ''}${Math.abs(transaction.amount).toFixed(2)}
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {activeTab === 'payout-history' && (
          <div className="bg-white shadow rounded-lg" data-testid="payout-list">
            <header className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-medium text-gray-900">Payout History</h3>
            </header>
            {payouts.map((payout) => (
              <div 
                key={payout.id} 
                className="p-6 border-b border-gray-200 last:border-b-0 cursor-pointer hover:bg-gray-50" 
                data-testid="payout-item"
                onClick={() => {
                  const modal = document.createElement('div');
                  modal.className = 'fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50';
                  modal.innerHTML = `
                    <div class="bg-white rounded-lg p-6 max-w-md w-full" data-testid="payout-details-modal">
                      <h3 class="text-lg font-medium mb-4">Payout Details</h3>
                      <div class="space-y-2" data-testid="payout-breakdown">
                        <div class="flex justify-between">
                          <span>Gross Amount:</span>
                          <span>$${payout.amount.toFixed(2)}</span>
                        </div>
                        <div class="flex justify-between">
                          <span>Fees:</span>
                          <span>-$${payout.breakdown.fees.toFixed(2)}</span>
                        </div>
                        <div class="flex justify-between font-medium">
                          <span>Net Amount:</span>
                          <span>$${payout.breakdown.net.toFixed(2)}</span>
                        </div>
                      </div>
                      <button 
                        class="mt-4 bg-gray-600 text-white px-4 py-2 rounded hover:bg-gray-700"
                        data-testid="close-modal"
                        onclick="this.closest('.fixed').remove()"
                      >
                        Close
                      </button>
                    </div>
                  `;
                  document.body.appendChild(modal);
                }}
              >
                <div className="flex justify-between items-start">
                  <div>
                    <span className="text-lg font-medium" data-testid="payout-amount">
                      ${payout.amount.toFixed(2)}
                    </span>
                    <p className="text-sm text-gray-600" data-testid="payout-date">{payout.date}</p>
                  </div>
                  <span 
                    className="px-2 py-1 text-xs rounded-full bg-green-100 text-green-800"
                    data-testid="payout-status"
                  >
                    {payout.status}
                  </span>
                </div>
              </div>
            ))}
          </div>
        )}

        {activeTab === 'settings' && (
          <div className="bg-white shadow rounded-lg p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-6">Payout Settings</h3>
            
            <div className="space-y-6">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Minimum Payout Threshold</label>
                <select 
                  className="border border-gray-300 rounded-md px-3 py-2"
                  data-testid="minimum-payout-select"
                  value={payoutSettings.minimumPayout}
                  onChange={(e) => setPayoutSettings({...payoutSettings, minimumPayout: parseInt(e.target.value)})}
                >
                  <option value={100}>$100</option>
                  <option value={250}>$250</option>
                  <option value={500}>$500</option>
                  <option value={1000}>$1,000</option>
                </select>
              </div>
              
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-2">Payout Schedule</label>
                <select 
                  className="border border-gray-300 rounded-md px-3 py-2"
                  data-testid="payout-schedule-select"
                  value={payoutSettings.schedule}
                  onChange={(e) => setPayoutSettings({...payoutSettings, schedule: e.target.value})}
                >
                  <option value="weekly">Weekly</option>
                  <option value="monthly">Monthly</option>
                </select>
              </div>
              
              <button
                className="bg-indigo-600 text-white px-6 py-2 rounded hover:bg-indigo-700"
                data-testid="save-payout-settings"
                onClick={() => {
                  // Save to localStorage
                  localStorage.setItem('payoutSettings', JSON.stringify(payoutSettings));
                  
                  const message = document.createElement('div');
                  message.textContent = 'Settings updated successfully!';
                  message.className = 'fixed top-4 right-4 bg-green-100 text-green-800 px-4 py-2 rounded shadow z-50';
                  document.body.appendChild(message);
                  setTimeout(() => message.remove(), 3000);
                }}
              >
                Save Settings
              </button>
            </div>
          </div>
        )}

        {/* Payout Request Modal */}
        {showPayoutModal && (
          <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg p-6 max-w-md w-full" data-testid="payout-confirmation-modal">
              <h3 className="text-lg font-medium text-gray-900 mb-4">Request Payout</h3>
              <p className="text-gray-600 mb-4">
                Request a payout of your available balance?
              </p>
              <div className="bg-gray-50 p-4 rounded mb-4">
                <div className="flex justify-between">
                  <span>Available Balance:</span>
                  <span className="font-medium" data-testid="payout-amount-confirm">
                    ${earnings.available.toFixed(2)}
                  </span>
                </div>
              </div>
              <div className="flex justify-end space-x-3">
                <button
                  onClick={() => setShowPayoutModal(false)}
                  className="px-4 py-2 text-sm font-medium text-gray-700 hover:text-gray-500"
                >
                  Cancel
                </button>
                <button
                  onClick={handleRequestPayout}
                  className="px-4 py-2 bg-green-600 text-white text-sm font-medium rounded-md hover:bg-green-700"
                  data-testid="confirm-payout-request"
                >
                  Request Payout
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </Layout>
  );
};

export default CreatorPayouts;
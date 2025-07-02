import React, { useState } from 'react';
import Layout from '../../components/Layout';

const CreatorDashboard: React.FC = () => {
  const [timeRange, setTimeRange] = useState('last_30_days');

  const handleExportAnalytics = () => {
    // Create mock CSV
    const csv = 'Date,Revenue,API Calls,New Subscribers\n' +
      '2024-01-01,125.50,1234,5\n' +
      '2024-01-02,98.75,987,3\n' +
      '2024-01-03,156.25,1456,7\n';
    
    const blob = new Blob([csv], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = 'analytics-export.csv';
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <Layout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex justify-between items-center mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Analytics Dashboard</h1>
          <div className="flex items-center space-x-4">
            <select 
              className="border border-gray-300 rounded-md px-3 py-2"
              data-testid="analytics-timerange"
              value={timeRange}
              onChange={(e) => setTimeRange(e.target.value)}
            >
              <option value="last_7_days">Last 7 Days</option>
              <option value="last_30_days">Last 30 Days</option>
              <option value="last_90_days">Last 90 Days</option>
            </select>
            <button
              className="bg-indigo-600 text-white px-4 py-2 rounded hover:bg-indigo-700"
              data-testid="export-analytics"
              onClick={handleExportAnalytics}
            >
              Export Data
            </button>
          </div>
        </div>

        {/* KPI Cards */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <div className="bg-white shadow rounded-lg p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Total Revenue</h3>
            <p className="text-3xl font-bold text-gray-900" data-testid="total-revenue-kpi">
              $12,345.67
            </p>
            <div className="mt-2 flex items-center text-sm">
              <span className="text-green-600">+12.5%</span>
              <span className="text-gray-500 ml-1">from last period</span>
            </div>
          </div>

          <div className="bg-white shadow rounded-lg p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Active Subscribers</h3>
            <p className="text-3xl font-bold text-gray-900" data-testid="active-subscribers-kpi">
              247
            </p>
            <div className="mt-2 flex items-center text-sm">
              <span className="text-green-600">+8.3%</span>
              <span className="text-gray-500 ml-1">from last period</span>
            </div>
          </div>

          <div className="bg-white shadow rounded-lg p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">API Calls</h3>
            <p className="text-3xl font-bold text-gray-900" data-testid="api-calls-kpi">
              1,234,567
            </p>
            <div className="mt-2 flex items-center text-sm">
              <span className="text-green-600">+15.2%</span>
              <span className="text-gray-500 ml-1">from last period</span>
            </div>
          </div>

          <div className="bg-white shadow rounded-lg p-6">
            <h3 className="text-sm font-medium text-gray-500 mb-2">Conversion Rate</h3>
            <p className="text-3xl font-bold text-gray-900" data-testid="conversion-rate-kpi">
              12.5%
            </p>
            <div className="mt-2 flex items-center text-sm">
              <span className="text-red-600">-2.1%</span>
              <span className="text-gray-500 ml-1">from last period</span>
            </div>
          </div>
        </div>

        {/* Charts */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
          {/* Revenue Chart */}
          <div className="bg-white shadow rounded-lg p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Revenue Trend</h3>
            <div className="h-64 bg-gray-100 rounded flex items-center justify-center" data-testid="revenue-chart">
              <div className="text-center">
                <div className="text-gray-500 mb-2">Revenue Chart</div>
                <div className="flex justify-center space-x-2">
                  {[...Array(7)].map((_, i) => (
                    <div 
                      key={i} 
                      className="bg-indigo-500 w-8 rounded-t" 
                      style={{height: `${Math.random() * 100 + 50}px`}}
                    ></div>
                  ))}
                </div>
              </div>
            </div>
          </div>

          {/* API Usage Chart */}
          <div className="bg-white shadow rounded-lg p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">API Usage</h3>
            <div className="h-64 bg-gray-100 rounded flex items-center justify-center" data-testid="api-usage-chart">
              <div className="text-center">
                <div className="text-gray-500 mb-2">API Usage Chart</div>
                <div className="flex justify-center space-x-2">
                  {[...Array(7)].map((_, i) => (
                    <div 
                      key={i} 
                      className="bg-blue-500 w-8 rounded-t" 
                      style={{height: `${Math.random() * 80 + 40}px`}}
                    ></div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Subscriber Growth Chart */}
        <div className="bg-white shadow rounded-lg p-6 mb-8">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Subscriber Growth</h3>
          <div className="h-64 bg-gray-100 rounded flex items-center justify-center" data-testid="subscriber-growth-chart">
            <div className="text-center">
              <div className="text-gray-500 mb-2">Subscriber Growth Chart</div>
              <div className="flex justify-center items-end space-x-2">
                {[...Array(12)].map((_, i) => (
                  <div 
                    key={i} 
                    className="bg-green-500 w-6 rounded-t" 
                    style={{height: `${Math.random() * 120 + 30}px`}}
                  ></div>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* Top APIs Performance */}
        <div className="bg-white shadow rounded-lg">
          <div className="px-6 py-4 border-b border-gray-200">
            <h3 className="text-lg font-medium text-gray-900">Top Performing APIs</h3>
          </div>
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">API Name</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Subscribers</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">API Calls</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Revenue</th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Growth</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                <tr>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                    Test Payment API
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">247</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">1,234,567</td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">
                    $12,345.67
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-green-600">+12.5%</td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </Layout>
  );
};

export default CreatorDashboard;
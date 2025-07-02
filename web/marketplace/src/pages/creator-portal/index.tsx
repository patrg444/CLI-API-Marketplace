import React, { useState } from 'react';
import { useRouter } from 'next/router';
import Layout from '../../components/Layout';

const CreatorPortal: React.FC = () => {
  const router = useRouter();
  const [isLoggedIn, setIsLoggedIn] = useState(() => {
    if (typeof window !== 'undefined') {
      return !!localStorage.getItem('mockCreator');
    }
    return false;
  });
  
  const testCreator = {
    email: 'test.creator@example.com',
    password: 'CreatorPass123!'
  };

  const handleLogin = (e: React.FormEvent) => {
    e.preventDefault();
    const formData = new FormData(e.target as HTMLFormElement);
    const email = formData.get('email');
    const password = formData.get('password');
    
    if (email === testCreator.email && password === testCreator.password) {
      setIsLoggedIn(true);
      localStorage.setItem('mockCreator', JSON.stringify({ email }));
    }
  };

  if (!isLoggedIn) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
          <div className="max-w-md w-full space-y-8">
            <div>
              <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
                Creator Portal Login
              </h2>
            </div>
            <form className="mt-8 space-y-6" onSubmit={handleLogin}>
              <div>
                <label htmlFor="email" className="sr-only">Email address</label>
                <input
                  id="email"
                  name="email"
                  type="email"
                  required
                  className="relative block w-full px-3 py-2 border border-gray-300 rounded-md"
                  placeholder="Email address"
                  data-testid="email-input"
                />
              </div>
              <div>
                <label htmlFor="password" className="sr-only">Password</label>
                <input
                  id="password"
                  name="password"
                  type="password"
                  required
                  className="relative block w-full px-3 py-2 border border-gray-300 rounded-md"
                  placeholder="Password"
                  data-testid="password-input"
                />
              </div>
              <div>
                <button
                  type="submit"
                  className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700"
                  data-testid="submit-login"
                >
                  Sign In
                </button>
              </div>
            </form>
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Creator Portal</h1>
          <p className="mt-2 text-lg text-gray-600">Manage your APIs and track earnings</p>
        </div>

        {/* Navigation */}
        <div className="border-b border-gray-200 mb-8">
          <nav className="-mb-px flex space-x-8">
            <button 
              className="border-indigo-500 text-indigo-600 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm"
              onClick={() => router.push('/creator-portal/dashboard')}
            >
              Dashboard
            </button>
            <button 
              className="border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm"
              data-testid="apis-nav"
              onClick={() => {
                window.location.href = '/creator-portal/apis';
              }}
            >
              My APIs
            </button>
            <button 
              className="border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300 whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm"
              onClick={() => router.push('/creator-portal/payouts')}
            >
              Payouts
            </button>
          </nav>
        </div>

        {/* Quick Stats */}
        <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
          <div className="bg-white overflow-hidden shadow rounded-lg">
            <div className="p-5">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <div className="w-8 h-8 bg-indigo-500 rounded-md flex items-center justify-center">
                    <span className="text-white text-sm font-medium">$</span>
                  </div>
                </div>
                <div className="ml-5 w-0 flex-1">
                  <dl>
                    <dt className="text-sm font-medium text-gray-500 truncate">Total Earnings</dt>
                    <dd className="text-lg font-medium text-gray-900" data-testid="total-revenue-kpi">$12,345.67</dd>
                  </dl>
                </div>
              </div>
            </div>
          </div>

          <div className="bg-white overflow-hidden shadow rounded-lg">
            <div className="p-5">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <div className="w-8 h-8 bg-green-500 rounded-md flex items-center justify-center">
                    <span className="text-white text-sm font-medium">👥</span>
                  </div>
                </div>
                <div className="ml-5 w-0 flex-1">
                  <dl>
                    <dt className="text-sm font-medium text-gray-500 truncate">Active Subscribers</dt>
                    <dd className="text-lg font-medium text-gray-900" data-testid="active-subscribers-kpi">247</dd>
                  </dl>
                </div>
              </div>
            </div>
          </div>

          <div className="bg-white overflow-hidden shadow rounded-lg">
            <div className="p-5">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <div className="w-8 h-8 bg-blue-500 rounded-md flex items-center justify-center">
                    <span className="text-white text-sm font-medium">📊</span>
                  </div>
                </div>
                <div className="ml-5 w-0 flex-1">
                  <dl>
                    <dt className="text-sm font-medium text-gray-500 truncate">API Calls</dt>
                    <dd className="text-lg font-medium text-gray-900" data-testid="api-calls-kpi">1,234,567</dd>
                  </dl>
                </div>
              </div>
            </div>
          </div>

          <div className="bg-white overflow-hidden shadow rounded-lg">
            <div className="p-5">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <div className="w-8 h-8 bg-yellow-500 rounded-md flex items-center justify-center">
                    <span className="text-white text-sm font-medium">%</span>
                  </div>
                </div>
                <div className="ml-5 w-0 flex-1">
                  <dl>
                    <dt className="text-sm font-medium text-gray-500 truncate">Conversion Rate</dt>
                    <dd className="text-lg font-medium text-gray-900" data-testid="conversion-rate-kpi">12.5%</dd>
                  </dl>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Charts Section */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
          {/* Revenue Chart */}
          <div className="bg-white shadow rounded-lg p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Revenue Trend</h3>
            <div className="h-64 bg-gray-100 rounded flex items-center justify-center" data-testid="revenue-chart">
              <span className="text-gray-500">Revenue Chart</span>
            </div>
          </div>

          {/* API Usage Chart */}
          <div className="bg-white shadow rounded-lg p-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">API Usage</h3>
            <div className="h-64 bg-gray-100 rounded flex items-center justify-center" data-testid="api-usage-chart">
              <span className="text-gray-500">API Usage Chart</span>
            </div>
          </div>
        </div>

        {/* Subscriber Growth Chart */}
        <div className="bg-white shadow rounded-lg p-6">
          <h3 className="text-lg font-medium text-gray-900 mb-4">Subscriber Growth</h3>
          <div className="h-64 bg-gray-100 rounded flex items-center justify-center" data-testid="subscriber-growth-chart">
            <span className="text-gray-500">Subscriber Growth Chart</span>
          </div>
        </div>
      </div>
    </Layout>
  );
};

export default CreatorPortal;
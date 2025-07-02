import React from 'react';
import Layout from '../components/Layout';
import Link from 'next/link';

const Documentation: React.FC = () => {
  return (
    <Layout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="max-w-4xl mx-auto">
          <div className="text-center mb-12">
            <h1 className="text-4xl font-bold text-gray-900 mb-4">API Documentation</h1>
            <p className="text-xl text-gray-600">
              Learn how to integrate and use APIs from our marketplace
            </p>
          </div>

          <div className="grid gap-8 md:grid-cols-2 lg:grid-cols-3 mb-12">
            {/* Getting Started */}
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow">
              <div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mb-4">
                <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 10V3L4 14h7v7l9-11h-7z" />
                </svg>
              </div>
              <h3 className="text-lg font-semibold text-gray-900 mb-2">Getting Started</h3>
              <p className="text-gray-600 mb-4">Learn the basics of using our API marketplace and how to get your first API key.</p>
              <Link href="/docs/getting-started" className="text-blue-600 hover:text-blue-500 font-medium">
                Read Guide →
              </Link>
            </div>

            {/* Authentication */}
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow">
              <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center mb-4">
                <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                </svg>
              </div>
              <h3 className="text-lg font-semibold text-gray-900 mb-2">Authentication</h3>
              <p className="text-gray-600 mb-4">Secure your API calls with proper authentication methods and API key management.</p>
              <Link href="/docs/authentication" className="text-blue-600 hover:text-blue-500 font-medium">
                Read Guide →
              </Link>
            </div>

            {/* API Reference */}
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow">
              <div className="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center mb-4">
                <svg className="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                </svg>
              </div>
              <h3 className="text-lg font-semibold text-gray-900 mb-2">API Reference</h3>
              <p className="text-gray-600 mb-4">Complete reference documentation for all available endpoints and parameters.</p>
              <Link href="/docs/api-reference" className="text-blue-600 hover:text-blue-500 font-medium">
                Read Guide →
              </Link>
            </div>

            {/* SDKs & Libraries */}
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow">
              <div className="w-12 h-12 bg-orange-100 rounded-lg flex items-center justify-center mb-4">
                <svg className="w-6 h-6 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z" />
                </svg>
              </div>
              <h3 className="text-lg font-semibold text-gray-900 mb-2">SDKs & Libraries</h3>
              <p className="text-gray-600 mb-4">Official SDKs and community libraries for popular programming languages.</p>
              <Link href="/docs/sdks" className="text-blue-600 hover:text-blue-500 font-medium">
                Read Guide →
              </Link>
            </div>

            {/* Examples */}
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow">
              <div className="w-12 h-12 bg-red-100 rounded-lg flex items-center justify-center mb-4">
                <svg className="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 20l4-16m4 4l4 4-4 4M6 16l-4-4 4-4" />
                </svg>
              </div>
              <h3 className="text-lg font-semibold text-gray-900 mb-2">Code Examples</h3>
              <p className="text-gray-600 mb-4">Practical examples and tutorials to help you implement APIs quickly.</p>
              <Link href="/docs/examples" className="text-blue-600 hover:text-blue-500 font-medium">
                Read Guide →
              </Link>
            </div>

            {/* Support */}
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 hover:shadow-md transition-shadow">
              <div className="w-12 h-12 bg-indigo-100 rounded-lg flex items-center justify-center mb-4">
                <svg className="w-6 h-6 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M18.364 5.636l-3.536 3.536m0 5.656l3.536 3.536M9.172 9.172L5.636 5.636m3.536 9.192L5.636 18.364M12 2.25a9.75 9.75 0 100 19.5 9.75 9.75 0 000-19.5z" />
                </svg>
              </div>
              <h3 className="text-lg font-semibold text-gray-900 mb-2">Support</h3>
              <p className="text-gray-600 mb-4">Get help from our support team and community developers.</p>
              <Link href="/docs/support" className="text-blue-600 hover:text-blue-500 font-medium">
                Get Help →
              </Link>
            </div>
          </div>

          {/* Quick Start Section */}
          <div className="bg-gradient-to-r from-blue-50 to-indigo-50 rounded-xl p-8 mb-12">
            <h2 className="text-2xl font-bold text-gray-900 mb-4">Quick Start</h2>
            <p className="text-gray-600 mb-6">
              Get up and running with your first API call in just a few minutes.
            </p>
            
            <div className="space-y-4">
              <div className="flex items-center">
                <div className="flex-shrink-0 w-8 h-8 bg-blue-600 text-white rounded-full flex items-center justify-center text-sm font-semibold">
                  1
                </div>
                <div className="ml-4">
                  <p className="font-medium text-gray-900">Sign up for an account</p>
                  <p className="text-gray-600">Create your free developer account to get started</p>
                </div>
              </div>
              
              <div className="flex items-center">
                <div className="flex-shrink-0 w-8 h-8 bg-blue-600 text-white rounded-full flex items-center justify-center text-sm font-semibold">
                  2
                </div>
                <div className="ml-4">
                  <p className="font-medium text-gray-900">Choose an API</p>
                  <p className="text-gray-600">Browse our marketplace and subscribe to an API</p>
                </div>
              </div>
              
              <div className="flex items-center">
                <div className="flex-shrink-0 w-8 h-8 bg-blue-600 text-white rounded-full flex items-center justify-center text-sm font-semibold">
                  3
                </div>
                <div className="ml-4">
                  <p className="font-medium text-gray-900">Get your API key</p>
                  <p className="text-gray-600">Generate your authentication credentials</p>
                </div>
              </div>
              
              <div className="flex items-center">
                <div className="flex-shrink-0 w-8 h-8 bg-blue-600 text-white rounded-full flex items-center justify-center text-sm font-semibold">
                  4
                </div>
                <div className="ml-4">
                  <p className="font-medium text-gray-900">Make your first call</p>
                  <p className="text-gray-600">Use our code examples to integrate the API</p>
                </div>
              </div>
            </div>

            <div className="mt-8">
              <Link 
                href="/auth/signup"
                className="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 transition-colors"
              >
                Get Started Free
              </Link>
            </div>
          </div>

          {/* Popular APIs Section */}
          <div className="mb-12">
            <h2 className="text-2xl font-bold text-gray-900 mb-6">Popular APIs</h2>
            <div className="grid gap-4 md:grid-cols-2">
              <div className="flex items-center p-4 bg-white rounded-lg border border-gray-200">
                <div className="w-10 h-10 bg-blue-100 rounded-lg flex items-center justify-center mr-4">
                  <svg className="w-5 h-5 text-blue-600" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M4 4a2 2 0 00-2 2v1h16V6a2 2 0 00-2-2H4zM18 9H2v5a2 2 0 002 2h12a2 2 0 002-2V9zM4 13a1 1 0 011-1h1a1 1 0 110 2H5a1 1 0 01-1-1zm5-1a1 1 0 100 2h1a1 1 0 100-2H9z" />
                  </svg>
                </div>
                <div>
                  <h3 className="font-semibold text-gray-900">Payment Processing API</h3>
                  <p className="text-sm text-gray-600">Accept payments securely</p>
                </div>
              </div>
              
              <div className="flex items-center p-4 bg-white rounded-lg border border-gray-200">
                <div className="w-10 h-10 bg-green-100 rounded-lg flex items-center justify-center mr-4">
                  <svg className="w-5 h-5 text-green-600" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M5.05 4.05a7 7 0 119.9 9.9L10 18.9l-4.95-4.95a7 7 0 010-9.9zM10 11a2 2 0 100-4 2 2 0 000 4z" clipRule="evenodd" />
                  </svg>
                </div>
                <div>
                  <h3 className="font-semibold text-gray-900">Geolocation API</h3>
                  <p className="text-sm text-gray-600">Location-based services</p>
                </div>
              </div>
            </div>
          </div>

          {/* Support Contact */}
          <div className="text-center bg-gray-50 rounded-xl p-8">
            <h2 className="text-xl font-bold text-gray-900 mb-2">Need Help?</h2>
            <p className="text-gray-600 mb-4">
              Our support team is here to help you succeed with our APIs.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <Link 
                href="/support"
                className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 transition-colors"
              >
                Contact Support
              </Link>
              <Link 
                href="/community"
                className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 transition-colors"
              >
                Join Community
              </Link>
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
};

export default Documentation;
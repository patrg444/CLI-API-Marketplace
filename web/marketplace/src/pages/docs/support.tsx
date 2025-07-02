import React, { useState } from 'react';
import Layout from '../../components/Layout';
import Link from 'next/link';

const Support: React.FC = () => {
  const [selectedCategory, setSelectedCategory] = useState('general');

  const categories = [
    { id: 'general', title: 'General Questions', icon: '❓' },
    { id: 'technical', title: 'Technical Support', icon: '🔧' },
    { id: 'billing', title: 'Billing & Subscriptions', icon: '💳' },
    { id: 'api-issues', title: 'API Issues', icon: '🐛' },
    { id: 'account', title: 'Account Management', icon: '👤' },
  ];

  const faqs = {
    general: [
      {
        question: "How do I get started with the API marketplace?",
        answer: "Start by creating a free account, browse our API catalog, subscribe to APIs that meet your needs, and generate API keys from your dashboard. Check out our Getting Started guide for detailed instructions."
      },
      {
        question: "What types of APIs are available?",
        answer: "We offer APIs across many categories including AI/ML, Finance, E-commerce, Analytics, Communication, Maps & Location, and more. Browse by category or use our advanced search to find specific APIs."
      },
      {
        question: "Do you offer free tiers?",
        answer: "Yes! Most APIs offer generous free tiers perfect for development and testing. You can always upgrade to paid plans when you're ready for production use."
      },
      {
        question: "How do I contact support?",
        answer: "You can reach us through email, our community forum, or schedule a call with our team. We typically respond within 24 hours for standard inquiries and within 4 hours for urgent technical issues."
      }
    ],
    technical: [
      {
        question: "I'm getting a 401 Unauthorized error",
        answer: "This usually means your API key is invalid, expired, or incorrectly formatted. Make sure you're using 'Bearer YOUR_API_KEY' in the Authorization header and that your key is active in your dashboard."
      },
      {
        question: "How do I handle rate limits?",
        answer: "Check the rate limit headers in API responses (X-RateLimit-Remaining, X-RateLimit-Reset). Implement exponential backoff and respect the Retry-After header when you get 429 responses."
      },
      {
        question: "Can I use the APIs in production?",
        answer: "Absolutely! Our APIs are production-ready with 99.9% uptime SLA. Make sure to use proper error handling, implement retries, and monitor your usage through our analytics dashboard."
      },
      {
        question: "Do you provide SDKs?",
        answer: "We offer official SDKs for JavaScript/Node.js and Python, with more languages coming soon. Community-maintained libraries are also available for other languages."
      }
    ],
    billing: [
      {
        question: "How does billing work?",
        answer: "You're billed monthly based on your usage and subscribed plans. Free tiers don't incur charges. Paid plans are billed at the beginning of each billing cycle, with usage-based charges calculated at the end."
      },
      {
        question: "Can I change my subscription plan?",
        answer: "Yes, you can upgrade or downgrade your plans anytime from your dashboard. Upgrades take effect immediately, while downgrades take effect at the next billing cycle."
      },
      {
        question: "What payment methods do you accept?",
        answer: "We accept all major credit cards (Visa, MasterCard, American Express) and support ACH transfers for enterprise customers. All payments are processed securely through Stripe."
      },
      {
        question: "Can I get a refund?",
        answer: "We offer prorated refunds for downgrades and cancellations within 30 days of subscription. Usage-based charges are non-refundable, but we can provide credits for service disruptions."
      }
    ],
    'api-issues': [
      {
        question: "An API is returning incorrect data",
        answer: "First, check the API documentation to ensure you're using the correct parameters. If the issue persists, contact our support team with your request details and we&apos;ll investigate with the API provider."
      },
      {
        question: "API response times are slow",
        answer: "Check our status page for any ongoing issues. If the problem is specific to your requests, contact support with examples. We monitor API performance and work with providers to resolve issues quickly."
      },
      {
        question: "How do I report an API bug?",
        answer: "Use our bug report form or email us with detailed information including API name, request/response examples, timestamps, and expected vs actual behavior. We'll triage and work with providers to fix issues."
      },
      {
        question: "Can I request new features for an API?",
        answer: "Yes! We work closely with API providers to gather user feedback. Submit feature requests through our portal and we&apos;ll advocate for improvements that benefit the community."
      }
    ],
    account: [
      {
        question: "How do I reset my password?",
        answer: "Click the 'Forgot Password' link on the login page and enter your email. You'll receive a reset link within a few minutes. If you don't see it, check your spam folder."
      },
      {
        question: "Can I change my email address?",
        answer: "Yes, you can update your email in your account settings. You'll need to verify the new email address before the change takes effect."
      },
      {
        question: "How do I delete my account?",
        answer: "Contact our support team to delete your account. We'll cancel all subscriptions and remove your data within 30 days. Note that some billing records may be retained for compliance purposes."
      },
      {
        question: "Can I have multiple team members on my account?",
        answer: "Yes! Pro and Enterprise plans support team management with role-based access controls. You can invite team members and manage their permissions from your dashboard."
      }
    ]
  };

  return (
    <Layout>
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8">
          <Link href="/docs" className="text-blue-600 hover:text-blue-500 font-medium">
            ← Back to Documentation
          </Link>
        </div>

        <div className="mb-12">
          <h1 className="text-4xl font-bold text-gray-900 mb-4">Support Center</h1>
          <p className="text-xl text-gray-600">
            Get help from our support team and community developers. We{'\''}re here to help you succeed.
          </p>
        </div>

        {/* Quick Contact Options */}
        <div className="grid md:grid-cols-3 gap-6 mb-12">
          <div className="bg-white border border-gray-200 rounded-lg p-6 text-center hover:shadow-md transition-shadow">
            <div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mx-auto mb-4">
              <svg className="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 8l7.89 4.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">Email Support</h3>
            <p className="text-gray-600 mb-4 text-sm">Get help via email. We typically respond within 24 hours.</p>
            <a href="mailto:support@marketplace.com" className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 transition-colors">
              Contact Support
            </a>
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-6 text-center hover:shadow-md transition-shadow">
            <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center mx-auto mb-4">
              <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M17 8h2a2 2 0 012 2v6a2 2 0 01-2 2h-2v4l-4-4H9a2 2 0 01-2-2v-6a2 2 0 012-2h8z" />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">Community Forum</h3>
            <p className="text-gray-600 mb-4 text-sm">Connect with other developers and get community support.</p>
            <Link href="/community" className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 transition-colors">
              Join Discussion
            </Link>
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-6 text-center hover:shadow-md transition-shadow">
            <div className="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center mx-auto mb-4">
              <svg className="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-gray-900 mb-2">Live Chat</h3>
            <p className="text-gray-600 mb-4 text-sm">Chat with our team in real-time during business hours.</p>
            <button className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 transition-colors">
              Start Chat
            </button>
          </div>
        </div>

        {/* Status & Resources */}
        <div className="grid md:grid-cols-2 gap-6 mb-12">
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Service Status</h3>
            <div className="flex items-center mb-3">
              <div className="w-3 h-3 bg-green-500 rounded-full mr-3"></div>
              <span className="text-gray-900 font-medium">All Systems Operational</span>
            </div>
            <p className="text-gray-600 text-sm mb-4">
              All APIs and services are running normally. Last updated: 2 minutes ago.
            </p>
            <Link href="/status" className="text-blue-600 hover:text-blue-500 text-sm font-medium">
              View Full Status Page →
            </Link>
          </div>

          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Helpful Resources</h3>
            <ul className="space-y-3">
              <li>
                <Link href="/docs/getting-started" className="text-blue-600 hover:text-blue-500 text-sm">
                  → Getting Started Guide
                </Link>
              </li>
              <li>
                <Link href="/docs/api-reference" className="text-blue-600 hover:text-blue-500 text-sm">
                  → API Reference Documentation
                </Link>
              </li>
              <li>
                <Link href="/docs/examples" className="text-blue-600 hover:text-blue-500 text-sm">
                  → Code Examples & Tutorials
                </Link>
              </li>
              <li>
                <Link href="/changelog" className="text-blue-600 hover:text-blue-500 text-sm">
                  → Changelog & Updates
                </Link>
              </li>
            </ul>
          </div>
        </div>

        {/* FAQ Section */}
        <section>
          <h2 className="text-2xl font-bold text-gray-900 mb-6">Frequently Asked Questions</h2>
          
          {/* Category Selector */}
          <div className="mb-6">
            <div className="flex flex-wrap gap-2">
              {categories.map((category) => (
                <button
                  key={category.id}
                  onClick={() => setSelectedCategory(category.id)}
                  className={`flex items-center px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                    selectedCategory === category.id
                      ? 'bg-blue-100 text-blue-700 border border-blue-200'
                      : 'bg-gray-100 text-gray-700 hover:bg-gray-200 border border-gray-200'
                  }`}
                >
                  <span className="mr-2">{category.icon}</span>
                  {category.title}
                </button>
              ))}
            </div>
          </div>

          {/* FAQ Items */}
          <div className="space-y-4">
            {faqs[selectedCategory as keyof typeof faqs]?.map((faq, index) => (
              <div key={index} className="bg-white border border-gray-200 rounded-lg p-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-3">{faq.question}</h3>
                <p className="text-gray-600 leading-relaxed">{faq.answer}</p>
              </div>
            ))}
          </div>
        </section>

        {/* Contact Form */}
        <section className="mt-16">
          <h2 className="text-2xl font-bold text-gray-900 mb-6">Still Need Help?</h2>
          <div className="bg-white border border-gray-200 rounded-lg p-8">
            <form className="space-y-6">
              <div className="grid md:grid-cols-2 gap-6">
                <div>
                  <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-2">
                    Your Name
                  </label>
                  <input
                    type="text"
                    id="name"
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    placeholder="Enter your full name"
                  />
                </div>
                <div>
                  <label htmlFor="email" className="block text-sm font-medium text-gray-700 mb-2">
                    Email Address
                  </label>
                  <input
                    type="email"
                    id="email"
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                    placeholder="Enter your email"
                  />
                </div>
              </div>

              <div>
                <label htmlFor="subject" className="block text-sm font-medium text-gray-700 mb-2">
                  Subject
                </label>
                <select
                  id="subject"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="">Select a topic</option>
                  <option value="technical">Technical Support</option>
                  <option value="billing">Billing Question</option>
                  <option value="api-issue">API Issue</option>
                  <option value="feature-request">Feature Request</option>
                  <option value="partnership">Partnership Inquiry</option>
                  <option value="other">Other</option>
                </select>
              </div>

              <div>
                <label htmlFor="priority" className="block text-sm font-medium text-gray-700 mb-2">
                  Priority Level
                </label>
                <select
                  id="priority"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="low">Low - General question</option>
                  <option value="medium">Medium - Non-urgent issue</option>
                  <option value="high">High - Affecting my work</option>
                  <option value="urgent">Urgent - Production issue</option>
                </select>
              </div>

              <div>
                <label htmlFor="message" className="block text-sm font-medium text-gray-700 mb-2">
                  Message
                </label>
                <textarea
                  id="message"
                  rows={6}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  placeholder="Please provide as much detail as possible about your question or issue..."
                ></textarea>
              </div>

              <div className="flex items-center">
                <input
                  id="updates"
                  type="checkbox"
                  className="h-4 w-4 text-blue-600 focus:ring-blue-500 border-gray-300 rounded"
                />
                <label htmlFor="updates" className="ml-2 block text-sm text-gray-700">
                  Send me updates about new features and API releases
                </label>
              </div>

              <div className="flex justify-end">
                <button
                  type="submit"
                  className="px-6 py-3 bg-blue-600 text-white text-sm font-medium rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition-colors"
                >
                  Send Message
                </button>
              </div>
            </form>
          </div>
        </section>

        {/* Enterprise Support */}
        <section className="mt-16">
          <div className="bg-gradient-to-r from-purple-50 to-indigo-50 rounded-xl p-8">
            <h2 className="text-2xl font-bold text-gray-900 mb-4">Enterprise Support</h2>
            <p className="text-gray-600 mb-6">
              Need dedicated support, custom SLAs, or priority assistance? Our Enterprise support team is here to help.
            </p>
            
            <div className="grid md:grid-cols-2 gap-6 mb-6">
              <div>
                <h3 className="font-semibold text-gray-900 mb-2">Includes:</h3>
                <ul className="text-gray-600 text-sm space-y-1">
                  <li>• Dedicated support manager</li>
                  <li>• 4-hour response time SLA</li>
                  <li>• Phone & video call support</li>
                  <li>• Custom integration assistance</li>
                  <li>• Priority feature requests</li>
                </ul>
              </div>
              <div>
                <h3 className="font-semibold text-gray-900 mb-2">Perfect for:</h3>
                <ul className="text-gray-600 text-sm space-y-1">
                  <li>• High-volume API usage</li>
                  <li>• Mission-critical applications</li>
                  <li>• Complex integrations</li>
                  <li>• Custom requirements</li>
                  <li>• Team training needs</li>
                </ul>
              </div>
            </div>

            <Link
              href="/contact?subject=Enterprise%20Support"
              className="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-purple-600 hover:bg-purple-700 transition-colors"
            >
              Contact Enterprise Sales
            </Link>
          </div>
        </section>
      </div>
    </Layout>
  );
};

export default Support;
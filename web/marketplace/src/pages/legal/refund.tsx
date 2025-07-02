import React from 'react';
import Head from 'next/head';
import Link from 'next/link';

const RefundPolicy: React.FC = () => {
  return (
    <>
      <Head>
        <title>Refund Policy | API Direct Marketplace</title>
        <meta name="description" content="Refund Policy for API Direct Marketplace" />
      </Head>

      <div className="min-h-screen bg-gray-900">
        <div className="max-w-4xl mx-auto px-4 py-16">
          <div className="bg-gray-800 rounded-lg shadow-lg p-8">
            <h1 className="text-3xl font-bold text-white mb-8">Refund Policy</h1>
            
            <p className="text-gray-300 mb-6 italic">
              Last updated: [DATE]
            </p>

            <div className="space-y-8 text-gray-300">
              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">1. Overview</h2>
                <p className="mb-4">
                  API Direct Marketplace strives to ensure customer satisfaction. This Refund Policy outlines the circumstances under which refunds may be issued for API subscriptions and usage charges.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">2. Subscription Refunds</h2>
                
                <h3 className="text-xl font-semibold text-white mb-2 mt-4">Monthly Subscriptions</h3>
                <ul className="list-disc list-inside ml-4 mb-4">
                  <li>No refunds for partial months</li>
                  <li>Cancellation takes effect at the end of the current billing period</li>
                  <li>You retain access until the end of the paid period</li>
                </ul>

                <h3 className="text-xl font-semibold text-white mb-2">Annual Subscriptions</h3>
                <ul className="list-disc list-inside ml-4 mb-4">
                  <li>Pro-rated refunds available within 30 days of purchase</li>
                  <li>No refunds after 30 days</li>
                  <li>Refund amount calculated based on unused months</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">3. Eligible Refund Circumstances</h2>
                <p className="mb-2">Refunds may be issued in the following situations:</p>
                
                <h3 className="text-xl font-semibold text-white mb-2 mt-4">Service Issues</h3>
                <ul className="list-disc list-inside ml-4 mb-4">
                  <li>API is non-functional for more than 72 consecutive hours</li>
                  <li>API does not match its documented functionality</li>
                  <li>Significant undisclosed changes to API functionality</li>
                  <li>Security breach affecting API service</li>
                </ul>

                <h3 className="text-xl font-semibold text-white mb-2">Billing Errors</h3>
                <ul className="list-disc list-inside ml-4 mb-4">
                  <li>Duplicate charges</li>
                  <li>Incorrect pricing applied</li>
                  <li>Charges after cancellation</li>
                  <li>Unauthorized charges</li>
                </ul>

                <h3 className="text-xl font-semibold text-white mb-2">First-Time Users</h3>
                <p className="mb-4">
                  New users may request a full refund within 48 hours of their first subscription if:
                </p>
                <ul className="list-disc list-inside ml-4">
                  <li>They have made fewer than 10 API calls</li>
                  <li>The API does not meet documented specifications</li>
                  <li>Technical issues prevent API usage</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">4. Non-Refundable Situations</h2>
                <p className="mb-2">Refunds will NOT be issued for:</p>
                <ul className="list-disc list-inside ml-4">
                  <li>Change of mind after the refund period</li>
                  <li>Lack of technical knowledge to implement the API</li>
                  <li>Issues with your own implementation or code</li>
                  <li>Rate limit violations or API abuse</li>
                  <li>Account suspension due to Terms of Service violations</li>
                  <li>Third-party integration issues outside our control</li>
                  <li>Usage-based charges that have been consumed</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">5. Refund Process</h2>
                
                <h3 className="text-xl font-semibold text-white mb-2 mt-4">How to Request a Refund</h3>
                <ol className="list-decimal list-inside ml-4 mb-4">
                  <li>Contact support at support@[DOMAIN]</li>
                  <li>Include your account email and transaction ID</li>
                  <li>Describe the reason for your refund request</li>
                  <li>Provide any relevant screenshots or documentation</li>
                </ol>

                <h3 className="text-xl font-semibold text-white mb-2">Processing Time</h3>
                <ul className="list-disc list-inside ml-4">
                  <li>Review: 2-3 business days</li>
                  <li>Decision notification: Within 5 business days</li>
                  <li>Refund processing: 5-10 business days after approval</li>
                  <li>Bank processing: Additional 3-5 business days</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">6. API Creator Responsibilities</h2>
                <p className="mb-4">
                  API creators are responsible for:
                </p>
                <ul className="list-disc list-inside ml-4">
                  <li>Maintaining accurate API documentation</li>
                  <li>Providing reasonable uptime (minimum 99.5%)</li>
                  <li>Notifying users of breaking changes 30 days in advance</li>
                  <li>Responding to support requests within 48 hours</li>
                </ul>
                <p className="mt-4">
                  Failure to meet these responsibilities may result in automatic refund approval.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">7. Dispute Resolution</h2>
                <p className="mb-4">
                  If your refund request is denied and you disagree with the decision:
                </p>
                <ol className="list-decimal list-inside ml-4">
                  <li>Request a review by senior support staff</li>
                  <li>Provide additional documentation</li>
                  <li>If still unresolved, request mediation</li>
                  <li>As a last resort, initiate a chargeback with your bank</li>
                </ol>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">8. Chargebacks</h2>
                <p className="mb-4">
                  Before initiating a chargeback with your bank or credit card company, please contact us to resolve the issue. Accounts with active chargebacks may be suspended until resolved.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">9. Credits Alternative</h2>
                <p className="mb-4">
                  In some cases, we may offer account credits instead of refunds:
                </p>
                <ul className="list-disc list-inside ml-4">
                  <li>Credits can be used for any API on the platform</li>
                  <li>Credits do not expire</li>
                  <li>Credits may exceed the refund amount as compensation</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">10. Changes to This Policy</h2>
                <p className="mb-4">
                  We reserve the right to modify this Refund Policy at any time. Changes will be effective immediately upon posting to this page.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">11. Contact Information</h2>
                <p className="mb-4">
                  For refund requests and questions about this policy:
                </p>
                <p className="mb-4">
                  Email: support@[DOMAIN]<br />
                  Billing inquiries: billing@[DOMAIN]<br />
                  Response time: 24-48 hours
                </p>
              </section>
            </div>

            <div className="mt-12 pt-8 border-t border-gray-700">
              <Link href="/" className="text-blue-400 hover:text-blue-300">
                ← Back to Home
              </Link>
            </div>
          </div>
        </div>
      </div>
    </>
  );
};

export default RefundPolicy;
import React from 'react';
import Head from 'next/head';
import Link from 'next/link';

const TermsOfService: React.FC = () => {
  return (
    <>
      <Head>
        <title>Terms of Service | API Direct Marketplace</title>
        <meta name="description" content="Terms of Service for API Direct Marketplace" />
      </Head>

      <div className="min-h-screen bg-gray-900">
        <div className="max-w-4xl mx-auto px-4 py-16">
          <div className="bg-gray-800 rounded-lg shadow-lg p-8">
            <h1 className="text-3xl font-bold text-white mb-8">Terms of Service</h1>
            
            <p className="text-gray-300 mb-6 italic">
              Last updated: [DATE]
            </p>

            <div className="space-y-8 text-gray-300">
              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">1. Acceptance of Terms</h2>
                <p className="mb-4">
                  By accessing or using the API Direct Marketplace (&quot;Service&quot;), you agree to be bound by these Terms of Service (&quot;Terms&quot;). If you disagree with any part of these terms, you may not access the Service.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">2. Description of Service</h2>
                <p className="mb-4">
                  API Direct Marketplace provides a platform for API creators to publish, manage, and monetize their APIs, and for consumers to discover, subscribe to, and use these APIs.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">3. User Accounts</h2>
                <p className="mb-4">
                  When you create an account with us, you must provide information that is accurate, complete, and current at all times. You are responsible for safeguarding the password and for all activities that occur under your account.
                </p>
                <ul className="list-disc list-inside ml-4">
                  <li>You must be at least 18 years old to use this Service</li>
                  <li>You are responsible for maintaining the security of your account</li>
                  <li>You must notify us immediately of any unauthorized access</li>
                  <li>One person or legal entity may maintain no more than one account</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">4. API Creator Terms</h2>
                <p className="mb-4">As an API creator, you agree to:</p>
                <ul className="list-disc list-inside ml-4">
                  <li>Provide accurate and complete information about your APIs</li>
                  <li>Maintain the security and availability of your APIs</li>
                  <li>Not publish APIs that violate any laws or regulations</li>
                  <li>Respond to support requests in a timely manner</li>
                  <li>Honor the pricing and terms you set for your APIs</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">5. API Consumer Terms</h2>
                <p className="mb-4">As an API consumer, you agree to:</p>
                <ul className="list-disc list-inside ml-4">
                  <li>Use APIs only for lawful purposes</li>
                  <li>Respect rate limits and usage restrictions</li>
                  <li>Not attempt to reverse engineer or bypass API security</li>
                  <li>Pay all fees associated with your API usage</li>
                  <li>Not resell or redistribute API access without permission</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">6. Payments and Fees</h2>
                <p className="mb-4">
                  All payments are processed through our third-party payment processor, Stripe. By using paid APIs, you agree to pay all applicable fees.
                </p>
                <ul className="list-disc list-inside ml-4">
                  <li>Platform fee: 20% of API revenue</li>
                  <li>Payment processing fees may apply</li>
                  <li>Creators receive payouts monthly</li>
                  <li>All fees are non-refundable unless otherwise stated</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">7. Prohibited Uses</h2>
                <p className="mb-4">You may not use the Service to:</p>
                <ul className="list-disc list-inside ml-4">
                  <li>Violate any laws or regulations</li>
                  <li>Infringe on intellectual property rights</li>
                  <li>Transmit malware or harmful code</li>
                  <li>Engage in fraudulent activities</li>
                  <li>Harass, abuse, or harm others</li>
                  <li>Attempt to gain unauthorized access to systems</li>
                  <li>Interfere with the proper functioning of the Service</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">8. Intellectual Property</h2>
                <p className="mb-4">
                  The Service and its original content, features, and functionality are owned by API Direct and are protected by international copyright, trademark, patent, trade secret, and other intellectual property laws.
                </p>
                <p className="mb-4">
                  API creators retain ownership of their API code and documentation, but grant us a license to host and distribute their APIs through our platform.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">9. Privacy</h2>
                <p className="mb-4">
                  Your use of the Service is also governed by our <Link href="/legal/privacy" className="text-blue-400 hover:text-blue-300">Privacy Policy</Link>.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">10. Disclaimers</h2>
                <p className="mb-4">
                  THE SERVICE IS PROVIDED &quot;AS IS&quot; WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, AND NON-INFRINGEMENT.
                </p>
                <p className="mb-4">
                  We do not warrant that the Service will be uninterrupted, secure, or error-free. We are not responsible for the availability, reliability, or performance of third-party APIs.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">11. Limitation of Liability</h2>
                <p className="mb-4">
                  TO THE MAXIMUM EXTENT PERMITTED BY LAW, API DIRECT SHALL NOT BE LIABLE FOR ANY INDIRECT, INCIDENTAL, SPECIAL, CONSEQUENTIAL, OR PUNITIVE DAMAGES, OR ANY LOSS OF PROFITS OR REVENUES, WHETHER INCURRED DIRECTLY OR INDIRECTLY, OR ANY LOSS OF DATA, USE, GOODWILL, OR OTHER INTANGIBLE LOSSES.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">12. Indemnification</h2>
                <p className="mb-4">
                  You agree to defend, indemnify, and hold harmless API Direct and its officers, directors, employees, and agents from any claims, damages, obligations, losses, liabilities, costs, or debt arising from your use of the Service or violation of these Terms.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">13. Termination</h2>
                <p className="mb-4">
                  We may terminate or suspend your account immediately, without prior notice or liability, for any reason, including breach of these Terms. Upon termination, your right to use the Service will cease immediately.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">14. Changes to Terms</h2>
                <p className="mb-4">
                  We reserve the right to modify these Terms at any time. We will notify users of any material changes by posting the new Terms on this page and updating the &quot;Last updated&quot; date.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">15. Governing Law</h2>
                <p className="mb-4">
                  These Terms shall be governed by and construed in accordance with the laws of [JURISDICTION], without regard to its conflict of law provisions.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">16. Contact Information</h2>
                <p className="mb-4">
                  If you have any questions about these Terms, please contact us at:
                </p>
                <p className="mb-4">
                  Email: legal@[DOMAIN]<br />
                  Address: [COMPANY ADDRESS]
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

export default TermsOfService;
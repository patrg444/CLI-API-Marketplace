import React from 'react';
import Head from 'next/head';
import Link from 'next/link';

const PrivacyPolicy: React.FC = () => {
  return (
    <>
      <Head>
        <title>Privacy Policy | API Direct Marketplace</title>
        <meta name="description" content="Privacy Policy for API Direct Marketplace" />
      </Head>

      <div className="min-h-screen bg-gray-900">
        <div className="max-w-4xl mx-auto px-4 py-16">
          <div className="bg-gray-800 rounded-lg shadow-lg p-8">
            <h1 className="text-3xl font-bold text-white mb-8">Privacy Policy</h1>
            
            <p className="text-gray-300 mb-6 italic">
              Last updated: [DATE]
            </p>

            <div className="space-y-8 text-gray-300">
              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">1. Introduction</h2>
                <p className="mb-4">
                  API Direct Marketplace (&quot;we,&quot; &quot;our,&quot; or &quot;us&quot;) is committed to protecting your privacy. This Privacy Policy explains how we collect, use, disclose, and safeguard your information when you use our service.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">2. Information We Collect</h2>
                
                <h3 className="text-xl font-semibold text-white mb-2 mt-4">Personal Information</h3>
                <p className="mb-2">We collect information you provide directly to us, such as:</p>
                <ul className="list-disc list-inside ml-4 mb-4">
                  <li>Name and email address</li>
                  <li>Account credentials</li>
                  <li>Payment information (processed by Stripe)</li>
                  <li>Company information</li>
                  <li>API documentation and code (for creators)</li>
                </ul>

                <h3 className="text-xl font-semibold text-white mb-2">Usage Information</h3>
                <p className="mb-2">We automatically collect information about your use of the Service:</p>
                <ul className="list-disc list-inside ml-4 mb-4">
                  <li>API usage metrics and logs</li>
                  <li>IP address and device information</li>
                  <li>Browser type and operating system</li>
                  <li>Pages visited and features used</li>
                  <li>Referral URLs</li>
                </ul>

                <h3 className="text-xl font-semibold text-white mb-2">Cookies and Tracking</h3>
                <p className="mb-4">
                  We use cookies and similar tracking technologies to track activity on our Service and hold certain information. You can instruct your browser to refuse all cookies or indicate when a cookie is being sent.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">3. How We Use Your Information</h2>
                <p className="mb-2">We use the information we collect to:</p>
                <ul className="list-disc list-inside ml-4">
                  <li>Provide, maintain, and improve our Service</li>
                  <li>Process transactions and send related information</li>
                  <li>Send technical notices, updates, and support messages</li>
                  <li>Respond to your comments and questions</li>
                  <li>Monitor and analyze usage patterns and trends</li>
                  <li>Detect, prevent, and address technical issues</li>
                  <li>Protect against fraudulent or illegal activity</li>
                  <li>Comply with legal obligations</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">4. How We Share Your Information</h2>
                <p className="mb-4">We may share your information in the following situations:</p>
                
                <h3 className="text-xl font-semibold text-white mb-2">With API Creators and Consumers</h3>
                <ul className="list-disc list-inside ml-4 mb-4">
                  <li>API creators can see usage data for their APIs</li>
                  <li>Your username may be visible in reviews and ratings</li>
                  <li>API usage logs may be shared with creators for support</li>
                </ul>

                <h3 className="text-xl font-semibold text-white mb-2">With Service Providers</h3>
                <ul className="list-disc list-inside ml-4 mb-4">
                  <li>Payment processing (Stripe)</li>
                  <li>Cloud infrastructure (AWS)</li>
                  <li>Analytics services</li>
                  <li>Customer support tools</li>
                </ul>

                <h3 className="text-xl font-semibold text-white mb-2">Legal Requirements</h3>
                <p className="mb-4">
                  We may disclose your information if required by law or in response to valid requests by public authorities.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">5. Data Security</h2>
                <p className="mb-4">
                  We implement appropriate technical and organizational measures to protect your personal information against unauthorized or unlawful processing, accidental loss, destruction, or damage.
                </p>
                <ul className="list-disc list-inside ml-4">
                  <li>Encryption of data in transit and at rest</li>
                  <li>Regular security audits and penetration testing</li>
                  <li>Access controls and authentication</li>
                  <li>Regular backups and disaster recovery procedures</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">6. Data Retention</h2>
                <p className="mb-4">
                  We retain your personal information for as long as necessary to provide our services and comply with legal obligations. Specifically:
                </p>
                <ul className="list-disc list-inside ml-4">
                  <li>Account information: Until account deletion</li>
                  <li>Transaction records: 7 years for tax purposes</li>
                  <li>API usage logs: 90 days</li>
                  <li>Support communications: 2 years</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">7. Your Rights</h2>
                <p className="mb-2">You have the right to:</p>
                <ul className="list-disc list-inside ml-4">
                  <li>Access your personal information</li>
                  <li>Correct inaccurate information</li>
                  <li>Request deletion of your information</li>
                  <li>Object to processing of your information</li>
                  <li>Request data portability</li>
                  <li>Withdraw consent at any time</li>
                </ul>
                <p className="mt-4">
                  To exercise these rights, please contact us at privacy@[DOMAIN].
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">8. International Data Transfers</h2>
                <p className="mb-4">
                  Your information may be transferred to and processed in countries other than your country of residence. We ensure appropriate safeguards are in place to protect your information in accordance with this Privacy Policy.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">9. Children&apos;s Privacy</h2>
                <p className="mb-4">
                  Our Service is not intended for individuals under the age of 18. We do not knowingly collect personal information from children under 18. If we become aware that we have collected personal information from a child under 18, we will take steps to delete such information.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">10. Third-Party Links</h2>
                <p className="mb-4">
                  Our Service may contain links to third-party websites. We are not responsible for the privacy practices of these external sites. We encourage you to review their privacy policies.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">11. California Privacy Rights</h2>
                <p className="mb-4">
                  If you are a California resident, you have additional rights under the California Consumer Privacy Act (CCPA), including the right to opt-out of the sale of personal information. We do not sell personal information.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">12. Changes to This Policy</h2>
                <p className="mb-4">
                  We may update this Privacy Policy from time to time. We will notify you of any changes by posting the new Privacy Policy on this page and updating the &quot;Last updated&quot; date.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">13. Contact Us</h2>
                <p className="mb-4">
                  If you have questions about this Privacy Policy, please contact us:
                </p>
                <p className="mb-4">
                  Email: privacy@[DOMAIN]<br />
                  Data Protection Officer: dpo@[DOMAIN]<br />
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

export default PrivacyPolicy;
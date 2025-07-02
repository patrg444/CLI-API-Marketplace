import React from 'react';
import Head from 'next/head';
import Link from 'next/link';

const APITerms: React.FC = () => {
  return (
    <>
      <Head>
        <title>API Usage Terms | API Direct Marketplace</title>
        <meta name="description" content="API Usage Terms for API Direct Marketplace" />
      </Head>

      <div className="min-h-screen bg-gray-900">
        <div className="max-w-4xl mx-auto px-4 py-16">
          <div className="bg-gray-800 rounded-lg shadow-lg p-8">
            <h1 className="text-3xl font-bold text-white mb-8">API Usage Terms</h1>
            
            <p className="text-gray-300 mb-6 italic">
              Last updated: [DATE]
            </p>

            <div className="space-y-8 text-gray-300">
              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">1. Introduction</h2>
                <p className="mb-4">
                  These API Usage Terms (&quot;API Terms&quot;) govern your use of APIs available through the API Direct Marketplace. By accessing or using any API, you agree to these terms in addition to our general Terms of Service.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">2. API Access and Authentication</h2>
                
                <h3 className="text-xl font-semibold text-white mb-2 mt-4">API Keys</h3>
                <ul className="list-disc list-inside ml-4 mb-4">
                  <li>Keep your API keys secure and confidential</li>
                  <li>Do not share API keys or embed them in client-side code</li>
                  <li>Rotate keys regularly and immediately if compromised</li>
                  <li>You are responsible for all usage under your API keys</li>
                </ul>

                <h3 className="text-xl font-semibold text-white mb-2">Access Restrictions</h3>
                <ul className="list-disc list-inside ml-4">
                  <li>Use APIs only for their intended purpose</li>
                  <li>Do not attempt to access APIs without valid credentials</li>
                  <li>Respect IP allowlists and geographic restrictions</li>
                  <li>Do not circumvent authentication mechanisms</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">3. Rate Limits and Fair Use</h2>
                
                <h3 className="text-xl font-semibold text-white mb-2 mt-4">Rate Limiting</h3>
                <p className="mb-2">All APIs are subject to rate limits:</p>
                <ul className="list-disc list-inside ml-4 mb-4">
                  <li>Respect the rate limits specified in API documentation</li>
                  <li>Implement exponential backoff for retry logic</li>
                  <li>Cache responses when appropriate</li>
                  <li>Contact support for higher limits if needed</li>
                </ul>

                <h3 className="text-xl font-semibold text-white mb-2">Fair Use Policy</h3>
                <ul className="list-disc list-inside ml-4">
                  <li>Do not use excessive bandwidth or resources</li>
                  <li>Avoid unnecessary API calls</li>
                  <li>Implement efficient polling intervals</li>
                  <li>Use webhooks when available instead of polling</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">4. Prohibited Uses</h2>
                <p className="mb-2">You may NOT use APIs to:</p>
                
                <h3 className="text-xl font-semibold text-white mb-2 mt-4">Illegal Activities</h3>
                <ul className="list-disc list-inside ml-4 mb-4">
                  <li>Violate any laws or regulations</li>
                  <li>Facilitate fraud or deception</li>
                  <li>Infringe on intellectual property rights</li>
                  <li>Process illegal content or transactions</li>
                </ul>

                <h3 className="text-xl font-semibold text-white mb-2">Technical Abuse</h3>
                <ul className="list-disc list-inside ml-4 mb-4">
                  <li>Perform denial-of-service attacks</li>
                  <li>Inject malicious code or malware</li>
                  <li>Attempt to reverse engineer APIs</li>
                  <li>Scrape or harvest data in bulk</li>
                  <li>Create derivative works without permission</li>
                </ul>

                <h3 className="text-xl font-semibold text-white mb-2">Commercial Restrictions</h3>
                <ul className="list-disc list-inside ml-4">
                  <li>Resell API access without authorization</li>
                  <li>Use APIs to compete with API Direct</li>
                  <li>Sublicense API access to third parties</li>
                  <li>Remove or obscure any proprietary notices</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">5. Data Usage and Privacy</h2>
                
                <h3 className="text-xl font-semibold text-white mb-2 mt-4">Data Processing</h3>
                <ul className="list-disc list-inside ml-4 mb-4">
                  <li>Process data in accordance with applicable privacy laws</li>
                  <li>Implement appropriate security measures</li>
                  <li>Delete data when no longer needed</li>
                  <li>Notify users of data collection and usage</li>
                </ul>

                <h3 className="text-xl font-semibold text-white mb-2">Data Restrictions</h3>
                <ul className="list-disc list-inside ml-4">
                  <li>Do not store sensitive data unless necessary</li>
                  <li>Encrypt sensitive data in transit and at rest</li>
                  <li>Do not share API data with unauthorized parties</li>
                  <li>Comply with data localization requirements</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">6. Service Level Agreement (SLA)</h2>
                
                <h3 className="text-xl font-semibold text-white mb-2 mt-4">Uptime Commitments</h3>
                <ul className="list-disc list-inside ml-4 mb-4">
                  <li>Standard APIs: 99.5% uptime</li>
                  <li>Premium APIs: 99.9% uptime</li>
                  <li>Uptime measured monthly</li>
                  <li>Excludes scheduled maintenance</li>
                </ul>

                <h3 className="text-xl font-semibold text-white mb-2">SLA Credits</h3>
                <table className="w-full border-collapse mb-4">
                  <thead>
                    <tr className="border-b border-gray-600">
                      <th className="text-left py-2">Uptime</th>
                      <th className="text-left py-2">Credit</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr className="border-b border-gray-700">
                      <td className="py-2">99.0% - 99.5%</td>
                      <td className="py-2">10%</td>
                    </tr>
                    <tr className="border-b border-gray-700">
                      <td className="py-2">95.0% - 99.0%</td>
                      <td className="py-2">25%</td>
                    </tr>
                    <tr className="border-b border-gray-700">
                      <td className="py-2">Below 95.0%</td>
                      <td className="py-2">50%</td>
                    </tr>
                  </tbody>
                </table>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">7. API Versioning and Changes</h2>
                
                <h3 className="text-xl font-semibold text-white mb-2 mt-4">Version Support</h3>
                <ul className="list-disc list-inside ml-4 mb-4">
                  <li>Current version: Full support</li>
                  <li>Previous version: 12 months support</li>
                  <li>Older versions: Best effort basis</li>
                  <li>Security patches for all supported versions</li>
                </ul>

                <h3 className="text-xl font-semibold text-white mb-2">Breaking Changes</h3>
                <ul className="list-disc list-inside ml-4">
                  <li>30 days advance notice for breaking changes</li>
                  <li>Migration guides provided</li>
                  <li>Deprecation warnings in API responses</li>
                  <li>Extended notice for major changes</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">8. Support and Maintenance</h2>
                
                <h3 className="text-xl font-semibold text-white mb-2 mt-4">Support Levels</h3>
                <table className="w-full border-collapse mb-4">
                  <thead>
                    <tr className="border-b border-gray-600">
                      <th className="text-left py-2">Plan</th>
                      <th className="text-left py-2">Response Time</th>
                      <th className="text-left py-2">Channels</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr className="border-b border-gray-700">
                      <td className="py-2">Free</td>
                      <td className="py-2">Best effort</td>
                      <td className="py-2">Community forum</td>
                    </tr>
                    <tr className="border-b border-gray-700">
                      <td className="py-2">Basic</td>
                      <td className="py-2">48 hours</td>
                      <td className="py-2">Email</td>
                    </tr>
                    <tr className="border-b border-gray-700">
                      <td className="py-2">Premium</td>
                      <td className="py-2">24 hours</td>
                      <td className="py-2">Email, Chat</td>
                    </tr>
                    <tr className="border-b border-gray-700">
                      <td className="py-2">Enterprise</td>
                      <td className="py-2">4 hours</td>
                      <td className="py-2">Email, Chat, Phone</td>
                    </tr>
                  </tbody>
                </table>

                <h3 className="text-xl font-semibold text-white mb-2">Maintenance Windows</h3>
                <ul className="list-disc list-inside ml-4">
                  <li>Scheduled: Sundays 2-4 AM UTC</li>
                  <li>Emergency: As needed with notice</li>
                  <li>Updates posted to status page</li>
                  <li>Email notifications for major maintenance</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">9. Liability and Indemnification</h2>
                <p className="mb-4">
                  API creators and API Direct disclaim liability for:
                </p>
                <ul className="list-disc list-inside ml-4">
                  <li>Indirect or consequential damages</li>
                  <li>Lost profits or revenue</li>
                  <li>Data loss or corruption</li>
                  <li>Third-party claims</li>
                </ul>
                <p className="mt-4">
                  You agree to indemnify API creators and API Direct against claims arising from your use of APIs.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">10. Termination</h2>
                <p className="mb-4">
                  API access may be terminated for:
                </p>
                <ul className="list-disc list-inside ml-4">
                  <li>Violation of these terms</li>
                  <li>Non-payment of fees</li>
                  <li>Excessive abuse or misuse</li>
                  <li>Legal or regulatory requirements</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">11. Contact Information</h2>
                <p className="mb-4">
                  For API-related questions and support:
                </p>
                <p className="mb-4">
                  Technical Support: api-support@[DOMAIN]<br />
                  Security Issues: security@[DOMAIN]<br />
                  Legal Questions: legal@[DOMAIN]
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

export default APITerms;
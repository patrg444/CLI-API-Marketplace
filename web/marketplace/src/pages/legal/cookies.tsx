import React from 'react';
import Head from 'next/head';
import Link from 'next/link';

const CookiePolicy: React.FC = () => {
  return (
    <>
      <Head>
        <title>Cookie Policy | API Direct Marketplace</title>
        <meta name="description" content="Cookie Policy for API Direct Marketplace" />
      </Head>

      <div className="min-h-screen bg-gray-900">
        <div className="max-w-4xl mx-auto px-4 py-16">
          <div className="bg-gray-800 rounded-lg shadow-lg p-8">
            <h1 className="text-3xl font-bold text-white mb-8">Cookie Policy</h1>
            
            <p className="text-gray-300 mb-6 italic">
              Last updated: [DATE]
            </p>

            <div className="space-y-8 text-gray-300">
              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">1. What Are Cookies</h2>
                <p className="mb-4">
                  Cookies are small text files that are placed on your computer or mobile device when you visit a website. They are widely used to make websites work efficiently and provide information to website owners.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">2. How We Use Cookies</h2>
                <p className="mb-4">
                  API Direct Marketplace uses cookies to:
                </p>
                <ul className="list-disc list-inside ml-4">
                  <li>Keep you signed in to your account</li>
                  <li>Remember your preferences and settings</li>
                  <li>Analyze how our service is used</li>
                  <li>Protect against fraud and improve security</li>
                  <li>Deliver personalized content</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">3. Types of Cookies We Use</h2>
                
                <h3 className="text-xl font-semibold text-white mb-2 mt-4">Essential Cookies</h3>
                <p className="mb-4">
                  These cookies are necessary for the website to function properly. They enable basic functions like page navigation and access to secure areas.
                </p>
                <table className="w-full border-collapse mb-4">
                  <thead>
                    <tr className="border-b border-gray-600">
                      <th className="text-left py-2">Cookie Name</th>
                      <th className="text-left py-2">Purpose</th>
                      <th className="text-left py-2">Duration</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr className="border-b border-gray-700">
                      <td className="py-2">session_id</td>
                      <td className="py-2">Maintains user session</td>
                      <td className="py-2">Session</td>
                    </tr>
                    <tr className="border-b border-gray-700">
                      <td className="py-2">auth_token</td>
                      <td className="py-2">Authentication</td>
                      <td className="py-2">30 days</td>
                    </tr>
                    <tr className="border-b border-gray-700">
                      <td className="py-2">csrf_token</td>
                      <td className="py-2">Security</td>
                      <td className="py-2">Session</td>
                    </tr>
                  </tbody>
                </table>

                <h3 className="text-xl font-semibold text-white mb-2 mt-4">Analytics Cookies</h3>
                <p className="mb-4">
                  These cookies help us understand how visitors interact with our website by collecting and reporting information anonymously.
                </p>
                <table className="w-full border-collapse mb-4">
                  <thead>
                    <tr className="border-b border-gray-600">
                      <th className="text-left py-2">Cookie Name</th>
                      <th className="text-left py-2">Provider</th>
                      <th className="text-left py-2">Duration</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr className="border-b border-gray-700">
                      <td className="py-2">_ga</td>
                      <td className="py-2">Google Analytics</td>
                      <td className="py-2">2 years</td>
                    </tr>
                    <tr className="border-b border-gray-700">
                      <td className="py-2">_gid</td>
                      <td className="py-2">Google Analytics</td>
                      <td className="py-2">24 hours</td>
                    </tr>
                    <tr className="border-b border-gray-700">
                      <td className="py-2">mp_token</td>
                      <td className="py-2">Mixpanel</td>
                      <td className="py-2">1 year</td>
                    </tr>
                  </tbody>
                </table>

                <h3 className="text-xl font-semibold text-white mb-2 mt-4">Functionality Cookies</h3>
                <p className="mb-4">
                  These cookies enable enhanced functionality and personalization, such as remembering your preferences.
                </p>
                <table className="w-full border-collapse mb-4">
                  <thead>
                    <tr className="border-b border-gray-600">
                      <th className="text-left py-2">Cookie Name</th>
                      <th className="text-left py-2">Purpose</th>
                      <th className="text-left py-2">Duration</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr className="border-b border-gray-700">
                      <td className="py-2">theme</td>
                      <td className="py-2">UI theme preference</td>
                      <td className="py-2">1 year</td>
                    </tr>
                    <tr className="border-b border-gray-700">
                      <td className="py-2">lang</td>
                      <td className="py-2">Language preference</td>
                      <td className="py-2">1 year</td>
                    </tr>
                  </tbody>
                </table>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">4. Third-Party Cookies</h2>
                <p className="mb-4">
                  Some cookies are placed by third-party services that appear on our pages. We do not control these cookies. Third-party cookies on our site include:
                </p>
                <ul className="list-disc list-inside ml-4">
                  <li>Stripe - Payment processing</li>
                  <li>Google Analytics - Usage analytics</li>
                  <li>Mixpanel - Product analytics</li>
                  <li>AWS Cognito - Authentication</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">5. Managing Cookies</h2>
                <p className="mb-4">
                  You can control and manage cookies in various ways:
                </p>
                
                <h3 className="text-xl font-semibold text-white mb-2 mt-4">Browser Settings</h3>
                <p className="mb-4">
                  Most browsers allow you to:
                </p>
                <ul className="list-disc list-inside ml-4 mb-4">
                  <li>See what cookies you have and delete them individually</li>
                  <li>Block third-party cookies</li>
                  <li>Block all cookies from specific sites</li>
                  <li>Block all cookies from being set</li>
                  <li>Delete all cookies when you close your browser</li>
                </ul>

                <h3 className="text-xl font-semibold text-white mb-2">Cookie Settings Links</h3>
                <ul className="list-disc list-inside ml-4">
                  <li><a href="https://support.google.com/chrome/answer/95647" className="text-blue-400 hover:text-blue-300" target="_blank" rel="noopener noreferrer">Chrome</a></li>
                  <li><a href="https://support.mozilla.org/en-US/kb/cookies" className="text-blue-400 hover:text-blue-300" target="_blank" rel="noopener noreferrer">Firefox</a></li>
                  <li><a href="https://support.apple.com/guide/safari/manage-cookies-and-website-data-sfri11471/mac" className="text-blue-400 hover:text-blue-300" target="_blank" rel="noopener noreferrer">Safari</a></li>
                  <li><a href="https://support.microsoft.com/en-us/help/17442/windows-internet-explorer-delete-manage-cookies" className="text-blue-400 hover:text-blue-300" target="_blank" rel="noopener noreferrer">Internet Explorer</a></li>
                  <li><a href="https://help.opera.com/en/latest/web-preferences/#cookies" className="text-blue-400 hover:text-blue-300" target="_blank" rel="noopener noreferrer">Opera</a></li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">6. Impact of Disabling Cookies</h2>
                <p className="mb-4">
                  Please note that if you disable cookies:
                </p>
                <ul className="list-disc list-inside ml-4">
                  <li>You may not be able to sign in to your account</li>
                  <li>Some features may not function properly</li>
                  <li>Your preferences may not be saved</li>
                  <li>You may see less relevant content</li>
                </ul>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">7. Changes to This Policy</h2>
                <p className="mb-4">
                  We may update this Cookie Policy from time to time. We will notify you of any changes by posting the new Cookie Policy on this page and updating the &quot;Last updated&quot; date.
                </p>
              </section>

              <section>
                <h2 className="text-2xl font-semibold text-white mb-4">8. Contact Us</h2>
                <p className="mb-4">
                  If you have questions about our use of cookies, please contact us:
                </p>
                <p className="mb-4">
                  Email: privacy@[DOMAIN]<br />
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

export default CookiePolicy;
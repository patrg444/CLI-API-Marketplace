import React, { useState } from 'react';
import Layout from '../../components/Layout';
import Link from 'next/link';

const SDKs: React.FC = () => {
  const [activeLanguage, setActiveLanguage] = useState('javascript');

  const languages = [
    { id: 'javascript', name: 'JavaScript/Node.js', icon: '🟨' },
    { id: 'python', name: 'Python', icon: '🐍' },
    { id: 'java', name: 'Java', icon: '☕' },
    { id: 'php', name: 'PHP', icon: '🐘' },
    { id: 'ruby', name: 'Ruby', icon: '💎' },
    { id: 'go', name: 'Go', icon: '🐹' },
    { id: 'csharp', name: 'C#', icon: '🔷' },
    { id: 'swift', name: 'Swift', icon: '🍎' },
  ];

  return (
    <Layout>
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="mb-8">
          <Link href="/docs" className="text-blue-600 hover:text-blue-500 font-medium">
            ← Back to Documentation
          </Link>
        </div>

        <div className="mb-12">
          <h1 className="text-4xl font-bold text-gray-900 mb-4">SDKs & Libraries</h1>
          <p className="text-xl text-gray-600">
            Official SDKs and community libraries for popular programming languages.
          </p>
        </div>

        {/* Language Selection */}
        <div className="mb-8">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Choose Your Language</h2>
          <div className="grid grid-cols-2 md:grid-cols-4 gap-3">
            {languages.map((lang) => (
              <button
                key={lang.id}
                onClick={() => setActiveLanguage(lang.id)}
                className={`p-4 rounded-lg border text-left transition-all duration-200 ${
                  activeLanguage === lang.id
                    ? 'border-blue-500 bg-blue-50 shadow-md'
                    : 'border-gray-200 bg-white hover:border-gray-300 hover:shadow-sm'
                }`}
              >
                <div className="text-2xl mb-2">{lang.icon}</div>
                <div className="font-medium text-gray-900 text-sm">{lang.name}</div>
              </button>
            ))}
          </div>
        </div>

        {/* JavaScript SDK */}
        {activeLanguage === 'javascript' && (
          <div className="space-y-8">
            <section>
              <h2 className="text-2xl font-bold text-gray-900 mb-4">JavaScript/Node.js SDK</h2>
              
              <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-3">Installation</h3>
                <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`npm install @marketplace/sdk
# or
yarn add @marketplace/sdk`}
                </pre>
              </div>

              <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-3">Quick Start</h3>
                <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`import { MarketplaceClient } from '@marketplace/sdk';

const client = new MarketplaceClient({
  apiKey: process.env.MARKETPLACE_API_KEY,
  environment: 'production' // or 'sandbox'
});

// List available APIs
const apis = await client.apis.list({
  category: 'AI/ML',
  limit: 10
});

// Subscribe to an API
const subscription = await client.subscriptions.create({
  apiId: 'api_12345',
  plan: 'pro'
});

// Make API calls through subscribed APIs
const result = await client.proxy.call('api_12345', {
  method: 'GET',
  endpoint: '/analyze',
  data: { image_url: 'https://example.com/image.jpg' }
});`}
                </pre>
              </div>

              <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-3">Configuration Options</h3>
                <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`const client = new MarketplaceClient({
  apiKey: 'your-api-key',
  environment: 'production',
  timeout: 30000,
  retries: 3,
  baseURL: 'https://api.marketplace.com/v1',
  userAgent: 'MyApp/1.0.0'
});`}
                </pre>
              </div>

              <div className="bg-white border border-gray-200 rounded-lg p-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-3">Error Handling</h3>
                <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`try {
  const result = await client.apis.get('api_12345');
} catch (error) {
  if (error instanceof MarketplaceError) {
    console.error('API Error:', error.code, error.message);
    console.error('Details:', error.details);
  } else {
    console.error('Network Error:', error.message);
  }
}`}
                </pre>
              </div>
            </section>
          </div>
        )}

        {/* Python SDK */}
        {activeLanguage === 'python' && (
          <div className="space-y-8">
            <section>
              <h2 className="text-2xl font-bold text-gray-900 mb-4">Python SDK</h2>
              
              <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-3">Installation</h3>
                <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`pip install marketplace-sdk`}
                </pre>
              </div>

              <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-3">Quick Start</h3>
                <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`from marketplace_sdk import MarketplaceClient

client = MarketplaceClient(
    api_key="your-api-key",
    environment="production"
)

# List available APIs
apis = client.apis.list(category="AI/ML", limit=10)

# Subscribe to an API
subscription = client.subscriptions.create(
    api_id="api_12345",
    plan="pro"
)

# Make API calls
result = client.proxy.call(
    api_id="api_12345",
    method="GET",
    endpoint="/analyze",
    data={"image_url": "https://example.com/image.jpg"}
)`}
                </pre>
              </div>

              <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-3">Async Support</h3>
                <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`import asyncio
from marketplace_sdk import AsyncMarketplaceClient

async def main():
    client = AsyncMarketplaceClient(api_key="your-api-key")
    
    apis = await client.apis.list(category="AI/ML")
    print(f"Found {len(apis.data)} APIs")
    
    await client.close()

asyncio.run(main())`}
                </pre>
              </div>

              <div className="bg-white border border-gray-200 rounded-lg p-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-3">Error Handling</h3>
                <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`from marketplace_sdk import MarketplaceError, RateLimitError

try:
    result = client.apis.get("api_12345")
except RateLimitError as e:
    print(f"Rate limit exceeded. Retry after: {e.retry_after}")
except MarketplaceError as e:
    print(f"API Error {e.code}: {e.message}")
except Exception as e:
    print(f"Unexpected error: {e}")`}
                </pre>
              </div>
            </section>
          </div>
        )}

        {/* Java SDK */}
        {activeLanguage === 'java' && (
          <div className="space-y-8">
            <section>
              <h2 className="text-2xl font-bold text-gray-900 mb-4">Java SDK</h2>
              
              <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-3">Installation (Maven)</h3>
                <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`<dependency>
    <groupId>com.marketplace</groupId>
    <artifactId>marketplace-sdk</artifactId>
    <version>1.0.0</version>
</dependency>`}
                </pre>
              </div>

              <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-3">Installation (Gradle)</h3>
                <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`implementation 'com.marketplace:marketplace-sdk:1.0.0'`}
                </pre>
              </div>

              <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-3">Quick Start</h3>
                <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`import com.marketplace.sdk.MarketplaceClient;
import com.marketplace.sdk.models.*;

MarketplaceClient client = MarketplaceClient.builder()
    .apiKey("your-api-key")
    .environment(Environment.PRODUCTION)
    .build();

// List APIs
ApiListRequest request = ApiListRequest.builder()
    .category("AI/ML")
    .limit(10)
    .build();

ApiListResponse apis = client.apis().list(request);

// Subscribe to API
SubscriptionCreateRequest subRequest = SubscriptionCreateRequest.builder()
    .apiId("api_12345")
    .plan("pro")
    .build();

Subscription subscription = client.subscriptions().create(subRequest);`}
                </pre>
              </div>

              <div className="bg-white border border-gray-200 rounded-lg p-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-3">Configuration</h3>
                <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`MarketplaceClient client = MarketplaceClient.builder()
    .apiKey("your-api-key")
    .environment(Environment.PRODUCTION)
    .timeout(Duration.ofSeconds(30))
    .retries(3)
    .baseUrl("https://api.marketplace.com/v1")
    .build();`}
                </pre>
              </div>
            </section>
          </div>
        )}

        {/* Other Languages */}
        {['php', 'ruby', 'go', 'csharp', 'swift'].includes(activeLanguage) && (
          <div className="space-y-8">
            <section>
              <h2 className="text-2xl font-bold text-gray-900 mb-4">
                {languages.find(l => l.id === activeLanguage)?.name} SDK
              </h2>
              
              <div className="bg-blue-50 border border-blue-200 rounded-lg p-6 mb-6">
                <div className="flex items-start">
                  <div className="flex-shrink-0">
                    <svg className="h-5 w-5 text-blue-400" viewBox="0 0 20 20" fill="currentColor">
                      <path fillRule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clipRule="evenodd" />
                    </svg>
                  </div>
                  <div className="ml-3">
                    <h3 className="text-sm font-medium text-blue-800">Coming Soon</h3>
                    <div className="mt-1 text-sm text-blue-700">
                      <p>The official {languages.find(l => l.id === activeLanguage)?.name} SDK is currently in development.</p>
                    </div>
                  </div>
                </div>
              </div>

              <div className="bg-white border border-gray-200 rounded-lg p-6 mb-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-3">HTTP API Usage</h3>
                <p className="text-gray-600 mb-4">
                  While we work on the official SDK, you can use our REST API directly with any HTTP client.
                </p>
                
                {activeLanguage === 'php' && (
                  <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`<?php
$ch = curl_init();
curl_setopt($ch, CURLOPT_URL, 'https://api.marketplace.com/v1/apis');
curl_setopt($ch, CURLOPT_HTTPHEADER, [
    'Authorization: Bearer your-api-key',
    'Content-Type: application/json'
]);
curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);

$response = curl_exec($ch);
$data = json_decode($response, true);

curl_close($ch);
?>`}
                  </pre>
                )}

                {activeLanguage === 'ruby' && (
                  <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`require 'net/http'
require 'json'

uri = URI('https://api.marketplace.com/v1/apis')
http = Net::HTTP.new(uri.host, uri.port)
http.use_ssl = true

request = Net::HTTP::Get.new(uri)
request['Authorization'] = 'Bearer your-api-key'
request['Content-Type'] = 'application/json'

response = http.request(request)
data = JSON.parse(response.body)`}
                  </pre>
                )}

                {activeLanguage === 'go' && (
                  <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`package main

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
)

func main() {
    client := &http.Client{}
    req, _ := http.NewRequest("GET", "https://api.marketplace.com/v1/apis", nil)
    req.Header.Add("Authorization", "Bearer your-api-key")
    req.Header.Add("Content-Type", "application/json")
    
    resp, _ := client.Do(req)
    defer resp.Body.Close()
    
    body, _ := io.ReadAll(resp.Body)
    var data map[string]interface{}
    json.Unmarshal(body, &data)
    
    fmt.Println(data)
}`}
                  </pre>
                )}

                {activeLanguage === 'csharp' && (
                  <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`using System;
using System.Net.Http;
using System.Threading.Tasks;
using Newtonsoft.Json;

class Program
{
    private static readonly HttpClient client = new HttpClient();
    
    static async Task Main()
    {
        client.DefaultRequestHeaders.Add("Authorization", "Bearer your-api-key");
        
        HttpResponseMessage response = await client.GetAsync("https://api.marketplace.com/v1/apis");
        string responseBody = await response.Content.ReadAsStringAsync();
        
        dynamic data = JsonConvert.DeserializeObject(responseBody);
        Console.WriteLine(data);
    }
}`}
                  </pre>
                )}

                {activeLanguage === 'swift' && (
                  <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg text-sm">
{`import Foundation

let url = URL(string: "https://api.marketplace.com/v1/apis")!
var request = URLRequest(url: url)
request.setValue("Bearer your-api-key", forHTTPHeaderField: "Authorization")
request.setValue("application/json", forHTTPHeaderField: "Content-Type")

let task = URLSession.shared.dataTask(with: request) { data, response, error in
    if let data = data {
        do {
            let json = try JSONSerialization.jsonObject(with: data, options: [])
            print(json)
        } catch {
            print("JSON parsing error: \\(error)")
        }
    }
}

task.resume()`}
                  </pre>
                )}
              </div>

              <div className="bg-white border border-gray-200 rounded-lg p-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-3">Get Notified</h3>
                <p className="text-gray-600 mb-4">
                  Want to be notified when the {languages.find(l => l.id === activeLanguage)?.name} SDK is released?
                </p>
                <Link 
                  href="/contact?subject=SDK%20Notification"
                  className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 transition-colors"
                >
                  Subscribe to Updates
                </Link>
              </div>
            </section>
          </div>
        )}

        {/* Community Libraries */}
        <section className="mt-16">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Community Libraries</h2>
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <p className="text-gray-600 mb-6">
              While we work on official SDKs for all languages, the community has created several helpful libraries:
            </p>
            
            <div className="grid md:grid-cols-2 gap-6">
              <div className="border border-gray-100 rounded-lg p-4">
                <h3 className="font-semibold text-gray-900 mb-2">marketplace-php-client</h3>
                <p className="text-sm text-gray-600 mb-3">Unofficial PHP client library</p>
                <div className="flex items-center justify-between">
                  <span className="text-xs text-gray-500">by @developer123</span>
                  <a href="#" className="text-blue-600 hover:text-blue-500 text-sm">View on GitHub</a>
                </div>
              </div>

              <div className="border border-gray-100 rounded-lg p-4">
                <h3 className="font-semibold text-gray-900 mb-2">marketplace-ruby-gem</h3>
                <p className="text-sm text-gray-600 mb-3">Ruby gem for Marketplace API</p>
                <div className="flex items-center justify-between">
                  <span className="text-xs text-gray-500">by @rubydev</span>
                  <a href="#" className="text-blue-600 hover:text-blue-500 text-sm">View on GitHub</a>
                </div>
              </div>

              <div className="border border-gray-100 rounded-lg p-4">
                <h3 className="font-semibold text-gray-900 mb-2">go-marketplace</h3>
                <p className="text-sm text-gray-600 mb-3">Go library for API integration</p>
                <div className="flex items-center justify-between">
                  <span className="text-xs text-gray-500">by @gopher</span>
                  <a href="#" className="text-blue-600 hover:text-blue-500 text-sm">View on GitHub</a>
                </div>
              </div>

              <div className="border border-gray-100 rounded-lg p-4">
                <h3 className="font-semibold text-gray-900 mb-2">marketplace-dotnet</h3>
                <p className="text-sm text-gray-600 mb-3">.NET library for C# developers</p>
                <div className="flex items-center justify-between">
                  <span className="text-xs text-gray-500">by @netdev</span>
                  <a href="#" className="text-blue-600 hover:text-blue-500 text-sm">View on GitHub</a>
                </div>
              </div>
            </div>

            <div className="mt-6 p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
              <p className="text-sm text-yellow-800">
                <strong>Note:</strong> Community libraries are maintained by third parties. We don&apos;t guarantee their functionality or provide support for them.
              </p>
            </div>
          </div>
        </section>

        {/* Contributing */}
        <section className="mt-16">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Contributing</h2>
          <div className="bg-white border border-gray-200 rounded-lg p-6">
            <p className="text-gray-600 mb-4">
              Want to contribute to our SDKs or create a community library? We&apos;d love your help!
            </p>
            
            <div className="space-y-4">
              <div>
                <h3 className="font-semibold text-gray-900 mb-2">Guidelines</h3>
                <ul className="list-disc list-inside space-y-1 text-gray-600 text-sm">
                  <li>Follow the language&apos;s standard conventions and best practices</li>
                  <li>Include comprehensive documentation and examples</li>
                  <li>Add unit tests with good coverage</li>
                  <li>Support async/await patterns where applicable</li>
                  <li>Handle errors gracefully with proper exception types</li>
                </ul>
              </div>

              <div className="flex gap-4">
                <Link 
                  href="/contact?subject=SDK%20Contribution"
                  className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 transition-colors"
                >
                  Get in Touch
                </Link>
                <a 
                  href="#"
                  className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 transition-colors"
                >
                  View on GitHub
                </a>
              </div>
            </div>
          </div>
        </section>

        {/* CTA */}
        <div className="mt-16 bg-gradient-to-r from-blue-50 to-indigo-50 rounded-xl p-8 text-center">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Ready to Start Coding?</h2>
          <p className="text-gray-600 mb-6">
            Choose your preferred language and start building with our APIs.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Link 
              href="/auth/signup"
              className="inline-flex items-center px-6 py-3 border border-transparent text-base font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 transition-colors"
            >
              Get Your API Key
            </Link>
            <Link 
              href="/docs/examples"
              className="inline-flex items-center px-6 py-3 border border-gray-300 text-base font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 transition-colors"
            >
              View Examples
            </Link>
          </div>
        </div>
      </div>
    </Layout>
  );
};

export default SDKs;
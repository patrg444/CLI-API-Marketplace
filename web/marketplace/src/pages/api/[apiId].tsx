import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import { useQuery } from 'react-query';
import Layout from '../../components/Layout';
import APIDocumentation from '../../components/APIDocumentation';
import ReviewSection from '../../components/ReviewSection';
import apiService from '../../services/api';
import { API, PricingPlan, APIDocumentation as APIDocType, Subscription } from '../../types/api';
import { useSwaggerInterceptor, useAPIBaseUrl } from '../../hooks/useSwaggerInterceptor';
import { Auth } from 'aws-amplify';

const APIDetails: React.FC = () => {
  const router = useRouter();
  const { apiId } = router.query;
  const [selectedPlan, setSelectedPlan] = useState<PricingPlan | null>(null);
  const [documentation, setDocumentation] = useState<APIDocType | null>(null);
  const [userSubscription, setUserSubscription] = useState<Subscription | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  const { data: api, isLoading, error } = useQuery<API>(
    ['api', apiId],
    () => apiService.getAPI(apiId as string),
    { enabled: !!apiId }
  );

  // Fetch API documentation
  const { data: docData } = useQuery(
    ['api-documentation', apiId],
    () => apiService.getAPIDocumentation(apiId as string),
    { 
      enabled: !!apiId,
      onSuccess: (data) => setDocumentation(data)
    }
  );

  // Check if user is authenticated and fetch their subscriptions
  useEffect(() => {
    const checkAuth = async () => {
      try {
        await Auth.currentAuthenticatedUser();
        setIsAuthenticated(true);
      } catch {
        setIsAuthenticated(false);
      }
    };
    checkAuth();
  }, []);

  // Fetch user's subscriptions if authenticated
  const { data: subscriptions } = useQuery(
    ['my-subscriptions'],
    () => apiService.listMySubscriptions(),
    {
      enabled: isAuthenticated,
      onSuccess: (subs) => {
        // Find subscription for this API
        const apiSub = subs.find(sub => sub.api_id === apiId);
        setUserSubscription(apiSub || null);
      }
    }
  );

  // Get API key and base URL for Swagger
  const { apiKey } = useSwaggerInterceptor({
    subscriptionId: userSubscription?.id,
    apiId: apiId as string
  });
  const apiBaseUrl = useAPIBaseUrl(apiId as string);

  const handleSubscribe = () => {
    if (!selectedPlan) return;
    
    // Navigate to subscription flow with selected plan
    router.push({
      pathname: `/subscribe/${apiId}`,
      query: { plan: selectedPlan.id }
    });
  };

  if (isLoading) {
    return (
      <Layout>
        <div className="flex justify-center items-center min-h-screen">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600"></div>
        </div>
      </Layout>
    );
  }

  if (error || !api) {
    return (
      <Layout>
        <div className="text-center py-12">
          <h2 className="text-2xl font-bold text-gray-900">API not found</h2>
          <button
            onClick={() => router.push('/')}
            className="mt-4 text-indigo-600 hover:text-indigo-500"
          >
            Back to marketplace
          </button>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="bg-white">
        {/* Header */}
        <div className="bg-gray-50 border-b">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
            <div className="md:flex md:items-center md:justify-between">
              <div className="flex-1 min-w-0">
                <div className="flex items-center">
                  {api.icon_url && (
                    <img 
                      src={api.icon_url} 
                      alt={`${api.name} icon`}
                      className="h-16 w-16 rounded-lg mr-4"
                    />
                  )}
                  <div>
                    <h1 className="text-3xl font-bold text-gray-900">{api.name}</h1>
                    <p className="mt-1 text-sm text-gray-500">
                      by Creator • {api.category}
                    </p>
                  </div>
                </div>
              </div>
              <div className="mt-4 flex md:mt-0 md:ml-4">
                <span className="inline-flex items-center px-3 py-0.5 rounded-full text-sm font-medium bg-green-100 text-green-800">
                  {api.is_published ? 'Active' : 'Inactive'}
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Content */}
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="lg:grid lg:grid-cols-3 lg:gap-8">
            {/* Main Content */}
            <div className="lg:col-span-2">
              {/* Description */}
              <div className="bg-white overflow-hidden">
                <div className="px-4 py-5 sm:p-6">
                  <h2 className="text-lg font-medium text-gray-900 mb-4">Description</h2>
                  <div className="prose max-w-none text-gray-500">
                    {api.description}
                  </div>
                </div>
              </div>

              {/* Tags */}
              {api.tags && api.tags.length > 0 && (
                <div className="mt-6">
                  <div className="flex flex-wrap gap-2">
                    {api.tags.map((tag, index) => (
                      <span
                        key={index}
                        className="inline-flex items-center px-3 py-0.5 rounded-full text-sm font-medium bg-indigo-100 text-indigo-800"
                      >
                        {tag}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              {/* API Documentation */}
              <div className="mt-8">
                <h2 className="text-lg font-medium text-gray-900 mb-4">API Documentation</h2>
                <APIDocumentation
                  documentation={documentation}
                  apiKey={apiKey}
                  apiBaseUrl={apiBaseUrl}
                  isSubscribed={!!userSubscription && userSubscription.status === 'active'}
                />
              </div>

              {/* Reviews Section */}
              <ReviewSection 
                apiId={apiId as string} 
                canReview={isAuthenticated && !!userSubscription && userSubscription.status === 'active'} 
              />
            </div>

            {/* Sidebar - Pricing & Subscribe */}
            <div className="mt-8 lg:mt-0">
              <div className="bg-white shadow rounded-lg">
                <div className="px-4 py-5 sm:p-6">
                  <h3 className="text-lg font-medium text-gray-900 mb-4">Pricing Plans</h3>
                  
                  {api.pricing_plans && api.pricing_plans.length > 0 ? (
                    <div className="space-y-4">
                      {api.pricing_plans.map((plan: PricingPlan) => (
                        <div
                          key={plan.id}
                          className={`border rounded-lg p-4 cursor-pointer transition-colors ${
                            selectedPlan?.id === plan.id
                              ? 'border-indigo-500 bg-indigo-50'
                              : 'border-gray-200 hover:border-gray-300'
                          }`}
                          onClick={() => setSelectedPlan(plan)}
                        >
                          <div className="flex items-center justify-between">
                            <div>
                              <h4 className="text-sm font-medium text-gray-900">{plan.name}</h4>
                              <p className="mt-1 text-sm text-gray-500">Type: {plan.type}</p>
                            </div>
                            <input
                              type="radio"
                              checked={selectedPlan?.id === plan.id}
                              onChange={() => setSelectedPlan(plan)}
                              className="h-4 w-4 text-indigo-600"
                            />
                          </div>
                          
                          <div className="mt-3">
                            <p className="text-2xl font-bold text-gray-900">
                              {plan.type === 'free' ? (
                                'Free'
                              ) : plan.type === 'pay_per_use' ? (
                                <>
                                  ${plan.price_per_call || 0}
                                  <span className="text-sm font-normal text-gray-500">/call</span>
                                </>
                              ) : (
                                <>
                                  ${plan.monthly_price || 0}
                                  <span className="text-sm font-normal text-gray-500">/month</span>
                                </>
                              )}
                            </p>
                            <ul className="mt-2 space-y-1">
                              {plan.call_limit && (
                                <li className="text-sm text-gray-500">
                                  {plan.call_limit.toLocaleString()} calls/month
                                </li>
                              )}
                              {plan.rate_limit_per_minute && (
                                <li className="text-sm text-gray-500">
                                  {plan.rate_limit_per_minute} requests/minute
                                </li>
                              )}
                              {plan.rate_limit_per_day && (
                                <li className="text-sm text-gray-500">
                                  {plan.rate_limit_per_day.toLocaleString()} requests/day
                                </li>
                              )}
                              {plan.features && Object.entries(plan.features).map(([key, value]) => (
                                <li key={key} className="text-sm text-gray-500">
                                  ✓ {key}: {String(value)}
                                </li>
                              ))}
                            </ul>
                          </div>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <p className="text-gray-500">No pricing plans available</p>
                  )}

                  <button
                    onClick={handleSubscribe}
                    disabled={!selectedPlan || (userSubscription?.status === 'active')}
                    className="mt-6 w-full bg-indigo-600 text-white py-2 px-4 rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {userSubscription?.status === 'active' 
                      ? 'Already Subscribed' 
                      : `Subscribe to ${selectedPlan?.name || 'Plan'}`}
                  </button>
                </div>
              </div>

              {/* API Stats */}
              <div className="mt-6 bg-white shadow rounded-lg">
                <div className="px-4 py-5 sm:p-6">
                  <h3 className="text-lg font-medium text-gray-900 mb-4">API Statistics</h3>
                  <dl className="space-y-3">
                    <div className="flex justify-between">
                      <dt className="text-sm text-gray-500">Active Subscriptions</dt>
                      <dd className="text-sm font-medium text-gray-900">
                        {api.total_subscriptions?.toLocaleString() || '0'}
                      </dd>
                    </div>
                    <div className="flex justify-between">
                      <dt className="text-sm text-gray-500">Average Rating</dt>
                      <dd className="text-sm font-medium text-gray-900">
                        {api.average_rating ? `${api.average_rating.toFixed(1)} / 5` : 'No ratings yet'}
                      </dd>
                    </div>
                    <div className="flex justify-between">
                      <dt className="text-sm text-gray-500">Total Reviews</dt>
                      <dd className="text-sm font-medium text-gray-900">
                        {api.total_reviews || '0'}
                      </dd>
                    </div>
                  </dl>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
};

export default APIDetails;

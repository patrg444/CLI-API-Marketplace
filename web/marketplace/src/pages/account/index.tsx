import React, { useState, useEffect } from 'react'
import { useRouter } from 'next/router'
import Layout from '@/components/Layout'
import { 
  UserCircleIcon, 
  CreditCardIcon, 
  KeyIcon, 
  ChartBarIcon,
  CodeBracketIcon,
  CurrencyDollarIcon,
  Cog6ToothIcon,
  ArrowPathIcon,
  PlusIcon,
  DocumentTextIcon,
  BellIcon
} from '@heroicons/react/24/outline'

interface UserData {
  id: string
  name: string
  email: string
  role: 'consumer' | 'creator' | 'both'
  isCreator: boolean
  createdAt: string
}

const AccountDashboard: React.FC = () => {
  const router = useRouter()
  const [activeView, setActiveView] = useState<'consumer' | 'creator'>('consumer')
  const [activeTab, setActiveTab] = useState('overview')
  const [user, setUser] = useState<UserData | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    // Check authentication
    const checkAuth = () => {
      try {
        const storedUser = localStorage.getItem('mockUser')
        if (storedUser) {
          const userData = JSON.parse(storedUser)
          // Add creator capabilities
          setUser({
            ...userData,
            role: userData.isCreator ? 'both' : 'consumer',
            isCreator: userData.isCreator || false
          })
        } else {
          router.push('/auth/login')
        }
      } catch {
        router.push('/auth/login')
      } finally {
        setLoading(false)
      }
    }

    checkAuth()
  }, [router])

  const handleBecomeCreator = () => {
    if (user) {
      const updatedUser = { ...user, isCreator: true, role: 'both' as const }
      setUser(updatedUser)
      localStorage.setItem('mockUser', JSON.stringify(updatedUser))
      setActiveView('creator')
    }
  }

  const consumerTabs = [
    { id: 'overview', label: 'Overview', icon: ChartBarIcon },
    { id: 'subscriptions', label: 'My Subscriptions', icon: CreditCardIcon },
    { id: 'api-keys', label: 'API Keys', icon: KeyIcon },
    { id: 'usage', label: 'Usage & Billing', icon: ChartBarIcon },
    { id: 'settings', label: 'Settings', icon: Cog6ToothIcon }
  ]

  const creatorTabs = [
    { id: 'overview', label: 'Overview', icon: ChartBarIcon },
    { id: 'my-apis', label: 'My APIs', icon: CodeBracketIcon },
    { id: 'analytics', label: 'Analytics', icon: ChartBarIcon },
    { id: 'earnings', label: 'Earnings', icon: CurrencyDollarIcon },
    { id: 'payouts', label: 'Payouts', icon: CreditCardIcon },
    { id: 'settings', label: 'Settings', icon: Cog6ToothIcon }
  ]

  const currentTabs = activeView === 'consumer' ? consumerTabs : creatorTabs

  if (loading) {
    return (
      <Layout>
        <div className="flex justify-center items-center min-h-screen">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
        </div>
      </Layout>
    )
  }

  if (!user) {
    return null
  }

  return (
    <Layout>
      <div className="min-h-screen bg-gray-900">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          {/* Header with Role Switcher */}
          <div className="flex justify-between items-center mb-8">
            <div>
              <h1 className="text-3xl font-bold text-white">My Account</h1>
              <p className="text-gray-400 mt-1">Welcome back, {user.name}</p>
            </div>
            
            {user.isCreator ? (
              <div className="flex items-center space-x-4">
                <span className="text-gray-400 text-sm">View as:</span>
                <div className="flex bg-gray-800 rounded-lg p-1">
                  <button
                    onClick={() => setActiveView('consumer')}
                    className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                      activeView === 'consumer' 
                        ? 'bg-blue-600 text-white' 
                        : 'text-gray-400 hover:text-white'
                    }`}
                  >
                    Consumer
                  </button>
                  <button
                    onClick={() => setActiveView('creator')}
                    className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${
                      activeView === 'creator' 
                        ? 'bg-blue-600 text-white' 
                        : 'text-gray-400 hover:text-white'
                    }`}
                  >
                    Creator
                  </button>
                </div>
              </div>
            ) : (
              <button
                onClick={handleBecomeCreator}
                className="flex items-center px-4 py-2 bg-gradient-to-r from-purple-600 to-blue-600 text-white rounded-lg hover:from-purple-700 hover:to-blue-700 transition-colors"
              >
                <PlusIcon className="h-5 w-5 mr-2" />
                Become a Creator
              </button>
            )}
          </div>

          <div className="flex gap-8">
            {/* Sidebar Navigation */}
            <div className="w-64 flex-shrink-0">
              <nav className="space-y-1">
                {currentTabs.map((tab) => {
                  const Icon = tab.icon
                  return (
                    <button
                      key={tab.id}
                      onClick={() => setActiveTab(tab.id)}
                      className={`w-full flex items-center px-4 py-3 text-sm font-medium rounded-lg transition-colors ${
                        activeTab === tab.id
                          ? 'bg-gray-800 text-white'
                          : 'text-gray-400 hover:text-white hover:bg-gray-800/50'
                      }`}
                    >
                      <Icon className="h-5 w-5 mr-3" />
                      {tab.label}
                    </button>
                  )
                })}
              </nav>

              {/* Quick Actions */}
              <div className="mt-8 p-4 bg-gray-800 rounded-lg">
                <h3 className="text-sm font-medium text-white mb-3">Quick Actions</h3>
                {activeView === 'consumer' ? (
                  <div className="space-y-2">
                    <button 
                      onClick={() => router.push('/')}
                      className="w-full text-left text-sm text-gray-400 hover:text-white"
                    >
                      → Browse APIs
                    </button>
                    <button className="w-full text-left text-sm text-gray-400 hover:text-white">
                      → View Documentation
                    </button>
                    <button className="w-full text-left text-sm text-gray-400 hover:text-white">
                      → Get Support
                    </button>
                  </div>
                ) : (
                  <div className="space-y-2">
                    <button className="w-full text-left text-sm text-gray-400 hover:text-white">
                      → Create New API
                    </button>
                    <button className="w-full text-left text-sm text-gray-400 hover:text-white">
                      → View Guidelines
                    </button>
                    <button className="w-full text-left text-sm text-gray-400 hover:text-white">
                      → API Analytics
                    </button>
                  </div>
                )}
              </div>
            </div>

            {/* Main Content Area */}
            <div className="flex-1">
              {/* Consumer Views */}
              {activeView === 'consumer' && (
                <>
                  {activeTab === 'overview' && (
                    <div className="space-y-6">
                      {/* Stats Grid */}
                      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                        <div className="bg-gray-800 rounded-lg p-6">
                          <h3 className="text-sm font-medium text-gray-400">Active Subscriptions</h3>
                          <p className="text-3xl font-bold text-white mt-2">3</p>
                          <p className="text-sm text-gray-500 mt-1">2 trials ending soon</p>
                        </div>
                        <div className="bg-gray-800 rounded-lg p-6">
                          <h3 className="text-sm font-medium text-gray-400">API Calls Today</h3>
                          <p className="text-3xl font-bold text-white mt-2">1,247</p>
                          <p className="text-sm text-green-500 mt-1">↑ 12% from yesterday</p>
                        </div>
                        <div className="bg-gray-800 rounded-lg p-6">
                          <h3 className="text-sm font-medium text-gray-400">Current Month Cost</h3>
                          <p className="text-3xl font-bold text-white mt-2">$67.43</p>
                          <p className="text-sm text-gray-500 mt-1">$132.57 remaining budget</p>
                        </div>
                      </div>

                      {/* Recent Activity */}
                      <div className="bg-gray-800 rounded-lg p-6">
                        <h2 className="text-lg font-medium text-white mb-4">Recent Activity</h2>
                        <div className="space-y-4">
                          <div className="flex items-center justify-between py-3 border-b border-gray-700">
                            <div>
                              <p className="text-white">Subscribed to Weather API</p>
                              <p className="text-sm text-gray-400">2 hours ago</p>
                            </div>
                            <span className="text-sm text-green-500">Active</span>
                          </div>
                          <div className="flex items-center justify-between py-3 border-b border-gray-700">
                            <div>
                              <p className="text-white">API Key generated for Translation API</p>
                              <p className="text-sm text-gray-400">1 day ago</p>
                            </div>
                            <span className="text-sm text-blue-500">New Key</span>
                          </div>
                          <div className="flex items-center justify-between py-3">
                            <div>
                              <p className="text-white">Usage limit alert - Stock Market API</p>
                              <p className="text-sm text-gray-400">3 days ago</p>
                            </div>
                            <span className="text-sm text-yellow-500">Warning</span>
                          </div>
                        </div>
                      </div>
                    </div>
                  )}

                  {activeTab === 'subscriptions' && (
                    <div className="bg-gray-800 rounded-lg p-6">
                      <h2 className="text-lg font-medium text-white mb-4">My Subscriptions</h2>
                      <div className="space-y-4">
                        {['Weather API', 'Translation API', 'Stock Market Data'].map((api, index) => (
                          <div key={index} className="flex items-center justify-between p-4 bg-gray-700 rounded-lg">
                            <div>
                              <h3 className="font-medium text-white">{api}</h3>
                              <p className="text-sm text-gray-400">Pro Plan • $49/month</p>
                            </div>
                            <div className="flex items-center space-x-3">
                              <span className="text-sm text-green-500">Active</span>
                              <button className="text-sm text-gray-400 hover:text-white">Manage</button>
                            </div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </>
              )}

              {/* Creator Views */}
              {activeView === 'creator' && (
                <>
                  {activeTab === 'overview' && (
                    <div className="space-y-6">
                      {/* Creator Stats */}
                      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
                        <div className="bg-gray-800 rounded-lg p-6">
                          <h3 className="text-sm font-medium text-gray-400">Published APIs</h3>
                          <p className="text-3xl font-bold text-white mt-2">2</p>
                          <p className="text-sm text-gray-500 mt-1">1 pending review</p>
                        </div>
                        <div className="bg-gray-800 rounded-lg p-6">
                          <h3 className="text-sm font-medium text-gray-400">Total Subscribers</h3>
                          <p className="text-3xl font-bold text-white mt-2">156</p>
                          <p className="text-sm text-green-500 mt-1">↑ 23 this month</p>
                        </div>
                        <div className="bg-gray-800 rounded-lg p-6">
                          <h3 className="text-sm font-medium text-gray-400">Monthly Revenue</h3>
                          <p className="text-3xl font-bold text-white mt-2">$3,450</p>
                          <p className="text-sm text-green-500 mt-1">↑ 18% from last month</p>
                        </div>
                        <div className="bg-gray-800 rounded-lg p-6">
                          <h3 className="text-sm font-medium text-gray-400">Avg Rating</h3>
                          <p className="text-3xl font-bold text-white mt-2">4.8</p>
                          <p className="text-sm text-gray-500 mt-1">from 89 reviews</p>
                        </div>
                      </div>

                      {/* API Performance */}
                      <div className="bg-gray-800 rounded-lg p-6">
                        <h2 className="text-lg font-medium text-white mb-4">API Performance</h2>
                        <div className="space-y-4">
                          <div className="p-4 bg-gray-700 rounded-lg">
                            <div className="flex justify-between items-start mb-2">
                              <div>
                                <h3 className="font-medium text-white">Weather Forecast API</h3>
                                <p className="text-sm text-gray-400">Published 3 months ago</p>
                              </div>
                              <span className="px-2 py-1 bg-green-500/20 text-green-500 text-xs rounded">Active</span>
                            </div>
                            <div className="grid grid-cols-3 gap-4 mt-4 text-sm">
                              <div>
                                <p className="text-gray-400">Subscribers</p>
                                <p className="text-white font-medium">89</p>
                              </div>
                              <div>
                                <p className="text-gray-400">Monthly Revenue</p>
                                <p className="text-white font-medium">$2,670</p>
                              </div>
                              <div>
                                <p className="text-gray-400">Success Rate</p>
                                <p className="text-white font-medium">99.8%</p>
                              </div>
                            </div>
                          </div>
                          <div className="p-4 bg-gray-700 rounded-lg">
                            <div className="flex justify-between items-start mb-2">
                              <div>
                                <h3 className="font-medium text-white">Currency Exchange API</h3>
                                <p className="text-sm text-gray-400">Published 1 month ago</p>
                              </div>
                              <span className="px-2 py-1 bg-green-500/20 text-green-500 text-xs rounded">Active</span>
                            </div>
                            <div className="grid grid-cols-3 gap-4 mt-4 text-sm">
                              <div>
                                <p className="text-gray-400">Subscribers</p>
                                <p className="text-white font-medium">67</p>
                              </div>
                              <div>
                                <p className="text-gray-400">Monthly Revenue</p>
                                <p className="text-white font-medium">$780</p>
                              </div>
                              <div>
                                <p className="text-gray-400">Success Rate</p>
                                <p className="text-white font-medium">99.5%</p>
                              </div>
                            </div>
                          </div>
                        </div>
                      </div>
                    </div>
                  )}

                  {activeTab === 'my-apis' && (
                    <div className="space-y-6">
                      <div className="flex justify-between items-center">
                        <h2 className="text-lg font-medium text-white">My APIs</h2>
                        <button className="flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
                          <PlusIcon className="h-5 w-5 mr-2" />
                          Create New API
                        </button>
                      </div>
                      <div className="bg-gray-800 rounded-lg p-6">
                        <p className="text-gray-400">Your published APIs will appear here.</p>
                      </div>
                    </div>
                  )}

                  {activeTab === 'earnings' && (
                    <div className="bg-gray-800 rounded-lg p-6">
                      <h2 className="text-lg font-medium text-white mb-4">Earnings Overview</h2>
                      <div className="space-y-6">
                        <div className="grid grid-cols-2 gap-6">
                          <div>
                            <p className="text-sm text-gray-400">Current Balance</p>
                            <p className="text-3xl font-bold text-white mt-1">$1,234.56</p>
                          </div>
                          <div>
                            <p className="text-sm text-gray-400">Lifetime Earnings</p>
                            <p className="text-3xl font-bold text-white mt-1">$12,450.00</p>
                          </div>
                        </div>
                        <div className="pt-4 border-t border-gray-700">
                          <p className="text-sm text-gray-400 mb-2">Next payout scheduled for:</p>
                          <p className="text-white">January 1, 2024 • $1,234.56</p>
                        </div>
                      </div>
                    </div>
                  )}
                </>
              )}

              {/* Settings (same for both views) */}
              {activeTab === 'settings' && (
                <div className="space-y-6">
                  <div className="bg-gray-800 rounded-lg p-6">
                    <h2 className="text-lg font-medium text-white mb-4">Account Settings</h2>
                    <div className="space-y-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-400 mb-2">Name</label>
                        <input 
                          type="text" 
                          value={user.name} 
                          className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg focus:ring-2 focus:ring-blue-500"
                          readOnly
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-400 mb-2">Email</label>
                        <input 
                          type="email" 
                          value={user.email} 
                          className="w-full px-3 py-2 bg-gray-700 text-white rounded-lg focus:ring-2 focus:ring-blue-500"
                          readOnly
                        />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-400 mb-2">Account Type</label>
                        <p className="text-white">{user.isCreator ? 'Consumer & Creator' : 'Consumer'}</p>
                      </div>
                    </div>
                  </div>

                  <div className="bg-gray-800 rounded-lg p-6">
                    <h2 className="text-lg font-medium text-white mb-4">Notifications</h2>
                    <div className="space-y-3">
                      <label className="flex items-center">
                        <input type="checkbox" className="mr-3 rounded bg-gray-700 border-gray-600" defaultChecked />
                        <span className="text-white">Email notifications for API updates</span>
                      </label>
                      <label className="flex items-center">
                        <input type="checkbox" className="mr-3 rounded bg-gray-700 border-gray-600" defaultChecked />
                        <span className="text-white">Usage alerts</span>
                      </label>
                      {user.isCreator && (
                        <>
                          <label className="flex items-center">
                            <input type="checkbox" className="mr-3 rounded bg-gray-700 border-gray-600" defaultChecked />
                            <span className="text-white">New subscriber notifications</span>
                          </label>
                          <label className="flex items-center">
                            <input type="checkbox" className="mr-3 rounded bg-gray-700 border-gray-600" defaultChecked />
                            <span className="text-white">Payout reminders</span>
                          </label>
                        </>
                      )}
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </Layout>
  )
}

export default AccountDashboard
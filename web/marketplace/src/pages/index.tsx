import React, { useState, useEffect } from 'react'
import Layout from '@/components/Layout'
import APICard from '@/components/APICard'
import SearchBar from '@/components/SearchBar'
import { Button } from '@/components/ui/Button'
import { SkeletonCard } from '@/components/ui/Skeleton'
import { useQuery } from 'react-query'
import apiService from '@/services/api'
import { marketplaceAPI } from '@/services/marketplace-api'
import { useRouter } from 'next/router'
import Link from 'next/link'

export default function MarketplacePage() {
  const router = useRouter()
  const [searchParams, setSearchParams] = useState<any>({})
  const [page, setPage] = useState(1)
  const [activeTab, setActiveTab] = useState<'browse' | 'search'>('browse')
  
  // Fetch featured APIs IDs for badge display
  const { data: featuredData } = useQuery<any>(
    ['featured-apis'],
    async () => {
      const featuredApis = await marketplaceAPI.getFeaturedAPIs()
      return featuredApis.slice(0, 3).map((api: any) => api.id)
    }
  )
  
  // Fetch platform statistics
  const { data: stats } = useQuery<any>(
    ['platform-stats'],
    async () => {
      // Get all APIs to calculate stats
      const result = await marketplaceAPI.getAPIs({ limit: 100 })
      const totalCalls = result.apis.reduce((sum, api) => sum + (api.calls || 0), 0)
      const totalReviews = result.apis.reduce((sum, api) => sum + (api.reviews || 0), 0)
      const avgRating = result.apis.reduce((sum, api) => sum + (api.rating || 0), 0) / result.apis.length
      return {
        totalAPIs: result.total,
        totalSubscriptions: totalCalls,
        totalReviews,
        averageRating: avgRating.toFixed(1)
      }
    }
  )

  // Parse URL query parameters on mount
  useEffect(() => {
    const params = new URLSearchParams(window.location.search)
    const parsedParams: any = {}
    
    params.forEach((value, key) => {
      if (key === 'tags') {
        parsedParams[key] = value.split(',')
      } else if (key === 'min_rating' || key === 'page') {
        parsedParams[key] = parseInt(value)
      } else if (key === 'has_free_tier') {
        parsedParams[key] = value === 'true'
      } else {
        parsedParams[key] = value
      }
    })
    
    if (Object.keys(parsedParams).length > 0) {
      setSearchParams(parsedParams)
      setActiveTab('search')
      if (parsedParams.page) {
        setPage(parsedParams.page)
      }
    }
  }, [])

  const { data, isLoading, error } = useQuery<any>(
    ['apis', searchParams, page, activeTab],
    async () => {
      if (activeTab === 'search' && Object.keys(searchParams).length > 0) {
        // Use marketplace API for search
        const result = await marketplaceAPI.getAPIs({
          search: searchParams.q,
          category: searchParams.category,
          maxPrice: searchParams.maxPrice,
          sort: searchParams.sort_by,
          page,
          limit: 12
        })
        // Transform to match expected format
        return {
          apis: result.apis.map(api => ({
            ...api,
            total_subscriptions: api.calls,
            average_rating: api.rating,
            total_reviews: api.reviews,
            pricing_plans: [{
              type: api.pricing.type === 'freemium' ? 'free' : 'subscription',
              monthly_price: api.pricing.monthlyPrice || 0,
              call_limit: api.pricing.freeCalls || 0
            }]
          })),
          total: result.total,
          facets: {
            categories: {},
            tags: {},
            price_ranges: {},
            ratings: {}
          }
        }
      } else {
        // Use marketplace API for browse
        const result = await marketplaceAPI.getAPIs({
          category: searchParams.category,
          page,
          limit: 12
        })
        // Transform to match expected format
        return {
          apis: result.apis.map(api => ({
            ...api,
            total_subscriptions: api.calls,
            average_rating: api.rating,
            total_reviews: api.reviews,
            pricing_plans: [{
              type: api.pricing.type === 'freemium' ? 'free' : 'subscription',
              monthly_price: api.pricing.monthlyPrice || 0,
              call_limit: api.pricing.freeCalls || 0
            }]
          })),
          total: result.total
        }
      }
    },
    {
      keepPreviousData: true,
    }
  )

  const handleSearch = (params: any) => {
    setSearchParams(params)
    setPage(1)
    setActiveTab('search')
    
    // Update URL with search params
    const queryString = new URLSearchParams(
      Object.entries(params).reduce((acc, [key, value]) => {
        if (value !== undefined) {
          acc[key] = Array.isArray(value) ? value.join(',') : String(value)
        }
        return acc
      }, {} as Record<string, string>)
    ).toString()
    
    // Use replace to avoid adding to history and prevent scroll jump
    router.replace(`/?${queryString}`, undefined, { scroll: false })
  }

  const handleCategoryClick = (category: string) => {
    if (category === '') {
      setSearchParams({})
      setActiveTab('browse')
      // Update URL without scrolling
      router.replace('/', undefined, { scroll: false })
    } else {
      setSearchParams({ category })
      setActiveTab('browse')
      // Update URL without scrolling
      router.replace(`/?category=${encodeURIComponent(category)}`, undefined, { scroll: false })
    }
    setPage(1)
  }

  return (
    <Layout>
      {/* Hero Section */}
      <div className="relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-primary-600 via-primary-700 to-indigo-800 opacity-90" />
        <div className="absolute inset-0 opacity-20" style={{ 
          backgroundImage: "url(\"data:image/svg+xml,%3Csvg width='60' height='60' viewBox='0 0 60 60' xmlns='http://www.w3.org/2000/svg'%3E%3Cg fill='none' fill-rule='evenodd'%3E%3Cg fill='%239C92AC' fill-opacity='0.05'%3E%3Cpath d='M36 34v-4h-2v4h-4v2h4v4h2v-4h4v-2h-4zm0-30V0h-2v4h-4v2h4v4h2V6h4V4h-4zM6 34v-4H4v4H0v2h4v4h2v-4h4v-2H6zM6 4V0H4v4H0v2h4v4h2V6h4V4H6z'/%3E%3C/g%3E%3C/g%3E%3C/svg%3E\")" 
        }} />
        
        <div className="relative py-16 sm:py-20 md:py-24 px-4 sm:px-6 lg:px-8">
          <div className="max-w-7xl mx-auto">
            <div className="text-center">
              <h1 className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-bold text-white mb-4 sm:mb-6 leading-tight">
                API Marketplace
              </h1>
              <p className="text-lg sm:text-xl md:text-2xl text-gray-100 mb-6 sm:mb-8 max-w-3xl mx-auto leading-relaxed px-4 sm:px-0">
                Discover, integrate, and scale with powerful APIs. Build amazing applications with our curated collection of modern APIs.
              </p>
              <div className="flex flex-col sm:flex-row gap-3 sm:gap-4 justify-center mb-8 sm:mb-12 px-4 sm:px-0">
                <Button 
                  size="lg" 
                  variant="secondary"
                  className="!bg-white !text-gray-900 hover:!bg-primary-50 hover:!text-primary-800 px-6 sm:px-8 py-3 sm:py-4 text-base sm:text-lg touch-manipulation w-full sm:w-auto font-medium border border-gray-200"
                  onClick={() => {
                    const searchSection = document.getElementById('search-section')
                    searchSection?.scrollIntoView({ behavior: 'smooth' })
                  }}
                  data-testid="hero-search-cta"
                >
                  Browse APIs
                </Button>
                <Button 
                  size="lg" 
                  variant="secondary" 
                  className="!text-white !border-2 !border-white hover:!bg-white hover:!text-primary-700 px-6 sm:px-8 py-3 sm:py-4 text-base sm:text-lg touch-manipulation w-full sm:w-auto font-medium transition-all duration-200 !bg-transparent"
                  onClick={() => router.push('/docs')}
                  data-testid="hero-learn-more"
                >
                  View Documentation
                </Button>
              </div>

              {/* Statistics */}
              {stats && (
                <div className="grid grid-cols-2 md:grid-cols-4 gap-4 sm:gap-6 max-w-4xl mx-auto px-4 sm:px-0">
                  <div className="text-center">
                    <div className="text-2xl sm:text-3xl font-bold text-white">{stats.totalAPIs}+</div>
                    <div className="text-xs sm:text-sm text-gray-200 mt-1">Available APIs</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl sm:text-3xl font-bold text-white">{stats.totalSubscriptions.toLocaleString()}</div>
                    <div className="text-xs sm:text-sm text-gray-200 mt-1">Active Subscriptions</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl sm:text-3xl font-bold text-white">{stats.totalReviews}+</div>
                    <div className="text-xs sm:text-sm text-gray-200 mt-1">Customer Reviews</div>
                  </div>
                  <div className="text-center">
                    <div className="text-2xl sm:text-3xl font-bold text-white">{stats.averageRating}</div>
                    <div className="text-xs sm:text-sm text-gray-200 mt-1">Average Rating</div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Search & Filter Section - Now prominently placed */}
      <div id="search-section" className="bg-white py-12 px-4 sm:px-6 lg:px-8">
        <div className="max-w-7xl mx-auto">
          <div className="text-center mb-8">
            <h2 className="text-3xl font-bold text-gray-900 mb-4">Find the Perfect API</h2>
            <p className="text-lg text-gray-600 mb-6">Search through our comprehensive collection of APIs</p>
            
            {/* Popular Tags */}
            <div className="mb-6">
              <div className="flex flex-wrap justify-center gap-2 px-4 sm:px-0">
                {['stripe', 'payments', 'weather', 'ai', 'ml', 'data', 'security', 'forecast'].map((tag) => (
                  <button
                    key={tag}
                    onClick={() => {
                      console.log('Tag clicked:', tag);
                      handleSearch({ tags: [tag] });
                    }}
                    className="px-3 sm:px-4 py-2 text-sm bg-primary-50 hover:bg-primary-100 text-primary-700 rounded-full transition-all duration-200 hover:shadow-sm font-medium touch-manipulation hover:scale-105"
                    data-testid={`tag-filter-${tag}`}
                  >
                    #{tag}
                  </button>
                ))}
              </div>
            </div>
          </div>

          {/* Search Bar */}
          <div className="max-w-4xl mx-auto mb-8">
            <SearchBar onSearch={handleSearch} showFilters={true} />
          </div>
        </div>
      </div>

      {/* Main Content Area */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        {/* Search Results Summary */}
        {activeTab === 'search' && data && (
          <div className="mb-8 bg-white p-6 rounded-xl shadow-sm border border-gray-100">
            <h2 className="text-xl font-semibold text-gray-900">
              {data.total} results found
              {searchParams.q && ` for "${searchParams.q}"`}
            </h2>
            
            {/* Facets */}
            {data.facets && (
              <div className="mt-4 flex flex-wrap gap-4">
                {data.facets.categories && Object.keys(data.facets.categories).length > 0 && (
                  <div className="flex flex-wrap gap-2">
                    <span className="text-sm text-gray-600">Categories:</span>
                    {Object.entries(data.facets.categories).slice(0, 5).map(([cat, count]) => (
                      <button
                        key={cat}
                        onClick={() => handleSearch({ ...searchParams, category: cat })}
                        className="text-sm px-3 py-1.5 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
                        data-testid="category-facet"
                      >
                        {cat} (<span data-testid="facet-count">{count as number}</span>)
                      </button>
                    ))}
                  </div>
                )}
                
                {data.facets.tags && Object.keys(data.facets.tags).length > 0 && (
                  <div className="flex flex-wrap gap-2">
                    <span className="text-sm text-gray-600">Tags:</span>
                    {Object.entries(data.facets.tags).slice(0, 5).map(([tag, count]) => (
                      <button
                        key={tag}
                        onClick={() => handleSearch({ ...searchParams, tags: [tag] })}
                        className="text-sm px-3 py-1.5 bg-primary-100 hover:bg-primary-200 text-primary-700 rounded-lg transition-colors"
                        data-testid={`tag-filter-${tag}`}
                      >
                        {tag} ({count as number})
                      </button>
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>
        )}

        <div className="flex flex-col lg:flex-row gap-6 lg:gap-8">
          {/* Sidebar for Browse Mode */}
          {activeTab === 'browse' && (
            <aside className="lg:w-64 flex-shrink-0">
              <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-4 sm:p-6">
                <h3 className="text-base sm:text-lg font-semibold text-gray-900 mb-4 sm:mb-6">Browse by Category</h3>
                <div className="space-y-1 sm:space-y-2">
                  <button
                    onClick={() => handleCategoryClick('')}
                    className={`w-full text-left px-3 sm:px-4 py-2 sm:py-2.5 rounded-lg text-sm transition-all duration-200 touch-manipulation ${
                      !searchParams.category
                        ? 'bg-primary-100 text-primary-700 font-medium shadow-sm'
                        : 'text-gray-700 hover:bg-gray-50 hover:text-gray-900'
                    }`}
                    data-testid="category-filter-all"
                  >
                    All Categories
                  </button>
                  {['AI/ML', 'Data', 'Finance', 'Weather'].map((category) => (
                    <button
                      key={category}
                      onClick={() => handleCategoryClick(category)}
                      className={`w-full text-left px-3 sm:px-4 py-2 sm:py-2.5 rounded-lg text-sm transition-all duration-200 touch-manipulation ${
                        searchParams.category === category
                          ? 'bg-primary-100 text-primary-700 font-medium shadow-sm'
                          : 'text-gray-700 hover:bg-gray-50 hover:text-gray-900'
                      }`}
                      data-testid={`category-filter-${category.toLowerCase().replace(/[^a-z0-9]/g, '-')}`}
                    >
                      {category}
                    </button>
                  ))}
                </div>
              </div>
            </aside>
          )}

          {/* API Grid */}
          <main className="flex-1">
            {isLoading ? (
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-2 xl:grid-cols-3 gap-4 sm:gap-6">
                {[...Array(6)].map((_, i) => (
                  <SkeletonCard key={i} showImage={false} />
                ))}
              </div>
            ) : error ? (
              <div className="text-center py-16 bg-white rounded-xl shadow-sm border border-gray-100">
                <svg className="mx-auto h-12 w-12 text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <h3 className="mt-4 text-lg font-medium text-gray-900">Error loading APIs</h3>
                <p className="mt-2 text-gray-600">Please try again later or contact support if the issue persists.</p>
                <Button 
                  className="mt-6"
                  onClick={() => window.location.reload()}
                >
                  Try Again
                </Button>
              </div>
            ) : data?.apis && data.apis.length > 0 ? (
              <>
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-2 xl:grid-cols-3 gap-4 sm:gap-6" data-testid="search-results">
                  {data.apis.map((api: any) => (
                    <APICard 
                      key={api.id} 
                      api={api} 
                      isFeatured={featuredData?.includes(api.id)}
                    />
                  ))}
                </div>

                {/* Pagination */}
                {data.total > 12 && (
                  <div className="mt-8 sm:mt-12 flex justify-center">
                    <nav className="flex items-center space-x-2">
                      <button
                        onClick={() => {
                          const newPage = Math.max(1, page - 1)
                          setPage(newPage)
                          const urlParams = new URLSearchParams(window.location.search)
                          if (newPage === 1) {
                            urlParams.delete('page')
                          } else {
                            urlParams.set('page', String(newPage))
                          }
                          router.push(`/?${urlParams.toString()}`)
                        }}
                        disabled={page === 1}
                        className="px-3 sm:px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors touch-manipulation"
                        data-testid="pagination-prev"
                      >
                        Previous
                      </button>
                      <span className="px-3 sm:px-4 py-2 text-sm font-medium text-gray-700 bg-gray-50 rounded-lg" data-testid="page-number">
                        Page {page} of {Math.ceil(data.total / 12)}
                      </span>
                      <button
                        onClick={() => {
                          const newPage = page + 1
                          setPage(newPage)
                          const urlParams = new URLSearchParams(window.location.search)
                          urlParams.set('page', String(newPage))
                          router.push(`/?${urlParams.toString()}`)
                        }}
                        disabled={page >= Math.ceil(data.total / 12)}
                        className="px-3 sm:px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed transition-colors touch-manipulation"
                        data-testid="pagination-next"
                      >
                        Next
                      </button>
                    </nav>
                  </div>
                )}
              </>
            ) : (
              <div className="text-center py-16 bg-white rounded-xl shadow-sm border border-gray-100">
                <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <h3 className="mt-4 text-lg font-medium text-gray-900">No APIs found</h3>
                <p className="mt-2 text-gray-600 max-w-sm mx-auto">
                  Try adjusting your search or filters to find what you&apos;re looking for.
                </p>
                <Button 
                  variant="secondary"
                  className="mt-6"
                  onClick={() => {
                    setSearchParams({})
                    setActiveTab('browse')
                  }}
                >
                  Clear Filters
                </Button>
              </div>
            )}
          </main>
        </div>

        {/* Strategic CTA Section - Creator Funnel */}
        <div className="bg-gradient-to-r from-primary-600 to-indigo-700 py-16 mt-16">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
            <h2 className="text-3xl font-bold text-white mb-4">
              Can&apos;t Find the API You Need?
            </h2>
            <p className="text-xl text-gray-100 mb-8 max-w-2xl mx-auto">
              Build and monetize it yourself with API-Direct. Turn your ideas into revenue by creating APIs for our marketplace.
            </p>
            <div className="flex flex-col sm:flex-row gap-3 sm:gap-4 justify-center px-4 sm:px-0">
              <Button 
                size="lg" 
                className="bg-white text-primary-700 hover:bg-gray-100 px-6 sm:px-8 py-3 sm:py-4 text-base sm:text-lg touch-manipulation w-full sm:w-auto"
                onClick={() => router.push('/create-api')}
              >
                Create Your API
              </Button>
              <Button 
                size="lg" 
                variant="ghost" 
                className="text-white border-white hover:bg-white/10 px-6 sm:px-8 py-3 sm:py-4 text-base sm:text-lg touch-manipulation w-full sm:w-auto"
                onClick={() => router.push('/register')}
              >
                Get Started Free
              </Button>
            </div>
          </div>
        </div>
      </div>
    </Layout>
  )
}
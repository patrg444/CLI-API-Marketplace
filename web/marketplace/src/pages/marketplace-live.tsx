import React, { useState, useEffect } from 'react'
import Layout from '@/components/Layout'
import APICard from '@/components/APICard'
import { Button } from '@/components/ui/Button'
import { SkeletonCard } from '@/components/ui/Skeleton'
import { marketplaceAPI, MarketplaceAPI, Category } from '@/services/marketplace-api'
import { useRouter } from 'next/router'

export default function MarketplaceLivePage() {
  const router = useRouter()
  const [categories, setCategories] = useState<Category[]>([])
  const [apis, setApis] = useState<MarketplaceAPI[]>([])
  const [featuredApis, setFeaturedApis] = useState<MarketplaceAPI[]>([])
  const [loading, setLoading] = useState(true)
  const [selectedCategory, setSelectedCategory] = useState<string>('')
  const [searchQuery, setSearchQuery] = useState('')
  const [currentPage, setCurrentPage] = useState(1)
  const [totalPages, setTotalPages] = useState(1)

  // Fetch categories on mount
  useEffect(() => {
    marketplaceAPI.getCategories().then(setCategories)
    marketplaceAPI.getFeaturedAPIs().then(setFeaturedApis)
  }, [])

  // Fetch APIs when filters change
  useEffect(() => {
    const fetchAPIs = async () => {
      setLoading(true)
      const result = await marketplaceAPI.getAPIs({
        category: selectedCategory,
        search: searchQuery,
        page: currentPage,
        limit: 12
      })
      setApis(result.apis)
      setTotalPages(result.totalPages)
      setLoading(false)
    }

    fetchAPIs()
  }, [selectedCategory, searchQuery, currentPage])

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    setCurrentPage(1) // Reset to first page on new search
  }

  return (
    <Layout>
      {/* Hero Section */}
      <div className="relative overflow-hidden bg-gradient-to-br from-indigo-600 to-purple-700">
        <div className="relative py-24 px-4 sm:px-6 lg:px-8">
          <div className="max-w-7xl mx-auto text-center">
            <h1 className="text-5xl font-bold text-white mb-6">
              API Marketplace
            </h1>
            <p className="text-xl text-gray-100 mb-8 max-w-3xl mx-auto">
              Discover and integrate powerful APIs to accelerate your development
            </p>
            
            {/* Search Bar */}
            <form onSubmit={handleSearch} className="max-w-2xl mx-auto">
              <div className="flex gap-2">
                <input
                  type="text"
                  placeholder="Search for APIs..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="flex-1 px-4 py-3 rounded-lg text-gray-900 focus:outline-none focus:ring-2 focus:ring-white"
                />
                <button
                  type="submit"
                  className="px-6 py-3 bg-white text-indigo-600 font-medium rounded-lg hover:bg-gray-100 transition"
                >
                  Search
                </button>
              </div>
            </form>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <div className="flex gap-8">
          {/* Categories Sidebar */}
          <aside className="w-64 flex-shrink-0">
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Categories</h3>
              <div className="space-y-2">
                <button
                  onClick={() => setSelectedCategory('')}
                  className={`w-full text-left px-3 py-2 rounded-lg transition ${
                    selectedCategory === '' 
                      ? 'bg-indigo-100 text-indigo-700' 
                      : 'text-gray-700 hover:bg-gray-50'
                  }`}
                >
                  All Categories
                </button>
                {categories.map((category) => (
                  <button
                    key={category.id}
                    onClick={() => setSelectedCategory(category.id)}
                    className={`w-full text-left px-3 py-2 rounded-lg transition ${
                      selectedCategory === category.id 
                        ? 'bg-indigo-100 text-indigo-700' 
                        : 'text-gray-700 hover:bg-gray-50'
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <span>{category.name}</span>
                      <span className="text-sm text-gray-500">{category.count}</span>
                    </div>
                  </button>
                ))}
              </div>
            </div>
          </aside>

          {/* API Grid */}
          <main className="flex-1">
            {/* Featured APIs (only show when no filters) */}
            {!selectedCategory && !searchQuery && featuredApis.length > 0 && (
              <div className="mb-8">
                <h2 className="text-2xl font-bold text-gray-900 mb-4">Featured APIs</h2>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  {featuredApis.slice(0, 3).map((api) => (
                    <div key={api.id} className="transform hover:scale-105 transition">
                      <APICard 
                        api={{
                          id: api.id,
                          name: api.name,
                          description: api.description,
                          category: api.category,
                          average_rating: api.rating,
                          total_reviews: api.reviews,
                          total_subscriptions: api.calls,
                          tags: api.tags,
                          pricing_plans: [{
                            type: api.pricing.type === 'freemium' ? 'free' : 'subscription',
                            monthly_price: api.pricing.monthlyPrice || 0,
                            call_limit: api.pricing.freeCalls || 0
                          }]
                        } as any}
                        isFeatured={true}
                      />
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* All APIs */}
            <div>
              <h2 className="text-2xl font-bold text-gray-900 mb-4">
                {selectedCategory 
                  ? categories.find(c => c.id === selectedCategory)?.name + ' APIs'
                  : searchQuery
                  ? 'Search Results'
                  : 'All APIs'
                }
              </h2>

              {loading ? (
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  {[...Array(6)].map((_, i) => (
                    <SkeletonCard key={i} showImage={false} />
                  ))}
                </div>
              ) : apis.length > 0 ? (
                <>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                    {apis.map((api) => (
                      <APICard 
                        key={api.id}
                        api={{
                          id: api.id,
                          name: api.name,
                          description: api.description,
                          category: api.category,
                          average_rating: api.rating,
                          total_reviews: api.reviews,
                          total_subscriptions: api.calls,
                          tags: api.tags,
                          pricing_plans: [{
                            type: api.pricing.type === 'freemium' ? 'free' : 'subscription',
                            monthly_price: api.pricing.monthlyPrice || 0,
                            call_limit: api.pricing.freeCalls || 0
                          }]
                        } as any}
                      />
                    ))}
                  </div>

                  {/* Pagination */}
                  {totalPages > 1 && (
                    <div className="mt-8 flex justify-center gap-2">
                      <Button
                        onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                        disabled={currentPage === 1}
                        variant="secondary"
                      >
                        Previous
                      </Button>
                      <span className="px-4 py-2 text-gray-700">
                        Page {currentPage} of {totalPages}
                      </span>
                      <Button
                        onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                        disabled={currentPage === totalPages}
                        variant="secondary"
                      >
                        Next
                      </Button>
                    </div>
                  )}
                </>
              ) : (
                <div className="text-center py-16">
                  <p className="text-gray-500">No APIs found</p>
                </div>
              )}
            </div>
          </main>
        </div>
      </div>
    </Layout>
  )
}
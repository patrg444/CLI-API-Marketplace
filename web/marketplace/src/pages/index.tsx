import React, { useState, useEffect } from 'react'
import Layout from '@/components/Layout'
import APICard from '@/components/APICard'
import SearchBar from '@/components/SearchBar'
import { useQuery } from 'react-query'
import apiService from '@/services/api'
import { useRouter } from 'next/router'

export default function MarketplacePage() {
  const router = useRouter()
  const [searchParams, setSearchParams] = useState<any>({})
  const [page, setPage] = useState(1)
  const [activeTab, setActiveTab] = useState<'browse' | 'search'>('browse')

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
        // Use advanced search endpoint
        return await apiService.searchAPIs({
          ...searchParams,
          page,
          limit: 12
        })
      } else {
        // Use regular browse endpoint
        return await apiService.listAPIs({
          category: searchParams.category,
          page,
          limit: 12
        })
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
  }

  const handleCategoryClick = (category: string) => {
    if (category === '') {
      setSearchParams({})
      setActiveTab('browse')
    } else {
      setSearchParams({ category })
      setActiveTab('browse')
    }
    setPage(1)
  }

  return (
    <Layout>
      <div className="bg-gradient-to-br from-indigo-50 to-white py-12 px-4 sm:px-6 lg:px-8">
        <div className="max-w-7xl mx-auto">
          <div className="text-center">
            <h1 className="text-4xl font-bold text-gray-900 sm:text-5xl">
              API Marketplace
            </h1>
            <p className="mt-4 text-xl text-gray-600">
              Discover and integrate powerful APIs for your applications
            </p>
          </div>

          {/* Search Bar */}
          <div className="mt-8 max-w-4xl mx-auto">
            <SearchBar onSearch={handleSearch} showFilters={true} />
          </div>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Search Results Summary */}
        {activeTab === 'search' && data && (
          <div className="mb-6">
            <h2 className="text-lg font-medium text-gray-900">
              {data.total} results found
              {searchParams.query && ` for "${searchParams.query}"`}
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
                        className="text-sm px-2 py-1 bg-gray-100 hover:bg-gray-200 rounded"
                      >
                        {cat} ({count as number})
                      </button>
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>
        )}

        <div className="flex flex-col lg:flex-row gap-8">
          {/* Sidebar for Browse Mode */}
          {activeTab === 'browse' && (
            <aside className="lg:w-64 flex-shrink-0">
              <div className="bg-white rounded-lg shadow p-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-4">Categories</h3>
                <div className="space-y-2">
                  <button
                    onClick={() => handleCategoryClick('')}
                    className={`w-full text-left px-3 py-2 rounded-md text-sm ${
                      !searchParams.category
                        ? 'bg-indigo-100 text-indigo-700 font-medium'
                        : 'text-gray-700 hover:bg-gray-100'
                    }`}
                  >
                    All Categories
                  </button>
                  {['AI/ML', 'Analytics', 'Authentication', 'Communication', 'Data', 'E-commerce', 'Finance', 'Maps & Location', 'Media', 'Social', 'Storage', 'Tools'].map((category) => (
                    <button
                      key={category}
                      onClick={() => handleCategoryClick(category)}
                      className={`w-full text-left px-3 py-2 rounded-md text-sm ${
                        searchParams.category === category
                          ? 'bg-indigo-100 text-indigo-700 font-medium'
                          : 'text-gray-700 hover:bg-gray-100'
                      }`}
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
              <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
                {[...Array(6)].map((_, i) => (
                  <div key={i} className="bg-white rounded-lg shadow-sm p-6 animate-pulse">
                    <div className="h-6 bg-gray-200 rounded w-3/4 mb-3"></div>
                    <div className="h-4 bg-gray-200 rounded w-full mb-2"></div>
                    <div className="h-4 bg-gray-200 rounded w-5/6"></div>
                  </div>
                ))}
              </div>
            ) : error ? (
              <div className="text-center py-12">
                <p className="text-red-600">Error loading APIs. Please try again later.</p>
              </div>
            ) : data?.apis && data.apis.length > 0 ? (
              <>
                <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
                  {data.apis.map((api: any) => (
                    <APICard key={api.id} api={api} />
                  ))}
                </div>

                {/* Pagination */}
                {data.total > 12 && (
                  <div className="mt-8 flex justify-center">
                    <nav className="flex space-x-2">
                      <button
                        onClick={() => setPage(p => Math.max(1, p - 1))}
                        disabled={page === 1}
                        className="px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        Previous
                      </button>
                      <span className="px-3 py-2 text-sm text-gray-700">
                        Page {page} of {Math.ceil(data.total / 12)}
                      </span>
                      <button
                        onClick={() => setPage(p => p + 1)}
                        disabled={page >= Math.ceil(data.total / 12)}
                        className="px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        Next
                      </button>
                    </nav>
                  </div>
                )}
              </>
            ) : (
              <div className="text-center py-12">
                <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9.172 16.172a4 4 0 015.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
                <h3 className="mt-2 text-sm font-medium text-gray-900">No APIs found</h3>
                <p className="mt-1 text-sm text-gray-500">
                  Try adjusting your search or filter to find what you're looking for.
                </p>
              </div>
            )}
          </main>
        </div>
      </div>
    </Layout>
  )
}

import React, { useState, useEffect, useCallback } from 'react'
import { useRouter } from 'next/router'
import { MagnifyingGlassIcon, AdjustmentsHorizontalIcon } from '@heroicons/react/24/outline'
import { ChevronDownIcon } from '@heroicons/react/24/solid'
import apiService from '@/services/api'
import { debounce } from 'lodash'

interface SearchBarProps {
  onSearch?: (params: any) => void
  showFilters?: boolean
}

const categories = [
  'All Categories',
  'AI/ML',
  'Analytics',
  'Authentication',
  'Communication',
  'Data',
  'E-commerce',
  'Finance',
  'Maps & Location',
  'Media',
  'Social',
  'Storage',
  'Tools',
]

const priceRanges = [
  { label: 'All Prices', value: '' },
  { label: 'Free', value: 'free' },
  { label: '$0-$50/mo', value: 'low' },
  { label: '$50-$200/mo', value: 'medium' },
  { label: '$200+/mo', value: 'high' },
]

const sortOptions = [
  { label: 'Relevance', value: 'relevance' },
  { label: 'Highest Rated', value: 'rating' },
  { label: 'Most Popular', value: 'subscriptions' },
  { label: 'Newest', value: 'newest' },
]

export default function SearchBar({ onSearch, showFilters = true }: SearchBarProps) {
  const router = useRouter()
  const [query, setQuery] = useState('')
  const [suggestions, setSuggestions] = useState<string[]>([])
  const [showSuggestions, setShowSuggestions] = useState(false)
  const [showFilterPanel, setShowFilterPanel] = useState(false)
  
  // Filters
  const [category, setCategory] = useState('All Categories')
  const [priceRange, setPriceRange] = useState('')
  const [minRating, setMinRating] = useState(0)
  const [hasFreeTier, setHasFreeTier] = useState(false)
  const [sortBy, setSortBy] = useState('relevance')
  const [tags, setTags] = useState<string[]>([])
  const [tagInput, setTagInput] = useState('')

  // Fetch suggestions with debounce
  const fetchSuggestions = useCallback(
    debounce(async (searchQuery: string) => {
      if (searchQuery.length > 2) {
        try {
          const results = await apiService.getSearchSuggestions(searchQuery)
          setSuggestions(results)
          setShowSuggestions(true)
        } catch (error) {
          console.error('Failed to fetch suggestions:', error)
        }
      } else {
        setSuggestions([])
        setShowSuggestions(false)
      }
    }, 300),
    []
  )

  useEffect(() => {
    fetchSuggestions(query)
  }, [query, fetchSuggestions])

  const handleSearch = (e?: React.FormEvent) => {
    e?.preventDefault()
    
    const searchParams = {
      query: query || undefined,
      category: category !== 'All Categories' ? category : undefined,
      price_range: priceRange || undefined,
      min_rating: minRating > 0 ? minRating : undefined,
      has_free_tier: hasFreeTier || undefined,
      sort_by: sortBy,
      tags: tags.length > 0 ? tags : undefined,
    }

    if (onSearch) {
      onSearch(searchParams)
    } else {
      // Update URL with search params
      const queryString = new URLSearchParams(
        Object.entries(searchParams).reduce((acc, [key, value]) => {
          if (value !== undefined) {
            acc[key] = Array.isArray(value) ? value.join(',') : String(value)
          }
          return acc
        }, {} as Record<string, string>)
      ).toString()
      
      router.push(`/?${queryString}`)
    }
    
    setShowSuggestions(false)
  }

  const handleSuggestionClick = (suggestion: string) => {
    setQuery(suggestion)
    setShowSuggestions(false)
    handleSearch()
  }

  const addTag = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && tagInput.trim()) {
      e.preventDefault()
      if (!tags.includes(tagInput.trim())) {
        setTags([...tags, tagInput.trim()])
      }
      setTagInput('')
    }
  }

  const removeTag = (tagToRemove: string) => {
    setTags(tags.filter(tag => tag !== tagToRemove))
  }

  const clearFilters = () => {
    setCategory('All Categories')
    setPriceRange('')
    setMinRating(0)
    setHasFreeTier(false)
    setSortBy('relevance')
    setTags([])
  }

  return (
    <div className="w-full">
      <form onSubmit={handleSearch} className="relative">
        <div className="flex gap-2">
          <div className="relative flex-1">
            <input
              type="text"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              onBlur={() => setTimeout(() => setShowSuggestions(false), 200)}
              placeholder="Search APIs..."
              className="w-full px-4 py-3 pl-12 pr-4 text-gray-900 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
            <MagnifyingGlassIcon className="absolute left-4 top-3.5 h-5 w-5 text-gray-400" />
            
            {/* Suggestions dropdown */}
            {showSuggestions && suggestions.length > 0 && (
              <div className="absolute z-10 w-full mt-1 bg-white border border-gray-200 rounded-lg shadow-lg">
                {suggestions.map((suggestion, index) => (
                  <button
                    key={index}
                    type="button"
                    onClick={() => handleSuggestionClick(suggestion)}
                    className="w-full px-4 py-2 text-left hover:bg-gray-50 focus:bg-gray-50 focus:outline-none"
                  >
                    {suggestion}
                  </button>
                ))}
              </div>
            )}
          </div>
          
          <button
            type="submit"
            className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            Search
          </button>
          
          {showFilters && (
            <button
              type="button"
              onClick={() => setShowFilterPanel(!showFilterPanel)}
              className="px-4 py-3 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-500"
            >
              <AdjustmentsHorizontalIcon className="h-5 w-5" />
            </button>
          )}
        </div>
      </form>

      {/* Filters Panel */}
      {showFilters && showFilterPanel && (
        <div className="mt-4 p-6 bg-gray-50 rounded-lg">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {/* Category */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Category
              </label>
              <select
                value={category}
                onChange={(e) => setCategory(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              >
                {categories.map((cat) => (
                  <option key={cat} value={cat}>
                    {cat}
                  </option>
                ))}
              </select>
            </div>

            {/* Price Range */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Price Range
              </label>
              <select
                value={priceRange}
                onChange={(e) => setPriceRange(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              >
                {priceRanges.map((range) => (
                  <option key={range.value} value={range.value}>
                    {range.label}
                  </option>
                ))}
              </select>
            </div>

            {/* Sort By */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Sort By
              </label>
              <select
                value={sortBy}
                onChange={(e) => setSortBy(e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              >
                {sortOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </div>

            {/* Minimum Rating */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Minimum Rating
              </label>
              <div className="flex items-center space-x-2">
                {[0, 1, 2, 3, 4, 5].map((rating) => (
                  <button
                    key={rating}
                    type="button"
                    onClick={() => setMinRating(rating)}
                    className={`px-3 py-1 rounded ${
                      minRating === rating
                        ? 'bg-blue-600 text-white'
                        : 'bg-gray-200 text-gray-700 hover:bg-gray-300'
                    }`}
                  >
                    {rating === 0 ? 'Any' : `${rating}+`}
                  </button>
                ))}
              </div>
            </div>

            {/* Free Tier */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Pricing Options
              </label>
              <label className="inline-flex items-center">
                <input
                  type="checkbox"
                  checked={hasFreeTier}
                  onChange={(e) => setHasFreeTier(e.target.checked)}
                  className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                />
                <span className="ml-2">Has free tier</span>
              </label>
            </div>

            {/* Tags */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Tags
              </label>
              <input
                type="text"
                value={tagInput}
                onChange={(e) => setTagInput(e.target.value)}
                onKeyDown={addTag}
                placeholder="Add tags..."
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-blue-500 focus:border-blue-500"
              />
              <div className="flex flex-wrap gap-2 mt-2">
                {tags.map((tag) => (
                  <span
                    key={tag}
                    className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800"
                  >
                    {tag}
                    <button
                      type="button"
                      onClick={() => removeTag(tag)}
                      className="ml-1 text-blue-600 hover:text-blue-800"
                    >
                      ×
                    </button>
                  </span>
                ))}
              </div>
            </div>
          </div>

          <div className="flex justify-between mt-6">
            <button
              type="button"
              onClick={clearFilters}
              className="px-4 py-2 text-gray-700 hover:text-gray-900"
            >
              Clear Filters
            </button>
            <button
              type="button"
              onClick={handleSearch}
              className="px-6 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
            >
              Apply Filters
            </button>
          </div>
        </div>
      )}
    </div>
  )
}

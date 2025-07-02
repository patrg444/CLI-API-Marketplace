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
  'Financial Services',
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
  { label: 'Most Popular', value: 'popularity' },
  { label: 'Newest', value: 'newest' },
]

export default function SearchBar({ onSearch, showFilters = true }: SearchBarProps) {
  const router = useRouter()
  const [query, setQuery] = useState('')
  const [suggestions, setSuggestions] = useState<string[]>([])
  const [showSuggestions, setShowSuggestions] = useState(false)
  const [showFilterPanel, setShowFilterPanel] = useState(true)
  
  // Filters
  const [category, setCategory] = useState('All Categories')
  const [priceRange, setPriceRange] = useState('')
  const [minRating, setMinRating] = useState(0)
  const [hasFreeTier, setHasFreeTier] = useState(false)
  const [sortBy, setSortBy] = useState('relevance')
  const [tags, setTags] = useState<string[]>([])
  const [tagInput, setTagInput] = useState('')

  // Initialize from URL parameters
  useEffect(() => {
    if (typeof window !== 'undefined') {
      const params = new URLSearchParams(window.location.search)
      
      if (params.get('q')) setQuery(params.get('q') || '')
      if (params.get('category')) setCategory(params.get('category') || 'All Categories')
      if (params.get('price_range')) setPriceRange(params.get('price_range') || '')
      if (params.get('min_rating')) setMinRating(parseInt(params.get('min_rating') || '0'))
      if (params.get('has_free_tier')) setHasFreeTier(params.get('has_free_tier') === 'true')
      if (params.get('sort_by')) setSortBy(params.get('sort_by') || 'relevance')
      if (params.get('tags')) setTags(params.get('tags')?.split(',') || [])
    }
  }, [])

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

  const handleSearch = (e?: React.FormEvent, overrides?: any) => {
    e?.preventDefault()
    
    const searchParams = {
      q: query || undefined,
      category: category !== 'All Categories' ? category : undefined,
      price_range: priceRange || undefined,
      min_rating: minRating > 0 ? minRating : undefined,
      has_free_tier: hasFreeTier ? true : undefined,
      sort_by: sortBy,
      tags: tags.length > 0 ? tags : undefined,
      ...overrides
    }

    // Update URL with search params
    const queryString = new URLSearchParams(
      Object.entries(searchParams).reduce((acc, [key, value]) => {
        if (value !== undefined) {
          acc[key] = Array.isArray(value) ? value.join(',') : String(value)
        }
        return acc
      }, {} as Record<string, string>)
    ).toString()
    
    // Use replace to avoid adding to history and prevent scroll jump
    router.replace(`/?${queryString}`, undefined, { scroll: false })
    
    if (onSearch) {
      onSearch(searchParams)
    }
    
    setShowSuggestions(false)
  }

  const handleSuggestionClick = (suggestion: string) => {
    setQuery(suggestion)
    setShowSuggestions(false)
    
    // Create search params with the selected suggestion
    const searchParams = {
      q: suggestion,
      category: category !== 'All Categories' ? category : undefined,
      price_range: priceRange || undefined,
      min_rating: minRating > 0 ? minRating : undefined,
      has_free_tier: hasFreeTier ? true : undefined,
      sort_by: sortBy,
      tags: tags.length > 0 ? tags : undefined,
    }

    // Update URL with search params
    const queryString = new URLSearchParams(
      Object.entries(searchParams).reduce((acc, [key, value]) => {
        if (value !== undefined) {
          acc[key] = Array.isArray(value) ? value.join(',') : String(value)
        }
        return acc
      }, {} as Record<string, string>)
    ).toString()
    
    // Use replace to avoid adding to history and prevent scroll jump
    router.replace(`/?${queryString}`, undefined, { scroll: false })
    
    if (onSearch) {
      onSearch(searchParams)
    }
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
              data-testid="search-input"
            />
            <MagnifyingGlassIcon className="absolute left-4 top-3.5 h-5 w-5 text-gray-400" />
            
            {/* Suggestions dropdown */}
            {showSuggestions && suggestions.length > 0 && (
              <div className="absolute z-10 w-full mt-1 bg-white border border-gray-200 rounded-lg shadow-lg" data-testid="search-suggestions">
                {suggestions.map((suggestion, index) => (
                  <button
                    key={index}
                    type="button"
                    onClick={() => handleSuggestionClick(suggestion)}
                    className="w-full px-4 py-2 text-left hover:bg-gray-50 focus:bg-gray-50 focus:outline-none"
                    data-testid="search-suggestion"
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
            data-testid="search-submit"
          >
            Search
          </button>
          
          {showFilters && (
            <button
              type="button"
              onClick={() => setShowFilterPanel(!showFilterPanel)}
              className="px-4 py-3 bg-gray-100 text-gray-700 rounded-lg hover:bg-gray-200 focus:outline-none focus:ring-2 focus:ring-gray-500"
              data-testid="filter-toggle-button"
            >
              <AdjustmentsHorizontalIcon className="h-5 w-5" />
            </button>
          )}
        </div>
      </form>

      {/* Flattened Filters - Horizontal Layout */}
      {showFilters && showFilterPanel && (
        <div className="mt-4 flex flex-wrap items-center gap-3 sm:gap-4">
          {/* Category Dropdown */}
          <div className="relative">
            <select
              value={category}
              onChange={(e) => {
                const newCategory = e.target.value
                setCategory(newCategory)
                handleSearch(undefined, { category: newCategory !== 'All Categories' ? newCategory : undefined })
              }}
              className="appearance-none bg-white border border-gray-300 rounded-lg px-4 py-2 pr-8 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              data-testid="category-filter"
            >
              {categories.map((cat) => (
                <option key={cat} value={cat}>
                  {cat}
                </option>
              ))}
            </select>
            <div className="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
              <ChevronDownIcon className="w-4 h-4 text-gray-400" />
            </div>
          </div>

          {/* Price Range Dropdown */}
          <div className="relative">
            <select
              value={priceRange}
              onChange={(e) => {
                setPriceRange(e.target.value)
                handleSearch(undefined, { price_range: e.target.value || undefined })
              }}
              className="appearance-none bg-white border border-gray-300 rounded-lg px-4 py-2 pr-8 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              data-testid="price-filter"
            >
              {priceRanges.map((range) => (
                <option key={range.value} value={range.value}>
                  {range.label}
                </option>
              ))}
            </select>
            <div className="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
              <ChevronDownIcon className="w-4 h-4 text-gray-400" />
            </div>
          </div>

          {/* Sort By Dropdown */}
          <div className="relative">
            <select
              value={sortBy}
              onChange={(e) => {
                const newSort = e.target.value
                setSortBy(newSort)
                handleSearch(undefined, { sort_by: newSort })
              }}
              className="appearance-none bg-white border border-gray-300 rounded-lg px-4 py-2 pr-8 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              data-testid="sort-select"
            >
              {sortOptions.map((option) => (
                <option key={option.value} value={option.value}>
                  {option.label}
                </option>
              ))}
            </select>
            <div className="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
              <ChevronDownIcon className="w-4 h-4 text-gray-400" />
            </div>
          </div>

          {/* Rating Filter */}
          <div className="relative">
            <select
              value={minRating}
              onChange={(e) => {
                const newRating = parseInt(e.target.value)
                setMinRating(newRating)
                handleSearch(undefined, { min_rating: newRating > 0 ? newRating : undefined })
              }}
              className="appearance-none bg-white border border-gray-300 rounded-lg px-4 py-2 pr-8 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              data-testid="rating-filter"
            >
              <option value={0}>Any Rating</option>
              <option value={4.5}>4.5+ Stars</option>
              <option value={4}>4.0+ Stars</option>
              <option value={3.5}>3.5+ Stars</option>
              <option value={3}>3.0+ Stars</option>
            </select>
            <div className="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
              <ChevronDownIcon className="w-4 h-4 text-gray-400" />
            </div>
          </div>

          {/* Free Tier Toggle */}
          <label className="flex items-center cursor-pointer">
            <input
              type="checkbox"
              checked={hasFreeTier}
              onChange={(e) => {
                const newFreeTier = e.target.checked
                setHasFreeTier(newFreeTier)
                handleSearch(undefined, { has_free_tier: newFreeTier ? true : undefined })
              }}
              className="sr-only"
              data-testid="free-tier-filter"
            />
            <div className={`relative w-12 h-6 rounded-full transition-colors duration-200 ${
              hasFreeTier ? 'bg-blue-600' : 'bg-gray-300'
            }`}>
              <div className={`absolute top-0.5 left-0.5 w-5 h-5 bg-white rounded-full transition-transform duration-200 ${
                hasFreeTier ? 'translate-x-6' : 'translate-x-0'
              }`}></div>
            </div>
            <span className="ml-2 text-sm text-gray-700">Has free tier</span>
          </label>

          {/* Tags Input */}
          <div className="flex items-center gap-2">
            <input
              type="text"
              value={tagInput}
              onChange={(e) => setTagInput(e.target.value)}
              onKeyDown={addTag}
              placeholder="Add tags..."
              className="px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 w-32"
            />
            {tags.length > 0 && (
              <div className="flex gap-1">
                {tags.map((tag) => (
                  <span
                    key={tag}
                    className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-blue-100 text-blue-800"
                  >
                    #{tag}
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
            )}
          </div>

          {/* Clear Filters */}
          {(category !== 'All Categories' || priceRange || minRating > 0 || hasFreeTier || tags.length > 0) && (
            <button
              type="button"
              onClick={() => {
                clearFilters()
                handleSearch()
              }}
              className="text-sm text-gray-500 hover:text-gray-700 border border-gray-300 rounded-lg px-3 py-2 hover:bg-gray-50 transition-colors"
            >
              Clear Filters
            </button>
          )}

          {/* Apply Filters */}
          <button
            type="button"
            onClick={handleSearch}
            className="px-4 py-2 bg-blue-600 text-white text-sm rounded-lg hover:bg-blue-700 transition-colors"
          >
            Apply Filters
          </button>
        </div>
      )}
    </div>
  )
}

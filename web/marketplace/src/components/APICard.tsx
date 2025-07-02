import React from 'react'
import Link from 'next/link'
import { API } from '@/types/api'

interface APICardProps {
  api: API
  isFeatured?: boolean
}

const StarIcon = () => (
  <svg className="h-4 w-4 text-yellow-400" fill="currentColor" viewBox="0 0 20 20">
    <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
  </svg>
)

const UsersIcon = () => (
  <svg className="h-4 w-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
  </svg>
)

const ArrowRightIcon = () => (
  <svg className="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
  </svg>
)

export default function APICard({ api, isFeatured }: APICardProps) {
  const freePlan = api.pricing_plans?.find(plan => plan.type === 'free')
  const paidPlans = api.pricing_plans?.filter(plan => plan.type !== 'free') || []
  const lowestPrice = paidPlans.length > 0 
    ? Math.min(...paidPlans.map(p => p.monthly_price || 0).filter(p => p > 0))
    : null

  return (
    <Link href={`/apis/${api.id}`} className="block group">
      <div 
        className="card card-hover h-full flex flex-col"
        data-testid="api-card"
      >
        <div className="p-4 sm:p-6 flex-1">
          {/* Header */}
          <div className="flex items-start justify-between mb-3 sm:mb-4">
            {api.icon_url ? (
              <img 
                src={api.icon_url} 
                alt={`${api.name} icon`}
                className="w-12 h-12 sm:w-14 sm:h-14 rounded-xl object-cover shadow-sm flex-shrink-0"
              />
            ) : (
              <div className="w-12 h-12 sm:w-14 sm:h-14 rounded-xl bg-gradient-to-br from-primary-500 to-primary-600 flex items-center justify-center text-white font-bold text-base sm:text-lg shadow-sm flex-shrink-0">
                {api.name.charAt(0).toUpperCase()}
              </div>
            )}
            
            <div className="flex flex-col sm:flex-row items-end sm:items-center gap-1 sm:gap-2 ml-3">
              {isFeatured && (
                <span className="badge badge-warning text-xs">Featured</span>
              )}
              {api.total_subscriptions && api.total_subscriptions > 100 && (
                <span className="badge badge-success text-xs">Popular</span>
              )}
            </div>
          </div>

          {/* Content */}
          <div className="space-y-2 sm:space-y-3">
            <div>
              <h3 className="text-base sm:text-lg font-semibold text-gray-900 group-hover:text-primary-600 transition-colors leading-tight">
                {api.name}
              </h3>
              <p className="text-sm text-gray-600 mt-1 line-clamp-2 leading-relaxed">
                {api.description}
              </p>
            </div>

            {/* Category and Tags */}
            <div className="flex flex-wrap items-center gap-1.5 sm:gap-2">
              <span className="badge badge-primary text-xs">
                {api.category}
              </span>
              {api.tags?.slice(0, window.innerWidth < 640 ? 1 : 2).map((tag, index) => (
                <span key={index} className="badge badge-gray text-xs">
                  {tag}
                </span>
              ))}
              {api.tags && api.tags.length > (window.innerWidth < 640 ? 1 : 2) && (
                <span className="text-xs text-gray-500">
                  +{api.tags.length - (window.innerWidth < 640 ? 1 : 2)} more
                </span>
              )}
            </div>

            {/* Stats */}
            <div className="flex items-center gap-3 sm:gap-4 text-sm">
              {api.average_rating && api.average_rating > 0 && (
                <div className="flex items-center gap-1">
                  <StarIcon />
                  <span className="font-medium text-gray-900" data-testid="api-rating">
                    {api.average_rating.toFixed(1)}
                  </span>
                  <span className="text-gray-500 hidden sm:inline">
                    ({api.total_reviews || 0})
                  </span>
                </div>
              )}
              
              {api.total_subscriptions && api.total_subscriptions > 0 && (
                <div className="flex items-center gap-1 text-gray-500">
                  <UsersIcon />
                  <span className="text-xs sm:text-sm">
                    {api.total_subscriptions > 1000 
                      ? `${Math.round(api.total_subscriptions / 1000)}k users`
                      : `${api.total_subscriptions.toLocaleString()} users`
                    }
                  </span>
                </div>
              )}
            </div>
          </div>
        </div>

        {/* Footer */}
        <div className="px-4 sm:px-6 py-3 sm:py-4 bg-gray-50 border-t border-gray-100 flex items-center justify-between">
          <div className="text-xs sm:text-sm">
            {freePlan ? (
              <span className="text-success-600 font-medium flex items-center gap-1.5" data-testid="free-tier-badge">
                <span className="w-2 h-2 bg-success-500 rounded-full"></span>
                <span className="hidden sm:inline">Free tier available</span>
                <span className="sm:hidden">Free tier</span>
              </span>
            ) : lowestPrice ? (
              <span className="text-gray-900">
                <span className="hidden sm:inline">From </span>
                <span className="font-semibold text-base sm:text-lg">${lowestPrice}</span>
                <span className="text-xs sm:text-sm">/mo</span>
              </span>
            ) : (
              <span className="text-gray-500 text-xs sm:text-sm">
                <span className="hidden sm:inline">Contact for pricing</span>
                <span className="sm:hidden">Contact us</span>
              </span>
            )}
          </div>
          
          <div className="text-gray-400 group-hover:text-primary-600 transition-colors">
            <ArrowRightIcon />
          </div>
        </div>
      </div>
    </Link>
  )
}
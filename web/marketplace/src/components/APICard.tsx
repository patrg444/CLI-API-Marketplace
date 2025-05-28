import React from 'react'
import Link from 'next/link'
import { API } from '@/types/api'

interface APICardProps {
  api: API
}

export default function APICard({ api }: APICardProps) {
  const freePlan = api.pricing_plans?.find(plan => plan.type === 'free')
  const paidPlans = api.pricing_plans?.filter(plan => plan.type !== 'free') || []
  const lowestPrice = paidPlans.length > 0 
    ? Math.min(...paidPlans.map(p => p.monthly_price || 0).filter(p => p > 0))
    : null

  return (
    <Link href={`/api/${api.id}`}>
      <div className="bg-white rounded-lg shadow-sm hover:shadow-md transition-shadow duration-200 p-6 cursor-pointer border border-gray-200">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <h3 className="text-lg font-semibold text-gray-900 mb-1">{api.name}</h3>
            <p className="text-sm text-gray-600 mb-3 line-clamp-2">{api.description}</p>
            
            <div className="flex items-center gap-4 text-sm">
              <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-indigo-100 text-indigo-800">
                {api.category}
              </span>
              
              {api.average_rating && (
                <div className="flex items-center">
                  <svg className="h-4 w-4 text-yellow-400 mr-1" fill="currentColor" viewBox="0 0 20 20">
                    <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                  </svg>
                  <span className="text-gray-600">{api.average_rating.toFixed(1)}</span>
                  <span className="text-gray-400 ml-1">({api.total_reviews})</span>
                </div>
              )}
            </div>
          </div>
          
          {api.icon_url && (
            <img 
              src={api.icon_url} 
              alt={`${api.name} icon`}
              className="w-12 h-12 rounded-lg object-cover ml-4"
            />
          )}
        </div>
        
        <div className="mt-4 pt-4 border-t border-gray-200 flex items-center justify-between">
          <div className="text-sm">
            {freePlan ? (
              <span className="text-green-600 font-medium">Free tier available</span>
            ) : lowestPrice ? (
              <span className="text-gray-900">
                From <span className="font-semibold">${lowestPrice}/mo</span>
              </span>
            ) : (
              <span className="text-gray-500">Pricing available</span>
            )}
          </div>
          
          {api.total_subscriptions && api.total_subscriptions > 0 && (
            <span className="text-xs text-gray-500">
              {api.total_subscriptions} active users
            </span>
          )}
        </div>
      </div>
    </Link>
  )
}

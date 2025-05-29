import React, { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from 'react-query'
import { Auth } from 'aws-amplify'
import apiService from '@/services/api'
import { StarIcon } from '@heroicons/react/24/solid'
import { StarIcon as StarOutlineIcon } from '@heroicons/react/24/outline'
import { HandThumbUpIcon, HandThumbDownIcon } from '@heroicons/react/24/outline'
import { CheckBadgeIcon } from '@heroicons/react/24/solid'

interface ReviewSectionProps {
  apiId: string
  canReview?: boolean
}

interface Review {
  id: string
  user_id: string
  user_name: string
  rating: number
  title: string
  comment: string
  helpful_count: number
  not_helpful_count: number
  verified_purchase: boolean
  creator_response?: string
  created_at: string
  user_vote?: 'helpful' | 'not_helpful'
}

interface ReviewStats {
  average_rating: number
  total_reviews: number
  rating_distribution: {
    [key: string]: number
  }
}

export default function ReviewSection({ apiId, canReview = false }: ReviewSectionProps) {
  const queryClient = useQueryClient()
  const [showReviewForm, setShowReviewForm] = useState(false)
  const [rating, setRating] = useState(5)
  const [title, setTitle] = useState('')
  const [comment, setComment] = useState('')
  const [sortBy, setSortBy] = useState<'newest' | 'oldest' | 'highest' | 'lowest' | 'most_helpful'>('most_helpful')
  const [page, setPage] = useState(1)

  // Fetch reviews
  const { data: reviewsData, isLoading: reviewsLoading } = useQuery(
    ['reviews', apiId, sortBy, page],
    () => apiService.getAPIReviews(apiId, { sort: sortBy, page, limit: 10 }),
    { keepPreviousData: true }
  )

  // Fetch review stats
  const { data: statsData } = useQuery(
    ['reviewStats', apiId],
    () => apiService.getReviewStats(apiId)
  )

  // Submit review mutation
  const submitReviewMutation = useMutation(
    (data: { rating: number; title: string; comment: string }) => 
      apiService.submitReview(apiId, data),
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['reviews', apiId])
        queryClient.invalidateQueries(['reviewStats', apiId])
        setShowReviewForm(false)
        setRating(5)
        setTitle('')
        setComment('')
      }
    }
  )

  // Vote on review mutation
  const voteOnReviewMutation = useMutation(
    ({ reviewId, helpful }: { reviewId: string; helpful: boolean }) =>
      apiService.voteOnReview(reviewId, helpful),
    {
      onSuccess: () => {
        queryClient.invalidateQueries(['reviews', apiId])
      }
    }
  )

  const handleSubmitReview = (e: React.FormEvent) => {
    e.preventDefault()
    submitReviewMutation.mutate({ rating, title, comment })
  }

  const handleVote = (reviewId: string, helpful: boolean) => {
    voteOnReviewMutation.mutate({ reviewId, helpful })
  }

  const renderStars = (rating: number, interactive = false, size = 'sm') => {
    const sizeClasses = size === 'sm' ? 'h-4 w-4' : 'h-5 w-5'
    
    return (
      <div className="flex items-center">
        {[1, 2, 3, 4, 5].map((star) => (
          <button
            key={star}
            type="button"
            disabled={!interactive}
            onClick={() => interactive && setRating(star)}
            className={interactive ? 'cursor-pointer' : 'cursor-default'}
          >
            {star <= rating ? (
              <StarIcon className={`${sizeClasses} text-yellow-400`} />
            ) : (
              <StarOutlineIcon className={`${sizeClasses} text-gray-300`} />
            )}
          </button>
        ))}
      </div>
    )
  }

  return (
    <div className="mt-12">
      <div className="border-t pt-8">
        <h2 className="text-2xl font-bold text-gray-900 mb-6">Customer Reviews</h2>

        {/* Review Stats */}
        {statsData && (
          <div className="bg-gray-50 rounded-lg p-6 mb-8">
            <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
              <div>
                <div className="flex items-center gap-3">
                  <span className="text-3xl font-bold text-gray-900">
                    {statsData.average_rating.toFixed(1)}
                  </span>
                  {renderStars(Math.round(statsData.average_rating), false, 'md')}
                  <span className="text-gray-600">
                    ({statsData.total_reviews} reviews)
                  </span>
                </div>
              </div>

              {/* Rating Distribution */}
              <div className="w-full sm:w-64">
                {[5, 4, 3, 2, 1].map((stars) => {
                  const count = statsData.rating_distribution[stars] || 0
                  const percentage = statsData.total_reviews > 0 
                    ? (count / statsData.total_reviews) * 100 
                    : 0

                  return (
                    <div key={stars} className="flex items-center gap-2 text-sm">
                      <span className="w-3">{stars}</span>
                      <StarIcon className="h-3 w-3 text-yellow-400" />
                      <div className="flex-1 bg-gray-200 rounded-full h-2">
                        <div
                          className="bg-yellow-400 h-2 rounded-full"
                          style={{ width: `${percentage}%` }}
                        />
                      </div>
                      <span className="w-12 text-right text-gray-600">
                        {count}
                      </span>
                    </div>
                  )
                })}
              </div>
            </div>
          </div>
        )}

        {/* Review Actions */}
        <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4 mb-6">
          {canReview && !showReviewForm && (
            <button
              onClick={() => setShowReviewForm(true)}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
            >
              Write a Review
            </button>
          )}

          <div className="flex items-center gap-2">
            <label className="text-sm text-gray-600">Sort by:</label>
            <select
              value={sortBy}
              onChange={(e) => setSortBy(e.target.value as any)}
              className="px-3 py-1 border border-gray-300 rounded-md text-sm"
            >
              <option value="most_helpful">Most Helpful</option>
              <option value="newest">Newest</option>
              <option value="oldest">Oldest</option>
              <option value="highest">Highest Rating</option>
              <option value="lowest">Lowest Rating</option>
            </select>
          </div>
        </div>

        {/* Review Form */}
        {showReviewForm && (
          <div className="bg-gray-50 rounded-lg p-6 mb-8">
            <h3 className="text-lg font-semibold mb-4">Write Your Review</h3>
            <form onSubmit={handleSubmitReview}>
              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Rating
                </label>
                {renderStars(rating, true, 'md')}
              </div>

              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Review Title
                </label>
                <input
                  type="text"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  required
                  className="w-full px-3 py-2 border border-gray-300 rounded-md"
                  placeholder="Sum up your experience in a few words"
                />
              </div>

              <div className="mb-4">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Your Review
                </label>
                <textarea
                  value={comment}
                  onChange={(e) => setComment(e.target.value)}
                  required
                  rows={4}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md"
                  placeholder="Share your experience with this API"
                />
              </div>

              <div className="flex gap-3">
                <button
                  type="submit"
                  disabled={submitReviewMutation.isLoading}
                  className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
                >
                  {submitReviewMutation.isLoading ? 'Submitting...' : 'Submit Review'}
                </button>
                <button
                  type="button"
                  onClick={() => {
                    setShowReviewForm(false)
                    setRating(5)
                    setTitle('')
                    setComment('')
                  }}
                  className="px-4 py-2 text-gray-700 hover:text-gray-900"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        )}

        {/* Reviews List */}
        {reviewsLoading ? (
          <div className="space-y-6">
            {[1, 2, 3].map((i) => (
              <div key={i} className="animate-pulse">
                <div className="h-4 bg-gray-200 rounded w-1/4 mb-2"></div>
                <div className="h-3 bg-gray-200 rounded w-3/4 mb-2"></div>
                <div className="h-3 bg-gray-200 rounded w-1/2"></div>
              </div>
            ))}
          </div>
        ) : reviewsData?.reviews && reviewsData.reviews.length > 0 ? (
          <>
            <div className="space-y-6">
              {reviewsData.reviews.map((review: Review) => (
                <div key={review.id} className="border-b pb-6">
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <div className="flex items-center gap-3 mb-2">
                        <div className="flex items-center">
                          {renderStars(review.rating)}
                          <span className="ml-2 font-semibold text-gray-900">
                            {review.title}
                          </span>
                        </div>
                        {review.verified_purchase && (
                          <div className="flex items-center text-green-600 text-sm">
                            <CheckBadgeIcon className="h-4 w-4 mr-1" />
                            Verified Purchase
                          </div>
                        )}
                      </div>
                      <p className="text-sm text-gray-600 mb-2">
                        By {review.user_name} on {new Date(review.created_at).toLocaleDateString()}
                      </p>
                      <p className="text-gray-700 mb-3">{review.comment}</p>

                      {/* Creator Response */}
                      {review.creator_response && (
                        <div className="mt-3 ml-4 p-3 bg-blue-50 rounded-md">
                          <p className="text-sm font-medium text-blue-900 mb-1">
                            Creator Response:
                          </p>
                          <p className="text-sm text-blue-800">{review.creator_response}</p>
                        </div>
                      )}

                      {/* Helpful Votes */}
                      <div className="flex items-center gap-4 mt-3">
                        <span className="text-sm text-gray-600">
                          {review.helpful_count} people found this helpful
                        </span>
                        <div className="flex items-center gap-2">
                          <button
                            onClick={() => handleVote(review.id, true)}
                            className={`flex items-center gap-1 px-3 py-1 rounded text-sm ${
                              review.user_vote === 'helpful'
                                ? 'bg-green-100 text-green-700'
                                : 'text-gray-600 hover:bg-gray-100'
                            }`}
                          >
                            <HandThumbUpIcon className="h-4 w-4" />
                            Helpful
                          </button>
                          <button
                            onClick={() => handleVote(review.id, false)}
                            className={`flex items-center gap-1 px-3 py-1 rounded text-sm ${
                              review.user_vote === 'not_helpful'
                                ? 'bg-red-100 text-red-700'
                                : 'text-gray-600 hover:bg-gray-100'
                            }`}
                          >
                            <HandThumbDownIcon className="h-4 w-4" />
                            Not Helpful
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {/* Pagination */}
            {reviewsData.total > 10 && (
              <div className="mt-8 flex justify-center">
                <nav className="flex space-x-2">
                  <button
                    onClick={() => setPage(p => Math.max(1, p - 1))}
                    disabled={page === 1}
                    className="px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50"
                  >
                    Previous
                  </button>
                  <span className="px-3 py-2 text-sm text-gray-700">
                    Page {page} of {Math.ceil(reviewsData.total / 10)}
                  </span>
                  <button
                    onClick={() => setPage(p => p + 1)}
                    disabled={page >= Math.ceil(reviewsData.total / 10)}
                    className="px-3 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 disabled:opacity-50"
                  >
                    Next
                  </button>
                </nav>
              </div>
            )}
          </>
        ) : (
          <div className="text-center py-8">
            <p className="text-gray-500">No reviews yet. Be the first to review this API!</p>
          </div>
        )}
      </div>
    </div>
  )
}

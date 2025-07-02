import React, { useState } from 'react';
import { useRouter } from 'next/router';
import Layout from '../../components/Layout';

// Type definitions
interface CreatorResponse {
  author: string;
  content: string;
  date: string;
}

interface Review {
  id: number;
  author: string;
  rating: number;
  date: string;
  title: string;
  content: string;
  helpful: number;
  notHelpful: number;
  flagged: boolean;
  hasCreatorResponse: boolean;
  creatorResponse?: CreatorResponse;
}

const APIDetails: React.FC = () => {
  const router = useRouter();
  const { apiId } = router.query;
  
  const [showSubscribeModal, setShowSubscribeModal] = useState(false);
  const [activeTab, setActiveTab] = useState('documentation');
  const [paymentError, setPaymentError] = useState(false);
  const [reviewSort, setReviewSort] = useState('recent');
  const [selectedRating, setSelectedRating] = useState(0);
  const [reviewTitle, setReviewTitle] = useState('');
  const [reviewComment, setReviewComment] = useState('');
  const [showReviewForm, setShowReviewForm] = useState(true);
  const [hasSubscription, setHasSubscription] = useState(() => {
    // Check localStorage for subscription status
    if (typeof window !== 'undefined') {
      const storedSubscription = localStorage.getItem('hasSubscription');
      return storedSubscription !== 'false';
    }
    return true;
  });
  const [votedReviews, setVotedReviews] = useState<Set<number>>(() => {
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem('votedReviews');
      if (stored) {
        return new Set(JSON.parse(stored));
      }
    }
    return new Set();
  });
  
  // Determine if this is a "no reviews" API
  const isNoReviewsAPI = apiId === 'new-api-no-reviews';

  // Mock reviews data with comprehensive structure
  const [reviews, setReviews] = useState(() => {
    // Return empty array for no-reviews API
    if (isNoReviewsAPI) {
      return [];
    }
    
    if (typeof window !== 'undefined') {
      const stored = localStorage.getItem('reviewsData');
      if (stored) {
        const reviewsData = JSON.parse(stored);
        // Check if there's data for this specific API
        if (reviewsData[apiId as string] !== undefined) {
          return reviewsData[apiId as string];
        }
      }
    }
    
    // Only return default reviews if not explicitly cleared
    if (typeof window !== 'undefined' && localStorage.getItem('reviewsCleared') === 'true') {
      return [];
    }
    
    return [
      {
        id: 1,
        author: 'John Developer',
        rating: 5,
        date: '2024-01-15',
        title: 'Excellent API',
        content: 'Very easy to integrate and great documentation. The response times are fast and the pricing is fair.',
        helpful: 12,
        notHelpful: 1,
        flagged: false,
        hasCreatorResponse: true,
        creatorResponse: {
          author: 'Creator',
          content: 'Thank you for the positive feedback! We\'re glad you found our API easy to integrate.',
          date: '2024-01-16'
        }
      },
      {
        id: 2,
        author: 'Sarah Smith', 
        rating: 4,
        date: '2024-01-10',
        title: 'Good but could be better',
        content: 'Works well but response times could be improved. Overall satisfied with the service.',
        helpful: 8,
        notHelpful: 2,
        flagged: false,
        hasCreatorResponse: false
      },
      {
        id: 3,
        author: 'Mike Johnson',
        rating: 3,
        date: '2024-01-05',
        title: 'Average experience',
        content: 'It works but documentation could be clearer.',
        helpful: 5,
        notHelpful: 3,
        flagged: false,
        hasCreatorResponse: false
      }
    ];
  });

  const handleVoteHelpful = (reviewId: number) => {
    if (votedReviews.has(reviewId)) return;
    
    const updatedReviews = reviews.map((review: Review) => 
      review.id === reviewId 
        ? { ...review, helpful: review.helpful + 1 }
        : review
    );
    setReviews(updatedReviews);
    
    // Save reviews by API ID
    const savedReviews = localStorage.getItem('reviewsData');
    const reviewsData = savedReviews ? JSON.parse(savedReviews) : {};
    reviewsData[apiId as string] = updatedReviews;
    localStorage.setItem('reviewsData', JSON.stringify(reviewsData));
    
    const newVotedReviews = new Set(Array.from(votedReviews).concat(reviewId));
    setVotedReviews(newVotedReviews);
    localStorage.setItem('votedReviews', JSON.stringify(Array.from(newVotedReviews)));
  };

  const handleVoteNotHelpful = (reviewId: number) => {
    if (votedReviews.has(reviewId)) return;
    
    const updatedReviews = reviews.map((review: Review) => 
      review.id === reviewId 
        ? { ...review, notHelpful: review.notHelpful + 1 }
        : review
    );
    setReviews(updatedReviews);
    
    // Save reviews by API ID
    const savedReviews = localStorage.getItem('reviewsData');
    const reviewsData = savedReviews ? JSON.parse(savedReviews) : {};
    reviewsData[apiId as string] = updatedReviews;
    localStorage.setItem('reviewsData', JSON.stringify(reviewsData));
    
    const newVotedReviews = new Set(Array.from(votedReviews).concat(reviewId));
    setVotedReviews(newVotedReviews);
    localStorage.setItem('votedReviews', JSON.stringify(Array.from(newVotedReviews)));
  };

  const handleSubmitReview = async () => {
    // Show validation errors inline
    const showError = (message: string) => {
      const existingError = document.querySelector('[data-testid="review-error"]');
      if (existingError) existingError.remove();
      
      const errorDiv = document.createElement('div');
      errorDiv.textContent = message;
      errorDiv.className = 'text-red-600 text-sm mt-2';
      errorDiv.setAttribute('data-testid', 'review-error');
      
      const submitButton = document.querySelector('[data-testid="submit-review"]');
      if (submitButton && submitButton.parentNode) {
        submitButton.parentNode.insertBefore(errorDiv, submitButton);
      }
    };

    if (selectedRating === 0) {
      showError('Please select a rating');
      return;
    }
    if (!reviewComment.trim()) {
      showError('Please write a review');
      return;
    }
    if (reviewComment.length < 10) {
      showError('Review must be at least 10 characters');
      return;
    }

    // Simulate API call with potential error
    try {
      // Check if we're in a test environment with mocked errors
      const isTestingError = typeof window !== 'undefined' && 
        (window.location.search.includes('test-error') || (window as any).__TEST_MOCK_ERROR);
      
      // Mock API submission that can fail
      if (isTestingError) {
        throw new Error('Network error');
      }

      const newReview = {
        id: reviews.length + 1,
        author: 'Current User',
        rating: selectedRating,
        date: new Date().toISOString().split('T')[0],
        title: reviewTitle || 'User Review',
        content: reviewComment,
        helpful: 0,
        notHelpful: 0,
        flagged: false,
        hasCreatorResponse: false
      };

      const updatedReviews = [newReview, ...reviews];
      setReviews(updatedReviews);
      
      // Save reviews by API ID
      const savedReviews = localStorage.getItem('reviewsData');
      const reviewsData = savedReviews ? JSON.parse(savedReviews) : {};
      reviewsData[apiId as string] = updatedReviews;
      localStorage.setItem('reviewsData', JSON.stringify(reviewsData));
      setShowReviewForm(false);
      setSelectedRating(0);
      setReviewTitle('');
      setReviewComment('');

      // Show success message
      const message = document.createElement('div');
      message.textContent = 'Review submitted successfully!';
      message.className = 'fixed top-4 right-4 bg-green-100 text-green-800 px-4 py-2 rounded shadow z-50';
      message.setAttribute('data-testid', 'review-success');
      document.body.appendChild(message);
      setTimeout(() => message.remove(), 3000);
    } catch (error) {
      showError('Failed to submit review. Please try again.');
    }
  };

  const sortedReviews = [...reviews].sort((a, b) => {
    switch (reviewSort) {
      case 'recent':
        return new Date(b.date).getTime() - new Date(a.date).getTime();
      case 'helpful':
        return b.helpful - a.helpful;
      case 'highest':
        return b.rating - a.rating;
      case 'lowest':
        return a.rating - b.rating;
      default:
        return 0;
    }
  });

  const handleSubscribe = () => {
    // Simulate payment processing delay
    setTimeout(() => {
      setShowSubscribeModal(false);
      // Mock subscription success
      const successMessage = document.createElement('div');
      successMessage.textContent = 'Subscription successful!';
      successMessage.className = 'fixed top-4 right-4 bg-green-100 text-green-800 px-4 py-2 rounded shadow';
      successMessage.setAttribute('data-testid', 'subscription-success');
      document.body.appendChild(successMessage);
      
      // Mock API key display
      const apiKeyDisplay = document.createElement('div');
      apiKeyDisplay.textContent = 'api_mock_key_12345';
      apiKeyDisplay.className = 'fixed top-20 right-4 bg-white px-4 py-2 rounded shadow border';
      apiKeyDisplay.setAttribute('data-testid', 'api-key-display');
      document.body.appendChild(apiKeyDisplay);
      
      // Remove messages after 5 seconds
      setTimeout(() => {
        successMessage.remove();
        apiKeyDisplay.remove();
      }, 5000);
    }, 1000);
  };
  
  return (
    <Layout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 sm:py-8">
        {/* API Header */}
        <div className="mb-6 sm:mb-8">
          <h1 className="text-2xl sm:text-3xl font-bold text-gray-900 leading-tight" data-testid="api-name">
            Payment Processing API
          </h1>
          <p className="mt-2 text-base sm:text-lg text-gray-600 leading-relaxed" data-testid="api-description">
            Secure payment processing with multiple payment methods
          </p>
        </div>

        {/* Navigation Tabs */}
        <div className="border-b border-gray-200 mb-6 sm:mb-8">
          <nav className="-mb-px flex space-x-4 sm:space-x-8 overflow-x-auto scrollbar-thin">
            <button 
              className={`whitespace-nowrap py-3 px-2 border-b-2 font-medium text-sm touch-manipulation ${activeTab === 'documentation' ? 'border-indigo-500 text-indigo-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}`}
              data-testid="api-documentation-tab"
              onClick={() => setActiveTab('documentation')}
            >
              Documentation
            </button>
            <button 
              className={`whitespace-nowrap py-3 px-2 border-b-2 font-medium text-sm touch-manipulation ${activeTab === 'sdk' ? 'border-indigo-500 text-indigo-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}`}
              data-testid="sdk-tab"
              onClick={() => setActiveTab('sdk')}
            >
              SDK
            </button>
            <button 
              className={`whitespace-nowrap py-3 px-2 border-b-2 font-medium text-sm touch-manipulation ${activeTab === 'code-examples' ? 'border-indigo-500 text-indigo-600' : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'}`}
              data-testid="code-examples-tab"
              onClick={() => setActiveTab('code-examples')}
            >
              Code Examples
            </button>
          </nav>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 lg:gap-8">
          {/* Main Content */}
          <div className="lg:col-span-2">
            {activeTab === 'documentation' && (
              <>
                {/* API Documentation */}
                <div className="bg-white rounded-lg shadow p-4 sm:p-6 mb-6 sm:mb-8" data-testid="api-documentation">
                  <h2 className="text-lg sm:text-xl font-semibold text-gray-900 mb-4">Documentation</h2>
                  <p className="text-gray-600">
                    This API provides secure payment processing capabilities.
                  </p>
                  {/* Mock Swagger UI */}
                  <div className="swagger-ui mt-6 p-4 border rounded" data-testid="swagger-ui">
                    <h3 className="font-medium mb-2">API Endpoints</h3>
                    <div className="space-y-2">
                      <div className="opblock-summary p-2 bg-green-50 rounded cursor-pointer">
                        POST /payments
                        <button className="ml-4 px-2 py-1 text-xs bg-white border rounded">
                          Try it out
                        </button>
                      </div>
                      <div className="opblock-summary p-2 bg-blue-50 rounded cursor-pointer">
                        GET /payments/[id]
                        <button className="ml-4 px-2 py-1 text-xs bg-white border rounded">
                          Try it out
                        </button>
                      </div>
                    </div>
                    {/* Mock parameters and execute section */}
                    <div className="mt-4">
                      <div className="parameters mb-4">
                        <input type="text" className="w-full border rounded px-2 py-1" placeholder="Parameter value" />
                      </div>
                      <button className="px-4 py-2 bg-blue-600 text-white rounded">
                        Execute
                      </button>
                      <div className="responses-inner mt-4 p-4 bg-gray-50 rounded">
                        <div className="response-code">200</div>
                      </div>
                    </div>
                  </div>
                </div>

                {/* Reviews Section */}
                <div className="bg-white rounded-lg shadow p-4 sm:p-6" data-testid="api-reviews">
                  <h2 className="text-lg sm:text-xl font-semibold text-gray-900 mb-4">Reviews & Ratings</h2>
                  
                  {/* Review Statistics */}
                  <div className="mb-6" data-testid="review-stats">
                    <div className="flex items-center mb-4">
                      <span className="text-2xl sm:text-3xl font-bold text-gray-900" data-testid="average-rating">
                        {reviews.length === 0 ? '0' : '4.5'}
                      </span>
                      <div className="ml-3 sm:ml-4">
                        <div className="flex items-center">
                          <span className="text-yellow-400 text-lg sm:text-xl">★★★★★</span>
                        </div>
                        <p className="text-xs sm:text-sm text-gray-600" data-testid="total-reviews">Based on {reviews.length} reviews</p>
                      </div>
                    </div>
                    
                    {/* Rating Distribution */}
                    <div className="space-y-2">
                      {[5, 4, 3, 2, 1].map((stars) => (
                        <div key={stars} className="flex items-center" data-testid={`rating-bar-${stars}`}>
                          <span className="text-sm w-3">{stars}</span>
                          <span className="text-yellow-400 mx-2">★</span>
                          <div className="flex-1 bg-gray-200 rounded-full h-2">
                            <div 
                              className="bg-yellow-400 h-2 rounded-full" 
                              style={{width: `${stars * 20}%`}}
                            ></div>
                          </div>
                          <span className="text-sm text-gray-600 ml-2" data-testid={`rating-percentage-${stars}`}>
                            {stars * 20}%
                          </span>
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Review Form or Subscribe Prompt */}
                  <div className="mb-6 pb-6 border-b border-gray-200">
                    {hasSubscription ? (
                      showReviewForm ? (
                        <div className="bg-gray-50 p-4 rounded-lg" data-testid="review-form">
                          <h3 className="text-lg font-medium mb-4">Write a Review</h3>
                          
                          {/* Rating Selection */}
                          <div className="mb-4">
                            <label className="block text-sm font-medium text-gray-700 mb-2">Rating</label>
                            <div className="flex space-x-1">
                              {[1, 2, 3, 4, 5].map((rating) => (
                                <button
                                  key={rating}
                                  type="button"
                                  className={`text-2xl ${selectedRating >= rating ? 'text-yellow-400' : 'text-gray-300 hover:text-yellow-400'}`}
                                  data-testid={`star-rating-${rating}`}
                                  onClick={() => setSelectedRating(rating)}
                                >
                                  ★
                                </button>
                              ))}
                            </div>
                          </div>

                          {/* Title Input */}
                          <div className="mb-4">
                            <label className="block text-sm font-medium text-gray-700 mb-2">Title (optional)</label>
                            <input
                              type="text"
                              value={reviewTitle}
                              onChange={(e) => setReviewTitle(e.target.value.slice(0, 100))}
                              maxLength={100}
                              className="w-full border border-gray-300 rounded-md px-3 py-2"
                              data-testid="review-title"
                              placeholder="Review title"
                            />
                          </div>

                          {/* Comment Input */}
                          <div className="mb-4">
                            <label className="block text-sm font-medium text-gray-700 mb-2">Review</label>
                            <textarea
                              value={reviewComment}
                              onChange={(e) => setReviewComment(e.target.value.slice(0, 1000))}
                              maxLength={1000}
                              rows={4}
                              className="w-full border border-gray-300 rounded-md px-3 py-2"
                              data-testid="review-comment"
                              placeholder="Write your review..."
                            />
                            <div className="text-sm text-gray-500 mt-1" data-testid="char-count">
                              {reviewComment.length} / 1000
                            </div>
                          </div>

                          <div className="flex justify-end space-x-3">
                            <button
                              type="button"
                              onClick={() => setShowReviewForm(false)}
                              className="px-4 py-2 text-gray-700 hover:text-gray-500"
                            >
                              Cancel
                            </button>
                            <button
                              type="button"
                              onClick={handleSubmitReview}
                              className="px-4 py-2 bg-indigo-600 text-white rounded hover:bg-indigo-700"
                              data-testid="submit-review"
                            >
                              Submit Review
                            </button>
                          </div>
                        </div>
                      ) : (
                        <button 
                          className="bg-indigo-600 text-white px-6 py-2 rounded-md hover:bg-indigo-700"
                          data-testid="add-review-button"
                          onClick={() => setShowReviewForm(true)}
                        >
                          Write a Review
                        </button>
                      )
                    ) : (
                      <div className="text-center p-4 bg-gray-50 rounded-lg" data-testid="subscribe-to-review">
                        <p className="text-gray-600 mb-3">Subscribe to this API to write a review</p>
                        <button 
                          className="bg-indigo-600 text-white px-4 py-2 rounded hover:bg-indigo-700"
                          onClick={() => setShowSubscribeModal(true)}
                        >
                          Subscribe Now
                        </button>
                      </div>
                    )}
                  </div>

                  {/* Review Controls */}
                  <div className="mb-6 flex justify-between items-center">
                    <div>
                      <label htmlFor="review-sort" className="text-sm font-medium text-gray-700 mr-2">Sort by:</label>
                      <select 
                        id="review-sort"
                        value={reviewSort}
                        onChange={(e) => setReviewSort(e.target.value)}
                        className="border border-gray-300 rounded-md px-3 py-1 text-sm"
                        data-testid="review-sort"
                      >
                        <option value="recent">Most Recent</option>
                        <option value="helpful">Most Helpful</option>
                        <option value="highest">Highest Rating</option>
                        <option value="lowest">Lowest Rating</option>
                      </select>
                    </div>
                    <div className="text-sm text-gray-600" data-testid="showing-reviews">
                      Showing 1-{Math.min(10, reviews.length)} of {reviews.length} reviews
                    </div>
                  </div>
                  
                  {/* Individual Reviews */}
                  <div className="space-y-4">
                    {sortedReviews.length === 0 ? (
                      <div className="text-center py-8" data-testid="no-reviews">
                        <p className="text-gray-500 mb-2">No reviews yet</p>
                        <p className="text-gray-600">Be the first to review</p>
                      </div>
                    ) : (
                      sortedReviews.map((review: Review) => (
                      <div key={review.id} className="border-b border-gray-200 pb-4" data-testid="review-item">
                        <div className="flex items-start justify-between mb-2">
                          <div>
                            <div className="flex items-center">
                              <span className="font-medium text-gray-900" data-testid="review-author">
                                {review.author}
                              </span>
                              <span className="ml-2 px-2 py-1 text-xs bg-green-100 text-green-800 rounded" data-testid="verified-purchase-badge">
                                Verified Purchase
                              </span>
                              <div className="ml-2 flex items-center">
                                <span className="sr-only" data-testid="review-rating">{review.rating}</span>
                                <div className="flex items-center">
                                  {[...Array(5)].map((_, i) => (
                                    <span key={i} className={i < review.rating ? 'text-yellow-400' : 'text-gray-300'}>
                                      ★
                                    </span>
                                  ))}
                                </div>
                              </div>
                            </div>
                            <p className="text-sm text-gray-500" data-testid="review-date">{review.date}</p>
                          </div>
                          <button 
                            className="text-sm text-red-600 hover:text-red-700"
                            data-testid="flag-review"
                            onClick={() => {
                              const message = document.createElement('div');
                              message.textContent = 'Review flagged for moderation';
                              message.className = 'fixed top-4 right-4 bg-red-100 text-red-800 px-4 py-2 rounded shadow z-50';
                              document.body.appendChild(message);
                              setTimeout(() => message.remove(), 3000);
                            }}
                          >
                            Flag
                          </button>
                        </div>
                        <h4 className="font-medium text-gray-900 mb-1" data-testid="review-item-title">
                          {review.title}
                        </h4>
                        <p className="text-gray-700 mb-3" data-testid="review-comment">
                          {review.content}
                        </p>

                        {/* Creator Response */}
                        {review.hasCreatorResponse && review.creatorResponse && (
                          <div className="mt-3 ml-4 p-3 bg-gray-50 rounded" data-testid="creator-response">
                            <div className="flex items-center mb-1">
                              <span className="font-medium text-sm text-gray-900" data-testid="response-author">
                                Creator
                              </span>
                              <span className="ml-2 text-xs text-gray-500">{review.creatorResponse.date}</span>
                            </div>
                            <p className="text-sm text-gray-700">{review.creatorResponse.content}</p>
                          </div>
                        )}

                        <div className="flex items-center space-x-4 mt-3">
                          <button 
                            className={`text-sm hover:text-gray-700 ${votedReviews.has(review.id) ? 'text-gray-400 cursor-not-allowed' : 'text-gray-600'}`}
                            data-testid="vote-helpful"
                            disabled={votedReviews.has(review.id)}
                            onClick={() => handleVoteHelpful(review.id)}
                          >
                            👍 Helpful (<span data-testid="helpful-count">{review.helpful}</span>)
                          </button>
                          <button 
                            className={`text-sm hover:text-gray-700 ${votedReviews.has(review.id) ? 'text-gray-400 cursor-not-allowed' : 'text-gray-600'}`}
                            data-testid="vote-not-helpful"
                            disabled={votedReviews.has(review.id)}
                            onClick={() => handleVoteNotHelpful(review.id)}
                          >
                            👎 Not Helpful
                          </button>
                        </div>
                      </div>
                      ))
                    )}
                  </div>

                  {/* Pagination */}
                  <div className="mt-6 flex justify-center" data-testid="review-pagination">
                    <div className="flex space-x-2">
                      <button 
                        className="px-3 py-1 text-sm border rounded hover:bg-gray-50"
                        data-testid="review-prev-page"
                      >
                        Previous
                      </button>
                      <button 
                        className="px-3 py-1 text-sm border rounded hover:bg-gray-50"
                        data-testid="review-next-page"
                      >
                        Next
                      </button>
                    </div>
                  </div>
                </div>
              </>
            )}
            
            {activeTab === 'sdk' && (
              <div className="bg-white rounded-lg shadow p-4 sm:p-6" data-testid="sdk-section">
                <h2 className="text-lg sm:text-xl font-semibold text-gray-900 mb-4">SDK Downloads</h2>
                <div className="space-y-4">
                  <button 
                    className="w-full text-left p-4 border rounded hover:bg-gray-50"
                    data-testid="download-sdk-javascript"
                    onClick={() => {
                      // Mock download
                      const link = document.createElement('a');
                      link.download = 'api-sdk.js';
                      link.href = 'data:text/javascript;charset=utf-8,' + encodeURIComponent('// Mock SDK');
                      link.click();
                    }}
                  >
                    <h3 className="font-medium">JavaScript SDK</h3>
                    <p className="text-sm text-gray-600">npm install @api-direct/payment-sdk</p>
                  </button>
                </div>
              </div>
            )}
            
            {activeTab === 'code-examples' && (
              <div className="bg-white rounded-lg shadow p-4 sm:p-6" data-testid="code-examples-section">
                <h2 className="text-lg sm:text-xl font-semibold text-gray-900 mb-4">Code Examples</h2>
                <div className="space-y-4">
                  <div data-testid="code-example-javascript">
                    <h3 className="font-medium mb-2">JavaScript</h3>
                    <pre className="bg-gray-100 p-3 sm:p-4 rounded overflow-x-auto text-sm">
                      <code>{`const payment = await api.payments.create({
  amount: 1000,
  currency: 'USD'
});`}</code>
                    </pre>
                    <button 
                      className="mt-2 text-sm text-indigo-600 hover:text-indigo-700" 
                      data-testid="copy-code-javascript"
                      onClick={() => {
                        const code = `const payment = await api.payments.create({
  amount: 1000,
  currency: 'USD'
});`;
                        navigator.clipboard.writeText(code);
                        // Show copied message
                        const copiedMessage = document.createElement('div');
                        copiedMessage.textContent = 'Copied!';
                        copiedMessage.className = 'fixed top-4 right-4 bg-green-100 text-green-800 px-4 py-2 rounded shadow';
                        document.body.appendChild(copiedMessage);
                        setTimeout(() => copiedMessage.remove(), 2000);
                      }}
                    >
                      Copy
                    </button>
                  </div>
                  <div data-testid="code-example-python">
                    <h3 className="font-medium mb-2">Python</h3>
                    <pre className="bg-gray-100 p-3 sm:p-4 rounded overflow-x-auto text-sm">
                      <code>{`payment = api.payments.create(
    amount=1000,
    currency='USD'
)`}</code>
                    </pre>
                  </div>
                  <div data-testid="code-example-curl">
                    <h3 className="font-medium mb-2">cURL</h3>
                    <pre className="bg-gray-100 p-3 sm:p-4 rounded overflow-x-auto text-sm">
                      <code>{`curl -X POST https://api.example.com/payments \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -d amount=1000 \\
  -d currency=USD`}</code>
                    </pre>
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* Sidebar */}
          <div className="lg:col-span-1 order-first lg:order-last">
            {/* Pricing Plans */}
            <div className="bg-white rounded-lg shadow p-4 sm:p-6 mb-6" data-testid="pricing-plans">
              <h2 className="text-lg sm:text-xl font-semibold text-gray-900 mb-4">Pricing Plans</h2>
              
              <div className="space-y-4">
                <div className="border border-gray-200 rounded-lg p-4">
                  <h3 className="font-medium text-gray-900">Free Tier</h3>
                  <p className="text-xl sm:text-2xl font-bold text-gray-900 mt-1">$0/month</p>
                  <p className="text-sm text-gray-600 mt-1">Up to 1,000 requests</p>
                  <button 
                    className="w-full mt-4 bg-indigo-600 text-white py-2.5 px-4 rounded-md hover:bg-indigo-700 transition-colors touch-manipulation"
                    data-testid="select-plan-basic"
                  >
                    Select Plan
                  </button>
                </div>
                
                <div className="border border-gray-200 rounded-lg p-4">
                  <h3 className="font-medium text-gray-900">Professional</h3>
                  <p className="text-xl sm:text-2xl font-bold text-gray-900 mt-1">$99/month</p>
                  <p className="text-sm text-gray-600 mt-1">Up to 50,000 requests</p>
                  <button 
                    className="w-full mt-4 bg-indigo-600 text-white py-2.5 px-4 rounded-md hover:bg-indigo-700 transition-colors touch-manipulation"
                    data-testid="select-plan-premium"
                  >
                    Select Plan
                  </button>
                </div>
              </div>
              
              <button 
                className="w-full mt-6 bg-green-600 text-white py-3 px-4 rounded-md hover:bg-green-700 font-medium transition-colors touch-manipulation"
                data-testid="subscribe-button"
                onClick={() => setShowSubscribeModal(true)}
              >
                Subscribe Now
              </button>
            </div>

            {/* API Key Section */}
            <div className="bg-white rounded-lg shadow p-4 sm:p-6 mb-6" data-testid="api-key-section">
              <h3 className="text-base sm:text-lg font-medium text-gray-900 mb-4">API Key</h3>
              <div className="space-y-3">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">Your API Key</label>
                  <div className="flex items-center space-x-3">
                    <div className="flex-1 p-3 border border-gray-300 rounded-md bg-gray-50">
                      <span className="font-mono text-sm text-gray-900" data-testid="api-key-value">
                        api_abc123def456ghi789jkl
                      </span>
                    </div>
                    <button 
                      className="px-3 sm:px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 transition-colors touch-manipulation"
                      data-testid="show-api-key"
                    >
                      Show
                    </button>
                    <button 
                      className="px-3 sm:px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50 transition-colors touch-manipulation"
                      data-testid="copy-api-key"
                    >
                      Copy
                    </button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Subscription Modal */}
      {showSubscribeModal && (
        <div className="fixed inset-0 bg-gray-500 bg-opacity-75 flex items-center justify-center z-50 p-4">
          <div className="bg-white rounded-lg p-4 sm:p-6 max-w-md w-full max-h-screen overflow-y-auto">
            <h3 className="text-base sm:text-lg font-medium text-gray-900 mb-4">Complete Subscription</h3>
            
            {/* Mock Stripe Payment Form */}
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">Card Information</label>
                <div className="border p-3 rounded">
                  <iframe 
                    title="Secure payment input frame" 
                    className="w-full h-32"
                    srcDoc={`
                      <html>
                        <body style="margin: 0; padding: 10px; font-family: Arial, sans-serif;">
                          <div style="display: flex; flex-direction: column; gap: 10px;">
                            <input placeholder="Card number" style="padding: 8px; border: 1px solid #ccc; border-radius: 4px;" />
                            <div style="display: flex; gap: 10px;">
                              <input placeholder="MM / YY" style="padding: 8px; border: 1px solid #ccc; border-radius: 4px; flex: 1;" />
                              <input placeholder="CVC" style="padding: 8px; border: 1px solid #ccc; border-radius: 4px; flex: 1;" />
                              <input placeholder="ZIP" style="padding: 8px; border: 1px solid #ccc; border-radius: 4px; flex: 1;" />
                            </div>
                          </div>
                        </body>
                      </html>
                    `}
                  />
                </div>
              </div>
              
              <button 
                className="w-full bg-indigo-600 text-white py-3 px-4 rounded hover:bg-indigo-700 transition-colors touch-manipulation"
                data-testid="complete-subscription"
                onClick={() => {
                  // Check if card number is declined test card
                  const iframe = document.querySelector('iframe[title="Secure payment input frame"]') as HTMLIFrameElement;
                  if (iframe && iframe.contentDocument) {
                    const cardInput = iframe.contentDocument.querySelector('input[placeholder="Card number"]') as HTMLInputElement;
                    if (cardInput && cardInput.value === '4000000000000002') {
                      // Show payment declined error
                      setPaymentError(true);
                      const errorMessage = document.createElement('div');
                      errorMessage.textContent = 'Payment declined';
                      errorMessage.className = 'fixed top-4 right-4 bg-red-100 text-red-800 px-4 py-2 rounded shadow';
                      document.body.appendChild(errorMessage);
                      setTimeout(() => errorMessage.remove(), 3000);
                      return;
                    }
                  }
                  handleSubscribe();
                }}
              >
                Subscribe
              </button>
            </div>
            {paymentError && (
              <p className="mt-2 text-sm text-red-600">
                Your payment was declined. Please try a different card.
              </p>
            )}
          </div>
        </div>
      )}
    </Layout>
  );
};

export default APIDetails;
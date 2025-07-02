import { useState, useEffect } from 'react';
import { useRouter } from 'next/router';

interface Review {
  id: string;
  rating: number;
  title: string;
  comment: string;
  author: string;
  date: string;
  verified: boolean;
  helpful: number;
  notHelpful: number;
  response?: {
    text: string;
    author: string;
    date: string;
  };
}

export default function CreatorReviews() {
  const router = useRouter();
  const { apiId } = router.query;
  const [reviews, setReviews] = useState<Review[]>([]);
  const [respondingTo, setRespondingTo] = useState<string | null>(null);
  const [responseText, setResponseText] = useState('');

  useEffect(() => {
    const savedReviews = localStorage.getItem('reviewsData');
    if (savedReviews) {
      const reviewsData = JSON.parse(savedReviews);
      setReviews(reviewsData[apiId as string] || []);
    }
  }, [apiId]);

  const handleRespondToReview = (reviewId: string) => {
    setRespondingTo(reviewId);
    setResponseText('');
  };

  const handleSubmitResponse = (reviewId: string) => {
    if (responseText.trim()) {
      const updatedReviews = reviews.map(review => {
        if (review.id === reviewId) {
          return {
            ...review,
            response: {
              text: responseText.trim(),
              author: 'Creator',
              date: new Date().toISOString().split('T')[0]
            }
          };
        }
        return review;
      });

      setReviews(updatedReviews);

      // Save to localStorage
      const savedReviews = localStorage.getItem('reviewsData');
      const reviewsData = savedReviews ? JSON.parse(savedReviews) : {};
      reviewsData[apiId as string] = updatedReviews;
      localStorage.setItem('reviewsData', JSON.stringify(reviewsData));

      setRespondingTo(null);
      setResponseText('');
    }
  };

  const handleCancelResponse = () => {
    setRespondingTo(null);
    setResponseText('');
  };

  return (
    <div className="max-w-4xl mx-auto p-6">
      <h1 className="text-3xl font-bold mb-6">API Reviews</h1>
      
      {reviews.length === 0 ? (
        <div className="text-center py-8">
          <p className="text-gray-500">No reviews yet for this API.</p>
        </div>
      ) : (
        <div className="space-y-6">
          {reviews.map((review) => (
            <div key={review.id} className="border border-gray-200 rounded-lg p-6" data-testid="review-item">
              <div className="flex items-start justify-between mb-4">
                <div>
                  <div className="flex items-center gap-2 mb-2">
                    <div className="flex text-yellow-400" data-testid="review-rating">
                      {Array.from({ length: 5 }, (_, i) => (
                        <span key={i}>
                          {i < review.rating ? '★' : '☆'}
                        </span>
                      ))}
                    </div>
                    <span className="font-semibold">{review.title}</span>
                  </div>
                  <div className="flex items-center gap-4 text-sm text-gray-600">
                    <span data-testid="review-author">{review.author}</span>
                    <span data-testid="review-date">{review.date}</span>
                    {review.verified && (
                      <span className="bg-green-100 text-green-800 px-2 py-1 rounded text-xs" data-testid="verified-purchase-badge">
                        Verified Purchase
                      </span>
                    )}
                  </div>
                </div>
              </div>

              <p className="text-gray-700 mb-4" data-testid="review-comment">{review.comment}</p>

              {/* Creator Response */}
              {review.response && (
                <div className="bg-blue-50 border-l-4 border-blue-200 p-4 mt-4" data-testid="creator-response">
                  <div className="flex items-center gap-2 mb-2">
                    <span className="font-semibold text-blue-700" data-testid="response-author">Creator</span>
                    <span className="text-sm text-blue-600">{review.response.date}</span>
                  </div>
                  <p className="text-blue-800">{review.response.text}</p>
                </div>
              )}

              {/* Response Form */}
              {respondingTo === review.id ? (
                <div className="mt-4 p-4 border-t border-gray-200">
                  <h4 className="font-semibold mb-3">Respond to this review</h4>
                  <textarea
                    value={responseText}
                    onChange={(e) => setResponseText(e.target.value)}
                    placeholder="Write your response..."
                    className="w-full border border-gray-300 rounded-md px-3 py-2 resize-none"
                    rows={4}
                    data-testid="response-text"
                  />
                  <div className="flex gap-2 mt-3">
                    <button
                      onClick={() => handleSubmitResponse(review.id)}
                      disabled={!responseText.trim()}
                      className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 disabled:bg-gray-300"
                      data-testid="submit-response"
                    >
                      Submit Response
                    </button>
                    <button
                      onClick={handleCancelResponse}
                      className="bg-gray-300 text-gray-700 px-4 py-2 rounded-md hover:bg-gray-400"
                    >
                      Cancel
                    </button>
                  </div>
                </div>
              ) : (
                !review.response && (
                  <div className="mt-4 pt-4 border-t border-gray-200">
                    <button
                      onClick={() => handleRespondToReview(review.id)}
                      className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700"
                      data-testid="respond-button"
                    >
                      Respond to Review
                    </button>
                  </div>
                )
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
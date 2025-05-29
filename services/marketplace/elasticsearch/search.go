package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/elastic/go-elasticsearch/v8"
)

// SearchService handles all search operations
type SearchService struct {
	client *elasticsearch.Client
}

// NewSearchService creates a new search service
func NewSearchService(client *elasticsearch.Client) *SearchService {
	return &SearchService{client: client}
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query       string            `json:"query"`
	Category    string            `json:"category"`
	Tags        []string          `json:"tags"`
	PriceRange  string            `json:"price_range"`
	MinRating   float32           `json:"min_rating"`
	HasFreeTier *bool             `json:"has_free_tier"`
	SortBy      string            `json:"sort_by"`
	Page        int               `json:"page"`
	Limit       int               `json:"limit"`
}

// SearchResponse represents search results
type SearchResponse struct {
	APIs   []APIDocument `json:"apis"`
	Total  int64         `json:"total"`
	Facets Facets        `json:"facets"`
}

// APIDocument represents an API in Elasticsearch
type APIDocument struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Tags        []string `json:"tags"`
	CreatorID   string  `json:"creator_id"`
	CreatorName string  `json:"creator_name"`
	Pricing     struct {
		HasFreeTier bool    `json:"has_free_tier"`
		MinPrice    float64 `json:"min_price"`
		MaxPrice    float64 `json:"max_price"`
		PriceRange  string  `json:"price_range"`
	} `json:"pricing"`
	Stats struct {
		AverageRating       float32 `json:"average_rating"`
		TotalReviews        int     `json:"total_reviews"`
		TotalSubscriptions  int     `json:"total_subscriptions"`
		MonthlyCalls        int64   `json:"monthly_calls"`
	} `json:"stats"`
	IsPublished bool    `json:"is_published"`
	BoostScore  float32 `json:"boost_score"`
}

// Facets represents aggregation results
type Facets struct {
	Categories   map[string]int64 `json:"categories"`
	PriceRanges  map[string]int64 `json:"price_ranges"`
	Tags         map[string]int64 `json:"tags"`
	RatingRanges map[string]int64 `json:"rating_ranges"`
}

// Search performs an API search
func (s *SearchService) Search(req SearchRequest) (*SearchResponse, error) {
	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}

	// Build query
	query := s.buildQuery(req)

	// Build request body
	body := map[string]interface{}{
		"query": query,
		"from":  (req.Page - 1) * req.Limit,
		"size":  req.Limit,
		"sort":  s.buildSort(req.SortBy),
		"aggs":  s.buildAggregations(),
	}

	// Execute search
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return nil, fmt.Errorf("error encoding query: %w", err)
	}

	res, err := s.client.Search(
		s.client.Search.WithContext(context.Background()),
		s.client.Search.WithIndex("apis"),
		s.client.Search.WithBody(&buf),
		s.client.Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("error executing search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search error: %s", body)
	}

	// Parse response
	var result struct {
		Hits struct {
			Total struct {
				Value int64 `json:"value"`
			} `json:"total"`
			Hits []struct {
				Source APIDocument `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
		Aggregations map[string]struct {
			Buckets []struct {
				Key      string `json:"key"`
				DocCount int64  `json:"doc_count"`
			} `json:"buckets"`
		} `json:"aggregations"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	// Extract APIs
	apis := make([]APIDocument, 0, len(result.Hits.Hits))
	for _, hit := range result.Hits.Hits {
		apis = append(apis, hit.Source)
	}

	// Extract facets
	facets := s.extractFacets(result.Aggregations)

	return &SearchResponse{
		APIs:   apis,
		Total:  result.Hits.Total.Value,
		Facets: facets,
	}, nil
}

// buildQuery constructs the Elasticsearch query
func (s *SearchService) buildQuery(req SearchRequest) map[string]interface{} {
	must := []interface{}{}
	filter := []interface{}{}

	// Full-text search
	if req.Query != "" {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query": req.Query,
				"fields": []string{
					"name^3",
					"description^2",
					"tags.text",
					"creator_name",
				},
				"type":      "best_fields",
				"fuzziness": "AUTO",
			},
		})
	}

	// Always filter for published APIs
	filter = append(filter, map[string]interface{}{
		"term": map[string]interface{}{
			"is_published": true,
		},
	})

	// Category filter
	if req.Category != "" {
		filter = append(filter, map[string]interface{}{
			"term": map[string]interface{}{
				"category": req.Category,
			},
		})
	}

	// Tags filter
	if len(req.Tags) > 0 {
		filter = append(filter, map[string]interface{}{
			"terms": map[string]interface{}{
				"tags": req.Tags,
			},
		})
	}

	// Price range filter
	if req.PriceRange != "" {
		filter = append(filter, map[string]interface{}{
			"term": map[string]interface{}{
				"pricing.price_range": req.PriceRange,
			},
		})
	}

	// Free tier filter
	if req.HasFreeTier != nil {
		filter = append(filter, map[string]interface{}{
			"term": map[string]interface{}{
				"pricing.has_free_tier": *req.HasFreeTier,
			},
		})
	}

	// Rating filter
	if req.MinRating > 0 {
		filter = append(filter, map[string]interface{}{
			"range": map[string]interface{}{
				"stats.average_rating": map[string]interface{}{
					"gte": req.MinRating,
				},
			},
		})
	}

	// Build final query
	query := map[string]interface{}{
		"bool": map[string]interface{}{},
	}

	if len(must) > 0 {
		query["bool"].(map[string]interface{})["must"] = must
	}
	if len(filter) > 0 {
		query["bool"].(map[string]interface{})["filter"] = filter
	}

	// If no conditions, match all
	if len(must) == 0 && len(filter) == 1 {
		return map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": filter,
			},
		}
	}

	return query
}

// buildSort constructs the sort criteria
func (s *SearchService) buildSort(sortBy string) []interface{} {
	switch sortBy {
	case "popularity":
		return []interface{}{
			map[string]interface{}{
				"stats.total_subscriptions": map[string]interface{}{
					"order": "desc",
				},
			},
		}
	case "rating":
		return []interface{}{
			map[string]interface{}{
				"stats.average_rating": map[string]interface{}{
					"order": "desc",
				},
			},
		}
	case "newest":
		return []interface{}{
			map[string]interface{}{
				"created_at": map[string]interface{}{
					"order": "desc",
				},
			},
		}
	case "relevance":
		fallthrough
	default:
		return []interface{}{
			"_score",
			map[string]interface{}{
				"boost_score": map[string]interface{}{
					"order": "desc",
				},
			},
		}
	}
}

// buildAggregations constructs the facet aggregations
func (s *SearchService) buildAggregations() map[string]interface{} {
	return map[string]interface{}{
		"categories": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "category",
				"size":  20,
			},
		},
		"price_ranges": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "pricing.price_range",
				"size":  10,
			},
		},
		"tags": map[string]interface{}{
			"terms": map[string]interface{}{
				"field": "tags",
				"size":  30,
			},
		},
		"rating_ranges": map[string]interface{}{
			"range": map[string]interface{}{
				"field": "stats.average_rating",
				"ranges": []interface{}{
					map[string]interface{}{"from": 4.5, "key": "4.5+"},
					map[string]interface{}{"from": 4.0, "to": 4.5, "key": "4.0-4.5"},
					map[string]interface{}{"from": 3.0, "to": 4.0, "key": "3.0-4.0"},
					map[string]interface{}{"to": 3.0, "key": "Below 3.0"},
				},
			},
		},
	}
}

// extractFacets extracts facet data from aggregations
func (s *SearchService) extractFacets(aggs map[string]struct {
	Buckets []struct {
		Key      string `json:"key"`
		DocCount int64  `json:"doc_count"`
	} `json:"buckets"`
}) Facets {
	facets := Facets{
		Categories:   make(map[string]int64),
		PriceRanges:  make(map[string]int64),
		Tags:         make(map[string]int64),
		RatingRanges: make(map[string]int64),
	}

	// Extract categories
	if cats, ok := aggs["categories"]; ok {
		for _, bucket := range cats.Buckets {
			facets.Categories[bucket.Key] = bucket.DocCount
		}
	}

	// Extract price ranges
	if prices, ok := aggs["price_ranges"]; ok {
		for _, bucket := range prices.Buckets {
			facets.PriceRanges[bucket.Key] = bucket.DocCount
		}
	}

	// Extract tags
	if tags, ok := aggs["tags"]; ok {
		for _, bucket := range tags.Buckets {
			facets.Tags[bucket.Key] = bucket.DocCount
		}
	}

	// Extract rating ranges
	if ratings, ok := aggs["rating_ranges"]; ok {
		for _, bucket := range ratings.Buckets {
			facets.RatingRanges[bucket.Key] = bucket.DocCount
		}
	}

	return facets
}

// GetSuggestions returns search suggestions based on partial input
func (s *SearchService) GetSuggestions(prefix string) ([]string, error) {
	query := map[string]interface{}{
		"suggest": map[string]interface{}{
			"api-suggest": map[string]interface{}{
				"prefix": prefix,
				"completion": map[string]interface{}{
					"field": "name.suggest",
					"size":  10,
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		return nil, fmt.Errorf("error encoding query: %w", err)
	}

	res, err := s.client.Search(
		s.client.Search.WithContext(context.Background()),
		s.client.Search.WithIndex("apis"),
		s.client.Search.WithBody(&buf),
	)
	if err != nil {
		return nil, fmt.Errorf("error executing search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search error: %s", body)
	}

	// Parse response
	var result struct {
		Suggest struct {
			APISuggest []struct {
				Options []struct {
					Text string `json:"text"`
				} `json:"options"`
			} `json:"api-suggest"`
		} `json:"suggest"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	// Extract suggestions
	suggestions := []string{}
	if len(result.Suggest.APISuggest) > 0 {
		for _, option := range result.Suggest.APISuggest[0].Options {
			suggestions = append(suggestions, option.Text)
		}
	}

	return suggestions, nil
}

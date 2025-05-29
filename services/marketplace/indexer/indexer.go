package indexer

import (
	"fmt"
	"log"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"marketplace/elasticsearch"
	"marketplace/store"
)

// APIIndexer handles indexing APIs to Elasticsearch
type APIIndexer struct {
	esClient *elasticsearch.Client
	apiStore *store.APIStore
}

// NewAPIIndexer creates a new API indexer
func NewAPIIndexer(esClient *elasticsearch.Client, apiStore *store.APIStore) *APIIndexer {
	return &APIIndexer{
		esClient: esClient,
		apiStore: apiStore,
	}
}

// IndexAPI indexes a single API
func (idx *APIIndexer) IndexAPI(apiID string) error {
	// Fetch API from database
	api, err := idx.apiStore.GetAPI(apiID)
	if err != nil {
		return fmt.Errorf("error fetching API: %w", err)
	}

	// Transform to Elasticsearch document
	doc := idx.transformToESDocument(api)

	// Index to Elasticsearch
	err = elasticsearch.IndexDocument(idx.esClient, "apis", apiID, doc)
	if err != nil {
		return fmt.Errorf("error indexing document: %w", err)
	}

	log.Printf("Successfully indexed API %s", apiID)
	return nil
}

// DeleteAPI removes an API from the index
func (idx *APIIndexer) DeleteAPI(apiID string) error {
	err := elasticsearch.DeleteDocument(idx.esClient, "apis", apiID)
	if err != nil {
		return fmt.Errorf("error deleting document: %w", err)
	}

	log.Printf("Successfully deleted API %s from index", apiID)
	return nil
}

// ReindexAll reindexes all published APIs
func (idx *APIIndexer) ReindexAll() error {
	log.Println("Starting full reindex...")

	// Fetch all published APIs
	apis, err := idx.apiStore.GetAllPublishedAPIs()
	if err != nil {
		return fmt.Errorf("error fetching APIs: %w", err)
	}

	// Transform to documents
	documents := make(map[string]interface{})
	for _, api := range apis {
		documents[api.ID] = idx.transformToESDocument(api)
	}

	// Bulk index
	err = elasticsearch.BulkIndex(idx.esClient, "apis", documents)
	if err != nil {
		return fmt.Errorf("error bulk indexing: %w", err)
	}

	log.Printf("Successfully reindexed %d APIs", len(apis))
	return nil
}

// transformToESDocument transforms an API to an Elasticsearch document
func (idx *APIIndexer) transformToESDocument(api *store.API) map[string]interface{} {
	// Calculate pricing info
	hasFreeTier := false
	minPrice := float64(0)
	maxPrice := float64(0)

	for _, plan := range api.PricingPlans {
		if plan.Type == "free" {
			hasFreeTier = true
		}
		
		if plan.MonthlyPrice != nil {
			if minPrice == 0 || *plan.MonthlyPrice < minPrice {
				minPrice = *plan.MonthlyPrice
			}
			if *plan.MonthlyPrice > maxPrice {
				maxPrice = *plan.MonthlyPrice
			}
		}
		
		if plan.PricePerCall != nil {
			// Estimate monthly cost based on 10k calls/month
			estimatedMonthly := *plan.PricePerCall * 10000
			if minPrice == 0 || estimatedMonthly < minPrice {
				minPrice = estimatedMonthly
			}
			if estimatedMonthly > maxPrice {
				maxPrice = estimatedMonthly
			}
		}
	}

	// Determine price range
	priceRange := api.PriceRange()

	// Calculate boost score (can be enhanced with more factors)
	boostScore := float32(1.0)
	if api.AverageRating != nil && *api.AverageRating > 4.0 {
		boostScore += (*api.AverageRating - 4.0) * 0.5
	}
	if api.TotalSubscriptions > 100 {
		boostScore += 0.5
	}
	if api.TotalReviews > 10 {
		boostScore += 0.3
	}

	// Build document
	doc := map[string]interface{}{
		"id":           api.ID,
		"name":         api.Name,
		"description":  api.Description,
		"category":     api.Category,
		"tags":         api.Tags,
		"creator_id":   api.UserID,
		"creator_name": api.CreatorName,
		"pricing": map[string]interface{}{
			"has_free_tier": hasFreeTier,
			"min_price":     minPrice,
			"max_price":     maxPrice,
			"price_range":   priceRange,
		},
		"stats": map[string]interface{}{
			"average_rating":      api.AverageRating,
			"total_reviews":       api.TotalReviews,
			"total_subscriptions": api.TotalSubscriptions,
			"monthly_calls":       0, // TODO: Calculate from metering data
		},
		"created_at":   api.CreatedAt,
		"updated_at":   api.UpdatedAt,
		"is_published": api.IsPublished,
		"boost_score":  boostScore,
	}

	// Add name to suggest field for autocomplete
	doc["name"] = map[string]interface{}{
		"input": []string{api.Name},
	}

	return doc
}

// IndexOnAPIEvent handles API events for indexing
func (idx *APIIndexer) IndexOnAPIEvent(event APIEvent) error {
	switch event.Type {
	case "api.published":
		return idx.IndexAPI(event.APIID)
	case "api.updated":
		// Only reindex if the API is published
		api, err := idx.apiStore.GetAPI(event.APIID)
		if err != nil {
			return err
		}
		if api.IsPublished {
			return idx.IndexAPI(event.APIID)
		}
	case "api.unpublished":
		return idx.DeleteAPI(event.APIID)
	case "api.deleted":
		return idx.DeleteAPI(event.APIID)
	}
	return nil
}

// APIEvent represents an API lifecycle event
type APIEvent struct {
	Type      string    `json:"type"`
	APIID     string    `json:"api_id"`
	Timestamp time.Time `json:"timestamp"`
}

package elasticsearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// NewClient creates a new Elasticsearch client
func NewClient(url string) (*elasticsearch.Client, error) {
	if url == "" {
		url = "http://localhost:9200"
	}

	cfg := elasticsearch.Config{
		Addresses: []string{url},
	}

	return elasticsearch.NewClient(cfg)
}

// InitializeIndices creates the necessary Elasticsearch indices with proper mappings
func InitializeIndices(client *elasticsearch.Client) error {
	// Check if index already exists
	res, err := client.Indices.Exists([]string{"apis"})
	if err != nil {
		return fmt.Errorf("error checking index existence: %w", err)
	}
	defer res.Body.Close()

	// If index exists, skip creation
	if res.StatusCode == 200 {
		log.Println("APIs index already exists")
		return nil
	}

	// Create index with mapping
	mapping := `{
		"mappings": {
			"properties": {
				"id": { "type": "keyword" },
				"name": {
					"type": "text",
					"analyzer": "standard",
					"fields": {
						"keyword": { "type": "keyword" },
						"suggest": { "type": "completion" }
					}
				},
				"description": {
					"type": "text",
					"analyzer": "standard"
				},
				"category": {
					"type": "keyword",
					"fields": {
						"text": { "type": "text" }
					}
				},
				"tags": {
					"type": "keyword",
					"fields": {
						"text": { "type": "text" }
					}
				},
				"creator_id": { "type": "keyword" },
				"creator_name": {
					"type": "text",
					"fields": {
						"keyword": { "type": "keyword" }
					}
				},
				"pricing": {
					"properties": {
						"has_free_tier": { "type": "boolean" },
						"min_price": { "type": "float" },
						"max_price": { "type": "float" },
						"price_range": { "type": "keyword" }
					}
				},
				"stats": {
					"properties": {
						"average_rating": { "type": "float" },
						"total_reviews": { "type": "integer" },
						"total_subscriptions": { "type": "integer" },
						"monthly_calls": { "type": "long" }
					}
				},
				"created_at": { "type": "date" },
				"updated_at": { "type": "date" },
				"is_published": { "type": "boolean" },
				"boost_score": { "type": "float" }
			}
		},
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 0,
			"analysis": {
				"analyzer": {
					"api_analyzer": {
						"tokenizer": "standard",
						"filter": ["lowercase", "api_synonyms"]
					}
				},
				"filter": {
					"api_synonyms": {
						"type": "synonym",
						"synonyms": [
							"api,service,endpoint",
							"auth,authentication,oauth",
							"db,database",
							"ml,machine learning,ai,artificial intelligence"
						]
					}
				}
			}
		}
	}`

	req := esapi.IndicesCreateRequest{
		Index: "apis",
		Body:  strings.NewReader(mapping),
	}

	res, err = req.Do(context.Background(), client)
	if err != nil {
		return fmt.Errorf("error creating index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("error creating index: %s", body)
	}

	log.Println("APIs index created successfully")
	return nil
}

// IndexDocument indexes a single document
func IndexDocument(client *elasticsearch.Client, index string, id string, doc interface{}) error {
	data, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("error marshaling document: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      index,
		DocumentID: id,
		Body:       bytes.NewReader(data),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), client)
	if err != nil {
		return fmt.Errorf("error indexing document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("error response: %s", body)
	}

	return nil
}

// DeleteDocument removes a document from the index
func DeleteDocument(client *elasticsearch.Client, index string, id string) error {
	req := esapi.DeleteRequest{
		Index:      index,
		DocumentID: id,
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), client)
	if err != nil {
		return fmt.Errorf("error deleting document: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("error response: %s", body)
	}

	return nil
}

// BulkIndex indexes multiple documents at once
func BulkIndex(client *elasticsearch.Client, index string, documents map[string]interface{}) error {
	var buf bytes.Buffer

	for id, doc := range documents {
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": index,
				"_id":    id,
			},
		}

		metaData, err := json.Marshal(meta)
		if err != nil {
			return fmt.Errorf("error marshaling metadata: %w", err)
		}

		docData, err := json.Marshal(doc)
		if err != nil {
			return fmt.Errorf("error marshaling document: %w", err)
		}

		buf.Write(metaData)
		buf.WriteByte('\n')
		buf.Write(docData)
		buf.WriteByte('\n')
	}

	req := esapi.BulkRequest{
		Body:    &buf,
		Refresh: "true",
	}

	res, err := req.Do(context.Background(), client)
	if err != nil {
		return fmt.Errorf("error performing bulk request: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("error response: %s", body)
	}

	return nil
}

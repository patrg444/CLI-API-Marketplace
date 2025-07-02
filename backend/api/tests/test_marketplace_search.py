"""
Test marketplace search functionality
"""

import pytest
from unittest.mock import Mock, AsyncMock, patch, MagicMock
from fastapi.testclient import TestClient
import sys
import os
import uuid
from decimal import Decimal
from datetime import datetime

sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

# Mock stripe before importing
mock_stripe = MagicMock()
sys.modules['stripe'] = mock_stripe

# Mock docker before importing
mock_docker = Mock()
sys.modules['docker'] = mock_docker

# Mock the database before importing main
with patch('asyncpg.create_pool', new_callable=AsyncMock) as mock_pool:
    with patch('redis.from_url') as mock_redis:
        mock_pool.return_value = AsyncMock()
        mock_redis.return_value = Mock()
        
        import main
        # Set the global db_pool variable
        main.db_pool = AsyncMock()
        main.payment_manager = Mock()
        main.deployment_manager = Mock()
        
        from main import app

client = TestClient(app)


class TestMarketplaceSearch:
    """Test marketplace search endpoints"""
    
    def test_search_with_query(self):
        """Test search with text query"""
        # Mock database response
        mock_conn = AsyncMock()
        mock_listings = [
            {
                "id": uuid.uuid4(),
                "api_id": uuid.uuid4(),
                "title": "Stripe Payment API",
                "description": "Process payments with Stripe",
                "category": "Financial Services",
                "creator_name": "John Doe",
                "pricing_model": "freemium",
                "price_per_call": Decimal("0.01"),
                "avg_rating": Decimal("4.5"),
                "review_count": 10,
                "monthly_calls": 5000,
                "uptime": Decimal("99.9"),
                "tags": "stripe,payment,finance",
                "featured": True
            }
        ]
        # Set up multiple return values for fetch calls
        # First call: main query results
        # Second call: facet results (if search is provided)
        mock_facets = [
            {"category": "Financial Services", "count": 5},
            {"category": "AI/ML", "count": 3}
        ]
        mock_conn.fetch.side_effect = [mock_listings, mock_facets]
        mock_conn.fetchval.return_value = 1  # Total count
        
        # Mock acquire context manager
        mock_acquire = MagicMock()
        mock_acquire.__aenter__ = AsyncMock(return_value=mock_conn)
        mock_acquire.__aexit__ = AsyncMock(return_value=None)
        main.db_pool.acquire = Mock(return_value=mock_acquire)
        
        # Test search
        response = client.get("/api/apis?search=payment")
        
        # Debug response
        print(f"Response status: {response.status_code}")
        print(f"Response body: {response.text}")
        
        assert response.status_code == 200
        data = response.json()
        assert data["success"] is True
        assert len(data["data"]) == 1
        assert data["data"][0]["name"] == "Stripe Payment API"
        assert data["meta"]["total"] == 1
        
        # Verify search parameter was passed to SQL
        # Check the first fetch call (main query)
        first_call = mock_conn.fetch.call_args_list[0]
        call_args = first_call[0]
        # The parameter should be in the args list
        assert any("%payment%" in str(arg) for arg in call_args)
    
    def test_filter_by_category(self):
        """Test filtering by category"""
        # Mock database response
        mock_conn = AsyncMock()
        mock_conn.fetch.return_value = []
        mock_conn.fetchval.return_value = 0
        
        # Mock acquire context manager
        mock_acquire = MagicMock()
        mock_acquire.__aenter__ = AsyncMock(return_value=mock_conn)
        mock_acquire.__aexit__ = AsyncMock(return_value=None)
        main.db_pool.acquire = Mock(return_value=mock_acquire)
        
        # Test category filter
        response = client.get("/api/apis?category=AI/ML")
        
        assert response.status_code == 200
        data = response.json()
        assert data["success"] is True
        
        # Verify category parameter was passed
        call_args = mock_conn.fetch.call_args[0]
        assert "AI/ML" in call_args  # Category should be in parameters
    
    def test_filter_by_price(self):
        """Test filtering by max price"""
        # Mock database response
        mock_conn = AsyncMock()
        mock_conn.fetch.return_value = []
        mock_conn.fetchval.return_value = 0
        
        # Mock acquire context manager
        mock_acquire = MagicMock()
        mock_acquire.__aenter__ = AsyncMock(return_value=mock_conn)
        mock_acquire.__aexit__ = AsyncMock(return_value=None)
        main.db_pool.acquire = Mock(return_value=mock_acquire)
        
        # Test price filter
        response = client.get("/api/apis?maxPrice=0.05")
        
        assert response.status_code == 200
        data = response.json()
        assert data["success"] is True
        
        # Verify price parameter was passed
        call_args = mock_conn.fetch.call_args[0]
        assert 0.05 in call_args  # Price should be in parameters
    
    def test_sort_by_rating(self):
        """Test sorting by rating"""
        # Mock database response
        mock_conn = AsyncMock()
        mock_conn.fetch.return_value = []
        mock_conn.fetchval.return_value = 0
        
        # Mock acquire context manager
        mock_acquire = MagicMock()
        mock_acquire.__aenter__ = AsyncMock(return_value=mock_conn)
        mock_acquire.__aexit__ = AsyncMock(return_value=None)
        main.db_pool.acquire = Mock(return_value=mock_acquire)
        
        # Test rating sort
        response = client.get("/api/apis?sort=rated")
        
        assert response.status_code == 200
        data = response.json()
        assert data["success"] is True
        
        # Verify ORDER BY clause includes rating
        query = mock_conn.fetch.call_args[0][0]
        assert "AVG(r.rating) DESC" in query
    
    def test_pagination(self):
        """Test pagination parameters"""
        # Mock database response
        mock_conn = AsyncMock()
        mock_conn.fetch.return_value = []
        mock_conn.fetchval.return_value = 50  # Total count
        
        # Mock acquire context manager
        mock_acquire = MagicMock()
        mock_acquire.__aenter__ = AsyncMock(return_value=mock_conn)
        mock_acquire.__aexit__ = AsyncMock(return_value=None)
        main.db_pool.acquire = Mock(return_value=mock_acquire)
        
        # Test pagination
        response = client.get("/api/apis?page=2&limit=10")
        
        assert response.status_code == 200
        data = response.json()
        assert data["success"] is True
        assert data["meta"]["page"] == 2
        assert data["meta"]["limit"] == 10
        assert data["meta"]["totalPages"] == 5
        
        # Verify LIMIT and OFFSET
        call_args = mock_conn.fetch.call_args[0]
        assert 10 in call_args  # LIMIT
        assert 10 in call_args  # OFFSET (page 2 with limit 10)
    
    def test_get_categories(self):
        """Test getting categories endpoint"""
        # Mock database response
        mock_conn = AsyncMock()
        mock_categories = [
            {"id": "AI/ML", "name": "AI/ML", "count": 15},
            {"id": "Financial Services", "name": "Financial Services", "count": 10}
        ]
        mock_conn.fetch.return_value = mock_categories
        
        # Mock acquire context manager
        mock_acquire = MagicMock()
        mock_acquire.__aenter__ = AsyncMock(return_value=mock_conn)
        mock_acquire.__aexit__ = AsyncMock(return_value=None)
        main.db_pool.acquire = Mock(return_value=mock_acquire)
        
        # Test categories
        response = client.get("/api/categories")
        
        assert response.status_code == 200
        data = response.json()
        assert data["success"] is True
        assert len(data["data"]) == 2
        assert data["data"][0]["name"] == "AI/ML"
        assert data["data"][0]["icon"] == "🤖"
        assert data["data"][0]["count"] == 15
    
    def test_get_featured_apis(self):
        """Test getting featured APIs"""
        # Mock database response
        mock_conn = AsyncMock()
        mock_featured = [
            {
                "id": uuid.uuid4(),
                "api_id": uuid.uuid4(),
                "title": "Featured API",
                "description": "A featured API",
                "category": "AI/ML",
                "creator_name": "Jane Doe",
                "pricing_model": "subscription",
                "price_per_call": None,
                "avg_rating": Decimal("5.0"),
                "review_count": 20,
                "monthly_calls": 10000,
                "tags": "ai,ml,featured",
                "created_at": datetime.now()
            }
        ]
        mock_conn.fetch.return_value = mock_featured
        
        # Mock acquire context manager
        mock_acquire = MagicMock()
        mock_acquire.__aenter__ = AsyncMock(return_value=mock_conn)
        mock_acquire.__aexit__ = AsyncMock(return_value=None)
        main.db_pool.acquire = Mock(return_value=mock_acquire)
        
        # Test featured APIs
        response = client.get("/api/apis/featured")
        
        assert response.status_code == 200
        data = response.json()
        assert data["success"] is True
        assert len(data["data"]) == 1
        assert data["data"][0]["name"] == "Featured API"
    
    def test_get_trending_apis(self):
        """Test getting trending APIs"""
        # Mock database response
        mock_conn = AsyncMock()
        mock_trending = [
            {
                "id": uuid.uuid4(),
                "api_id": uuid.uuid4(),
                "title": "Trending API",
                "description": "A trending API",
                "category": "Developer Tools",
                "creator_name": "Bob Smith",
                "pricing_model": "freemium",
                "price_per_call": Decimal("0.001"),
                "avg_rating": Decimal("4.8"),
                "review_count": 50,
                "monthly_calls": 50000,
                "monthly_calls_prev": 25000,
                "tags": "tools,trending",
                "featured": False,
                "growth": 100.0
            }
        ]
        mock_conn.fetch.return_value = mock_trending
        
        # Mock acquire context manager
        mock_acquire = MagicMock()
        mock_acquire.__aenter__ = AsyncMock(return_value=mock_conn)
        mock_acquire.__aexit__ = AsyncMock(return_value=None)
        main.db_pool.acquire = Mock(return_value=mock_acquire)
        
        # Test trending APIs
        response = client.get("/api/apis/trending")
        
        assert response.status_code == 200
        data = response.json()
        assert data["success"] is True
        assert len(data["data"]) == 1
        assert data["data"][0]["name"] == "Trending API"
        assert data["data"][0]["trending"] is True
        assert data["data"][0]["growth"] == 100.0
    
    def test_search_with_facets(self):
        """Test search returns category facets"""
        # Mock database response
        mock_conn = AsyncMock()
        mock_listings = []
        mock_facets = [
            {"category": "AI/ML", "count": 5},
            {"category": "Financial Services", "count": 3}
        ]
        
        # Set up multiple return values for fetch
        mock_conn.fetch.side_effect = [mock_listings, mock_facets]
        mock_conn.fetchval.return_value = 0
        
        # Mock acquire context manager
        mock_acquire = MagicMock()
        mock_acquire.__aenter__ = AsyncMock(return_value=mock_conn)
        mock_acquire.__aexit__ = AsyncMock(return_value=None)
        main.db_pool.acquire = Mock(return_value=mock_acquire)
        
        # Test search with facets
        response = client.get("/api/apis?search=machine learning")
        
        assert response.status_code == 200
        data = response.json()
        assert data["success"] is True
        assert "facets" in data
        assert "categories" in data["facets"]
        assert len(data["facets"]["categories"]) == 2


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
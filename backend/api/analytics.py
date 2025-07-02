"""
Analytics module for API-Direct
Provides comprehensive analytics for API creators
"""

import os
import logging
from datetime import datetime, timedelta
from typing import Dict, Any, Optional, List
import asyncpg
from fastapi import HTTPException
from decimal import Decimal
import json

logger = logging.getLogger(__name__)


class AnalyticsManager:
    """Manages analytics and reporting for APIs"""
    
    def __init__(self, db_pool: asyncpg.Pool):
        self.db_pool = db_pool
    
    async def get_usage_by_consumer(
        self,
        user_id: str,
        api_id: Optional[str] = None,
        period: str = "30d"
    ) -> Dict[str, Any]:
        """Get API usage breakdown by consumer"""
        # Determine time range
        if period == "7d":
            start_date = datetime.utcnow() - timedelta(days=7)
        elif period == "30d":
            start_date = datetime.utcnow() - timedelta(days=30)
        elif period == "90d":
            start_date = datetime.utcnow() - timedelta(days=90)
        else:
            start_date = datetime.utcnow() - timedelta(days=30)
        
        async with self.db_pool.acquire() as conn:
            # Base query
            query = """
                SELECT 
                    ac.consumer_id,
                    u.name as consumer_name,
                    u.company as consumer_company,
                    COUNT(*) as total_calls,
                    COUNT(DISTINCT DATE(ac.created_at)) as active_days,
                    AVG(ac.response_time_ms) as avg_response_time,
                    COUNT(CASE WHEN ac.status_code >= 400 THEN 1 END) as error_count,
                    SUM(ac.amount_charged) as revenue_generated,
                    MAX(ac.created_at) as last_call_at
                FROM api_calls ac
                JOIN apis a ON ac.api_id = a.id
                LEFT JOIN users u ON ac.consumer_id = u.id
                WHERE a.user_id = $1 
                AND ac.created_at >= $2
            """
            
            params = [user_id, start_date]
            
            if api_id:
                query += " AND ac.api_id = $3"
                params.append(api_id)
            
            query += """
                GROUP BY ac.consumer_id, u.name, u.company
                ORDER BY total_calls DESC
                LIMIT 50
            """
            
            consumers = await conn.fetch(query, *params)
            
            # Get total unique consumers
            total_consumers_query = """
                SELECT COUNT(DISTINCT ac.consumer_id) as total
                FROM api_calls ac
                JOIN apis a ON ac.api_id = a.id
                WHERE a.user_id = $1 AND ac.created_at >= $2
            """
            
            if api_id:
                total_consumers_query += " AND ac.api_id = $3"
            
            total_consumers = await conn.fetchval(total_consumers_query, *params)
            
            return {
                "period": period,
                "total_consumers": total_consumers or 0,
                "consumers": [
                    {
                        "consumer_id": str(c["consumer_id"]) if c["consumer_id"] else "anonymous",
                        "consumer_name": c["consumer_name"] or "Anonymous",
                        "consumer_company": c["consumer_company"],
                        "total_calls": c["total_calls"],
                        "active_days": c["active_days"],
                        "avg_response_time": float(c["avg_response_time"]) if c["avg_response_time"] else 0,
                        "error_rate": (c["error_count"] / c["total_calls"] * 100) if c["total_calls"] > 0 else 0,
                        "revenue_generated": float(c["revenue_generated"]) if c["revenue_generated"] else 0,
                        "last_call_at": c["last_call_at"].isoformat() if c["last_call_at"] else None
                    }
                    for c in consumers
                ]
            }
    
    async def get_geographic_analytics(
        self,
        user_id: str,
        api_id: Optional[str] = None,
        period: str = "30d"
    ) -> Dict[str, Any]:
        """Get geographic distribution of API calls"""
        # Determine time range
        if period == "7d":
            start_date = datetime.utcnow() - timedelta(days=7)
        elif period == "30d":
            start_date = datetime.utcnow() - timedelta(days=30)
        else:
            start_date = datetime.utcnow() - timedelta(days=30)
        
        async with self.db_pool.acquire() as conn:
            query = """
                SELECT 
                    ac.country,
                    COUNT(*) as call_count,
                    COUNT(DISTINCT ac.consumer_id) as unique_consumers,
                    AVG(ac.response_time_ms) as avg_response_time,
                    SUM(ac.amount_charged) as revenue
                FROM api_calls ac
                JOIN apis a ON ac.api_id = a.id
                WHERE a.user_id = $1 
                AND ac.created_at >= $2
                AND ac.country IS NOT NULL
            """
            
            params = [user_id, start_date]
            
            if api_id:
                query += " AND ac.api_id = $3"
                params.append(api_id)
            
            query += """
                GROUP BY ac.country
                ORDER BY call_count DESC
                LIMIT 50
            """
            
            countries = await conn.fetch(query, *params)
            
            # Get top cities
            city_query = """
                SELECT 
                    ac.city,
                    ac.country,
                    COUNT(*) as call_count
                FROM api_calls ac
                JOIN apis a ON ac.api_id = a.id
                WHERE a.user_id = $1 
                AND ac.created_at >= $2
                AND ac.city IS NOT NULL
            """
            
            if api_id:
                city_query += " AND ac.api_id = $3"
            
            city_query += """
                GROUP BY ac.city, ac.country
                ORDER BY call_count DESC
                LIMIT 20
            """
            
            cities = await conn.fetch(city_query, *params)
            
            return {
                "period": period,
                "countries": [
                    {
                        "country": c["country"],
                        "call_count": c["call_count"],
                        "unique_consumers": c["unique_consumers"],
                        "avg_response_time": float(c["avg_response_time"]) if c["avg_response_time"] else 0,
                        "revenue": float(c["revenue"]) if c["revenue"] else 0
                    }
                    for c in countries
                ],
                "top_cities": [
                    {
                        "city": c["city"],
                        "country": c["country"],
                        "call_count": c["call_count"]
                    }
                    for c in cities
                ]
            }
    
    async def get_error_analytics(
        self,
        user_id: str,
        api_id: Optional[str] = None,
        period: str = "7d"
    ) -> Dict[str, Any]:
        """Get detailed error analytics"""
        # Determine time range
        if period == "24h":
            start_date = datetime.utcnow() - timedelta(hours=24)
        elif period == "7d":
            start_date = datetime.utcnow() - timedelta(days=7)
        elif period == "30d":
            start_date = datetime.utcnow() - timedelta(days=30)
        else:
            start_date = datetime.utcnow() - timedelta(days=7)
        
        async with self.db_pool.acquire() as conn:
            # Error distribution by status code
            status_query = """
                SELECT 
                    ac.status_code,
                    COUNT(*) as count,
                    ac.path,
                    ac.method
                FROM api_calls ac
                JOIN apis a ON ac.api_id = a.id
                WHERE a.user_id = $1 
                AND ac.created_at >= $2
                AND ac.status_code >= 400
            """
            
            params = [user_id, start_date]
            
            if api_id:
                status_query += " AND ac.api_id = $3"
                params.append(api_id)
            
            status_query += """
                GROUP BY ac.status_code, ac.path, ac.method
                ORDER BY count DESC
                LIMIT 50
            """
            
            error_details = await conn.fetch(status_query, *params)
            
            # Error trends over time
            trend_query = """
                SELECT 
                    DATE_TRUNC('hour', ac.created_at) as time_bucket,
                    COUNT(CASE WHEN ac.status_code >= 400 AND ac.status_code < 500 THEN 1 END) as client_errors,
                    COUNT(CASE WHEN ac.status_code >= 500 THEN 1 END) as server_errors,
                    COUNT(*) as total_calls
                FROM api_calls ac
                JOIN apis a ON ac.api_id = a.id
                WHERE a.user_id = $1 
                AND ac.created_at >= $2
            """
            
            if api_id:
                trend_query += " AND ac.api_id = $3"
            
            trend_query += """
                GROUP BY time_bucket
                ORDER BY time_bucket
            """
            
            error_trends = await conn.fetch(trend_query, *params)
            
            # Most common error messages (if tracked)
            message_query = """
                SELECT 
                    ac.error_message,
                    COUNT(*) as count
                FROM api_calls ac
                JOIN apis a ON ac.api_id = a.id
                WHERE a.user_id = $1 
                AND ac.created_at >= $2
                AND ac.status_code >= 400
                AND ac.error_message IS NOT NULL
            """
            
            if api_id:
                message_query += " AND ac.api_id = $3"
            
            message_query += """
                GROUP BY ac.error_message
                ORDER BY count DESC
                LIMIT 20
            """
            
            error_messages = await conn.fetch(message_query, *params)
            
            return {
                "period": period,
                "error_details": [
                    {
                        "status_code": e["status_code"],
                        "count": e["count"],
                        "endpoint": f"{e['method']} {e['path']}",
                        "error_type": self._get_error_type(e["status_code"])
                    }
                    for e in error_details
                ],
                "error_trends": [
                    {
                        "timestamp": t["time_bucket"].isoformat(),
                        "client_errors": t["client_errors"],
                        "server_errors": t["server_errors"],
                        "error_rate": ((t["client_errors"] + t["server_errors"]) / t["total_calls"] * 100) if t["total_calls"] > 0 else 0
                    }
                    for t in error_trends
                ],
                "common_errors": [
                    {
                        "message": m["error_message"],
                        "count": m["count"]
                    }
                    for m in error_messages
                ]
            }
    
    async def get_revenue_by_api(
        self,
        user_id: str,
        period: str = "30d"
    ) -> Dict[str, Any]:
        """Get revenue breakdown by API"""
        # Determine time range
        if period == "7d":
            start_date = datetime.utcnow() - timedelta(days=7)
        elif period == "30d":
            start_date = datetime.utcnow() - timedelta(days=30)
        elif period == "90d":
            start_date = datetime.utcnow() - timedelta(days=90)
        else:
            start_date = datetime.utcnow() - timedelta(days=30)
        
        async with self.db_pool.acquire() as conn:
            # Revenue by API
            api_revenue = await conn.fetch("""
                SELECT 
                    a.id,
                    a.name,
                    a.pricing_model,
                    a.price_per_request,
                    COUNT(ac.*) as total_calls,
                    COUNT(CASE WHEN ac.billable THEN 1 END) as billable_calls,
                    SUM(ac.amount_charged) as revenue,
                    AVG(ac.amount_charged) as avg_revenue_per_call,
                    COUNT(DISTINCT ac.consumer_id) as unique_consumers
                FROM apis a
                LEFT JOIN api_calls ac ON a.id = ac.api_id AND ac.created_at >= $2
                WHERE a.user_id = $1
                GROUP BY a.id, a.name, a.pricing_model, a.price_per_request
                ORDER BY revenue DESC NULLS LAST
            """, user_id, start_date)
            
            # Revenue trends
            revenue_trends = await conn.fetch("""
                SELECT 
                    DATE_TRUNC('day', ac.created_at) as date,
                    a.name as api_name,
                    SUM(ac.amount_charged) as daily_revenue,
                    COUNT(*) as call_count
                FROM api_calls ac
                JOIN apis a ON ac.api_id = a.id
                WHERE a.user_id = $1 
                AND ac.created_at >= $2
                AND ac.billable = true
                GROUP BY date, a.name
                ORDER BY date, a.name
            """, user_id, start_date)
            
            # Group trends by date
            trends_by_date = {}
            for trend in revenue_trends:
                date_key = trend["date"].isoformat()
                if date_key not in trends_by_date:
                    trends_by_date[date_key] = {
                        "date": date_key,
                        "total_revenue": 0,
                        "apis": {}
                    }
                
                trends_by_date[date_key]["total_revenue"] += float(trend["daily_revenue"]) if trend["daily_revenue"] else 0
                trends_by_date[date_key]["apis"][trend["api_name"]] = {
                    "revenue": float(trend["daily_revenue"]) if trend["daily_revenue"] else 0,
                    "calls": trend["call_count"]
                }
            
            return {
                "period": period,
                "apis": [
                    {
                        "api_id": str(api["id"]),
                        "api_name": api["name"],
                        "pricing_model": api["pricing_model"],
                        "price_per_request": float(api["price_per_request"]) if api["price_per_request"] else 0,
                        "total_calls": api["total_calls"],
                        "billable_calls": api["billable_calls"],
                        "revenue": float(api["revenue"]) if api["revenue"] else 0,
                        "avg_revenue_per_call": float(api["avg_revenue_per_call"]) if api["avg_revenue_per_call"] else 0,
                        "unique_consumers": api["unique_consumers"]
                    }
                    for api in api_revenue
                ],
                "revenue_trends": list(trends_by_date.values())
            }
    
    async def get_endpoint_analytics(
        self,
        user_id: str,
        api_id: str,
        period: str = "7d"
    ) -> Dict[str, Any]:
        """Get detailed analytics for API endpoints"""
        # Determine time range
        if period == "24h":
            start_date = datetime.utcnow() - timedelta(hours=24)
        elif period == "7d":
            start_date = datetime.utcnow() - timedelta(days=7)
        elif period == "30d":
            start_date = datetime.utcnow() - timedelta(days=30)
        else:
            start_date = datetime.utcnow() - timedelta(days=7)
        
        async with self.db_pool.acquire() as conn:
            # Verify API ownership
            api_owner = await conn.fetchval(
                "SELECT user_id FROM apis WHERE id = $1",
                api_id
            )
            
            if api_owner != user_id:
                raise HTTPException(status_code=403, detail="Not authorized to view this API's analytics")
            
            # Endpoint performance
            endpoints = await conn.fetch("""
                SELECT 
                    method,
                    path,
                    COUNT(*) as call_count,
                    AVG(response_time_ms) as avg_response_time,
                    PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY response_time_ms) as p50_response_time,
                    PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY response_time_ms) as p95_response_time,
                    PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY response_time_ms) as p99_response_time,
                    COUNT(CASE WHEN status_code >= 200 AND status_code < 300 THEN 1 END) * 100.0 / COUNT(*) as success_rate,
                    AVG(request_size_bytes) as avg_request_size,
                    AVG(response_size_bytes) as avg_response_size
                FROM api_calls
                WHERE api_id = $1 
                AND created_at >= $2
                GROUP BY method, path
                ORDER BY call_count DESC
            """, api_id, start_date)
            
            # Status code distribution by endpoint
            status_distribution = await conn.fetch("""
                SELECT 
                    method,
                    path,
                    status_code,
                    COUNT(*) as count
                FROM api_calls
                WHERE api_id = $1 
                AND created_at >= $2
                GROUP BY method, path, status_code
                ORDER BY method, path, status_code
            """, api_id, start_date)
            
            # Group status codes by endpoint
            status_by_endpoint = {}
            for s in status_distribution:
                endpoint_key = f"{s['method']} {s['path']}"
                if endpoint_key not in status_by_endpoint:
                    status_by_endpoint[endpoint_key] = {}
                status_by_endpoint[endpoint_key][str(s['status_code'])] = s['count']
            
            return {
                "period": period,
                "api_id": api_id,
                "endpoints": [
                    {
                        "method": e["method"],
                        "path": e["path"],
                        "call_count": e["call_count"],
                        "performance": {
                            "avg_response_time": float(e["avg_response_time"]) if e["avg_response_time"] else 0,
                            "p50_response_time": float(e["p50_response_time"]) if e["p50_response_time"] else 0,
                            "p95_response_time": float(e["p95_response_time"]) if e["p95_response_time"] else 0,
                            "p99_response_time": float(e["p99_response_time"]) if e["p99_response_time"] else 0
                        },
                        "success_rate": float(e["success_rate"]) if e["success_rate"] else 0,
                        "avg_request_size": e["avg_request_size"] or 0,
                        "avg_response_size": e["avg_response_size"] or 0,
                        "status_codes": status_by_endpoint.get(f"{e['method']} {e['path']}", {})
                    }
                    for e in endpoints
                ]
            }
    
    async def get_usage_patterns(
        self,
        user_id: str,
        api_id: Optional[str] = None
    ) -> Dict[str, Any]:
        """Get API usage patterns (hourly, daily, weekly)"""
        async with self.db_pool.acquire() as conn:
            # Base query conditions
            base_conditions = "a.user_id = $1 AND ac.created_at >= NOW() - INTERVAL '30 days'"
            params = [user_id]
            
            if api_id:
                base_conditions += " AND ac.api_id = $2"
                params.append(api_id)
            
            # Hourly pattern (which hours are busiest)
            hourly_query = f"""
                SELECT 
                    EXTRACT(HOUR FROM ac.created_at) as hour,
                    COUNT(*) as call_count,
                    AVG(ac.response_time_ms) as avg_response_time
                FROM api_calls ac
                JOIN apis a ON ac.api_id = a.id
                WHERE {base_conditions}
                GROUP BY hour
                ORDER BY hour
            """
            
            hourly_pattern = await conn.fetch(hourly_query, *params)
            
            # Daily pattern (which days of week are busiest)
            daily_query = f"""
                SELECT 
                    EXTRACT(DOW FROM ac.created_at) as day_of_week,
                    TO_CHAR(ac.created_at, 'Day') as day_name,
                    COUNT(*) as call_count,
                    AVG(ac.response_time_ms) as avg_response_time
                FROM api_calls ac
                JOIN apis a ON ac.api_id = a.id
                WHERE {base_conditions}
                GROUP BY day_of_week, day_name
                ORDER BY day_of_week
            """
            
            daily_pattern = await conn.fetch(daily_query, *params)
            
            # Peak usage times
            peak_query = f"""
                SELECT 
                    DATE_TRUNC('hour', ac.created_at) as hour_bucket,
                    COUNT(*) as call_count
                FROM api_calls ac
                JOIN apis a ON ac.api_id = a.id
                WHERE {base_conditions}
                GROUP BY hour_bucket
                ORDER BY call_count DESC
                LIMIT 10
            """
            
            peak_hours = await conn.fetch(peak_query, *params)
            
            return {
                "hourly_pattern": [
                    {
                        "hour": int(h["hour"]),
                        "call_count": h["call_count"],
                        "avg_response_time": float(h["avg_response_time"]) if h["avg_response_time"] else 0
                    }
                    for h in hourly_pattern
                ],
                "daily_pattern": [
                    {
                        "day_of_week": int(d["day_of_week"]),
                        "day_name": d["day_name"].strip(),
                        "call_count": d["call_count"],
                        "avg_response_time": float(d["avg_response_time"]) if d["avg_response_time"] else 0
                    }
                    for d in daily_pattern
                ],
                "peak_hours": [
                    {
                        "timestamp": p["hour_bucket"].isoformat(),
                        "call_count": p["call_count"]
                    }
                    for p in peak_hours
                ]
            }
    
    def _get_error_type(self, status_code: int) -> str:
        """Get human-readable error type from status code"""
        if status_code == 400:
            return "Bad Request"
        elif status_code == 401:
            return "Unauthorized"
        elif status_code == 403:
            return "Forbidden"
        elif status_code == 404:
            return "Not Found"
        elif status_code == 429:
            return "Too Many Requests"
        elif status_code >= 400 and status_code < 500:
            return "Client Error"
        elif status_code == 500:
            return "Internal Server Error"
        elif status_code == 502:
            return "Bad Gateway"
        elif status_code == 503:
            return "Service Unavailable"
        elif status_code >= 500:
            return "Server Error"
        else:
            return "Unknown"
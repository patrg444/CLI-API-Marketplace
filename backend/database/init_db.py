#!/usr/bin/env python3
"""
Database initialization script
Creates database and runs schema.sql
"""

import asyncio
import asyncpg
import os
import sys
from pathlib import Path

# Database configuration
DATABASE_HOST = os.getenv("DATABASE_HOST", "localhost")
DATABASE_PORT = os.getenv("DATABASE_PORT", "5432")
DATABASE_NAME = os.getenv("DATABASE_NAME", "apidirect")
DATABASE_USER = os.getenv("DATABASE_USER", "postgres")
DATABASE_PASSWORD = os.getenv("DATABASE_PASSWORD", "")

# Connection URLs
ADMIN_DATABASE_URL = f"postgresql://{DATABASE_USER}:{DATABASE_PASSWORD}@{DATABASE_HOST}:{DATABASE_PORT}/postgres"
DATABASE_URL = f"postgresql://{DATABASE_USER}:{DATABASE_PASSWORD}@{DATABASE_HOST}:{DATABASE_PORT}/{DATABASE_NAME}"


async def create_database():
    """Create the database if it doesn't exist"""
    try:
        # Connect to postgres database to create our database
        conn = await asyncpg.connect(ADMIN_DATABASE_URL)
        
        # Check if database exists
        exists = await conn.fetchval(
            "SELECT EXISTS(SELECT datname FROM pg_database WHERE datname = $1)",
            DATABASE_NAME
        )
        
        if not exists:
            await conn.execute(f'CREATE DATABASE "{DATABASE_NAME}"')
            print(f"Database '{DATABASE_NAME}' created successfully")
        else:
            print(f"Database '{DATABASE_NAME}' already exists")
        
        await conn.close()
        return True
        
    except Exception as e:
        print(f"Error creating database: {e}")
        return False


async def run_schema():
    """Run the schema.sql file to create tables"""
    try:
        # Connect to our database
        conn = await asyncpg.connect(DATABASE_URL)
        
        # Read schema file
        schema_path = Path(__file__).parent / "schema.sql"
        with open(schema_path, 'r') as f:
            schema_sql = f.read()
        
        # Execute schema
        await conn.execute(schema_sql)
        print("Schema created successfully")
        
        await conn.close()
        return True
        
    except Exception as e:
        print(f"Error running schema: {e}")
        import traceback
        traceback.print_exc()
        return False


async def create_test_data():
    """Create some test data for development"""
    try:
        conn = await asyncpg.connect(DATABASE_URL)
        
        # Check if we already have users
        user_count = await conn.fetchval("SELECT COUNT(*) FROM users")
        
        if user_count == 0:
            # Create a test user
            test_user_id = await conn.fetchval("""
                INSERT INTO users (email, password_hash, name, company, email_verified)
                VALUES ($1, $2, $3, $4, $5)
                RETURNING id
            """, 
                "test@example.com",
                "$2b$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewKyNiGU/FOAXBJa",  # password: "password123"
                "Test User",
                "Test Company",
                True
            )
            
            print(f"Created test user: test@example.com (password: password123)")
            
            # Create a test API
            await conn.execute("""
                INSERT INTO apis (user_id, name, description, deployment_type, status, endpoint_url)
                VALUES ($1, $2, $3, $4, $5, $6)
            """,
                test_user_id,
                "Test Weather API",
                "A test API for weather data",
                "hosted",
                "running",
                "https://api.example.com/weather"
            )
            
            print("Created test API")
        else:
            print("Test data already exists")
        
        await conn.close()
        return True
        
    except Exception as e:
        print(f"Error creating test data: {e}")
        return False


async def verify_database():
    """Verify database setup"""
    try:
        conn = await asyncpg.connect(DATABASE_URL)
        
        # Check tables
        tables = await conn.fetch("""
            SELECT tablename FROM pg_tables 
            WHERE schemaname = 'public'
            ORDER BY tablename
        """)
        
        print("\nDatabase tables:")
        for table in tables:
            count = await conn.fetchval(f"SELECT COUNT(*) FROM {table['tablename']}")
            print(f"  - {table['tablename']}: {count} rows")
        
        await conn.close()
        return True
        
    except Exception as e:
        print(f"Error verifying database: {e}")
        return False


async def main():
    """Initialize the database"""
    print("API-Direct Database Initialization")
    print("==================================")
    print(f"Database: {DATABASE_NAME}")
    print(f"Host: {DATABASE_HOST}:{DATABASE_PORT}")
    print()
    
    # Create database
    if not await create_database():
        print("Failed to create database")
        sys.exit(1)
    
    # Run schema
    if not await run_schema():
        print("Failed to create schema")
        sys.exit(1)
    
    # Create test data (only in development)
    if os.getenv("ENVIRONMENT", "development") == "development":
        await create_test_data()
    
    # Verify setup
    await verify_database()
    
    print("\nDatabase initialization complete!")


if __name__ == "__main__":
    asyncio.run(main())
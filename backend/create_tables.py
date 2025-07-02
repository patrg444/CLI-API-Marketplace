#!/usr/bin/env python3
"""
Create database tables
"""
import asyncio
import os
from sqlalchemy import create_engine
from api.database import Base

def create_tables():
    # Use synchronous engine for table creation
    db_url = os.getenv('DATABASE_URL', 'postgresql://apidirect:localpassword@localhost:5432/apidirect?sslmode=disable')
    
    # Convert async URL to sync
    if db_url.startswith('postgresql+asyncpg://'):
        db_url = db_url.replace('postgresql+asyncpg://', 'postgresql://')
    
    engine = create_engine(db_url)
    
    print("Creating database tables...")
    Base.metadata.create_all(bind=engine)
    print("Tables created successfully!")
    
    # Show created tables
    from sqlalchemy import inspect
    inspector = inspect(engine)
    tables = inspector.get_table_names()
    print(f"\nCreated tables: {', '.join(tables)}")

if __name__ == "__main__":
    create_tables()
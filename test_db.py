#!/usr/bin/env python3
"""Test database connection"""
import asyncio
from backend.api.database import get_db_manager

async def test_db():
    db_manager = get_db_manager()
    async with db_manager.get_session() as session:
        print("Database connection successful!")
        
if __name__ == "__main__":
    asyncio.run(test_db())
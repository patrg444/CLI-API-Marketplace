"""add_marketplace_search_columns

Revision ID: 1995d068809f
Revises: 89f904e2bc1c
Create Date: 2025-06-29 18:16:23.504587

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = '1995d068809f'
down_revision: Union[str, None] = '89f904e2bc1c'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Add monthly_calls_prev column to apis table for growth calculation
    op.add_column('apis', sa.Column('monthly_calls_prev', sa.Integer(), nullable=True))
    
    # Add indexes for better search performance
    op.create_index('idx_marketplace_listings_status', 'marketplace_listings', ['status'])
    op.create_index('idx_marketplace_listings_category', 'marketplace_listings', ['category'])
    op.create_index('idx_marketplace_listings_featured', 'marketplace_listings', ['featured'])
    op.create_index('idx_marketplace_listings_title', 'marketplace_listings', ['title'])
    
    # Add GIN index for full-text search on title, description, and tags
    op.execute("""
        CREATE EXTENSION IF NOT EXISTS pg_trgm;
        CREATE INDEX idx_marketplace_listings_search 
        ON marketplace_listings 
        USING gin((title || ' ' || description || ' ' || COALESCE(tags, '')) gin_trgm_ops);
    """)


def downgrade() -> None:
    # Remove indexes
    op.drop_index('idx_marketplace_listings_search', 'marketplace_listings')
    op.drop_index('idx_marketplace_listings_title', 'marketplace_listings')
    op.drop_index('idx_marketplace_listings_featured', 'marketplace_listings')
    op.drop_index('idx_marketplace_listings_category', 'marketplace_listings')
    op.drop_index('idx_marketplace_listings_status', 'marketplace_listings')
    
    # Remove column
    op.drop_column('apis', 'monthly_calls_prev')

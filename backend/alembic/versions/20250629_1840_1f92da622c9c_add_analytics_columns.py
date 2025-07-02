"""add_analytics_columns

Revision ID: 1f92da622c9c
Revises: 1995d068809f
Create Date: 2025-06-29 18:40:32.415469

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = '1f92da622c9c'
down_revision: Union[str, None] = '1995d068809f'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Add error_message column to api_calls for better error tracking
    op.add_column('api_calls', sa.Column('error_message', sa.Text(), nullable=True))
    
    # Add city column for more detailed geographic analytics
    op.add_column('api_calls', sa.Column('city', sa.String(100), nullable=True))
    
    # Add indexes for analytics queries
    op.create_index('idx_api_calls_consumer_id', 'api_calls', ['consumer_id'])
    op.create_index('idx_api_calls_country', 'api_calls', ['country'])
    op.create_index('idx_api_calls_status_code', 'api_calls', ['status_code'])
    op.create_index('idx_api_calls_created_at_api_id', 'api_calls', ['created_at', 'api_id'])
    
    # Add composite index for revenue queries
    op.create_index('idx_api_calls_billable_created_at', 'api_calls', ['billable', 'created_at'])


def downgrade() -> None:
    # Remove indexes
    op.drop_index('idx_api_calls_billable_created_at', 'api_calls')
    op.drop_index('idx_api_calls_created_at_api_id', 'api_calls')
    op.drop_index('idx_api_calls_status_code', 'api_calls')
    op.drop_index('idx_api_calls_country', 'api_calls')
    op.drop_index('idx_api_calls_consumer_id', 'api_calls')
    
    # Remove columns
    op.drop_column('api_calls', 'city')
    op.drop_column('api_calls', 'error_message')

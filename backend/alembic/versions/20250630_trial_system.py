"""add api trial and sandbox system

Revision ID: trial_system_001
Revises: 980c7ef8e5d9
Create Date: 2025-06-30 10:00:00.000000

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import postgresql

# revision identifiers, used by Alembic.
revision: str = 'trial_system_001'
down_revision: Union[str, None] = '980c7ef8e5d9'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Add trial configuration to apis table
    op.add_column('apis', sa.Column('trial_enabled', sa.Boolean(), nullable=False, server_default='false'))
    op.add_column('apis', sa.Column('trial_requests', sa.Integer(), nullable=True))
    op.add_column('apis', sa.Column('trial_duration_days', sa.Integer(), nullable=True))
    op.add_column('apis', sa.Column('trial_rate_limit', sa.Integer(), nullable=True))  # requests per minute
    op.add_column('apis', sa.Column('sandbox_enabled', sa.Boolean(), nullable=False, server_default='false'))
    op.add_column('apis', sa.Column('sandbox_base_url', sa.String(500), nullable=True))
    
    # Create api_trials table to track user trials
    op.create_table(
        'api_trials',
        sa.Column('id', sa.UUID(), nullable=False, server_default=sa.text('gen_random_uuid()')),
        sa.Column('user_id', sa.UUID(), nullable=False),
        sa.Column('api_id', sa.UUID(), nullable=False),
        sa.Column('status', sa.String(20), nullable=False, server_default='active'),  # active, expired, converted
        sa.Column('requests_used', sa.Integer(), nullable=False, server_default='0'),
        sa.Column('requests_limit', sa.Integer(), nullable=True),
        sa.Column('started_at', sa.TIMESTAMP(timezone=True), nullable=False, server_default=sa.text('NOW()')),
        sa.Column('expires_at', sa.TIMESTAMP(timezone=True), nullable=True),
        sa.Column('converted_at', sa.TIMESTAMP(timezone=True), nullable=True),
        sa.Column('created_at', sa.TIMESTAMP(timezone=True), nullable=False, server_default=sa.text('NOW()')),
        sa.PrimaryKeyConstraint('id'),
        sa.ForeignKeyConstraint(['user_id'], ['users.id'], ondelete='CASCADE'),
        sa.ForeignKeyConstraint(['api_id'], ['apis.id'], ondelete='CASCADE'),
        sa.UniqueConstraint('user_id', 'api_id', name='uq_user_api_trial'),
        sa.Index('idx_api_trials_user_id', 'user_id'),
        sa.Index('idx_api_trials_api_id', 'api_id'),
        sa.Index('idx_api_trials_status', 'status'),
        sa.Index('idx_api_trials_expires_at', 'expires_at')
    )
    
    # Create sandbox_requests table for test mode tracking
    op.create_table(
        'sandbox_requests',
        sa.Column('id', sa.UUID(), nullable=False, server_default=sa.text('gen_random_uuid()')),
        sa.Column('api_id', sa.UUID(), nullable=False),
        sa.Column('user_id', sa.UUID(), nullable=False),
        sa.Column('endpoint', sa.String(500), nullable=False),
        sa.Column('method', sa.String(10), nullable=False),
        sa.Column('request_headers', sa.JSON(), nullable=True),
        sa.Column('request_body', sa.JSON(), nullable=True),
        sa.Column('response_status', sa.Integer(), nullable=True),
        sa.Column('response_headers', sa.JSON(), nullable=True),
        sa.Column('response_body', sa.JSON(), nullable=True),
        sa.Column('response_time_ms', sa.Integer(), nullable=True),
        sa.Column('is_mocked', sa.Boolean(), nullable=False, server_default='false'),
        sa.Column('created_at', sa.TIMESTAMP(timezone=True), nullable=False, server_default=sa.text('NOW()')),
        sa.PrimaryKeyConstraint('id'),
        sa.ForeignKeyConstraint(['api_id'], ['apis.id'], ondelete='CASCADE'),
        sa.ForeignKeyConstraint(['user_id'], ['users.id'], ondelete='CASCADE'),
        sa.Index('idx_sandbox_requests_api_id', 'api_id'),
        sa.Index('idx_sandbox_requests_user_id', 'user_id'),
        sa.Index('idx_sandbox_requests_created_at', 'created_at')
    )
    
    # Add mock responses table for sandbox mode
    op.create_table(
        'api_mock_responses',
        sa.Column('id', sa.UUID(), nullable=False, server_default=sa.text('gen_random_uuid()')),
        sa.Column('api_id', sa.UUID(), nullable=False),
        sa.Column('endpoint_pattern', sa.String(500), nullable=False),  # regex pattern
        sa.Column('method', sa.String(10), nullable=False),
        sa.Column('response_status', sa.Integer(), nullable=False, server_default='200'),
        sa.Column('response_headers', sa.JSON(), nullable=True),
        sa.Column('response_body', sa.JSON(), nullable=False),
        sa.Column('description', sa.Text(), nullable=True),
        sa.Column('is_active', sa.Boolean(), nullable=False, server_default='true'),
        sa.Column('created_at', sa.TIMESTAMP(timezone=True), nullable=False, server_default=sa.text('NOW()')),
        sa.Column('updated_at', sa.TIMESTAMP(timezone=True), nullable=False, server_default=sa.text('NOW()')),
        sa.PrimaryKeyConstraint('id'),
        sa.ForeignKeyConstraint(['api_id'], ['apis.id'], ondelete='CASCADE'),
        sa.Index('idx_mock_responses_api_id', 'api_id'),
        sa.Index('idx_mock_responses_active', 'is_active')
    )


def downgrade() -> None:
    # Drop tables
    op.drop_table('api_mock_responses')
    op.drop_table('sandbox_requests')
    op.drop_table('api_trials')
    
    # Remove columns from apis table
    op.drop_column('apis', 'sandbox_base_url')
    op.drop_column('apis', 'sandbox_enabled')
    op.drop_column('apis', 'trial_rate_limit')
    op.drop_column('apis', 'trial_duration_days')
    op.drop_column('apis', 'trial_requests')
    op.drop_column('apis', 'trial_enabled')
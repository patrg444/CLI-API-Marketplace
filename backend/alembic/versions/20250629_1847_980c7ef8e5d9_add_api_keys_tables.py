"""add_api_keys_tables

Revision ID: 980c7ef8e5d9
Revises: 1f92da622c9c
Create Date: 2025-06-29 18:47:23.685218

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = '980c7ef8e5d9'
down_revision: Union[str, None] = '1f92da622c9c'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Create API keys table
    op.create_table(
        'api_keys',
        sa.Column('id', sa.UUID(), nullable=False, server_default=sa.text('gen_random_uuid()')),
        sa.Column('user_id', sa.UUID(), nullable=False),
        sa.Column('name', sa.String(100), nullable=False),
        sa.Column('description', sa.Text(), nullable=True),
        sa.Column('key_hash', sa.String(255), nullable=False),
        sa.Column('scopes', sa.ARRAY(sa.String()), nullable=True, server_default='{}'),
        sa.Column('status', sa.String(20), nullable=False, server_default='active'),
        sa.Column('expires_at', sa.TIMESTAMP(timezone=True), nullable=True),
        sa.Column('last_used_at', sa.TIMESTAMP(timezone=True), nullable=True),
        sa.Column('created_at', sa.TIMESTAMP(timezone=True), nullable=False, server_default=sa.text('NOW()')),
        sa.Column('updated_at', sa.TIMESTAMP(timezone=True), nullable=False, server_default=sa.text('NOW()')),
        sa.PrimaryKeyConstraint('id'),
        sa.ForeignKeyConstraint(['user_id'], ['users.id'], ondelete='CASCADE'),
        sa.Index('idx_api_keys_user_id', 'user_id'),
        sa.Index('idx_api_keys_status', 'status')
    )
    
    # Create API key logs table for tracking usage
    op.create_table(
        'api_key_logs',
        sa.Column('id', sa.UUID(), nullable=False, server_default=sa.text('gen_random_uuid()')),
        sa.Column('key_id', sa.UUID(), nullable=False),
        sa.Column('action', sa.String(50), nullable=False),  # created, used, revoked
        sa.Column('ip_address', sa.String(45), nullable=True),
        sa.Column('user_agent', sa.Text(), nullable=True),
        sa.Column('created_at', sa.TIMESTAMP(timezone=True), nullable=False, server_default=sa.text('NOW()')),
        sa.PrimaryKeyConstraint('id'),
        sa.ForeignKeyConstraint(['key_id'], ['api_keys.id'], ondelete='CASCADE'),
        sa.Index('idx_api_key_logs_key_id', 'key_id'),
        sa.Index('idx_api_key_logs_created_at', 'created_at')
    )
    
    # Create user_api_keys table (alternative naming used in code)
    op.create_table(
        'user_api_keys',
        sa.Column('id', sa.UUID(), nullable=False, server_default=sa.text('gen_random_uuid()')),
        sa.Column('user_id', sa.UUID(), nullable=False),
        sa.Column('name', sa.String(100), nullable=False),
        sa.Column('key_hash', sa.String(255), nullable=False),
        sa.Column('scopes', sa.ARRAY(sa.String()), nullable=True, server_default='{}'),
        sa.Column('expires_at', sa.TIMESTAMP(timezone=True), nullable=True),
        sa.Column('last_used_at', sa.TIMESTAMP(timezone=True), nullable=True),
        sa.Column('created_at', sa.TIMESTAMP(timezone=True), nullable=False, server_default=sa.text('NOW()')),
        sa.PrimaryKeyConstraint('id'),
        sa.ForeignKeyConstraint(['user_id'], ['users.id'], ondelete='CASCADE'),
        sa.Index('idx_user_api_keys_user_id', 'user_id')
    )


def downgrade() -> None:
    # Drop tables
    op.drop_table('user_api_keys')
    op.drop_table('api_key_logs')
    op.drop_table('api_keys')

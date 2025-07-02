"""add api documentation schema

Revision ID: api_docs_001
Revises: trial_system_001
Create Date: 2025-06-30 11:00:00.000000

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import postgresql

# revision identifiers, used by Alembic.
revision: str = 'api_docs_001'
down_revision: Union[str, None] = 'trial_system_001'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Create api_endpoints table
    op.create_table(
        'api_endpoints',
        sa.Column('id', sa.UUID(), nullable=False, server_default=sa.text('gen_random_uuid()')),
        sa.Column('api_id', sa.UUID(), nullable=False),
        sa.Column('path', sa.String(500), nullable=False),
        sa.Column('method', sa.String(10), nullable=False),
        sa.Column('summary', sa.String(200), nullable=True),
        sa.Column('description', sa.Text(), nullable=True),
        sa.Column('operation_id', sa.String(100), nullable=True),
        sa.Column('category', sa.String(50), nullable=True),
        sa.Column('deprecated', sa.Boolean(), nullable=False, server_default='false'),
        sa.Column('requires_auth', sa.Boolean(), nullable=False, server_default='true'),
        sa.Column('created_at', sa.TIMESTAMP(timezone=True), nullable=False, server_default=sa.text('NOW()')),
        sa.Column('updated_at', sa.TIMESTAMP(timezone=True), nullable=False, server_default=sa.text('NOW()')),
        sa.PrimaryKeyConstraint('id'),
        sa.ForeignKeyConstraint(['api_id'], ['apis.id'], ondelete='CASCADE'),
        sa.UniqueConstraint('api_id', 'path', 'method', name='uq_api_endpoint'),
        sa.Index('idx_api_endpoints_api_id', 'api_id')
    )
    
    # Create api_parameters table
    op.create_table(
        'api_parameters',
        sa.Column('id', sa.UUID(), nullable=False, server_default=sa.text('gen_random_uuid()')),
        sa.Column('endpoint_id', sa.UUID(), nullable=False),
        sa.Column('name', sa.String(100), nullable=False),
        sa.Column('location', sa.String(20), nullable=False),  # path, query, header, cookie
        sa.Column('description', sa.Text(), nullable=True),
        sa.Column('required', sa.Boolean(), nullable=False, server_default='false'),
        sa.Column('schema', sa.JSON(), nullable=True),
        sa.Column('example', sa.Text(), nullable=True),
        sa.Column('deprecated', sa.Boolean(), nullable=False, server_default='false'),
        sa.Column('created_at', sa.TIMESTAMP(timezone=True), nullable=False, server_default=sa.text('NOW()')),
        sa.PrimaryKeyConstraint('id'),
        sa.ForeignKeyConstraint(['endpoint_id'], ['api_endpoints.id'], ondelete='CASCADE'),
        sa.Index('idx_api_parameters_endpoint_id', 'endpoint_id')
    )
    
    # Create api_request_bodies table
    op.create_table(
        'api_request_bodies',
        sa.Column('id', sa.UUID(), nullable=False, server_default=sa.text('gen_random_uuid()')),
        sa.Column('endpoint_id', sa.UUID(), nullable=False),
        sa.Column('description', sa.Text(), nullable=True),
        sa.Column('required', sa.Boolean(), nullable=False, server_default='true'),
        sa.Column('schema', sa.JSON(), nullable=False),
        sa.Column('examples', sa.JSON(), nullable=True),
        sa.Column('created_at', sa.TIMESTAMP(timezone=True), nullable=False, server_default=sa.text('NOW()')),
        sa.PrimaryKeyConstraint('id'),
        sa.ForeignKeyConstraint(['endpoint_id'], ['api_endpoints.id'], ondelete='CASCADE'),
        sa.UniqueConstraint('endpoint_id', name='uq_endpoint_request_body'),
        sa.Index('idx_api_request_bodies_endpoint_id', 'endpoint_id')
    )
    
    # Create api_responses table
    op.create_table(
        'api_responses',
        sa.Column('id', sa.UUID(), nullable=False, server_default=sa.text('gen_random_uuid()')),
        sa.Column('endpoint_id', sa.UUID(), nullable=False),
        sa.Column('status_code', sa.Integer(), nullable=False),
        sa.Column('description', sa.Text(), nullable=True),
        sa.Column('schema', sa.JSON(), nullable=True),
        sa.Column('examples', sa.JSON(), nullable=True),
        sa.Column('created_at', sa.TIMESTAMP(timezone=True), nullable=False, server_default=sa.text('NOW()')),
        sa.PrimaryKeyConstraint('id'),
        sa.ForeignKeyConstraint(['endpoint_id'], ['api_endpoints.id'], ondelete='CASCADE'),
        sa.UniqueConstraint('endpoint_id', 'status_code', name='uq_endpoint_response'),
        sa.Index('idx_api_responses_endpoint_id', 'endpoint_id')
    )
    
    # Create api_schemas table for reusable schemas
    op.create_table(
        'api_schemas',
        sa.Column('id', sa.UUID(), nullable=False, server_default=sa.text('gen_random_uuid()')),
        sa.Column('api_id', sa.UUID(), nullable=False),
        sa.Column('name', sa.String(100), nullable=False),
        sa.Column('description', sa.Text(), nullable=True),
        sa.Column('schema', sa.JSON(), nullable=False),
        sa.Column('created_at', sa.TIMESTAMP(timezone=True), nullable=False, server_default=sa.text('NOW()')),
        sa.PrimaryKeyConstraint('id'),
        sa.ForeignKeyConstraint(['api_id'], ['apis.id'], ondelete='CASCADE'),
        sa.UniqueConstraint('api_id', 'name', name='uq_api_schema_name'),
        sa.Index('idx_api_schemas_api_id', 'api_id')
    )
    
    # Add documentation-related columns to apis table
    op.add_column('apis', sa.Column('version', sa.String(20), nullable=True, server_default='1.0.0'))
    op.add_column('apis', sa.Column('logo_url', sa.String(500), nullable=True))
    op.add_column('apis', sa.Column('auth_type', sa.String(20), nullable=True))  # apiKey, bearer, oauth2, basic
    op.add_column('apis', sa.Column('auth_location', sa.String(20), nullable=True))  # header, query
    op.add_column('apis', sa.Column('auth_header', sa.String(50), nullable=True))  # e.g., X-API-Key
    op.add_column('apis', sa.Column('oauth_auth_url', sa.String(500), nullable=True))
    op.add_column('apis', sa.Column('oauth_token_url', sa.String(500), nullable=True))
    op.add_column('apis', sa.Column('oauth_scopes', sa.JSON(), nullable=True))


def downgrade() -> None:
    # Remove columns from apis table
    op.drop_column('apis', 'oauth_scopes')
    op.drop_column('apis', 'oauth_token_url')
    op.drop_column('apis', 'oauth_auth_url')
    op.drop_column('apis', 'auth_header')
    op.drop_column('apis', 'auth_location')
    op.drop_column('apis', 'auth_type')
    op.drop_column('apis', 'logo_url')
    op.drop_column('apis', 'version')
    
    # Drop tables
    op.drop_table('api_schemas')
    op.drop_table('api_responses')
    op.drop_table('api_request_bodies')
    op.drop_table('api_parameters')
    op.drop_table('api_endpoints')
"""Initial schema

Revision ID: 89f904e2bc1c
Revises: 
Create Date: 2025-06-29 16:29:21.401015

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import postgresql


# revision identifiers, used by Alembic.
revision: str = '89f904e2bc1c'
down_revision: Union[str, None] = None
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Enable UUID extension for PostgreSQL
    op.execute('CREATE EXTENSION IF NOT EXISTS "uuid-ossp"')
    
    # Create users table
    op.create_table('users',
        sa.Column('id', sa.UUID(), server_default=sa.text('uuid_generate_v4()'), nullable=False),
        sa.Column('email', sa.String(255), nullable=False),
        sa.Column('password_hash', sa.String(255), nullable=False),
        sa.Column('name', sa.String(255), nullable=False),
        sa.Column('company', sa.String(255), nullable=True),
        sa.Column('phone', sa.String(50), nullable=True),
        sa.Column('bio', sa.Text(), nullable=True),
        sa.Column('avatar_url', sa.String(500), nullable=True),
        sa.Column('email_verified', sa.Boolean(), server_default='false', nullable=True),
        sa.Column('is_active', sa.Boolean(), server_default='true', nullable=True),
        sa.Column('is_premium', sa.Boolean(), server_default='false', nullable=True),
        sa.Column('default_deployment_type', sa.String(20), server_default='hosted', nullable=True),
        sa.Column('timezone', sa.String(50), server_default='UTC', nullable=True),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.text('NOW()'), nullable=True),
        sa.Column('updated_at', sa.DateTime(timezone=True), server_default=sa.text('NOW()'), nullable=True),
        sa.Column('last_login_at', sa.DateTime(timezone=True), nullable=True),
        sa.PrimaryKeyConstraint('id'),
        sa.UniqueConstraint('email')
    )
    
    # Create user_api_keys table
    op.create_table('user_api_keys',
        sa.Column('id', sa.UUID(), server_default=sa.text('uuid_generate_v4()'), nullable=False),
        sa.Column('user_id', sa.UUID(), nullable=False),
        sa.Column('key_hash', sa.String(255), nullable=False),
        sa.Column('name', sa.String(100), nullable=False),
        sa.Column('scopes', sa.JSON(), server_default='[]', nullable=True),
        sa.Column('last_used_at', sa.DateTime(timezone=True), nullable=True),
        sa.Column('expires_at', sa.DateTime(timezone=True), nullable=True),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.text('NOW()'), nullable=True),
        sa.ForeignKeyConstraint(['user_id'], ['users.id'], ondelete='CASCADE'),
        sa.PrimaryKeyConstraint('id')
    )
    
    # Create password_reset_tokens table
    op.create_table('password_reset_tokens',
        sa.Column('id', sa.UUID(), server_default=sa.text('uuid_generate_v4()'), nullable=False),
        sa.Column('user_id', sa.UUID(), nullable=False),
        sa.Column('token', sa.String(100), nullable=False),
        sa.Column('expires_at', sa.DateTime(timezone=True), nullable=False),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.text('NOW()'), nullable=True),
        sa.ForeignKeyConstraint(['user_id'], ['users.id'], ondelete='CASCADE'),
        sa.PrimaryKeyConstraint('id'),
        sa.UniqueConstraint('token'),
        sa.UniqueConstraint('user_id')
    )
    
    # Create apis table
    op.create_table('apis',
        sa.Column('id', sa.UUID(), server_default=sa.text('uuid_generate_v4()'), nullable=False),
        sa.Column('user_id', sa.UUID(), nullable=False),
        sa.Column('name', sa.String(100), nullable=False),
        sa.Column('description', sa.Text(), nullable=True),
        sa.Column('version', sa.String(20), server_default='1.0.0', nullable=True),
        sa.Column('deployment_type', sa.String(20), nullable=False),
        sa.Column('status', sa.String(20), server_default='building', nullable=True),
        sa.Column('endpoint_url', sa.String(500), nullable=True),
        sa.Column('custom_domain', sa.String(255), nullable=True),
        sa.Column('template_id', sa.String(50), nullable=True),
        sa.Column('runtime_config', sa.JSON(), server_default='{}', nullable=True),
        sa.Column('scaling_config', sa.JSON(), server_default='{}', nullable=True),
        sa.Column('pricing_model', sa.String(20), server_default='per_request', nullable=True),
        sa.Column('price_per_request', sa.DECIMAL(10, 6), nullable=True),
        sa.Column('is_public', sa.Boolean(), server_default='false', nullable=True),
        sa.Column('marketplace_category', sa.String(50), nullable=True),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.text('NOW()'), nullable=True),
        sa.Column('updated_at', sa.DateTime(timezone=True), server_default=sa.text('NOW()'), nullable=True),
        sa.Column('deployed_at', sa.DateTime(timezone=True), nullable=True),
        sa.ForeignKeyConstraint(['user_id'], ['users.id'], ondelete='CASCADE'),
        sa.PrimaryKeyConstraint('id'),
        sa.UniqueConstraint('user_id', 'name')
    )
    
    # Create deployments table
    op.create_table('deployments',
        sa.Column('id', sa.UUID(), server_default=sa.text('uuid_generate_v4()'), nullable=False),
        sa.Column('api_id', sa.UUID(), nullable=False),
        sa.Column('version', sa.String(20), nullable=False),
        sa.Column('status', sa.String(20), nullable=False),
        sa.Column('deployment_method', sa.String(20), nullable=False),
        sa.Column('config_snapshot', sa.JSON(), nullable=True),
        sa.Column('build_logs', sa.Text(), nullable=True),
        sa.Column('build_duration_seconds', sa.Integer(), nullable=True),
        sa.Column('started_at', sa.DateTime(timezone=True), server_default=sa.text('NOW()'), nullable=True),
        sa.Column('completed_at', sa.DateTime(timezone=True), nullable=True),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.text('NOW()'), nullable=True),
        sa.ForeignKeyConstraint(['api_id'], ['apis.id'], ondelete='CASCADE'),
        sa.PrimaryKeyConstraint('id')
    )
    
    # Create subscriptions table
    op.create_table('subscriptions',
        sa.Column('id', sa.UUID(), server_default=sa.text('uuid_generate_v4()'), nullable=False),
        sa.Column('user_id', sa.UUID(), nullable=False),
        sa.Column('plan', sa.String(50), nullable=False),
        sa.Column('status', sa.String(20), server_default='active', nullable=True),
        sa.Column('stripe_subscription_id', sa.String(100), nullable=True),
        sa.Column('stripe_customer_id', sa.String(100), nullable=True),
        sa.Column('amount', sa.DECIMAL(10, 2), nullable=False),
        sa.Column('currency', sa.String(3), server_default='USD', nullable=True),
        sa.Column('billing_interval', sa.String(20), server_default='month', nullable=True),
        sa.Column('current_period_start', sa.DateTime(timezone=True), nullable=True),
        sa.Column('current_period_end', sa.DateTime(timezone=True), nullable=True),
        sa.Column('created_at', sa.DateTime(timezone=True), server_default=sa.text('NOW()'), nullable=True),
        sa.Column('updated_at', sa.DateTime(timezone=True), server_default=sa.text('NOW()'), nullable=True),
        sa.ForeignKeyConstraint(['user_id'], ['users.id'], ondelete='CASCADE'),
        sa.PrimaryKeyConstraint('id'),
        sa.UniqueConstraint('stripe_subscription_id')
    )
    
    # Create indexes
    op.create_index('idx_users_email', 'users', ['email'])
    op.create_index('idx_apis_user_id', 'apis', ['user_id'])
    op.create_index('idx_apis_status', 'apis', ['status'])
    
    # Create update_updated_at_column function
    op.execute("""
        CREATE OR REPLACE FUNCTION update_updated_at_column()
        RETURNS TRIGGER AS $$
        BEGIN
            NEW.updated_at = NOW();
            RETURN NEW;
        END;
        $$ language 'plpgsql';
    """)
    
    # Create triggers
    op.execute("""
        CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
        FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    """)
    
    op.execute("""
        CREATE TRIGGER update_apis_updated_at BEFORE UPDATE ON apis
        FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    """)
    
    op.execute("""
        CREATE TRIGGER update_subscriptions_updated_at BEFORE UPDATE ON subscriptions
        FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
    """)


def downgrade() -> None:
    # Drop triggers
    op.execute('DROP TRIGGER IF EXISTS update_subscriptions_updated_at ON subscriptions')
    op.execute('DROP TRIGGER IF EXISTS update_apis_updated_at ON apis')
    op.execute('DROP TRIGGER IF EXISTS update_users_updated_at ON users')
    
    # Drop function
    op.execute('DROP FUNCTION IF EXISTS update_updated_at_column()')
    
    # Drop indexes
    op.drop_index('idx_apis_status', table_name='apis')
    op.drop_index('idx_apis_user_id', table_name='apis')
    op.drop_index('idx_users_email', table_name='users')
    
    # Drop tables
    op.drop_table('subscriptions')
    op.drop_table('deployments')
    op.drop_table('apis')
    op.drop_table('password_reset_tokens')
    op.drop_table('user_api_keys')
    op.drop_table('users')
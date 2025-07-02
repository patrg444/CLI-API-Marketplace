#!/bin/bash

# Generate secure secrets for production environment

echo "🔐 Generating secure secrets for production environment..."
echo ""

# Function to generate secure random string
generate_secret() {
    openssl rand -base64 $1 | tr -d "=+/" | cut -c1-$1
}

# Generate all required secrets
POSTGRES_PASSWORD=$(generate_secret 32)
REDIS_PASSWORD=$(generate_secret 32)
INFLUXDB_PASSWORD=$(generate_secret 32)
INFLUXDB_TOKEN=$(generate_secret 64)
JWT_SECRET=$(generate_secret 64)
NEXTAUTH_SECRET=$(generate_secret 64)
GRAFANA_PASSWORD=$(generate_secret 20)

# Display generated secrets
echo "Copy these values to your .env.production file:"
echo "================================================"
echo ""
echo "# Database"
echo "POSTGRES_PASSWORD=$POSTGRES_PASSWORD"
echo ""
echo "# Redis"
echo "REDIS_PASSWORD=$REDIS_PASSWORD"
echo ""
echo "# InfluxDB"
echo "INFLUXDB_PASSWORD=$INFLUXDB_PASSWORD"
echo "INFLUXDB_TOKEN=$INFLUXDB_TOKEN"
echo ""
echo "# Security"
echo "JWT_SECRET=$JWT_SECRET"
echo "NEXTAUTH_SECRET=$NEXTAUTH_SECRET"
echo ""
echo "# Monitoring"
echo "GRAFANA_PASSWORD=$GRAFANA_PASSWORD"
echo ""
echo "================================================"
echo ""
echo "⚠️  IMPORTANT: Save these values securely!"
echo "⚠️  You won't be able to recover them once lost."
echo ""
echo "Next steps:"
echo "1. Update .env.production with these values"
echo "2. Configure your AWS credentials"
echo "3. Set up Stripe API keys"
echo "4. Configure email service credentials"
echo ""
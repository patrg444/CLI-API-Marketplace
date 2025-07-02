#!/bin/bash
# Setup Monitoring Stack for API Direct Marketplace

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Setting up monitoring stack...${NC}"

# Create necessary directories
echo "Creating directories..."
mkdir -p "$PROJECT_ROOT/monitoring/grafana/provisioning/dashboards"
mkdir -p "$PROJECT_ROOT/monitoring/grafana/provisioning/datasources"
mkdir -p "$PROJECT_ROOT/monitoring/grafana/provisioning/alerting"
mkdir -p "$PROJECT_ROOT/monitoring/prometheus"
mkdir -p "$PROJECT_ROOT/monitoring/alertmanager"

# Copy Prometheus configuration
echo "Configuring Prometheus..."
cp "$SCRIPT_DIR/prometheus.yml" "$PROJECT_ROOT/monitoring/prometheus/"
cp "$SCRIPT_DIR/prometheus-alerts.yml" "$PROJECT_ROOT/monitoring/prometheus/"

# Update Prometheus config to include alerts
cat >> "$PROJECT_ROOT/monitoring/prometheus/prometheus.yml" <<EOF

# Alert rules
rule_files:
  - '/etc/prometheus/prometheus-alerts.yml'

# Alertmanager configuration
alerting:
  alertmanagers:
    - static_configs:
        - targets:
            - alertmanager:9093
EOF

# Setup Alertmanager
echo "Configuring Alertmanager..."
cp "$SCRIPT_DIR/alertmanager.yml" "$PROJECT_ROOT/monitoring/alertmanager/"

# Create Grafana datasource provisioning
echo "Configuring Grafana datasources..."
cat > "$PROJECT_ROOT/monitoring/grafana/provisioning/datasources/prometheus.yml" <<EOF
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: false
EOF

# Create Grafana dashboard provisioning
echo "Configuring Grafana dashboards..."
cat > "$PROJECT_ROOT/monitoring/grafana/provisioning/dashboards/dashboards.yml" <<EOF
apiVersion: 1

providers:
  - name: 'API Direct Dashboards'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /etc/grafana/provisioning/dashboards
EOF

# Copy dashboards
cp "$SCRIPT_DIR/grafana/dashboards/"*.json "$PROJECT_ROOT/monitoring/grafana/provisioning/dashboards/" 2>/dev/null || true

# Create docker-compose for monitoring stack
echo "Creating monitoring docker-compose..."
cat > "$PROJECT_ROOT/monitoring/docker-compose.monitoring.yml" <<'EOF'
version: '3.8'

services:
  prometheus:
    image: prom/prometheus:latest
    container_name: apidirect-prometheus
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
      - ./prometheus/prometheus-alerts.yml:/etc/prometheus/prometheus-alerts.yml
      - prometheus_data:/prometheus
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=30d'
      - '--web.enable-lifecycle'
    ports:
      - "9090:9090"
    networks:
      - apidirect-network
    restart: unless-stopped

  alertmanager:
    image: prom/alertmanager:latest
    container_name: apidirect-alertmanager
    volumes:
      - ./alertmanager/alertmanager.yml:/etc/alertmanager/alertmanager.yml
      - alertmanager_data:/alertmanager
    command:
      - '--config.file=/etc/alertmanager/alertmanager.yml'
      - '--storage.path=/alertmanager'
    ports:
      - "9093:9093"
    networks:
      - apidirect-network
    restart: unless-stopped

  grafana:
    image: grafana/grafana:latest
    container_name: apidirect-grafana
    environment:
      GF_SECURITY_ADMIN_PASSWORD: ${GRAFANA_PASSWORD:-admin}
      GF_USERS_ALLOW_SIGN_UP: false
      GF_INSTALL_PLUGINS: grafana-clock-panel,grafana-simple-json-datasource,grafana-piechart-panel
    volumes:
      - grafana_data:/var/lib/grafana
      - ./grafana/provisioning:/etc/grafana/provisioning
    ports:
      - "3000:3000"
    networks:
      - apidirect-network
    depends_on:
      - prometheus
    restart: unless-stopped

  node-exporter:
    image: prom/node-exporter:latest
    container_name: apidirect-node-exporter
    volumes:
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - /:/rootfs:ro
    command:
      - '--path.procfs=/host/proc'
      - '--path.rootfs=/rootfs'
      - '--path.sysfs=/host/sys'
      - '--collector.filesystem.mount-points-exclude=^/(sys|proc|dev|host|etc)($$|/)'
    ports:
      - "9100:9100"
    networks:
      - apidirect-network
    restart: unless-stopped

  postgres-exporter:
    image: prometheuscommunity/postgres-exporter
    container_name: apidirect-postgres-exporter
    environment:
      DATA_SOURCE_NAME: "postgresql://apidirect:${POSTGRES_PASSWORD}@postgres:5432/apidirect?sslmode=disable"
    ports:
      - "9187:9187"
    networks:
      - apidirect-network
    restart: unless-stopped

  redis-exporter:
    image: oliver006/redis_exporter
    container_name: apidirect-redis-exporter
    environment:
      REDIS_ADDR: "redis://redis:6379"
      REDIS_PASSWORD: "${REDIS_PASSWORD}"
    ports:
      - "9121:9121"
    networks:
      - apidirect-network
    restart: unless-stopped

volumes:
  prometheus_data:
  alertmanager_data:
  grafana_data:

networks:
  apidirect-network:
    external: true
EOF

# Create monitoring start script
echo "Creating monitoring start script..."
cat > "$PROJECT_ROOT/monitoring/start-monitoring.sh" <<'EOF'
#!/bin/bash
# Start monitoring stack

cd "$(dirname "$0")"

echo "Starting monitoring stack..."
docker-compose -f docker-compose.monitoring.yml up -d

echo "Waiting for services to start..."
sleep 10

echo "Monitoring stack status:"
docker-compose -f docker-compose.monitoring.yml ps

echo ""
echo "Access points:"
echo "- Prometheus: http://localhost:9090"
echo "- Alertmanager: http://localhost:9093"
echo "- Grafana: http://localhost:3000 (admin/${GRAFANA_PASSWORD:-admin})"
echo ""
echo "To view logs: docker-compose -f docker-compose.monitoring.yml logs -f"
EOF

chmod +x "$PROJECT_ROOT/monitoring/start-monitoring.sh"

# Create example alert testing script
echo "Creating alert testing script..."
cat > "$PROJECT_ROOT/monitoring/test-alerts.sh" <<'EOF'
#!/bin/bash
# Test monitoring alerts

echo "Testing alerts..."

# Test critical CPU alert
echo "Generating high CPU load for 2 minutes..."
stress --cpu 8 --timeout 120s &

# Test API error rate
echo "Generating API errors..."
for i in {1..100}; do
    curl -X POST http://localhost:8000/api/test/error -H "X-Test: true" 2>/dev/null
    sleep 0.1
done

echo "Check Alertmanager UI at http://localhost:9093 to see if alerts fired"
echo "Check Grafana at http://localhost:3000 to see metrics"
EOF

chmod +x "$PROJECT_ROOT/monitoring/test-alerts.sh"

echo -e "${GREEN}Monitoring setup complete!${NC}"
echo ""
echo "Next steps:"
echo "1. Update alertmanager.yml with your notification settings"
echo "2. Start monitoring: cd monitoring && ./start-monitoring.sh"
echo "3. Import additional Grafana dashboards as needed"
echo "4. Configure alert notification channels in Grafana"
echo ""
echo -e "${YELLOW}Important: Remember to update alert thresholds based on your baseline metrics${NC}"
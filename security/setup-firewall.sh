#!/bin/bash
# UFW Firewall Setup for API Direct Marketplace

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Setting up UFW firewall for API Direct Marketplace...${NC}"

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}This script must be run as root (use sudo)${NC}"
   exit 1
fi

# Install UFW if not present
if ! command -v ufw &> /dev/null; then
    echo "Installing UFW..."
    apt-get update
    apt-get install -y ufw
fi

# Reset UFW to defaults
echo "Resetting UFW to defaults..."
ufw --force reset

# Default policies
echo "Setting default policies..."
ufw default deny incoming
ufw default allow outgoing

# Allow SSH (important - don't lock yourself out!)
echo -e "${YELLOW}Allowing SSH access...${NC}"
ufw allow 22/tcp comment 'SSH'

# Allow HTTP and HTTPS
echo "Allowing web traffic..."
ufw allow 80/tcp comment 'HTTP'
ufw allow 443/tcp comment 'HTTPS'

# Allow specific ports for monitoring (restricted to local network)
echo "Configuring monitoring access..."
# Prometheus (restrict to internal network)
ufw allow from 10.0.0.0/8 to any port 9090 comment 'Prometheus - Internal'
ufw allow from 172.16.0.0/12 to any port 9090 comment 'Prometheus - Docker'
ufw allow from 192.168.0.0/16 to any port 9090 comment 'Prometheus - Private'

# Grafana (restrict to internal network)
ufw allow from 10.0.0.0/8 to any port 3000 comment 'Grafana - Internal'
ufw allow from 172.16.0.0/12 to any port 3000 comment 'Grafana - Docker'
ufw allow from 192.168.0.0/16 to any port 3000 comment 'Grafana - Private'

# Database ports (only from Docker network)
echo "Restricting database access..."
ufw allow from 172.16.0.0/12 to any port 5432 comment 'PostgreSQL - Docker only'
ufw allow from 172.16.0.0/12 to any port 6379 comment 'Redis - Docker only'
ufw allow from 172.16.0.0/12 to any port 8086 comment 'InfluxDB - Docker only'

# Block common attack ports
echo "Blocking common attack vectors..."
ufw deny 23/tcp comment 'Telnet'
ufw deny 135/tcp comment 'RPC'
ufw deny 139/tcp comment 'NetBIOS'
ufw deny 445/tcp comment 'SMB'
ufw deny 1433/tcp comment 'MSSQL'
ufw deny 3306/tcp comment 'MySQL'
ufw deny 3389/tcp comment 'RDP'

# Rate limiting for SSH
echo "Setting up SSH rate limiting..."
ufw limit ssh/tcp comment 'SSH rate limit'

# Docker specific rules
echo "Configuring Docker rules..."
# Docker API (block external access)
ufw deny 2375/tcp comment 'Docker API'
ufw deny 2376/tcp comment 'Docker API TLS'

# Allow Docker internal communication
if [ -f /etc/default/ufw ]; then
    sed -i 's/DEFAULT_FORWARD_POLICY="DROP"/DEFAULT_FORWARD_POLICY="ACCEPT"/' /etc/default/ufw
fi

# Configure UFW to allow forwarding
if [ -f /etc/ufw/sysctl.conf ]; then
    sed -i 's|#net/ipv4/ip_forward=1|net/ipv4/ip_forward=1|' /etc/ufw/sysctl.conf
fi

# Add Docker rules to UFW
cat > /etc/ufw/applications.d/docker <<EOF
[Docker]
title=Docker
description=Docker container communication
ports=2375,2376/tcp
EOF

# Create custom chains for additional protection
echo "Creating custom firewall chains..."

# DDoS protection rules
iptables -N RATE_LIMIT 2>/dev/null || true
iptables -A RATE_LIMIT -m limit --limit 100/sec --limit-burst 100 -j RETURN
iptables -A RATE_LIMIT -j DROP

# SYN flood protection
iptables -N SYN_FLOOD 2>/dev/null || true
iptables -A SYN_FLOOD -m limit --limit 25/sec --limit-burst 50 -j RETURN
iptables -A SYN_FLOOD -j DROP

# Apply chains to INPUT
iptables -I INPUT -p tcp --syn -j SYN_FLOOD
iptables -I INPUT -p tcp -m tcp --tcp-flags FIN,SYN,RST,ACK SYN -m conntrack --ctstate NEW -j RATE_LIMIT

# Save iptables rules
echo "Saving iptables rules..."
if command -v netfilter-persistent &> /dev/null; then
    netfilter-persistent save
else
    iptables-save > /etc/iptables/rules.v4 2>/dev/null || true
fi

# Enable UFW
echo -e "${GREEN}Enabling UFW...${NC}"
ufw --force enable

# Show status
echo -e "\n${GREEN}Firewall configuration complete!${NC}"
echo "Current firewall status:"
ufw status verbose

# Additional security recommendations
echo -e "\n${YELLOW}Additional security recommendations:${NC}"
echo "1. Change SSH port from 22 to a custom port"
echo "2. Disable root SSH login"
echo "3. Use SSH key authentication only"
echo "4. Set up fail2ban for intrusion prevention"
echo "5. Regularly review firewall logs: /var/log/ufw.log"
echo "6. Configure sysctl for additional kernel-level protection"

# Create monitoring script
cat > /usr/local/bin/monitor-firewall.sh <<'EOF'
#!/bin/bash
# Monitor firewall activity

echo "=== UFW Status ==="
ufw status numbered

echo -e "\n=== Recent Blocked Connections (last 50) ==="
grep -i ufw /var/log/syslog | grep -i block | tail -50

echo -e "\n=== Connection Statistics ==="
ss -s

echo -e "\n=== Top 10 IPs by connection count ==="
netstat -ntu | awk '{print $5}' | cut -d: -f1 | sort | uniq -c | sort -n | tail -10

echo -e "\n=== Firewall Rules Hit Count ==="
iptables -L -v -n | grep -E "^Chain|pkts"
EOF

chmod +x /usr/local/bin/monitor-firewall.sh

echo -e "\n${GREEN}Firewall monitoring script created: /usr/local/bin/monitor-firewall.sh${NC}"
echo "Run it periodically to monitor firewall activity."
# API Direct CLI Reference

## Overview

The API Direct CLI (`apidirect`) is a command-line tool for deploying, managing, and monetizing APIs with minimal DevOps overhead.

## Global Options

These options can be used with any command:

- `--config <file>` - Specify config file (default: `~/.apidirect/config.yaml`)
- `--verbose, -v` - Enable verbose output
- `--help, -h` - Show help for a command
- `--version` - Display version information

## Authentication Commands

### `apidirect auth login`
Authenticate with API Direct platform.

```bash
apidirect auth login
```

Options:
- `--browser` - Open browser for authentication (default)
- `--token <token>` - Use API token directly

### `apidirect auth logout`
Log out from API Direct platform.

```bash
apidirect auth logout
```

### `apidirect auth whoami`
Display current authenticated user.

```bash
apidirect auth whoami
```

## API Development Commands

### `apidirect init`
Initialize a new API project.

```bash
apidirect init [project-name]
```

Options:
- `--template <name>` - Use specific template (python-flask, nodejs-express, go-gin, ruby-sinatra)
- `--minimal` - Create minimal project structure

### `apidirect import`
Import an existing API project.

```bash
apidirect import <path>
```

Options:
- `--name <name>` - Override detected API name
- `--framework <name>` - Specify framework manually
- `--no-docker` - Skip Dockerfile generation

### `apidirect validate`
Validate API configuration before deployment.

```bash
apidirect validate
```

Options:
- `--fix` - Automatically fix common issues
- `--strict` - Enable strict validation

### `apidirect run`
Run API locally for development.

```bash
apidirect run [api-name]
```

Options:
- `--port <port>` - Local port (default: 8080)
- `--watch` - Auto-reload on file changes
- `--env <file>` - Environment file to use

## Deployment Commands

### `apidirect deploy`
Deploy API to cloud infrastructure.

```bash
apidirect deploy [api-name]
```

Options:
- `--hosted` - Use managed infrastructure
- `--environment <env>` - Target environment (dev, staging, production)
- `--no-build` - Skip build step
- `--force` - Force deployment even with warnings

### `apidirect status`
View deployment status.

```bash
apidirect status [api-name]
```

Options:
- `--watch` - Continuously monitor status
- `--json` - Output in JSON format

### `apidirect logs`
View API logs.

```bash
apidirect logs [api-name]
```

Options:
- `--follow, -f` - Stream logs in real-time
- `--since <duration>` - Show logs since duration (e.g., 5m, 1h)
- `--tail <lines>` - Number of lines to show
- `--filter <pattern>` - Filter logs by pattern

### `apidirect scale`
Scale API instances.

```bash
apidirect scale [api-name]
```

Options:
- `--replicas <count>` - Number of instances
- `--min <count>` - Minimum instances for auto-scaling
- `--max <count>` - Maximum instances for auto-scaling
- `--cpu-target <percent>` - Target CPU for auto-scaling

## Environment Management

### `apidirect env list`
List environment variables.

```bash
apidirect env list [--environment <env>]
```

### `apidirect env set`
Set environment variables.

```bash
apidirect env set KEY=value [KEY2=value2...]
```

Options:
- `--environment <env>` - Target environment
- `--file <file>` - Load from .env file

### `apidirect env get`
Get environment variable value.

```bash
apidirect env get KEY
```

### `apidirect env delete`
Delete environment variables.

```bash
apidirect env delete KEY [KEY2...]
```

### `apidirect env pull`
Pull environment variables to local file.

```bash
apidirect env pull [--environment <env>]
```

### `apidirect env push`
Push local environment file to platform.

```bash
apidirect env push <file> [--environment <env>]
```

## Marketplace Commands

### `apidirect publish`
Publish API to marketplace.

```bash
apidirect publish [api-name]
```

Options:
- `--category <category>` - API category
- `--tags <tags>` - Comma-separated tags
- `--private` - Make API private

### `apidirect unpublish`
Remove API from marketplace.

```bash
apidirect unpublish [api-name]
```

### `apidirect pricing set`
Configure API pricing.

```bash
apidirect pricing set [api-name]
```

Options:
- `--plan-file <file>` - Pricing configuration file

### `apidirect pricing get`
View current pricing plans.

```bash
apidirect pricing get [api-name]
```

## Consumer Commands

### `apidirect search`
Search marketplace for APIs.

```bash
apidirect search <query>
```

Options:
- `--category <category>` - Filter by category
- `--sort <field>` - Sort results (relevance, rating, subscribers)
- `--limit <count>` - Number of results

### `apidirect browse`
Browse APIs by category.

```bash
apidirect browse [category]
```

### `apidirect info`
View detailed API information.

```bash
apidirect info <api-name>
```

Options:
- `--show-reviews` - Include recent reviews
- `--json` - Output in JSON format

### `apidirect subscribe`
Subscribe to an API.

```bash
apidirect subscribe <api-name>
```

Options:
- `--plan <plan-id>` - Specific plan to subscribe
- `--trial` - Start with free trial
- `--yes` - Skip confirmation

### `apidirect subscriptions`
Manage your API subscriptions.

```bash
apidirect subscriptions <subcommand>
```

Subcommands:
- `list` - List all subscriptions
- `show <id>` - Show subscription details
- `cancel <id>` - Cancel subscription
- `usage <id>` - View usage statistics
- `keys <id>` - Manage API keys

## Analytics Commands

### `apidirect analytics`
View API analytics.

```bash
apidirect analytics <subcommand> [api-name]
```

Subcommands:
- `usage` - API usage statistics
- `revenue` - Revenue analytics
- `consumers` - Consumer insights
- `performance` - Performance metrics

Options:
- `--period <period>` - Time period (24h, 7d, 30d, custom)
- `--format <format>` - Output format (table, json, csv)

### `apidirect earnings`
Track API earnings.

```bash
apidirect earnings <subcommand>
```

Subcommands:
- `summary` - Earnings overview
- `details` - Detailed breakdown
- `payout` - Request payout
- `history` - Payout history
- `setup` - Configure payout settings

## Review Commands

### `apidirect review`
Manage API reviews.

```bash
apidirect review <subcommand>
```

Subcommands:
- `submit <api-name>` - Submit a review
- `list <api-name>` - List API reviews
- `my` - View your reviews
- `respond <review-id>` - Respond to review
- `report <review-id>` - Report inappropriate review

## Documentation Commands

### `apidirect docs generate`
Generate API documentation.

```bash
apidirect docs generate [api-name]
```

Options:
- `--format <format>` - Output format (openapi, markdown, html, postman)
- `--output <dir>` - Output directory
- `--theme <theme>` - Documentation theme

### `apidirect docs preview`
Preview documentation locally.

```bash
apidirect docs preview [api-name]
```

Options:
- `--port <port>` - Preview server port

### `apidirect docs publish`
Publish documentation online.

```bash
apidirect docs publish [api-name]
```

Options:
- `--custom-domain <domain>` - Use custom domain
- `--private` - Require authentication

## Configuration Commands

### `apidirect config`
Manage CLI configuration.

```bash
apidirect config <subcommand>
```

Subcommands:
- `list` - Show all settings
- `get <key>` - Get setting value
- `set <key> <value>` - Set configuration
- `unset <key>` - Remove setting

### `apidirect completion`
Generate shell completions.

```bash
apidirect completion <shell>
```

Supported shells:
- `bash`
- `zsh`
- `fish`
- `powershell`

### `apidirect self-update`
Update CLI to latest version.

```bash
apidirect self-update
```

Options:
- `--check` - Only check for updates
- `--force` - Force update

## Examples

### Deploy a Python Flask API

```bash
# Import existing Flask app
apidirect import ./my-flask-api

# Set environment variables
apidirect env set DATABASE_URL=postgres://...
apidirect env set SECRET_KEY=mysecret

# Deploy to production
apidirect deploy --environment production

# Monitor deployment
apidirect status --watch
```

### Monetize an API

```bash
# Publish to marketplace
apidirect publish my-api --category "Weather" --tags "forecast,climate"

# Set pricing plans
cat > pricing.json << EOF
{
  "plans": [
    {
      "name": "Free",
      "type": "free",
      "requests_per_month": 1000
    },
    {
      "name": "Pro",
      "type": "subscription",
      "price": 49.99,
      "requests_per_month": 100000
    }
  ]
}
EOF

apidirect pricing set my-api --plan-file pricing.json

# View analytics
apidirect analytics revenue my-api --period 30d
```

### Subscribe to an API

```bash
# Search for weather APIs
apidirect search "weather forecast"

# Get API details
apidirect info weather-pro-api --show-reviews

# Subscribe with free trial
apidirect subscribe weather-pro-api --trial

# View API key
apidirect subscriptions keys <subscription-id>
```

## Environment Variables

- `APIDIRECT_API_ENDPOINT` - API endpoint URL
- `APIDIRECT_AUTH_TOKEN` - Authentication token
- `APIDIRECT_CONFIG_DIR` - Configuration directory
- `APIDIRECT_NO_COLOR` - Disable colored output
- `APIDIRECT_DEBUG` - Enable debug logging

## Exit Codes

- `0` - Success
- `1` - General error
- `2` - Command syntax error
- `3` - Authentication required
- `4` - Resource not found
- `5` - Validation error
- `6` - Network error

## Getting Help

- Run `apidirect --help` for general help
- Run `apidirect <command> --help` for command-specific help
- Visit https://docs.apidirect.io for full documentation
- Report issues at https://github.com/api-direct/cli/issues
# 🚀 API Direct CLI - Quick Start

Get your API deployed in under 5 minutes!

## Installation

### macOS (Homebrew)
```bash
brew install apidirect
```

### One-line installer (All platforms)
```bash
curl -fsSL https://raw.githubusercontent.com/api-direct/cli/main/cli/install.sh | bash
```

### Manual Download
Download from [GitHub Releases](https://github.com/api-direct/cli/releases/latest)

## 🎯 Your First API in 3 Steps

### 1. Create Your API
```bash
# Initialize a new API project
apidirect init my-weather-api

# Follow the interactive prompts to:
# - Choose your framework (Python/FastAPI, Node.js/Express, Go, etc.)
# - Set up authentication
# - Configure endpoints
```

### 2. Deploy to the Cloud
```bash
cd my-weather-api

# Deploy your API
apidirect deploy

# Your API is now live! 🎉
# Example: https://my-weather-api.api-direct.com
```

### 3. Publish to Marketplace (Optional)
```bash
# Set pricing and publish
apidirect pricing set --plan basic --price 9.99 --calls 1000

# Publish to marketplace
apidirect publish
```

## 🔥 Core Commands

| Command | Description | Example |
|---------|-------------|---------|
| `apidirect init` | Create new API project | `apidirect init my-api` |
| `apidirect deploy` | Deploy to production | `apidirect deploy --env production` |
| `apidirect status` | Check deployment status | `apidirect status` |
| `apidirect logs` | View API logs | `apidirect logs --tail 100` |
| `apidirect scale` | Scale your API | `apidirect scale --replicas 5` |

## 💰 Monetization Commands

| Command | Description | Example |
|---------|-------------|---------|
| `apidirect pricing` | Set API pricing | `apidirect pricing set --plan pro --price 29.99` |
| `apidirect publish` | Publish to marketplace | `apidirect publish` |
| `apidirect analytics` | View usage analytics | `apidirect analytics usage` |
| `apidirect earnings` | Check earnings | `apidirect earnings summary` |

## 🛒 Consumer Commands

| Command | Description | Example |
|---------|-------------|---------|
| `apidirect search` | Find APIs | `apidirect search weather` |
| `apidirect subscribe` | Subscribe to API | `apidirect subscribe weather-api` |
| `apidirect subscriptions` | Manage subscriptions | `apidirect subscriptions list` |

## 🔧 Development Workflow

```bash
# 1. Create and develop locally
apidirect init my-api
cd my-api
apidirect run  # Start local dev server

# 2. Test your API
curl http://localhost:8000/api/health

# 3. Deploy when ready
apidirect deploy

# 4. Monitor and scale
apidirect status
apidirect logs
apidirect scale --auto --min 2 --max 10
```

## 🎨 Framework Templates

Choose from popular frameworks:

- **Python**: FastAPI, Flask, Django
- **Node.js**: Express, NestJS, Fastify  
- **Go**: Gin, Echo, Chi
- **Ruby**: Rails API, Sinatra
- **Java**: Spring Boot
- **PHP**: Laravel, Symfony

## 📊 Built-in Features

✅ **Auto-scaling** - Handles traffic spikes automatically  
✅ **Monitoring** - Real-time metrics and logs  
✅ **Security** - API keys, rate limiting, HTTPS  
✅ **Documentation** - Auto-generated from your code  
✅ **Marketplace** - Monetize your APIs instantly  
✅ **Multi-language** - Support for all major languages  

## 🔗 Helpful Links

- 📖 [Full Documentation](https://docs.api-direct.com)
- 🎮 [Interactive Tutorial](https://tutorial.api-direct.com)
- 💬 [Community Discord](https://discord.gg/api-direct)
- 🐛 [Report Issues](https://github.com/api-direct/cli/issues)
- 📧 [Support](mailto:support@api-direct.com)

## 💡 Pro Tips

- Use `apidirect completion bash` to set up shell autocompletion
- Run `apidirect --help` for detailed command information
- Set `APIDIRECT_ENV=development` for local testing
- Use `apidirect validate` to check your configuration

---

**Need help?** Run `apidirect help` or visit our [documentation](https://docs.api-direct.com)!
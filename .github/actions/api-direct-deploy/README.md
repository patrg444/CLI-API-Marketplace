# API-Direct Deploy Action

Deploy your APIs to the API-Direct platform and optionally publish them to the marketplace directly from GitHub Actions.

## Features

- 🚀 **One-click deployment** from GitHub to API-Direct
- 📢 **Automatic marketplace publishing** with metadata
- 🧪 **Built-in API testing** after deployment
- 📋 **Rich deployment summaries** in GitHub
- 💬 **PR preview deployments** with automatic comments
- 🔄 **Multi-environment support** (staging, production)

## Quick Start

### 1. Set up your API-Direct token

Add your API-Direct token as a repository secret:

1. Go to your repository settings
2. Navigate to "Secrets and variables" → "Actions"
3. Add a new secret named `API_DIRECT_TOKEN`
4. Paste your API-Direct API key as the value

### 2. Create a workflow file

Create `.github/workflows/deploy.yml`:

```yaml
name: Deploy to API-Direct

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Deploy to API-Direct
        uses: api-direct/deploy-action@v1
        with:
          api-key: ${{ secrets.API_DIRECT_TOKEN }}
          publish: true
          description: "My awesome API"
          category: "Web Services"
          tags: "api,rest,production"
```

## Usage

### Basic Deployment

```yaml
- name: Deploy to API-Direct
  uses: api-direct/deploy-action@v1
  with:
    api-key: ${{ secrets.API_DIRECT_TOKEN }}
```

### Deploy and Publish to Marketplace

```yaml
- name: Deploy and Publish
  uses: api-direct/deploy-action@v1
  with:
    api-key: ${{ secrets.API_DIRECT_TOKEN }}
    publish: true
    description: "Real-time weather data API"
    category: "Weather"
    tags: "weather,forecast,realtime"
```

### PR Preview Deployments

```yaml
- name: Deploy Preview
  if: github.event_name == 'pull_request'
  uses: api-direct/deploy-action@v1
  with:
    api-key: ${{ secrets.API_DIRECT_TOKEN }}
    api-name: "${{ github.event.repository.name }}-pr-${{ github.event.number }}"
    publish: false
```

### Multi-Environment Setup

```yaml
- name: Deploy to Staging
  if: github.ref == 'refs/heads/develop'
  uses: api-direct/deploy-action@v1
  with:
    api-key: ${{ secrets.API_DIRECT_STAGING_TOKEN }}
    api-name: "${{ github.event.repository.name }}-staging"
    
- name: Deploy to Production
  if: github.ref == 'refs/heads/main'
  uses: api-direct/deploy-action@v1
  with:
    api-key: ${{ secrets.API_DIRECT_TOKEN }}
    publish: true
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `api-key` | API-Direct API key for authentication | ✅ | |
| `api-name` | Name of the API to deploy | ❌ | Repository name |
| `publish` | Whether to publish to marketplace | ❌ | `false` |
| `description` | API description for marketplace | ❌ | |
| `category` | API category for marketplace | ❌ | `General` |
| `tags` | Comma-separated tags for marketplace | ❌ | |
| `working-directory` | Directory containing apidirect.yaml | ❌ | `.` |
| `cli-version` | Version of API-Direct CLI to use | ❌ | `latest` |

## Outputs

| Output | Description |
|--------|-------------|
| `api-url` | URL of the deployed API |
| `marketplace-url` | URL of the marketplace listing (if published) |
| `deployment-id` | Unique deployment identifier |

## Example Workflows

### Complete CI/CD Pipeline

```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.9'
          
      - name: Install dependencies
        run: |
          pip install -r requirements.txt
          
      - name: Run tests
        run: pytest tests/ -v
        
      - name: Run linting
        run: flake8 .

  deploy-staging:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/develop'
    steps:
      - uses: actions/checkout@v4
      
      - name: Deploy to Staging
        uses: api-direct/deploy-action@v1
        with:
          api-key: ${{ secrets.API_DIRECT_STAGING_TOKEN }}
          api-name: "${{ github.event.repository.name }}-staging"

  deploy-production:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v4
      
      - name: Deploy to Production
        uses: api-direct/deploy-action@v1
        with:
          api-key: ${{ secrets.API_DIRECT_TOKEN }}
          publish: true
          description: "Production API for ${{ github.event.repository.name }}"
          category: "Production APIs"
          tags: "production,stable,api"

  preview:
    needs: test
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request'
    steps:
      - uses: actions/checkout@v4
      
      - name: Deploy Preview
        uses: api-direct/deploy-action@v1
        with:
          api-key: ${{ secrets.API_DIRECT_TOKEN }}
          api-name: "${{ github.event.repository.name }}-pr-${{ github.event.number }}"
```

### Release-based Deployment

```yaml
name: Release Deployment

on:
  release:
    types: [published]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Deploy Release
        uses: api-direct/deploy-action@v1
        with:
          api-key: ${{ secrets.API_DIRECT_TOKEN }}
          publish: true
          description: "Release ${{ github.event.release.tag_name }}"
          tags: "release,v${{ github.event.release.tag_name }},stable"
```

## Advanced Configuration

### Custom CLI Version

```yaml
- name: Deploy with specific CLI version
  uses: api-direct/deploy-action@v1
  with:
    api-key: ${{ secrets.API_DIRECT_TOKEN }}
    cli-version: "v1.2.3"
```

### Monorepo Support

```yaml
- name: Deploy API from subdirectory
  uses: api-direct/deploy-action@v1
  with:
    api-key: ${{ secrets.API_DIRECT_TOKEN }}
    working-directory: "./apis/user-service"
    api-name: "user-service"
```

### Using Outputs

```yaml
- name: Deploy API
  id: deploy
  uses: api-direct/deploy-action@v1
  with:
    api-key: ${{ secrets.API_DIRECT_TOKEN }}
    publish: true

- name: Run integration tests
  run: |
    API_URL="${{ steps.deploy.outputs.api-url }}"
    curl -f "$API_URL/health"
    
- name: Notify team
  run: |
    echo "API deployed to: ${{ steps.deploy.outputs.api-url }}"
    echo "Marketplace: ${{ steps.deploy.outputs.marketplace-url }}"
```

## Troubleshooting

### Common Issues

**Authentication Failed**
- Ensure your `API_DIRECT_TOKEN` secret is set correctly
- Verify the token has deployment permissions

**apidirect.yaml Not Found**
- Make sure your repository contains an `apidirect.yaml` file
- Use `working-directory` input if the file is in a subdirectory

**Deployment Timeout**
- Large APIs may take longer to deploy
- Check the API-Direct status page for service issues

**Marketplace Publishing Failed**
- Ensure `description` is provided when `publish: true`
- Check that your API meets marketplace requirements

### Debug Mode

Enable debug logging by setting the `ACTIONS_STEP_DEBUG` secret to `true` in your repository.

## Security

- Never commit API keys to your repository
- Use GitHub secrets for sensitive information
- Rotate API keys regularly
- Use different keys for different environments

## Support

- 📖 [API-Direct Documentation](https://docs.api-direct.io)
- 💬 [Community Discord](https://discord.gg/api-direct)
- 🐛 [Report Issues](https://github.com/api-direct/deploy-action/issues)
- 📧 [Support Email](mailto:support@api-direct.io)

## License

This action is licensed under the MIT License. See [LICENSE](LICENSE) for details.

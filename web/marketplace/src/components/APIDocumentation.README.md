# API Documentation Component

This component provides interactive API documentation using Swagger UI, fully integrated with the marketplace's authentication and subscription system.

## Features

- **Interactive Documentation**: Renders OpenAPI/Swagger specifications with an interactive UI
- **Try It Out**: Subscribed users can test API endpoints directly from the documentation
- **Automatic Authentication**: API keys are automatically injected for authenticated requests
- **Custom Styling**: Themed to match the marketplace design
- **Multiple Format Support**: Handles both JSON and YAML OpenAPI specifications
- **Subscription Aware**: Shows appropriate UI based on user's subscription status

## Usage

```typescript
import APIDocumentation from '@/components/APIDocumentation';

<APIDocumentation
  documentation={apiDocumentation}
  apiKey={userApiKey}
  apiBaseUrl="https://gateway.api-direct.com/api/v1/apis/123"
  isSubscribed={true}
/>
```

## Props

- `documentation`: The API documentation object containing OpenAPI spec
- `apiKey`: The user's API key for this API (optional)
- `apiBaseUrl`: The base URL for API requests
- `isSubscribed`: Whether the user has an active subscription

## OpenAPI Specification Requirements

The OpenAPI specification should include:

1. **Basic Information**:
```yaml
openapi: 3.0.0
info:
  title: Your API Name
  version: 1.0.0
  description: API description
```

2. **Server Configuration** (will be overridden by apiBaseUrl):
```yaml
servers:
  - url: https://api.example.com/v1
    description: Production server
```

3. **Security Schemes**:
```yaml
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key
```

4. **Endpoints**:
```yaml
paths:
  /users:
    get:
      summary: List users
      security:
        - ApiKeyAuth: []
      responses:
        '200':
          description: Successful response
```

## Creator Upload Process

API creators should upload their OpenAPI specification through:
1. CLI: `api-direct publish --openapi ./openapi.yaml`
2. Creator Portal: Upload in the API settings page

## Testing

To test the documentation locally:

1. Ensure you have a valid OpenAPI spec in the database
2. Subscribe to an API to enable "Try it out" functionality
3. The component will automatically inject your API key for requests

## Customization

The component includes extensive CSS customization to match the marketplace theme. To modify styles, update the `swaggerCustomStyles` constant in the component.

## Security Considerations

- API keys are only injected for subscribed users
- The gateway validates API keys before proxying requests
- CORS is handled by the API gateway
- Request URLs are rewritten to use the gateway endpoint

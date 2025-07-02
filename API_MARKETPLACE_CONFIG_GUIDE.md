# API Marketplace Configuration Guide

## Overview
The API-Direct marketplace configuration interface provides comprehensive controls for publishing and monetizing APIs. This guide covers all configuration options available through the console.

## Configuration Sections

### 1. General Information

#### API Logo
- **Format**: JPEG/PNG
- **Max Size**: 500x500px
- **Upload**: Drag & drop or click to browse

#### Categories
Select from predefined marketplace categories:
- Tools
- Data
- Financial
- Social
- Location
- AI/ML
- Media
- Business
- Sports
- Science
- Gaming
- Other

#### Descriptions
- **Short Description** (200 chars max)
  - Used for API card display in marketplace
  - Plain text only
  - Should be concise and compelling

- **Long Description** (optional)
  - Supports Markdown formatting
  - Appears on API listing page
  - Can include:
    - Feature lists
    - Use cases
    - Code examples
    - Technical details

#### Additional Info
- **Website URL**: Link to API documentation or company site
- **Terms of Use URL**: Legal terms for API usage

### 2. Visibility Settings

#### Public/Private Toggle
- **Private**: Only accessible to invited users
- **Public**: Searchable and accessible on API Hub
  - Triggers marketplace listing
  - Subject to marketplace terms

### 3. Base URL Configuration

#### Primary URL
- Production API endpoint
- HTTPS required for public APIs
- Example: `https://api.yourdomain.com`

#### Multiple URLs (Advanced)
- Load balancing across regions
- Failover URLs
- Development/staging endpoints

#### Health Check
- **Path**: Endpoint for monitoring (e.g., `/health`)
- **Frequency**: Daily automated checks
- **Status Indicators**:
  - SUCCESS (green)
  - FAILURE (red)
  - PENDING (yellow)
- **Alerts**: Email notifications on failures

### 4. Endpoints Configuration

#### Endpoint Definition
For each endpoint, specify:
- **Method**: GET, POST, PUT, PATCH, DELETE
- **Path**: URL path with parameters (e.g., `/users/{id}`)
- **Name**: Human-readable endpoint name
- **Description**: What the endpoint does
- **Parameters**:
  - Path parameters
  - Query parameters
  - Request body schema
  - Response schema

#### Endpoint Groups
- Organize related endpoints
- Improve documentation structure
- Example groups:
  - Authentication
  - User Management
  - Data Operations

### 5. Gateway Configuration

#### Gateway DNS
- Provided by API-Direct
- Format: `https://api.api-direct.com/v1/your-api-name`
- Automatic SSL/TLS

#### Firewall Settings
- **X-API-Direct-Proxy-Secret**: Unique header for verification
- **IP Whitelisting**: Allow only API-Direct infrastructure
- **Benefits**:
  - Prevent direct access
  - DDoS protection
  - Request validation

### 6. Security Settings

#### Threat Protection
- **SQL Injection Protection**: Block malicious queries
- **JavaScript Injection Protection**: Sanitize inputs
- **RegEx Pattern Matching**: Custom security rules
- **Content-Type Enforcement**: Require proper headers

#### Request Schema Validation
Three modes available:
1. **Passthrough Everything** (default)
   - Unknown parameters allowed
   - Maximum compatibility

2. **Strip and Passthrough**
   - Remove undefined parameters
   - Clean request forwarding

3. **Block**
   - Reject requests with undefined parameters
   - Strictest validation

#### Request Configurations
- **Size Limit**: 1-50 MB (default: 10 MB)
- **Timeout**: 1-180 seconds (default: 60s)
- **Rate Limiting**: Per-tier configuration

### 7. Authorization

#### API-Direct Standard Auth
- Single API key per developer
- Automatic key management
- Cross-API authentication

#### Custom Authorization
- **OAuth 2.0**: Configure client credentials
- **JWT**: Custom token validation
- **API Keys**: Additional layer
- **Basic Auth**: Username/password

#### Secret Headers & Parameters
Add custom headers/parameters to all requests:
- **Header Name**: Custom header key
- **Value**: Static or dynamic value
- **Type**: Header or Query parameter
- **Use Cases**:
  - Backend API keys
  - Service identifiers
  - Custom auth tokens

### 8. Transformations

#### Request Transformations
- **Add Parameters**: Inject values
- **Remove Parameters**: Strip sensitive data
- **Remap Parameters**: Change names/structure
- **Examples**:
  ```
  Add Header: X-Service-ID = "marketplace"
  Remove Query: internal_debug
  Remap Body: user_name → userName
  ```

#### Response Transformations
- **Filter Fields**: Remove internal data
- **Add Metadata**: Inject response headers
- **Format Conversion**: JSON ↔ XML

### 9. Pricing Configuration

#### Pricing Models

1. **Free**
   - No payment required
   - Optional rate limits
   - Good for open source/demos

2. **Freemium**
   - Free tier with limits
   - Paid tiers for higher usage
   - Most popular model

3. **Pay Per Use**
   - No monthly fee
   - Charge per request
   - Usage-based billing

4. **Paid Only**
   - All tiers require payment
   - No free usage
   - Enterprise focus

#### Tier Configuration (4 Tiers Standard)

For each tier, configure:

##### Basic Settings
- **Tier Name**: Customizable (e.g., BASIC, PRO, ULTRA, MEGA)
- **Monthly Price**: $0.00 - $9999.99
- **Position**: Order in pricing table

##### Quotas
- **Requests/Month**: Included requests
  - Example: 100, 1000, 5000, 20000
- **Overage Pricing**: Cost per additional request
  - Example: $0.01, $0.008, $0.005, $0.003
- **Rate Limits**: Requests per second/minute/hour
  - Example: 10/sec, 30/sec, 60/sec, 100/sec

##### Features
Tier-specific features:
- **Support Level**: Basic, Priority, Premium, Dedicated
- **SSL Encryption**: Standard for all
- **Analytics**: Basic, Advanced, Custom
- **SLA**: Uptime guarantees
- **Custom Domain**: Higher tiers only
- **Dedicated Resources**: Enterprise tiers

#### Object Definition
Define what counts as one billable request:
- Standard: "A call to any endpoint is one request"
- Custom definitions:
  - Batch operations counting
  - Data volume considerations
  - Compute time factors

#### Global Features
Features available across tiers:
- **All Tiers**: Core functionality
- **Paid Tiers Only**: Premium features
- **Custom Selection**: Specific tier combinations

### 10. Subscriber Management

#### Subscriber View
- **Username**: Developer account
- **Status**: Active, Suspended, Cancelled
- **Subscription Date**: When they joined
- **Plan**: Current pricing tier
- **Total Paid**: Lifetime value
- **Last Active**: Recent API usage

#### Subscriber Actions
- **Upgrade/Downgrade**: Change tiers
- **Usage Reports**: Detailed analytics
- **Communication**: Direct messaging
- **Suspension**: For violations

## Best Practices

### 1. Pricing Strategy
- Start with competitive free tier
- Clear value progression between tiers
- Reasonable overage charges
- Consider market standards

### 2. Documentation
- Complete endpoint documentation
- Clear authentication guide
- Code examples in multiple languages
- Error response documentation

### 3. Security
- Enable all protection features
- Use secret headers for backend auth
- Regular health check monitoring
- IP whitelisting when possible

### 4. Performance
- Set appropriate timeouts
- Configure rate limits per tier
- Monitor usage patterns
- Scale resources as needed

## Publishing Checklist

Before making your API public:

- [ ] Logo uploaded and looks professional
- [ ] Descriptions are clear and compelling
- [ ] All endpoints documented
- [ ] Pricing tiers configured
- [ ] Features clearly differentiated
- [ ] Health check passing
- [ ] Security settings enabled
- [ ] Terms of use provided
- [ ] Test with sample requests
- [ ] Review competitor pricing

## Revenue Optimization

### Tier Design
- **Free Tier**: Generous enough to attract users
- **First Paid Tier**: Clear value jump from free
- **Higher Tiers**: Enterprise features
- **Overage Pricing**: Encourage upgrades

### Feature Differentiation
- **Performance**: Higher rate limits
- **Support**: Response time SLAs
- **Features**: Advanced functionality
- **Resources**: Dedicated infrastructure

### Monitoring & Adjustment
- Track conversion rates
- Monitor tier distribution
- Adjust limits based on usage
- A/B test pricing changes

## Support

For additional help:
- Documentation: docs.api-direct.com
- Support: support@api-direct.com
- Community: forum.api-direct.com
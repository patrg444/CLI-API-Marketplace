// Test Data Factory for generating consistent test data
// Uses faker.js for realistic data generation

const { faker } = require('@faker-js/faker');
const crypto = require('crypto');

class TestDataFactory {
  constructor(seed = null) {
    if (seed) {
      faker.seed(seed);
    }
  }
  
  // User factory
  createUser(overrides = {}) {
    const firstName = faker.person.firstName();
    const lastName = faker.person.lastName();
    const username = faker.internet.userName({ firstName, lastName }).toLowerCase();
    
    return {
      id: this.generateId('user'),
      username,
      email: faker.internet.email({ firstName, lastName }).toLowerCase(),
      password: 'TestPassword123!', // Default test password
      firstName,
      lastName,
      avatar: faker.image.avatar(),
      bio: faker.lorem.paragraph(),
      role: 'consumer',
      createdAt: faker.date.past(),
      updatedAt: faker.date.recent(),
      emailVerified: true,
      twoFactorEnabled: false,
      preferences: {
        theme: faker.helpers.arrayElement(['light', 'dark', 'auto']),
        emailNotifications: true,
        marketingEmails: false
      },
      ...overrides
    };
  }
  
  // API Creator factory
  createCreator(overrides = {}) {
    const user = this.createUser({ role: 'creator' });
    
    return {
      ...user,
      company: faker.company.name(),
      website: faker.internet.url(),
      github: `https://github.com/${user.username}`,
      verified: faker.datatype.boolean(),
      totalAPIs: faker.number.int({ min: 1, max: 20 }),
      totalRevenue: faker.number.float({ min: 0, max: 50000, precision: 0.01 }),
      payoutDetails: {
        method: faker.helpers.arrayElement(['bank_transfer', 'paypal', 'stripe']),
        accountId: this.generateId('payout'),
        verified: true
      },
      ...overrides
    };
  }
  
  // API factory
  createAPI(overrides = {}) {
    const name = faker.helpers.slugify(faker.commerce.productName()).toLowerCase();
    const category = faker.helpers.arrayElement(['ai-ml', 'data', 'finance', 'weather', 'media', 'tools']);
    
    return {
      id: name,
      name,
      displayName: faker.commerce.productName(),
      author: `@${faker.internet.userName().toLowerCase()}`,
      description: faker.commerce.productDescription(),
      category,
      icon: this.getIconForCategory(category),
      color: faker.color.human(),
      version: faker.system.semver(),
      status: faker.helpers.arrayElement(['active', 'beta', 'deprecated']),
      visibility: faker.helpers.arrayElement(['public', 'private', 'unlisted']),
      
      // Metrics
      rating: faker.number.float({ min: 3.5, max: 5, precision: 0.1 }),
      reviews: faker.number.int({ min: 0, max: 500 }),
      calls: faker.number.int({ min: 0, max: 1000000 }),
      uniqueUsers: faker.number.int({ min: 0, max: 10000 }),
      
      // Pricing
      pricing: this.generatePricing(),
      
      // Features
      tags: this.generateTags(category),
      featured: faker.datatype.boolean(0.3),
      trending: faker.datatype.boolean(0.2),
      growth: faker.number.int({ min: -50, max: 500 }),
      
      // Technical details
      endpoints: this.generateEndpoints(),
      authentication: faker.helpers.arrayElement(['api_key', 'oauth2', 'jwt']),
      rateLimits: {
        requests: faker.number.int({ min: 100, max: 10000 }),
        period: 'hour'
      },
      
      // Timestamps
      createdAt: faker.date.past(),
      updatedAt: faker.date.recent(),
      lastDeployment: faker.date.recent(),
      
      ...overrides
    };
  }
  
  // Subscription factory
  createSubscription(userId, apiId, overrides = {}) {
    const plan = faker.helpers.arrayElement(['free', 'starter', 'pro', 'enterprise']);
    
    return {
      id: this.generateId('sub'),
      userId,
      apiId,
      plan,
      status: faker.helpers.arrayElement(['active', 'cancelled', 'past_due']),
      
      // Billing
      billingCycle: plan === 'free' ? null : faker.helpers.arrayElement(['monthly', 'yearly']),
      amount: this.getPriceForPlan(plan),
      currency: 'USD',
      nextBillingDate: faker.date.future(),
      
      // Usage
      usage: {
        calls: faker.number.int({ min: 0, max: 10000 }),
        period: 'current_month',
        limit: this.getLimitForPlan(plan)
      },
      
      // Keys
      apiKeys: [
        {
          id: this.generateId('key'),
          key: this.generateAPIKey(),
          name: 'Production Key',
          createdAt: faker.date.past(),
          lastUsed: faker.date.recent()
        }
      ],
      
      createdAt: faker.date.past(),
      updatedAt: faker.date.recent(),
      
      ...overrides
    };
  }
  
  // Review factory
  createReview(userId, apiId, overrides = {}) {
    return {
      id: this.generateId('review'),
      userId,
      apiId,
      rating: faker.number.int({ min: 1, max: 5 }),
      title: faker.lorem.sentence(),
      comment: faker.lorem.paragraph(),
      helpful: faker.number.int({ min: 0, max: 50 }),
      verified: true,
      
      // Response from creator
      response: faker.datatype.boolean(0.3) ? {
        comment: faker.lorem.paragraph(),
        createdAt: faker.date.recent()
      } : null,
      
      createdAt: faker.date.past(),
      updatedAt: faker.date.recent(),
      
      ...overrides
    };
  }
  
  // Transaction factory
  createTransaction(userId, overrides = {}) {
    const type = faker.helpers.arrayElement(['subscription', 'usage', 'payout', 'refund']);
    
    return {
      id: this.generateId('txn'),
      userId,
      type,
      status: faker.helpers.arrayElement(['completed', 'pending', 'failed']),
      amount: faker.number.float({ min: 0.01, max: 1000, precision: 0.01 }),
      currency: 'USD',
      description: this.getTransactionDescription(type),
      
      // Payment details
      paymentMethod: type === 'payout' ? null : {
        type: faker.helpers.arrayElement(['card', 'paypal', 'bank']),
        last4: faker.finance.creditCardNumber('####'),
        brand: faker.helpers.arrayElement(['visa', 'mastercard', 'amex'])
      },
      
      // Metadata
      metadata: {
        apiId: type === 'subscription' ? this.generateId('api') : null,
        invoiceId: this.generateId('inv'),
        period: type === 'subscription' ? faker.date.month() : null
      },
      
      createdAt: faker.date.past(),
      
      ...overrides
    };
  }
  
  // Analytics data factory
  createAnalyticsData(apiId, dateRange = 30) {
    const data = [];
    const endDate = new Date();
    
    for (let i = 0; i < dateRange; i++) {
      const date = new Date(endDate);
      date.setDate(date.getDate() - i);
      
      data.push({
        date: date.toISOString().split('T')[0],
        apiId,
        metrics: {
          calls: faker.number.int({ min: 100, max: 5000 }),
          uniqueUsers: faker.number.int({ min: 10, max: 500 }),
          errors: faker.number.int({ min: 0, max: 50 }),
          avgLatency: faker.number.int({ min: 50, max: 500 }),
          successRate: faker.number.float({ min: 95, max: 100, precision: 0.1 })
        },
        endpoints: this.generateEndpointMetrics(),
        geography: this.generateGeographyMetrics()
      });
    }
    
    return data.reverse(); // Chronological order
  }
  
  // Helper methods
  generateId(prefix) {
    return `${prefix}_${crypto.randomBytes(12).toString('hex')}`;
  }
  
  generateAPIKey() {
    return `api_${crypto.randomBytes(32).toString('hex')}`;
  }
  
  generatePricing() {
    const type = faker.helpers.arrayElement(['free', 'freemium', 'subscription', 'usage']);
    
    switch (type) {
      case 'free':
        return { type, limits: { calls: 1000, rateLimit: 10 } };
      
      case 'freemium':
        return {
          type,
          freeCalls: faker.number.int({ min: 100, max: 5000 }),
          pricePerCall: faker.number.float({ min: 0.0001, max: 0.01, precision: 0.0001 }),
          currency: 'USD'
        };
      
      case 'subscription':
        return {
          type,
          plans: [
            { name: 'Starter', price: 9.99, calls: 10000 },
            { name: 'Pro', price: 49.99, calls: 100000 },
            { name: 'Enterprise', price: 299.99, calls: 1000000 }
          ],
          currency: 'USD'
        };
      
      case 'usage':
        return {
          type,
          pricePerCall: faker.number.float({ min: 0.001, max: 0.1, precision: 0.001 }),
          minimumCharge: faker.number.float({ min: 1, max: 10, precision: 0.01 }),
          currency: 'USD'
        };
    }
  }
  
  generateTags(category) {
    const baseTags = {
      'ai-ml': ['ai', 'machine-learning', 'nlp', 'computer-vision'],
      'data': ['database', 'analytics', 'etl', 'big-data'],
      'finance': ['fintech', 'payments', 'crypto', 'trading'],
      'weather': ['forecast', 'climate', 'meteorology', 'real-time'],
      'media': ['image', 'video', 'audio', 'streaming'],
      'tools': ['utility', 'automation', 'productivity', 'developer-tools']
    };
    
    const tags = baseTags[category] || [];
    const additionalTags = faker.helpers.arrayElements(
      ['api', 'rest', 'graphql', 'websocket', 'real-time', 'batch', 'async'],
      faker.number.int({ min: 1, max: 3 })
    );
    
    return [...tags, ...additionalTags];
  }
  
  generateEndpoints() {
    const count = faker.number.int({ min: 1, max: 10 });
    const endpoints = [];
    
    for (let i = 0; i < count; i++) {
      endpoints.push({
        path: `/${faker.helpers.slugify(faker.hacker.noun())}`,
        method: faker.helpers.arrayElement(['GET', 'POST', 'PUT', 'DELETE']),
        description: faker.hacker.phrase()
      });
    }
    
    return endpoints;
  }
  
  generateEndpointMetrics() {
    return {
      '/search': faker.number.int({ min: 100, max: 2000 }),
      '/get': faker.number.int({ min: 50, max: 1000 }),
      '/create': faker.number.int({ min: 10, max: 500 }),
      '/update': faker.number.int({ min: 5, max: 200 }),
      '/delete': faker.number.int({ min: 1, max: 50 })
    };
  }
  
  generateGeographyMetrics() {
    return {
      'US': faker.number.int({ min: 100, max: 2000 }),
      'EU': faker.number.int({ min: 50, max: 1500 }),
      'ASIA': faker.number.int({ min: 50, max: 1000 }),
      'OTHER': faker.number.int({ min: 10, max: 500 })
    };
  }
  
  getIconForCategory(category) {
    const icons = {
      'ai-ml': 'brain',
      'data': 'database',
      'finance': 'credit-card',
      'weather': 'cloud-sun',
      'media': 'image',
      'tools': 'wrench'
    };
    return icons[category] || 'cube';
  }
  
  getPriceForPlan(plan) {
    const prices = {
      free: 0,
      starter: 9.99,
      pro: 49.99,
      enterprise: 299.99
    };
    return prices[plan] || 0;
  }
  
  getLimitForPlan(plan) {
    const limits = {
      free: 1000,
      starter: 10000,
      pro: 100000,
      enterprise: 1000000
    };
    return limits[plan] || 0;
  }
  
  getTransactionDescription(type) {
    const descriptions = {
      subscription: 'Monthly subscription payment',
      usage: 'Usage-based charges',
      payout: 'Creator earnings payout',
      refund: 'Refund for cancelled subscription'
    };
    return descriptions[type] || 'Transaction';
  }
  
  // Batch creation methods
  createUsers(count, overrides = {}) {
    return Array.from({ length: count }, () => this.createUser(overrides));
  }
  
  createAPIs(count, overrides = {}) {
    return Array.from({ length: count }, () => this.createAPI(overrides));
  }
  
  createCompleteTestDataset() {
    // Create interconnected test data
    const users = this.createUsers(20);
    const creators = this.createUsers(5, { role: 'creator' });
    const apis = this.createAPIs(15);
    
    const subscriptions = [];
    const reviews = [];
    const transactions = [];
    
    // Create subscriptions and reviews
    users.forEach(user => {
      // Each user subscribes to 1-3 APIs
      const apiCount = faker.number.int({ min: 1, max: 3 });
      const selectedApis = faker.helpers.arrayElements(apis, apiCount);
      
      selectedApis.forEach(api => {
        subscriptions.push(this.createSubscription(user.id, api.id));
        
        // 50% chance of leaving a review
        if (faker.datatype.boolean()) {
          reviews.push(this.createReview(user.id, api.id));
        }
        
        // Create transaction for subscription
        transactions.push(this.createTransaction(user.id, {
          type: 'subscription',
          metadata: { apiId: api.id }
        }));
      });
    });
    
    // Create payout transactions for creators
    creators.forEach(creator => {
      transactions.push(this.createTransaction(creator.id, {
        type: 'payout',
        amount: faker.number.float({ min: 100, max: 5000, precision: 0.01 })
      }));
    });
    
    return {
      users: [...users, ...creators],
      apis,
      subscriptions,
      reviews,
      transactions,
      analytics: apis.map(api => ({
        apiId: api.id,
        data: this.createAnalyticsData(api.id)
      }))
    };
  }
}

module.exports = TestDataFactory;
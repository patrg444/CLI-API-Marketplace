// API Contract Testing using Joi for schema validation
// This ensures API responses match expected contracts

const Joi = require('joi');
const axios = require('axios');

// API Base URL
const API_BASE_URL = process.env.API_URL || 'http://localhost:3000/api';

// Define API response schemas
const schemas = {
  // Category schema
  category: Joi.object({
    id: Joi.string().required(),
    name: Joi.string().required(),
    icon: Joi.string().required(),
    count: Joi.number().integer().min(0).required()
  }),
  
  // API pricing schema
  pricing: Joi.alternatives().try(
    Joi.object({
      type: Joi.string().valid('freemium').required(),
      freeCalls: Joi.number().integer().min(0).required(),
      pricePerCall: Joi.number().min(0).required(),
      currency: Joi.string().valid('USD', 'EUR', 'GBP').required()
    }),
    Joi.object({
      type: Joi.string().valid('subscription').required(),
      monthlyPrice: Joi.number().min(0).required(),
      calls: Joi.number().integer().min(0).required(),
      currency: Joi.string().valid('USD', 'EUR', 'GBP').required()
    })
  ),
  
  // Full API schema
  api: Joi.object({
    id: Joi.string().required(),
    name: Joi.string().required(),
    author: Joi.string().pattern(/^@[\w-]+$/).required(),
    description: Joi.string().min(10).max(500).required(),
    category: Joi.string().required(),
    icon: Joi.string().required(),
    color: Joi.string().required(),
    rating: Joi.number().min(0).max(5).required(),
    reviews: Joi.number().integer().min(0).required(),
    calls: Joi.number().integer().min(0).required(),
    pricing: Joi.alternatives().try(
      Joi.object({
        type: Joi.string().valid('freemium').required(),
        freeCalls: Joi.number().integer().min(0).required(),
        pricePerCall: Joi.number().min(0).required(),
        currency: Joi.string().valid('USD', 'EUR', 'GBP').required()
      }),
      Joi.object({
        type: Joi.string().valid('subscription').required(),
        monthlyPrice: Joi.number().min(0).required(),
        calls: Joi.number().integer().min(0).required(),
        currency: Joi.string().valid('USD', 'EUR', 'GBP').required()
      })
    ).required(),
    tags: Joi.array().items(Joi.string()).min(1).required(),
    featured: Joi.boolean().required(),
    trending: Joi.boolean().required(),
    growth: Joi.number().optional()
  }),
  
  // Response wrapper schemas
  successResponse: (dataSchema) => Joi.object({
    success: Joi.boolean().valid(true).required(),
    data: dataSchema.required()
  }),
  
  paginatedResponse: (dataSchema) => Joi.object({
    success: Joi.boolean().valid(true).required(),
    data: Joi.array().items(dataSchema).required(),
    meta: Joi.object({
      total: Joi.number().integer().min(0).required(),
      page: Joi.number().integer().min(1).required(),
      limit: Joi.number().integer().min(1).required(),
      totalPages: Joi.number().integer().min(0).required()
    }).required()
  }),
  
  errorResponse: Joi.object({
    success: Joi.boolean().valid(false).required(),
    error: Joi.string().required()
  })
};

// Contract test suite
class APIContractTester {
  constructor(baseUrl = API_BASE_URL) {
    this.baseUrl = baseUrl;
    this.client = axios.create({
      baseURL: baseUrl,
      timeout: 5000,
      validateStatus: () => true // Don't throw on any status
    });
  }
  
  async runAllTests() {
    console.log('🤝 Running API Contract Tests...\n');
    
    const results = {
      total: 0,
      passed: 0,
      failed: 0,
      tests: []
    };
    
    // Run all contract tests
    const tests = [
      this.testCategoriesEndpoint(),
      this.testApisEndpoint(),
      this.testApiFiltering(),
      this.testApiPagination(),
      this.testFeaturedApisEndpoint(),
      this.testTrendingApisEndpoint(),
      this.testSpecificApiEndpoint(),
      this.testErrorResponses(),
      this.testEdgeCases()
    ];
    
    for (const test of tests) {
      const result = await test;
      results.total++;
      if (result.passed) {
        results.passed++;
        console.log(`✅ ${result.name}`);
      } else {
        results.failed++;
        console.log(`❌ ${result.name}: ${result.error}`);
      }
      results.tests.push(result);
    }
    
    // Print summary
    console.log('\n📊 Contract Test Summary:');
    console.log(`Total: ${results.total}`);
    console.log(`Passed: ${results.passed}`);
    console.log(`Failed: ${results.failed}`);
    console.log(`Success Rate: ${((results.passed / results.total) * 100).toFixed(2)}%`);
    
    return results;
  }
  
  async testCategoriesEndpoint() {
    const testName = 'GET /categories contract';
    try {
      const response = await this.client.get('/categories');
      
      // Validate response schema
      const schema = schemas.successResponse(
        Joi.array().items(schemas.category)
      );
      
      const { error } = schema.validate(response.data);
      if (error) throw new Error(error.details[0].message);
      
      // Additional business logic validations
      const categories = response.data.data;
      if (categories.length === 0) {
        throw new Error('No categories returned');
      }
      
      // Check for duplicate IDs
      const ids = categories.map(c => c.id);
      if (new Set(ids).size !== ids.length) {
        throw new Error('Duplicate category IDs found');
      }
      
      return { name: testName, passed: true };
    } catch (error) {
      return { name: testName, passed: false, error: error.message };
    }
  }
  
  async testApisEndpoint() {
    const testName = 'GET /apis contract';
    try {
      const response = await this.client.get('/apis');
      
      // Validate paginated response
      const schema = schemas.paginatedResponse(schemas.api);
      const { error } = schema.validate(response.data);
      if (error) throw new Error(error.details[0].message);
      
      // Validate pagination logic
      const { data, meta } = response.data;
      if (data.length > meta.limit) {
        throw new Error('Returned more items than limit');
      }
      
      if (meta.totalPages !== Math.ceil(meta.total / meta.limit)) {
        throw new Error('Invalid totalPages calculation');
      }
      
      return { name: testName, passed: true };
    } catch (error) {
      return { name: testName, passed: false, error: error.message };
    }
  }
  
  async testApiFiltering() {
    const testName = 'API filtering contract';
    try {
      // Test category filter
      const categoryResponse = await this.client.get('/apis?category=ai-ml');
      const { error: catError } = schemas.paginatedResponse(schemas.api)
        .validate(categoryResponse.data);
      if (catError) throw new Error(catError.details[0].message);
      
      // Verify all returned APIs match category
      const wrongCategory = categoryResponse.data.data.find(
        api => api.category !== 'ai-ml'
      );
      if (wrongCategory) {
        throw new Error('API with wrong category returned');
      }
      
      // Test search filter
      const searchResponse = await this.client.get('/apis?search=weather');
      const { error: searchError } = schemas.paginatedResponse(schemas.api)
        .validate(searchResponse.data);
      if (searchError) throw new Error(searchError.details[0].message);
      
      // Test price filter
      const priceResponse = await this.client.get('/apis?maxPrice=0.005');
      const { error: priceError } = schemas.paginatedResponse(schemas.api)
        .validate(priceResponse.data);
      if (priceError) throw new Error(priceError.details[0].message);
      
      return { name: testName, passed: true };
    } catch (error) {
      return { name: testName, passed: false, error: error.message };
    }
  }
  
  async testApiPagination() {
    const testName = 'API pagination contract';
    try {
      // Test different pages
      const page1 = await this.client.get('/apis?page=1&limit=5');
      const page2 = await this.client.get('/apis?page=2&limit=5');
      
      // Validate both responses
      const schema = schemas.paginatedResponse(schemas.api);
      const { error: error1 } = schema.validate(page1.data);
      if (error1) throw new Error(`Page 1: ${error1.details[0].message}`);
      
      const { error: error2 } = schema.validate(page2.data);
      if (error2) throw new Error(`Page 2: ${error2.details[0].message}`);
      
      // Ensure no duplicate items across pages
      const ids1 = page1.data.data.map(api => api.id);
      const ids2 = page2.data.data.map(api => api.id);
      const duplicates = ids1.filter(id => ids2.includes(id));
      
      if (duplicates.length > 0) {
        throw new Error('Duplicate items across pages');
      }
      
      return { name: testName, passed: true };
    } catch (error) {
      return { name: testName, passed: false, error: error.message };
    }
  }
  
  async testFeaturedApisEndpoint() {
    const testName = 'GET /apis/featured contract';
    try {
      const response = await this.client.get('/apis/featured');
      
      const schema = schemas.successResponse(
        Joi.array().items(schemas.api)
      );
      const { error } = schema.validate(response.data);
      if (error) throw new Error(error.details[0].message);
      
      // Verify all APIs are featured
      const nonFeatured = response.data.data.find(api => !api.featured);
      if (nonFeatured) {
        throw new Error('Non-featured API in featured endpoint');
      }
      
      return { name: testName, passed: true };
    } catch (error) {
      return { name: testName, passed: false, error: error.message };
    }
  }
  
  async testTrendingApisEndpoint() {
    const testName = 'GET /apis/trending contract';
    try {
      const response = await this.client.get('/apis/trending');
      
      const schema = schemas.successResponse(
        Joi.array().items(schemas.api)
      );
      const { error } = schema.validate(response.data);
      if (error) throw new Error(error.details[0].message);
      
      // Verify all APIs are trending
      const nonTrending = response.data.data.find(api => !api.trending);
      if (nonTrending) {
        throw new Error('Non-trending API in trending endpoint');
      }
      
      // Verify sorted by growth
      const growthValues = response.data.data.map(api => api.growth || 0);
      const isSorted = growthValues.every((val, i, arr) => 
        i === 0 || arr[i-1] >= val
      );
      if (!isSorted) {
        throw new Error('Trending APIs not sorted by growth');
      }
      
      return { name: testName, passed: true };
    } catch (error) {
      return { name: testName, passed: false, error: error.message };
    }
  }
  
  async testSpecificApiEndpoint() {
    const testName = 'GET /apis/:id contract';
    try {
      // Test with known API
      const response = await this.client.get('/apis/global-weather-api');
      
      if (response.status === 200) {
        const schema = schemas.successResponse(schemas.api);
        const { error } = schema.validate(response.data);
        if (error) throw new Error(error.details[0].message);
        
        if (response.data.data.id !== 'global-weather-api') {
          throw new Error('Returned wrong API');
        }
      } else {
        throw new Error(`Unexpected status: ${response.status}`);
      }
      
      // Test with non-existent API
      const errorResponse = await this.client.get('/apis/non-existent');
      if (errorResponse.status !== 404) {
        throw new Error('Should return 404 for non-existent API');
      }
      
      const { error: errorSchemaError } = schemas.errorResponse
        .validate(errorResponse.data);
      if (errorSchemaError) {
        throw new Error(`Error response: ${errorSchemaError.details[0].message}`);
      }
      
      return { name: testName, passed: true };
    } catch (error) {
      return { name: testName, passed: false, error: error.message };
    }
  }
  
  async testErrorResponses() {
    const testName = 'Error response contracts';
    try {
      // Test various error scenarios
      const errorScenarios = [
        { url: '/invalid-endpoint', expectedStatus: 404 },
        { url: '/apis?page=-1', expectedStatus: 200 }, // Should handle gracefully
        { url: '/apis/../../etc/passwd', expectedStatus: 404 } // Path traversal
      ];
      
      for (const scenario of errorScenarios) {
        const response = await this.client.get(scenario.url);
        
        if (response.status !== scenario.expectedStatus) {
          throw new Error(
            `${scenario.url} returned ${response.status}, expected ${scenario.expectedStatus}`
          );
        }
        
        // Validate error response format for 4xx/5xx
        if (response.status >= 400) {
          const { error } = schemas.errorResponse.validate(response.data);
          if (error) {
            throw new Error(`Invalid error format: ${error.details[0].message}`);
          }
        }
      }
      
      return { name: testName, passed: true };
    } catch (error) {
      return { name: testName, passed: false, error: error.message };
    }
  }
  
  async testEdgeCases() {
    const testName = 'Edge case contracts';
    try {
      // Test empty results
      const emptySearch = await this.client.get('/apis?search=xyz123abc456');
      const { error: emptyError } = schemas.paginatedResponse(schemas.api)
        .validate(emptySearch.data);
      if (emptyError) throw new Error(`Empty search: ${emptyError.details[0].message}`);
      
      if (emptySearch.data.data.length !== 0) {
        throw new Error('Should return empty array for no matches');
      }
      
      // Test large pagination
      const largePage = await this.client.get('/apis?page=999&limit=100');
      const { error: pageError } = schemas.paginatedResponse(schemas.api)
        .validate(largePage.data);
      if (pageError) throw new Error(`Large page: ${pageError.details[0].message}`);
      
      // Test special characters in search
      const specialChars = await this.client.get('/apis?search=' + encodeURIComponent('test!@#$%'));
      if (specialChars.status !== 200) {
        throw new Error('Should handle special characters in search');
      }
      
      return { name: testName, passed: true };
    } catch (error) {
      return { name: testName, passed: false, error: error.message };
    }
  }
}

// Export for use in other tests
module.exports = {
  APIContractTester,
  schemas
};

// Run tests if called directly
if (require.main === module) {
  const tester = new APIContractTester();
  tester.runAllTests().then(results => {
    process.exit(results.failed > 0 ? 1 : 0);
  });
}
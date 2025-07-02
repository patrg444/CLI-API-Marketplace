// Test script to verify marketplace API endpoints
const handler = require('./api/index.js').default;

// Mock request/response objects
const mockReq = (url) => ({
  url,
  method: 'GET',
  headers: { host: 'localhost:3000' }
});

const mockRes = () => {
  const res = {
    statusCode: 200,
    headers: {},
    status: function(code) {
      this.statusCode = code;
      return this;
    },
    setHeader: function(key, value) {
      this.headers[key] = value;
      return this;
    },
    setHeaders: function(headers) {
      Object.assign(this.headers, headers);
      return this;
    },
    json: function(data) {
      console.log(`Status: ${this.statusCode}`);
      console.log('Response:', JSON.stringify(data, null, 2));
      return this;
    },
    end: function() {
      console.log('Response ended');
      return this;
    }
  };
  return res;
};

// Test endpoints
console.log('Testing Marketplace API Endpoints...\n');

console.log('1. Testing /api/categories:');
handler(mockReq('/api/categories'), mockRes());

console.log('\n2. Testing /api/apis:');
handler(mockReq('/api/apis'), mockRes());

console.log('\n3. Testing /api/apis/featured:');
handler(mockReq('/api/apis/featured'), mockRes());

console.log('\n4. Testing /api/apis/trending:');
handler(mockReq('/api/apis/trending'), mockRes());

console.log('\n5. Testing /api/apis with search:');
handler(mockReq('/api/apis?search=weather'), mockRes());

console.log('\n6. Testing specific API:');
handler(mockReq('/api/apis/global-weather-api'), mockRes());
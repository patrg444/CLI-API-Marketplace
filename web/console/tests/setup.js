// Test setup file

// Add any global test setup here
global.console = {
  ...console,
  // Uncomment to suppress console logs during tests
  // log: jest.fn(),
  error: jest.fn(),
  warn: jest.fn(),
};

// Mock window.location
delete window.location;
window.location = {
  href: '',
  hostname: 'localhost',
  pathname: '/',
  search: '',
  hash: '',
  reload: jest.fn(),
};

// Mock Chart.js if used
global.Chart = jest.fn();

// Add custom matchers if needed
expect.extend({
  toBeWithinRange(received, floor, ceiling) {
    const pass = received >= floor && received <= ceiling;
    if (pass) {
      return {
        message: () =>
          `expected ${received} not to be within range ${floor} - ${ceiling}`,
        pass: true,
      };
    } else {
      return {
        message: () =>
          `expected ${received} to be within range ${floor} - ${ceiling}`,
        pass: false,
      };
    }
  },
});
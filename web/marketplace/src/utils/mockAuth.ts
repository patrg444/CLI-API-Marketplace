// Mock authentication utilities for testing
export interface MockUser {
  id: string;
  username: string;
  email: string;
  name: string;
}

export const mockAuthUtils = {
  // Set mock user in localStorage
  setMockUser: (user: MockUser) => {
    localStorage.setItem('mockUser', JSON.stringify(user));
  },

  // Get current mock user
  getCurrentUser: async (): Promise<MockUser> => {
    const mockUser = localStorage.getItem('mockUser');
    if (mockUser) {
      return JSON.parse(mockUser);
    }
    throw new Error('No authenticated user');
  },

  // Get current mock user synchronously
  getCurrentUserSync: (): MockUser | null => {
    const mockUser = localStorage.getItem('mockUser');
    return mockUser ? JSON.parse(mockUser) : null;
  },

  // Clear mock user (logout)
  clearMockUser: () => {
    localStorage.removeItem('mockUser');
  },

  // Check if user is authenticated
  isAuthenticated: (): boolean => {
    return localStorage.getItem('mockUser') !== null;
  },

  // Get mock session
  getCurrentSession: async () => {
    const mockUser = localStorage.getItem('mockUser');
    if (mockUser) {
      return {
        getIdToken: () => ({
          getJwtToken: () => 'mock-jwt-token-' + Date.now()
        })
      };
    }
    throw new Error('No authenticated user');
  },

  // Initialize with a default test user (for development)
  initializeTestUser: () => {
    const testUser: MockUser = {
      id: 'test-user-123',
      username: 'testuser',
      email: 'test@example.com',
      name: 'Test User'
    };
    mockAuthUtils.setMockUser(testUser);
    return testUser;
  }
};

// Auto-initialize test user in development
if (typeof window !== 'undefined' && process.env.NODE_ENV === 'development') {
  if (!mockAuthUtils.isAuthenticated()) {
    mockAuthUtils.initializeTestUser();
  }
}
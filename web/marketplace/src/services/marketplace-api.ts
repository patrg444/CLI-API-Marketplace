// Real Marketplace API Service
// This connects to our Vercel backend API

const API_BASE_URL = process.env.NODE_ENV === 'production' 
  ? '/api'  // Use relative URL in production to work with Vercel Functions
  : 'http://localhost:3000/api';

export interface MarketplaceAPI {
  id: string;
  name: string;
  author: string;
  description: string;
  category: string;
  icon: string;
  color: string;
  rating: number;
  reviews: number;
  calls: number;
  pricing: {
    type: 'freemium' | 'subscription';
    freeCalls?: number;
    pricePerCall?: number;
    monthlyPrice?: number;
    currency: string;
  };
  tags: string[];
  featured: boolean;
  trending: boolean;
  growth?: number;
}

export interface Category {
  id: string;
  name: string;
  icon: string;
  count: number;
}

class MarketplaceAPIService {
  private baseURL: string;

  constructor() {
    this.baseURL = API_BASE_URL;
  }

  async getCategories(): Promise<Category[]> {
    try {
      const response = await fetch(`${this.baseURL}/categories`);
      const data = await response.json();
      return data.success ? data.data : [];
    } catch (error) {
      console.error('Error fetching categories:', error);
      return [];
    }
  }

  async getAPIs(params?: {
    category?: string;
    search?: string;
    maxPrice?: number;
    sort?: string;
    page?: number;
    limit?: number;
  }): Promise<{
    apis: MarketplaceAPI[];
    total: number;
    page: number;
    totalPages: number;
  }> {
    try {
      const queryParams = new URLSearchParams();
      if (params) {
        Object.entries(params).forEach(([key, value]) => {
          if (value !== undefined) {
            queryParams.append(key, value.toString());
          }
        });
      }

      const response = await fetch(`${this.baseURL}/apis?${queryParams}`);
      const data = await response.json();
      
      if (data.success) {
        return {
          apis: data.data,
          total: data.meta?.total || 0,
          page: data.meta?.page || 1,
          totalPages: data.meta?.totalPages || 1
        };
      }
      
      return { apis: [], total: 0, page: 1, totalPages: 1 };
    } catch (error) {
      console.error('Error fetching APIs:', error);
      return { apis: [], total: 0, page: 1, totalPages: 1 };
    }
  }

  async getFeaturedAPIs(): Promise<MarketplaceAPI[]> {
    try {
      const response = await fetch(`${this.baseURL}/apis/featured`);
      const data = await response.json();
      return data.success ? data.data : [];
    } catch (error) {
      console.error('Error fetching featured APIs:', error);
      return [];
    }
  }

  async getTrendingAPIs(): Promise<MarketplaceAPI[]> {
    try {
      const response = await fetch(`${this.baseURL}/apis/trending`);
      const data = await response.json();
      return data.success ? data.data : [];
    } catch (error) {
      console.error('Error fetching trending APIs:', error);
      return [];
    }
  }

  async getAPI(id: string): Promise<MarketplaceAPI | null> {
    try {
      const response = await fetch(`${this.baseURL}/apis/${id}`);
      const data = await response.json();
      return data.success ? data.data : null;
    } catch (error) {
      console.error('Error fetching API:', error);
      return null;
    }
  }
}

export const marketplaceAPI = new MarketplaceAPIService();
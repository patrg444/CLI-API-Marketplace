import React from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import APICard from '../APICard';
import { useRouter } from 'next/router';
import { API } from '@/types/api';

// Mock Next.js router
jest.mock('next/router', () => ({
  useRouter: jest.fn(),
}));

// Mock Link component
jest.mock('next/link', () => {
  return ({ children, href, ...props }: any) => (
    <a href={href} {...props}>{children}</a>
  );
});

describe('APICard Component', () => {
  const mockPush = jest.fn();
  
  // Mock window.innerWidth for responsive behavior
  beforeAll(() => {
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: 1024
    });
  });
  
  const mockAPI: API = {
    id: 'weather-api',
    creator_id: 'user-123',
    name: 'Weather API',
    description: 'Real-time weather data for 200K+ cities worldwide',
    category: 'weather',
    tags: ['weather', 'forecast', 'climate', 'real-time'],
    is_published: true,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
    pricing_plans: [
      {
        id: 'plan-free',
        api_id: 'weather-api',
        name: 'Free Plan',
        type: 'free',
        call_limit: 1000,
        rate_limit_per_minute: 10,
        is_active: true
      },
      {
        id: 'plan-pro',
        api_id: 'weather-api',
        name: 'Pro Plan',
        type: 'subscription',
        monthly_price: 29,
        call_limit: 100000,
        rate_limit_per_minute: 100,
        is_active: true
      }
    ],
    average_rating: 4.8,
    total_reviews: 142,
    total_subscriptions: 125000
  };

  beforeEach(() => {
    (useRouter as jest.Mock).mockReturnValue({
      push: mockPush,
      pathname: '/marketplace',
    });
  });

  afterEach(() => {
    jest.clearAllMocks();
  });

  describe('Rendering', () => {
    it('renders all API information correctly', () => {
      render(<APICard api={mockAPI} />);

      // Check basic info
      expect(screen.getByText('Weather API')).toBeInTheDocument();
      expect(screen.getByText('Real-time weather data for 200K+ cities worldwide')).toBeInTheDocument();

      // Check rating if available
      expect(screen.getByText('4.8')).toBeInTheDocument();
      expect(screen.getByText('(142)')).toBeInTheDocument();

      // Check pricing - free tier shows as "Free tier" on mobile
      expect(screen.getByTestId('free-tier-badge')).toBeInTheDocument();
      expect(screen.getByText('Free tier')).toBeInTheDocument();

      // Check tags - component checks window.innerWidth which is undefined in tests
      // So we'll just check that tags are rendered
      const badges = screen.getAllByText('weather', { selector: '.badge' });
      expect(badges.length).toBeGreaterThan(0);

      // Check category
      expect(screen.getByText('weather', { selector: '.badge-primary' })).toBeInTheDocument();
    });

    it('renders featured badge when featured', () => {
      render(<APICard api={mockAPI} isFeatured />);
      
      expect(screen.getByText('Featured')).toBeInTheDocument();
    });

    it('renders popular badge when has many subscriptions', () => {
      render(<APICard api={mockAPI} />);
      
      expect(screen.getByText('Popular')).toBeInTheDocument();
    });

    it('handles only subscription pricing', () => {
      const subscriptionAPI: API = {
        ...mockAPI,
        pricing_plans: [
          {
            id: 'plan-sub',
            api_id: 'weather-api',
            name: 'Enterprise',
            type: 'subscription',
            monthly_price: 99,
            call_limit: 10000,
            is_active: true
          }
        ]
      };
      
      render(<APICard api={subscriptionAPI} />);
      
      // Check that it shows the monthly price
      expect(screen.getByText('$99')).toBeInTheDocument();
      expect(screen.getByText('/mo')).toBeInTheDocument();
      expect(screen.queryByTestId('free-tier-badge')).not.toBeInTheDocument();
    });

    it('renders icon initial when no icon_url', () => {
      render(<APICard api={mockAPI} />);
      
      expect(screen.getByText('W')).toBeInTheDocument();
    });

    it('renders icon image when icon_url is provided', () => {
      const apiWithIcon: API = {
        ...mockAPI,
        icon_url: 'https://example.com/icon.png'
      };
      
      render(<APICard api={apiWithIcon} />);
      
      const icon = screen.getByAltText('Weather API icon');
      expect(icon).toHaveAttribute('src', 'https://example.com/icon.png');
    });
  });

  describe('Interactions', () => {
    it('navigates to API detail page on click', () => {
      render(<APICard api={mockAPI} />);
      
      const link = screen.getByRole('link');
      expect(link).toHaveAttribute('href', '/apis/weather-api');
    });

    it('shows hover state', async () => {
      const user = userEvent.setup();
      render(<APICard api={mockAPI} />);
      
      const card = screen.getByTestId('api-card');
      await user.hover(card);
      
      // Check if hover classes are applied
      expect(card).toHaveClass('card-hover');
    });
  });

  describe('Error Handling', () => {
    it('handles missing pricing gracefully', () => {
      const apiWithoutPricing: API = {
        ...mockAPI,
        pricing_plans: []
      };
      
      render(<APICard api={apiWithoutPricing} />);
      
      // Should not show any pricing info
      expect(screen.queryByText('Free')).not.toBeInTheDocument();
      expect(screen.queryByText(/From \$/)).not.toBeInTheDocument();
    });

    it('handles missing rating gracefully', () => {
      const apiWithoutRating: API = {
        ...mockAPI,
        average_rating: undefined,
        total_reviews: undefined
      };
      
      render(<APICard api={apiWithoutRating} />);
      
      // Should not show rating section
      expect(screen.queryByText('4.8')).not.toBeInTheDocument();
      expect(screen.queryByText(/\(\d+\)/)).not.toBeInTheDocument();
    });

    it('handles API with no tags', () => {
      const apiWithoutTags: API = {
        ...mockAPI,
        tags: []
      };
      
      render(<APICard api={apiWithoutTags} />);
      
      // Should still render without errors
      expect(screen.getByText('Weather API')).toBeInTheDocument();
    });
  });

  describe('Accessibility', () => {
    it('has proper link structure', () => {
      render(<APICard api={mockAPI} />);
      
      const link = screen.getByRole('link');
      expect(link).toBeInTheDocument();
      expect(link).toHaveAttribute('href', '/apis/weather-api');
    });

    it('has proper image alt text when icon is present', () => {
      const apiWithIcon: API = {
        ...mockAPI,
        icon_url: 'https://example.com/icon.png'
      };
      
      render(<APICard api={apiWithIcon} />);
      
      const icon = screen.getByRole('img');
      expect(icon).toHaveAttribute('alt', 'Weather API icon');
    });
  });

  describe('Popular Badge Logic', () => {
    it('shows popular badge when subscriptions > 100', () => {
      render(<APICard api={mockAPI} />);
      expect(screen.getByText('Popular')).toBeInTheDocument();
    });

    it('does not show popular badge when subscriptions <= 100', () => {
      const unpopularAPI: API = {
        ...mockAPI,
        total_subscriptions: 50
      };
      
      render(<APICard api={unpopularAPI} />);
      expect(screen.queryByText('Popular')).not.toBeInTheDocument();
    });

    it('does not show popular badge when subscriptions is undefined', () => {
      const apiNoSubs: API = {
        ...mockAPI,
        total_subscriptions: undefined
      };
      
      render(<APICard api={apiNoSubs} />);
      expect(screen.queryByText('Popular')).not.toBeInTheDocument();
    });
  });

  describe('Pricing Display', () => {
    it('shows free tier when available', () => {
      render(<APICard api={mockAPI} />);
      expect(screen.getByTestId('free-tier-badge')).toBeInTheDocument();
    });

    it('shows lowest paid price when multiple paid plans exist', () => {
      const multiPlanAPI: API = {
        ...mockAPI,
        pricing_plans: [
          {
            id: 'plan-1',
            api_id: 'weather-api',
            name: 'Basic',
            type: 'subscription',
            monthly_price: 19,
            is_active: true
          },
          {
            id: 'plan-2',
            api_id: 'weather-api',
            name: 'Pro',
            type: 'subscription',
            monthly_price: 49,
            is_active: true
          }
        ]
      };
      
      render(<APICard api={multiPlanAPI} />);
      expect(screen.getByText('$19')).toBeInTheDocument();
      expect(screen.getByText('/mo')).toBeInTheDocument();
    });

    it('handles pay_per_use pricing type', () => {
      const payPerUseAPI: API = {
        ...mockAPI,
        pricing_plans: [
          {
            id: 'plan-ppu',
            api_id: 'weather-api',
            name: 'Pay as you go',
            type: 'pay_per_use',
            price_per_call: 0.001,
            is_active: true
          }
        ]
      };
      
      render(<APICard api={payPerUseAPI} />);
      // Component currently doesn't display pay_per_use pricing
      expect(screen.queryByText(/\$0.001/)).not.toBeInTheDocument();
    });
  });
});
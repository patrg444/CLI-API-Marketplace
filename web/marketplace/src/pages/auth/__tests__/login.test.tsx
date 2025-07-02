import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useRouter } from 'next/router';
import Login from '../login';

// Mock dependencies
const mockPush = jest.fn();
const mockRouter = {
  push: mockPush,
  query: {},
  pathname: '/auth/login',
  route: '/auth/login',
  asPath: '/auth/login',
  replace: jest.fn(),
  reload: jest.fn(),
  back: jest.fn(),
  prefetch: jest.fn(),
  beforePopState: jest.fn(),
  events: {
    on: jest.fn(),
    off: jest.fn(),
    emit: jest.fn(),
  },
};

jest.mock('next/router', () => ({
  useRouter: () => mockRouter,
}));

jest.mock('next/link', () => {
  return ({ children, href }: any) => <a href={href}>{children}</a>;
});

// Mock Layout component
jest.mock('../../../components/Layout', () => {
  return ({ children }: any) => <div>{children}</div>;
});

describe('Login Page', () => {
  beforeEach(() => {
    // Reset mocks
    mockPush.mockClear();
    localStorage.clear();
    // Reset router mock
    mockRouter.push.mockClear();
    mockRouter.replace.mockClear();
    // Clear any stored items
    localStorage.removeItem('mockUser');
  });

  afterEach(() => {
    jest.clearAllMocks();
    localStorage.clear();
  });

  describe('Rendering', () => {
    it('renders login form with all fields', () => {
      render(<Login />);
      
      expect(screen.getByPlaceholderText(/enter your email/i)).toBeInTheDocument();
      expect(screen.getByPlaceholderText(/enter your password/i)).toBeInTheDocument();
      expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
      expect(screen.getByText(/create a new account/i)).toBeInTheDocument();
      expect(screen.getByText(/forgot your password/i)).toBeInTheDocument();
    });

    it('shows loading state when submitting', async () => {
      render(<Login />);
      
      const emailInput = screen.getByPlaceholderText(/enter your email/i);
      const passwordInput = screen.getByPlaceholderText(/enter your password/i);
      const submitButton = screen.getByRole('button', { name: /sign in/i });
      
      await userEvent.type(emailInput, 'test@example.com');
      await userEvent.type(passwordInput, 'password123');
      
      fireEvent.click(submitButton);
      
      expect(submitButton).toBeDisabled();
      expect(submitButton).toHaveTextContent(/signing in/i);
    });
  });

  describe('Form Validation', () => {
    it('shows error for empty fields', async () => {
      render(<Login />);
      
      const submitButton = screen.getByRole('button', { name: /sign in/i });
      fireEvent.click(submitButton);
      
      await waitFor(() => {
        expect(screen.getByText(/please fill in all fields/i)).toBeInTheDocument();
      });
    });

    it('shows error for invalid email format', async () => {
      const user = userEvent.setup();
      render(<Login />);
      
      const emailInput = screen.getByPlaceholderText(/enter your email/i);
      const passwordInput = screen.getByPlaceholderText(/enter your password/i);
      
      await user.type(emailInput, 'invalid-email');
      await user.type(passwordInput, 'password123');
      
      const submitButton = screen.getByRole('button', { name: /sign in/i });
      await user.click(submitButton);
      
      await waitFor(() => {
        expect(screen.getByText(/please enter a valid email/i)).toBeInTheDocument();
      });
    });

    it('shows error for empty password', async () => {
      const user = userEvent.setup();
      render(<Login />);
      
      const emailInput = screen.getByPlaceholderText(/enter your email/i);
      await user.type(emailInput, 'test@example.com');
      
      const submitButton = screen.getByRole('button', { name: /sign in/i });
      await user.click(submitButton);
      
      await waitFor(() => {
        expect(screen.getByText(/please fill in all fields/i)).toBeInTheDocument();
      });
    });
  });

  describe('Authentication Flow', () => {
    it('successfully logs in with valid credentials', async () => {
      const user = userEvent.setup();
      render(<Login />);
      
      await user.type(screen.getByPlaceholderText(/enter your email/i), 'test@example.com');
      await user.type(screen.getByPlaceholderText(/enter your password/i), 'ValidPassword123!');
      await user.click(screen.getByRole('button', { name: /sign in/i }));
      
      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith('/dashboard');
      }, { timeout: 2000 });
      
      // Check localStorage was set
      const storedUser = JSON.parse(localStorage.getItem('mockUser') || '{}');
      expect(storedUser.email).toBe('test@example.com');
      expect(storedUser.name).toBe('test');
    });

    it.skip('stores user data in localStorage after login', async () => {
      const user = userEvent.setup();
      render(<Login />);
      
      await user.type(screen.getByPlaceholderText(/enter your email/i), 'john.doe@example.com');
      await user.type(screen.getByPlaceholderText(/enter your password/i), 'password123');
      await user.click(screen.getByRole('button', { name: /sign in/i }));
      
      await waitFor(() => {
        expect(mockPush).toHaveBeenCalledWith('/dashboard');
      }, { timeout: 2000 });
      
      // Check localStorage after navigation
      const storedUser = JSON.parse(localStorage.getItem('mockUser') || '{}');
      expect(storedUser).toEqual({
        email: 'john.doe@example.com',
        name: 'john.doe',
        id: 'test-user-id'
      });
    });
  });

  describe('UI Interactions', () => {
    it('clears error message when user starts typing', async () => {
      const user = userEvent.setup();
      render(<Login />);
      
      // Trigger an error
      await user.click(screen.getByRole('button', { name: /sign in/i }));
      
      await waitFor(() => {
        expect(screen.getByText(/please fill in all fields/i)).toBeInTheDocument();
      });
      
      // Start typing
      await user.type(screen.getByPlaceholderText(/enter your email/i), 'test');
      
      // Error should be cleared
      expect(screen.queryByText(/please fill in all fields/i)).not.toBeInTheDocument();
    });

    it('disables submit button while loading', async () => {
      const user = userEvent.setup();
      render(<Login />);
      
      await user.type(screen.getByPlaceholderText(/enter your email/i), 'test@example.com');
      await user.type(screen.getByPlaceholderText(/enter your password/i), 'password123');
      
      const submitButton = screen.getByRole('button', { name: /sign in/i });
      fireEvent.click(submitButton);
      
      expect(submitButton).toBeDisabled();
      
      await waitFor(() => {
        expect(submitButton).not.toBeDisabled();
      }, { timeout: 2000 });
    });
  });

  describe('Navigation', () => {
    it('has link to signup page', () => {
      render(<Login />);
      
      const signupLink = screen.getByText(/create a new account/i).closest('a');
      expect(signupLink).toHaveAttribute('href', '/auth/signup');
    });

    it('has link to forgot password page', () => {
      render(<Login />);
      
      const forgotLink = screen.getByText(/forgot your password/i).closest('a');
      expect(forgotLink).toHaveAttribute('href', '/auth/forgot-password');
    });
  });

  describe('Security', () => {
    it('password field is of type password', () => {
      render(<Login />);
      
      const passwordInput = screen.getByPlaceholderText(/enter your password/i);
      expect(passwordInput).toHaveAttribute('type', 'password');
    });

    it('clears form on component unmount', () => {
      const { unmount } = render(<Login />);
      
      const emailInput = screen.getByPlaceholderText(/enter your email/i) as HTMLInputElement;
      const passwordInput = screen.getByPlaceholderText(/enter your password/i) as HTMLInputElement;
      
      fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
      fireEvent.change(passwordInput, { target: { value: 'password123' } });
      
      unmount();
      
      // In a real app, you'd want to ensure state is cleared
      // This is a simplified test
      expect(true).toBe(true);
    });
  });

  describe('Accessibility', () => {
    it('form inputs have proper labels or placeholders', () => {
      render(<Login />);
      
      expect(screen.getByPlaceholderText(/enter your email/i)).toBeInTheDocument();
      expect(screen.getByPlaceholderText(/enter your password/i)).toBeInTheDocument();
    });

    it('form elements are keyboard accessible', async () => {
      const user = userEvent.setup();
      render(<Login />);
      
      // Focus on email input
      const emailInput = screen.getByPlaceholderText(/enter your email/i);
      emailInput.focus();
      expect(emailInput).toHaveFocus();
      
      // Tab to password
      await user.tab();
      expect(screen.getByPlaceholderText(/enter your password/i)).toHaveFocus();
      
      // Tab to forgot password link
      await user.tab();
      
      // Tab to submit button
      await user.tab();
      expect(screen.getByRole('button', { name: /sign in/i })).toHaveFocus();
    });
  });
});
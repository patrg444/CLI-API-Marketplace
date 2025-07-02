import React, { ReactNode, useState, useEffect } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/router'
import { Bars3Icon, XMarkIcon } from '@heroicons/react/24/outline'

interface LayoutProps {
  children: ReactNode
}

export default function Layout({ children }: LayoutProps) {
  const router = useRouter()
  const [user, setUser] = useState<any>(null)
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)

  useEffect(() => {
    // Simple mock authentication for testing
    const storedUser = localStorage.getItem('mockUser')
    if (storedUser) {
      setUser(JSON.parse(storedUser))
    }
  }, [])

  const handleSignOut = () => {
    localStorage.removeItem('mockUser')
    setUser(null)
    router.push('/')
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <Link href="/" className="flex items-center">
                <span className="text-lg sm:text-xl font-bold text-indigo-600">API Direct</span>
                <span className="ml-1 sm:ml-2 text-xs sm:text-sm text-gray-500">Marketplace</span>
              </Link>
              
              {/* Desktop Navigation */}
              <div className="hidden md:ml-8 md:flex md:space-x-8">
                <Link
                  href="/"
                  className={`inline-flex items-center px-1 pt-1 text-sm font-medium ${
                    router.pathname === '/' 
                      ? 'text-gray-900 border-b-2 border-indigo-500'
                      : 'text-gray-500 hover:text-gray-700'
                  }`}
                  data-testid="browse-apis"
                >
                  Browse APIs
                </Link>
                {user && (
                  <>
                    <Link
                      href="/dashboard"
                      className={`inline-flex items-center px-1 pt-1 text-sm font-medium ${
                        router.pathname === '/dashboard'
                          ? 'text-gray-900 border-b-2 border-indigo-500'
                          : 'text-gray-500 hover:text-gray-700'
                      }`}
                    >
                      My Dashboard
                    </Link>
                    <Link
                      href="/account"
                      className={`inline-flex items-center px-1 pt-1 text-sm font-medium ${
                        router.pathname.startsWith('/account')
                          ? 'text-gray-900 border-b-2 border-indigo-500'
                          : 'text-gray-500 hover:text-gray-700'
                      }`}
                    >
                      My Account
                    </Link>
                  </>
                )}
              </div>
            </div>

            {/* Desktop User Menu */}
            <div className="hidden md:flex md:items-center md:space-x-4">
              {user ? (
                <>
                  <span className="text-sm text-gray-700 truncate max-w-32">{user?.email}</span>
                  <button
                    onClick={handleSignOut}
                    className="text-sm text-gray-500 hover:text-gray-700 px-3 py-2 rounded-md transition-colors"
                    data-testid="signout-button"
                  >
                    Sign Out
                  </button>
                </>
              ) : (
                <div className="flex items-center space-x-4">
                  <Link
                    href="/auth/login"
                    className="text-sm text-gray-700 hover:text-gray-900 px-3 py-2 rounded-md transition-colors"
                    data-testid="login-button"
                  >
                    Sign In
                  </Link>
                  <Link
                    href="/auth/signup"
                    className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 transition-colors touch-manipulation"
                    data-testid="signup-button"
                  >
                    Sign Up
                  </Link>
                </div>
              )}
            </div>

            {/* Mobile menu button */}
            <div className="md:hidden flex items-center">
              <button
                type="button"
                className="inline-flex items-center justify-center p-2 rounded-md text-gray-400 hover:text-gray-500 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-inset focus:ring-indigo-500 touch-manipulation"
                aria-controls="mobile-menu"
                aria-expanded={mobileMenuOpen}
                onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
                data-testid="mobile-menu-button"
              >
                <span className="sr-only">Open main menu</span>
                {mobileMenuOpen ? (
                  <XMarkIcon className="block h-6 w-6" aria-hidden="true" />
                ) : (
                  <Bars3Icon className="block h-6 w-6" aria-hidden="true" />
                )}
              </button>
            </div>
          </div>
        </div>

        {/* Mobile menu */}
        <div className={`md:hidden ${mobileMenuOpen ? 'block' : 'hidden'}`} id="mobile-menu">
          <div className="px-2 pt-2 pb-3 space-y-1 sm:px-3 bg-white border-t border-gray-200">
            <Link
              href="/"
              className={`block px-3 py-2 rounded-md text-base font-medium transition-colors touch-manipulation ${
                router.pathname === '/'
                  ? 'text-indigo-700 bg-indigo-50'
                  : 'text-gray-700 hover:text-gray-900 hover:bg-gray-50'
              }`}
              onClick={() => setMobileMenuOpen(false)}
              data-testid="mobile-browse-apis"
            >
              Browse APIs
            </Link>
            {user && (
              <Link
                href="/dashboard"
                className={`block px-3 py-2 rounded-md text-base font-medium transition-colors touch-manipulation ${
                  router.pathname === '/dashboard'
                    ? 'text-indigo-700 bg-indigo-50'
                    : 'text-gray-700 hover:text-gray-900 hover:bg-gray-50'
                }`}
                onClick={() => setMobileMenuOpen(false)}
                data-testid="mobile-dashboard"
              >
                My Dashboard
              </Link>
            )}
          </div>
          
          {/* Mobile User Section */}
          <div className="pt-4 pb-3 border-t border-gray-200">
            {user ? (
              <div className="space-y-1">
                <div className="px-4 py-2">
                  <div className="text-base font-medium text-gray-800 truncate">{user?.email}</div>
                </div>
                <button
                  onClick={() => {
                    handleSignOut()
                    setMobileMenuOpen(false)
                  }}
                  className="block px-4 py-2 text-base font-medium text-gray-500 hover:text-gray-800 hover:bg-gray-100 w-full text-left touch-manipulation"
                  data-testid="mobile-signout-button"
                >
                  Sign Out
                </button>
              </div>
            ) : (
              <div className="space-y-1 px-4">
                <Link
                  href="/auth/login"
                  className="block py-2 text-base font-medium text-gray-700 hover:text-gray-900 touch-manipulation"
                  onClick={() => setMobileMenuOpen(false)}
                  data-testid="mobile-login-button"
                >
                  Sign In
                </Link>
                <Link
                  href="/auth/signup"
                  className="block w-full text-center py-3 px-4 mt-2 border border-transparent text-base font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700 transition-colors touch-manipulation"
                  onClick={() => setMobileMenuOpen(false)}
                  data-testid="mobile-signup-button"
                >
                  Sign Up
                </Link>
              </div>
            )}
          </div>
        </div>
      </nav>

      <main>{children}</main>

      <footer className="bg-white mt-auto">
        <div className="max-w-7xl mx-auto py-8 sm:py-12 px-4 sm:px-6 lg:px-8">
          <div className="text-center text-xs sm:text-sm text-gray-500">
            © 2025 API Direct. All rights reserved.
          </div>
        </div>
      </footer>
    </div>
  )
}

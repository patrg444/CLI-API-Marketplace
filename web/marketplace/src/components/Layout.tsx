import React, { ReactNode } from 'react'
import Link from 'next/link'
import { useRouter } from 'next/router'
import { getCurrentUser, signOut, fetchUserAttributes } from 'aws-amplify/auth'
import { useQuery } from 'react-query'

interface LayoutProps {
  children: ReactNode
}

export default function Layout({ children }: LayoutProps) {
  const router = useRouter()
  const { data: user } = useQuery('currentUser', async () => {
    try {
      const currentUser = await getCurrentUser()
      const attributes = await fetchUserAttributes()
      return { ...currentUser, attributes }
    } catch {
      return null
    }
  })

  const handleSignOut = async () => {
    try {
      await signOut()
      router.push('/')
    } catch (error) {
      console.error('Error signing out:', error)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <nav className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex">
              <Link href="/" className="flex items-center">
                <span className="text-xl font-bold text-indigo-600">API Direct</span>
                <span className="ml-2 text-sm text-gray-500">Marketplace</span>
              </Link>
              
              <div className="hidden sm:ml-8 sm:flex sm:space-x-8">
                <Link
                  href="/"
                  className={`inline-flex items-center px-1 pt-1 text-sm font-medium ${
                    router.pathname === '/' 
                      ? 'text-gray-900 border-b-2 border-indigo-500'
                      : 'text-gray-500 hover:text-gray-700'
                  }`}
                >
                  Browse APIs
                </Link>
                {user && (
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
                )}
              </div>
            </div>

            <div className="flex items-center space-x-4">
              {user ? (
                <>
                  <span className="text-sm text-gray-700">{user.attributes?.email}</span>
                  <button
                    onClick={handleSignOut}
                    className="text-sm text-gray-500 hover:text-gray-700"
                  >
                    Sign Out
                  </button>
                </>
              ) : (
                <div className="space-x-4">
                  <Link
                    href="/auth/login"
                    className="text-sm text-gray-700 hover:text-gray-900"
                  >
                    Sign In
                  </Link>
                  <Link
                    href="/auth/signup"
                    className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md text-white bg-indigo-600 hover:bg-indigo-700"
                  >
                    Sign Up
                  </Link>
                </div>
              )}
            </div>
          </div>
        </div>
      </nav>

      <main>{children}</main>

      <footer className="bg-white mt-auto">
        <div className="max-w-7xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
          <div className="text-center text-sm text-gray-500">
            © 2025 API Direct. All rights reserved.
          </div>
        </div>
      </footer>
    </div>
  )
}

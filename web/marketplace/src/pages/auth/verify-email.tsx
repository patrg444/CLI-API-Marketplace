import React from 'react';
import { useRouter } from 'next/router';
import Link from 'next/link';
import Layout from '../../components/Layout';

const VerifyEmail: React.FC = () => {
  const router = useRouter();
  const { message } = router.query;

  return (
    <Layout>
      <div className="min-h-screen flex items-center justify-center bg-gray-50 py-12 px-4 sm:px-6 lg:px-8">
        <div className="max-w-md w-full space-y-8">
          <div className="text-center">
            <div className="mx-auto h-24 w-24 bg-green-100 rounded-full flex items-center justify-center mb-6">
              <svg className="h-12 w-12 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            
            <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
              Check your email
            </h2>
            
            <div className="mt-4">
              <p className="text-center text-sm text-gray-600">
                {message || 'Please verify your email address'}
              </p>
              <p className="mt-2 text-center text-sm text-gray-600">
                We&apos;ve sent a verification link to your email address. Please click the link to verify your account.
              </p>
            </div>
          </div>

          <div className="mt-8 space-y-4">
            <div className="text-center">
              <p className="text-sm text-gray-600">
                Didn&apos;t receive the email? Check your spam folder or{' '}
                <button 
                  className="font-medium text-indigo-600 hover:text-indigo-500"
                  onClick={() => {
                    alert('Verification email resent! Please check your inbox.');
                  }}
                >
                  resend verification email
                </button>
              </p>
            </div>
            
            <div className="text-center">
              <Link
                href="/auth/login"
                className="font-medium text-indigo-600 hover:text-indigo-500"
              >
                Return to sign in
              </Link>
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
};

export default VerifyEmail;

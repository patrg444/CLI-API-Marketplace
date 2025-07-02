import React from 'react'

export default function TestPage() {
  return (
    <div className="min-h-screen bg-gray-100 p-8">
      <div className="max-w-4xl mx-auto">
        <h1 className="text-4xl font-bold text-blue-600 mb-4">Test Page</h1>
        <div className="test-gradient">
          This should have a gradient background if CSS is loading
        </div>
        <div className="bg-white p-6 rounded-lg shadow-lg">
          <h2 className="text-2xl font-semibold text-gray-800 mb-4">Tailwind Test</h2>
          <p className="text-gray-600 mb-4">This text should be gray if Tailwind is working.</p>
          <button className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded">
            Test Button
          </button>
        </div>
      </div>
    </div>
  )
}
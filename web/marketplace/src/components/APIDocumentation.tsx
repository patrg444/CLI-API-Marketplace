import React, { useEffect, useState } from 'react'
import SwaggerUI from 'swagger-ui-react'
import 'swagger-ui-react/swagger-ui.css'
import { APIDocumentation as APIDocType, APIKey } from '@/types/api'

interface APIDocumentationProps {
  documentation: APIDocType | null
  apiKey?: APIKey | null
  apiBaseUrl?: string
  isSubscribed: boolean
}

const APIDocumentation: React.FC<APIDocumentationProps> = ({
  documentation,
  apiKey,
  apiBaseUrl,
  isSubscribed
}) => {
  const [spec, setSpec] = useState<any>(null)

  useEffect(() => {
    if (documentation?.openapi_spec) {
      // Parse the spec if it's a string
      if (typeof documentation.openapi_spec === 'string') {
        try {
          setSpec(JSON.parse(documentation.openapi_spec))
        } catch (e) {
          // If parsing fails, assume it's already an object
          setSpec(documentation.openapi_spec)
        }
      } else {
        setSpec(documentation.openapi_spec)
      }
    }
  }, [documentation])

  // Add custom styles
  useEffect(() => {
    const styleId = 'swagger-ui-custom-styles'
    if (!document.getElementById(styleId)) {
      const style = document.createElement('style')
      style.id = styleId
      style.innerHTML = swaggerCustomStyles
      document.head.appendChild(style)
    }

    return () => {
      const existingStyle = document.getElementById(styleId)
      if (existingStyle) {
        existingStyle.remove()
      }
    }
  }, [])

  if (!documentation || !documentation.has_openapi) {
    return (
      <div className="bg-gray-50 rounded-lg p-8 text-center">
        <svg
          className="mx-auto h-12 w-12 text-gray-400"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
          />
        </svg>
        <h3 className="mt-2 text-sm font-medium text-gray-900">No API documentation available</h3>
        <p className="mt-1 text-sm text-gray-500">
          The API creator hasn't uploaded OpenAPI documentation yet.
        </p>
        {documentation?.markdown_content && (
          <div className="mt-6 text-left prose prose-sm max-w-none">
            <div dangerouslySetInnerHTML={{ __html: documentation.markdown_content }} />
          </div>
        )}
      </div>
    )
  }

  if (!spec) {
    return (
      <div className="flex justify-center items-center p-8">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600"></div>
      </div>
    )
  }

  // Custom plugin to inject API key
  const requestInterceptor = (request: any) => {
    if (apiKey && isSubscribed) {
      request.headers['X-API-Key'] = apiKey.key_prefix
    }
    
    // Update the URL if we have a custom base URL
    if (apiBaseUrl && request.url) {
      try {
        const url = new URL(request.url)
        const baseUrl = new URL(apiBaseUrl)
        url.host = baseUrl.host
        url.protocol = baseUrl.protocol
        request.url = url.toString()
      } catch (e) {
        console.error('Error updating request URL:', e)
      }
    }
    
    return request
  }

  // Custom layout plugin to control which sections are shown
  const CustomLayoutPlugin = () => {
    return {
      wrapComponents: {
        InfoContainer: () => () => null, // Hide info section
      }
    }
  }

  return (
    <div className="api-documentation">
      {!isSubscribed && (
        <div className="mb-4 p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-yellow-800">
                Limited Access
              </h3>
              <div className="mt-2 text-sm text-yellow-700">
                <p>
                  Subscribe to this API to enable the "Try it out" functionality and test endpoints directly from this documentation.
                </p>
              </div>
            </div>
          </div>
        </div>
      )}

      <div className="swagger-ui-wrapper">
        <SwaggerUI
          spec={spec}
          docExpansion="list"
          defaultModelsExpandDepth={1}
          showExtensions={false}
          showCommonExtensions={false}
          tryItOutEnabled={isSubscribed}
          requestInterceptor={requestInterceptor}
          plugins={[CustomLayoutPlugin]}
        />
      </div>
    </div>
  )
}

const swaggerCustomStyles = `
  .api-documentation .swagger-ui {
    font-family: inherit;
  }

  .api-documentation .swagger-ui .topbar {
    display: none;
  }

  .api-documentation .swagger-ui .info {
    margin-bottom: 2rem;
  }

  .api-documentation .swagger-ui .info .title {
    font-size: 1.875rem;
    font-weight: 700;
    color: #111827;
  }

  .api-documentation .swagger-ui .info .description {
    font-size: 0.875rem;
    color: #6b7280;
    margin-top: 0.5rem;
  }

  .api-documentation .swagger-ui .scheme-container {
    background: #f9fafb;
    padding: 1rem;
    border-radius: 0.5rem;
    margin-bottom: 1.5rem;
  }

  .api-documentation .swagger-ui .btn {
    background: #4f46e5;
    color: white;
    border: none;
    padding: 0.5rem 1rem;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .api-documentation .swagger-ui .btn:hover {
    background: #4338ca;
  }

  .api-documentation .swagger-ui .btn.cancel {
    background: #f3f4f6;
    color: #374151;
  }

  .api-documentation .swagger-ui .btn.cancel:hover {
    background: #e5e7eb;
  }

  .api-documentation .swagger-ui .opblock {
    margin-bottom: 1rem;
    border: 1px solid #e5e7eb;
    border-radius: 0.5rem;
    overflow: hidden;
  }

  .api-documentation .swagger-ui .opblock-summary {
    border: none;
    padding: 0.75rem 1rem;
  }

  .api-documentation .swagger-ui .opblock.opblock-get .opblock-summary {
    background: #eff6ff;
  }

  .api-documentation .swagger-ui .opblock.opblock-post .opblock-summary {
    background: #f0fdf4;
  }

  .api-documentation .swagger-ui .opblock.opblock-put .opblock-summary {
    background: #fef3c7;
  }

  .api-documentation .swagger-ui .opblock.opblock-delete .opblock-summary {
    background: #fee2e2;
  }

  .api-documentation .swagger-ui .opblock-summary-method {
    font-weight: 600;
    padding: 0.25rem 0.5rem;
    border-radius: 0.25rem;
    font-size: 0.75rem;
    min-width: 4rem;
    text-align: center;
  }

  .api-documentation .swagger-ui .opblock.opblock-get .opblock-summary-method {
    background: #3b82f6;
    color: white;
  }

  .api-documentation .swagger-ui .opblock.opblock-post .opblock-summary-method {
    background: #10b981;
    color: white;
  }

  .api-documentation .swagger-ui .opblock.opblock-put .opblock-summary-method {
    background: #f59e0b;
    color: white;
  }

  .api-documentation .swagger-ui .opblock.opblock-delete .opblock-summary-method {
    background: #ef4444;
    color: white;
  }

  .api-documentation .swagger-ui .opblock-body {
    background: white;
  }

  .api-documentation .swagger-ui .parameter__name,
  .api-documentation .swagger-ui .parameter__type {
    font-size: 0.875rem;
  }

  .api-documentation .swagger-ui .parameter__name.required::after {
    content: " *";
    color: #ef4444;
  }

  .api-documentation .swagger-ui table tbody tr td {
    padding: 0.75rem;
  }

  .api-documentation .swagger-ui .response-col_status {
    font-weight: 600;
  }

  .api-documentation .swagger-ui .responses-table {
    margin-top: 1rem;
  }

  .api-documentation .swagger-ui .model-box {
    background: #f9fafb;
    border: 1px solid #e5e7eb;
    border-radius: 0.375rem;
    padding: 1rem;
    margin: 0.5rem 0;
  }

  .api-documentation .swagger-ui select,
  .api-documentation .swagger-ui input[type=text],
  .api-documentation .swagger-ui input[type=password],
  .api-documentation .swagger-ui input[type=email],
  .api-documentation .swagger-ui input[type=file],
  .api-documentation .swagger-ui textarea {
    border: 1px solid #d1d5db;
    border-radius: 0.375rem;
    padding: 0.5rem 0.75rem;
    font-size: 0.875rem;
  }

  .api-documentation .swagger-ui select:focus,
  .api-documentation .swagger-ui input:focus,
  .api-documentation .swagger-ui textarea:focus {
    outline: none;
    border-color: #4f46e5;
    box-shadow: 0 0 0 3px rgba(79, 70, 229, 0.1);
  }

  .api-documentation .swagger-ui .execute-wrapper {
    padding: 1rem;
    border-top: 1px solid #e5e7eb;
  }

  .api-documentation .swagger-ui .btn.execute {
    background: #4f46e5;
  }

  .api-documentation .swagger-ui .btn.execute:hover {
    background: #4338ca;
  }

  .api-documentation .swagger-ui .responses-wrapper {
    padding: 1rem;
  }

  .api-documentation .swagger-ui .response {
    margin-top: 1rem;
  }

  .api-documentation .swagger-ui pre.microlight {
    background: #1f2937;
    color: #f3f4f6;
    padding: 1rem;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    overflow-x: auto;
  }

  .api-documentation .swagger-ui .loading-container {
    padding: 2rem;
  }

  /* Hide try it out button when not subscribed */
  .api-documentation .swagger-ui .try-out__btn:disabled {
    display: none;
  }

  /* Additional responsive design */
  @media (max-width: 768px) {
    .api-documentation .swagger-ui .opblock-summary-path {
      font-size: 0.75rem;
    }
    
    .api-documentation .swagger-ui .opblock-summary-description {
      font-size: 0.75rem;
    }
  }
`

export default APIDocumentation

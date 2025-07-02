import React, { useState, useEffect } from 'react';
import { ChevronDownIcon, PlayIcon, CodeBracketIcon } from '@heroicons/react/24/outline';
import { ClipboardIcon } from '@heroicons/react/24/outline';

interface APIEndpoint {
  path: string;
  method: string;
  description?: string;
  parameters?: Parameter[];
  requestBody?: RequestBodySchema;
  responses?: { [key: string]: ResponseSchema };
}

interface Parameter {
  name: string;
  in: 'query' | 'path' | 'header';
  required: boolean;
  type: string;
  description?: string;
  example?: any;
}

interface RequestBodySchema {
  type: string;
  properties?: { [key: string]: any };
  example?: any;
}

interface ResponseSchema {
  description: string;
  example?: any;
}

interface APIPlaygroundProps {
  apiId: string;
  apiUrl: string;
  endpoints: APIEndpoint[];
  apiKey?: string;
  onSubscribe?: () => void;
}

const APIPlayground: React.FC<APIPlaygroundProps> = ({
  apiId,
  apiUrl,
  endpoints,
  apiKey,
  onSubscribe
}) => {
  const [selectedEndpoint, setSelectedEndpoint] = useState<APIEndpoint | null>(
    endpoints.length > 0 ? endpoints[0] : null
  );
  const [requestData, setRequestData] = useState<{
    headers: { [key: string]: string };
    queryParams: { [key: string]: string };
    pathParams: { [key: string]: string };
    body: string;
  }>({
    headers: {},
    queryParams: {},
    pathParams: {},
    body: ''
  });
  const [response, setResponse] = useState<{
    status?: number;
    headers?: { [key: string]: string };
    body?: string;
    error?: string;
    loading?: boolean;
  }>({});
  const [showCodeGen, setShowCodeGen] = useState(false);
  const [selectedLanguage, setSelectedLanguage] = useState('curl');

  useEffect(() => {
    if (selectedEndpoint) {
      // Initialize request data with defaults
      const newRequestData = {
        headers: apiKey ? { 'X-API-Key': apiKey } : {} as Record<string, string>,
        queryParams: {} as Record<string, string>,
        pathParams: {} as Record<string, string>,
        body: selectedEndpoint.requestBody?.example ? 
          JSON.stringify(selectedEndpoint.requestBody.example, null, 2) : ''
      };

      // Set default values from endpoint parameters
      selectedEndpoint.parameters?.forEach(param => {
        if (param.example !== undefined) {
          if (param.in === 'query') {
            newRequestData.queryParams[param.name] = String(param.example);
          } else if (param.in === 'path') {
            newRequestData.pathParams[param.name] = String(param.example);
          } else if (param.in === 'header') {
            newRequestData.headers[param.name] = String(param.example);
          }
        }
      });

      setRequestData(newRequestData);
    }
  }, [selectedEndpoint, apiKey]);

  const executeRequest = async () => {
    if (!selectedEndpoint) return;

    setResponse({ loading: true });

    try {
      // Build the URL with path and query parameters
      let url = apiUrl + selectedEndpoint.path;
      
      // Replace path parameters
      Object.entries(requestData.pathParams).forEach(([key, value]) => {
        url = url.replace(`{${key}}`, encodeURIComponent(value));
      });

      // Add query parameters
      const queryString = new URLSearchParams(requestData.queryParams).toString();
      if (queryString) {
        url += '?' + queryString;
      }

      // Prepare request options
      const options: RequestInit = {
        method: selectedEndpoint.method.toUpperCase(),
        headers: {
          'Content-Type': 'application/json',
          ...requestData.headers
        }
      };

      // Add body for POST/PUT/PATCH requests
      if (['POST', 'PUT', 'PATCH'].includes(selectedEndpoint.method.toUpperCase()) && requestData.body) {
        options.body = requestData.body;
      }

      // Make the request through our proxy to handle CORS
      const proxyUrl = `/api/playground/proxy`;
      const proxyResponse = await fetch(proxyUrl, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify({
          url,
          options
        })
      });

      const result = await proxyResponse.json();

      setResponse({
        status: result.status,
        headers: result.headers,
        body: JSON.stringify(result.data, null, 2),
        loading: false
      });
    } catch (error) {
      setResponse({
        error: error instanceof Error ? error.message : 'Request failed',
        loading: false
      });
    }
  };

  const generateCode = (language: string) => {
    if (!selectedEndpoint) return '';

    let url = apiUrl + selectedEndpoint.path;
    Object.entries(requestData.pathParams).forEach(([key, value]) => {
      url = url.replace(`{${key}}`, value);
    });

    const queryString = new URLSearchParams(requestData.queryParams).toString();
    if (queryString) {
      url += '?' + queryString;
    }

    switch (language) {
      case 'curl':
        let curlCmd = `curl -X ${selectedEndpoint.method.toUpperCase()} "${url}"`;
        Object.entries(requestData.headers).forEach(([key, value]) => {
          curlCmd += ` \\\n  -H "${key}: ${value}"`;
        });
        if (requestData.body && ['POST', 'PUT', 'PATCH'].includes(selectedEndpoint.method.toUpperCase())) {
          curlCmd += ` \\\n  -d '${requestData.body}'`;
        }
        return curlCmd;

      case 'javascript':
        return `const response = await fetch('${url}', {
  method: '${selectedEndpoint.method.toUpperCase()}',
  headers: ${JSON.stringify(requestData.headers, null, 2)},${
    requestData.body && ['POST', 'PUT', 'PATCH'].includes(selectedEndpoint.method.toUpperCase()) 
      ? `\n  body: ${JSON.stringify(requestData.body)}`
      : ''
  }
});

const data = await response.json();
console.log(data);`;

      case 'python':
        return `import requests

response = requests.${selectedEndpoint.method.toLowerCase()}(
    '${url}',
    headers=${JSON.stringify(requestData.headers)},${
      requestData.body && ['POST', 'PUT', 'PATCH'].includes(selectedEndpoint.method.toUpperCase())
        ? `\n    json=${requestData.body}`
        : ''
    }
)

print(response.json())`;

      default:
        return '';
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
  };

  const getMethodColor = (method: string) => {
    switch (method.toUpperCase()) {
      case 'GET': return 'text-green-600 bg-green-50';
      case 'POST': return 'text-blue-600 bg-blue-50';
      case 'PUT': return 'text-yellow-600 bg-yellow-50';
      case 'DELETE': return 'text-red-600 bg-red-50';
      case 'PATCH': return 'text-purple-600 bg-purple-50';
      default: return 'text-gray-600 bg-gray-50';
    }
  };

  if (!apiKey && onSubscribe) {
    return (
      <div className="bg-gradient-to-r from-blue-50 to-indigo-50 rounded-lg p-8 text-center">
        <div className="max-w-md mx-auto">
          <CodeBracketIcon className="h-16 w-16 text-blue-500 mx-auto mb-4" />
          <h3 className="text-xl font-semibold text-gray-900 mb-2">
            Try This API
          </h3>
          <p className="text-gray-600 mb-6">
            Subscribe to get an API key and test all endpoints in our interactive playground.
          </p>
          <button
            onClick={onSubscribe}
            className="bg-blue-600 text-white px-6 py-3 rounded-lg font-medium hover:bg-blue-700 transition-colors"
          >
            Subscribe to Test API
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg border border-gray-200 overflow-hidden">
      {/* Header */}
      <div className="bg-gray-50 px-6 py-4 border-b border-gray-200">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold text-gray-900">API Playground</h3>
          <div className="flex items-center space-x-2">
            <button
              onClick={() => setShowCodeGen(!showCodeGen)}
              className="flex items-center space-x-2 px-3 py-1 text-sm text-gray-600 hover:text-gray-900 transition-colors"
            >
              <CodeBracketIcon className="h-4 w-4" />
              <span>Code</span>
            </button>
          </div>
        </div>
      </div>

      <div className="flex">
        {/* Endpoint Selection */}
        <div className="w-1/3 border-r border-gray-200">
          <div className="p-4">
            <h4 className="text-sm font-medium text-gray-900 mb-3">Endpoints</h4>
            <div className="space-y-2">
              {endpoints.map((endpoint, index) => (
                <button
                  key={index}
                  onClick={() => setSelectedEndpoint(endpoint)}
                  className={`w-full text-left p-3 rounded-lg border transition-colors ${
                    selectedEndpoint === endpoint
                      ? 'border-blue-200 bg-blue-50'
                      : 'border-gray-200 hover:border-gray-300'
                  }`}
                >
                  <div className="flex items-center space-x-2 mb-1">
                    <span className={`px-2 py-1 text-xs font-medium rounded ${getMethodColor(endpoint.method)}`}>
                      {endpoint.method.toUpperCase()}
                    </span>
                    <span className="text-sm font-mono text-gray-900">{endpoint.path}</span>
                  </div>
                  {endpoint.description && (
                    <p className="text-xs text-gray-600">{endpoint.description}</p>
                  )}
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* Request Configuration */}
        <div className="flex-1">
          {selectedEndpoint && (
            <div className="p-6">
              {/* Endpoint Info */}
              <div className="mb-6">
                <div className="flex items-center space-x-3 mb-2">
                  <span className={`px-3 py-1 text-sm font-medium rounded ${getMethodColor(selectedEndpoint.method)}`}>
                    {selectedEndpoint.method.toUpperCase()}
                  </span>
                  <span className="text-lg font-mono text-gray-900">{selectedEndpoint.path}</span>
                </div>
                {selectedEndpoint.description && (
                  <p className="text-gray-600">{selectedEndpoint.description}</p>
                )}
              </div>

              {/* Parameters */}
              {selectedEndpoint.parameters && selectedEndpoint.parameters.length > 0 && (
                <div className="mb-6">
                  <h5 className="text-sm font-medium text-gray-900 mb-3">Parameters</h5>
                  <div className="space-y-3">
                    {selectedEndpoint.parameters.map((param, index) => (
                      <div key={index}>
                        <label className="block text-sm font-medium text-gray-700 mb-1">
                          {param.name}
                          {param.required && <span className="text-red-500 ml-1">*</span>}
                          <span className="text-xs text-gray-500 ml-2">({param.in})</span>
                        </label>
                        <input
                          type="text"
                          value={
                            param.in === 'query' ? requestData.queryParams[param.name] || '' :
                            param.in === 'path' ? requestData.pathParams[param.name] || '' :
                            requestData.headers[param.name] || ''
                          }
                          onChange={(e) => {
                            const value = e.target.value;
                            setRequestData(prev => ({
                              ...prev,
                              [param.in === 'query' ? 'queryParams' : 
                               param.in === 'path' ? 'pathParams' : 'headers']: {
                                ...prev[param.in === 'query' ? 'queryParams' : 
                                       param.in === 'path' ? 'pathParams' : 'headers'],
                                [param.name]: value
                              }
                            }));
                          }}
                          placeholder={param.description || `Enter ${param.name}`}
                          className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        />
                        {param.description && (
                          <p className="text-xs text-gray-500 mt-1">{param.description}</p>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              )}

              {/* Request Body */}
              {['POST', 'PUT', 'PATCH'].includes(selectedEndpoint.method.toUpperCase()) && (
                <div className="mb-6">
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Request Body
                  </label>
                  <textarea
                    value={requestData.body}
                    onChange={(e) => setRequestData(prev => ({ ...prev, body: e.target.value }))}
                    placeholder="Enter JSON request body"
                    rows={8}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent font-mono text-sm"
                  />
                </div>
              )}

              {/* Headers */}
              <div className="mb-6">
                <h5 className="text-sm font-medium text-gray-900 mb-3">Headers</h5>
                <div className="space-y-2">
                  {Object.entries(requestData.headers).map(([key, value]) => (
                    <div key={key} className="flex space-x-2">
                      <input
                        type="text"
                        value={key}
                        onChange={(e) => {
                          const newKey = e.target.value;
                          setRequestData(prev => {
                            const newHeaders = { ...prev.headers };
                            delete newHeaders[key];
                            newHeaders[newKey] = value;
                            return { ...prev, headers: newHeaders };
                          });
                        }}
                        placeholder="Header name"
                        className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                      <input
                        type="text"
                        value={value}
                        onChange={(e) => {
                          setRequestData(prev => ({
                            ...prev,
                            headers: { ...prev.headers, [key]: e.target.value }
                          }));
                        }}
                        placeholder="Header value"
                        className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      />
                      <button
                        onClick={() => {
                          setRequestData(prev => {
                            const newHeaders = { ...prev.headers };
                            delete newHeaders[key];
                            return { ...prev, headers: newHeaders };
                          });
                        }}
                        className="px-3 py-2 text-red-600 hover:text-red-800"
                      >
                        ×
                      </button>
                    </div>
                  ))}
                  <button
                    onClick={() => {
                      setRequestData(prev => ({
                        ...prev,
                        headers: { ...prev.headers, '': '' }
                      }));
                    }}
                    className="text-sm text-blue-600 hover:text-blue-800"
                  >
                    + Add Header
                  </button>
                </div>
              </div>

              {/* Execute Button */}
              <button
                onClick={executeRequest}
                disabled={response.loading}
                className="w-full bg-blue-600 text-white py-3 px-4 rounded-lg font-medium hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center space-x-2"
              >
                {response.loading ? (
                  <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-white"></div>
                ) : (
                  <>
                    <PlayIcon className="h-5 w-5" />
                    <span>Send Request</span>
                  </>
                )}
              </button>

              {/* Response */}
              {(response.status || response.error) && (
                <div className="mt-6">
                  <div className="flex items-center justify-between mb-3">
                    <h5 className="text-sm font-medium text-gray-900">Response</h5>
                    {response.body && (
                      <button
                        onClick={() => copyToClipboard(response.body!)}
                        className="flex items-center space-x-1 text-sm text-gray-600 hover:text-gray-900"
                      >
                        <ClipboardIcon className="h-4 w-4" />
                        <span>Copy</span>
                      </button>
                    )}
                  </div>
                  
                  {response.error ? (
                    <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                      <p className="text-red-800">{response.error}</p>
                    </div>
                  ) : (
                    <div className="border border-gray-200 rounded-lg">
                      <div className="bg-gray-50 px-4 py-2 border-b border-gray-200">
                        <span className={`inline-flex items-center px-2 py-1 rounded text-sm font-medium ${
                          response.status && response.status < 300 ? 'bg-green-100 text-green-800' :
                          response.status && response.status < 400 ? 'bg-yellow-100 text-yellow-800' :
                          'bg-red-100 text-red-800'
                        }`}>
                          {response.status}
                        </span>
                      </div>
                      <pre className="p-4 text-sm font-mono text-gray-900 overflow-x-auto">
                        {response.body}
                      </pre>
                    </div>
                  )}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Code Generation Modal */}
      {showCodeGen && selectedEndpoint && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg max-w-4xl w-full mx-4 max-h-[80vh] overflow-hidden">
            <div className="flex items-center justify-between p-6 border-b border-gray-200">
              <h3 className="text-lg font-semibold text-gray-900">Code Examples</h3>
              <button
                onClick={() => setShowCodeGen(false)}
                className="text-gray-400 hover:text-gray-600"
              >
                ×
              </button>
            </div>
            <div className="p-6">
              <div className="flex space-x-4 mb-4">
                {['curl', 'javascript', 'python'].map((lang) => (
                  <button
                    key={lang}
                    onClick={() => setSelectedLanguage(lang)}
                    className={`px-3 py-2 text-sm font-medium rounded ${
                      selectedLanguage === lang
                        ? 'bg-blue-100 text-blue-700'
                        : 'text-gray-600 hover:text-gray-900'
                    }`}
                  >
                    {lang === 'curl' ? 'cURL' : lang.charAt(0).toUpperCase() + lang.slice(1)}
                  </button>
                ))}
              </div>
              <div className="relative">
                <pre className="bg-gray-900 text-gray-100 p-4 rounded-lg overflow-x-auto text-sm">
                  {generateCode(selectedLanguage)}
                </pre>
                <button
                  onClick={() => copyToClipboard(generateCode(selectedLanguage))}
                  className="absolute top-2 right-2 p-2 text-gray-400 hover:text-gray-200"
                >
                  <ClipboardIcon className="h-4 w-4" />
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default APIPlayground;

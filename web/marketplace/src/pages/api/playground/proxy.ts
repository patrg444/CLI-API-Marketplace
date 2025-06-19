import { NextApiRequest, NextApiResponse } from 'next';

interface ProxyRequest {
  url: string;
  options: {
    method: string;
    headers: { [key: string]: string };
    body?: string;
  };
}

export default async function handler(req: NextApiRequest, res: NextApiResponse) {
  if (req.method !== 'POST') {
    return res.status(405).json({ error: 'Method not allowed' });
  }

  try {
    const { url, options }: ProxyRequest = req.body;

    if (!url || !options) {
      return res.status(400).json({ error: 'Missing url or options' });
    }

    // Validate URL to prevent SSRF attacks
    const parsedUrl = new URL(url);
    const allowedHosts = [
      'localhost',
      '127.0.0.1',
      // Add your API-Direct platform domains here
      'api.api-direct.io',
      'staging-api.api-direct.io'
    ];

    // Allow any subdomain of api-direct.io
    const isApiDirectDomain = parsedUrl.hostname.endsWith('.api-direct.io') || 
                             parsedUrl.hostname === 'api-direct.io';
    
    const isLocalhost = parsedUrl.hostname === 'localhost' || 
                       parsedUrl.hostname === '127.0.0.1' ||
                       parsedUrl.hostname.startsWith('192.168.') ||
                       parsedUrl.hostname.startsWith('10.') ||
                       parsedUrl.hostname.startsWith('172.');

    if (!isApiDirectDomain && !isLocalhost) {
      return res.status(403).json({ 
        error: 'Forbidden: Only API-Direct platform URLs are allowed' 
      });
    }

    // Prepare the request
    const fetchOptions: RequestInit = {
      method: options.method,
      headers: {
        ...options.headers,
        // Add user agent to identify playground requests
        'User-Agent': 'API-Direct-Playground/1.0'
      }
    };

    // Add body for POST/PUT/PATCH requests
    if (options.body && ['POST', 'PUT', 'PATCH'].includes(options.method.toUpperCase())) {
      fetchOptions.body = options.body;
    }

    // Set timeout for the request
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 30000); // 30 second timeout
    fetchOptions.signal = controller.signal;

    try {
      // Make the request
      const response = await fetch(url, fetchOptions);
      clearTimeout(timeoutId);

      // Get response headers
      const responseHeaders: { [key: string]: string } = {};
      response.headers.forEach((value, key) => {
        responseHeaders[key] = value;
      });

      // Get response body
      let responseData;
      const contentType = response.headers.get('content-type');
      
      if (contentType && contentType.includes('application/json')) {
        try {
          responseData = await response.json();
        } catch (e) {
          responseData = await response.text();
        }
      } else {
        responseData = await response.text();
      }

      // Return the response
      return res.status(200).json({
        status: response.status,
        statusText: response.statusText,
        headers: responseHeaders,
        data: responseData
      });

    } catch (fetchError) {
      clearTimeout(timeoutId);
      
      if (fetchError instanceof Error) {
        if (fetchError.name === 'AbortError') {
          return res.status(408).json({ 
            error: 'Request timeout',
            details: 'The API request took too long to respond'
          });
        }
        
        return res.status(500).json({ 
          error: 'Request failed',
          details: fetchError.message
        });
      }
      
      return res.status(500).json({ 
        error: 'Unknown error occurred'
      });
    }

  } catch (error) {
    console.error('Proxy error:', error);
    
    if (error instanceof TypeError && error.message.includes('Invalid URL')) {
      return res.status(400).json({ 
        error: 'Invalid URL',
        details: 'The provided URL is not valid'
      });
    }
    
    return res.status(500).json({ 
      error: 'Internal server error',
      details: error instanceof Error ? error.message : 'Unknown error'
    });
  }
}

// Increase the body size limit for larger request bodies
export const config = {
  api: {
    bodyParser: {
      sizeLimit: '1mb',
    },
  },
};

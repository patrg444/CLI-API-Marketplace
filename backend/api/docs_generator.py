"""
API Documentation Generator
Generates OpenAPI specs, interactive docs, and code examples
"""

import asyncpg
import logging
import json
import yaml
from typing import Dict, Any, List, Optional
from datetime import datetime
from fastapi import HTTPException
import httpx
from pydantic import BaseModel, Field
from uuid import UUID

logger = logging.getLogger(__name__)


class EndpointDoc(BaseModel):
    """Documentation for a single endpoint"""
    path: str
    method: str
    summary: str
    description: Optional[str] = None
    parameters: List[Dict[str, Any]] = []
    request_body: Optional[Dict[str, Any]] = None
    responses: Dict[str, Dict[str, Any]] = {}
    tags: List[str] = []
    deprecated: bool = False
    security: List[Dict[str, List[str]]] = []


class APIDocumentation(BaseModel):
    """Complete API documentation"""
    openapi: str = "3.0.0"
    info: Dict[str, Any]
    servers: List[Dict[str, str]] = []
    paths: Dict[str, Dict[str, Any]] = {}
    components: Dict[str, Any] = Field(default_factory=dict)
    security: List[Dict[str, List[str]]] = []
    tags: List[Dict[str, str]] = []


class DocsGenerator:
    """Generates API documentation and code examples"""
    
    def __init__(self, db_pool: asyncpg.Pool):
        self.db_pool = db_pool
        
        # Code templates for different languages
        self.code_templates = {
            "curl": self._generate_curl,
            "python": self._generate_python,
            "javascript": self._generate_javascript,
            "go": self._generate_go,
            "java": self._generate_java,
            "php": self._generate_php
        }
    
    async def generate_openapi_spec(self, api_id: str) -> Dict[str, Any]:
        """
        Generate OpenAPI 3.0 specification for an API
        """
        async with self.db_pool.acquire() as conn:
            # Get API details
            api = await conn.fetchrow("""
                SELECT a.*, u.name as owner_name, u.email as owner_email
                FROM apis a
                JOIN users u ON a.user_id = u.id
                WHERE a.id = $1
            """, api_id)
            
            if not api:
                raise HTTPException(status_code=404, detail="API not found")
            
            # Get endpoints
            endpoints = await conn.fetch("""
                SELECT * FROM api_endpoints
                WHERE api_id = $1
                ORDER BY path, method
            """, api_id)
            
            # Build OpenAPI spec
            spec = APIDocumentation(
                info={
                    "title": api['name'],
                    "version": api.get('version', '1.0.0'),
                    "description": api.get('description', ''),
                    "contact": {
                        "name": api['owner_name'],
                        "email": api['owner_email']
                    },
                    "x-logo": {
                        "url": api.get('logo_url', ''),
                        "altText": f"{api['name']} logo"
                    }
                },
                servers=[
                    {
                        "url": api['base_url'],
                        "description": "Production server"
                    }
                ]
            )
            
            # Add sandbox server if available
            if api.get('sandbox_enabled') and api.get('sandbox_base_url'):
                spec.servers.append({
                    "url": api['sandbox_base_url'],
                    "description": "Sandbox server (for testing)"
                })
            
            # Add authentication
            if api.get('auth_type'):
                spec.components['securitySchemes'] = self._generate_security_schemes(api)
                spec.security = [{"apiKey": []}]
            
            # Add endpoints
            for endpoint in endpoints:
                path = endpoint['path']
                method = endpoint['method'].lower()
                
                if path not in spec.paths:
                    spec.paths[path] = {}
                
                spec.paths[path][method] = await self._generate_endpoint_doc(
                    api, endpoint, conn
                )
            
            # Add schemas
            schemas = await conn.fetch("""
                SELECT * FROM api_schemas
                WHERE api_id = $1
            """, api_id)
            
            if schemas:
                spec.components['schemas'] = {}
                for schema in schemas:
                    spec.components['schemas'][schema['name']] = schema['schema']
            
            # Add tags
            categories = await conn.fetch("""
                SELECT DISTINCT category FROM api_endpoints
                WHERE api_id = $1 AND category IS NOT NULL
            """, api_id)
            
            spec.tags = [
                {"name": cat['category'], "description": f"{cat['category']} endpoints"}
                for cat in categories
            ]
            
            return spec.dict()
    
    async def _generate_endpoint_doc(
        self, 
        api: Dict, 
        endpoint: Dict, 
        conn: asyncpg.Connection
    ) -> Dict[str, Any]:
        """Generate documentation for a single endpoint"""
        doc = {
            "summary": endpoint.get('summary', f"{endpoint['method']} {endpoint['path']}"),
            "description": endpoint.get('description', ''),
            "operationId": endpoint.get('operation_id', f"{endpoint['method'].lower()}_{endpoint['path'].replace('/', '_')}"),
            "tags": [endpoint['category']] if endpoint.get('category') else [],
            "parameters": [],
            "responses": {}
        }
        
        # Add parameters
        params = await conn.fetch("""
            SELECT * FROM api_parameters
            WHERE endpoint_id = $1
            ORDER BY required DESC, name
        """, endpoint['id'])
        
        for param in params:
            param_doc = {
                "name": param['name'],
                "in": param['location'],  # path, query, header, cookie
                "description": param.get('description', ''),
                "required": param.get('required', False),
                "schema": param.get('schema', {"type": "string"})
            }
            
            if param.get('example'):
                param_doc['example'] = param['example']
            
            doc['parameters'].append(param_doc)
        
        # Add request body
        if endpoint['method'] in ['POST', 'PUT', 'PATCH']:
            request_body = await conn.fetchrow("""
                SELECT * FROM api_request_bodies
                WHERE endpoint_id = $1
            """, endpoint['id'])
            
            if request_body:
                doc['requestBody'] = {
                    "description": request_body.get('description', ''),
                    "required": request_body.get('required', True),
                    "content": {
                        "application/json": {
                            "schema": request_body['schema'],
                            "examples": request_body.get('examples', {})
                        }
                    }
                }
        
        # Add responses
        responses = await conn.fetch("""
            SELECT * FROM api_responses
            WHERE endpoint_id = $1
            ORDER BY status_code
        """, endpoint['id'])
        
        for response in responses:
            doc['responses'][str(response['status_code'])] = {
                "description": response.get('description', ''),
                "content": {
                    "application/json": {
                        "schema": response.get('schema', {}),
                        "examples": response.get('examples', {})
                    }
                }
            }
        
        # Add default responses if none specified
        if not doc['responses']:
            doc['responses'] = {
                "200": {
                    "description": "Successful response",
                    "content": {
                        "application/json": {
                            "schema": {"type": "object"}
                        }
                    }
                },
                "400": {"description": "Bad request"},
                "401": {"description": "Unauthorized"},
                "500": {"description": "Internal server error"}
            }
        
        # Add security if required
        if endpoint.get('requires_auth', True):
            doc['security'] = [{"apiKey": []}]
        
        return doc
    
    def _generate_security_schemes(self, api: Dict) -> Dict[str, Any]:
        """Generate security schemes based on API auth type"""
        auth_type = api.get('auth_type', 'apiKey')
        
        if auth_type == 'apiKey':
            return {
                "apiKey": {
                    "type": "apiKey",
                    "in": api.get('auth_location', 'header'),
                    "name": api.get('auth_header', 'X-API-Key')
                }
            }
        elif auth_type == 'bearer':
            return {
                "bearerAuth": {
                    "type": "http",
                    "scheme": "bearer",
                    "bearerFormat": "JWT"
                }
            }
        elif auth_type == 'oauth2':
            return {
                "oauth2": {
                    "type": "oauth2",
                    "flows": {
                        "authorizationCode": {
                            "authorizationUrl": api.get('oauth_auth_url', ''),
                            "tokenUrl": api.get('oauth_token_url', ''),
                            "scopes": api.get('oauth_scopes', {})
                        }
                    }
                }
            }
        else:
            return {}
    
    async def generate_code_examples(
        self, 
        api_id: str, 
        endpoint_path: str,
        method: str,
        languages: List[str] = None
    ) -> Dict[str, str]:
        """
        Generate code examples for an endpoint in multiple languages
        """
        if languages is None:
            languages = ["curl", "python", "javascript"]
        
        async with self.db_pool.acquire() as conn:
            # Get API and endpoint details
            api = await conn.fetchrow("""
                SELECT * FROM apis WHERE id = $1
            """, api_id)
            
            endpoint = await conn.fetchrow("""
                SELECT e.*, 
                       array_agg(
                           jsonb_build_object(
                               'name', p.name,
                               'location', p.location,
                               'required', p.required,
                               'example', p.example
                           )
                       ) FILTER (WHERE p.id IS NOT NULL) as parameters
                FROM api_endpoints e
                LEFT JOIN api_parameters p ON e.id = p.endpoint_id
                WHERE e.api_id = $1 AND e.path = $2 AND e.method = $3
                GROUP BY e.id
            """, api_id, endpoint_path, method)
            
            if not api or not endpoint:
                raise HTTPException(status_code=404, detail="Endpoint not found")
            
            # Get request body example
            request_body = None
            if method in ['POST', 'PUT', 'PATCH']:
                body_data = await conn.fetchrow("""
                    SELECT examples FROM api_request_bodies
                    WHERE endpoint_id = $1
                    LIMIT 1
                """, endpoint['id'])
                
                if body_data and body_data['examples']:
                    # Get first example
                    examples = body_data['examples']
                    if isinstance(examples, dict) and examples:
                        request_body = list(examples.values())[0]
            
            # Generate code for each language
            examples = {}
            for lang in languages:
                if lang in self.code_templates:
                    examples[lang] = self.code_templates[lang](
                        api, endpoint, request_body
                    )
            
            return examples
    
    def _generate_curl(self, api: Dict, endpoint: Dict, request_body: Any) -> str:
        """Generate cURL example"""
        url = f"{api['base_url']}{endpoint['path']}"
        method = endpoint['method']
        
        # Build cURL command
        cmd = [f"curl -X {method}"]
        cmd.append(f'"{url}"')
        
        # Add headers
        cmd.append(f'-H "Accept: application/json"')
        
        if api.get('auth_type') == 'apiKey':
            cmd.append(f'-H "{api.get("auth_header", "X-API-Key")}: YOUR_API_KEY"')
        elif api.get('auth_type') == 'bearer':
            cmd.append(f'-H "Authorization: Bearer YOUR_TOKEN"')
        
        # Add request body
        if request_body and method in ['POST', 'PUT', 'PATCH']:
            cmd.append(f'-H "Content-Type: application/json"')
            cmd.append(f"-d '{json.dumps(request_body, indent=2)}'")
        
        # Add parameters
        if endpoint.get('parameters'):
            query_params = [
                p for p in endpoint['parameters'] 
                if p['location'] == 'query' and p.get('example')
            ]
            if query_params:
                query_string = "&".join([
                    f"{p['name']}={p['example']}" 
                    for p in query_params
                ])
                url += f"?{query_string}"
        
        return " \\\n  ".join(cmd)
    
    def _generate_python(self, api: Dict, endpoint: Dict, request_body: Any) -> str:
        """Generate Python example using requests"""
        code = ["import requests", ""]
        
        # URL
        code.append(f'url = "{api["base_url"]}{endpoint["path"]}"')
        
        # Headers
        code.append("\nheaders = {")
        code.append('    "Accept": "application/json"')
        
        if api.get('auth_type') == 'apiKey':
            code.append(f',    "{api.get("auth_header", "X-API-Key")}": "YOUR_API_KEY"')
        elif api.get('auth_type') == 'bearer':
            code.append(',    "Authorization": "Bearer YOUR_TOKEN"')
        
        code.append("}")
        
        # Parameters
        params = [
            p for p in endpoint.get('parameters', []) 
            if p['location'] == 'query' and p.get('example')
        ]
        if params:
            code.append("\nparams = {")
            for p in params:
                code.append(f'    "{p["name"]}": "{p["example"]}"')
            code.append("}")
        
        # Request body
        if request_body and endpoint['method'] in ['POST', 'PUT', 'PATCH']:
            code.append("\ndata = " + json.dumps(request_body, indent=4))
        
        # Make request
        code.append("")
        method = endpoint['method'].lower()
        
        if endpoint['method'] == 'GET':
            if params:
                code.append(f"response = requests.{method}(url, headers=headers, params=params)")
            else:
                code.append(f"response = requests.{method}(url, headers=headers)")
        elif request_body:
            if params:
                code.append(f"response = requests.{method}(url, headers=headers, params=params, json=data)")
            else:
                code.append(f"response = requests.{method}(url, headers=headers, json=data)")
        else:
            code.append(f"response = requests.{method}(url, headers=headers)")
        
        code.append("\nprint(response.json())")
        
        return "\n".join(code)
    
    def _generate_javascript(self, api: Dict, endpoint: Dict, request_body: Any) -> str:
        """Generate JavaScript example using fetch"""
        code = []
        
        # URL with parameters
        url = f"{api['base_url']}{endpoint['path']}"
        params = [
            p for p in endpoint.get('parameters', []) 
            if p['location'] == 'query' and p.get('example')
        ]
        if params:
            query_string = "&".join([
                f"{p['name']}={p['example']}" 
                for p in params
            ])
            url += f"?{query_string}"
        
        code.append(f'const url = "{url}";')
        
        # Options
        code.append("\nconst options = {")
        code.append(f'  method: "{endpoint["method"]}",')
        code.append("  headers: {")
        code.append('    "Accept": "application/json"')
        
        if api.get('auth_type') == 'apiKey':
            code.append(f',    "{api.get("auth_header", "X-API-Key")}": "YOUR_API_KEY"')
        elif api.get('auth_type') == 'bearer':
            code.append(',    "Authorization": "Bearer YOUR_TOKEN"')
        
        if request_body and endpoint['method'] in ['POST', 'PUT', 'PATCH']:
            code.append(',    "Content-Type": "application/json"')
        
        code.append("  }")
        
        if request_body and endpoint['method'] in ['POST', 'PUT', 'PATCH']:
            code.append(f",  body: JSON.stringify({json.dumps(request_body, indent=2)})")
        
        code.append("};")
        
        # Fetch
        code.append("\nfetch(url, options)")
        code.append("  .then(response => response.json())")
        code.append("  .then(data => console.log(data))")
        code.append("  .catch(error => console.error('Error:', error));")
        
        return "\n".join(code)
    
    def _generate_go(self, api: Dict, endpoint: Dict, request_body: Any) -> str:
        """Generate Go example"""
        code = ['package main', '', 'import (', '    "bytes"', '    "encoding/json"',
                '    "fmt"', '    "io/ioutil"', '    "net/http"', ')', '', 'func main() {']
        
        # URL
        code.append(f'    url := "{api["base_url"]}{endpoint["path"]}"')
        
        # Request body
        if request_body and endpoint['method'] in ['POST', 'PUT', 'PATCH']:
            code.append(f'\n    payload, _ := json.Marshal({json.dumps(request_body)})')
            code.append('    req, _ := http.NewRequest("' + endpoint['method'] + '", url, bytes.NewBuffer(payload))')
        else:
            code.append(f'\n    req, _ := http.NewRequest("{endpoint["method"]}", url, nil)')
        
        # Headers
        code.append('\n    req.Header.Set("Accept", "application/json")')
        
        if api.get('auth_type') == 'apiKey':
            code.append(f'    req.Header.Set("{api.get("auth_header", "X-API-Key")}", "YOUR_API_KEY")')
        
        # Make request
        code.append('\n    client := &http.Client{}')
        code.append('    resp, err := client.Do(req)')
        code.append('    if err != nil {')
        code.append('        panic(err)')
        code.append('    }')
        code.append('    defer resp.Body.Close()')
        code.append('\n    body, _ := ioutil.ReadAll(resp.Body)')
        code.append('    fmt.Println(string(body))')
        code.append('}')
        
        return "\n".join(code)
    
    def _generate_java(self, api: Dict, endpoint: Dict, request_body: Any) -> str:
        """Generate Java example"""
        code = ['import java.io.IOException;', 'import java.net.URI;', 'import java.net.http.HttpClient;',
                'import java.net.http.HttpRequest;', 'import java.net.http.HttpResponse;', '',
                'public class APIExample {', '    public static void main(String[] args) throws IOException, InterruptedException {']
        
        # URL
        code.append(f'        String url = "{api["base_url"]}{endpoint["path"]}";')
        
        # Build request
        code.append('\n        HttpRequest.Builder requestBuilder = HttpRequest.newBuilder()')
        code.append('            .uri(URI.create(url))')
        code.append(f'            .method("{endpoint["method"]}", HttpRequest.BodyPublishers.ofString(""))')
        code.append('            .header("Accept", "application/json")')
        
        if api.get('auth_type') == 'apiKey':
            code.append(f'            .header("{api.get("auth_header", "X-API-Key")}", "YOUR_API_KEY")')
        
        if request_body and endpoint['method'] in ['POST', 'PUT', 'PATCH']:
            body_json = json.dumps(request_body)
            code.append(f'            .method("{endpoint["method"]}", HttpRequest.BodyPublishers.ofString("{body_json}"))')
            code.append('            .header("Content-Type", "application/json")')
        
        code.append('            ;')
        
        # Send request
        code.append('\n        HttpClient client = HttpClient.newHttpClient();')
        code.append('        HttpResponse<String> response = client.send(requestBuilder.build(), HttpResponse.BodyHandlers.ofString());')
        code.append('\n        System.out.println(response.body());')
        code.append('    }')
        code.append('}')
        
        return "\n".join(code)
    
    def _generate_php(self, api: Dict, endpoint: Dict, request_body: Any) -> str:
        """Generate PHP example"""
        code = ['<?php', '']
        
        # URL
        code.append(f'$url = "{api["base_url"]}{endpoint["path"]}";')
        
        # Headers
        code.append('$headers = [')
        code.append('    "Accept: application/json"')
        
        if api.get('auth_type') == 'apiKey':
            code.append(f',    "{api.get("auth_header", "X-API-Key")}: YOUR_API_KEY"')
        
        code.append('];')
        
        # Setup cURL
        code.append('\n$ch = curl_init();')
        code.append('curl_setopt($ch, CURLOPT_URL, $url);')
        code.append('curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);')
        code.append(f'curl_setopt($ch, CURLOPT_CUSTOMREQUEST, "{endpoint["method"]}");')
        code.append('curl_setopt($ch, CURLOPT_HTTPHEADER, $headers);')
        
        # Request body
        if request_body and endpoint['method'] in ['POST', 'PUT', 'PATCH']:
            code.append(f'\n$data = \'{json.dumps(request_body)}\';')
            code.append('curl_setopt($ch, CURLOPT_POSTFIELDS, $data);')
        
        # Execute
        code.append('\n$response = curl_exec($ch);')
        code.append('curl_close($ch);')
        code.append('\necho $response;')
        code.append('?>')
        
        return "\n".join(code)
    
    async def export_postman_collection(self, api_id: str) -> Dict[str, Any]:
        """
        Export API as Postman collection
        """
        async with self.db_pool.acquire() as conn:
            # Get API details
            api = await conn.fetchrow("""
                SELECT * FROM apis WHERE id = $1
            """, api_id)
            
            if not api:
                raise HTTPException(status_code=404, detail="API not found")
            
            # Get endpoints
            endpoints = await conn.fetch("""
                SELECT * FROM api_endpoints
                WHERE api_id = $1
                ORDER BY category, path, method
            """, api_id)
            
            # Build Postman collection
            collection = {
                "info": {
                    "name": api['name'],
                    "description": api.get('description', ''),
                    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
                },
                "auth": {
                    "type": "apikey",
                    "apikey": [{
                        "key": "key",
                        "value": "{{api_key}}",
                        "type": "string"
                    }, {
                        "key": "in",
                        "value": api.get('auth_location', 'header'),
                        "type": "string"
                    }, {
                        "key": "name",
                        "value": api.get('auth_header', 'X-API-Key'),
                        "type": "string"
                    }]
                },
                "variable": [
                    {
                        "key": "base_url",
                        "value": api['base_url'],
                        "type": "string"
                    },
                    {
                        "key": "api_key",
                        "value": "YOUR_API_KEY",
                        "type": "string"
                    }
                ],
                "item": []
            }
            
            # Group endpoints by category
            categories = {}
            for endpoint in endpoints:
                category = endpoint.get('category', 'General')
                if category not in categories:
                    categories[category] = {
                        "name": category,
                        "item": []
                    }
                
                # Build request
                request = {
                    "name": endpoint.get('summary', f"{endpoint['method']} {endpoint['path']}"),
                    "request": {
                        "method": endpoint['method'],
                        "header": [],
                        "url": {
                            "raw": "{{base_url}}" + endpoint['path'],
                            "host": ["{{base_url}}"],
                            "path": endpoint['path'].strip('/').split('/')
                        }
                    }
                }
                
                # Add description
                if endpoint.get('description'):
                    request['request']['description'] = endpoint['description']
                
                # Add parameters
                params = await conn.fetch("""
                    SELECT * FROM api_parameters
                    WHERE endpoint_id = $1
                """, endpoint['id'])
                
                query_params = []
                for param in params:
                    if param['location'] == 'query':
                        query_params.append({
                            "key": param['name'],
                            "value": param.get('example', ''),
                            "description": param.get('description', ''),
                            "disabled": not param.get('required', False)
                        })
                    elif param['location'] == 'header':
                        request['request']['header'].append({
                            "key": param['name'],
                            "value": param.get('example', ''),
                            "description": param.get('description', ''),
                            "disabled": not param.get('required', False)
                        })
                
                if query_params:
                    request['request']['url']['query'] = query_params
                
                # Add request body
                if endpoint['method'] in ['POST', 'PUT', 'PATCH']:
                    body_data = await conn.fetchrow("""
                        SELECT * FROM api_request_bodies
                        WHERE endpoint_id = $1
                    """, endpoint['id'])
                    
                    if body_data:
                        request['request']['body'] = {
                            "mode": "raw",
                            "raw": json.dumps(
                                list(body_data.get('examples', {}).values())[0] 
                                if body_data.get('examples') 
                                else {},
                                indent=2
                            ),
                            "options": {
                                "raw": {
                                    "language": "json"
                                }
                            }
                        }
                        request['request']['header'].append({
                            "key": "Content-Type",
                            "value": "application/json"
                        })
                
                categories[category]['item'].append(request)
            
            # Add categories to collection
            collection['item'] = list(categories.values())
            
            return collection
    
    async def generate_sdk(self, api_id: str, language: str) -> str:
        """
        Generate SDK code for an API
        Currently supports Python and JavaScript
        """
        if language not in ['python', 'javascript']:
            raise HTTPException(
                status_code=400,
                detail="SDK generation currently supports Python and JavaScript"
            )
        
        async with self.db_pool.acquire() as conn:
            # Get API and endpoints
            api = await conn.fetchrow("""
                SELECT * FROM apis WHERE id = $1
            """, api_id)
            
            endpoints = await conn.fetch("""
                SELECT * FROM api_endpoints
                WHERE api_id = $1
                ORDER BY category, path
            """, api_id)
            
            if language == 'python':
                return self._generate_python_sdk(api, endpoints)
            else:
                return self._generate_javascript_sdk(api, endpoints)
    
    def _generate_python_sdk(self, api: Dict, endpoints: List[Dict]) -> str:
        """Generate Python SDK"""
        class_name = ''.join(word.capitalize() for word in api['name'].split())
        
        code = [
            f'"""',
            f'{api["name"]} Python SDK',
            f'',
            f'{api.get("description", "")}',
            f'"""',
            '',
            'import requests',
            'from typing import Dict, Any, Optional',
            '',
            '',
            f'class {class_name}Client:',
            f'    """Client for {api["name"]} API"""',
            '',
            f'    def __init__(self, api_key: str, base_url: str = "{api["base_url"]}"):',
            '        self.api_key = api_key',
            '        self.base_url = base_url.rstrip("/")',
            '        self.session = requests.Session()',
            f'        self.session.headers.update({{',
            f'            "{api.get("auth_header", "X-API-Key")}": api_key,',
            f'            "Accept": "application/json"',
            f'        }})',
            ''
        ]
        
        # Group endpoints by category
        categories = {}
        for endpoint in endpoints:
            cat = endpoint.get('category', 'general')
            if cat not in categories:
                categories[cat] = []
            categories[cat].append(endpoint)
        
        # Generate methods for each endpoint
        for category, eps in categories.items():
            code.append(f'    # {category.title()} endpoints')
            
            for ep in eps:
                method_name = self._endpoint_to_method_name(ep['path'], ep['method'])
                params = ['self']
                
                # Add path parameters
                path_params = re.findall(r'{(\w+)}', ep['path'])
                params.extend(path_params)
                
                # Add optional parameters
                if ep['method'] in ['POST', 'PUT', 'PATCH']:
                    params.append('data: Dict[str, Any]')
                
                params.append('**kwargs')
                
                code.append(f'    def {method_name}({", ".join(params)}) -> Dict[str, Any]:')
                code.append(f'        """')
                code.append(f'        {ep.get("summary", ep["method"] + " " + ep["path"])}')
                if ep.get('description'):
                    code.append(f'        ')
                    code.append(f'        {ep["description"]}')
                code.append(f'        """')
                
                # Build URL
                path = ep['path']
                for param in path_params:
                    path = path.replace(f'{{{param}}}', f'{{str({param})}}')
                
                code.append(f'        url = f"{{self.base_url}}{path}"')
                
                # Make request
                if ep['method'] == 'GET':
                    code.append(f'        response = self.session.get(url, params=kwargs)')
                elif ep['method'] == 'DELETE':
                    code.append(f'        response = self.session.delete(url)')
                elif ep['method'] in ['POST', 'PUT', 'PATCH']:
                    code.append(f'        response = self.session.{ep["method"].lower()}(url, json=data, params=kwargs)')
                
                code.append('        response.raise_for_status()')
                code.append('        return response.json()')
                code.append('')
            
            code.append('')
        
        return '\n'.join(code)
    
    def _generate_javascript_sdk(self, api: Dict, endpoints: List[Dict]) -> str:
        """Generate JavaScript SDK"""
        class_name = ''.join(word.capitalize() for word in api['name'].split())
        
        code = [
            f'/**',
            f' * {api["name"]} JavaScript SDK',
            f' * {api.get("description", "")}',
            f' */',
            '',
            f'class {class_name}Client {{',
            f'  constructor(apiKey, baseUrl = "{api["base_url"]}") {{',
            '    this.apiKey = apiKey;',
            '    this.baseUrl = baseUrl.replace(/\\/$/, "");',
            '    this.headers = {',
            f'      "{api.get("auth_header", "X-API-Key")}": apiKey,',
            '      "Accept": "application/json",',
            '      "Content-Type": "application/json"',
            '    };',
            '  }',
            '',
            '  async request(method, path, data = null, params = {}) {',
            '    const url = new URL(this.baseUrl + path);',
            '    ',
            '    // Add query parameters',
            '    Object.entries(params).forEach(([key, value]) => {',
            '      if (value !== undefined && value !== null) {',
            '        url.searchParams.append(key, value);',
            '      }',
            '    });',
            '',
            '    const options = {',
            '      method,',
            '      headers: this.headers',
            '    };',
            '',
            '    if (data && ["POST", "PUT", "PATCH"].includes(method)) {',
            '      options.body = JSON.stringify(data);',
            '    }',
            '',
            '    const response = await fetch(url.toString(), options);',
            '    ',
            '    if (!response.ok) {',
            '      throw new Error(`API error: ${response.status} ${response.statusText}`);',
            '    }',
            '',
            '    return response.json();',
            '  }',
            ''
        ]
        
        # Generate methods
        for endpoint in endpoints:
            method_name = self._endpoint_to_method_name(endpoint['path'], endpoint['method'])
            
            # Extract path parameters
            path_params = re.findall(r'{(\w+)}', endpoint['path'])
            
            # Build method signature
            params = path_params.copy()
            if endpoint['method'] in ['POST', 'PUT', 'PATCH']:
                params.append('data')
            params.append('queryParams = {}')
            
            code.append(f'  async {method_name}({", ".join(params)}) {{')
            
            # Build path
            path = endpoint['path']
            for param in path_params:
                path = path.replace(f'{{{param}}}', f'${{{param}}}')
            
            # Make request
            if endpoint['method'] in ['POST', 'PUT', 'PATCH']:
                code.append(f'    return this.request("{endpoint["method"]}", `{path}`, data, queryParams);')
            else:
                code.append(f'    return this.request("{endpoint["method"]}", `{path}`, null, queryParams);')
            
            code.append('  }')
            code.append('')
        
        code.append('}')
        code.append('')
        code.append(f'export default {class_name}Client;')
        
        return '\n'.join(code)
    
    def _endpoint_to_method_name(self, path: str, method: str) -> str:
        """Convert endpoint path to method name"""
        # Remove leading slash and parameters
        parts = path.strip('/').split('/')
        clean_parts = []
        
        for part in parts:
            if '{' not in part:
                clean_parts.append(part)
            else:
                # Extract parameter name
                param = re.search(r'{(\w+)}', part).group(1)
                clean_parts.append('by_' + param)
        
        # Build method name
        method_name = method.lower()
        if method == 'GET':
            if parts[-1].endswith('s'):
                method_name = 'list'
            else:
                method_name = 'get'
        elif method == 'POST':
            method_name = 'create'
        elif method == 'PUT':
            method_name = 'update'
        elif method == 'DELETE':
            method_name = 'delete'
        
        # Combine parts
        name_parts = [method_name] + [p.replace('-', '_') for p in clean_parts]
        return '_'.join(name_parts)
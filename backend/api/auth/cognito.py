"""
AWS Cognito authentication for production use
"""
import os
import jwt
from typing import Optional, Dict, Any
from datetime import datetime
from fastapi import HTTPException, status
import httpx
from jwt.algorithms import RSAAlgorithm
import json

class CognitoAuth:
    def __init__(self):
        self.region = os.getenv("AWS_REGION", "us-east-1")
        self.user_pool_id = os.getenv("COGNITO_USER_POOL_ID")
        self.client_id = os.getenv("COGNITO_CLIENT_ID")
        self.jwks_url = f"https://cognito-idp.{self.region}.amazonaws.com/{self.user_pool_id}/.well-known/jwks.json"
        self.issuer = f"https://cognito-idp.{self.region}.amazonaws.com/{self.user_pool_id}"
        self._jwks_client = None
        self._keys = {}

    async def _get_jwks(self):
        """Fetch JSON Web Key Set from Cognito"""
        if self._jwks_client is None:
            self._jwks_client = httpx.AsyncClient()
        
        response = await self._jwks_client.get(self.jwks_url)
        response.raise_for_status()
        return response.json()

    async def _get_signing_key(self, kid: str):
        """Get the signing key for a given key ID"""
        if kid not in self._keys:
            jwks = await self._get_jwks()
            for key in jwks["keys"]:
                if key["kid"] == kid:
                    self._keys[kid] = RSAAlgorithm.from_jwk(json.dumps(key))
                    break
            else:
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="Unable to find a signing key that matches"
                )
        return self._keys[kid]

    async def verify_token(self, token: str) -> Dict[str, Any]:
        """Verify a Cognito JWT token"""
        try:
            # Decode token header without verification to get the key ID
            unverified_header = jwt.get_unverified_header(token)
            kid = unverified_header["kid"]
            
            # Get the signing key
            signing_key = await self._get_signing_key(kid)
            
            # Verify and decode the token
            payload = jwt.decode(
                token,
                signing_key,
                algorithms=["RS256"],
                audience=self.client_id,
                issuer=self.issuer,
                options={"verify_exp": True}
            )
            
            # Verify token use
            if payload.get("token_use") not in ["id", "access"]:
                raise HTTPException(
                    status_code=status.HTTP_401_UNAUTHORIZED,
                    detail="Invalid token use"
                )
            
            return payload
            
        except jwt.ExpiredSignatureError:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Token has expired"
            )
        except jwt.InvalidTokenError as e:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail=f"Invalid token: {str(e)}"
            )
        except Exception as e:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail=f"Token verification failed: {str(e)}"
            )

    async def get_user_from_token(self, token: str) -> Dict[str, Any]:
        """Extract user information from a verified token"""
        payload = await self.verify_token(token)
        
        # Build user object from token claims
        user = {
            "id": payload.get("sub"),
            "email": payload.get("email"),
            "name": payload.get("name") or payload.get("given_name", "") + " " + payload.get("family_name", ""),
            "username": payload.get("cognito:username", payload.get("email")),
            "groups": payload.get("cognito:groups", []),
            "token_use": payload.get("token_use"),
            "exp": payload.get("exp"),
            "iat": payload.get("iat")
        }
        
        # Clean up empty values
        user = {k: v for k, v in user.items() if v}
        
        return user

    async def close(self):
        """Close the HTTP client"""
        if self._jwks_client:
            await self._jwks_client.aclose()

# Global instance
cognito_auth = None

def get_cognito_auth() -> Optional[CognitoAuth]:
    """Get or create Cognito auth instance"""
    global cognito_auth
    
    # Check if Cognito is configured
    if not os.getenv("COGNITO_USER_POOL_ID") or not os.getenv("COGNITO_CLIENT_ID"):
        return None
    
    if cognito_auth is None:
        cognito_auth = CognitoAuth()
    
    return cognito_auth
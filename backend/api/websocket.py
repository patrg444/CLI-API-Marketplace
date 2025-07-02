"""
WebSocket handler for real-time updates in the Creator Portal
"""

import json
import asyncio
import logging
from typing import Dict, Set
from fastapi import WebSocket, WebSocketDisconnect
from datetime import datetime

logger = logging.getLogger(__name__)

class WebSocketManager:
    def __init__(self):
        # Store active connections by user_id
        self.active_connections: Dict[str, Set[WebSocket]] = {}
        # Store connection metadata
        self.connection_metadata: Dict[WebSocket, dict] = {}
    
    async def connect(self, websocket: WebSocket, user_id: str):
        """Accept a new WebSocket connection"""
        await websocket.accept()
        
        # Initialize user connections if not exists
        if user_id not in self.active_connections:
            self.active_connections[user_id] = set()
        
        # Add connection to user's set
        self.active_connections[user_id].add(websocket)
        
        # Store metadata
        self.connection_metadata[websocket] = {
            'user_id': user_id,
            'connected_at': datetime.utcnow(),
            'page': None
        }
        
        logger.info(f"WebSocket connected for user {user_id}")
        await self.send_personal_message(websocket, {
            'type': 'connection_established',
            'payload': {'status': 'connected'}
        })
    
    def disconnect(self, websocket: WebSocket):
        """Remove a WebSocket connection"""
        if websocket in self.connection_metadata:
            user_id = self.connection_metadata[websocket]['user_id']
            
            # Remove from user's connections
            if user_id in self.active_connections:
                self.active_connections[user_id].discard(websocket)
                
                # Clean up empty user sets
                if not self.active_connections[user_id]:
                    del self.active_connections[user_id]
            
            # Remove metadata
            del self.connection_metadata[websocket]
            logger.info(f"WebSocket disconnected for user {user_id}")
    
    async def send_personal_message(self, websocket: WebSocket, message: dict):
        """Send a message to a specific WebSocket connection"""
        try:
            await websocket.send_text(json.dumps(message))
        except Exception as e:
            logger.error(f"Failed to send message to WebSocket: {e}")
    
    async def send_to_user(self, user_id: str, message: dict):
        """Send a message to all connections for a specific user"""
        if user_id in self.active_connections:
            # Send to all user's connections
            dead_connections = []
            for websocket in self.active_connections[user_id].copy():
                try:
                    await websocket.send_text(json.dumps(message))
                except Exception as e:
                    logger.error(f"Failed to send message to user {user_id}: {e}")
                    dead_connections.append(websocket)
            
            # Clean up dead connections
            for websocket in dead_connections:
                self.disconnect(websocket)
    
    async def broadcast_to_all(self, message: dict):
        """Send a message to all connected clients"""
        for user_id in list(self.active_connections.keys()):
            await self.send_to_user(user_id, message)
    
    async def handle_message(self, websocket: WebSocket, data: dict):
        """Handle incoming WebSocket messages"""
        message_type = data.get('type')
        payload = data.get('payload', {})
        
        if message_type == 'auth':
            # Handle authentication
            token = data.get('token')
            if token:
                # In a real app, validate the JWT token here
                user_id = self.extract_user_from_token(token)
                if user_id:
                    self.connection_metadata[websocket]['user_id'] = user_id
                    await self.send_personal_message(websocket, {
                        'type': 'auth_success',
                        'payload': {'user_id': user_id}
                    })
        
        elif message_type == 'page_change':
            # Track which page the user is currently viewing
            page = payload.get('page')
            if websocket in self.connection_metadata:
                self.connection_metadata[websocket]['page'] = page
        
        elif message_type == 'ping':
            # Respond to ping with pong
            await self.send_personal_message(websocket, {
                'type': 'pong',
                'payload': {'timestamp': datetime.utcnow().isoformat()}
            })
    
    def extract_user_from_token(self, token: str) -> str:
        """Extract user ID from JWT token"""
        # Simplified implementation - in real app, decode and validate JWT
        try:
            # For demo purposes, return a mock user ID
            return "user_123"
        except Exception:
            return None
    
    async def notify_api_status_change(self, user_id: str, api_id: str, api_name: str, status: str, endpoint_url: str = None):
        """Notify user about API status changes"""
        message = {
            'type': 'api_status_update',
            'payload': {
                'api_id': api_id,
                'api_name': api_name,
                'status': status,
                'endpoint_url': endpoint_url,
                'timestamp': datetime.utcnow().isoformat()
            }
        }
        await self.send_to_user(user_id, message)
    
    async def notify_analytics_update(self, user_id: str, analytics_data: dict):
        """Notify user about analytics updates"""
        message = {
            'type': 'analytics_update',
            'payload': analytics_data
        }
        await self.send_to_user(user_id, message)
    
    async def notify_billing_update(self, user_id: str, billing_data: dict):
        """Notify user about billing/revenue updates"""
        message = {
            'type': 'billing_update',
            'payload': billing_data
        }
        await self.send_to_user(user_id, message)
    
    async def notify_transaction_created(self, user_id: str, transaction: dict):
        """Notify user about new transactions"""
        message = {
            'type': 'transaction_created',
            'payload': {
                'transaction': transaction,
                'timestamp': datetime.utcnow().isoformat()
            }
        }
        await self.send_to_user(user_id, message)
    
    async def notify_payout_completed(self, user_id: str, amount: float, payout_id: str):
        """Notify user about completed payouts"""
        message = {
            'type': 'payout_completed',
            'payload': {
                'amount': amount,
                'payout_id': payout_id,
                'timestamp': datetime.utcnow().isoformat()
            }
        }
        await self.send_to_user(user_id, message)
    
    async def notify_api_published(self, user_id: str, publication_data: dict):
        """Notify user about API marketplace publication"""
        message = {
            'type': 'api_published',
            'payload': publication_data
        }
        await self.send_to_user(user_id, message)
    
    def get_connection_count(self) -> int:
        """Get total number of active connections"""
        return sum(len(connections) for connections in self.active_connections.values())
    
    def get_user_count(self) -> int:
        """Get number of unique connected users"""
        return len(self.active_connections)

# Global WebSocket manager instance
websocket_manager = WebSocketManager()

async def websocket_endpoint(websocket: WebSocket, user_id: str = "anonymous"):
    """Main WebSocket endpoint handler"""
    await websocket_manager.connect(websocket, user_id)
    
    try:
        while True:
            # Receive message from client
            data = await websocket.receive_text()
            try:
                message = json.loads(data)
                await websocket_manager.handle_message(websocket, message)
            except json.JSONDecodeError:
                await websocket_manager.send_personal_message(websocket, {
                    'type': 'error',
                    'payload': {'message': 'Invalid JSON format'}
                })
    
    except WebSocketDisconnect:
        websocket_manager.disconnect(websocket)
    except Exception as e:
        logger.error(f"WebSocket error: {e}")
        websocket_manager.disconnect(websocket)
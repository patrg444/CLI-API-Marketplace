"""
Simple WebSocket tests that work with actual implementation
"""

import pytest
import asyncio
import json
from unittest.mock import Mock, AsyncMock
from fastapi import WebSocket

import sys
import os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from websocket import WebSocketManager


class TestWebSocketManager:
    """Test WebSocket manager functionality"""
    
    @pytest.fixture
    def manager(self):
        return WebSocketManager()
    
    @pytest.fixture
    def mock_websocket(self):
        """Create a mock WebSocket"""
        ws = Mock(spec=WebSocket)
        ws.accept = AsyncMock()
        ws.send_text = AsyncMock()
        ws.send_json = AsyncMock()
        ws.close = AsyncMock()
        return ws
    
    @pytest.mark.asyncio
    async def test_connect_user(self, manager, mock_websocket):
        """Test connecting a user"""
        user_id = "user-123"
        
        # Connect user
        await manager.connect(mock_websocket, user_id)
        
        # Verify WebSocket was accepted
        mock_websocket.accept.assert_called_once()
        
        # Verify user is in active connections
        assert user_id in manager.active_connections
        assert mock_websocket in manager.active_connections[user_id]
        
        # Verify metadata stored
        assert mock_websocket in manager.connection_metadata
        assert manager.connection_metadata[mock_websocket]['user_id'] == user_id
        
        # Verify welcome message sent via send_text
        mock_websocket.send_text.assert_called()
    
    def test_disconnect_user(self, manager, mock_websocket):
        """Test disconnecting a user"""
        user_id = "user-123"
        
        # Setup - manually add connection
        manager.active_connections[user_id] = {mock_websocket}
        manager.connection_metadata[mock_websocket] = {'user_id': user_id}
        
        # Disconnect
        manager.disconnect(mock_websocket)
        
        # Verify connection removed
        assert user_id not in manager.active_connections or mock_websocket not in manager.active_connections.get(user_id, set())
        assert mock_websocket not in manager.connection_metadata
    
    @pytest.mark.asyncio
    async def test_send_personal_message(self, manager, mock_websocket):
        """Test sending personal message to user"""
        message = {"type": "test", "data": "hello"}
        
        await manager.send_personal_message(mock_websocket, message)
        
        # Verify message sent via send_text with JSON string
        mock_websocket.send_text.assert_called_once_with(json.dumps(message))
    
    @pytest.mark.asyncio
    async def test_broadcast_to_user(self, manager):
        """Test broadcasting to all connections of a user"""
        user_id = "user-123"
        
        # Create multiple connections for same user
        ws1 = Mock(spec=WebSocket)
        ws1.send_text = AsyncMock()
        ws2 = Mock(spec=WebSocket)
        ws2.send_text = AsyncMock()
        
        manager.active_connections[user_id] = {ws1, ws2}
        
        # Broadcast message
        message = {"type": "broadcast", "data": "test"}
        await manager.send_to_user(user_id, message)  # Method is send_to_user, not broadcast_to_user
        
        # Verify both connections received message via send_text
        expected_text = json.dumps(message)
        ws1.send_text.assert_called_once_with(expected_text)
        ws2.send_text.assert_called_once_with(expected_text)
    
    @pytest.mark.asyncio
    async def test_broadcast_to_all(self, manager):
        """Test broadcasting to all connected users"""
        # Setup multiple users with connections
        ws1 = Mock(spec=WebSocket)
        ws1.send_text = AsyncMock()
        ws2 = Mock(spec=WebSocket)
        ws2.send_text = AsyncMock()
        ws3 = Mock(spec=WebSocket)
        ws3.send_text = AsyncMock()
        
        manager.active_connections = {
            "user-1": {ws1},
            "user-2": {ws2, ws3}
        }
        
        # Broadcast
        message = {"type": "announcement", "data": "system update"}
        await manager.broadcast_to_all(message)
        
        # Verify all connections received message via send_text
        expected_text = json.dumps(message)
        ws1.send_text.assert_called_once_with(expected_text)
        ws2.send_text.assert_called_once_with(expected_text)
        ws3.send_text.assert_called_once_with(expected_text)


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
"""
WebSocket functionality tests for real-time features
Tests connection lifecycle, authentication, message broadcasting, and error handling
"""

import pytest
import asyncio
import json
from unittest.mock import Mock, patch, AsyncMock
from fastapi.testclient import TestClient
from fastapi import WebSocket
import websockets

import sys
import os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
from websocket import WebSocketManager


class TestWebSocketManager:
    """Test WebSocket connection management"""
    
    @pytest.fixture
    def manager(self):
        return WebSocketManager()
    
    @pytest.fixture
    def mock_websocket(self):
        """Create a mock WebSocket connection"""
        ws = Mock(spec=WebSocket)
        ws.accept = AsyncMock()
        ws.send_text = AsyncMock()
        ws.send_json = AsyncMock()
        ws.receive_text = AsyncMock()
        ws.receive_json = AsyncMock()
        ws.close = AsyncMock()
        return ws
    
    @pytest.mark.asyncio
    async def test_connect_authenticated_user(self, manager, mock_websocket):
        """Test connecting an authenticated user"""
        user_id = "user-123"
        token = "valid-jwt-token"
        
        # Mock token validation
        with patch('websocket.validate_token', return_value={'sub': user_id}):
            await manager.connect(mock_websocket, token)
            
            # Verify connection accepted
            mock_websocket.accept.assert_called_once()
            
            # Verify user added to active connections
            assert user_id in manager.active_connections
            assert mock_websocket in manager.active_connections[user_id]
            
            # Verify welcome message sent
            mock_websocket.send_json.assert_called_with({
                'type': 'connection',
                'status': 'connected',
                'user_id': user_id
            })
    
    @pytest.mark.asyncio
    async def test_connect_invalid_token(self, manager, mock_websocket):
        """Test connection rejection with invalid token"""
        token = "invalid-token"
        
        with patch('backend.api.websocket.validate_token', side_effect=Exception("Invalid token")):
            await manager.connect(mock_websocket, token)
            
            # Should close connection with error
            mock_websocket.close.assert_called_once_with(code=1008, reason="Invalid authentication")
            
            # Should not be in active connections
            assert len(manager.active_connections) == 0
    
    @pytest.mark.asyncio
    async def test_disconnect_user(self, manager, mock_websocket):
        """Test disconnecting a user"""
        user_id = "user-123"
        
        # First connect
        manager.active_connections[user_id] = [mock_websocket]
        
        # Then disconnect
        await manager.disconnect(user_id, mock_websocket)
        
        # Verify removed from connections
        assert user_id not in manager.active_connections
        
        # Verify close called
        mock_websocket.close.assert_called_once()
    
    @pytest.mark.asyncio
    async def test_broadcast_to_user(self, manager, mock_websocket):
        """Test broadcasting message to specific user"""
        user_id = "user-123"
        message = {"type": "notification", "content": "Test message"}
        
        # Setup connection
        manager.active_connections[user_id] = [mock_websocket]
        
        # Broadcast
        await manager.send_to_user(user_id, message)
        
        # Verify message sent
        mock_websocket.send_json.assert_called_once_with(message)
    
    @pytest.mark.asyncio
    async def test_broadcast_to_all(self, manager):
        """Test broadcasting to all connected users"""
        # Setup multiple connections
        ws1, ws2, ws3 = Mock(), Mock(), Mock()
        for ws in [ws1, ws2, ws3]:
            ws.send_json = AsyncMock()
        
        manager.active_connections = {
            "user-1": [ws1],
            "user-2": [ws2, ws3],  # User with multiple connections
        }
        
        message = {"type": "announcement", "content": "System update"}
        
        # Broadcast
        await manager.broadcast(message)
        
        # Verify all connections received message
        ws1.send_json.assert_called_once_with(message)
        ws2.send_json.assert_called_once_with(message)
        ws3.send_json.assert_called_once_with(message)
    
    @pytest.mark.asyncio
    async def test_handle_connection_error(self, manager, mock_websocket):
        """Test handling connection errors gracefully"""
        user_id = "user-123"
        manager.active_connections[user_id] = [mock_websocket]
        
        # Simulate send failure
        mock_websocket.send_json.side_effect = Exception("Connection lost")
        
        # Should handle error gracefully
        await manager.send_to_user(user_id, {"test": "message"})
        
        # Connection should be removed
        assert user_id not in manager.active_connections
    
    @pytest.mark.asyncio
    async def test_connection_limit_per_user(self, manager):
        """Test connection limit per user"""
        user_id = "user-123"
        max_connections = 5
        
        # Create multiple connections
        connections = []
        for i in range(max_connections + 1):
            ws = Mock()
            ws.accept = AsyncMock()
            ws.send_json = AsyncMock()
            ws.close = AsyncMock()
            connections.append(ws)
        
        # Add connections up to limit
        with patch('websocket.validate_token', return_value={'sub': user_id}):
            for i in range(max_connections):
                await manager.connect(connections[i], "valid-token")
            
            # Verify all connected
            assert len(manager.active_connections[user_id]) == max_connections
            
            # Try to exceed limit
            await manager.connect(connections[max_connections], "valid-token")
            
            # Should reject the extra connection
            connections[max_connections].close.assert_called_once()
            assert len(manager.active_connections[user_id]) == max_connections


class TestWebSocketMessages:
    """Test WebSocket message handling"""
    
    @pytest.fixture
    def connection_manager(self):
        return ConnectionManager()
    
    @pytest.mark.asyncio
    async def test_handle_ping_pong(self, connection_manager, mock_websocket):
        """Test ping/pong heartbeat"""
        user_id = "user-123"
        
        # Receive ping
        mock_websocket.receive_json.return_value = {"type": "ping"}
        
        await connection_manager.handle_message(user_id, mock_websocket)
        
        # Should respond with pong
        mock_websocket.send_json.assert_called_with({
            "type": "pong",
            "timestamp": pytest.approx(asyncio.get_event_loop().time(), rel=1)
        })
    
    @pytest.mark.asyncio
    async def test_handle_subscribe_to_api(self, connection_manager, mock_websocket):
        """Test subscribing to API updates"""
        user_id = "user-123"
        api_id = "weather-api"
        
        # Send subscribe message
        mock_websocket.receive_json.return_value = {
            "type": "subscribe",
            "api_id": api_id
        }
        
        await connection_manager.handle_message(user_id, mock_websocket)
        
        # Verify subscription added
        assert api_id in connection_manager.subscriptions
        assert user_id in connection_manager.subscriptions[api_id]
        
        # Should confirm subscription
        mock_websocket.send_json.assert_called_with({
            "type": "subscribed",
            "api_id": api_id,
            "status": "success"
        })
    
    @pytest.mark.asyncio
    async def test_handle_unsubscribe(self, connection_manager, mock_websocket):
        """Test unsubscribing from API updates"""
        user_id = "user-123"
        api_id = "weather-api"
        
        # Setup existing subscription
        connection_manager.subscriptions[api_id] = {user_id}
        
        # Send unsubscribe message
        mock_websocket.receive_json.return_value = {
            "type": "unsubscribe",
            "api_id": api_id
        }
        
        await connection_manager.handle_message(user_id, mock_websocket)
        
        # Verify unsubscribed
        assert user_id not in connection_manager.subscriptions.get(api_id, set())
    
    @pytest.mark.asyncio
    async def test_api_status_update_broadcast(self, connection_manager):
        """Test broadcasting API status updates to subscribers"""
        api_id = "weather-api"
        
        # Setup subscribers
        ws1, ws2, ws3 = Mock(), Mock(), Mock()
        for ws in [ws1, ws2, ws3]:
            ws.send_json = AsyncMock()
        
        connection_manager.subscriptions[api_id] = {"user-1", "user-2"}
        connection_manager.active_connections = {
            "user-1": [ws1],
            "user-2": [ws2, ws3],
        }
        
        # Broadcast API update
        update = {
            "type": "api_update",
            "api_id": api_id,
            "status": "maintenance",
            "message": "API under maintenance"
        }
        
        await connection_manager.broadcast_to_api_subscribers(api_id, update)
        
        # Verify only subscribers received update
        ws1.send_json.assert_called_once_with(update)
        ws2.send_json.assert_called_once_with(update)
        ws3.send_json.assert_called_once_with(update)
    
    @pytest.mark.asyncio
    async def test_rate_limit_messages(self, connection_manager, mock_websocket):
        """Test rate limiting on message sending"""
        user_id = "user-123"
        
        # Send many messages quickly
        for i in range(100):
            mock_websocket.receive_json.return_value = {
                "type": "message",
                "content": f"Message {i}"
            }
            
            result = await connection_manager.handle_message(user_id, mock_websocket)
            
            if i < 50:  # Assuming 50 messages per minute limit
                assert result is True
            else:
                # Should be rate limited
                mock_websocket.send_json.assert_called_with({
                    "type": "error",
                    "code": "RATE_LIMIT",
                    "message": "Too many messages. Please slow down."
                })
                break


class TestWebSocketReconnection:
    """Test WebSocket reconnection handling"""
    
    @pytest.mark.asyncio
    async def test_reconnect_with_session_id(self):
        """Test reconnection using session ID"""
        manager = WebSocketManager()
        user_id = "user-123"
        session_id = "session-456"
        
        # First connection
        ws1 = Mock()
        ws1.accept = AsyncMock()
        ws1.send_json = AsyncMock()
        
        with patch('websocket.validate_token', return_value={'sub': user_id}):
            await manager.connect(ws1, "token", session_id=session_id)
        
        # Simulate disconnect
        await manager.disconnect(user_id, ws1)
        
        # Reconnect with same session
        ws2 = Mock()
        ws2.accept = AsyncMock()
        ws2.send_json = AsyncMock()
        
        with patch('websocket.validate_token', return_value={'sub': user_id}):
            await manager.connect(ws2, "token", session_id=session_id)
        
        # Should restore session state
        ws2.send_json.assert_any_call({
            "type": "reconnected",
            "session_id": session_id,
            "missed_messages": []  # In real implementation, would include missed messages
        })
    
    @pytest.mark.asyncio
    async def test_exponential_backoff_on_errors(self):
        """Test exponential backoff for reconnection attempts"""
        backoff_times = []
        
        async def mock_connect_with_failure():
            raise Exception("Connection failed")
        
        # Track backoff times
        original_sleep = asyncio.sleep
        async def track_sleep(seconds):
            backoff_times.append(seconds)
            if len(backoff_times) < 3:  # Fail first 3 attempts
                raise Exception("Still failing")
            return await original_sleep(0)  # Don't actually sleep in tests
        
        with patch('asyncio.sleep', side_effect=track_sleep):
            reconnect_handler = ReconnectHandler()
            
            try:
                await reconnect_handler.connect_with_retry(mock_connect_with_failure)
            except:
                pass
        
        # Verify exponential backoff
        assert len(backoff_times) >= 3
        assert backoff_times[0] < backoff_times[1] < backoff_times[2]
        assert backoff_times[1] / backoff_times[0] >= 1.5  # Exponential growth


class TestWebSocketSecurity:
    """Test WebSocket security features"""
    
    @pytest.mark.asyncio
    async def test_message_size_limit(self):
        """Test message size limits to prevent DoS"""
        manager = WebSocketManager()
        ws = Mock()
        ws.receive_text = AsyncMock()
        
        # Large message
        large_message = "x" * (1024 * 1024 + 1)  # 1MB + 1 byte
        ws.receive_text.return_value = large_message
        
        with pytest.raises(ValueError, match="Message too large"):
            await manager.receive_message(ws)
    
    @pytest.mark.asyncio
    async def test_origin_validation(self):
        """Test WebSocket origin validation"""
        manager = WebSocketManager()
        
        # Valid origins
        valid_origins = [
            "https://apidirect.dev",
            "https://www.apidirect.dev",
            "http://localhost:3000"  # Development
        ]
        
        # Invalid origins
        invalid_origins = [
            "https://evil-site.com",
            "http://phishing-site.net"
        ]
        
        for origin in valid_origins:
            assert manager.validate_origin(origin) is True
        
        for origin in invalid_origins:
            assert manager.validate_origin(origin) is False
    
    @pytest.mark.asyncio
    async def test_token_refresh_during_connection(self):
        """Test handling token refresh while connected"""
        manager = WebSocketManager()
        user_id = "user-123"
        ws = Mock()
        ws.send_json = AsyncMock()
        
        # Connect with token about to expire
        with patch('backend.api.websocket.validate_token', return_value={
            'sub': user_id,
            'exp': asyncio.get_event_loop().time() + 60  # Expires in 1 minute
        }):
            await manager.connect(ws, "token")
        
        # Should request token refresh
        ws.send_json.assert_any_call({
            "type": "token_refresh_required",
            "expires_in": pytest.approx(60, rel=1)
        })


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
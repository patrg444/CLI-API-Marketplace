"""
API Version Management System

Handles semantic versioning, version comparison, deprecation,
and version-specific routing for APIs.
"""

import re
from typing import Dict, List, Optional, Tuple, Any
from datetime import datetime, timedelta
from uuid import UUID
import asyncpg
from enum import Enum


class VersionType(Enum):
    MAJOR = "major"  # Breaking changes
    MINOR = "minor"  # New features, backward compatible
    PATCH = "patch"  # Bug fixes


class VersionStatus(Enum):
    DRAFT = "draft"
    ACTIVE = "active"
    DEPRECATED = "deprecated"
    RETIRED = "retired"


class VersionManager:
    def __init__(self, db_pool: asyncpg.Pool):
        self.db_pool = db_pool
        self.version_pattern = re.compile(r'^(\d+)\.(\d+)\.(\d+)(?:-(.+))?$')
    
    async def create_version_table(self):
        """Create API versions table if it doesn't exist."""
        async with self.db_pool.acquire() as conn:
            await conn.execute('''
                CREATE TABLE IF NOT EXISTS api_versions (
                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                    api_id UUID NOT NULL REFERENCES apis(id) ON DELETE CASCADE,
                    version VARCHAR(20) NOT NULL,
                    major INTEGER NOT NULL,
                    minor INTEGER NOT NULL,
                    patch INTEGER NOT NULL,
                    prerelease VARCHAR(50),
                    status VARCHAR(20) NOT NULL DEFAULT 'draft',
                    release_notes TEXT,
                    breaking_changes TEXT[],
                    deprecated_features TEXT[],
                    minimum_supported_version VARCHAR(20),
                    deprecation_date TIMESTAMP,
                    retirement_date TIMESTAMP,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                    created_by UUID REFERENCES users(id),
                    published_at TIMESTAMP,
                    UNIQUE(api_id, version)
                );
                
                CREATE INDEX IF NOT EXISTS idx_api_versions_api_id ON api_versions(api_id);
                CREATE INDEX IF NOT EXISTS idx_api_versions_status ON api_versions(status);
                CREATE INDEX IF NOT EXISTS idx_api_versions_created_at ON api_versions(created_at DESC);
            ''')
    
    def parse_version(self, version: str) -> Tuple[int, int, int, Optional[str]]:
        """Parse semantic version string into components."""
        match = self.version_pattern.match(version)
        if not match:
            raise ValueError(f"Invalid version format: {version}. Use semantic versioning (e.g., 1.0.0)")
        
        major, minor, patch, prerelease = match.groups()
        return int(major), int(minor), int(patch), prerelease
    
    def compare_versions(self, v1: str, v2: str) -> int:
        """
        Compare two version strings.
        Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
        """
        v1_parts = self.parse_version(v1)
        v2_parts = self.parse_version(v2)
        
        # Compare major, minor, patch
        for i in range(3):
            if v1_parts[i] < v2_parts[i]:
                return -1
            elif v1_parts[i] > v2_parts[i]:
                return 1
        
        # Compare prerelease
        if v1_parts[3] is None and v2_parts[3] is None:
            return 0
        elif v1_parts[3] is None:  # v1 is release, v2 is prerelease
            return 1
        elif v2_parts[3] is None:  # v1 is prerelease, v2 is release
            return -1
        else:  # Both are prereleases
            return -1 if v1_parts[3] < v2_parts[3] else (1 if v1_parts[3] > v2_parts[3] else 0)
    
    async def create_version(
        self,
        api_id: str,
        version_type: VersionType,
        user_id: str,
        release_notes: Optional[str] = None,
        breaking_changes: Optional[List[str]] = None,
        deprecated_features: Optional[List[str]] = None,
        prerelease: Optional[str] = None
    ) -> Dict[str, Any]:
        """Create a new version for an API."""
        async with self.db_pool.acquire() as conn:
            # Get current version
            current = await conn.fetchrow('''
                SELECT version, major, minor, patch
                FROM api_versions
                WHERE api_id = $1 AND status != 'draft'
                ORDER BY major DESC, minor DESC, patch DESC
                LIMIT 1
            ''', UUID(api_id))
            
            if current:
                major, minor, patch = current['major'], current['minor'], current['patch']
            else:
                # First version
                major, minor, patch = 0, 0, 0
            
            # Increment version based on type
            if version_type == VersionType.MAJOR:
                major += 1
                minor = 0
                patch = 0
            elif version_type == VersionType.MINOR:
                minor += 1
                patch = 0
            else:  # PATCH
                patch += 1
            
            # Build version string
            version = f"{major}.{minor}.{patch}"
            if prerelease:
                version += f"-{prerelease}"
            
            # Create version record
            result = await conn.fetchrow('''
                INSERT INTO api_versions (
                    api_id, version, major, minor, patch, prerelease,
                    status, release_notes, breaking_changes, deprecated_features,
                    created_by
                ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
                RETURNING *
            ''', UUID(api_id), version, major, minor, patch, prerelease,
                VersionStatus.DRAFT.value, release_notes, breaking_changes or [],
                deprecated_features or [], UUID(user_id))
            
            return dict(result)
    
    async def publish_version(
        self,
        api_id: str,
        version: str,
        deprecation_period_days: int = 90
    ) -> Dict[str, Any]:
        """Publish a draft version, making it active."""
        async with self.db_pool.acquire() as conn:
            async with conn.transaction():
                # Update current active versions to deprecated
                await conn.execute('''
                    UPDATE api_versions
                    SET status = $1,
                        deprecation_date = CURRENT_TIMESTAMP,
                        retirement_date = CURRENT_TIMESTAMP + INTERVAL '%s days'
                    WHERE api_id = $2 AND status = $3
                ''', VersionStatus.DEPRECATED.value, deprecation_period_days * 2,
                    UUID(api_id), VersionStatus.ACTIVE.value)
                
                # Publish the new version
                result = await conn.fetchrow('''
                    UPDATE api_versions
                    SET status = $1,
                        published_at = CURRENT_TIMESTAMP
                    WHERE api_id = $2 AND version = $3 AND status = $4
                    RETURNING *
                ''', VersionStatus.ACTIVE.value, UUID(api_id), version,
                    VersionStatus.DRAFT.value)
                
                if not result:
                    raise ValueError(f"Version {version} not found or not in draft status")
                
                # Update API table with new version
                await conn.execute('''
                    UPDATE apis
                    SET version = $1,
                        updated_at = CURRENT_TIMESTAMP
                    WHERE id = $2
                ''', version, UUID(api_id))
                
                return dict(result)
    
    async def get_versions(
        self,
        api_id: str,
        include_drafts: bool = False,
        limit: int = 20,
        offset: int = 0
    ) -> List[Dict[str, Any]]:
        """Get all versions for an API."""
        async with self.db_pool.acquire() as conn:
            query = '''
                SELECT v.*, u.username as created_by_username
                FROM api_versions v
                LEFT JOIN users u ON v.created_by = u.id
                WHERE v.api_id = $1
            '''
            
            params = [UUID(api_id)]
            if not include_drafts:
                query += ' AND v.status != $2'
                params.append(VersionStatus.DRAFT.value)
            
            query += ' ORDER BY v.major DESC, v.minor DESC, v.patch DESC'
            query += f' LIMIT {limit} OFFSET {offset}'
            
            rows = await conn.fetch(query, *params)
            return [dict(row) for row in rows]
    
    async def get_version_changelog(
        self,
        api_id: str,
        from_version: Optional[str] = None,
        to_version: Optional[str] = None
    ) -> Dict[str, Any]:
        """Get changelog between versions."""
        async with self.db_pool.acquire() as conn:
            query = '''
                SELECT version, release_notes, breaking_changes,
                       deprecated_features, published_at
                FROM api_versions
                WHERE api_id = $1 AND status != 'draft'
            '''
            params = [UUID(api_id)]
            
            if from_version and to_version:
                # Get versions between from and to
                query += ' AND (major, minor, patch) > (SELECT major, minor, patch FROM api_versions WHERE api_id = $1 AND version = $2)'
                query += ' AND (major, minor, patch) <= (SELECT major, minor, patch FROM api_versions WHERE api_id = $1 AND version = $3)'
                params.extend([from_version, to_version])
            elif from_version:
                # Get all versions after from_version
                query += ' AND (major, minor, patch) > (SELECT major, minor, patch FROM api_versions WHERE api_id = $1 AND version = $2)'
                params.append(from_version)
            
            query += ' ORDER BY major DESC, minor DESC, patch DESC'
            
            rows = await conn.fetch(query, *params)
            
            changelog = {
                'versions': [],
                'all_breaking_changes': [],
                'all_deprecated_features': []
            }
            
            for row in rows:
                version_info = {
                    'version': row['version'],
                    'published_at': row['published_at'].isoformat() if row['published_at'] else None,
                    'release_notes': row['release_notes'],
                    'breaking_changes': row['breaking_changes'],
                    'deprecated_features': row['deprecated_features']
                }
                changelog['versions'].append(version_info)
                changelog['all_breaking_changes'].extend(row['breaking_changes'] or [])
                changelog['all_deprecated_features'].extend(row['deprecated_features'] or [])
            
            return changelog
    
    async def check_version_compatibility(
        self,
        api_id: str,
        client_version: str,
        required_version: Optional[str] = None
    ) -> Dict[str, Any]:
        """Check if a client version is compatible with the API."""
        async with self.db_pool.acquire() as conn:
            # Get current active version
            active = await conn.fetchrow('''
                SELECT version, minimum_supported_version
                FROM api_versions
                WHERE api_id = $1 AND status = $2
                ORDER BY major DESC, minor DESC, patch DESC
                LIMIT 1
            ''', UUID(api_id), VersionStatus.ACTIVE.value)
            
            if not active:
                return {
                    'compatible': False,
                    'reason': 'No active version found'
                }
            
            current_version = active['version']
            min_supported = active['minimum_supported_version'] or '1.0.0'
            
            # Check if client version is within supported range
            if self.compare_versions(client_version, min_supported) < 0:
                return {
                    'compatible': False,
                    'reason': f'Client version {client_version} is below minimum supported version {min_supported}',
                    'minimum_version': min_supported,
                    'current_version': current_version,
                    'upgrade_required': True
                }
            
            # Check if client version is ahead of current
            if self.compare_versions(client_version, current_version) > 0:
                return {
                    'compatible': False,
                    'reason': f'Client version {client_version} is ahead of current version {current_version}',
                    'current_version': current_version
                }
            
            # Check deprecation status
            client_version_info = await conn.fetchrow('''
                SELECT status, deprecation_date, retirement_date
                FROM api_versions
                WHERE api_id = $1 AND version = $2
            ''', UUID(api_id), client_version)
            
            if client_version_info:
                if client_version_info['status'] == VersionStatus.DEPRECATED.value:
                    return {
                        'compatible': True,
                        'warning': 'This version is deprecated',
                        'deprecation_date': client_version_info['deprecation_date'].isoformat(),
                        'retirement_date': client_version_info['retirement_date'].isoformat() if client_version_info['retirement_date'] else None,
                        'current_version': current_version
                    }
                elif client_version_info['status'] == VersionStatus.RETIRED.value:
                    return {
                        'compatible': False,
                        'reason': 'This version has been retired',
                        'current_version': current_version,
                        'upgrade_required': True
                    }
            
            return {
                'compatible': True,
                'current_version': current_version,
                'client_version': client_version
            }
    
    async def auto_retire_versions(self) -> List[str]:
        """Automatically retire versions past their retirement date."""
        async with self.db_pool.acquire() as conn:
            rows = await conn.fetch('''
                UPDATE api_versions
                SET status = $1
                WHERE status = $2 
                AND retirement_date IS NOT NULL 
                AND retirement_date < CURRENT_TIMESTAMP
                RETURNING api_id, version
            ''', VersionStatus.RETIRED.value, VersionStatus.DEPRECATED.value)
            
            return [f"{row['api_id']}:{row['version']}" for row in rows]
    
    async def generate_version_diff(
        self,
        api_id: str,
        version1: str,
        version2: str
    ) -> Dict[str, Any]:
        """Generate a diff between two versions."""
        async with self.db_pool.acquire() as conn:
            # Get both versions
            v1 = await conn.fetchrow('''
                SELECT * FROM api_versions
                WHERE api_id = $1 AND version = $2
            ''', UUID(api_id), version1)
            
            v2 = await conn.fetchrow('''
                SELECT * FROM api_versions
                WHERE api_id = $1 AND version = $2
            ''', UUID(api_id), version2)
            
            if not v1 or not v2:
                raise ValueError("One or both versions not found")
            
            # Determine which is newer
            comparison = self.compare_versions(version1, version2)
            if comparison < 0:
                old_version, new_version = v1, v2
            else:
                old_version, new_version = v2, v1
            
            # Get all versions between them
            changelog = await self.get_version_changelog(
                api_id,
                old_version['version'],
                new_version['version']
            )
            
            return {
                'old_version': old_version['version'],
                'new_version': new_version['version'],
                'version_jump': {
                    'major': new_version['major'] - old_version['major'],
                    'minor': new_version['minor'] - old_version['minor'],
                    'patch': new_version['patch'] - old_version['patch']
                },
                'changelog': changelog,
                'is_breaking_change': new_version['major'] > old_version['major']
            }
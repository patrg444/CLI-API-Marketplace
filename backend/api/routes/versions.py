from fastapi import APIRouter, Depends, HTTPException, Query
from sqlalchemy.orm import Session
from typing import List, Optional
from datetime import datetime
from pydantic import BaseModel

from backend.api.auth.auth import get_current_user, require_api_owner
from backend.api.database import get_db, User, API, APIVersion

router = APIRouter()

# Pydantic models for request/response
class VersionCreate(BaseModel):
    version_number: str
    version_type: str = "draft"  # draft, beta, stable
    release_notes: Optional[str] = None
    breaking_changes: Optional[List[str]] = []
    
class VersionUpdate(BaseModel):
    version_type: Optional[str] = None
    release_notes: Optional[str] = None
    breaking_changes: Optional[List[str]] = None
    
class VersionRollback(BaseModel):
    reason: str
    notify_users: bool = True

class VersionResponse(BaseModel):
    id: str
    api_id: str
    version_number: str
    version_type: str
    release_notes: Optional[str]
    breaking_changes: List[str]
    created_at: datetime
    published_at: Optional[datetime]
    deprecated_at: Optional[datetime]
    is_active: bool
    usage_stats: dict
    
    class Config:
        orm_mode = True

class VersionComparison(BaseModel):
    version_a: str
    version_b: str
    endpoints_added: List[dict]
    endpoints_removed: List[dict]
    endpoints_modified: List[dict]
    breaking_changes: List[str]

@router.get("/apis/{api_id}/versions", response_model=List[VersionResponse])
async def list_api_versions(
    api_id: str,
    version_type: Optional[str] = None,
    include_deprecated: bool = False,
    current_user: User = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """List all versions for an API"""
    # Check if user has access to this API
    api = db.query(API).filter(API.id == api_id).first()
    if not api:
        raise HTTPException(status_code=404, detail="API not found")
    
    # Allow read access for public APIs or owner
    if not api.is_public and api.user_id != current_user.id:
        raise HTTPException(status_code=403, detail="Access denied")
    
    # Query versions
    query = db.query(APIVersion).filter(APIVersion.api_id == api_id)
    
    if version_type:
        query = query.filter(APIVersion.version_type == version_type)
    
    if not include_deprecated:
        query = query.filter(APIVersion.deprecated_at.is_(None))
    
    versions = query.order_by(APIVersion.created_at.desc()).all()
    
    # Enrich with usage stats
    for version in versions:
        version.usage_stats = {
            "active_users": 1247 if version.is_active else 892,  # Mock data
            "api_calls_24h": 52300 if version.is_active else 12400,
            "error_rate": 0.02 if version.is_active else 0.05,
            "avg_response_time": 234 if version.is_active else 312
        }
    
    return versions

@router.get("/apis/{api_id}/versions/current", response_model=VersionResponse)
async def get_current_version(
    api_id: str,
    current_user: User = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """Get the current active version of an API"""
    api = db.query(API).filter(API.id == api_id).first()
    if not api:
        raise HTTPException(status_code=404, detail="API not found")
    
    version = db.query(APIVersion).filter(
        APIVersion.api_id == api_id,
        APIVersion.is_active == True
    ).first()
    
    if not version:
        raise HTTPException(status_code=404, detail="No active version found")
    
    version.usage_stats = {
        "active_users": 1247,
        "api_calls_24h": 52300,
        "error_rate": 0.02,
        "avg_response_time": 234
    }
    
    return version

@router.post("/apis/{api_id}/versions", response_model=VersionResponse)
async def create_version(
    api_id: str,
    version_data: VersionCreate,
    current_user: User = Depends(require_api_owner),
    db: Session = Depends(get_db)
):
    """Create a new version for an API"""
    # Validate version number format (semantic versioning)
    import re
    if not re.match(r'^\d+\.\d+\.\d+(-[\w\.]+)?$', version_data.version_number):
        raise HTTPException(
            status_code=400, 
            detail="Invalid version number format. Use semantic versioning (e.g., 2.1.0)"
        )
    
    # Check if version already exists
    existing = db.query(APIVersion).filter(
        APIVersion.api_id == api_id,
        APIVersion.version_number == version_data.version_number
    ).first()
    
    if existing:
        raise HTTPException(
            status_code=400,
            detail=f"Version {version_data.version_number} already exists"
        )
    
    # Create new version
    new_version = APIVersion(
        api_id=api_id,
        version_number=version_data.version_number,
        version_type=version_data.version_type,
        release_notes=version_data.release_notes,
        breaking_changes=version_data.breaking_changes or [],
        created_at=datetime.utcnow(),
        is_active=False  # New versions start as inactive
    )
    
    db.add(new_version)
    db.commit()
    db.refresh(new_version)
    
    new_version.usage_stats = {
        "active_users": 0,
        "api_calls_24h": 0,
        "error_rate": 0,
        "avg_response_time": 0
    }
    
    return new_version

@router.put("/apis/{api_id}/versions/{version_id}", response_model=VersionResponse)
async def update_version(
    api_id: str,
    version_id: str,
    version_update: VersionUpdate,
    current_user: User = Depends(require_api_owner),
    db: Session = Depends(get_db)
):
    """Update version details"""
    version = db.query(APIVersion).filter(
        APIVersion.id == version_id,
        APIVersion.api_id == api_id
    ).first()
    
    if not version:
        raise HTTPException(status_code=404, detail="Version not found")
    
    # Update fields if provided
    if version_update.version_type is not None:
        version.version_type = version_update.version_type
    
    if version_update.release_notes is not None:
        version.release_notes = version_update.release_notes
    
    if version_update.breaking_changes is not None:
        version.breaking_changes = version_update.breaking_changes
    
    db.commit()
    db.refresh(version)
    
    return version

@router.post("/apis/{api_id}/versions/{version_id}/promote")
async def promote_version(
    api_id: str,
    version_id: str,
    current_user: User = Depends(require_api_owner),
    db: Session = Depends(get_db)
):
    """Promote a version to stable/production"""
    version = db.query(APIVersion).filter(
        APIVersion.id == version_id,
        APIVersion.api_id == api_id
    ).first()
    
    if not version:
        raise HTTPException(status_code=404, detail="Version not found")
    
    if version.version_type == "stable":
        raise HTTPException(status_code=400, detail="Version is already stable")
    
    # Deactivate current active version
    current_active = db.query(APIVersion).filter(
        APIVersion.api_id == api_id,
        APIVersion.is_active == True
    ).first()
    
    if current_active:
        current_active.is_active = False
    
    # Promote this version
    version.version_type = "stable"
    version.is_active = True
    version.published_at = datetime.utcnow()
    
    db.commit()
    
    return {"message": f"Version {version.version_number} promoted to stable"}

@router.post("/apis/{api_id}/versions/{version_id}/rollback")
async def rollback_version(
    api_id: str,
    version_id: str,
    rollback_data: VersionRollback,
    current_user: User = Depends(require_api_owner),
    db: Session = Depends(get_db)
):
    """Rollback to a previous version"""
    target_version = db.query(APIVersion).filter(
        APIVersion.id == version_id,
        APIVersion.api_id == api_id
    ).first()
    
    if not target_version:
        raise HTTPException(status_code=404, detail="Version not found")
    
    if target_version.is_active:
        raise HTTPException(status_code=400, detail="Version is already active")
    
    # Deactivate current version
    current_active = db.query(APIVersion).filter(
        APIVersion.api_id == api_id,
        APIVersion.is_active == True
    ).first()
    
    if current_active:
        current_active.is_active = False
        # Log rollback event
        # TODO: Add rollback logging
    
    # Activate target version
    target_version.is_active = True
    
    db.commit()
    
    # TODO: Send notifications if rollback_data.notify_users is True
    
    return {
        "message": f"Successfully rolled back to version {target_version.version_number}",
        "reason": rollback_data.reason
    }

@router.post("/apis/{api_id}/versions/{version_id}/deprecate")
async def deprecate_version(
    api_id: str,
    version_id: str,
    sunset_date: Optional[datetime] = None,
    current_user: User = Depends(require_api_owner),
    db: Session = Depends(get_db)
):
    """Mark a version as deprecated"""
    version = db.query(APIVersion).filter(
        APIVersion.id == version_id,
        APIVersion.api_id == api_id
    ).first()
    
    if not version:
        raise HTTPException(status_code=404, detail="Version not found")
    
    if version.is_active:
        raise HTTPException(
            status_code=400, 
            detail="Cannot deprecate the active version"
        )
    
    version.deprecated_at = datetime.utcnow()
    version.sunset_date = sunset_date
    
    db.commit()
    
    return {"message": f"Version {version.version_number} marked as deprecated"}

@router.get("/apis/{api_id}/versions/compare")
async def compare_versions(
    api_id: str,
    version_a: str = Query(..., description="First version number"),
    version_b: str = Query(..., description="Second version number"),
    current_user: User = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """Compare two versions of an API"""
    # Get both versions
    v_a = db.query(APIVersion).filter(
        APIVersion.api_id == api_id,
        APIVersion.version_number == version_a
    ).first()
    
    v_b = db.query(APIVersion).filter(
        APIVersion.api_id == api_id,
        APIVersion.version_number == version_b
    ).first()
    
    if not v_a or not v_b:
        raise HTTPException(status_code=404, detail="One or both versions not found")
    
    # Mock comparison data - in real implementation, this would compare actual API specs
    comparison = VersionComparison(
        version_a=version_a,
        version_b=version_b,
        endpoints_added=[
            {"method": "POST", "path": "/api/v2/images/batch", "description": "Batch processing endpoint"}
        ],
        endpoints_removed=[
            {"method": "DELETE", "path": "/api/v1/images/:id", "description": "Legacy delete endpoint"}
        ],
        endpoints_modified=[
            {"method": "POST", "path": "/api/v2/images/enhance", "change": "Added quality parameter"}
        ],
        breaking_changes=v_b.breaking_changes if v_b.breaking_changes else []
    )
    
    return comparison

@router.get("/apis/{api_id}/versions/{version_id}/changelog")
async def get_version_changelog(
    api_id: str,
    version_id: str,
    current_user: User = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    """Get detailed changelog for a version"""
    version = db.query(APIVersion).filter(
        APIVersion.id == version_id,
        APIVersion.api_id == api_id
    ).first()
    
    if not version:
        raise HTTPException(status_code=404, detail="Version not found")
    
    # Mock changelog - in real implementation, this would be tracked
    changelog = {
        "version": version.version_number,
        "release_date": version.published_at,
        "type": version.version_type,
        "release_notes": version.release_notes,
        "breaking_changes": version.breaking_changes,
        "features": [
            "Added batch processing endpoints",
            "Improved watermark positioning",
            "Enhanced error handling"
        ],
        "fixes": [
            "Fixed memory leak in image processing",
            "Resolved authentication timeout issues"
        ],
        "improvements": [
            "Performance optimizations for large images",
            "Better error messages"
        ]
    }
    
    return changelog
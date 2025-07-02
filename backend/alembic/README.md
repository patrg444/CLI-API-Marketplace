# API-Direct Database Migrations

This directory contains database migrations managed by Alembic.

## Setup

1. Install dependencies:
   ```bash
   pip install alembic psycopg2-binary
   ```

2. Set database URL:
   ```bash
   export DATABASE_URL="postgresql://user:password@localhost/apidirect"
   ```

## Common Commands

### Create a new migration:
```bash
# Auto-generate migration from model changes
alembic revision --autogenerate -m "Add new feature"

# Create empty migration
alembic revision -m "Manual migration"
```

### Run migrations:
```bash
# Upgrade to latest version
alembic upgrade head

# Upgrade to specific version
alembic upgrade <revision>

# Show current version
alembic current

# Show migration history
alembic history
```

### Rollback migrations:
```bash
# Downgrade one revision
alembic downgrade -1

# Downgrade to specific revision
alembic downgrade <revision>

# Downgrade all (careful!)
alembic downgrade base
```

## Migration Files

Migrations are stored in the `versions/` directory with timestamps in the filename for better organization.

## Development Workflow

1. Make changes to models in `api/database.py`
2. Generate migration: `alembic revision --autogenerate -m "Description"`
3. Review generated migration and edit if needed
4. Apply migration: `alembic upgrade head`
5. Test the changes
6. Commit both model changes and migration files

## Production Deployment

1. Back up the database before running migrations
2. Test migrations on a staging environment first
3. Run migrations during low-traffic periods
4. Have a rollback plan ready

## Troubleshooting

### Connection Issues
- Ensure PostgreSQL is running
- Check DATABASE_URL environment variable
- Verify network connectivity

### Migration Conflicts
- Use `alembic merge` to resolve branched migrations
- Always pull latest migrations before creating new ones

### Schema Out of Sync
- Use `alembic stamp head` to mark database as up-to-date
- Use `alembic check` to verify migration status
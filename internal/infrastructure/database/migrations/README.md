# Database Migrations

This directory contains database migration files for the log parser application.

## Migration Structure

Each migration consists of two files:
- `XXX_migration_name_up.sql` - Forward migration (applies changes)
- `XXX_migration_name_down.sql` - Rollback migration (reverts changes)

Where `XXX` is a 3-digit sequence number (e.g., 001, 002, etc.).

## Current Migrations

### 001_initial_schema
**Purpose:** Creates the initial database schema with all base tables.

**Tables created:**
- `download_info` - Flash/download information with unique `tcu_pcba_number`
- `logistic_data` - Logistics and hardware information with unique `pcba_number`  
- `test_station_record` - Test station records linked to logistic data
- `test_step` - Individual test steps linked to test station records

**Applied:** Initial database setup

### 002_remove_unique_constraints
**Purpose:** Removes incorrect unique constraints that prevented legitimate duplicate PCBA numbers.

**Changes:**
- Removes unique constraint on `logistic_data.pcba_number`
- Removes unique constraint on `download_info.tcu_pcba_number`

**Rationale:** PCBA numbers can legitimately appear multiple times due to:
- Retests and quality checks
- Different test stations processing the same unit
- Multiple test cycles for the same hardware

**Applied:** After troubleshooting 100% extraction rate issues

## Running Migrations

### Manual Application (PostgreSQL)

**Apply migrations:**
```bash
# Apply initial schema
psql -h localhost -U admino -d pandora_logs -f 001_initial_schema_up.sql

# Apply constraint fixes  
psql -h localhost -U admino -d pandora_logs -f 002_remove_unique_constraints_up.sql
```

**Rollback migrations:**
```bash
# Rollback constraint changes
psql -h localhost -U admino -d pandora_logs -f 002_remove_unique_constraints_down.sql

# Rollback initial schema
psql -h localhost -U admino -d pandora_logs -f 001_initial_schema_down.sql
```

### Migration Tool Integration

For automated migration management, consider integrating tools like:
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [Goose](https://github.com/pressly/goose)
- [Atlas](https://atlasgo.io/)

## Best Practices

1. **Sequential Numbering:** Always use sequential numbers for migration files
2. **Descriptive Names:** Use clear, descriptive names for migration purposes
3. **Rollback Safety:** Ensure down migrations can safely revert changes
4. **Data Preservation:** Consider data preservation when dropping tables or constraints
5. **Testing:** Test both up and down migrations in development environment
6. **Backup:** Always backup production data before applying migrations

## Migration History

| Migration | Date Applied | Purpose | Status |
|-----------|-------------|---------|---------|
| 001 | Initial | Base schema creation | ✅ Applied |
| 002 | 2025-11-07 | Remove unique constraints | ✅ Applied |

## Notes

- The original `create_tables_schema.sql` file has been migrated to `001_initial_schema_up.sql`
- Migration 002 addresses critical issues discovered during log parser troubleshooting
- All migrations have been tested against the production log parsing workflow
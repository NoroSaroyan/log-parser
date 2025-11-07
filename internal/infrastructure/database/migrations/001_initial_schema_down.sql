-- Rollback initial database schema
-- Drops all tables created in 001_initial_schema_up.sql

-- Drop tables in reverse order to respect foreign key constraints
DROP TABLE IF EXISTS test_step CASCADE;
DROP TABLE IF EXISTS test_station_record CASCADE;
DROP TABLE IF EXISTS logistic_data CASCADE;
DROP TABLE IF EXISTS download_info CASCADE;
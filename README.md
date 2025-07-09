# Log Parser

## Project Overview

Log Parser is a modular Go application designed to extract, parse, and store data from large log files containing
embedded JSON about devices, tests, and logistics. The parsed data is persisted in a PostgreSQL database and made
accessible via a REST API.

## Technologies Used

- **Go 1.24** — application language
- **PostgreSQL 16** — relational database for storing parsed data
- **Docker & Docker Compose** — containerization and orchestration for easy deployment
- **Raw SQL** — direct database interaction without ORM
- **REST API** — data exposure for querying parsed log information

## Architecture & Structure

### Overview

The project follows a clean layered architecture to separate concerns clearly. It processes raw log files, transforms
the data into structured models, stores it in the database, and exposes retrieval APIs.

### Main Components

- `cmd/api` — REST API server entry point
- `cmd/cli` — CLI tool for manual or batch log file parsing
- `internal/domain/models` — DTOs and DB models representing parsed and stored data
- `internal/services` — business logic handling parsing, data transformation, and coordination
- `internal/repository` — data access layer executing raw SQL queries against PostgreSQL
- `deployments/` — Dockerfile and docker-compose configuration for containerized environment
- `configs/` — configuration files such as database credentials, ports, and app settings

### Data Flow

1. Log files are parsed into Data Transfer Objects (DTOs).
2. DTOs are converted to database models with proper linking, especially by `PCBANumber`.
3. Data is stored into PostgreSQL using raw SQL queries, maintaining strict control over database interactions.
4. REST API endpoints provide access to the parsed data for querying PCBA numbers and related information.

### Design Highlights

- Clear separation between DTOs and DB models ensures modularity and testability.
- Raw SQL usage optimizes database performance and gives fine-grained control.
- Containerization with Docker enables consistent deployment environments.
- Comprehensive logging and error handling across layers.

## Setup & Usage

### Prerequisites

- Docker and Docker Compose installed
- Access to a Unix-like terminal (Linux/macOS/WSL on Windows)

### Running with Docker Compose

```bash
# Start the entire application stack (API + PostgreSQL)
make up

# Stop the application stack
make down

# Restart the stack
make restart

# Build Docker images manually
make docker-build

# View real-time logs of all services
make logs

# List running containers in the stack
make ps

# Execute an interactive shell inside the app container
make exec-app

# Connect to the PostgreSQL database inside its container
make exec-postgres
```

### API Usage Examples

Ensure the app is running (default port 8080) and use these curl commands to query data:

```bash
# Get list of all PCBA numbers
curl -i "http://localhost:8080/api/v1/pcbanumbers"

# Get download info for a specific PCBA number
curl -i "http://localhost:8080/api/v1/download?pcbanumber=H8444A11100S60305140"

# Get final aggregated info for a PCBA number
curl -i "http://localhost:8080/api/v1/final?pcbanumber=H8444A11100S60305140"

# Get detailed PCBA info for a specific PCBA number
curl -i "http://localhost:8080/api/v1/pcba?pcbanumber=H8444A11100S60305140"
```

### Running the CLI parser locally

You can parse log files directly via the CLI:

```bash
go run cmd/main.go corporate_resources/
go run cmd/main.go corporate_resources/mesrestapi2.log
go run cmd/main.go corporate_resources/mesrestapi.log corporate_resources/mesrestapi2.log
```

Replace `corporate_resources/` with your log files directory or specific log file paths.

## Makefile Commands

For convenience, here are the Makefile commands available:

```makefile
up: ## Start the application stack
	@docker compose -f deployments/docker-compose.yml --project-directory . up -d

down: ## Stop the application stack
	@docker compose -f deployments/docker-compose.yml --project-directory . down

restart: ## Restart the application stack
	down up
docker-build: ## Build Docker images
	@docker compose -f deployments/docker-compose.yml --project-directory . build

logs: ## View real-time logs
	@docker compose -f deployments/docker-compose.yml --project-directory . logs -f

ps: ## List running containers
	@docker compose -f deployments/docker-compose.yml --project-directory . ps

exec-app: ## Enter app container shell
	@docker exec -it log-parser sh

exec-postgres: ## Enter PostgreSQL container shell
	@docker exec -it my_postgres psql -U $(POSTGRES_USER) -d $(POSTGRES_DB)

```

## Database Schema

The PostgreSQL database consists of the following main tables, designed to represent different stages and aspects of the
log data:

### `download_info`

Stores metadata about download operations linked to specific PCBA units.

- `id` (SERIAL PRIMARY KEY) — unique record identifier
- `test_station` (TEXT, NOT NULL) — station where the download took place
- `flash_entity_type` (TEXT) — type of flash entity
- `tcu_pcba_number` (TEXT, NOT NULL, UNIQUE) — unique PCBA number for the TCU (Telematics Control Unit)
- `flash_elapsed_time` (INTEGER) — duration of flashing in seconds
- `tcu_entity_flash_state` (TEXT) — state of the flash entity
- `part_number` (TEXT) — part number identifier
- `product_line` (TEXT) — product line name
- `download_tool_version` (TEXT) — version of the download tool used
- `download_finished_time` (TEXT) — timestamp when download finished

### `logistic_data`

Captures hardware, software, and device logistics information associated with each PCBA.

- `id` (SERIAL PRIMARY KEY) — unique record identifier
- `pcba_number` (TEXT, NOT NULL, UNIQUE) — unique PCBA number
- `product_sn` (TEXT) — product serial number
- `part_number` (TEXT) — part number
- `vp_app_version` (TEXT) — VP application version
- `vp_boot_loader_version` (TEXT) — VP boot loader version
- `vp_core_version` (TEXT) — VP core software version
- `supplier_hardware_version` (TEXT) — hardware version from the supplier
- `manufacturer_hardware_version` (TEXT) — hardware version from the manufacturer
- `manufacturer_software_version` (TEXT) — software version from the manufacturer
- `ble_mac` (TEXT) — Bluetooth MAC address
- `ble_sn` (TEXT) — Bluetooth serial number
- `ble_version` (TEXT) — Bluetooth version
- `ble_passwork_key` (TEXT) — Bluetooth passkey
- `ap_app_version` (TEXT) — application processor application version
- `ap_kernel_version` (TEXT) — application processor kernel version
- `tcu_iccid` (TEXT) — ICCID for TCU SIM card
- `phone_number` (TEXT) — associated phone number
- `imei` (TEXT) — device IMEI number
- `imsi` (TEXT) — IMSI number
- `production_date` (TEXT) — production date

### `test_station_record`

Represents individual test sessions performed on a PCBA, linked to logistic data.

- `id` (SERIAL PRIMARY KEY) — unique record identifier
- `part_number` (TEXT) — part number
- `test_station` (TEXT, NOT NULL) — station where the test was conducted
- `entity_type` (TEXT) — type of entity tested
- `product_line` (TEXT) — product line
- `test_tool_version` (TEXT) — version of test tool used
- `test_finished_time` (TEXT) — test completion timestamp
- `is_all_passed` (BOOLEAN) — overall pass/fail status of the test
- `error_codes` (TEXT) — any error codes generated
- `logistic_data_id` (INTEGER, NOT NULL) — foreign key referencing `logistic_data(id)`; enforces cascading delete on
  logistic data removal

### `test_step`

Details each step within a test session, linked to the corresponding test station record.

- `id` (SERIAL PRIMARY KEY) — unique record identifier
- `test_step_name` (TEXT, NOT NULL) — name of the test step
- `test_threshold_value` (TEXT) — threshold value for the test step
- `test_measured_value` (TEXT) — measured value recorded
- `test_step_elapsed_time` (INTEGER) — duration of the step in seconds
- `test_step_result` (TEXT) — result status (e.g., pass/fail)
- `test_step_error_code` (TEXT) — error code if any
- `test_station_record_id` (INTEGER, NOT NULL) — foreign key referencing `test_station_record(id)`; enforces cascading
  delete on test station record removal

---

### Table Relationships

- `logistic_data.pcba_number` uniquely identifies a PCBA unit.
- `test_station_record` links to `logistic_data` via `logistic_data_id`.
- `test_step` links to `test_station_record` via `test_station_record_id`.
- `download_info.tcu_pcba_number` stores unique PCBA numbers related to download operations.

This schema supports a normalized relational model linking raw device info, test sessions, and individual test steps
tied to their respective PCBA identifiers.

## Error Handling

The application implements comprehensive error handling at all layers:

- File parsing errors are logged and skipped
- Database errors are propagated with context
- API errors return appropriate HTTP status codes
- All errors are logged with stack traces when available

## Logging

The application uses structured logging with:

- Different log levels (DEBUG, INFO, WARN, ERROR)
- Contextual information in each log entry
- Correlation IDs for tracing requests through the system
- Separate log files for different components

## Performance Considerations

The parser is optimized for:

- Batch database inserts for better performance
- Connection pooling for database access

## Future Enhancements

Planned future improvements include:

## Future Enhancements

Planned improvements and upcoming features:

- **Stream-Based JSON Parsing for Large Logs**  
  Shift from in-memory parsing to a streaming model. This will allow the app to process multi-gigabyte log files line by
  line, dramatically reducing memory usage and supporting constrained environments.

- **Parallel Processing for Faster Ingestion**  
  Add worker pools to process and insert parsed data concurrently. This will unlock better CPU utilization and cut
  processing times for massive datasets.

- **Secure API with Authentication**  
  Introduce API key or JWT-based authentication to protect endpoints and control access to sensitive data.

- **Richer Query Capabilities**  
  Extend API endpoints with filtering (date ranges, test outcomes, metadata fields) and pagination to handle large
  result sets more elegantly.

- **Data Export Tools**  
  Enable users to export parsed data in common formats (CSV, JSON, XML) for integration with external systems or
  reporting pipelines.

- **Web Dashboard for Visualization**  
  Build a lightweight dashboard to visualize test results, monitor parsing progress, and surface aggregated insights on
  PCBA numbers and test stations.

- **Alerts and Monitoring**  
  Implement real-time alerting (via Slack, email, or webhooks) for parsing failures or critical errors. Expose
  Prometheus-compatible metrics for resource monitoring in production.

- **Incremental & Real-Time Parsing Support**  
  Support incremental parsing of newly appended log data without reprocessing already-ingested content, laying the
  foundation for real-time ingestion pipelines.

- **Customizable Resource Controls**  
  Let users fine-tune memory and CPU usage through configuration (e.g., worker counts, buffer sizes), adapting the app
  to a wide range of deployment environments.

- **Smarter Error Handling & Recovery**  
  Improve resilience by skipping over corrupted JSON blocks gracefully and retrying partial DB transactions instead of
  failing the entire process.

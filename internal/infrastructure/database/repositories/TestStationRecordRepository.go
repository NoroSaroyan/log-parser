package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"log-parser/internal/domain/models/db"
)

// testStationRecordRepository provides methods for accessing and manipulating
// TestStationRecordDB entities in the database.
type testStationRecordRepository struct {
	db *sql.DB
}

// NewTestStationRecordRepository initializes a new TestStationRecord repository.
//
// It accepts an *sql.DB instance and returns a repository capable of performing
// CRUD operations on TestStationRecordDB models.
func NewTestStationRecordRepository(db *sql.DB) *testStationRecordRepository {
	return &testStationRecordRepository{db: db}
}

// Insert adds a new TestStationRecordDB into the database.
//
// It populates the given record's ID field with the auto-generated primary key.
// Returns an error if the insert fails or if no ID is returned.
func (r *testStationRecordRepository) Insert(ctx context.Context, rec *db.TestStationRecordDB) error {
	query := `
    INSERT INTO test_station_record 
    (part_number, test_station, entity_type, product_line, test_tool_version, test_finished_time, is_all_passed, error_codes, logistic_data_id)
    VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
    RETURNING id
    `
	err := r.db.QueryRowContext(ctx, query,
		rec.PartNumber, rec.TestStation, rec.EntityType, rec.ProductLine,
		rec.TestToolVersion, rec.TestFinishedTime, rec.IsAllPassed, rec.ErrorCodes, rec.LogisticDataID,
	).Scan(&rec.ID)
	if err != nil {
		return fmt.Errorf("failed to insert TestStationRecord and retrieve ID: %w", err)
	}
	if rec.ID == 0 {
		return fmt.Errorf("unexpected: inserted TestStationRecord returned ID=0")
	}
	return nil
}

// GetByPCBANumber retrieves all TestStationRecordDB entries that are linked
// to a LogisticData record with the specified PCBA number.
//
// Returns a slice of TestStationRecordDB pointers or an error if the query fails.
func (r *testStationRecordRepository) GetByPCBANumber(ctx context.Context, pcba string) ([]*db.TestStationRecordDB, error) {
	query := `
    SELECT tsr.id, tsr.part_number, tsr.test_station, tsr.entity_type, tsr.product_line, tsr.test_tool_version,
           tsr.test_finished_time, tsr.is_all_passed, tsr.error_codes, tsr.logistic_data_id
    FROM test_station_record tsr
    JOIN logistic_data ld ON tsr.logistic_data_id = ld.id
    WHERE ld.pcba_number = $1
    `
	rows, err := r.db.QueryContext(ctx, query, pcba)
	if err != nil {
		return nil, fmt.Errorf("failed to query TestStationRecords by PCBA number: %w", err)
	}
	defer rows.Close()

	var results []*db.TestStationRecordDB
	for rows.Next() {
		var rec db.TestStationRecordDB
		if err := rows.Scan(
			&rec.ID, &rec.PartNumber, &rec.TestStation, &rec.EntityType, &rec.ProductLine,
			&rec.TestToolVersion, &rec.TestFinishedTime, &rec.IsAllPassed, &rec.ErrorCodes, &rec.LogisticDataID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan TestStationRecord row: %w", err)
		}
		results = append(results, &rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return results, nil
}

// GetAllPCBANumbers retrieves all distinct PCBA numbers present in the database.
//
// This method performs a JOIN with the LogisticData table to extract unique PCBA numbers
// associated with TestStationRecords. Returns an error if the query or row scanning fails.
func (r *testStationRecordRepository) GetAllPCBANumbers(ctx context.Context) ([]string, error) {
	query := `
		SELECT DISTINCT l.pcba_number
		FROM test_station_record tsr
		JOIN logistic_data l ON tsr.logistic_data_id = l.id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query PCBA numbers: %w", err)
	}
	defer rows.Close()

	var pcbas []string
	for rows.Next() {
		var pcba string
		if err := rows.Scan(&pcba); err != nil {
			return nil, fmt.Errorf("failed to scan PCBA number: %w", err)
		}
		pcbas = append(pcbas, pcba)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return pcbas, nil
}

// GetByPartNumber retrieves all TestStationRecordDB entries that match the given part number.
//
// Returns a slice of TestStationRecordDB pointers or an error if the query fails or scanning fails.
func (r *testStationRecordRepository) GetByPartNumber(ctx context.Context, partNumber string) ([]*db.TestStationRecordDB, error) {
	query := `
	SELECT id, part_number, test_station, entity_type, product_line, test_tool_version,
	       test_finished_time, is_all_passed, error_codes, logistic_data_id
	FROM test_station_record
	WHERE part_number = $1
	`
	rows, err := r.db.QueryContext(ctx, query, partNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to query TestStationRecords by part number: %w", err)
	}
	defer rows.Close()

	var results []*db.TestStationRecordDB
	for rows.Next() {
		var rec db.TestStationRecordDB
		if err := rows.Scan(
			&rec.ID, &rec.PartNumber, &rec.TestStation, &rec.EntityType, &rec.ProductLine,
			&rec.TestToolVersion, &rec.TestFinishedTime, &rec.IsAllPassed, &rec.ErrorCodes, &rec.LogisticDataID,
		); err != nil {
			return nil, fmt.Errorf("failed to scan TestStationRecord row: %w", err)
		}
		results = append(results, &rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}
	return results, nil
}

// GetByID retrieves a single TestStationRecordDB entry by its unique ID.
//
// If no record is found, it returns (nil, nil). Returns an error if the query fails
// or scanning the row encounters an issue.
func (r *testStationRecordRepository) GetByID(ctx context.Context, id int) (*db.TestStationRecordDB, error) {
	query := `
	SELECT id, part_number, test_station, entity_type, product_line, test_tool_version,
	       test_finished_time, is_all_passed, error_codes, logistic_data_id
	FROM test_station_record
	WHERE id = $1
	`
	var rec db.TestStationRecordDB
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&rec.ID, &rec.PartNumber, &rec.TestStation, &rec.EntityType, &rec.ProductLine,
		&rec.TestToolVersion, &rec.TestFinishedTime, &rec.IsAllPassed, &rec.ErrorCodes, &rec.LogisticDataID,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Return nil if no record was found
		}
		return nil, fmt.Errorf("failed to fetch TestStationRecord by ID: %w", err)
	}
	return &rec, nil
}

// Package repositories provides database persistence implementations for domain entities.
//
// This package defines repository types responsible for managing data access and manipulation
// of core domain models within the application using raw SQL queries and the standard
// database/sql package.
//
// The repositories encapsulate all CRUD (Create, Read, Update, Delete) operations,
// enabling separation of concerns by isolating data storage logic from business logic.
//
// Features:
//
// - Transactional batch inserts ensuring atomic operations.
// - Querying by foreign keys and filtering by domain-specific fields.
// - Explicit usage of context.Context to support cancellation and timeouts.
// - Proper management of prepared statements and SQL transactions for efficiency and safety.
// - Direct SQL query usage for maximum control and performance without ORM overhead.
//
// Usage:
//
// To use a repository, create an instance by providing a configured *sql.DB connection:
//
//	db, err := sql.Open("postgres", dsn)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer db.Close()
//
//	testStepRepo := repositories.NewTestStepRepository(db)
//
// Then, invoke repository methods with a context to perform operations:
//
//	ctx := context.Background()
//	steps, err := testStepRepo.GetByTestStationRecordID(ctx, someRecordID)
//	if err != nil {
//	    // handle error
//	}
//
// This design allows easy testing, swapping out database implementations, and clear
// domain-driven design by decoupling persistence concerns from service and domain layers.
package repositories

import (
	"context"
	"database/sql"
	"log-parser/internal/domain/models/db"
)

// testStepRepository provides methods to interact with TestStepDB records in the database.
type testStepRepository struct {
	db *sql.DB
}

// NewTestStepRepository creates a new instance of testStepRepository.
//
// It takes an *sql.DB instance and returns a repository that can perform
// CRUD operations on TestStepDB models.
func NewTestStepRepository(db *sql.DB) *testStepRepository {
	return &testStepRepository{db: db}
}

// InsertBatch inserts multiple TestStepDB records in a single database transaction.
//
// Each step in the provided slice is linked to the specified testStationRecordID.
// If any insertion fails, the entire transaction is rolled back.
//
// Parameters:
//   - ctx: context for cancellation and timeout.
//   - steps: slice of TestStepDB pointers to insert.
//   - testStationRecordID: foreign key ID linking steps to a TestStationRecord.
//
// Returns an error if the transaction fails or if preparing/executing the statement fails.
func (r *testStepRepository) InsertBatch(ctx context.Context, steps []*db.TestStepDB, testStationRecordID int) error {
	query := `
    INSERT INTO test_step 
    (test_step_name, test_threshold_value, test_measured_value, test_step_elapsed_time, test_step_result, test_step_error_code, test_station_record_id)
    VALUES ($1,$2,$3,$4,$5,$6,$7)
    `
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, step := range steps {
		if _, err := stmt.ExecContext(ctx,
			step.TestStepName, step.TestThresholdValue, step.TestMeasuredValue, step.TestStepElapsedTime,
			step.TestStepResult, step.TestStepErrorCode, testStationRecordID,
		); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// GetByTestStationRecordID retrieves all TestStepDB records associated with the given testStationRecordID.
//
// Parameters:
//   - ctx: context for cancellation and timeout.
//   - recordID: ID of the TestStationRecord to fetch steps for.
//
// Returns a slice of TestStepDB pointers or an error if the query or row scanning fails.
func (r *testStepRepository) GetByTestStationRecordID(ctx context.Context, recordID int) ([]*db.TestStepDB, error) {
	query := `
    SELECT test_step_name, test_threshold_value, test_measured_value, test_step_elapsed_time, test_step_result, test_step_error_code
    FROM test_step
    WHERE test_station_record_id = $1
    `
	rows, err := r.db.QueryContext(ctx, query, recordID)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var results []*db.TestStepDB
	for rows.Next() {
		var s db.TestStepDB
		if err := rows.Scan(
			&s.TestStepName, &s.TestThresholdValue, &s.TestMeasuredValue, &s.TestStepElapsedTime,
			&s.TestStepResult, &s.TestStepErrorCode,
		); err != nil {
			return nil, err
		}
		results = append(results, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// GetByPartNumber retrieves all TestStepDB records linked to a specific part number.
//
// This performs an INNER JOIN with the TestStationRecord table to filter steps
// by the provided part number.
//
// Parameters:
//   - ctx: context for cancellation and timeout.
//   - partNumber: the part number to filter TestStepDB records by.
//
// Returns a slice of TestStepDB pointers or an error if the query or scanning fails.
func (r *testStepRepository) GetByPartNumber(ctx context.Context, partNumber string) ([]*db.TestStepDB, error) {
	query := `
	SELECT ts.test_step_name, ts.test_threshold_value, ts.test_measured_value, 
	       ts.test_step_elapsed_time, ts.test_step_result, ts.test_step_error_code
	FROM test_step ts
	INNER JOIN test_station_record tsr ON ts.test_station_record_id = tsr.id
	WHERE tsr.part_number = $1
	`
	rows, err := r.db.QueryContext(ctx, query, partNumber)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = rows.Close()
	}()

	var results []*db.TestStepDB
	for rows.Next() {
		var s db.TestStepDB
		if err := rows.Scan(
			&s.TestStepName, &s.TestThresholdValue, &s.TestMeasuredValue, &s.TestStepElapsedTime,
			&s.TestStepResult, &s.TestStepErrorCode,
		); err != nil {
			return nil, err
		}
		results = append(results, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

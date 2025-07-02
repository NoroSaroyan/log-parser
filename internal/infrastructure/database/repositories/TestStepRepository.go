package database

import (
	"context"
	"database/sql"
	"log-parser/internal/domain/models/db"
)

type testStepRepo struct {
	db *sql.DB
}

func NewTestStepRepo(db *sql.DB) *testStepRepo {
	return &testStepRepo{db: db}
}

func (r *testStepRepo) InsertBatch(ctx context.Context, steps []*db.TestStepDB, testStationRecordID int) error {
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
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, step := range steps {
		if _, err := stmt.ExecContext(ctx,
			step.TestStepName, step.TestThresholdValue, step.TestMeasuredValue, step.TestStepElapsedTime,
			step.TestStepResult, step.TestStepErrorCode, testStationRecordID,
		); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (r *testStepRepo) GetByTestStationRecordID(ctx context.Context, recordID int) ([]*db.TestStepDB, error) {
	query := `
    SELECT test_step_name, test_threshold_value, test_measured_value, test_step_elapsed_time, test_step_result, test_step_error_code
    FROM test_step
    WHERE test_station_record_id = $1
    `
	rows, err := r.db.QueryContext(ctx, query, recordID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

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

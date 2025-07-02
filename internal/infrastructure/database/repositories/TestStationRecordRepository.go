package database

import (
	"context"
	"database/sql"
	"log-parser/internal/domain/models/db"
)

type testStationRecordRepo struct {
	db *sql.DB
}

func NewTestStationRecordRepo(db *sql.DB) *testStationRecordRepo {
	return &testStationRecordRepo{db: db}
}

func (r *testStationRecordRepo) Insert(ctx context.Context, rec *db.TestStationRecordDB) error {
	query := `
    INSERT INTO test_station_record 
    (part_number, test_station, entity_type, product_line, test_tool_version, test_finished_time, is_all_passed, error_codes, logistic_data_id)
    VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
    `
	_, err := r.db.ExecContext(ctx, query,
		rec.PartNumber, rec.TestStation, rec.EntityType, rec.ProductLine,
		rec.TestToolVersion, rec.TestFinishedTime, rec.IsAllPassed, rec.ErrorCodes, rec.LogisticDataID,
	)
	return err
}

func (r *testStationRecordRepo) GetByPCBANumber(ctx context.Context, pcba string) ([]*db.TestStationRecordDB, error) {
	// Join with logistic_data to filter by pcba_number
	query := `
    SELECT tsr.id, tsr.part_number, tsr.test_station, tsr.entity_type, tsr.product_line, tsr.test_tool_version,
           tsr.test_finished_time, tsr.is_all_passed, tsr.error_codes, tsr.logistic_data_id
    FROM test_station_record tsr
    JOIN logistic_data ld ON tsr.logistic_data_id = ld.id
    WHERE ld.pcba_number = $1
    `
	rows, err := r.db.QueryContext(ctx, query, pcba)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*db.TestStationRecordDB
	for rows.Next() {
		var rec db.TestStationRecordDB
		if err := rows.Scan(
			&rec.ID, &rec.PartNumber, &rec.TestStation, &rec.EntityType, &rec.ProductLine,
			&rec.TestToolVersion, &rec.TestFinishedTime, &rec.IsAllPassed, &rec.ErrorCodes, &rec.LogisticDataID,
		); err != nil {
			return nil, err
		}
		results = append(results, &rec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

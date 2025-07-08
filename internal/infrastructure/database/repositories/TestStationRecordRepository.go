package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"log-parser/internal/domain/models/db"
)

type testStationRecordRepository struct {
	db *sql.DB
}

func (r *testStationRecordRepository) GetByPartNumber(ctx context.Context, partNumber string) ([]*db.TestStationRecordDB, error) {
	//TODO implement me
	panic("implement me")
}

func NewTestStationRecordRepository(db *sql.DB) *testStationRecordRepository {
	return &testStationRecordRepository{db: db}
}

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

package repositories

import (
	"database/sql"
	"log-parser/internal/domain/models"
)

type testStationRecordRepo struct {
	db *sql.DB
}

func NewTestStationRecordRepo(db *sql.DB) TestStationRecordRepository {
	return &testStationRecordRepo{db: db}
}

func (r *testStationRecordRepo) Insert(record *models.TestStationRecordDTO) (int, error) {
	query := `INSERT INTO test_station_record
		(part_number, test_station, entity_type, product_line, test_tool_version, test_finished_time, is_all_passed, error_codes, logistic_data_id)
		VALUES
		($1,$2,$3,$4,$5,$6,$7,$8,$9)
		RETURNING id`
	var id int
	err := r.db.QueryRow(query,
		record.PartNumber,
		record.TestStation,
		record.EntityType,
		record.ProductLine,
		record.TestToolVersion,
		record.TestFinishedTime,
		record.IsAllPassed,
		record.ErrorCodes,
		record.LogisticDataID, // **You will need to add LogisticDataID field in your model for DB foreign key**
	).Scan(&id)
	return id, err
}

func (r *testStationRecordRepo) GetByPCBANumber(pcba string) ([]*models.TestStationRecordDTO, error) {
	// Join with logistic_data to filter by PCBANumber
	query := `SELECT tsr.id, tsr.part_number, tsr.test_station, tsr.entity_type, tsr.product_line, tsr.test_tool_version,
		tsr.test_finished_time, tsr.is_all_passed, tsr.error_codes, tsr.logistic_data_id
		FROM test_station_record tsr
		INNER JOIN logistic_data ld ON tsr.logistic_data_id = ld.id
		WHERE ld.pcba_number = $1`

	rows, err := r.db.Query(query, pcba)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.TestStationRecordDTO
	for rows.Next() {
		rec := new(models.TestStationRecordDTO)
		err := rows.Scan(
			&rec.ID, &rec.PartNumber, &rec.TestStation, &rec.EntityType, &rec.ProductLine,
			&rec.TestToolVersion, &rec.TestFinishedTime, &rec.IsAllPassed, &rec.ErrorCodes, &rec.LogisticDataID,
		)
		if err != nil {
			return nil, err
		}
		results = append(results, rec)
	}
	return results, nil
}

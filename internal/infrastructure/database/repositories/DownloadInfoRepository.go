package repositories

import (
	"context"
	"database/sql"
	"log"
	"log-parser/internal/domain/models/db"
	"log-parser/internal/domain/repositories"
)

// DownloadInfoRepository provides methods for persisting and retrieving
// DownloadInfo records in the database.
//
// It implements the repositories.DownloadInfoRepository interface.
type DownloadInfoRepository struct {
	db *sql.DB
}

// NewDownloadInfoRepository creates a new DownloadInfoRepository.
//
// It panics if the provided *sql.DB is nil.
func NewDownloadInfoRepository(db *sql.DB) *DownloadInfoRepository {
	if db == nil {
		log.Fatal("DB is nil in NewDownloadInfoRepository")
	}
	return &DownloadInfoRepository{db: db}
}

// Insert adds a new DownloadInfoDB record to the download_info table.
//
// If a record with the same tcu_pcba_number already exists, the insertion
// is skipped due to ON CONFLICT DO NOTHING. Returns any database error encountered.
func (r *DownloadInfoRepository) Insert(ctx context.Context, d *db.DownloadInfoDB) error {
	query := `
	INSERT INTO download_info 
	(test_station, flash_entity_type, tcu_pcba_number, flash_elapsed_time, tcu_entity_flash_state, part_number, product_line, download_tool_version, download_finished_time)
	VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	ON CONFLICT (tcu_pcba_number) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query,
		d.TestStation, d.FlashEntityType, d.TcuPCBANumber, d.FlashElapsedTime,
		d.TcuEntityFlashState, d.PartNumber, d.ProductLine, d.DownloadToolVersion, d.DownloadFinishedTime,
	)
	return err
}

// GetByPCBANumber retrieves a single DownloadInfoDB record by tcu_pcba_number.
//
// Returns:
//   - (*DownloadInfoDB, nil) if a matching record is found.
//   - (nil, nil) if no record exists for the given PCBA number.
//   - (nil, error) if a database error occurs.
func (r *DownloadInfoRepository) GetByPCBANumber(ctx context.Context, pcba string) (*db.DownloadInfoDB, error) {
	query := `
	SELECT test_station, flash_entity_type, tcu_pcba_number, flash_elapsed_time, 
	       tcu_entity_flash_state, part_number, product_line, download_tool_version, download_finished_time
	FROM download_info
	WHERE tcu_pcba_number = $1
	LIMIT 1
	`

	var d db.DownloadInfoDB
	err := r.db.QueryRowContext(ctx, query, pcba).Scan(
		&d.TestStation,
		&d.FlashEntityType,
		&d.TcuPCBANumber,
		&d.FlashElapsedTime,
		&d.TcuEntityFlashState,
		&d.PartNumber,
		&d.ProductLine,
		&d.DownloadToolVersion,
		&d.DownloadFinishedTime,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &d, nil
}

// GetByPartNumber retrieves a single DownloadInfoDB record by part_number.
//
// TODO: Implement this method.
func (r *DownloadInfoRepository) GetByPartNumber(ctx context.Context, partNumber string) (*db.DownloadInfoDB, error) {
	// TODO: implement me
	panic("implement me")
}

// Ensure DownloadInfoRepository implements the repositories.DownloadInfoRepository interface.
var _ repositories.DownloadInfoRepository = (*DownloadInfoRepository)(nil)

package repositories

import (
	"context"
	"database/sql"
	"log-parser/internal/domain/models/db"
	"log-parser/internal/domain/repositories"
)

type downloadInfoRepository struct {
	db *sql.DB
}

func (r *downloadInfoRepository) GetByPartNumber(ctx context.Context, partNumber string) (*db.DownloadInfoDB, error) {
	//TODO implement me
	panic("implement me")
}

func NewDownloadInfoRepository(db *sql.DB) *downloadInfoRepository {
	return &downloadInfoRepository{db: db}
}

func (r *downloadInfoRepository) Insert(ctx context.Context, d *db.DownloadInfoDB) error {
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

func (r *downloadInfoRepository) GetByPCBANumber(ctx context.Context, pcba string) (*db.DownloadInfoDB, error) {
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

var _ repositories.DownloadInfoRepository = (*downloadInfoRepository)(nil)

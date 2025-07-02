package repositories

import (
	"database/sql"
	"log-parser/internal/domain/models"
)

type downloadInfoRepo struct {
	db *sql.DB
}

func NewDownloadInfoRepo(db *sql.DB) DownloadInfoRepository {
	return &downloadInfoRepo{db: db}
}

func (r *downloadInfoRepo) Insert(info *models.DownloadInfoDTO) error {
	query := `INSERT INTO download_info 
		(test_station, flash_entity_type, tcu_pcba_number, flash_elapsed_time, tcu_entity_flash_state, part_number, product_line, download_tool_version, download_finished_time)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`

	_, err := r.db.Exec(query,
		info.TestStation, info.FlashEntityType, info.TcuPCBANumber, info.FlashElapsedTime, info.TcuEntityFlashState,
		info.PartNumber, info.ProductLine, info.DownloadToolVersion, info.DownloadFinishedTime,
	)
	return err
}

func (r *downloadInfoRepo) GetByPCBANumber(pcba string) (*models.DownloadInfoDTO, error) {
	query := `SELECT test_station, flash_entity_type, tcu_pcba_number, flash_elapsed_time, tcu_entity_flash_state, part_number, product_line, download_tool_version, download_finished_time 
			  FROM download_info WHERE tcu_pcba_number = $1`
	row := r.db.QueryRow(query, pcba)

	var info models.DownloadInfoDTO
	err := row.Scan(
		&info.TestStation, &info.FlashEntityType, &info.TcuPCBANumber, &info.FlashElapsedTime, &info.TcuEntityFlashState,
		&info.PartNumber, &info.ProductLine, &info.DownloadToolVersion, &info.DownloadFinishedTime,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &info, nil
}

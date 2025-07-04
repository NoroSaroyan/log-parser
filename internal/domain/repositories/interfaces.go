package repositories

import (
	"context"
	"log-parser/internal/domain/models/db"
)

type DownloadInfoRepository interface {
	Insert(ctx context.Context, info *db.DownloadInfoDB) error
	GetByPCBANumber(ctx context.Context, pcba string) ([]*db.DownloadInfoDB, error)
	GetByPartNumber(ctx context.Context, partNumber string) ([]*db.DownloadInfoDB, error)
}

type LogisticDataRepository interface {
	Insert(ctx context.Context, data *db.LogisticDataDB) error
	GetByPCBANumber(ctx context.Context, pcba string) ([]*db.LogisticDataDB, error)
	GetByPartNumber(ctx context.Context, partNumber string) ([]*db.LogisticDataDB, error)
	GetIDByPCBANumber(ctx context.Context, pcba string) (int, error)
}

type TestStationRecordRepository interface {
	Insert(ctx context.Context, record *db.TestStationRecordDB) error
	GetByPCBANumber(ctx context.Context, pcba string) ([]*db.TestStationRecordDB, error)
	GetByPartNumber(ctx context.Context, partNumber string) ([]*db.TestStationRecordDB, error)
}

type TestStepRepository interface {
	InsertBatch(ctx context.Context, steps []*db.TestStepDB, testStationRecordID int) error
	GetByTestStationRecordID(ctx context.Context, recordID int) ([]*db.TestStepDB, error)
	GetByPartNumber(ctx context.Context, partNumber string) ([]*db.TestStepDB, error)
}

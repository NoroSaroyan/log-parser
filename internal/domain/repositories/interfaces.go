package repositories

import (
	"log-parser/internal/domain/models/db"
)

type DownloadInfoRepository interface {
	Insert(info *db.DownloadInfoDB) error
	GetByPCBANumber(pcba string) ([]*db.DownloadInfoDB, error)
}

type LogisticDataRepository interface {
	Insert(data *db.LogisticDataDB) error
	GetByPCBANumber(pcba string) ([]*db.LogisticDataDB, error)
}

type TestStationRecordRepository interface {
	Insert(record *db.TestStationRecordDB) error
	GetByPCBANumber(pcba string) ([]*db.TestStationRecordDB, error)
}

type TestStepRepository interface {
	InsertBatch(steps []*db.TestStepDB, testStationRecordID int) error
	GetByTestStationRecordID(recordID int) ([]*db.TestStepDB, error)
}

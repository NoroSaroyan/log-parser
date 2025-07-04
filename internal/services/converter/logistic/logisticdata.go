package logistic

import (
	db "log-parser/internal/domain/models/db"
	dto "log-parser/internal/domain/models/dto"
)

func ConvertToDB(dto dto.LogisticDataDTO) db.LogisticDataDB {
	return db.LogisticDataDB{
		PCBANumber:                  dto.PCBANumber,
		ProductSN:                   dto.ProductSN,
		PartNumber:                  dto.PartNumber,
		VPAppVersion:                dto.VPAppVersion,
		VPBootLoaderVersion:         dto.VPBootLoaderVersion,
		VPCoreVersion:               dto.VPCoreVersion,
		SupplierHardwareVersion:     dto.SupplierHardwareVersion,
		ManufacturerHardwareVersion: dto.ManufacturerHardwareVersion,
		ManufacturerSoftwareVersion: dto.ManufacturerSoftwareVersion,
		BleMac:                      dto.BleMac,
		BleSN:                       dto.BleSN,
		BleVersion:                  dto.BleVersion,
		BlePassworkKey:              dto.BlePassworkKey,
		APAppVersion:                dto.APAppVersion,
		APKernelVersion:             dto.APKernelVersion,
		TcuICCID:                    dto.TcuICCID,
		PhoneNumber:                 dto.PhoneNumber,
		IMEI:                        dto.IMEI,
		IMSI:                        dto.IMSI,
		ProductionDate:              dto.ProductionDate,
	}
}

func ConvertToDTO(db db.LogisticDataDB) dto.LogisticDataDTO {
	return dto.LogisticDataDTO{
		PCBANumber:                  db.PCBANumber,
		ProductSN:                   db.ProductSN,
		PartNumber:                  db.PartNumber,
		VPAppVersion:                db.VPAppVersion,
		VPBootLoaderVersion:         db.VPBootLoaderVersion,
		VPCoreVersion:               db.VPCoreVersion,
		SupplierHardwareVersion:     db.SupplierHardwareVersion,
		ManufacturerHardwareVersion: db.ManufacturerHardwareVersion,
		ManufacturerSoftwareVersion: db.ManufacturerSoftwareVersion,
		BleMac:                      db.BleMac,
		BleSN:                       db.BleSN,
		BleVersion:                  db.BleVersion,
		BlePassworkKey:              db.BlePassworkKey,
		APAppVersion:                db.APAppVersion,
		APKernelVersion:             db.APKernelVersion,
		TcuICCID:                    db.TcuICCID,
		PhoneNumber:                 db.PhoneNumber,
		IMEI:                        db.IMEI,
		IMSI:                        db.IMSI,
		ProductionDate:              db.ProductionDate,
	}
}

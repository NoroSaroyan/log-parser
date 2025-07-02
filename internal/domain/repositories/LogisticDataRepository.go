package repositories

import (
	"database/sql"
	"log-parser/internal/domain/models"
)

type logisticDataRepo struct {
	db *sql.DB
}

func NewLogisticDataRepo(db *sql.DB) LogisticDataRepository {
	return &logisticDataRepo{db: db}
}

func (r *logisticDataRepo) Insert(data *models.LogisticDataDTO) (int, error) {
	query := `INSERT INTO logistic_data 
		(pcba_number, product_sn, part_number, vp_app_version, vp_boot_loader_version, vp_core_version,
		supplier_hardware_version, manufacturer_hardware_version, manufacturer_software_version,
		ble_mac, ble_sn, ble_version, ble_passwork_key,
		ap_app_version, ap_kernel_version, tcu_iccid, phone_number, imei, imsi, production_date)
		VALUES
		($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
		RETURNING id`
	var id int
	err := r.db.QueryRow(query,
		data.PCBANumber, data.ProductSN, data.PartNumber, data.VPAppVersion, data.VPBootLoaderVersion, data.VPCoreVersion,
		data.SupplierHardwareVersion, data.ManufacturerHardwareVersion, data.ManufacturerSoftwareVersion,
		data.BleMac, data.BleSN, data.BleVersion, data.BlePassworkKey,
		data.APAppVersion, data.APKernelVersion, data.TcuICCID, data.PhoneNumber, data.IMEI, data.IMSI, data.ProductionDate,
	).Scan(&id)
	return id, err
}

func (r *logisticDataRepo) GetByPCBANumber(pcba string) (*models.LogisticDataDTO, error) {
	query := `SELECT id, pcba_number, product_sn, part_number, vp_app_version, vp_boot_loader_version, vp_core_version,
		supplier_hardware_version, manufacturer_hardware_version, manufacturer_software_version,
		ble_mac, ble_sn, ble_version, ble_passwork_key,
		ap_app_version, ap_kernel_version, tcu_iccid, phone_number, imei, imsi, production_date
		FROM logistic_data WHERE pcba_number = $1`

	row := r.db.QueryRow(query, pcba)
	var data models.LogisticDataDTO
	var id int
	err := row.Scan(
		&id,
		&data.PCBANumber, &data.ProductSN, &data.PartNumber, &data.VPAppVersion, &data.VPBootLoaderVersion, &data.VPCoreVersion,
		&data.SupplierHardwareVersion, &data.ManufacturerHardwareVersion, &data.ManufacturerSoftwareVersion,
		&data.BleMac, &data.BleSN, &data.BleVersion, &data.BlePassworkKey,
		&data.APAppVersion, &data.APKernelVersion, &data.TcuICCID, &data.PhoneNumber, &data.IMEI, &data.IMSI, &data.ProductionDate,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &data, nil
}

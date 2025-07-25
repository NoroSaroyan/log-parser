package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/NoroSaroyan/log-parser/internal/domain/models/db"
	"github.com/NoroSaroyan/log-parser/internal/domain/repositories"
)

// logisticDataRepository provides methods to perform CRUD operations
// on the logistic_data table.
type logisticDataRepository struct {
	db *sql.DB
}

// NewLogisticDataRepository creates a new instance of logisticDataRepository.
func NewLogisticDataRepository(db *sql.DB) *logisticDataRepository {
	return &logisticDataRepository{db: db}
}

// Insert inserts a new LogisticDataDB record into the logistic_data table.
// It updates the ID field of the provided model with the generated database ID.
// Returns an error if the insert fails.
func (r *logisticDataRepository) Insert(ctx context.Context, d *db.LogisticDataDB) error {
	query := `
        INSERT INTO logistic_data 
        (pcba_number, product_sn, part_number, vp_app_version, vp_boot_loader_version, vp_core_version,
        supplier_hardware_version, manufacturer_hardware_version, manufacturer_software_version,
        ble_mac, ble_sn, ble_version, ble_passwork_key, ap_app_version, ap_kernel_version,
        tcu_iccid, phone_number, imei, imsi, production_date)
        VALUES
        ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)
        RETURNING id
    `
	params := []interface{}{
		d.PCBANumber, d.ProductSN, d.PartNumber, d.VPAppVersion, d.VPBootLoaderVersion, d.VPCoreVersion,
		d.SupplierHardwareVersion, d.ManufacturerHardwareVersion, d.ManufacturerSoftwareVersion,
		d.BleMac, d.BleSN, d.BleVersion, d.BlePassworkKey, d.APAppVersion, d.APKernelVersion,
		d.TcuICCID, d.PhoneNumber, d.IMEI, d.IMSI, d.ProductionDate,
	}

	var id int
	err := r.db.QueryRowContext(ctx, query, params...).Scan(&id)
	if err != nil {
		return fmt.Errorf("failed to insert LogisticData and retrieve ID: %w", err)
	}
	if id == 0 {
		return fmt.Errorf("unexpected: inserted LogisticData returned ID=0")
	}
	d.ID = id
	return nil
}

// GetByPartNumber retrieves all LogisticDataDB records matching the given part number.
// Returns a slice of LogisticDataDB pointers or an error.
func (r *logisticDataRepository) GetByPartNumber(ctx context.Context, partNumber string) ([]*db.LogisticDataDB, error) {
	query := `
	SELECT pcba_number, product_sn, part_number, vp_app_version, vp_boot_loader_version, vp_core_version,
	       supplier_hardware_version, manufacturer_hardware_version, manufacturer_software_version,
	       ble_mac, ble_sn, ble_version, ble_passwork_key, ap_app_version, ap_kernel_version,
	       tcu_iccid, phone_number, imei, imsi, production_date
	FROM logistic_data
	WHERE part_number = $1
	`

	rows, err := r.db.QueryContext(ctx, query, partNumber)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*db.LogisticDataDB
	for rows.Next() {
		var d db.LogisticDataDB
		if err := rows.Scan(
			&d.PCBANumber, &d.ProductSN, &d.PartNumber, &d.VPAppVersion, &d.VPBootLoaderVersion, &d.VPCoreVersion,
			&d.SupplierHardwareVersion, &d.ManufacturerHardwareVersion, &d.ManufacturerSoftwareVersion,
			&d.BleMac, &d.BleSN, &d.BleVersion, &d.BlePassworkKey, &d.APAppVersion, &d.APKernelVersion,
			&d.TcuICCID, &d.PhoneNumber, &d.IMEI, &d.IMSI, &d.ProductionDate,
		); err != nil {
			return nil, err
		}
		results = append(results, &d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// GetAllByPCBANumber retrieves all LogisticDataDB records matching the given PCBA number.
// Returns a slice of LogisticDataDB pointers or an error.
func (r *logisticDataRepository) GetAllByPCBANumber(ctx context.Context, pcba string) ([]*db.LogisticDataDB, error) {
	query := `
    SELECT pcba_number, product_sn, part_number, vp_app_version, vp_boot_loader_version, vp_core_version,
    supplier_hardware_version, manufacturer_hardware_version, manufacturer_software_version,
    ble_mac, ble_sn, ble_version, ble_passwork_key, ap_app_version, ap_kernel_version,
    tcu_iccid, phone_number, imei, imsi, production_date
    FROM logistic_data
    WHERE pcba_number = $1
    `
	rows, err := r.db.QueryContext(ctx, query, pcba)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*db.LogisticDataDB
	for rows.Next() {
		var d db.LogisticDataDB
		if err := rows.Scan(
			&d.PCBANumber, &d.ProductSN, &d.PartNumber, &d.VPAppVersion, &d.VPBootLoaderVersion, &d.VPCoreVersion,
			&d.SupplierHardwareVersion, &d.ManufacturerHardwareVersion, &d.ManufacturerSoftwareVersion,
			&d.BleMac, &d.BleSN, &d.BleVersion, &d.BlePassworkKey, &d.APAppVersion, &d.APKernelVersion,
			&d.TcuICCID, &d.PhoneNumber, &d.IMEI, &d.IMSI, &d.ProductionDate,
		); err != nil {
			return nil, err
		}
		results = append(results, &d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return results, nil
}

// GetByPCBANumber retrieves the first LogisticDataDB record matching the given PCBA number.
// Returns a pointer to LogisticDataDB or nil if no record is found.
func (r *logisticDataRepository) GetByPCBANumber(ctx context.Context, pcba string) (*db.LogisticDataDB, error) {
	query := `
	SELECT pcba_number, product_sn, part_number, vp_app_version, vp_boot_loader_version, vp_core_version,
	       supplier_hardware_version, manufacturer_hardware_version, manufacturer_software_version,
	       ble_mac, ble_sn, ble_version, ble_passwork_key, ap_app_version, ap_kernel_version,
	       tcu_iccid, phone_number, imei, imsi, production_date
	FROM logistic_data
	WHERE pcba_number = $1
	LIMIT 1
	`

	var d db.LogisticDataDB
	err := r.db.QueryRowContext(ctx, query, pcba).Scan(
		&d.PCBANumber, &d.ProductSN, &d.PartNumber, &d.VPAppVersion, &d.VPBootLoaderVersion, &d.VPCoreVersion,
		&d.SupplierHardwareVersion, &d.ManufacturerHardwareVersion, &d.ManufacturerSoftwareVersion,
		&d.BleMac, &d.BleSN, &d.BleVersion, &d.BlePassworkKey, &d.APAppVersion, &d.APKernelVersion,
		&d.TcuICCID, &d.PhoneNumber, &d.IMEI, &d.IMSI, &d.ProductionDate,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &d, nil
}

// GetIDByPCBANumber retrieves the ID of a logistic_data record by PCBA number.
// Returns 0 if no record is found.
func (r *logisticDataRepository) GetIDByPCBANumber(ctx context.Context, pcba string) (int, error) {
	query := `
        SELECT id
        FROM logistic_data
        WHERE pcba_number = $1
        LIMIT 1
    `
	var id int
	err := r.db.QueryRowContext(ctx, query, pcba).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, err
		}
		return 0, err
	}
	return id, nil
}

// GetById retrieves a single LogisticDataDB record by its database ID.
// Returns a pointer to LogisticDataDB or nil if no record is found.
func (r *logisticDataRepository) GetById(ctx context.Context, id int) (*db.LogisticDataDB, error) {
	query := `
		SELECT 
			pcba_number, product_sn, part_number, vp_app_version, vp_boot_loader_version, vp_core_version,
			supplier_hardware_version, manufacturer_hardware_version, manufacturer_software_version,
			ble_mac, ble_sn, ble_version, ble_passwork_key, ap_app_version, ap_kernel_version,
			tcu_iccid, phone_number, imei, imsi, production_date 
		FROM logistic_data 
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var d db.LogisticDataDB
	err := row.Scan(
		&d.PCBANumber, &d.ProductSN, &d.PartNumber, &d.VPAppVersion, &d.VPBootLoaderVersion, &d.VPCoreVersion,
		&d.SupplierHardwareVersion, &d.ManufacturerHardwareVersion, &d.ManufacturerSoftwareVersion,
		&d.BleMac, &d.BleSN, &d.BleVersion, &d.BlePassworkKey, &d.APAppVersion, &d.APKernelVersion,
		&d.TcuICCID, &d.PhoneNumber, &d.IMEI, &d.IMSI, &d.ProductionDate,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &d, nil
}

// Ensure logisticDataRepository satisfies the LogisticDataRepository interface.
var _ repositories.LogisticDataRepository = (*logisticDataRepository)(nil)

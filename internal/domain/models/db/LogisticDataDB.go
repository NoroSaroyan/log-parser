package db

type LogisticDataDB struct {
	ID                          int    `db:"id"`
	PCBANumber                  string `db:"pcba_number"`
	ProductSN                   string `db:"product_sn"`
	PartNumber                  string `db:"part_number"`
	VPAppVersion                string `db:"vp_app_version"`
	VPBootLoaderVersion         string `db:"vp_boot_loader_version"`
	VPCoreVersion               string `db:"vp_core_version"`
	SupplierHardwareVersion     string `db:"supplier_hardware_version"`
	ManufacturerHardwareVersion string `db:"manufacturer_hardware_version"`
	ManufacturerSoftwareVersion string `db:"manufacturer_software_version"`
	BleMac                      string `db:"ble_mac"`
	BleSN                       string `db:"ble_sn"`
	BleVersion                  string `db:"ble_version"`
	BlePassworkKey              string `db:"ble_passwork_key"`
	APAppVersion                string `db:"ap_app_version"`
	APKernelVersion             string `db:"ap_kernel_version"`
	TcuICCID                    string `db:"tcu_iccid"`
	PhoneNumber                 string `db:"phone_number"`
	IMEI                        string `db:"imei"`
	IMSI                        string `db:"imsi"`
	ProductionDate              string `db:"production_date"`
}

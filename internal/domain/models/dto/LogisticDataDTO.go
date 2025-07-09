package dto

// LogisticDataDTO holds logistic info for a PCBA
//
// swagger:model
type LogisticDataDTO struct {
	PCBANumber                  string `json:"PCBANumber"`
	ProductSN                   string `json:"ProductSN"`
	PartNumber                  string `json:"PartNumber"`
	VPAppVersion                string `json:"VPAppVersion"`
	VPBootLoaderVersion         string `json:"VPBootLoaderVersion"`
	VPCoreVersion               string `json:"VPCoreVersion"`
	SupplierHardwareVersion     string `json:"SupplierHardwareVersion"`
	ManufacturerHardwareVersion string `json:"ManufacturerHardwareVersion"`
	ManufacturerSoftwareVersion string `json:"ManufacturerSoftwareVersion"`
	BleMac                      string `json:"BleMac"`
	BleSN                       string `json:"BleSN"`
	BleVersion                  string `json:"BleVersion"`
	BlePassworkKey              string `json:"BlePassworkKey"`
	APAppVersion                string `json:"APAppVersion"`
	APKernelVersion             string `json:"APKernelVersion"`
	TcuICCID                    string `json:"TcuICCID"`
	PhoneNumber                 string `json:"PhoneNumber"`
	IMEI                        string `json:"IMEI"`
	IMSI                        string `json:"IMSI"`
	ProductionDate              string `json:"ProductionDate"`
}

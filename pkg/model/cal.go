package model

import (
	"time"
)

// Cal contains the calibration data for a device.
type Cal struct {
	ID           int64     `db:"id" json:"-"`
	DeviceID     string    `db:"device_id" json:"deviceID"`
	SysID        uint64    `db:"sysid" json:"sysID"` // System id, CPU id or similar
	CollectionID string    `db:"collection_id" json:"collectionID"`
	ValidFrom    time.Time `db:"valid_from" json:"from"`

	// New fields
	CircuitType   string    `db:"circuit_type" json:"circuitType"`
	AFESerial     string    `db:"afe_serial" json:"afeSerial"`
	AFEType       string    `db:"afe_type" json:"afeType"`
	Sensor1Serial string    `db:"sensor1_serial" json:"sensor1Serial"`
	Sensor2Serial string    `db:"sensor2_serial" json:"sensor2Serial"`
	Sensor3Serial string    `db:"sensor3_serial" json:"sensor3Serial"`
	AFECalDate    time.Time `db:"afe_cal_date" json:"AFECalDate"` // When was the sensor calibrated
	Vt20Offset    float64   `db:"vt20_offset" json:"vt20Offset"`  // Temperature offset for probe at 20C

	Sensor1WEe           int32   `db:"sensor1_we_e" json:"sensor1WEe"`                     // Unit: mV
	Sensor1WE0           int32   `db:"sensor1_we_0" json:"sensor1WE0"`                     // Unit: mV
	Sensor1AEe           int32   `db:"sensor1_ae_e" json:"sensor1AEe"`                     // Unit: mV
	Sensor1AE0           int32   `db:"sensor1_ae_0" json:"sensor1AE0"`                     // Unit: mV
	Sensor1PCBGain       float64 `db:"sensor1_pcb_gain" json:"sensor1PCBGain"`             // Unit: mV / nA
	Sensor1WESensitivity float64 `db:"sensor1_we_sensitivity" json:"sensor1WESensitivity"` // Unit: mV / ppb

	Sensor2WEe           int32   `db:"sensor2_we_e" json:"sensor2WEe"`                     // Unit: mV
	Sensor2WE0           int32   `db:"sensor2_we_0" json:"sensor2WE0"`                     // Unit: mV
	Sensor2AEe           int32   `db:"sensor2_ae_e" json:"sensor2AEe"`                     // Unit: mV
	Sensor2AE0           int32   `db:"sensor2_ae_0" json:"sensor2AE0"`                     // Unit: mV
	Sensor2PCBGain       float64 `db:"sensor2_pcb_gain" json:"sensor2PCBGain"`             // Unit: mV / nA
	Sensor2WESensitivity float64 `db:"sensor2_we_sensitivity" json:"sensor2WESensitivity"` // Unit: mV / ppb

	Sensor3WEe           int32   `db:"sensor3_we_e" json:"sensor3WEe"`                     // Unit: mV
	Sensor3WE0           int32   `db:"sensor3_we_0" json:"sensor3WE0"`                     // Unit: mV
	Sensor3AEe           int32   `db:"sensor3_ae_e" json:"sensor3AEe"`                     // Unit: mV
	Sensor3AE0           int32   `db:"sensor3_ae_0" json:"sensor3AE0"`                     // Unit: mV
	Sensor3PCBGain       float64 `db:"sensor3_pcb_gain" json:"sensor3PCBGain"`             // Unit: mV / nA
	Sensor3WESensitivity float64 `db:"sensor3_we_sensitivity" json:"sensor3WESensitivity"` // Unit: mV / ppb
}

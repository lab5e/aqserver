package model

import (
	"github.com/sgreben/piecewiselinear"
)

type sensorLut struct {
	LUT []float64
}

var (
	afe3ScalingFactor = float64(0.0000005960464478) // named "lsb" in datasheet.

	// Temperatures used in the LUT
	afe3LutTemperatures = []float64{-30.0, -20.0, -10.0, 0.0, 10.0, 20.0, 30.0, 40.0, 50.0}

	// Lookup tables according to Appendix 1 of Alphasense Application
	// Note AAN 803.  Included the whole table although we only need 3
	// entries in case we will use any of these sensors in the future.
	afe3Luts = map[string]sensorLut{
		"CO-A4":  sensorLut{LUT: []float64{1.0, 1.0, 1.0, 1.0, 1.0, -1.0, -0.76, -0.76 - 0.76}},
		"CO2-B4": sensorLut{LUT: []float64{-1.0, -1.0, -1.0, -1.0, -1.0, -1.0, -3.8 - 3.8 - 3.8}},
		"NO-A4":  sensorLut{LUT: []float64{1.48, 1.48, 1.48, 1.48, 1.48, 2.02, 1.72, 1.72, 1.72}},
		"NO-B4":  sensorLut{LUT: []float64{1.04, 1.04, 1.04, 1.04, 1.04, 1.82, 2.0, 2.0, 2.0}},
		"NO2-A4": sensorLut{LUT: []float64{1.09, 1.09, 1.09, 1.09, 1.09, 1.35, 3.0, 3.0, 3.0}},
		"NO2-B4": sensorLut{LUT: []float64{0.76, 0.76, 0.76, 0.76, 0.76, 0.68, 0.23, 0.23, 0.23}},
		"SO2-A4": sensorLut{LUT: []float64{1.15, 1.15, 1.15, 1.15, 1.15, 1.82, 3.93, 3.93, 3.93}},
		"SO2-B4": sensorLut{LUT: []float64{0.96, 0.96, 0.96, 0.96, 0.96, 1.34, 1.10, 1.10, 1.10}},
		"O3-A4":  sensorLut{LUT: []float64{0.75, 0.75, 0.75, 0.75, 1.28, 1.28, 1.28, 1.28 /*, no value */}},
		"O3-B4":  sensorLut{LUT: []float64{0.77, 0.77, 0.77, 0.77, 1.56, 1.56, 1.56, 2.85 /*, no value */}},
	}

	// Create correction functions for the sensors we use
	correctSensor1 = correctionFuncFromName("NO-A4")
	correctSensor2 = correctionFuncFromName("NO2-A4")
	correctSensor3 = correctionFuncFromName("O3-A4")
)

// CalculateSensorValues calculates sensor values using measured data
// and calibration data specific to the the device.
func CalculateSensorValues(m *Message, cal *Cal) {

	// Calculate the temperature.
	// TODO(borud): have @tlan and @hansj double-check this
	m.AFE3TempValue = ((float64(m.AFE3TempRaw) * afe3ScalingFactor) - cal.Vt20Offset + 0.02) * 1000.0

	var sensor1TempCorrectionFactor = correctSensor1(m.AFE3TempValue)
	var sensor2TempCorrectionFactor = correctSensor1(m.AFE3TempValue)
	var sensor3TempCorrectionFactor = correctSensor1(m.AFE3TempValue)

	// Sensor 1 - NO2 sensor
	{
		wmV := voltage(m.Sensor1Work, cal.Sensor1WEe, cal.Sensor1WE0)
		amV := voltage(m.Sensor1Aux, cal.Sensor1AEe, cal.Sensor1AE0) * sensor1TempCorrectionFactor
		m.NO2PPB = (wmV - amV) / cal.Sensor1WESensitivity
	}

	// Sensor 2 - O3 + NO2 sensor, calculate O3 by subtracting NO2 sensor value
	{
		wmV := voltage(m.Sensor2Work, cal.Sensor2WEe, cal.Sensor2WE0)
		amV := voltage(m.Sensor2Aux, cal.Sensor2AEe, cal.Sensor2AE0) * sensor2TempCorrectionFactor
		m.O3PPB = ((wmV - amV) / cal.Sensor2WESensitivity) - m.NO2PPB
	}

	// Sensor 3 - NO sensor
	{
		wmV := voltage(m.Sensor3Work, cal.Sensor3WEe, cal.Sensor3WE0)
		amV := voltage(m.Sensor3Aux, cal.Sensor3AEe, cal.Sensor3AE0) * sensor3TempCorrectionFactor
		m.NOPPB = (wmV - amV) / cal.Sensor3WESensitivity
	}
}

func voltage(w uint32, offset int32, zero int32) float64 {
	return (float64(w) * afe3ScalingFactor * 1000) - float64(offset+zero)
}

func correctionFuncFromName(name string) func(float64) float64 {
	lut, ok := afe3Luts[name]
	if !ok {
		panic("Sensor name not found: " + name)
	}

	f := piecewiselinear.Function{Y: lut.LUT}
	f.X = afe3LutTemperatures[:len(lut.LUT)]

	return func(t float64) float64 {
		return f.At(t)
	}
}

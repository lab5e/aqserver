package model

import (
	"github.com/golang/protobuf/proto"
	aqv1 "github.com/lab5e/aqserver/pkg/aq/v1"
)

// MessageFromProtobuf takes a Sample protobuffer and returns a
// Message
func MessageFromProtobuf(s *aqv1.Sample) *Message {
	return &Message{
		// Board fields
		SysID:            s.Sysid,
		FirmwareVersion:  s.FirmwareVersion,
		Uptime:           s.Uptime,
		BoardTemp:        s.BoardTemp,
		BoardRelHumidity: s.BoardRelHumidity,
		Status:           s.Status,

		// GPS fields
		GPSTimeStamp: s.GpsTimestamp,
		Lat:          s.Lat,
		Lon:          s.Lon,
		Alt:          s.Alt,

		// AFE3 fields
		Sensor1Work: s.Sensor_1Work,
		Sensor1Aux:  s.Sensor_1Aux,
		Sensor2Work: s.Sensor_2Work,
		Sensor2Aux:  s.Sensor_2Aux,
		Sensor3Work: s.Sensor_3Work,
		Sensor3Aux:  s.Sensor_3Aux,
		AFE3TempRaw: s.Afe3TempRaw,

		// OPC-N3 fields
		OPCPMA:            s.OpcPmA,
		OPCPMB:            s.OpcPmB,
		OPCPMC:            s.OpcPmC,
		PM1:               s.Pm1,
		PM10:              s.Pm10,
		PM25:              s.Pm25,
		OPCSamplePeriod:   uint16(s.OpcSamplePeriod),
		OPCSampleFlowRate: uint16(s.OpcSampleFlowRate),
		OPCTemp:           uint16(s.OpcTemp),
		OPCHum:            uint16(s.OpcHum),
		OPCFanRevcount:    uint16(s.OpcFanRevcount),
		OPCLaserStatus:    uint16(s.OpcLaserStatus),
		OPCSampleValid:    uint8(s.OpcSampleValid),
		OPCBin0:           uint16(s.OpcBin_0),
		OPCBin1:           uint16(s.OpcBin_1),
		OPCBin2:           uint16(s.OpcBin_2),
		OPCBin3:           uint16(s.OpcBin_3),
		OPCBin4:           uint16(s.OpcBin_4),
		OPCBin5:           uint16(s.OpcBin_5),
		OPCBin6:           uint16(s.OpcBin_6),
		OPCBin7:           uint16(s.OpcBin_7),
		OPCBin8:           uint16(s.OpcBin_8),
		OPCBin9:           uint16(s.OpcBin_9),
		OPCBin10:          uint16(s.OpcBin_10),
		OPCBin11:          uint16(s.OpcBin_11),
		OPCBin12:          uint16(s.OpcBin_12),
		OPCBin13:          uint16(s.OpcBin_13),
		OPCBin14:          uint16(s.OpcBin_14),
		OPCBin15:          uint16(s.OpcBin_15),
		OPCBin16:          uint16(s.OpcBin_16),
		OPCBin17:          uint16(s.OpcBin_17),
		OPCBin18:          uint16(s.OpcBin_18),
		OPCBin19:          uint16(s.OpcBin_19),
		OPCBin20:          uint16(s.OpcBin_20),
		OPCBin21:          uint16(s.OpcBin_21),
		OPCBin22:          uint16(s.OpcBin_22),
		OPCBin23:          uint16(s.OpcBin_23),
	}
}

// ProtobufFromData unmarshals a protobuffer from a byte slice
func ProtobufFromData(buf []byte) (*aqv1.Sample, error) {
	pb := aqv1.Sample{}
	return &pb, proto.Unmarshal(buf, &pb)
}

// DataFromProtobuf marshals a aqv1.Sample into a byte slice
func DataFromProtobuf(pb *aqv1.Sample) ([]byte, error) {
	return proto.Marshal(pb)
}

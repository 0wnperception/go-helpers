package wbmioConnectorModbus

import (
	"time"

	modbus "github.com/goburrow/modbus"
)

const (
	COIL_ON  = 0xFF00
	COIL_OFF = 0x0000
)
const (
	INPUT_START_ADDR  = 1000
	OUTPUT_START_ADDR = 2500
)

type WBMIO struct {
	deviceName    string
	modbusClient  modbus.Client
	modbusHandler *modbus.RTUClientHandler
}

func NewWBMIO(deviceName, serial string, addr byte) *WBMIO {
	handler := modbus.NewRTUClientHandler(serial)
	handler.BaudRate = 9600
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 2
	handler.SlaveId = addr
	handler.Timeout = 5 * time.Second

	client := modbus.NewClient(handler)
	return &WBMIO{
		deviceName:    deviceName,
		modbusClient:  client,
		modbusHandler: handler,
	}
}

func (wb *WBMIO) SetState(idx uint16, state bool) (err error) {
	if state {
		_, err = wb.modbusClient.WriteSingleCoil(OUTPUT_START_ADDR+idx, COIL_ON)
	} else {
		_, err = wb.modbusClient.WriteSingleCoil(OUTPUT_START_ADDR+idx, COIL_OFF)
	}
	return
}

func (wb *WBMIO) GetState(idx uint16) (in bool, err error) {
	res, err := wb.modbusClient.ReadCoils(INPUT_START_ADDR+idx, 1)
	if err != nil || len(res) == 0 {
		return false, err
	}
	return res[0] > 0, err
}

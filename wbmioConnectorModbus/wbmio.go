package wbmioConnectorModbus

import (
	"context"
	"go-helpers/concurrent"
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
	concurrent    *concurrent.Concurrent
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
		concurrent:    concurrent.NewConcurrent(concurrent.ConcurrentConfig{SimCapacity: 1}),
		deviceName:    deviceName,
		modbusClient:  client,
		modbusHandler: handler,
	}
}

func (wb *WBMIO) SetState(ctx context.Context, idx uint16, state bool) (err error) {
	if ok := wb.concurrent.Borrow(ctx); ok {
		if state {
			_, err = wb.modbusClient.WriteSingleCoil(OUTPUT_START_ADDR+idx, COIL_ON)
		} else {
			_, err = wb.modbusClient.WriteSingleCoil(OUTPUT_START_ADDR+idx, COIL_OFF)
		}
		wb.concurrent.SettleUp()
	}
	return
}

func (wb *WBMIO) GetState(ctx context.Context, idx uint16) (in bool, err error) {
	if ok := wb.concurrent.Borrow(ctx); ok {
		var res []byte
		if res, err = wb.modbusClient.ReadCoils(INPUT_START_ADDR+idx, 1); err == nil {
			if len(res) > 0 {
				in = res[0] > 0
			}
		}
		wb.concurrent.SettleUp()
	}
	return
}

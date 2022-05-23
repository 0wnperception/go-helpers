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

	INPUT_START_ADDR  = 1000
	OUTPUT_START_ADDR = 2500

	RECONNECT_ATTEMPTS = 10
)

type WBMIO struct {
	deviceName    string
	modbusClient  modbus.Client
	modbusHandler *modbus.RTUClientHandler
	concurrent    *concurrent.Concurrent
}

func NewWBMIO(deviceName, serial string, addr byte, serialConcurrent *concurrent.Concurrent) *WBMIO {
	handler := modbus.NewRTUClientHandler(serial)
	handler.BaudRate = 9600
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 2
	handler.SlaveId = addr
	handler.Timeout = 100 * time.Millisecond

	client := modbus.NewClient(handler)
	return &WBMIO{
		concurrent:    serialConcurrent,
		deviceName:    deviceName,
		modbusClient:  client,
		modbusHandler: handler,
	}
}

func (wb *WBMIO) SetState(ctx context.Context, idx uint16, state bool) (err error) {
	if ok := wb.concurrent.Borrow(ctx); ok {
		for i := 0; i < RECONNECT_ATTEMPTS; i++ {
			// log.Printf("wbmio %s set attempt %d", wb.deviceName, i+1)
			if state {
				_, err = wb.modbusClient.WriteSingleCoil(OUTPUT_START_ADDR+idx, COIL_ON)
			} else {
				_, err = wb.modbusClient.WriteSingleCoil(OUTPUT_START_ADDR+idx, COIL_OFF)
			}
			if err == nil {
				break
			}
			time.Sleep(3 * time.Millisecond)
		}
		wb.concurrent.SettleUp()
	}
	return
}

func (wb *WBMIO) GetState(ctx context.Context, idx uint16) (in bool, err error) {
	if ok := wb.concurrent.Borrow(ctx); ok {
		for i := 0; i < RECONNECT_ATTEMPTS; i++ {
			// log.Printf("wbmio %s get attempt %d", wb.deviceName, i+1)
			var res []byte
			if res, err = wb.modbusClient.ReadCoils(INPUT_START_ADDR+idx, 1); err == nil {
				if len(res) > 0 {
					in = res[0] > 0
				}
			}
			if err == nil {
				break
			}
			time.Sleep(7 * time.Millisecond)
		}
		wb.concurrent.SettleUp()
	}
	return
}

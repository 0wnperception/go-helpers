package epr60ConnectorModbus

import (
	"context"
	"errors"
	"fmt"
	"time"

	modbus "github.com/things-go/go-modbus"
)

type EPR60SetupConfig struct {
	CurrentmAmps uint16
	EmergencyDec uint16
}

type EPR60 struct {
	deviceName    string
	registers     EPR60Registers
	controlParams EPR60ControlParams
	modbusClient  modbus.Client
}

type EPR60ControlParams struct {
	RequestedPos int
}

type EPR60Registers struct {
	cfg       EPR60RegistersConfig
	registers EPR60RegistersVal
}

type EPR60RegistersVal struct {
	AbsolutePosLow,
	AbsolutePosHigh,
	ModeControl,
	EmergencyStopDec,
	Reverse,
	PosControlPosLow16bit,
	PosControlPosHigh16bit,
	PosControlAcc,
	PosControlDec,
	PosControlSpeed,
	PositionAbsolute,
	SpeedControlAcc, SpeedControlDec, SpeedControlSpeed,
	ClearPos,
	CurrentSettings uint16
	InputState1, InputState2, InputState3, InputState4, InputState5, InputState6,
	StateRunning bool
}

func NewEPR60(deviceName string, addr string) *EPR60 {
	opts := func(p modbus.ClientProvider) {
		p.LogMode(false)
	}
	p := modbus.NewTCPClientProvider(addr, opts)
	client := modbus.NewClient(p)
	return &EPR60{
		deviceName:   deviceName,
		modbusClient: client,
		registers:    EPR60Registers{cfg: NewEPR60RegistersConfig()},
	}
}

func (e *EPR60) Connect(cfg EPR60SetupConfig) error {
	if err := e.modbusClient.Connect(); err != nil {
		return err
	}
	if err := e.Setup(cfg); err != nil {
		return err
	}
	return nil
}

func (e *EPR60) Setup(cfg EPR60SetupConfig) error {
	var inputs []uint16
	for i := 0; i < INPUTS_AMOUNT; i++ {
		inputs = append(inputs, uint16(EPR60_INPUT_MODE_DEFAULT))
	}
	if err := e.SetupInputs(inputs); err != nil {
		return err
	}
	if err := e.SetupCurrent(cfg.CurrentmAmps); err != nil {
		return err
	}
	if err := e.SetupPresetModeControl(); err != nil {
		return err
	}
	if err := e.SetupOperatingMode(); err != nil {
		return err
	}
	if err := e.SetupEmergencyDec(cfg.EmergencyDec); err != nil {
		return err
	}
	return nil
}

func (e *EPR60) Release() {
	e.modbusClient.Close()
}

func (e *EPR60) PositionMove(ctx context.Context, pos int, speed, acc uint16, dir bool) error {
	if err := e.SetPosConfig(pos, speed, acc, dir); err != nil {
		return err
	}
	if err := e.RunPosConfig(); err != nil {
		return err
	}
	if err := e.CheckPosConfig(ctx); err != nil {
		return err
	}
	return nil
}

func (e *EPR60) CheckPosConfig(ctx context.Context) error {
	pollTicker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-pollTicker.C:
			if err := e.PollCheckPosCompleted(); err != nil {
				return err
			}
			if e.controlParams.RequestedPos == e.GetAbsolutePos() {
				return nil
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (e *EPR60) RunPosConfig() error {
	return e.modbusClient.WriteSingleRegister(1, e.registers.cfg.ModeControlAddr, uint16(EPR60_MODE_CONTROL_POS_CONTROL))
}

func (e *EPR60) SetPosConfig(pos int, speed, acc uint16, dir bool) error {
	lpos := uint16(pos & 0xffff)
	hpos := uint16(pos >> 16)
	edir := uint16(0)
	if dir {
		edir = 1
	}

	if err := e.modbusClient.WriteMultipleRegisters(1, e.registers.cfg.PosControlAccAddr, 5, []uint16{acc, acc, speed, lpos, hpos}); err != nil {
		return err
	}
	if err := e.modbusClient.WriteSingleRegister(1, e.registers.cfg.ReverseAddr, edir); err != nil {
		return err
	}
	if err := e.modbusClient.WriteSingleRegister(1, e.registers.cfg.PositionAbsoluteAddr, 1); err != nil {
		return err
	}
	e.controlParams.RequestedPos = pos
	return nil
}

func (e *EPR60) ClearPos() error {
	return e.modbusClient.WriteSingleRegister(1, e.registers.cfg.ClearPosAddr, uint16(1))
}

func (e *EPR60) PollCheckPosCompleted() error {
	if poll8_9, err := e.modbusClient.ReadHoldingRegisters(1, e.registers.cfg.AbsolutePosLowAddr, 2); len(poll8_9) < 2 || err != nil {
		return err
	} else {
		e.registers.registers.AbsolutePosLow = poll8_9[0]
		e.registers.registers.AbsolutePosHigh = poll8_9[1]
	}
	if poll18, err := e.modbusClient.ReadHoldingRegisters(1, e.registers.cfg.ModeControlAddr, 1); len(poll18) < 1 || err != nil {
		return err
	} else {
		e.registers.registers.ModeControl = poll18[0]
	}
	return nil
}

func (e *EPR60) SpeedMove(speed, acc uint16, dir bool) error {
	if err := e.SetSpeedConfig(speed, acc, dir); err != nil {
		panic(err)
	}
	return e.RunSpeedConfig()
}

func (e *EPR60) SetSpeedConfig(speed, acc uint16, dir bool) error {
	edir := uint16(0)
	if dir {
		edir = 1
	}
	if err := e.modbusClient.WriteMultipleRegisters(1, e.registers.cfg.SpeedControlAccAddr, 3, []uint16{acc, acc, speed}); err != nil {
		return err
	}
	if err := e.modbusClient.WriteSingleRegister(1, e.registers.cfg.ReverseAddr, edir); err != nil {
		return err
	}
	return nil
}

func (e *EPR60) RunSpeedConfig() error {
	return e.modbusClient.WriteSingleRegister(1, e.registers.cfg.ModeControlAddr, uint16(EPR60_MODE_CONTROL_SPEED_CONTROL))
}

func (e *EPR60) DecStop(ctx context.Context) error {
	if err := e.modbusClient.WriteSingleRegister(1, e.registers.cfg.ModeControlAddr, uint16(EPR60_MODE_CONTROL_DECELERATION_STOP)); err != nil {
		return err
	}
	return e.CheckStoped(ctx)
}

func (e *EPR60) EmergencyStop(ctx context.Context) error {
	if err := e.modbusClient.WriteSingleRegister(1, e.registers.cfg.ModeControlAddr, uint16(EPR60_MODE_CONTROL_EMERGENCY_STOP)); err != nil {
		return err
	}
	return e.CheckStoped(ctx)
}

func (e *EPR60) DecStopAxisOnSensor(ctx context.Context, idx uint16, state bool) (err error) {
	pollTicker := time.NewTicker(1 * time.Millisecond)
	for {
		select {
		case <-pollTicker.C:
			if sensor, tmperr := e.GetInputState(idx); tmperr != nil {
				err = tmperr
				return
			} else {
				if sensor == state {
					e.DecStop(ctx)
					return
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (e *EPR60) EmergencyStopAxisOnSensor(ctx context.Context, idx uint16, state bool) (err error) {
	pollTicker := time.NewTicker(1 * time.Millisecond)
	for {
		select {
		case <-pollTicker.C:
			if sensor, tmperr := e.GetInputState(idx); tmperr != nil {
				err = tmperr
				return
			} else {
				if sensor == state {
					e.EmergencyStop(ctx)
					return
				}
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (e *EPR60) CheckStoped(ctx context.Context) error {
	pollTicker := time.NewTicker(1 * time.Millisecond)
	for {
		select {
		case <-pollTicker.C:
			if err := e.PollRunning(); err != nil {
				return err
			}
			if !e.registers.registers.StateRunning {
				return nil
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (e *EPR60) PollRunning() (err error) {
	if poll1, err := e.modbusClient.ReadHoldingRegisters(1, e.registers.cfg.StateRunningAddr, 1); err == nil {
		e.registers.registers.StateRunning = (poll1[0] & (1 << EPR60_RUNNING_BIT)) == (1 << EPR60_RUNNING_BIT)
	}
	return
}

func (e *EPR60) SetupCurrent(mAmps uint16) (err error) {
	return e.modbusClient.WriteSingleRegister(1, e.registers.cfg.CurrentSettingsAddr, mAmps)
}

func (e *EPR60) SetupPresetModeControl() (err error) {
	return e.modbusClient.WriteSingleRegister(1, e.registers.cfg.PresetModeControlAddr, 0)
}

func (e *EPR60) SetupOperatingMode() (err error) {
	return e.modbusClient.WriteSingleRegister(1, e.registers.cfg.MotorOperatingModeAddr, uint16(EPR60_OPERATING_MODE_OPENLOOP))
}

func (e *EPR60) SetupEmergencyDec(dec uint16) (err error) {
	return e.modbusClient.WriteSingleRegister(1, e.registers.cfg.EmergencyStopDecAddr, dec)
}

func (e *EPR60) SetupInputs(modes []uint16) (err error) {
	return e.modbusClient.WriteMultipleRegisters(1, e.registers.cfg.SettingInput1Addr, INPUTS_AMOUNT, modes)
}

func (e *EPR60) GetAbsolutePos() int {
	return (int(e.registers.registers.AbsolutePosHigh) << 16) + int(e.registers.registers.AbsolutePosLow)
}

func (e *EPR60) GetName() string {
	return e.deviceName
}

func (e *EPR60) GetInputState(idx uint16) (in bool, err error) {
	if err = e.PollInput(); err == nil {
		switch idx {
		case 1:
			in = e.registers.registers.InputState1
		case 2:
			in = e.registers.registers.InputState2
		case 3:
			in = e.registers.registers.InputState3
		case 4:
			in = e.registers.registers.InputState4
		case 5:
			in = e.registers.registers.InputState5
		case 6:
			in = e.registers.registers.InputState6
		default:
			err = errors.New(fmt.Sprintf("no such input with idx %d", idx))
		}
	}
	return
}

func (e *EPR60) PollInput() (err error) {
	if poll1_2, err := e.modbusClient.ReadHoldingRegisters(1, e.registers.cfg.InputStateAddr, 1); err == nil {
		e.registers.registers.InputState1 = (poll1_2[0] & (1 << EPR60_INPUT_BIT_1)) == 0
		e.registers.registers.InputState2 = (poll1_2[0] & (1 << EPR60_INPUT_BIT_2)) == 0
		e.registers.registers.InputState3 = (poll1_2[0] & (1 << EPR60_INPUT_BIT_3)) == 0
		e.registers.registers.InputState4 = (poll1_2[0] & (1 << EPR60_INPUT_BIT_4)) == 0
		e.registers.registers.InputState5 = (poll1_2[0] & (1 << EPR60_INPUT_BIT_5)) == 0
		e.registers.registers.InputState6 = (poll1_2[0] & (1 << EPR60_INPUT_BIT_6)) == 0
	}
	return
}

/*
func (e *EPR60) PollRegisters() error {
	if poll1_2, err := e.modbusClient.ReadHoldingRegisters(1, e.registers.cfg.StateRunningAddr, 2); len(poll1_2) < 2 || err != nil {
		return err
	} else {
		e.registers.registers.StateRunning = poll1_2[0]
		e.registers.registers.InputState1 = (poll1_2[1] & (1 << EPR60_INPUT_BIT_1))
		e.registers.registers.InputState2 = (poll1_2[1] & (1 << EPR60_INPUT_BIT_2))
		e.registers.registers.InputState3 = (poll1_2[1] & (1 << EPR60_INPUT_BIT_3))
		e.registers.registers.InputState4 = (poll1_2[1] & (1 << EPR60_INPUT_BIT_4))
		e.registers.registers.InputState5 = (poll1_2[1] & (1 << EPR60_INPUT_BIT_5))
		e.registers.registers.InputState6 = (poll1_2[1] & (1 << EPR60_INPUT_BIT_6))
	}
	if poll8_9, err := e.modbusClient.ReadHoldingRegisters(1, e.registers.cfg.AbsolutePosLowAddr, 2); len(poll8_9) < 2 || err != nil {
		return err
	} else {
		e.registers.registers.AbsolutePosLow = poll8_9[0]
		e.registers.registers.AbsolutePosHigh = poll8_9[1]
	}
	if poll18, err := e.modbusClient.ReadHoldingRegisters(1, e.registers.cfg.ModeControlAddr, 1); len(poll18) < 1 || err != nil {
		return err
	} else {
		e.registers.registers.ModeControl = poll18[0]
	}
	if poll23, err := e.modbusClient.ReadHoldingRegisters(1, e.registers.cfg.ReverseAddr, 1); len(poll23) < 1 || err != nil {
		return err
	} else {
		e.registers.registers.Reverse = poll23[0]
	}

	if poll70_74, err := e.modbusClient.ReadHoldingRegisters(1, e.registers.cfg.PosControlAccAddr, 5); len(poll70_74) < 5 || err != nil {
		return err
	} else {
		e.registers.registers.PosControlAcc = poll70_74[0]
		e.registers.registers.PosControlDec = poll70_74[1]
		e.registers.registers.PosControlSpeed = poll70_74[2]
		e.registers.registers.PosControlPosLow16bit = poll70_74[3]
		e.registers.registers.PosControlPosHigh16bit = poll70_74[4]
	}
	if poll84, err := e.modbusClient.ReadHoldingRegisters(1, e.registers.cfg.PositionAbsoluteAddr, 1); len(poll84) < 1 || err != nil {
		return err
	} else {
		e.registers.registers.PositionAbsolute = poll84[0]
	}

	return nil
}
*/

package epr60ConnectorModbus

import (
	"context"
	"errors"
	"fmt"
	"go-helpers/convertions"
	"time"

	modbus "github.com/goburrow/modbus"
)

type EPR60 struct {
	cfg           EPR60Config
	modbusClient  modbus.Client
	registers     EPR60Registers
	controlParams EPR60ControlParams
}

type EPR60Config struct {
	DeviceName   string
	Addr         string
	CurrentmAmps uint16
	EmergencyDec uint16
}

type EPR60ControlParams struct {
	RequestedPos int
}

type EPR60Registers struct {
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

func NewEPR60(cfg EPR60Config) (e *EPR60, err error) {
	p := modbus.NewTCPClientHandler(cfg.Addr)
	client := modbus.NewClient(p)
	e = &EPR60{
		cfg:          cfg,
		modbusClient: client,
	}
	err = e.Setup(cfg)
	return
}

func (e *EPR60) Setup(cfg EPR60Config) error {
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

func (e *EPR60) RunPosConfig() (err error) {
	_, err = e.modbusClient.WriteSingleRegister(ModeControlAddr, uint16(EPR60_MODE_CONTROL_POS_CONTROL))
	return
}

func (e *EPR60) SetPosConfig(pos int, speed, acc uint16, dir bool) error {
	lpos := uint16(pos & 0xffff)
	hpos := uint16(pos >> 16)
	edir := uint16(0)
	if dir {
		edir = 1
	}

	if _, err := e.modbusClient.WriteMultipleRegisters(PosControlAccAddr, 5, convertions.Uint162Bytes(acc, acc, speed, lpos, hpos)); err != nil {
		return err
	}
	if _, err := e.modbusClient.WriteSingleRegister(ReverseAddr, edir); err != nil {
		return err
	}
	if _, err := e.modbusClient.WriteSingleRegister(PositionAbsoluteAddr, 1); err != nil {
		return err
	}
	e.controlParams.RequestedPos = pos
	return nil
}

func (e *EPR60) ClearPos() (err error) {
	_, err = e.modbusClient.WriteSingleRegister(ClearPosAddr, uint16(1))
	return
}

func (e *EPR60) PollCheckPosCompleted() error {
	if binPoll8_9, err := e.modbusClient.ReadHoldingRegisters(AbsolutePosLowAddr, 2); len(binPoll8_9) < 4 || err != nil {
		return err
	} else {
		poll8_9 := convertions.Bytes2Uint16(binPoll8_9)
		e.registers.AbsolutePosLow = poll8_9[0]
		e.registers.AbsolutePosHigh = poll8_9[1]
	}
	if binPoll18, err := e.modbusClient.ReadHoldingRegisters(ModeControlAddr, 1); len(binPoll18) < 2 || err != nil {
		return err
	} else {
		poll18 := convertions.Bytes2Uint16(binPoll18)
		e.registers.ModeControl = poll18[0]
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
	if _, err := e.modbusClient.WriteMultipleRegisters(SpeedControlAccAddr, 3, convertions.Uint162Bytes(acc, acc, speed)); err != nil {
		return err
	}
	if _, err := e.modbusClient.WriteSingleRegister(ReverseAddr, edir); err != nil {
		return err
	}
	return nil
}

func (e *EPR60) RunSpeedConfig() (err error) {
	_, err = e.modbusClient.WriteSingleRegister(ModeControlAddr, uint16(EPR60_MODE_CONTROL_SPEED_CONTROL))
	return
}

func (e *EPR60) DecStop(ctx context.Context) error {
	if _, err := e.modbusClient.WriteSingleRegister(ModeControlAddr, uint16(EPR60_MODE_CONTROL_DECELERATION_STOP)); err != nil {
		return err
	}
	return e.CheckStoped(ctx)
}

func (e *EPR60) EmergencyStop(ctx context.Context) error {
	if _, err := e.modbusClient.WriteSingleRegister(ModeControlAddr, uint16(EPR60_MODE_CONTROL_EMERGENCY_STOP)); err != nil {
		return err
	}
	return e.CheckStoped(ctx)
}

func (e *EPR60) DecStopAxisOnSensor(ctx context.Context, getSensorState func(ctx context.Context) (in bool, err error), state bool) (err error) {
	pollTicker := time.NewTicker(1 * time.Millisecond)
	for {
		select {
		case <-pollTicker.C:
			if sensor, tmperr := getSensorState(ctx); tmperr != nil {
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

func (e *EPR60) EmergencyStopAxisOnSensor(ctx context.Context, getSensorState func(ctx context.Context) (in bool, err error), state bool) (err error) {
	pollTicker := time.NewTicker(1 * time.Millisecond)
	for {
		select {
		case <-pollTicker.C:
			if sensor, tmperr := getSensorState(ctx); tmperr != nil {
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
			if !e.registers.StateRunning {
				return nil
			}
		case <-ctx.Done():
			return nil
		}
	}
}

func (e *EPR60) PollRunning() (err error) {
	if binPoll1, err := e.modbusClient.ReadHoldingRegisters(StateRunningAddr, 1); err == nil {
		poll1 := convertions.Bytes2Uint16(binPoll1)
		e.registers.StateRunning = (poll1[0] & (1 << EPR60_RUNNING_BIT)) == (1 << EPR60_RUNNING_BIT)
	}
	return
}

func (e *EPR60) SetupCurrent(mAmps uint16) (err error) {
	_, err = e.modbusClient.WriteSingleRegister(CurrentSettingsAddr, mAmps)
	return
}

func (e *EPR60) SetupPresetModeControl() (err error) {
	_, err = e.modbusClient.WriteSingleRegister(PresetModeControlAddr, 0)
	return
}

func (e *EPR60) SetupOperatingMode() (err error) {
	_, err = e.modbusClient.WriteSingleRegister(MotorOperatingModeAddr, uint16(EPR60_OPERATING_MODE_OPENLOOP))
	return
}

func (e *EPR60) SetupEmergencyDec(dec uint16) (err error) {
	_, err = e.modbusClient.WriteSingleRegister(EmergencyStopDecAddr, dec)
	return
}

func (e *EPR60) SetupInputs(modes []uint16) (err error) {
	_, err = e.modbusClient.WriteMultipleRegisters(SettingInput1Addr, INPUTS_AMOUNT, convertions.Uint162Bytes(modes...))
	return
}

func (e *EPR60) GetAbsolutePos() int {
	return (int(e.registers.AbsolutePosHigh) << 16) + int(e.registers.AbsolutePosLow)
}

func (e *EPR60) GetName() string {
	return e.cfg.DeviceName
}

func (e *EPR60) GetInputState(idx uint16) (in bool, err error) {
	if err = e.PollInput(); err == nil {
		switch idx {
		case 1:
			in = e.registers.InputState1
		case 2:
			in = e.registers.InputState2
		case 3:
			in = e.registers.InputState3
		case 4:
			in = e.registers.InputState4
		case 5:
			in = e.registers.InputState5
		case 6:
			in = e.registers.InputState6
		default:
			err = errors.New(fmt.Sprintf("no such input with idx %d", idx))
		}
	}
	return
}

func (e *EPR60) PollInput() (err error) {
	if binPoll1_2, err := e.modbusClient.ReadHoldingRegisters(InputStateAddr, 1); err == nil {
		poll1_2 := convertions.Bytes2Uint16(binPoll1_2)
		e.registers.InputState1 = (poll1_2[0] & (1 << EPR60_INPUT_BIT_1)) == 0
		e.registers.InputState2 = (poll1_2[0] & (1 << EPR60_INPUT_BIT_2)) == 0
		e.registers.InputState3 = (poll1_2[0] & (1 << EPR60_INPUT_BIT_3)) == 0
		e.registers.InputState4 = (poll1_2[0] & (1 << EPR60_INPUT_BIT_4)) == 0
		e.registers.InputState5 = (poll1_2[0] & (1 << EPR60_INPUT_BIT_5)) == 0
		e.registers.InputState6 = (poll1_2[0] & (1 << EPR60_INPUT_BIT_6)) == 0
	}
	return
}

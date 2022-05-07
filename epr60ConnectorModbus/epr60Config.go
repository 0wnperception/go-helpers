package epr60ConnectorModbus

type EPR60Iface interface {
	PositionMove(pos int, speed, acc uint16, dir uint16)
	SpeedMove(speed, acc uint16, dir uint16) error
	DecStop() error
	EmergencyStop() error
	GetAbsolutePos() int
	GetInputState(idx int) (in bool, err error)
	GetName() string
}

type EPR60RegistersConfig struct {
	AbsolutePosLowAddr,
	AbsolutePosHighAddr,
	StateRunningAddr,
	ModeControlAddr,
	EmergencyStopDecAddr,
	ReverseAddr,
	InputStateAddr,
	PosControlPosLow16bitAddr,
	PosControlPosHigh16bitAddr,
	PosControlAccAddr,
	PosControlDecAddr,
	PosControlSpeedAddr,
	ClearPosAddr,
	PositionAbsoluteAddr,
	SpeedControlAccAddr,
	SpeedControlDecAddr,
	SpeedControlSpeedAddr,

	PresetModeControlAddr,
	MotorTypeSettingsAddr,
	MotorOperatingModeAddr,
	CurrentSettingsAddr,
	SettingInput1Addr,
	SettingInput2Addr,
	SettingInput3Addr,
	SettingInput4Addr uint16
}

func NewEPR60RegistersConfig() EPR60RegistersConfig {
	return EPR60RegistersConfig{
		StateRunningAddr:           1,
		InputStateAddr:             2,
		AbsolutePosLowAddr:         8,
		AbsolutePosHighAddr:        9,
		ModeControlAddr:            18,
		PresetModeControlAddr:      20,
		MotorTypeSettingsAddr:      21,
		MotorOperatingModeAddr:     22,
		ReverseAddr:                23,
		CurrentSettingsAddr:        25,
		SettingInput1Addr:          60,
		SettingInput2Addr:          61,
		SettingInput3Addr:          62,
		SettingInput4Addr:          63,
		PosControlAccAddr:          70,
		PosControlDecAddr:          71,
		PosControlSpeedAddr:        72,
		PosControlPosLow16bitAddr:  73,
		PosControlPosHigh16bitAddr: 74,
		SpeedControlAccAddr:        75,
		SpeedControlDecAddr:        76,
		SpeedControlSpeedAddr:      77,
		EmergencyStopDecAddr:       78,
		ClearPosAddr:               85,
		PositionAbsoluteAddr:       84,
	}
}

type EPR60_MODE_CONTROL int

const (
	EPR60_MODE_CONTROL_EXPECTATION       EPR60_MODE_CONTROL = 0
	EPR60_MODE_CONTROL_POS_CONTROL       EPR60_MODE_CONTROL = 1
	EPR60_MODE_CONTROL_SPEED_CONTROL     EPR60_MODE_CONTROL = 3
	EPR60_MODE_CONTROL_EMERGENCY_STOP    EPR60_MODE_CONTROL = 5
	EPR60_MODE_CONTROL_DECELERATION_STOP EPR60_MODE_CONTROL = 6
)

const INPUTS_AMOUNT = 6

type EPR60_INPUT_BIT int

const (
	EPR60_INPUT_BIT_1 EPR60_INPUT_BIT = 0
	EPR60_INPUT_BIT_2 EPR60_INPUT_BIT = 1
	EPR60_INPUT_BIT_3 EPR60_INPUT_BIT = 2
	EPR60_INPUT_BIT_4 EPR60_INPUT_BIT = 3
	EPR60_INPUT_BIT_5 EPR60_INPUT_BIT = 4
	EPR60_INPUT_BIT_6 EPR60_INPUT_BIT = 5
)

type EPR60_INPUT_MODE int

const (
	EPR60_INPUT_MODE_DEFAULT EPR60_INPUT_MODE = 54
)

type EPR60_OPERATING_MODE int

const (
	EPR60_OPERATING_MODE_OPENLOOP EPR60_OPERATING_MODE = 0
)

const (
	EPR60_RUNNING_BIT = 3
)

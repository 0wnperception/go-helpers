package epr60ConnectorModbus

const POLL_TIMEOUT = 1

const (
	StateRunningAddr           = 1
	InputStateAddr             = 2
	AbsolutePosLowAddr         = 8
	AbsolutePosHighAddr        = 9
	ModeControlAddr            = 18
	PresetModeControlAddr      = 20
	MotorTypeSettingsAddr      = 21
	MotorOperatingModeAddr     = 22
	ReverseAddr                = 23
	CurrentSettingsAddr        = 25
	SettingInput1Addr          = 60
	SettingInput2Addr          = 61
	SettingInput3Addr          = 62
	SettingInput4Addr          = 63
	SettingInput5Addr          = 64
	SettingInput6Addr          = 65
	PosControlAccAddr          = 70
	PosControlDecAddr          = 71
	PosControlSpeedAddr        = 72
	PosControlPosLow16bitAddr  = 73
	PosControlPosHigh16bitAddr = 74
	SpeedControlAccAddr        = 75
	SpeedControlDecAddr        = 76
	SpeedControlSpeedAddr      = 77
	EmergencyStopDecAddr       = 78
	ClearPosAddr               = 85
	PositionAbsoluteAddr       = 84

	MultiSegmentMode       = 221
	MultiSegmentNumber     = 222
	MultiSegmentTimeFormat = 223

	MultiSegmentStagePosLow1      = 125
	MultiSegmentStagePosHigh1     = 126
	MultiSegmentStageSpeed1       = 224
	MultiSegmentStageAcc1         = 225
	MultiSegmentStageWaitingTime1 = 226

	MultiSegmentStagePosLow2      = 127
	MultiSegmentStagePosHigh2     = 128
	MultiSegmentStageSpeed2       = 227
	MultiSegmentStageAcc2         = 228
	MultiSegmentStageWaitingTime2 = 229

	MultiSegmentStagePosLow3      = 129
	MultiSegmentStagePosHigh3     = 130
	MultiSegmentStageSpeed3       = 230
	MultiSegmentStageAcc3         = 231
	MultiSegmentStageWaitingTime3 = 232

	MultiSegmentStagePosLow4      = 131
	MultiSegmentStagePosHigh4     = 132
	MultiSegmentStageSpeed4       = 233
	MultiSegmentStageAcc4         = 234
	MultiSegmentStageWaitingTime4 = 235

	MultiSegmentStagePosLow5      = 133
	MultiSegmentStagePosHigh5     = 134
	MultiSegmentStageSpeed5       = 236
	MultiSegmentStageAcc5         = 237
	MultiSegmentStageWaitingTime5 = 238
)

const MULTISEGMENT_AMOUNT = 5

type EPR60_MULTISEGMENT_TIME_FORMAT uint16

const (
	EPR60_MULTISEGMENT_TIME_FORMAT_MS EPR60_MULTISEGMENT_TIME_FORMAT = 0
	EPR60_MULTISEGMENT_TIME_FORMAT_S  EPR60_MULTISEGMENT_TIME_FORMAT = 1
)

type EPR60_MULTISEGMENT_MODE uint16

const (
	EPR60_MULTISEGMENT_MODE_SINGLE EPR60_MULTISEGMENT_MODE = 0
	EPR60_MULTISEGMENT_MODE_CYCLE  EPR60_MULTISEGMENT_MODE = 1
	EPR60_MULTISEGMENT_MODE_INPUT  EPR60_MULTISEGMENT_MODE = 2
)

type EPR60_MODE_CONTROL uint16

const (
	EPR60_MODE_CONTROL_EXPECTATION       EPR60_MODE_CONTROL = 0
	EPR60_MODE_CONTROL_POS_CONTROL       EPR60_MODE_CONTROL = 1
	EPR60_MODE_CONTROL_SPEED_CONTROL     EPR60_MODE_CONTROL = 3
	EPR60_MODE_CONTROL_EMERGENCY_STOP    EPR60_MODE_CONTROL = 5
	EPR60_MODE_CONTROL_DECELERATION_STOP EPR60_MODE_CONTROL = 6
)

const INPUTS_AMOUNT = 3

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
	EPR60_INPUT_MODE_DEFAULT           EPR60_INPUT_MODE = 54
	EPR60_INPUT_MODE_MULTI_SEGMENT_ON  EPR60_INPUT_MODE = 56
	EPR60_INPUT_MODE_MULTI_SEGMENT_OFF EPR60_INPUT_MODE = 24
)

type EPR60_OPERATING_MODE int

const (
	EPR60_OPERATING_MODE_OPENLOOP EPR60_OPERATING_MODE = 0
)

const (
	EPR60_RUNNING_BIT = 3
)

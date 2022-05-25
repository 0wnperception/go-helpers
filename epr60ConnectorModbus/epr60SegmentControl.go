package epr60ConnectorModbus

import (
	"go-helpers/convertions"
	"log"
	"time"
)

type MultiSegmentOpts struct {
	Acc   uint16
	Speed uint16
	Pos   int
	Dir   bool
}

func (e *EPR60) SetMultisegmentConfig(opts ...MultiSegmentOpts) (err error) {
	amount := len(opts)
	if amount > MULTISEGMENT_AMOUNT {
		amount = MULTISEGMENT_AMOUNT
	}
	if amount > 0 {
		log.Println("write input setting")
		_, err = e.modbusClient.WriteSingleRegister(SettingInput3Addr, uint16(EPR60_INPUT_MODE_MULTI_SEGMENT_OFF))

		log.Println("write segment mode")
		if _, err := e.modbusClient.WriteMultipleRegisters(MultiSegmentMode, 3, convertions.Uint162Bytes(
			uint16(EPR60_MULTISEGMENT_MODE_SINGLE),
			uint16(amount),
			uint16(EPR60_MULTISEGMENT_TIME_FORMAT_MS),
		)); err != nil {
			return err
		}

		edir := uint16(0)
		if opts[0].Dir {
			edir = 1
		}
		log.Println("write segment dir")
		if _, err := e.modbusClient.WriteSingleRegister(ReverseAddr, edir); err != nil {
			return err
		}

		delay := uint16(1)

		for i := 0; i < len(opts) && i < MULTISEGMENT_AMOUNT; i++ {
			lpos := uint16(opts[i].Pos & 0xffff)
			hpos := uint16(opts[i].Pos >> 16)
			log.Printf("write segment %d opts %v", i, opts[i])
			switch i {
			case 0:
				if _, err := e.modbusClient.WriteMultipleRegisters(MultiSegmentStagePosLow1, 2, convertions.Uint162Bytes(
					uint16(lpos),
					uint16(hpos),
				)); err != nil {
					return err
				}

				if _, err := e.modbusClient.WriteMultipleRegisters(MultiSegmentStageSpeed1, 3, convertions.Uint162Bytes(
					opts[i].Speed,
					opts[i].Acc,
					delay,
				)); err != nil {
					return err
				}

			case 1:
				if _, err := e.modbusClient.WriteMultipleRegisters(MultiSegmentStagePosLow2, 2, convertions.Uint162Bytes(
					uint16(lpos),
					uint16(hpos),
				)); err != nil {
					return err
				}

				if _, err := e.modbusClient.WriteMultipleRegisters(MultiSegmentStageSpeed2, 3, convertions.Uint162Bytes(
					opts[i].Speed,
					opts[i].Acc,
					delay,
				)); err != nil {
					return err
				}
			case 2:
				if _, err := e.modbusClient.WriteMultipleRegisters(MultiSegmentStagePosLow3, 2, convertions.Uint162Bytes(
					uint16(lpos),
					uint16(hpos),
				)); err != nil {
					return err
				}

				if _, err := e.modbusClient.WriteMultipleRegisters(MultiSegmentStageSpeed3, 3, convertions.Uint162Bytes(
					opts[i].Speed,
					opts[i].Acc,
					delay,
				)); err != nil {
					return err
				}
			case 3:
				if _, err := e.modbusClient.WriteMultipleRegisters(MultiSegmentStagePosLow4, 2, convertions.Uint162Bytes(
					uint16(lpos),
					uint16(hpos),
				)); err != nil {
					return err
				}

				if _, err := e.modbusClient.WriteMultipleRegisters(MultiSegmentStageSpeed4, 3, convertions.Uint162Bytes(
					opts[i].Speed,
					opts[i].Acc,
					delay,
				)); err != nil {
					return err
				}
			case 4:
				if _, err := e.modbusClient.WriteMultipleRegisters(MultiSegmentStagePosLow5, 2, convertions.Uint162Bytes(
					uint16(lpos),
					uint16(hpos),
				)); err != nil {
					return err
				}

				if _, err := e.modbusClient.WriteMultipleRegisters(MultiSegmentStageSpeed5, 3, convertions.Uint162Bytes(
					opts[i].Speed,
					opts[i].Acc,
					delay,
				)); err != nil {
					return err
				}
			}
		}
	}

	return
}

func (e *EPR60) RunMultisegmentConfig() (err error) {
	log.Printf("write input run ")
	_, err = e.modbusClient.WriteSingleRegister(SettingInput6Addr, uint16(EPR60_INPUT_MODE_MULTI_SEGMENT_ON))
	time.Sleep(100 * time.Millisecond)
	_, err = e.modbusClient.WriteSingleRegister(SettingInput6Addr, uint16(EPR60_INPUT_MODE_MULTI_SEGMENT_OFF))
	return
}

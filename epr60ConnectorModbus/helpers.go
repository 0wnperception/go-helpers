package epr60ConnectorModbus

import (
	"context"
	"log"
	"time"
)

func (e *EPR60) CalibrateFast(ctx context.Context,
	GetSensorState func(ctx context.Context) (in bool, err error),
	Acc uint16,
	Speed uint16,
	Dir bool) (err error) {

	log.Printf("start calibration %s", e.GetName())
	log.Println("check calibration sensor..")
	sensor, err := GetSensorState(ctx)
	if err != nil {
		return err
	}
	log.Printf("calib sensor %v", sensor)

	if !sensor {
		log.Printf("run %s forward..", e.GetName())
		if err := e.SpeedMove(Speed, Acc, Dir); err != nil {
			return err
		}
		log.Printf("wait %s calibration sensor..", e.GetName())
		if err = e.EmergencyStopAxisOnSensor(ctx, GetSensorState, true); err != nil {
			return err
		}
	}
	e.ClearPos()
	log.Printf("%s calibration completed ", e.GetName())
	return
}

func (e *EPR60) CalibrateSlow(ctx context.Context,
	GetSensorState func(ctx context.Context) (in bool, err error),
	Acc uint16,
	SpeedHigh uint16,
	SpeedLow uint16,
	Dir bool) (err error) {

	log.Printf("start calibration %s", e.GetName())
	log.Println("check calibration sensor..")
	sensor, err := GetSensorState(ctx)
	if err != nil {
		return err
	}
	log.Printf("calib sensor %v", sensor)

	if !sensor {
		log.Printf("run %s forward..", e.GetName())
		if err := e.SpeedMove(SpeedHigh, Acc, Dir); err != nil {
			return err
		}
		log.Printf("wait %s calibration sensor..", e.GetName())
		if err = e.EmergencyStopAxisOnSensor(ctx, GetSensorState, true); err != nil {
			return err
		}
	}

	log.Printf("run %s back..", e.GetName())
	if err = e.SpeedMove(SpeedLow, Acc, !Dir); err != nil {
		return err
	}
	log.Printf("wait %s calibration sensor..", e.GetName())
	if err = e.EmergencyStopAxisOnSensor(ctx, GetSensorState, false); err != nil {
		return err
	}
	e.ClearPos()
	log.Printf("%s calibration completed", e.GetName())
	return
}

type SyncOpts struct {
	Axis  *EPR60
	Acc   uint16
	Speed uint16
	Pos   int
	Dir   bool
}

func (e *EPR60) PositionMoveSync(ctx context.Context, acc, speed uint16, pos int, dir bool, opts ...SyncOpts) (err error) {
	//setting config of movement
	if err = e.SetPosConfig(acc, speed, pos, dir); err != nil {
		return
	}
	for _, opt := range opts {
		if err = opt.Axis.SetPosConfig(opt.Acc, opt.Speed, opt.Pos, opt.Dir); err != nil {
			return
		}
	}

	tStart := time.Now()
	//running motors
	go func(err *error) {
		*err = e.RunPosConfig()
		log.Printf("run %s %v", e.GetName(), time.Now().Sub(tStart))
	}(&err)
	for _, opt := range opts {
		if err != nil {
			break
		}
		go func(op SyncOpts, err *error) {
			*err = op.Axis.RunPosConfig()
			log.Printf("run %s %v", op.Axis.GetName(), time.Now().Sub(tStart))
		}(opt, &err)
	}

	if err != nil {
		return
	}
	//check movement completed
	if err := e.CheckPosConfig(ctx); err != nil {
		return err
	}
	for _, opt := range opts {
		if err := opt.Axis.CheckPosConfig(ctx); err != nil {
			return err
		}
	}
	return nil
}

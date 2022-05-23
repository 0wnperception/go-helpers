package epr60ConnectorModbus

import (
	"context"
	"sync"
)

func (e *EPR60) RunAxisSpeedMoveAsync(wg *sync.WaitGroup, speed, acc uint16, dir bool) (err error) {
	wg.Add(1)
	go func(wg *sync.WaitGroup, e *EPR60, speed, acc uint16, dir bool, err *error) {
		*err = e.SpeedMove(speed, acc, dir)
		wg.Done()
	}(wg, e, speed, acc, dir, &err)
	return
}

func (e *EPR60) RunAxisPositionMoveAsync(ctx context.Context, wg *sync.WaitGroup, pos int, speed, acc uint16, dir bool) (err error) {
	wg.Add(1)
	go func(ctx context.Context, wg *sync.WaitGroup, e *EPR60, pos int, speed, acc uint16, dir bool, err *error) {
		*err = e.PositionMove(ctx, pos, speed, acc, dir)
		wg.Done()
	}(ctx, wg, e, pos, speed, acc, dir, &err)
	return
}

func (e *EPR60) DecStopAxisAsync(ctx context.Context, wg *sync.WaitGroup) (err error) {
	wg.Add(1)
	go func(wg *sync.WaitGroup, e *EPR60, err *error) {
		*err = e.DecStop(ctx)
		wg.Done()
	}(wg, e, &err)
	return
}

func (e *EPR60) EmergencyStopAxisAsync(ctx context.Context, wg *sync.WaitGroup) (err error) {
	wg.Add(1)
	go func(wg *sync.WaitGroup, e *EPR60, err *error) {
		*err = e.EmergencyStop(ctx)
		wg.Done()
	}(wg, e, &err)
	return
}

func (e *EPR60) DecStopAxisOnSensorAsync(
	ctx context.Context,
	wg *sync.WaitGroup,
	getSensorState func(ctx context.Context) (in bool, err error),
	state bool) (err error) {

	wg.Add(1)
	go func(ctx context.Context, wg *sync.WaitGroup, e *EPR60, getSensorState func(ctx context.Context) (in bool, err error), state bool, err *error) {
		*err = e.DecStopAxisOnSensor(ctx, getSensorState, state)
		wg.Done()
	}(ctx, wg, e, getSensorState, state, &err)
	return
}

func (e *EPR60) EmergencyStopAxisOnSensorAsync(
	ctx context.Context,
	wg *sync.WaitGroup,
	getSensorState func(ctx context.Context) (in bool, err error),
	state bool) (err error) {

	wg.Add(1)
	go func(wg *sync.WaitGroup, e *EPR60, getSensorState func(ctx context.Context) (in bool, err error), state bool, err *error) {
		*err = e.EmergencyStopAxisOnSensor(ctx, getSensorState, state)
		wg.Done()
	}(wg, e, getSensorState, state, &err)
	return
}

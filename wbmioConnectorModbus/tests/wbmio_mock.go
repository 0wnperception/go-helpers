package tests

// Code generated by http://github.com/gojuno/minimock (dev). DO NOT EDIT.

//go:generate minimock -i go-helpers/wbmioConnectorModbus.WBMIOIface -o ./tests/wbmio_mock.go -n Wbmio_mock

import (
	"sync"
	mm_atomic "sync/atomic"
	mm_time "time"

	"github.com/gojuno/minimock/v3"
)

// Wbmio_mock implements wbmioConnectorModbus.WBMIOIface
type Wbmio_mock struct {
	t minimock.Tester

	funcGetState          func(idx uint16) (in bool, err error)
	inspectFuncGetState   func(idx uint16)
	afterGetStateCounter  uint64
	beforeGetStateCounter uint64
	GetStateMock          mWbmio_mockGetState

	funcSetState          func(idx uint16, state bool) (err error)
	inspectFuncSetState   func(idx uint16, state bool)
	afterSetStateCounter  uint64
	beforeSetStateCounter uint64
	SetStateMock          mWbmio_mockSetState
}

// NewWbmio_mock returns a mock for wbmioConnectorModbus.WBMIOIface
func NewWbmio_mock(t minimock.Tester) *Wbmio_mock {
	m := &Wbmio_mock{t: t}
	if controller, ok := t.(minimock.MockController); ok {
		controller.RegisterMocker(m)
	}

	m.GetStateMock = mWbmio_mockGetState{mock: m}
	m.GetStateMock.callArgs = []*Wbmio_mockGetStateParams{}

	m.SetStateMock = mWbmio_mockSetState{mock: m}
	m.SetStateMock.callArgs = []*Wbmio_mockSetStateParams{}

	return m
}

type mWbmio_mockGetState struct {
	mock               *Wbmio_mock
	defaultExpectation *Wbmio_mockGetStateExpectation
	expectations       []*Wbmio_mockGetStateExpectation

	callArgs []*Wbmio_mockGetStateParams
	mutex    sync.RWMutex
}

// Wbmio_mockGetStateExpectation specifies expectation struct of the WBMIOIface.GetState
type Wbmio_mockGetStateExpectation struct {
	mock    *Wbmio_mock
	params  *Wbmio_mockGetStateParams
	results *Wbmio_mockGetStateResults
	Counter uint64
}

// Wbmio_mockGetStateParams contains parameters of the WBMIOIface.GetState
type Wbmio_mockGetStateParams struct {
	idx uint16
}

// Wbmio_mockGetStateResults contains results of the WBMIOIface.GetState
type Wbmio_mockGetStateResults struct {
	in  bool
	err error
}

// Expect sets up expected params for WBMIOIface.GetState
func (mmGetState *mWbmio_mockGetState) Expect(idx uint16) *mWbmio_mockGetState {
	if mmGetState.mock.funcGetState != nil {
		mmGetState.mock.t.Fatalf("Wbmio_mock.GetState mock is already set by Set")
	}

	if mmGetState.defaultExpectation == nil {
		mmGetState.defaultExpectation = &Wbmio_mockGetStateExpectation{}
	}

	mmGetState.defaultExpectation.params = &Wbmio_mockGetStateParams{idx}
	for _, e := range mmGetState.expectations {
		if minimock.Equal(e.params, mmGetState.defaultExpectation.params) {
			mmGetState.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmGetState.defaultExpectation.params)
		}
	}

	return mmGetState
}

// Inspect accepts an inspector function that has same arguments as the WBMIOIface.GetState
func (mmGetState *mWbmio_mockGetState) Inspect(f func(idx uint16)) *mWbmio_mockGetState {
	if mmGetState.mock.inspectFuncGetState != nil {
		mmGetState.mock.t.Fatalf("Inspect function is already set for Wbmio_mock.GetState")
	}

	mmGetState.mock.inspectFuncGetState = f

	return mmGetState
}

// Return sets up results that will be returned by WBMIOIface.GetState
func (mmGetState *mWbmio_mockGetState) Return(in bool, err error) *Wbmio_mock {
	if mmGetState.mock.funcGetState != nil {
		mmGetState.mock.t.Fatalf("Wbmio_mock.GetState mock is already set by Set")
	}

	if mmGetState.defaultExpectation == nil {
		mmGetState.defaultExpectation = &Wbmio_mockGetStateExpectation{mock: mmGetState.mock}
	}
	mmGetState.defaultExpectation.results = &Wbmio_mockGetStateResults{in, err}
	return mmGetState.mock
}

//Set uses given function f to mock the WBMIOIface.GetState method
func (mmGetState *mWbmio_mockGetState) Set(f func(idx uint16) (in bool, err error)) *Wbmio_mock {
	if mmGetState.defaultExpectation != nil {
		mmGetState.mock.t.Fatalf("Default expectation is already set for the WBMIOIface.GetState method")
	}

	if len(mmGetState.expectations) > 0 {
		mmGetState.mock.t.Fatalf("Some expectations are already set for the WBMIOIface.GetState method")
	}

	mmGetState.mock.funcGetState = f
	return mmGetState.mock
}

// When sets expectation for the WBMIOIface.GetState which will trigger the result defined by the following
// Then helper
func (mmGetState *mWbmio_mockGetState) When(idx uint16) *Wbmio_mockGetStateExpectation {
	if mmGetState.mock.funcGetState != nil {
		mmGetState.mock.t.Fatalf("Wbmio_mock.GetState mock is already set by Set")
	}

	expectation := &Wbmio_mockGetStateExpectation{
		mock:   mmGetState.mock,
		params: &Wbmio_mockGetStateParams{idx},
	}
	mmGetState.expectations = append(mmGetState.expectations, expectation)
	return expectation
}

// Then sets up WBMIOIface.GetState return parameters for the expectation previously defined by the When method
func (e *Wbmio_mockGetStateExpectation) Then(in bool, err error) *Wbmio_mock {
	e.results = &Wbmio_mockGetStateResults{in, err}
	return e.mock
}

// GetState implements wbmioConnectorModbus.WBMIOIface
func (mmGetState *Wbmio_mock) GetState(idx uint16) (in bool, err error) {
	mm_atomic.AddUint64(&mmGetState.beforeGetStateCounter, 1)
	defer mm_atomic.AddUint64(&mmGetState.afterGetStateCounter, 1)

	if mmGetState.inspectFuncGetState != nil {
		mmGetState.inspectFuncGetState(idx)
	}

	mm_params := &Wbmio_mockGetStateParams{idx}

	// Record call args
	mmGetState.GetStateMock.mutex.Lock()
	mmGetState.GetStateMock.callArgs = append(mmGetState.GetStateMock.callArgs, mm_params)
	mmGetState.GetStateMock.mutex.Unlock()

	for _, e := range mmGetState.GetStateMock.expectations {
		if minimock.Equal(e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return e.results.in, e.results.err
		}
	}

	if mmGetState.GetStateMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmGetState.GetStateMock.defaultExpectation.Counter, 1)
		mm_want := mmGetState.GetStateMock.defaultExpectation.params
		mm_got := Wbmio_mockGetStateParams{idx}
		if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmGetState.t.Errorf("Wbmio_mock.GetState got unexpected parameters, want: %#v, got: %#v%s\n", *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmGetState.GetStateMock.defaultExpectation.results
		if mm_results == nil {
			mmGetState.t.Fatal("No results are set for the Wbmio_mock.GetState")
		}
		return (*mm_results).in, (*mm_results).err
	}
	if mmGetState.funcGetState != nil {
		return mmGetState.funcGetState(idx)
	}
	mmGetState.t.Fatalf("Unexpected call to Wbmio_mock.GetState. %v", idx)
	return
}

// GetStateAfterCounter returns a count of finished Wbmio_mock.GetState invocations
func (mmGetState *Wbmio_mock) GetStateAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmGetState.afterGetStateCounter)
}

// GetStateBeforeCounter returns a count of Wbmio_mock.GetState invocations
func (mmGetState *Wbmio_mock) GetStateBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmGetState.beforeGetStateCounter)
}

// Calls returns a list of arguments used in each call to Wbmio_mock.GetState.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmGetState *mWbmio_mockGetState) Calls() []*Wbmio_mockGetStateParams {
	mmGetState.mutex.RLock()

	argCopy := make([]*Wbmio_mockGetStateParams, len(mmGetState.callArgs))
	copy(argCopy, mmGetState.callArgs)

	mmGetState.mutex.RUnlock()

	return argCopy
}

// MinimockGetStateDone returns true if the count of the GetState invocations corresponds
// the number of defined expectations
func (m *Wbmio_mock) MinimockGetStateDone() bool {
	for _, e := range m.GetStateMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.GetStateMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterGetStateCounter) < 1 {
		return false
	}
	// if func was set then invocations count should be greater than zero
	if m.funcGetState != nil && mm_atomic.LoadUint64(&m.afterGetStateCounter) < 1 {
		return false
	}
	return true
}

// MinimockGetStateInspect logs each unmet expectation
func (m *Wbmio_mock) MinimockGetStateInspect() {
	for _, e := range m.GetStateMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to Wbmio_mock.GetState with params: %#v", *e.params)
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.GetStateMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterGetStateCounter) < 1 {
		if m.GetStateMock.defaultExpectation.params == nil {
			m.t.Error("Expected call to Wbmio_mock.GetState")
		} else {
			m.t.Errorf("Expected call to Wbmio_mock.GetState with params: %#v", *m.GetStateMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcGetState != nil && mm_atomic.LoadUint64(&m.afterGetStateCounter) < 1 {
		m.t.Error("Expected call to Wbmio_mock.GetState")
	}
}

type mWbmio_mockSetState struct {
	mock               *Wbmio_mock
	defaultExpectation *Wbmio_mockSetStateExpectation
	expectations       []*Wbmio_mockSetStateExpectation

	callArgs []*Wbmio_mockSetStateParams
	mutex    sync.RWMutex
}

// Wbmio_mockSetStateExpectation specifies expectation struct of the WBMIOIface.SetState
type Wbmio_mockSetStateExpectation struct {
	mock    *Wbmio_mock
	params  *Wbmio_mockSetStateParams
	results *Wbmio_mockSetStateResults
	Counter uint64
}

// Wbmio_mockSetStateParams contains parameters of the WBMIOIface.SetState
type Wbmio_mockSetStateParams struct {
	idx   uint16
	state bool
}

// Wbmio_mockSetStateResults contains results of the WBMIOIface.SetState
type Wbmio_mockSetStateResults struct {
	err error
}

// Expect sets up expected params for WBMIOIface.SetState
func (mmSetState *mWbmio_mockSetState) Expect(idx uint16, state bool) *mWbmio_mockSetState {
	if mmSetState.mock.funcSetState != nil {
		mmSetState.mock.t.Fatalf("Wbmio_mock.SetState mock is already set by Set")
	}

	if mmSetState.defaultExpectation == nil {
		mmSetState.defaultExpectation = &Wbmio_mockSetStateExpectation{}
	}

	mmSetState.defaultExpectation.params = &Wbmio_mockSetStateParams{idx, state}
	for _, e := range mmSetState.expectations {
		if minimock.Equal(e.params, mmSetState.defaultExpectation.params) {
			mmSetState.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmSetState.defaultExpectation.params)
		}
	}

	return mmSetState
}

// Inspect accepts an inspector function that has same arguments as the WBMIOIface.SetState
func (mmSetState *mWbmio_mockSetState) Inspect(f func(idx uint16, state bool)) *mWbmio_mockSetState {
	if mmSetState.mock.inspectFuncSetState != nil {
		mmSetState.mock.t.Fatalf("Inspect function is already set for Wbmio_mock.SetState")
	}

	mmSetState.mock.inspectFuncSetState = f

	return mmSetState
}

// Return sets up results that will be returned by WBMIOIface.SetState
func (mmSetState *mWbmio_mockSetState) Return(err error) *Wbmio_mock {
	if mmSetState.mock.funcSetState != nil {
		mmSetState.mock.t.Fatalf("Wbmio_mock.SetState mock is already set by Set")
	}

	if mmSetState.defaultExpectation == nil {
		mmSetState.defaultExpectation = &Wbmio_mockSetStateExpectation{mock: mmSetState.mock}
	}
	mmSetState.defaultExpectation.results = &Wbmio_mockSetStateResults{err}
	return mmSetState.mock
}

//Set uses given function f to mock the WBMIOIface.SetState method
func (mmSetState *mWbmio_mockSetState) Set(f func(idx uint16, state bool) (err error)) *Wbmio_mock {
	if mmSetState.defaultExpectation != nil {
		mmSetState.mock.t.Fatalf("Default expectation is already set for the WBMIOIface.SetState method")
	}

	if len(mmSetState.expectations) > 0 {
		mmSetState.mock.t.Fatalf("Some expectations are already set for the WBMIOIface.SetState method")
	}

	mmSetState.mock.funcSetState = f
	return mmSetState.mock
}

// When sets expectation for the WBMIOIface.SetState which will trigger the result defined by the following
// Then helper
func (mmSetState *mWbmio_mockSetState) When(idx uint16, state bool) *Wbmio_mockSetStateExpectation {
	if mmSetState.mock.funcSetState != nil {
		mmSetState.mock.t.Fatalf("Wbmio_mock.SetState mock is already set by Set")
	}

	expectation := &Wbmio_mockSetStateExpectation{
		mock:   mmSetState.mock,
		params: &Wbmio_mockSetStateParams{idx, state},
	}
	mmSetState.expectations = append(mmSetState.expectations, expectation)
	return expectation
}

// Then sets up WBMIOIface.SetState return parameters for the expectation previously defined by the When method
func (e *Wbmio_mockSetStateExpectation) Then(err error) *Wbmio_mock {
	e.results = &Wbmio_mockSetStateResults{err}
	return e.mock
}

// SetState implements wbmioConnectorModbus.WBMIOIface
func (mmSetState *Wbmio_mock) SetState(idx uint16, state bool) (err error) {
	mm_atomic.AddUint64(&mmSetState.beforeSetStateCounter, 1)
	defer mm_atomic.AddUint64(&mmSetState.afterSetStateCounter, 1)

	if mmSetState.inspectFuncSetState != nil {
		mmSetState.inspectFuncSetState(idx, state)
	}

	mm_params := &Wbmio_mockSetStateParams{idx, state}

	// Record call args
	mmSetState.SetStateMock.mutex.Lock()
	mmSetState.SetStateMock.callArgs = append(mmSetState.SetStateMock.callArgs, mm_params)
	mmSetState.SetStateMock.mutex.Unlock()

	for _, e := range mmSetState.SetStateMock.expectations {
		if minimock.Equal(e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return e.results.err
		}
	}

	if mmSetState.SetStateMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmSetState.SetStateMock.defaultExpectation.Counter, 1)
		mm_want := mmSetState.SetStateMock.defaultExpectation.params
		mm_got := Wbmio_mockSetStateParams{idx, state}
		if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmSetState.t.Errorf("Wbmio_mock.SetState got unexpected parameters, want: %#v, got: %#v%s\n", *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmSetState.SetStateMock.defaultExpectation.results
		if mm_results == nil {
			mmSetState.t.Fatal("No results are set for the Wbmio_mock.SetState")
		}
		return (*mm_results).err
	}
	if mmSetState.funcSetState != nil {
		return mmSetState.funcSetState(idx, state)
	}
	mmSetState.t.Fatalf("Unexpected call to Wbmio_mock.SetState. %v %v", idx, state)
	return
}

// SetStateAfterCounter returns a count of finished Wbmio_mock.SetState invocations
func (mmSetState *Wbmio_mock) SetStateAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmSetState.afterSetStateCounter)
}

// SetStateBeforeCounter returns a count of Wbmio_mock.SetState invocations
func (mmSetState *Wbmio_mock) SetStateBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmSetState.beforeSetStateCounter)
}

// Calls returns a list of arguments used in each call to Wbmio_mock.SetState.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmSetState *mWbmio_mockSetState) Calls() []*Wbmio_mockSetStateParams {
	mmSetState.mutex.RLock()

	argCopy := make([]*Wbmio_mockSetStateParams, len(mmSetState.callArgs))
	copy(argCopy, mmSetState.callArgs)

	mmSetState.mutex.RUnlock()

	return argCopy
}

// MinimockSetStateDone returns true if the count of the SetState invocations corresponds
// the number of defined expectations
func (m *Wbmio_mock) MinimockSetStateDone() bool {
	for _, e := range m.SetStateMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.SetStateMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterSetStateCounter) < 1 {
		return false
	}
	// if func was set then invocations count should be greater than zero
	if m.funcSetState != nil && mm_atomic.LoadUint64(&m.afterSetStateCounter) < 1 {
		return false
	}
	return true
}

// MinimockSetStateInspect logs each unmet expectation
func (m *Wbmio_mock) MinimockSetStateInspect() {
	for _, e := range m.SetStateMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to Wbmio_mock.SetState with params: %#v", *e.params)
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.SetStateMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterSetStateCounter) < 1 {
		if m.SetStateMock.defaultExpectation.params == nil {
			m.t.Error("Expected call to Wbmio_mock.SetState")
		} else {
			m.t.Errorf("Expected call to Wbmio_mock.SetState with params: %#v", *m.SetStateMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcSetState != nil && mm_atomic.LoadUint64(&m.afterSetStateCounter) < 1 {
		m.t.Error("Expected call to Wbmio_mock.SetState")
	}
}

// MinimockFinish checks that all mocked methods have been called the expected number of times
func (m *Wbmio_mock) MinimockFinish() {
	if !m.minimockDone() {
		m.MinimockGetStateInspect()

		m.MinimockSetStateInspect()
		m.t.FailNow()
	}
}

// MinimockWait waits for all mocked methods to be called the expected number of times
func (m *Wbmio_mock) MinimockWait(timeout mm_time.Duration) {
	timeoutCh := mm_time.After(timeout)
	for {
		if m.minimockDone() {
			return
		}
		select {
		case <-timeoutCh:
			m.MinimockFinish()
			return
		case <-mm_time.After(10 * mm_time.Millisecond):
		}
	}
}

func (m *Wbmio_mock) minimockDone() bool {
	done := true
	return done &&
		m.MinimockGetStateDone() &&
		m.MinimockSetStateDone()
}
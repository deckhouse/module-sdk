// Code generated by http://github.com/gojuno/minimock (v3.4.3). DO NOT EDIT.

package hook_test

//go:generate minimock -i github.com/deckhouse/module-sdk/internal/hook.HookRequest -o hook_request_mock_test.go -n HookRequestMock -p hook_test

import (
	"sync"
	mm_atomic "sync/atomic"
	mm_time "time"

	bindingcontext "github.com/deckhouse/module-sdk/internal/binding-context"
	"github.com/deckhouse/module-sdk/pkg"
	"github.com/gojuno/minimock/v3"
)

// HookRequestMock implements HookRequest
type HookRequestMock struct {
	t          minimock.Tester
	finishOnce sync.Once

	funcGetBindingContexts          func() (ba1 []bindingcontext.BindingContext, err error)
	funcGetBindingContextsOrigin    string
	inspectFuncGetBindingContexts   func()
	afterGetBindingContextsCounter  uint64
	beforeGetBindingContextsCounter uint64
	GetBindingContextsMock          mHookRequestMockGetBindingContexts

	funcGetConfigValues          func() (m1 map[string]any, err error)
	funcGetConfigValuesOrigin    string
	inspectFuncGetConfigValues   func()
	afterGetConfigValuesCounter  uint64
	beforeGetConfigValuesCounter uint64
	GetConfigValuesMock          mHookRequestMockGetConfigValues

	funcGetDependencyContainer          func() (d1 pkg.DependencyContainer)
	funcGetDependencyContainerOrigin    string
	inspectFuncGetDependencyContainer   func()
	afterGetDependencyContainerCounter  uint64
	beforeGetDependencyContainerCounter uint64
	GetDependencyContainerMock          mHookRequestMockGetDependencyContainer

	funcGetValues          func() (m1 map[string]any, err error)
	funcGetValuesOrigin    string
	inspectFuncGetValues   func()
	afterGetValuesCounter  uint64
	beforeGetValuesCounter uint64
	GetValuesMock          mHookRequestMockGetValues
}

// NewHookRequestMock returns a mock for HookRequest
func NewHookRequestMock(t minimock.Tester) *HookRequestMock {
	m := &HookRequestMock{t: t}

	if controller, ok := t.(minimock.MockController); ok {
		controller.RegisterMocker(m)
	}

	m.GetBindingContextsMock = mHookRequestMockGetBindingContexts{mock: m}

	m.GetConfigValuesMock = mHookRequestMockGetConfigValues{mock: m}

	m.GetDependencyContainerMock = mHookRequestMockGetDependencyContainer{mock: m}

	m.GetValuesMock = mHookRequestMockGetValues{mock: m}

	t.Cleanup(m.MinimockFinish)

	return m
}

type mHookRequestMockGetBindingContexts struct {
	optional           bool
	mock               *HookRequestMock
	defaultExpectation *HookRequestMockGetBindingContextsExpectation
	expectations       []*HookRequestMockGetBindingContextsExpectation

	expectedInvocations       uint64
	expectedInvocationsOrigin string
}

// HookRequestMockGetBindingContextsExpectation specifies expectation struct of the HookRequest.GetBindingContexts
type HookRequestMockGetBindingContextsExpectation struct {
	mock *HookRequestMock

	results      *HookRequestMockGetBindingContextsResults
	returnOrigin string
	Counter      uint64
}

// HookRequestMockGetBindingContextsResults contains results of the HookRequest.GetBindingContexts
type HookRequestMockGetBindingContextsResults struct {
	ba1 []bindingcontext.BindingContext
	err error
}

// Marks this method to be optional. The default behavior of any method with Return() is '1 or more', meaning
// the test will fail minimock's automatic final call check if the mocked method was not called at least once.
// Optional() makes method check to work in '0 or more' mode.
// It is NOT RECOMMENDED to use this option unless you really need it, as default behaviour helps to
// catch the problems when the expected method call is totally skipped during test run.
func (mmGetBindingContexts *mHookRequestMockGetBindingContexts) Optional() *mHookRequestMockGetBindingContexts {
	mmGetBindingContexts.optional = true
	return mmGetBindingContexts
}

// Expect sets up expected params for HookRequest.GetBindingContexts
func (mmGetBindingContexts *mHookRequestMockGetBindingContexts) Expect() *mHookRequestMockGetBindingContexts {
	if mmGetBindingContexts.mock.funcGetBindingContexts != nil {
		mmGetBindingContexts.mock.t.Fatalf("HookRequestMock.GetBindingContexts mock is already set by Set")
	}

	if mmGetBindingContexts.defaultExpectation == nil {
		mmGetBindingContexts.defaultExpectation = &HookRequestMockGetBindingContextsExpectation{}
	}

	return mmGetBindingContexts
}

// Inspect accepts an inspector function that has same arguments as the HookRequest.GetBindingContexts
func (mmGetBindingContexts *mHookRequestMockGetBindingContexts) Inspect(f func()) *mHookRequestMockGetBindingContexts {
	if mmGetBindingContexts.mock.inspectFuncGetBindingContexts != nil {
		mmGetBindingContexts.mock.t.Fatalf("Inspect function is already set for HookRequestMock.GetBindingContexts")
	}

	mmGetBindingContexts.mock.inspectFuncGetBindingContexts = f

	return mmGetBindingContexts
}

// Return sets up results that will be returned by HookRequest.GetBindingContexts
func (mmGetBindingContexts *mHookRequestMockGetBindingContexts) Return(ba1 []bindingcontext.BindingContext, err error) *HookRequestMock {
	if mmGetBindingContexts.mock.funcGetBindingContexts != nil {
		mmGetBindingContexts.mock.t.Fatalf("HookRequestMock.GetBindingContexts mock is already set by Set")
	}

	if mmGetBindingContexts.defaultExpectation == nil {
		mmGetBindingContexts.defaultExpectation = &HookRequestMockGetBindingContextsExpectation{mock: mmGetBindingContexts.mock}
	}
	mmGetBindingContexts.defaultExpectation.results = &HookRequestMockGetBindingContextsResults{ba1, err}
	mmGetBindingContexts.defaultExpectation.returnOrigin = minimock.CallerInfo(1)
	return mmGetBindingContexts.mock
}

// Set uses given function f to mock the HookRequest.GetBindingContexts method
func (mmGetBindingContexts *mHookRequestMockGetBindingContexts) Set(f func() (ba1 []bindingcontext.BindingContext, err error)) *HookRequestMock {
	if mmGetBindingContexts.defaultExpectation != nil {
		mmGetBindingContexts.mock.t.Fatalf("Default expectation is already set for the HookRequest.GetBindingContexts method")
	}

	if len(mmGetBindingContexts.expectations) > 0 {
		mmGetBindingContexts.mock.t.Fatalf("Some expectations are already set for the HookRequest.GetBindingContexts method")
	}

	mmGetBindingContexts.mock.funcGetBindingContexts = f
	mmGetBindingContexts.mock.funcGetBindingContextsOrigin = minimock.CallerInfo(1)
	return mmGetBindingContexts.mock
}

// Times sets number of times HookRequest.GetBindingContexts should be invoked
func (mmGetBindingContexts *mHookRequestMockGetBindingContexts) Times(n uint64) *mHookRequestMockGetBindingContexts {
	if n == 0 {
		mmGetBindingContexts.mock.t.Fatalf("Times of HookRequestMock.GetBindingContexts mock can not be zero")
	}
	mm_atomic.StoreUint64(&mmGetBindingContexts.expectedInvocations, n)
	mmGetBindingContexts.expectedInvocationsOrigin = minimock.CallerInfo(1)
	return mmGetBindingContexts
}

func (mmGetBindingContexts *mHookRequestMockGetBindingContexts) invocationsDone() bool {
	if len(mmGetBindingContexts.expectations) == 0 && mmGetBindingContexts.defaultExpectation == nil && mmGetBindingContexts.mock.funcGetBindingContexts == nil {
		return true
	}

	totalInvocations := mm_atomic.LoadUint64(&mmGetBindingContexts.mock.afterGetBindingContextsCounter)
	expectedInvocations := mm_atomic.LoadUint64(&mmGetBindingContexts.expectedInvocations)

	return totalInvocations > 0 && (expectedInvocations == 0 || expectedInvocations == totalInvocations)
}

// GetBindingContexts implements HookRequest
func (mmGetBindingContexts *HookRequestMock) GetBindingContexts() (ba1 []bindingcontext.BindingContext, err error) {
	mm_atomic.AddUint64(&mmGetBindingContexts.beforeGetBindingContextsCounter, 1)
	defer mm_atomic.AddUint64(&mmGetBindingContexts.afterGetBindingContextsCounter, 1)

	mmGetBindingContexts.t.Helper()

	if mmGetBindingContexts.inspectFuncGetBindingContexts != nil {
		mmGetBindingContexts.inspectFuncGetBindingContexts()
	}

	if mmGetBindingContexts.GetBindingContextsMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmGetBindingContexts.GetBindingContextsMock.defaultExpectation.Counter, 1)

		mm_results := mmGetBindingContexts.GetBindingContextsMock.defaultExpectation.results
		if mm_results == nil {
			mmGetBindingContexts.t.Fatal("No results are set for the HookRequestMock.GetBindingContexts")
		}
		return (*mm_results).ba1, (*mm_results).err
	}
	if mmGetBindingContexts.funcGetBindingContexts != nil {
		return mmGetBindingContexts.funcGetBindingContexts()
	}
	mmGetBindingContexts.t.Fatalf("Unexpected call to HookRequestMock.GetBindingContexts.")
	return
}

// GetBindingContextsAfterCounter returns a count of finished HookRequestMock.GetBindingContexts invocations
func (mmGetBindingContexts *HookRequestMock) GetBindingContextsAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmGetBindingContexts.afterGetBindingContextsCounter)
}

// GetBindingContextsBeforeCounter returns a count of HookRequestMock.GetBindingContexts invocations
func (mmGetBindingContexts *HookRequestMock) GetBindingContextsBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmGetBindingContexts.beforeGetBindingContextsCounter)
}

// MinimockGetBindingContextsDone returns true if the count of the GetBindingContexts invocations corresponds
// the number of defined expectations
func (m *HookRequestMock) MinimockGetBindingContextsDone() bool {
	if m.GetBindingContextsMock.optional {
		// Optional methods provide '0 or more' call count restriction.
		return true
	}

	for _, e := range m.GetBindingContextsMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	return m.GetBindingContextsMock.invocationsDone()
}

// MinimockGetBindingContextsInspect logs each unmet expectation
func (m *HookRequestMock) MinimockGetBindingContextsInspect() {
	for _, e := range m.GetBindingContextsMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Error("Expected call to HookRequestMock.GetBindingContexts")
		}
	}

	afterGetBindingContextsCounter := mm_atomic.LoadUint64(&m.afterGetBindingContextsCounter)
	// if default expectation was set then invocations count should be greater than zero
	if m.GetBindingContextsMock.defaultExpectation != nil && afterGetBindingContextsCounter < 1 {
		m.t.Errorf("Expected call to HookRequestMock.GetBindingContexts at\n%s", m.GetBindingContextsMock.defaultExpectation.returnOrigin)
	}
	// if func was set then invocations count should be greater than zero
	if m.funcGetBindingContexts != nil && afterGetBindingContextsCounter < 1 {
		m.t.Errorf("Expected call to HookRequestMock.GetBindingContexts at\n%s", m.funcGetBindingContextsOrigin)
	}

	if !m.GetBindingContextsMock.invocationsDone() && afterGetBindingContextsCounter > 0 {
		m.t.Errorf("Expected %d calls to HookRequestMock.GetBindingContexts at\n%s but found %d calls",
			mm_atomic.LoadUint64(&m.GetBindingContextsMock.expectedInvocations), m.GetBindingContextsMock.expectedInvocationsOrigin, afterGetBindingContextsCounter)
	}
}

type mHookRequestMockGetConfigValues struct {
	optional           bool
	mock               *HookRequestMock
	defaultExpectation *HookRequestMockGetConfigValuesExpectation
	expectations       []*HookRequestMockGetConfigValuesExpectation

	expectedInvocations       uint64
	expectedInvocationsOrigin string
}

// HookRequestMockGetConfigValuesExpectation specifies expectation struct of the HookRequest.GetConfigValues
type HookRequestMockGetConfigValuesExpectation struct {
	mock *HookRequestMock

	results      *HookRequestMockGetConfigValuesResults
	returnOrigin string
	Counter      uint64
}

// HookRequestMockGetConfigValuesResults contains results of the HookRequest.GetConfigValues
type HookRequestMockGetConfigValuesResults struct {
	m1  map[string]any
	err error
}

// Marks this method to be optional. The default behavior of any method with Return() is '1 or more', meaning
// the test will fail minimock's automatic final call check if the mocked method was not called at least once.
// Optional() makes method check to work in '0 or more' mode.
// It is NOT RECOMMENDED to use this option unless you really need it, as default behaviour helps to
// catch the problems when the expected method call is totally skipped during test run.
func (mmGetConfigValues *mHookRequestMockGetConfigValues) Optional() *mHookRequestMockGetConfigValues {
	mmGetConfigValues.optional = true
	return mmGetConfigValues
}

// Expect sets up expected params for HookRequest.GetConfigValues
func (mmGetConfigValues *mHookRequestMockGetConfigValues) Expect() *mHookRequestMockGetConfigValues {
	if mmGetConfigValues.mock.funcGetConfigValues != nil {
		mmGetConfigValues.mock.t.Fatalf("HookRequestMock.GetConfigValues mock is already set by Set")
	}

	if mmGetConfigValues.defaultExpectation == nil {
		mmGetConfigValues.defaultExpectation = &HookRequestMockGetConfigValuesExpectation{}
	}

	return mmGetConfigValues
}

// Inspect accepts an inspector function that has same arguments as the HookRequest.GetConfigValues
func (mmGetConfigValues *mHookRequestMockGetConfigValues) Inspect(f func()) *mHookRequestMockGetConfigValues {
	if mmGetConfigValues.mock.inspectFuncGetConfigValues != nil {
		mmGetConfigValues.mock.t.Fatalf("Inspect function is already set for HookRequestMock.GetConfigValues")
	}

	mmGetConfigValues.mock.inspectFuncGetConfigValues = f

	return mmGetConfigValues
}

// Return sets up results that will be returned by HookRequest.GetConfigValues
func (mmGetConfigValues *mHookRequestMockGetConfigValues) Return(m1 map[string]any, err error) *HookRequestMock {
	if mmGetConfigValues.mock.funcGetConfigValues != nil {
		mmGetConfigValues.mock.t.Fatalf("HookRequestMock.GetConfigValues mock is already set by Set")
	}

	if mmGetConfigValues.defaultExpectation == nil {
		mmGetConfigValues.defaultExpectation = &HookRequestMockGetConfigValuesExpectation{mock: mmGetConfigValues.mock}
	}
	mmGetConfigValues.defaultExpectation.results = &HookRequestMockGetConfigValuesResults{m1, err}
	mmGetConfigValues.defaultExpectation.returnOrigin = minimock.CallerInfo(1)
	return mmGetConfigValues.mock
}

// Set uses given function f to mock the HookRequest.GetConfigValues method
func (mmGetConfigValues *mHookRequestMockGetConfigValues) Set(f func() (m1 map[string]any, err error)) *HookRequestMock {
	if mmGetConfigValues.defaultExpectation != nil {
		mmGetConfigValues.mock.t.Fatalf("Default expectation is already set for the HookRequest.GetConfigValues method")
	}

	if len(mmGetConfigValues.expectations) > 0 {
		mmGetConfigValues.mock.t.Fatalf("Some expectations are already set for the HookRequest.GetConfigValues method")
	}

	mmGetConfigValues.mock.funcGetConfigValues = f
	mmGetConfigValues.mock.funcGetConfigValuesOrigin = minimock.CallerInfo(1)
	return mmGetConfigValues.mock
}

// Times sets number of times HookRequest.GetConfigValues should be invoked
func (mmGetConfigValues *mHookRequestMockGetConfigValues) Times(n uint64) *mHookRequestMockGetConfigValues {
	if n == 0 {
		mmGetConfigValues.mock.t.Fatalf("Times of HookRequestMock.GetConfigValues mock can not be zero")
	}
	mm_atomic.StoreUint64(&mmGetConfigValues.expectedInvocations, n)
	mmGetConfigValues.expectedInvocationsOrigin = minimock.CallerInfo(1)
	return mmGetConfigValues
}

func (mmGetConfigValues *mHookRequestMockGetConfigValues) invocationsDone() bool {
	if len(mmGetConfigValues.expectations) == 0 && mmGetConfigValues.defaultExpectation == nil && mmGetConfigValues.mock.funcGetConfigValues == nil {
		return true
	}

	totalInvocations := mm_atomic.LoadUint64(&mmGetConfigValues.mock.afterGetConfigValuesCounter)
	expectedInvocations := mm_atomic.LoadUint64(&mmGetConfigValues.expectedInvocations)

	return totalInvocations > 0 && (expectedInvocations == 0 || expectedInvocations == totalInvocations)
}

// GetConfigValues implements HookRequest
func (mmGetConfigValues *HookRequestMock) GetConfigValues() (m1 map[string]any, err error) {
	mm_atomic.AddUint64(&mmGetConfigValues.beforeGetConfigValuesCounter, 1)
	defer mm_atomic.AddUint64(&mmGetConfigValues.afterGetConfigValuesCounter, 1)

	mmGetConfigValues.t.Helper()

	if mmGetConfigValues.inspectFuncGetConfigValues != nil {
		mmGetConfigValues.inspectFuncGetConfigValues()
	}

	if mmGetConfigValues.GetConfigValuesMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmGetConfigValues.GetConfigValuesMock.defaultExpectation.Counter, 1)

		mm_results := mmGetConfigValues.GetConfigValuesMock.defaultExpectation.results
		if mm_results == nil {
			mmGetConfigValues.t.Fatal("No results are set for the HookRequestMock.GetConfigValues")
		}
		return (*mm_results).m1, (*mm_results).err
	}
	if mmGetConfigValues.funcGetConfigValues != nil {
		return mmGetConfigValues.funcGetConfigValues()
	}
	mmGetConfigValues.t.Fatalf("Unexpected call to HookRequestMock.GetConfigValues.")
	return
}

// GetConfigValuesAfterCounter returns a count of finished HookRequestMock.GetConfigValues invocations
func (mmGetConfigValues *HookRequestMock) GetConfigValuesAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmGetConfigValues.afterGetConfigValuesCounter)
}

// GetConfigValuesBeforeCounter returns a count of HookRequestMock.GetConfigValues invocations
func (mmGetConfigValues *HookRequestMock) GetConfigValuesBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmGetConfigValues.beforeGetConfigValuesCounter)
}

// MinimockGetConfigValuesDone returns true if the count of the GetConfigValues invocations corresponds
// the number of defined expectations
func (m *HookRequestMock) MinimockGetConfigValuesDone() bool {
	if m.GetConfigValuesMock.optional {
		// Optional methods provide '0 or more' call count restriction.
		return true
	}

	for _, e := range m.GetConfigValuesMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	return m.GetConfigValuesMock.invocationsDone()
}

// MinimockGetConfigValuesInspect logs each unmet expectation
func (m *HookRequestMock) MinimockGetConfigValuesInspect() {
	for _, e := range m.GetConfigValuesMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Error("Expected call to HookRequestMock.GetConfigValues")
		}
	}

	afterGetConfigValuesCounter := mm_atomic.LoadUint64(&m.afterGetConfigValuesCounter)
	// if default expectation was set then invocations count should be greater than zero
	if m.GetConfigValuesMock.defaultExpectation != nil && afterGetConfigValuesCounter < 1 {
		m.t.Errorf("Expected call to HookRequestMock.GetConfigValues at\n%s", m.GetConfigValuesMock.defaultExpectation.returnOrigin)
	}
	// if func was set then invocations count should be greater than zero
	if m.funcGetConfigValues != nil && afterGetConfigValuesCounter < 1 {
		m.t.Errorf("Expected call to HookRequestMock.GetConfigValues at\n%s", m.funcGetConfigValuesOrigin)
	}

	if !m.GetConfigValuesMock.invocationsDone() && afterGetConfigValuesCounter > 0 {
		m.t.Errorf("Expected %d calls to HookRequestMock.GetConfigValues at\n%s but found %d calls",
			mm_atomic.LoadUint64(&m.GetConfigValuesMock.expectedInvocations), m.GetConfigValuesMock.expectedInvocationsOrigin, afterGetConfigValuesCounter)
	}
}

type mHookRequestMockGetDependencyContainer struct {
	optional           bool
	mock               *HookRequestMock
	defaultExpectation *HookRequestMockGetDependencyContainerExpectation
	expectations       []*HookRequestMockGetDependencyContainerExpectation

	expectedInvocations       uint64
	expectedInvocationsOrigin string
}

// HookRequestMockGetDependencyContainerExpectation specifies expectation struct of the HookRequest.GetDependencyContainer
type HookRequestMockGetDependencyContainerExpectation struct {
	mock *HookRequestMock

	results      *HookRequestMockGetDependencyContainerResults
	returnOrigin string
	Counter      uint64
}

// HookRequestMockGetDependencyContainerResults contains results of the HookRequest.GetDependencyContainer
type HookRequestMockGetDependencyContainerResults struct {
	d1 pkg.DependencyContainer
}

// Marks this method to be optional. The default behavior of any method with Return() is '1 or more', meaning
// the test will fail minimock's automatic final call check if the mocked method was not called at least once.
// Optional() makes method check to work in '0 or more' mode.
// It is NOT RECOMMENDED to use this option unless you really need it, as default behaviour helps to
// catch the problems when the expected method call is totally skipped during test run.
func (mmGetDependencyContainer *mHookRequestMockGetDependencyContainer) Optional() *mHookRequestMockGetDependencyContainer {
	mmGetDependencyContainer.optional = true
	return mmGetDependencyContainer
}

// Expect sets up expected params for HookRequest.GetDependencyContainer
func (mmGetDependencyContainer *mHookRequestMockGetDependencyContainer) Expect() *mHookRequestMockGetDependencyContainer {
	if mmGetDependencyContainer.mock.funcGetDependencyContainer != nil {
		mmGetDependencyContainer.mock.t.Fatalf("HookRequestMock.GetDependencyContainer mock is already set by Set")
	}

	if mmGetDependencyContainer.defaultExpectation == nil {
		mmGetDependencyContainer.defaultExpectation = &HookRequestMockGetDependencyContainerExpectation{}
	}

	return mmGetDependencyContainer
}

// Inspect accepts an inspector function that has same arguments as the HookRequest.GetDependencyContainer
func (mmGetDependencyContainer *mHookRequestMockGetDependencyContainer) Inspect(f func()) *mHookRequestMockGetDependencyContainer {
	if mmGetDependencyContainer.mock.inspectFuncGetDependencyContainer != nil {
		mmGetDependencyContainer.mock.t.Fatalf("Inspect function is already set for HookRequestMock.GetDependencyContainer")
	}

	mmGetDependencyContainer.mock.inspectFuncGetDependencyContainer = f

	return mmGetDependencyContainer
}

// Return sets up results that will be returned by HookRequest.GetDependencyContainer
func (mmGetDependencyContainer *mHookRequestMockGetDependencyContainer) Return(d1 pkg.DependencyContainer) *HookRequestMock {
	if mmGetDependencyContainer.mock.funcGetDependencyContainer != nil {
		mmGetDependencyContainer.mock.t.Fatalf("HookRequestMock.GetDependencyContainer mock is already set by Set")
	}

	if mmGetDependencyContainer.defaultExpectation == nil {
		mmGetDependencyContainer.defaultExpectation = &HookRequestMockGetDependencyContainerExpectation{mock: mmGetDependencyContainer.mock}
	}
	mmGetDependencyContainer.defaultExpectation.results = &HookRequestMockGetDependencyContainerResults{d1}
	mmGetDependencyContainer.defaultExpectation.returnOrigin = minimock.CallerInfo(1)
	return mmGetDependencyContainer.mock
}

// Set uses given function f to mock the HookRequest.GetDependencyContainer method
func (mmGetDependencyContainer *mHookRequestMockGetDependencyContainer) Set(f func() (d1 pkg.DependencyContainer)) *HookRequestMock {
	if mmGetDependencyContainer.defaultExpectation != nil {
		mmGetDependencyContainer.mock.t.Fatalf("Default expectation is already set for the HookRequest.GetDependencyContainer method")
	}

	if len(mmGetDependencyContainer.expectations) > 0 {
		mmGetDependencyContainer.mock.t.Fatalf("Some expectations are already set for the HookRequest.GetDependencyContainer method")
	}

	mmGetDependencyContainer.mock.funcGetDependencyContainer = f
	mmGetDependencyContainer.mock.funcGetDependencyContainerOrigin = minimock.CallerInfo(1)
	return mmGetDependencyContainer.mock
}

// Times sets number of times HookRequest.GetDependencyContainer should be invoked
func (mmGetDependencyContainer *mHookRequestMockGetDependencyContainer) Times(n uint64) *mHookRequestMockGetDependencyContainer {
	if n == 0 {
		mmGetDependencyContainer.mock.t.Fatalf("Times of HookRequestMock.GetDependencyContainer mock can not be zero")
	}
	mm_atomic.StoreUint64(&mmGetDependencyContainer.expectedInvocations, n)
	mmGetDependencyContainer.expectedInvocationsOrigin = minimock.CallerInfo(1)
	return mmGetDependencyContainer
}

func (mmGetDependencyContainer *mHookRequestMockGetDependencyContainer) invocationsDone() bool {
	if len(mmGetDependencyContainer.expectations) == 0 && mmGetDependencyContainer.defaultExpectation == nil && mmGetDependencyContainer.mock.funcGetDependencyContainer == nil {
		return true
	}

	totalInvocations := mm_atomic.LoadUint64(&mmGetDependencyContainer.mock.afterGetDependencyContainerCounter)
	expectedInvocations := mm_atomic.LoadUint64(&mmGetDependencyContainer.expectedInvocations)

	return totalInvocations > 0 && (expectedInvocations == 0 || expectedInvocations == totalInvocations)
}

// GetDependencyContainer implements HookRequest
func (mmGetDependencyContainer *HookRequestMock) GetDependencyContainer() (d1 pkg.DependencyContainer) {
	mm_atomic.AddUint64(&mmGetDependencyContainer.beforeGetDependencyContainerCounter, 1)
	defer mm_atomic.AddUint64(&mmGetDependencyContainer.afterGetDependencyContainerCounter, 1)

	mmGetDependencyContainer.t.Helper()

	if mmGetDependencyContainer.inspectFuncGetDependencyContainer != nil {
		mmGetDependencyContainer.inspectFuncGetDependencyContainer()
	}

	if mmGetDependencyContainer.GetDependencyContainerMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmGetDependencyContainer.GetDependencyContainerMock.defaultExpectation.Counter, 1)

		mm_results := mmGetDependencyContainer.GetDependencyContainerMock.defaultExpectation.results
		if mm_results == nil {
			mmGetDependencyContainer.t.Fatal("No results are set for the HookRequestMock.GetDependencyContainer")
		}
		return (*mm_results).d1
	}
	if mmGetDependencyContainer.funcGetDependencyContainer != nil {
		return mmGetDependencyContainer.funcGetDependencyContainer()
	}
	mmGetDependencyContainer.t.Fatalf("Unexpected call to HookRequestMock.GetDependencyContainer.")
	return
}

// GetDependencyContainerAfterCounter returns a count of finished HookRequestMock.GetDependencyContainer invocations
func (mmGetDependencyContainer *HookRequestMock) GetDependencyContainerAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmGetDependencyContainer.afterGetDependencyContainerCounter)
}

// GetDependencyContainerBeforeCounter returns a count of HookRequestMock.GetDependencyContainer invocations
func (mmGetDependencyContainer *HookRequestMock) GetDependencyContainerBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmGetDependencyContainer.beforeGetDependencyContainerCounter)
}

// MinimockGetDependencyContainerDone returns true if the count of the GetDependencyContainer invocations corresponds
// the number of defined expectations
func (m *HookRequestMock) MinimockGetDependencyContainerDone() bool {
	if m.GetDependencyContainerMock.optional {
		// Optional methods provide '0 or more' call count restriction.
		return true
	}

	for _, e := range m.GetDependencyContainerMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	return m.GetDependencyContainerMock.invocationsDone()
}

// MinimockGetDependencyContainerInspect logs each unmet expectation
func (m *HookRequestMock) MinimockGetDependencyContainerInspect() {
	for _, e := range m.GetDependencyContainerMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Error("Expected call to HookRequestMock.GetDependencyContainer")
		}
	}

	afterGetDependencyContainerCounter := mm_atomic.LoadUint64(&m.afterGetDependencyContainerCounter)
	// if default expectation was set then invocations count should be greater than zero
	if m.GetDependencyContainerMock.defaultExpectation != nil && afterGetDependencyContainerCounter < 1 {
		m.t.Errorf("Expected call to HookRequestMock.GetDependencyContainer at\n%s", m.GetDependencyContainerMock.defaultExpectation.returnOrigin)
	}
	// if func was set then invocations count should be greater than zero
	if m.funcGetDependencyContainer != nil && afterGetDependencyContainerCounter < 1 {
		m.t.Errorf("Expected call to HookRequestMock.GetDependencyContainer at\n%s", m.funcGetDependencyContainerOrigin)
	}

	if !m.GetDependencyContainerMock.invocationsDone() && afterGetDependencyContainerCounter > 0 {
		m.t.Errorf("Expected %d calls to HookRequestMock.GetDependencyContainer at\n%s but found %d calls",
			mm_atomic.LoadUint64(&m.GetDependencyContainerMock.expectedInvocations), m.GetDependencyContainerMock.expectedInvocationsOrigin, afterGetDependencyContainerCounter)
	}
}

type mHookRequestMockGetValues struct {
	optional           bool
	mock               *HookRequestMock
	defaultExpectation *HookRequestMockGetValuesExpectation
	expectations       []*HookRequestMockGetValuesExpectation

	expectedInvocations       uint64
	expectedInvocationsOrigin string
}

// HookRequestMockGetValuesExpectation specifies expectation struct of the HookRequest.GetValues
type HookRequestMockGetValuesExpectation struct {
	mock *HookRequestMock

	results      *HookRequestMockGetValuesResults
	returnOrigin string
	Counter      uint64
}

// HookRequestMockGetValuesResults contains results of the HookRequest.GetValues
type HookRequestMockGetValuesResults struct {
	m1  map[string]any
	err error
}

// Marks this method to be optional. The default behavior of any method with Return() is '1 or more', meaning
// the test will fail minimock's automatic final call check if the mocked method was not called at least once.
// Optional() makes method check to work in '0 or more' mode.
// It is NOT RECOMMENDED to use this option unless you really need it, as default behaviour helps to
// catch the problems when the expected method call is totally skipped during test run.
func (mmGetValues *mHookRequestMockGetValues) Optional() *mHookRequestMockGetValues {
	mmGetValues.optional = true
	return mmGetValues
}

// Expect sets up expected params for HookRequest.GetValues
func (mmGetValues *mHookRequestMockGetValues) Expect() *mHookRequestMockGetValues {
	if mmGetValues.mock.funcGetValues != nil {
		mmGetValues.mock.t.Fatalf("HookRequestMock.GetValues mock is already set by Set")
	}

	if mmGetValues.defaultExpectation == nil {
		mmGetValues.defaultExpectation = &HookRequestMockGetValuesExpectation{}
	}

	return mmGetValues
}

// Inspect accepts an inspector function that has same arguments as the HookRequest.GetValues
func (mmGetValues *mHookRequestMockGetValues) Inspect(f func()) *mHookRequestMockGetValues {
	if mmGetValues.mock.inspectFuncGetValues != nil {
		mmGetValues.mock.t.Fatalf("Inspect function is already set for HookRequestMock.GetValues")
	}

	mmGetValues.mock.inspectFuncGetValues = f

	return mmGetValues
}

// Return sets up results that will be returned by HookRequest.GetValues
func (mmGetValues *mHookRequestMockGetValues) Return(m1 map[string]any, err error) *HookRequestMock {
	if mmGetValues.mock.funcGetValues != nil {
		mmGetValues.mock.t.Fatalf("HookRequestMock.GetValues mock is already set by Set")
	}

	if mmGetValues.defaultExpectation == nil {
		mmGetValues.defaultExpectation = &HookRequestMockGetValuesExpectation{mock: mmGetValues.mock}
	}
	mmGetValues.defaultExpectation.results = &HookRequestMockGetValuesResults{m1, err}
	mmGetValues.defaultExpectation.returnOrigin = minimock.CallerInfo(1)
	return mmGetValues.mock
}

// Set uses given function f to mock the HookRequest.GetValues method
func (mmGetValues *mHookRequestMockGetValues) Set(f func() (m1 map[string]any, err error)) *HookRequestMock {
	if mmGetValues.defaultExpectation != nil {
		mmGetValues.mock.t.Fatalf("Default expectation is already set for the HookRequest.GetValues method")
	}

	if len(mmGetValues.expectations) > 0 {
		mmGetValues.mock.t.Fatalf("Some expectations are already set for the HookRequest.GetValues method")
	}

	mmGetValues.mock.funcGetValues = f
	mmGetValues.mock.funcGetValuesOrigin = minimock.CallerInfo(1)
	return mmGetValues.mock
}

// Times sets number of times HookRequest.GetValues should be invoked
func (mmGetValues *mHookRequestMockGetValues) Times(n uint64) *mHookRequestMockGetValues {
	if n == 0 {
		mmGetValues.mock.t.Fatalf("Times of HookRequestMock.GetValues mock can not be zero")
	}
	mm_atomic.StoreUint64(&mmGetValues.expectedInvocations, n)
	mmGetValues.expectedInvocationsOrigin = minimock.CallerInfo(1)
	return mmGetValues
}

func (mmGetValues *mHookRequestMockGetValues) invocationsDone() bool {
	if len(mmGetValues.expectations) == 0 && mmGetValues.defaultExpectation == nil && mmGetValues.mock.funcGetValues == nil {
		return true
	}

	totalInvocations := mm_atomic.LoadUint64(&mmGetValues.mock.afterGetValuesCounter)
	expectedInvocations := mm_atomic.LoadUint64(&mmGetValues.expectedInvocations)

	return totalInvocations > 0 && (expectedInvocations == 0 || expectedInvocations == totalInvocations)
}

// GetValues implements HookRequest
func (mmGetValues *HookRequestMock) GetValues() (m1 map[string]any, err error) {
	mm_atomic.AddUint64(&mmGetValues.beforeGetValuesCounter, 1)
	defer mm_atomic.AddUint64(&mmGetValues.afterGetValuesCounter, 1)

	mmGetValues.t.Helper()

	if mmGetValues.inspectFuncGetValues != nil {
		mmGetValues.inspectFuncGetValues()
	}

	if mmGetValues.GetValuesMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmGetValues.GetValuesMock.defaultExpectation.Counter, 1)

		mm_results := mmGetValues.GetValuesMock.defaultExpectation.results
		if mm_results == nil {
			mmGetValues.t.Fatal("No results are set for the HookRequestMock.GetValues")
		}
		return (*mm_results).m1, (*mm_results).err
	}
	if mmGetValues.funcGetValues != nil {
		return mmGetValues.funcGetValues()
	}
	mmGetValues.t.Fatalf("Unexpected call to HookRequestMock.GetValues.")
	return
}

// GetValuesAfterCounter returns a count of finished HookRequestMock.GetValues invocations
func (mmGetValues *HookRequestMock) GetValuesAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmGetValues.afterGetValuesCounter)
}

// GetValuesBeforeCounter returns a count of HookRequestMock.GetValues invocations
func (mmGetValues *HookRequestMock) GetValuesBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmGetValues.beforeGetValuesCounter)
}

// MinimockGetValuesDone returns true if the count of the GetValues invocations corresponds
// the number of defined expectations
func (m *HookRequestMock) MinimockGetValuesDone() bool {
	if m.GetValuesMock.optional {
		// Optional methods provide '0 or more' call count restriction.
		return true
	}

	for _, e := range m.GetValuesMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	return m.GetValuesMock.invocationsDone()
}

// MinimockGetValuesInspect logs each unmet expectation
func (m *HookRequestMock) MinimockGetValuesInspect() {
	for _, e := range m.GetValuesMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Error("Expected call to HookRequestMock.GetValues")
		}
	}

	afterGetValuesCounter := mm_atomic.LoadUint64(&m.afterGetValuesCounter)
	// if default expectation was set then invocations count should be greater than zero
	if m.GetValuesMock.defaultExpectation != nil && afterGetValuesCounter < 1 {
		m.t.Errorf("Expected call to HookRequestMock.GetValues at\n%s", m.GetValuesMock.defaultExpectation.returnOrigin)
	}
	// if func was set then invocations count should be greater than zero
	if m.funcGetValues != nil && afterGetValuesCounter < 1 {
		m.t.Errorf("Expected call to HookRequestMock.GetValues at\n%s", m.funcGetValuesOrigin)
	}

	if !m.GetValuesMock.invocationsDone() && afterGetValuesCounter > 0 {
		m.t.Errorf("Expected %d calls to HookRequestMock.GetValues at\n%s but found %d calls",
			mm_atomic.LoadUint64(&m.GetValuesMock.expectedInvocations), m.GetValuesMock.expectedInvocationsOrigin, afterGetValuesCounter)
	}
}

// MinimockFinish checks that all mocked methods have been called the expected number of times
func (m *HookRequestMock) MinimockFinish() {
	m.finishOnce.Do(func() {
		if !m.minimockDone() {
			m.MinimockGetBindingContextsInspect()

			m.MinimockGetConfigValuesInspect()

			m.MinimockGetDependencyContainerInspect()

			m.MinimockGetValuesInspect()
		}
	})
}

// MinimockWait waits for all mocked methods to be called the expected number of times
func (m *HookRequestMock) MinimockWait(timeout mm_time.Duration) {
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

func (m *HookRequestMock) minimockDone() bool {
	done := true
	return done &&
		m.MinimockGetBindingContextsDone() &&
		m.MinimockGetConfigValuesDone() &&
		m.MinimockGetDependencyContainerDone() &&
		m.MinimockGetValuesDone()
}

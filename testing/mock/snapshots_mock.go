// Code generated by http://github.com/gojuno/minimock (v3.4.4). DO NOT EDIT.

package mock

//go:generate minimock -i github.com/deckhouse/module-sdk/pkg.Snapshots -o snapshots_mock.go -n SnapshotsMock -p mock

import (
	"sync"
	mm_atomic "sync/atomic"
	mm_time "time"

	mm_pkg "github.com/deckhouse/module-sdk/pkg"
	"github.com/gojuno/minimock/v3"
)

// SnapshotsMock implements mm_pkg.Snapshots
type SnapshotsMock struct {
	t          minimock.Tester
	finishOnce sync.Once

	funcGet          func(key string) (sa1 []mm_pkg.Snapshot)
	funcGetOrigin    string
	inspectFuncGet   func(key string)
	afterGetCounter  uint64
	beforeGetCounter uint64
	GetMock          mSnapshotsMockGet
}

// NewSnapshotsMock returns a mock for mm_pkg.Snapshots
func NewSnapshotsMock(t minimock.Tester) *SnapshotsMock {
	m := &SnapshotsMock{t: t}

	if controller, ok := t.(minimock.MockController); ok {
		controller.RegisterMocker(m)
	}

	m.GetMock = mSnapshotsMockGet{mock: m}
	m.GetMock.callArgs = []*SnapshotsMockGetParams{}

	t.Cleanup(m.MinimockFinish)

	return m
}

type mSnapshotsMockGet struct {
	optional           bool
	mock               *SnapshotsMock
	defaultExpectation *SnapshotsMockGetExpectation
	expectations       []*SnapshotsMockGetExpectation

	callArgs []*SnapshotsMockGetParams
	mutex    sync.RWMutex

	expectedInvocations       uint64
	expectedInvocationsOrigin string
}

// SnapshotsMockGetExpectation specifies expectation struct of the Snapshots.Get
type SnapshotsMockGetExpectation struct {
	mock               *SnapshotsMock
	params             *SnapshotsMockGetParams
	paramPtrs          *SnapshotsMockGetParamPtrs
	expectationOrigins SnapshotsMockGetExpectationOrigins
	results            *SnapshotsMockGetResults
	returnOrigin       string
	Counter            uint64
}

// SnapshotsMockGetParams contains parameters of the Snapshots.Get
type SnapshotsMockGetParams struct {
	key string
}

// SnapshotsMockGetParamPtrs contains pointers to parameters of the Snapshots.Get
type SnapshotsMockGetParamPtrs struct {
	key *string
}

// SnapshotsMockGetResults contains results of the Snapshots.Get
type SnapshotsMockGetResults struct {
	sa1 []mm_pkg.Snapshot
}

// SnapshotsMockGetOrigins contains origins of expectations of the Snapshots.Get
type SnapshotsMockGetExpectationOrigins struct {
	origin    string
	originKey string
}

// Marks this method to be optional. The default behavior of any method with Return() is '1 or more', meaning
// the test will fail minimock's automatic final call check if the mocked method was not called at least once.
// Optional() makes method check to work in '0 or more' mode.
// It is NOT RECOMMENDED to use this option unless you really need it, as default behaviour helps to
// catch the problems when the expected method call is totally skipped during test run.
func (mmGet *mSnapshotsMockGet) Optional() *mSnapshotsMockGet {
	mmGet.optional = true
	return mmGet
}

// Expect sets up expected params for Snapshots.Get
func (mmGet *mSnapshotsMockGet) Expect(key string) *mSnapshotsMockGet {
	if mmGet.mock.funcGet != nil {
		mmGet.mock.t.Fatalf("SnapshotsMock.Get mock is already set by Set")
	}

	if mmGet.defaultExpectation == nil {
		mmGet.defaultExpectation = &SnapshotsMockGetExpectation{}
	}

	if mmGet.defaultExpectation.paramPtrs != nil {
		mmGet.mock.t.Fatalf("SnapshotsMock.Get mock is already set by ExpectParams functions")
	}

	mmGet.defaultExpectation.params = &SnapshotsMockGetParams{key}
	mmGet.defaultExpectation.expectationOrigins.origin = minimock.CallerInfo(1)
	for _, e := range mmGet.expectations {
		if minimock.Equal(e.params, mmGet.defaultExpectation.params) {
			mmGet.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmGet.defaultExpectation.params)
		}
	}

	return mmGet
}

// ExpectKeyParam1 sets up expected param key for Snapshots.Get
func (mmGet *mSnapshotsMockGet) ExpectKeyParam1(key string) *mSnapshotsMockGet {
	if mmGet.mock.funcGet != nil {
		mmGet.mock.t.Fatalf("SnapshotsMock.Get mock is already set by Set")
	}

	if mmGet.defaultExpectation == nil {
		mmGet.defaultExpectation = &SnapshotsMockGetExpectation{}
	}

	if mmGet.defaultExpectation.params != nil {
		mmGet.mock.t.Fatalf("SnapshotsMock.Get mock is already set by Expect")
	}

	if mmGet.defaultExpectation.paramPtrs == nil {
		mmGet.defaultExpectation.paramPtrs = &SnapshotsMockGetParamPtrs{}
	}
	mmGet.defaultExpectation.paramPtrs.key = &key
	mmGet.defaultExpectation.expectationOrigins.originKey = minimock.CallerInfo(1)

	return mmGet
}

// Inspect accepts an inspector function that has same arguments as the Snapshots.Get
func (mmGet *mSnapshotsMockGet) Inspect(f func(key string)) *mSnapshotsMockGet {
	if mmGet.mock.inspectFuncGet != nil {
		mmGet.mock.t.Fatalf("Inspect function is already set for SnapshotsMock.Get")
	}

	mmGet.mock.inspectFuncGet = f

	return mmGet
}

// Return sets up results that will be returned by Snapshots.Get
func (mmGet *mSnapshotsMockGet) Return(sa1 []mm_pkg.Snapshot) *SnapshotsMock {
	if mmGet.mock.funcGet != nil {
		mmGet.mock.t.Fatalf("SnapshotsMock.Get mock is already set by Set")
	}

	if mmGet.defaultExpectation == nil {
		mmGet.defaultExpectation = &SnapshotsMockGetExpectation{mock: mmGet.mock}
	}
	mmGet.defaultExpectation.results = &SnapshotsMockGetResults{sa1}
	mmGet.defaultExpectation.returnOrigin = minimock.CallerInfo(1)
	return mmGet.mock
}

// Set uses given function f to mock the Snapshots.Get method
func (mmGet *mSnapshotsMockGet) Set(f func(key string) (sa1 []mm_pkg.Snapshot)) *SnapshotsMock {
	if mmGet.defaultExpectation != nil {
		mmGet.mock.t.Fatalf("Default expectation is already set for the Snapshots.Get method")
	}

	if len(mmGet.expectations) > 0 {
		mmGet.mock.t.Fatalf("Some expectations are already set for the Snapshots.Get method")
	}

	mmGet.mock.funcGet = f
	mmGet.mock.funcGetOrigin = minimock.CallerInfo(1)
	return mmGet.mock
}

// When sets expectation for the Snapshots.Get which will trigger the result defined by the following
// Then helper
func (mmGet *mSnapshotsMockGet) When(key string) *SnapshotsMockGetExpectation {
	if mmGet.mock.funcGet != nil {
		mmGet.mock.t.Fatalf("SnapshotsMock.Get mock is already set by Set")
	}

	expectation := &SnapshotsMockGetExpectation{
		mock:               mmGet.mock,
		params:             &SnapshotsMockGetParams{key},
		expectationOrigins: SnapshotsMockGetExpectationOrigins{origin: minimock.CallerInfo(1)},
	}
	mmGet.expectations = append(mmGet.expectations, expectation)
	return expectation
}

// Then sets up Snapshots.Get return parameters for the expectation previously defined by the When method
func (e *SnapshotsMockGetExpectation) Then(sa1 []mm_pkg.Snapshot) *SnapshotsMock {
	e.results = &SnapshotsMockGetResults{sa1}
	return e.mock
}

// Times sets number of times Snapshots.Get should be invoked
func (mmGet *mSnapshotsMockGet) Times(n uint64) *mSnapshotsMockGet {
	if n == 0 {
		mmGet.mock.t.Fatalf("Times of SnapshotsMock.Get mock can not be zero")
	}
	mm_atomic.StoreUint64(&mmGet.expectedInvocations, n)
	mmGet.expectedInvocationsOrigin = minimock.CallerInfo(1)
	return mmGet
}

func (mmGet *mSnapshotsMockGet) invocationsDone() bool {
	if len(mmGet.expectations) == 0 && mmGet.defaultExpectation == nil && mmGet.mock.funcGet == nil {
		return true
	}

	totalInvocations := mm_atomic.LoadUint64(&mmGet.mock.afterGetCounter)
	expectedInvocations := mm_atomic.LoadUint64(&mmGet.expectedInvocations)

	return totalInvocations > 0 && (expectedInvocations == 0 || expectedInvocations == totalInvocations)
}

// Get implements mm_pkg.Snapshots
func (mmGet *SnapshotsMock) Get(key string) (sa1 []mm_pkg.Snapshot) {
	mm_atomic.AddUint64(&mmGet.beforeGetCounter, 1)
	defer mm_atomic.AddUint64(&mmGet.afterGetCounter, 1)

	mmGet.t.Helper()

	if mmGet.inspectFuncGet != nil {
		mmGet.inspectFuncGet(key)
	}

	mm_params := SnapshotsMockGetParams{key}

	// Record call args
	mmGet.GetMock.mutex.Lock()
	mmGet.GetMock.callArgs = append(mmGet.GetMock.callArgs, &mm_params)
	mmGet.GetMock.mutex.Unlock()

	for _, e := range mmGet.GetMock.expectations {
		if minimock.Equal(*e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return e.results.sa1
		}
	}

	if mmGet.GetMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmGet.GetMock.defaultExpectation.Counter, 1)
		mm_want := mmGet.GetMock.defaultExpectation.params
		mm_want_ptrs := mmGet.GetMock.defaultExpectation.paramPtrs

		mm_got := SnapshotsMockGetParams{key}

		if mm_want_ptrs != nil {

			if mm_want_ptrs.key != nil && !minimock.Equal(*mm_want_ptrs.key, mm_got.key) {
				mmGet.t.Errorf("SnapshotsMock.Get got unexpected parameter key, expected at\n%s:\nwant: %#v\n got: %#v%s\n",
					mmGet.GetMock.defaultExpectation.expectationOrigins.originKey, *mm_want_ptrs.key, mm_got.key, minimock.Diff(*mm_want_ptrs.key, mm_got.key))
			}

		} else if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmGet.t.Errorf("SnapshotsMock.Get got unexpected parameters, expected at\n%s:\nwant: %#v\n got: %#v%s\n",
				mmGet.GetMock.defaultExpectation.expectationOrigins.origin, *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmGet.GetMock.defaultExpectation.results
		if mm_results == nil {
			mmGet.t.Fatal("No results are set for the SnapshotsMock.Get")
		}
		return (*mm_results).sa1
	}
	if mmGet.funcGet != nil {
		return mmGet.funcGet(key)
	}
	mmGet.t.Fatalf("Unexpected call to SnapshotsMock.Get. %v", key)
	return
}

// GetAfterCounter returns a count of finished SnapshotsMock.Get invocations
func (mmGet *SnapshotsMock) GetAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmGet.afterGetCounter)
}

// GetBeforeCounter returns a count of SnapshotsMock.Get invocations
func (mmGet *SnapshotsMock) GetBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmGet.beforeGetCounter)
}

// Calls returns a list of arguments used in each call to SnapshotsMock.Get.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmGet *mSnapshotsMockGet) Calls() []*SnapshotsMockGetParams {
	mmGet.mutex.RLock()

	argCopy := make([]*SnapshotsMockGetParams, len(mmGet.callArgs))
	copy(argCopy, mmGet.callArgs)

	mmGet.mutex.RUnlock()

	return argCopy
}

// MinimockGetDone returns true if the count of the Get invocations corresponds
// the number of defined expectations
func (m *SnapshotsMock) MinimockGetDone() bool {
	if m.GetMock.optional {
		// Optional methods provide '0 or more' call count restriction.
		return true
	}

	for _, e := range m.GetMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	return m.GetMock.invocationsDone()
}

// MinimockGetInspect logs each unmet expectation
func (m *SnapshotsMock) MinimockGetInspect() {
	for _, e := range m.GetMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to SnapshotsMock.Get at\n%s with params: %#v", e.expectationOrigins.origin, *e.params)
		}
	}

	afterGetCounter := mm_atomic.LoadUint64(&m.afterGetCounter)
	// if default expectation was set then invocations count should be greater than zero
	if m.GetMock.defaultExpectation != nil && afterGetCounter < 1 {
		if m.GetMock.defaultExpectation.params == nil {
			m.t.Errorf("Expected call to SnapshotsMock.Get at\n%s", m.GetMock.defaultExpectation.returnOrigin)
		} else {
			m.t.Errorf("Expected call to SnapshotsMock.Get at\n%s with params: %#v", m.GetMock.defaultExpectation.expectationOrigins.origin, *m.GetMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcGet != nil && afterGetCounter < 1 {
		m.t.Errorf("Expected call to SnapshotsMock.Get at\n%s", m.funcGetOrigin)
	}

	if !m.GetMock.invocationsDone() && afterGetCounter > 0 {
		m.t.Errorf("Expected %d calls to SnapshotsMock.Get at\n%s but found %d calls",
			mm_atomic.LoadUint64(&m.GetMock.expectedInvocations), m.GetMock.expectedInvocationsOrigin, afterGetCounter)
	}
}

// MinimockFinish checks that all mocked methods have been called the expected number of times
func (m *SnapshotsMock) MinimockFinish() {
	m.finishOnce.Do(func() {
		if !m.minimockDone() {
			m.MinimockGetInspect()
		}
	})
}

// MinimockWait waits for all mocked methods to be called the expected number of times
func (m *SnapshotsMock) MinimockWait(timeout mm_time.Duration) {
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

func (m *SnapshotsMock) minimockDone() bool {
	done := true
	return done &&
		m.MinimockGetDone()
}

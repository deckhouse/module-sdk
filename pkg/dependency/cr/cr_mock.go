package cr

// Code generated by http://github.com/gojuno/minimock (dev). DO NOT EDIT.

//go:generate minimock -i github.com/deckhouse/module-sdk/pkg/dependency/cr.Client -o ./cr_mock.go

import (
	"context"
	"sync"
	mm_atomic "sync/atomic"
	mm_time "time"

	"github.com/gojuno/minimock/v3"
	v1 "github.com/google/go-containerregistry/pkg/v1"
)

// ClientMock implements Client
type ClientMock struct {
	t minimock.Tester

	funcDigest          func(tag string) (s1 string, err error)
	inspectFuncDigest   func(tag string)
	afterDigestCounter  uint64
	beforeDigestCounter uint64
	DigestMock          mClientMockDigest

	funcImage          func(tag string) (i1 v1.Image, err error)
	inspectFuncImage   func(tag string)
	afterImageCounter  uint64
	beforeImageCounter uint64
	ImageMock          mClientMockImage

	funcListTags          func() (sa1 []string, err error)
	inspectFuncListTags   func()
	afterListTagsCounter  uint64
	beforeListTagsCounter uint64
	ListTagsMock          mClientMockListTags
}

// NewClientMock returns a mock for Client
func NewClientMock(t minimock.Tester) *ClientMock {
	m := &ClientMock{t: t}
	if controller, ok := t.(minimock.MockController); ok {
		controller.RegisterMocker(m)
	}

	m.DigestMock = mClientMockDigest{mock: m}
	m.DigestMock.callArgs = []*ClientMockDigestParams{}

	m.ImageMock = mClientMockImage{mock: m}
	m.ImageMock.callArgs = []*ClientMockImageParams{}

	m.ListTagsMock = mClientMockListTags{mock: m}

	return m
}

type mClientMockDigest struct {
	mock               *ClientMock
	defaultExpectation *ClientMockDigestExpectation
	expectations       []*ClientMockDigestExpectation

	callArgs []*ClientMockDigestParams
	mutex    sync.RWMutex
}

// ClientMockDigestExpectation specifies expectation struct of the Client.Digest
type ClientMockDigestExpectation struct {
	mock    *ClientMock
	params  *ClientMockDigestParams
	results *ClientMockDigestResults
	Counter uint64
}

// ClientMockDigestParams contains parameters of the Client.Digest
type ClientMockDigestParams struct {
	tag string
}

// ClientMockDigestResults contains results of the Client.Digest
type ClientMockDigestResults struct {
	s1  string
	err error
}

// Expect sets up expected params for Client.Digest
func (mmDigest *mClientMockDigest) Expect(tag string) *mClientMockDigest {
	if mmDigest.mock.funcDigest != nil {
		mmDigest.mock.t.Fatalf("ClientMock.Digest mock is already set by Set")
	}

	if mmDigest.defaultExpectation == nil {
		mmDigest.defaultExpectation = &ClientMockDigestExpectation{}
	}

	mmDigest.defaultExpectation.params = &ClientMockDigestParams{tag}
	for _, e := range mmDigest.expectations {
		if minimock.Equal(e.params, mmDigest.defaultExpectation.params) {
			mmDigest.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmDigest.defaultExpectation.params)
		}
	}

	return mmDigest
}

// Inspect accepts an inspector function that has same arguments as the Client.Digest
func (mmDigest *mClientMockDigest) Inspect(f func(tag string)) *mClientMockDigest {
	if mmDigest.mock.inspectFuncDigest != nil {
		mmDigest.mock.t.Fatalf("Inspect function is already set for ClientMock.Digest")
	}

	mmDigest.mock.inspectFuncDigest = f

	return mmDigest
}

// Return sets up results that will be returned by Client.Digest
func (mmDigest *mClientMockDigest) Return(s1 string, err error) *ClientMock {
	if mmDigest.mock.funcDigest != nil {
		mmDigest.mock.t.Fatalf("ClientMock.Digest mock is already set by Set")
	}

	if mmDigest.defaultExpectation == nil {
		mmDigest.defaultExpectation = &ClientMockDigestExpectation{mock: mmDigest.mock}
	}
	mmDigest.defaultExpectation.results = &ClientMockDigestResults{s1, err}
	return mmDigest.mock
}

// Set uses given function f to mock the Client.Digest method
func (mmDigest *mClientMockDigest) Set(f func(tag string) (s1 string, err error)) *ClientMock {
	if mmDigest.defaultExpectation != nil {
		mmDigest.mock.t.Fatalf("Default expectation is already set for the Client.Digest method")
	}

	if len(mmDigest.expectations) > 0 {
		mmDigest.mock.t.Fatalf("Some expectations are already set for the Client.Digest method")
	}

	mmDigest.mock.funcDigest = f
	return mmDigest.mock
}

// When sets expectation for the Client.Digest which will trigger the result defined by the following
// Then helper
func (mmDigest *mClientMockDigest) When(tag string) *ClientMockDigestExpectation {
	if mmDigest.mock.funcDigest != nil {
		mmDigest.mock.t.Fatalf("ClientMock.Digest mock is already set by Set")
	}

	expectation := &ClientMockDigestExpectation{
		mock:   mmDigest.mock,
		params: &ClientMockDigestParams{tag},
	}
	mmDigest.expectations = append(mmDigest.expectations, expectation)
	return expectation
}

// Then sets up Client.Digest return parameters for the expectation previously defined by the When method
func (e *ClientMockDigestExpectation) Then(s1 string, err error) *ClientMock {
	e.results = &ClientMockDigestResults{s1, err}
	return e.mock
}

// Digest implements Client
func (mmDigest *ClientMock) Digest(_ context.Context, tag string) (s1 string, err error) {
	mm_atomic.AddUint64(&mmDigest.beforeDigestCounter, 1)
	defer mm_atomic.AddUint64(&mmDigest.afterDigestCounter, 1)

	if mmDigest.inspectFuncDigest != nil {
		mmDigest.inspectFuncDigest(tag)
	}

	mm_params := &ClientMockDigestParams{tag}

	// Record call args
	mmDigest.DigestMock.mutex.Lock()
	mmDigest.DigestMock.callArgs = append(mmDigest.DigestMock.callArgs, mm_params)
	mmDigest.DigestMock.mutex.Unlock()

	for _, e := range mmDigest.DigestMock.expectations {
		if minimock.Equal(e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return e.results.s1, e.results.err
		}
	}

	if mmDigest.DigestMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmDigest.DigestMock.defaultExpectation.Counter, 1)
		mm_want := mmDigest.DigestMock.defaultExpectation.params
		mm_got := ClientMockDigestParams{tag}
		if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmDigest.t.Errorf("ClientMock.Digest got unexpected parameters, want: %#v, got: %#v%s\n", *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmDigest.DigestMock.defaultExpectation.results
		if mm_results == nil {
			mmDigest.t.Fatal("No results are set for the ClientMock.Digest")
		}
		return (*mm_results).s1, (*mm_results).err
	}
	if mmDigest.funcDigest != nil {
		return mmDigest.funcDigest(tag)
	}
	mmDigest.t.Fatalf("Unexpected call to ClientMock.Digest. %v", tag)
	return
}

// DigestAfterCounter returns a count of finished ClientMock.Digest invocations
func (mmDigest *ClientMock) DigestAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmDigest.afterDigestCounter)
}

// DigestBeforeCounter returns a count of ClientMock.Digest invocations
func (mmDigest *ClientMock) DigestBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmDigest.beforeDigestCounter)
}

// Calls returns a list of arguments used in each call to ClientMock.Digest.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmDigest *mClientMockDigest) Calls() []*ClientMockDigestParams {
	mmDigest.mutex.RLock()

	argCopy := make([]*ClientMockDigestParams, len(mmDigest.callArgs))
	copy(argCopy, mmDigest.callArgs)

	mmDigest.mutex.RUnlock()

	return argCopy
}

// MinimockDigestDone returns true if the count of the Digest invocations corresponds
// the number of defined expectations
func (m *ClientMock) MinimockDigestDone() bool {
	for _, e := range m.DigestMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.DigestMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterDigestCounter) < 1 {
		return false
	}
	// if func was set then invocations count should be greater than zero
	if m.funcDigest != nil && mm_atomic.LoadUint64(&m.afterDigestCounter) < 1 {
		return false
	}
	return true
}

// MinimockDigestInspect logs each unmet expectation
func (m *ClientMock) MinimockDigestInspect() {
	for _, e := range m.DigestMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to ClientMock.Digest with params: %#v", *e.params)
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.DigestMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterDigestCounter) < 1 {
		if m.DigestMock.defaultExpectation.params == nil {
			m.t.Error("Expected call to ClientMock.Digest")
		} else {
			m.t.Errorf("Expected call to ClientMock.Digest with params: %#v", *m.DigestMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcDigest != nil && mm_atomic.LoadUint64(&m.afterDigestCounter) < 1 {
		m.t.Error("Expected call to ClientMock.Digest")
	}
}

type mClientMockImage struct {
	mock               *ClientMock
	defaultExpectation *ClientMockImageExpectation
	expectations       []*ClientMockImageExpectation

	callArgs []*ClientMockImageParams
	mutex    sync.RWMutex
}

// ClientMockImageExpectation specifies expectation struct of the Client.Image
type ClientMockImageExpectation struct {
	mock    *ClientMock
	params  *ClientMockImageParams
	results *ClientMockImageResults
	Counter uint64
}

// ClientMockImageParams contains parameters of the Client.Image
type ClientMockImageParams struct {
	tag string
}

// ClientMockImageResults contains results of the Client.Image
type ClientMockImageResults struct {
	i1  v1.Image
	err error
}

// Expect sets up expected params for Client.Image
func (mmImage *mClientMockImage) Expect(tag string) *mClientMockImage {
	if mmImage.mock.funcImage != nil {
		mmImage.mock.t.Fatalf("ClientMock.Image mock is already set by Set")
	}

	if mmImage.defaultExpectation == nil {
		mmImage.defaultExpectation = &ClientMockImageExpectation{}
	}

	mmImage.defaultExpectation.params = &ClientMockImageParams{tag}
	for _, e := range mmImage.expectations {
		if minimock.Equal(e.params, mmImage.defaultExpectation.params) {
			mmImage.mock.t.Fatalf("Expectation set by When has same params: %#v", *mmImage.defaultExpectation.params)
		}
	}

	return mmImage
}

// Inspect accepts an inspector function that has same arguments as the Client.Image
func (mmImage *mClientMockImage) Inspect(f func(tag string)) *mClientMockImage {
	if mmImage.mock.inspectFuncImage != nil {
		mmImage.mock.t.Fatalf("Inspect function is already set for ClientMock.Image")
	}

	mmImage.mock.inspectFuncImage = f

	return mmImage
}

// Return sets up results that will be returned by Client.Image
func (mmImage *mClientMockImage) Return(i1 v1.Image, err error) *ClientMock {
	if mmImage.mock.funcImage != nil {
		mmImage.mock.t.Fatalf("ClientMock.Image mock is already set by Set")
	}

	if mmImage.defaultExpectation == nil {
		mmImage.defaultExpectation = &ClientMockImageExpectation{mock: mmImage.mock}
	}
	mmImage.defaultExpectation.results = &ClientMockImageResults{i1, err}
	return mmImage.mock
}

// Set uses given function f to mock the Client.Image method
func (mmImage *mClientMockImage) Set(f func(tag string) (i1 v1.Image, err error)) *ClientMock {
	if mmImage.defaultExpectation != nil {
		mmImage.mock.t.Fatalf("Default expectation is already set for the Client.Image method")
	}

	if len(mmImage.expectations) > 0 {
		mmImage.mock.t.Fatalf("Some expectations are already set for the Client.Image method")
	}

	mmImage.mock.funcImage = f
	return mmImage.mock
}

// When sets expectation for the Client.Image which will trigger the result defined by the following
// Then helper
func (mmImage *mClientMockImage) When(tag string) *ClientMockImageExpectation {
	if mmImage.mock.funcImage != nil {
		mmImage.mock.t.Fatalf("ClientMock.Image mock is already set by Set")
	}

	expectation := &ClientMockImageExpectation{
		mock:   mmImage.mock,
		params: &ClientMockImageParams{tag},
	}
	mmImage.expectations = append(mmImage.expectations, expectation)
	return expectation
}

// Then sets up Client.Image return parameters for the expectation previously defined by the When method
func (e *ClientMockImageExpectation) Then(i1 v1.Image, err error) *ClientMock {
	e.results = &ClientMockImageResults{i1, err}
	return e.mock
}

// Image implements Client
func (mmImage *ClientMock) Image(_ context.Context, tag string) (i1 v1.Image, err error) {
	mm_atomic.AddUint64(&mmImage.beforeImageCounter, 1)
	defer mm_atomic.AddUint64(&mmImage.afterImageCounter, 1)

	if mmImage.inspectFuncImage != nil {
		mmImage.inspectFuncImage(tag)
	}

	mm_params := &ClientMockImageParams{tag}

	// Record call args
	mmImage.ImageMock.mutex.Lock()
	mmImage.ImageMock.callArgs = append(mmImage.ImageMock.callArgs, mm_params)
	mmImage.ImageMock.mutex.Unlock()

	for _, e := range mmImage.ImageMock.expectations {
		if minimock.Equal(e.params, mm_params) {
			mm_atomic.AddUint64(&e.Counter, 1)
			return e.results.i1, e.results.err
		}
	}

	if mmImage.ImageMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmImage.ImageMock.defaultExpectation.Counter, 1)
		mm_want := mmImage.ImageMock.defaultExpectation.params
		mm_got := ClientMockImageParams{tag}
		if mm_want != nil && !minimock.Equal(*mm_want, mm_got) {
			mmImage.t.Errorf("ClientMock.Image got unexpected parameters, want: %#v, got: %#v%s\n", *mm_want, mm_got, minimock.Diff(*mm_want, mm_got))
		}

		mm_results := mmImage.ImageMock.defaultExpectation.results
		if mm_results == nil {
			mmImage.t.Fatal("No results are set for the ClientMock.Image")
		}
		return (*mm_results).i1, (*mm_results).err
	}
	if mmImage.funcImage != nil {
		return mmImage.funcImage(tag)
	}
	mmImage.t.Fatalf("Unexpected call to ClientMock.Image. %v", tag)
	return
}

// ImageAfterCounter returns a count of finished ClientMock.Image invocations
func (mmImage *ClientMock) ImageAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmImage.afterImageCounter)
}

// ImageBeforeCounter returns a count of ClientMock.Image invocations
func (mmImage *ClientMock) ImageBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmImage.beforeImageCounter)
}

// Calls returns a list of arguments used in each call to ClientMock.Image.
// The list is in the same order as the calls were made (i.e. recent calls have a higher index)
func (mmImage *mClientMockImage) Calls() []*ClientMockImageParams {
	mmImage.mutex.RLock()

	argCopy := make([]*ClientMockImageParams, len(mmImage.callArgs))
	copy(argCopy, mmImage.callArgs)

	mmImage.mutex.RUnlock()

	return argCopy
}

// MinimockImageDone returns true if the count of the Image invocations corresponds
// the number of defined expectations
func (m *ClientMock) MinimockImageDone() bool {
	for _, e := range m.ImageMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.ImageMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterImageCounter) < 1 {
		return false
	}
	// if func was set then invocations count should be greater than zero
	if m.funcImage != nil && mm_atomic.LoadUint64(&m.afterImageCounter) < 1 {
		return false
	}
	return true
}

// MinimockImageInspect logs each unmet expectation
func (m *ClientMock) MinimockImageInspect() {
	for _, e := range m.ImageMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Errorf("Expected call to ClientMock.Image with params: %#v", *e.params)
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.ImageMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterImageCounter) < 1 {
		if m.ImageMock.defaultExpectation.params == nil {
			m.t.Error("Expected call to ClientMock.Image")
		} else {
			m.t.Errorf("Expected call to ClientMock.Image with params: %#v", *m.ImageMock.defaultExpectation.params)
		}
	}
	// if func was set then invocations count should be greater than zero
	if m.funcImage != nil && mm_atomic.LoadUint64(&m.afterImageCounter) < 1 {
		m.t.Error("Expected call to ClientMock.Image")
	}
}

type mClientMockListTags struct {
	mock               *ClientMock
	defaultExpectation *ClientMockListTagsExpectation
	expectations       []*ClientMockListTagsExpectation
}

// ClientMockListTagsExpectation specifies expectation struct of the Client.ListTags
type ClientMockListTagsExpectation struct {
	mock *ClientMock

	results *ClientMockListTagsResults
	Counter uint64
}

// ClientMockListTagsResults contains results of the Client.ListTags
type ClientMockListTagsResults struct {
	sa1 []string
	err error
}

// Expect sets up expected params for Client.ListTags
func (mmListTags *mClientMockListTags) Expect() *mClientMockListTags {
	if mmListTags.mock.funcListTags != nil {
		mmListTags.mock.t.Fatalf("ClientMock.ListTags mock is already set by Set")
	}

	if mmListTags.defaultExpectation == nil {
		mmListTags.defaultExpectation = &ClientMockListTagsExpectation{}
	}

	return mmListTags
}

// Inspect accepts an inspector function that has same arguments as the Client.ListTags
func (mmListTags *mClientMockListTags) Inspect(f func()) *mClientMockListTags {
	if mmListTags.mock.inspectFuncListTags != nil {
		mmListTags.mock.t.Fatalf("Inspect function is already set for ClientMock.ListTags")
	}

	mmListTags.mock.inspectFuncListTags = f

	return mmListTags
}

// Return sets up results that will be returned by Client.ListTags
func (mmListTags *mClientMockListTags) Return(sa1 []string, err error) *ClientMock {
	if mmListTags.mock.funcListTags != nil {
		mmListTags.mock.t.Fatalf("ClientMock.ListTags mock is already set by Set")
	}

	if mmListTags.defaultExpectation == nil {
		mmListTags.defaultExpectation = &ClientMockListTagsExpectation{mock: mmListTags.mock}
	}
	mmListTags.defaultExpectation.results = &ClientMockListTagsResults{sa1, err}
	return mmListTags.mock
}

// Set uses given function f to mock the Client.ListTags method
func (mmListTags *mClientMockListTags) Set(f func() (sa1 []string, err error)) *ClientMock {
	if mmListTags.defaultExpectation != nil {
		mmListTags.mock.t.Fatalf("Default expectation is already set for the Client.ListTags method")
	}

	if len(mmListTags.expectations) > 0 {
		mmListTags.mock.t.Fatalf("Some expectations are already set for the Client.ListTags method")
	}

	mmListTags.mock.funcListTags = f
	return mmListTags.mock
}

// ListTags implements Client
func (mmListTags *ClientMock) ListTags(_ context.Context) (sa1 []string, err error) {
	mm_atomic.AddUint64(&mmListTags.beforeListTagsCounter, 1)
	defer mm_atomic.AddUint64(&mmListTags.afterListTagsCounter, 1)

	if mmListTags.inspectFuncListTags != nil {
		mmListTags.inspectFuncListTags()
	}

	if mmListTags.ListTagsMock.defaultExpectation != nil {
		mm_atomic.AddUint64(&mmListTags.ListTagsMock.defaultExpectation.Counter, 1)

		mm_results := mmListTags.ListTagsMock.defaultExpectation.results
		if mm_results == nil {
			mmListTags.t.Fatal("No results are set for the ClientMock.ListTags")
		}
		return (*mm_results).sa1, (*mm_results).err
	}
	if mmListTags.funcListTags != nil {
		return mmListTags.funcListTags()
	}
	mmListTags.t.Fatalf("Unexpected call to ClientMock.ListTags.")
	return
}

// ListTagsAfterCounter returns a count of finished ClientMock.ListTags invocations
func (mmListTags *ClientMock) ListTagsAfterCounter() uint64 {
	return mm_atomic.LoadUint64(&mmListTags.afterListTagsCounter)
}

// ListTagsBeforeCounter returns a count of ClientMock.ListTags invocations
func (mmListTags *ClientMock) ListTagsBeforeCounter() uint64 {
	return mm_atomic.LoadUint64(&mmListTags.beforeListTagsCounter)
}

// MinimockListTagsDone returns true if the count of the ListTags invocations corresponds
// the number of defined expectations
func (m *ClientMock) MinimockListTagsDone() bool {
	for _, e := range m.ListTagsMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			return false
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.ListTagsMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterListTagsCounter) < 1 {
		return false
	}
	// if func was set then invocations count should be greater than zero
	if m.funcListTags != nil && mm_atomic.LoadUint64(&m.afterListTagsCounter) < 1 {
		return false
	}
	return true
}

// MinimockListTagsInspect logs each unmet expectation
func (m *ClientMock) MinimockListTagsInspect() {
	for _, e := range m.ListTagsMock.expectations {
		if mm_atomic.LoadUint64(&e.Counter) < 1 {
			m.t.Error("Expected call to ClientMock.ListTags")
		}
	}

	// if default expectation was set then invocations count should be greater than zero
	if m.ListTagsMock.defaultExpectation != nil && mm_atomic.LoadUint64(&m.afterListTagsCounter) < 1 {
		m.t.Error("Expected call to ClientMock.ListTags")
	}
	// if func was set then invocations count should be greater than zero
	if m.funcListTags != nil && mm_atomic.LoadUint64(&m.afterListTagsCounter) < 1 {
		m.t.Error("Expected call to ClientMock.ListTags")
	}
}

// MinimockFinish checks that all mocked methods have been called the expected number of times
func (m *ClientMock) MinimockFinish() {
	if !m.minimockDone() {
		m.MinimockDigestInspect()

		m.MinimockImageInspect()

		m.MinimockListTagsInspect()
		m.t.FailNow()
	}
}

// MinimockWait waits for all mocked methods to be called the expected number of times
func (m *ClientMock) MinimockWait(timeout mm_time.Duration) {
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

func (m *ClientMock) minimockDone() bool {
	done := true
	return done &&
		m.MinimockDigestDone() &&
		m.MinimockImageDone() &&
		m.MinimockListTagsDone()
}

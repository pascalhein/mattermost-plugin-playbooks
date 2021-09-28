// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/mattermost/mattermost-plugin-playbooks/server/app (interfaces: PlaybookRunStore)

// Package mock_app is a generated GoMock package.
package mock_app

import (
	gomock "github.com/golang/mock/gomock"
	app "github.com/mattermost/mattermost-plugin-playbooks/server/app"
	reflect "reflect"
	time "time"
)

// MockPlaybookRunStore is a mock of PlaybookRunStore interface
type MockPlaybookRunStore struct {
	ctrl     *gomock.Controller
	recorder *MockPlaybookRunStoreMockRecorder
}

// MockPlaybookRunStoreMockRecorder is the mock recorder for MockPlaybookRunStore
type MockPlaybookRunStoreMockRecorder struct {
	mock *MockPlaybookRunStore
}

// NewMockPlaybookRunStore creates a new mock instance
func NewMockPlaybookRunStore(ctrl *gomock.Controller) *MockPlaybookRunStore {
	mock := &MockPlaybookRunStore{ctrl: ctrl}
	mock.recorder = &MockPlaybookRunStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockPlaybookRunStore) EXPECT() *MockPlaybookRunStoreMockRecorder {
	return m.recorder
}

// ChangeCreationDate mocks base method
func (m *MockPlaybookRunStore) ChangeCreationDate(arg0 string, arg1 time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ChangeCreationDate", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// ChangeCreationDate indicates an expected call of ChangeCreationDate
func (mr *MockPlaybookRunStoreMockRecorder) ChangeCreationDate(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ChangeCreationDate", reflect.TypeOf((*MockPlaybookRunStore)(nil).ChangeCreationDate), arg0, arg1)
}

// CreatePlaybookRun mocks base method
func (m *MockPlaybookRunStore) CreatePlaybookRun(arg0 *app.PlaybookRun) (*app.PlaybookRun, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePlaybookRun", arg0)
	ret0, _ := ret[0].(*app.PlaybookRun)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreatePlaybookRun indicates an expected call of CreatePlaybookRun
func (mr *MockPlaybookRunStoreMockRecorder) CreatePlaybookRun(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePlaybookRun", reflect.TypeOf((*MockPlaybookRunStore)(nil).CreatePlaybookRun), arg0)
}

// CreateTimelineEvent mocks base method
func (m *MockPlaybookRunStore) CreateTimelineEvent(arg0 *app.TimelineEvent) (*app.TimelineEvent, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateTimelineEvent", arg0)
	ret0, _ := ret[0].(*app.TimelineEvent)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateTimelineEvent indicates an expected call of CreateTimelineEvent
func (mr *MockPlaybookRunStoreMockRecorder) CreateTimelineEvent(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateTimelineEvent", reflect.TypeOf((*MockPlaybookRunStore)(nil).CreateTimelineEvent), arg0)
}

// FinishPlaybookRun mocks base method
func (m *MockPlaybookRunStore) FinishPlaybookRun(arg0 string, arg1 int64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FinishPlaybookRun", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// FinishPlaybookRun indicates an expected call of FinishPlaybookRun
func (mr *MockPlaybookRunStoreMockRecorder) FinishPlaybookRun(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FinishPlaybookRun", reflect.TypeOf((*MockPlaybookRunStore)(nil).FinishPlaybookRun), arg0, arg1)
}

// GetAssignedTasks mocks base method
func (m *MockPlaybookRunStore) GetAssignedTasks(arg0 string) ([]app.AssignedRun, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAssignedTasks", arg0)
	ret0, _ := ret[0].([]app.AssignedRun)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAssignedTasks indicates an expected call of GetAssignedTasks
func (mr *MockPlaybookRunStoreMockRecorder) GetAssignedTasks(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAssignedTasks", reflect.TypeOf((*MockPlaybookRunStore)(nil).GetAssignedTasks), arg0)
}

// GetBroadcastChannelIDsToRootIDs mocks base method
func (m *MockPlaybookRunStore) GetBroadcastChannelIDsToRootIDs(arg0 string) (map[string]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBroadcastChannelIDsToRootIDs", arg0)
	ret0, _ := ret[0].(map[string]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBroadcastChannelIDsToRootIDs indicates an expected call of GetBroadcastChannelIDsToRootIDs
func (mr *MockPlaybookRunStoreMockRecorder) GetBroadcastChannelIDsToRootIDs(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBroadcastChannelIDsToRootIDs", reflect.TypeOf((*MockPlaybookRunStore)(nil).GetBroadcastChannelIDsToRootIDs), arg0)
}

// GetHistoricalPlaybookRunParticipantsCount mocks base method
func (m *MockPlaybookRunStore) GetHistoricalPlaybookRunParticipantsCount(arg0 string) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHistoricalPlaybookRunParticipantsCount", arg0)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHistoricalPlaybookRunParticipantsCount indicates an expected call of GetHistoricalPlaybookRunParticipantsCount
func (mr *MockPlaybookRunStoreMockRecorder) GetHistoricalPlaybookRunParticipantsCount(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHistoricalPlaybookRunParticipantsCount", reflect.TypeOf((*MockPlaybookRunStore)(nil).GetHistoricalPlaybookRunParticipantsCount), arg0)
}

// GetOwners mocks base method
func (m *MockPlaybookRunStore) GetOwners(arg0 app.RequesterInfo, arg1 app.PlaybookRunFilterOptions) ([]app.OwnerInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOwners", arg0, arg1)
	ret0, _ := ret[0].([]app.OwnerInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOwners indicates an expected call of GetOwners
func (mr *MockPlaybookRunStoreMockRecorder) GetOwners(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOwners", reflect.TypeOf((*MockPlaybookRunStore)(nil).GetOwners), arg0, arg1)
}

// GetPlaybookRun mocks base method
func (m *MockPlaybookRunStore) GetPlaybookRun(arg0 string) (*app.PlaybookRun, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPlaybookRun", arg0)
	ret0, _ := ret[0].(*app.PlaybookRun)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPlaybookRun indicates an expected call of GetPlaybookRun
func (mr *MockPlaybookRunStoreMockRecorder) GetPlaybookRun(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPlaybookRun", reflect.TypeOf((*MockPlaybookRunStore)(nil).GetPlaybookRun), arg0)
}

// GetPlaybookRunIDForChannel mocks base method
func (m *MockPlaybookRunStore) GetPlaybookRunIDForChannel(arg0 string) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPlaybookRunIDForChannel", arg0)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPlaybookRunIDForChannel indicates an expected call of GetPlaybookRunIDForChannel
func (mr *MockPlaybookRunStoreMockRecorder) GetPlaybookRunIDForChannel(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPlaybookRunIDForChannel", reflect.TypeOf((*MockPlaybookRunStore)(nil).GetPlaybookRunIDForChannel), arg0)
}

// GetPlaybookRuns mocks base method
func (m *MockPlaybookRunStore) GetPlaybookRuns(arg0 app.RequesterInfo, arg1 app.PlaybookRunFilterOptions) (*app.GetPlaybookRunsResults, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPlaybookRuns", arg0, arg1)
	ret0, _ := ret[0].(*app.GetPlaybookRunsResults)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPlaybookRuns indicates an expected call of GetPlaybookRuns
func (mr *MockPlaybookRunStoreMockRecorder) GetPlaybookRuns(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPlaybookRuns", reflect.TypeOf((*MockPlaybookRunStore)(nil).GetPlaybookRuns), arg0, arg1)
}

// GetTimelineEvent mocks base method
func (m *MockPlaybookRunStore) GetTimelineEvent(arg0, arg1 string) (*app.TimelineEvent, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetTimelineEvent", arg0, arg1)
	ret0, _ := ret[0].(*app.TimelineEvent)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetTimelineEvent indicates an expected call of GetTimelineEvent
func (mr *MockPlaybookRunStoreMockRecorder) GetTimelineEvent(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetTimelineEvent", reflect.TypeOf((*MockPlaybookRunStore)(nil).GetTimelineEvent), arg0, arg1)
}

// HasViewedChannel mocks base method
func (m *MockPlaybookRunStore) HasViewedChannel(arg0, arg1 string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasViewedChannel", arg0, arg1)
	ret0, _ := ret[0].(bool)
	return ret0
}

// HasViewedChannel indicates an expected call of HasViewedChannel
func (mr *MockPlaybookRunStoreMockRecorder) HasViewedChannel(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasViewedChannel", reflect.TypeOf((*MockPlaybookRunStore)(nil).HasViewedChannel), arg0, arg1)
}

// NukeDB mocks base method
func (m *MockPlaybookRunStore) NukeDB() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NukeDB")
	ret0, _ := ret[0].(error)
	return ret0
}

// NukeDB indicates an expected call of NukeDB
func (mr *MockPlaybookRunStoreMockRecorder) NukeDB() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NukeDB", reflect.TypeOf((*MockPlaybookRunStore)(nil).NukeDB))
}

// SetBroadcastChannelIDsToRootID mocks base method
func (m *MockPlaybookRunStore) SetBroadcastChannelIDsToRootID(arg0 string, arg1 map[string]string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetBroadcastChannelIDsToRootID", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetBroadcastChannelIDsToRootID indicates an expected call of SetBroadcastChannelIDsToRootID
func (mr *MockPlaybookRunStoreMockRecorder) SetBroadcastChannelIDsToRootID(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetBroadcastChannelIDsToRootID", reflect.TypeOf((*MockPlaybookRunStore)(nil).SetBroadcastChannelIDsToRootID), arg0, arg1)
}

// SetViewedChannel mocks base method
func (m *MockPlaybookRunStore) SetViewedChannel(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetViewedChannel", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetViewedChannel indicates an expected call of SetViewedChannel
func (mr *MockPlaybookRunStoreMockRecorder) SetViewedChannel(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetViewedChannel", reflect.TypeOf((*MockPlaybookRunStore)(nil).SetViewedChannel), arg0, arg1)
}

// UpdatePlaybookRun mocks base method
func (m *MockPlaybookRunStore) UpdatePlaybookRun(arg0 *app.PlaybookRun) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatePlaybookRun", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdatePlaybookRun indicates an expected call of UpdatePlaybookRun
func (mr *MockPlaybookRunStoreMockRecorder) UpdatePlaybookRun(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePlaybookRun", reflect.TypeOf((*MockPlaybookRunStore)(nil).UpdatePlaybookRun), arg0)
}

// UpdateStatus mocks base method
func (m *MockPlaybookRunStore) UpdateStatus(arg0 *app.SQLStatusPost) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateStatus", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateStatus indicates an expected call of UpdateStatus
func (mr *MockPlaybookRunStoreMockRecorder) UpdateStatus(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateStatus", reflect.TypeOf((*MockPlaybookRunStore)(nil).UpdateStatus), arg0)
}

// UpdateTimelineEvent mocks base method
func (m *MockPlaybookRunStore) UpdateTimelineEvent(arg0 *app.TimelineEvent) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateTimelineEvent", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateTimelineEvent indicates an expected call of UpdateTimelineEvent
func (mr *MockPlaybookRunStoreMockRecorder) UpdateTimelineEvent(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateTimelineEvent", reflect.TypeOf((*MockPlaybookRunStore)(nil).UpdateTimelineEvent), arg0)
}

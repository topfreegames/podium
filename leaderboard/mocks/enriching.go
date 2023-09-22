// Code generated by MockGen. DO NOT EDIT.
// Source: leaderboard/enriching/interfaces.go

// Package mock_enriching is a generated GoMock package.
package mock_enriching

import (
	context "context"
	reflect "reflect"
	time "time"

	gomock "github.com/golang/mock/gomock"
	model "github.com/topfreegames/podium/leaderboard/v2/model"
)

// MockEnricher is a mock of Enricher interface.
type MockEnricher struct {
	ctrl     *gomock.Controller
	recorder *MockEnricherMockRecorder
}

// MockEnricherMockRecorder is the mock recorder for MockEnricher.
type MockEnricherMockRecorder struct {
	mock *MockEnricher
}

// NewMockEnricher creates a new mock instance.
func NewMockEnricher(ctrl *gomock.Controller) *MockEnricher {
	mock := &MockEnricher{ctrl: ctrl}
	mock.recorder = &MockEnricherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEnricher) EXPECT() *MockEnricherMockRecorder {
	return m.recorder
}

// Enrich mocks base method.
func (m *MockEnricher) Enrich(ctx context.Context, tenantID, leaderboardID string, members []*model.Member) ([]*model.Member, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Enrich", ctx, tenantID, leaderboardID, members)
	ret0, _ := ret[0].([]*model.Member)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Enrich indicates an expected call of Enrich.
func (mr *MockEnricherMockRecorder) Enrich(ctx, tenantID, leaderboardID, members interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Enrich", reflect.TypeOf((*MockEnricher)(nil).Enrich), ctx, tenantID, leaderboardID, members)
}

// MockEnricherCache is a mock of EnricherCache interface.
type MockEnricherCache struct {
	ctrl     *gomock.Controller
	recorder *MockEnricherCacheMockRecorder
}

// MockEnricherCacheMockRecorder is the mock recorder for MockEnricherCache.
type MockEnricherCacheMockRecorder struct {
	mock *MockEnricherCache
}

// NewMockEnricherCache creates a new mock instance.
func NewMockEnricherCache(ctrl *gomock.Controller) *MockEnricherCache {
	mock := &MockEnricherCache{ctrl: ctrl}
	mock.recorder = &MockEnricherCacheMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockEnricherCache) EXPECT() *MockEnricherCacheMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockEnricherCache) Get(ctx context.Context, tenantID, leaderboardID string, members []*model.Member) (map[string]map[string]string, bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, tenantID, leaderboardID, members)
	ret0, _ := ret[0].(map[string]map[string]string)
	ret1, _ := ret[1].(bool)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// Get indicates an expected call of Get.
func (mr *MockEnricherCacheMockRecorder) Get(ctx, tenantID, leaderboardID, members interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockEnricherCache)(nil).Get), ctx, tenantID, leaderboardID, members)
}

// Set mocks base method.
func (m *MockEnricherCache) Set(ctx context.Context, tenantID, leaderboardID string, members []*model.Member, ttl time.Duration) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Set", ctx, tenantID, leaderboardID, members, ttl)
	ret0, _ := ret[0].(error)
	return ret0
}

// Set indicates an expected call of Set.
func (mr *MockEnricherCacheMockRecorder) Set(ctx, tenantID, leaderboardID, members, ttl interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Set", reflect.TypeOf((*MockEnricherCache)(nil).Set), ctx, tenantID, leaderboardID, members, ttl)
}

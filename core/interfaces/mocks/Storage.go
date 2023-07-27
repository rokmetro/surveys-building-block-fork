// Code generated by mockery v2.26.1. DO NOT EDIT.

package mocks

import (
	interfaces "application/core/interfaces"

	mock "github.com/stretchr/testify/mock"

	model "application/core/model"

	time "time"
)

// Storage is an autogenerated mock type for the Storage type
type Storage struct {
	mock.Mock
}

// CreateAlertContact provides a mock function with given fields: alertContact
func (_m *Storage) CreateAlertContact(alertContact model.AlertContact) (*model.AlertContact, error) {
	ret := _m.Called(alertContact)

	var r0 *model.AlertContact
	var r1 error
	if rf, ok := ret.Get(0).(func(model.AlertContact) (*model.AlertContact, error)); ok {
		return rf(alertContact)
	}
	if rf, ok := ret.Get(0).(func(model.AlertContact) *model.AlertContact); ok {
		r0 = rf(alertContact)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AlertContact)
		}
	}

	if rf, ok := ret.Get(1).(func(model.AlertContact) error); ok {
		r1 = rf(alertContact)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateSurvey provides a mock function with given fields: survey
func (_m *Storage) CreateSurvey(survey model.Survey) (*model.Survey, error) {
	ret := _m.Called(survey)

	var r0 *model.Survey
	var r1 error
	if rf, ok := ret.Get(0).(func(model.Survey) (*model.Survey, error)); ok {
		return rf(survey)
	}
	if rf, ok := ret.Get(0).(func(model.Survey) *model.Survey); ok {
		r0 = rf(survey)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Survey)
		}
	}

	if rf, ok := ret.Get(1).(func(model.Survey) error); ok {
		r1 = rf(survey)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateSurveyResponse provides a mock function with given fields: surveyResponse
func (_m *Storage) CreateSurveyResponse(surveyResponse model.SurveyResponse) (*model.SurveyResponse, error) {
	ret := _m.Called(surveyResponse)

	var r0 *model.SurveyResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(model.SurveyResponse) (*model.SurveyResponse, error)); ok {
		return rf(surveyResponse)
	}
	if rf, ok := ret.Get(0).(func(model.SurveyResponse) *model.SurveyResponse); ok {
		r0 = rf(surveyResponse)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.SurveyResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(model.SurveyResponse) error); ok {
		r1 = rf(surveyResponse)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteAlertContact provides a mock function with given fields: id, orgID, appID
func (_m *Storage) DeleteAlertContact(id string, orgID string, appID string) error {
	ret := _m.Called(id, orgID, appID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string) error); ok {
		r0 = rf(id, orgID, appID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteConfig provides a mock function with given fields: id
func (_m *Storage) DeleteConfig(id string) error {
	ret := _m.Called(id)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteSurvey provides a mock function with given fields: id, orgID, appID, creatorID
func (_m *Storage) DeleteSurvey(id string, orgID string, appID string, creatorID *string) error {
	ret := _m.Called(id, orgID, appID, creatorID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, *string) error); ok {
		r0 = rf(id, orgID, appID, creatorID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteSurveyResponse provides a mock function with given fields: id, orgID, appID, userID
func (_m *Storage) DeleteSurveyResponse(id string, orgID string, appID string, userID string) error {
	ret := _m.Called(id, orgID, appID, userID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string) error); ok {
		r0 = rf(id, orgID, appID, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteSurveyResponses provides a mock function with given fields: orgID, appID, userID, surveyIDs, surveyTypes, startDate, endDate
func (_m *Storage) DeleteSurveyResponses(orgID string, appID string, userID string, surveyIDs []string, surveyTypes []string, startDate *time.Time, endDate *time.Time) error {
	ret := _m.Called(orgID, appID, userID, surveyIDs, surveyTypes, startDate, endDate)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, []string, []string, *time.Time, *time.Time) error); ok {
		r0 = rf(orgID, appID, userID, surveyIDs, surveyTypes, startDate, endDate)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FindConfig provides a mock function with given fields: configType, appID, orgID
func (_m *Storage) FindConfig(configType string, appID string, orgID string) (*model.Config, error) {
	ret := _m.Called(configType, appID, orgID)

	var r0 *model.Config
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, string) (*model.Config, error)); ok {
		return rf(configType, appID, orgID)
	}
	if rf, ok := ret.Get(0).(func(string, string, string) *model.Config); ok {
		r0 = rf(configType, appID, orgID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Config)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(configType, appID, orgID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindConfigByID provides a mock function with given fields: id
func (_m *Storage) FindConfigByID(id string) (*model.Config, error) {
	ret := _m.Called(id)

	var r0 *model.Config
	var r1 error
	if rf, ok := ret.Get(0).(func(string) (*model.Config, error)); ok {
		return rf(id)
	}
	if rf, ok := ret.Get(0).(func(string) *model.Config); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Config)
		}
	}

	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindConfigs provides a mock function with given fields: configType
func (_m *Storage) FindConfigs(configType *string) ([]model.Config, error) {
	ret := _m.Called(configType)

	var r0 []model.Config
	var r1 error
	if rf, ok := ret.Get(0).(func(*string) ([]model.Config, error)); ok {
		return rf(configType)
	}
	if rf, ok := ret.Get(0).(func(*string) []model.Config); ok {
		r0 = rf(configType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.Config)
		}
	}

	if rf, ok := ret.Get(1).(func(*string) error); ok {
		r1 = rf(configType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAlertContact provides a mock function with given fields: id, orgID, appID
func (_m *Storage) GetAlertContact(id string, orgID string, appID string) (*model.AlertContact, error) {
	ret := _m.Called(id, orgID, appID)

	var r0 *model.AlertContact
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, string) (*model.AlertContact, error)); ok {
		return rf(id, orgID, appID)
	}
	if rf, ok := ret.Get(0).(func(string, string, string) *model.AlertContact); ok {
		r0 = rf(id, orgID, appID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AlertContact)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(id, orgID, appID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAlertContacts provides a mock function with given fields: orgID, appID
func (_m *Storage) GetAlertContacts(orgID string, appID string) ([]model.AlertContact, error) {
	ret := _m.Called(orgID, appID)

	var r0 []model.AlertContact
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) ([]model.AlertContact, error)); ok {
		return rf(orgID, appID)
	}
	if rf, ok := ret.Get(0).(func(string, string) []model.AlertContact); ok {
		r0 = rf(orgID, appID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.AlertContact)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(orgID, appID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAlertContactsByKey provides a mock function with given fields: key, orgID, appID
func (_m *Storage) GetAlertContactsByKey(key string, orgID string, appID string) ([]model.AlertContact, error) {
	ret := _m.Called(key, orgID, appID)

	var r0 []model.AlertContact
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, string) ([]model.AlertContact, error)); ok {
		return rf(key, orgID, appID)
	}
	if rf, ok := ret.Get(0).(func(string, string, string) []model.AlertContact); ok {
		r0 = rf(key, orgID, appID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.AlertContact)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(key, orgID, appID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSurvey provides a mock function with given fields: id, orgID, appID
func (_m *Storage) GetSurvey(id string, orgID string, appID string) (*model.Survey, error) {
	ret := _m.Called(id, orgID, appID)

	var r0 *model.Survey
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, string) (*model.Survey, error)); ok {
		return rf(id, orgID, appID)
	}
	if rf, ok := ret.Get(0).(func(string, string, string) *model.Survey); ok {
		r0 = rf(id, orgID, appID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Survey)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(id, orgID, appID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSurveyResponse provides a mock function with given fields: id, orgID, appID, userID
func (_m *Storage) GetSurveyResponse(id string, orgID string, appID string, userID string) (*model.SurveyResponse, error) {
	ret := _m.Called(id, orgID, appID, userID)

	var r0 *model.SurveyResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, string, string) (*model.SurveyResponse, error)); ok {
		return rf(id, orgID, appID, userID)
	}
	if rf, ok := ret.Get(0).(func(string, string, string, string) *model.SurveyResponse); ok {
		r0 = rf(id, orgID, appID, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.SurveyResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, string, string) error); ok {
		r1 = rf(id, orgID, appID, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSurveyResponses provides a mock function with given fields: orgID, appID, userID, surveyIDs, surveyTypes, startDate, endDate, limit, offset
func (_m *Storage) GetSurveyResponses(orgID *string, appID *string, userID *string, surveyIDs []string, surveyTypes []string, startDate *time.Time, endDate *time.Time, limit *int, offset *int) ([]model.SurveyResponse, error) {
	ret := _m.Called(orgID, appID, userID, surveyIDs, surveyTypes, startDate, endDate, limit, offset)

	var r0 []model.SurveyResponse
	var r1 error
	if rf, ok := ret.Get(0).(func(*string, *string, *string, []string, []string, *time.Time, *time.Time, *int, *int) ([]model.SurveyResponse, error)); ok {
		return rf(orgID, appID, userID, surveyIDs, surveyTypes, startDate, endDate, limit, offset)
	}
	if rf, ok := ret.Get(0).(func(*string, *string, *string, []string, []string, *time.Time, *time.Time, *int, *int) []model.SurveyResponse); ok {
		r0 = rf(orgID, appID, userID, surveyIDs, surveyTypes, startDate, endDate, limit, offset)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.SurveyResponse)
		}
	}

	if rf, ok := ret.Get(1).(func(*string, *string, *string, []string, []string, *time.Time, *time.Time, *int, *int) error); ok {
		r1 = rf(orgID, appID, userID, surveyIDs, surveyTypes, startDate, endDate, limit, offset)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSurveys provides a mock function with given fields: orgID, appID, creatorID, surveyIDs, surveyTypes, limit, offset
func (_m *Storage) GetSurveys(orgID string, appID string, creatorID *string, surveyIDs []string, surveyTypes []string, limit *int, offset *int) ([]model.Survey, error) {
	ret := _m.Called(orgID, appID, creatorID, surveyIDs, surveyTypes, limit, offset)

	var r0 []model.Survey
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string, *string, []string, []string, *int, *int) ([]model.Survey, error)); ok {
		return rf(orgID, appID, creatorID, surveyIDs, surveyTypes, limit, offset)
	}
	if rf, ok := ret.Get(0).(func(string, string, *string, []string, []string, *int, *int) []model.Survey); ok {
		r0 = rf(orgID, appID, creatorID, surveyIDs, surveyTypes, limit, offset)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.Survey)
		}
	}

	if rf, ok := ret.Get(1).(func(string, string, *string, []string, []string, *int, *int) error); ok {
		r1 = rf(orgID, appID, creatorID, surveyIDs, surveyTypes, limit, offset)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// InsertConfig provides a mock function with given fields: config
func (_m *Storage) InsertConfig(config model.Config) error {
	ret := _m.Called(config)

	var r0 error
	if rf, ok := ret.Get(0).(func(model.Config) error); ok {
		r0 = rf(config)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// PerformTransaction provides a mock function with given fields: _a0
func (_m *Storage) PerformTransaction(_a0 func(interfaces.Storage) error) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(func(interfaces.Storage) error) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RegisterStorageListener provides a mock function with given fields: listener
func (_m *Storage) RegisterStorageListener(listener interfaces.StorageListener) {
	_m.Called(listener)
}

// UpdateAlertContact provides a mock function with given fields: alertContact
func (_m *Storage) UpdateAlertContact(alertContact model.AlertContact) error {
	ret := _m.Called(alertContact)

	var r0 error
	if rf, ok := ret.Get(0).(func(model.AlertContact) error); ok {
		r0 = rf(alertContact)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateConfig provides a mock function with given fields: config
func (_m *Storage) UpdateConfig(config model.Config) error {
	ret := _m.Called(config)

	var r0 error
	if rf, ok := ret.Get(0).(func(model.Config) error); ok {
		r0 = rf(config)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateSurvey provides a mock function with given fields: survey, admin
func (_m *Storage) UpdateSurvey(survey model.Survey, admin bool) error {
	ret := _m.Called(survey, admin)

	var r0 error
	if rf, ok := ret.Get(0).(func(model.Survey, bool) error); ok {
		r0 = rf(survey, admin)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateSurveyResponse provides a mock function with given fields: surveyResponse
func (_m *Storage) UpdateSurveyResponse(surveyResponse model.SurveyResponse) error {
	ret := _m.Called(surveyResponse)

	var r0 error
	if rf, ok := ret.Get(0).(func(model.SurveyResponse) error); ok {
		r0 = rf(surveyResponse)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewStorage interface {
	mock.TestingT
	Cleanup(func())
}

// NewStorage creates a new instance of Storage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewStorage(t mockConstructorTestingTNewStorage) *Storage {
	mock := &Storage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

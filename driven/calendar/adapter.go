package calendar

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rokwire/core-auth-library-go/v3/authservice"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

const (
	// EventRoleAdmin represents the role string assigned to event admins
	EventRoleAdmin string = "admin"

	// TypeCalendarUser calendar.User type
	TypeCalendarUser logutils.MessageDataType = "calendar user"
)

// Adapter implements the Calendar interface
type Adapter struct {
	baseURL               string
	serviceAccountManager *authservice.ServiceAccountManager

	logger *logs.Logger
}

// TODO: adjust API definitions because it is work in progress

// EventPerson is defined from Calendar API
type EventPerson struct {
	User             User   `json:"user"`
	Registered       bool   `json:"registered"`
	Role             string `json:"role"`
	RegistrationType string `json:"registration_type"`
	Attended         bool   `json:"attended"`
	Time             int    `json:"time"`
}

// User defines users from EventPerson struct
type User struct {
	AccountID  string `json:"account_id"`
	ExternalID string `json:"external_id"`
}

// NewCalendarAdapter creates a new Calendar BB adapter instance
func NewCalendarAdapter(notificationHost string, serviceAccountManager *authservice.ServiceAccountManager, logger *logs.Logger) (*Adapter, error) {
	return &Adapter{baseURL: notificationHost, serviceAccountManager: serviceAccountManager, logger: logger}, nil
}

// GetEventUsers gets the event users through Calendar BB
func (a *Adapter) GetEventUsers(orgID string, appID string, eventID string, users []User, registered *bool, role string, attended *bool) ([]EventPerson, error) {
	if a.serviceAccountManager == nil {
		return nil, errors.Newf("service account manager is nil")
	}

	return a.getEventUsers(orgID, appID, eventID, users, registered, role, attended)
}

// gets the event users through Calendar BB
func (a *Adapter) getEventUsers(orgID string, appID string, eventID string, users []User, registered *bool, role string, attended *bool) ([]EventPerson, error) {
	url := fmt.Sprintf("%s/api/bbs/event/%s/users", a.baseURL, eventID)

	bodyData := map[string]interface{}{}

	bodyData["users"] = users

	if registered != nil {
		bodyData["registered"] = *registered
	}

	if len(role) > 0 {
		bodyData["role"] = role
	}

	if attended != nil {
		bodyData["attended"] = *attended
	}

	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionMarshal, logutils.TypeRequestBody, nil, err)
	}

	req, err := http.NewRequest("GET", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, logutils.TypeRequest, nil, err)
	}

	resp, err := a.serviceAccountManager.MakeRequest(req, appID, orgID)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionSend, logutils.TypeRequest, nil, err)
	}

	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Newf("request error response code (%d)", resp.Status)
	}
	if resp.StatusCode != 200 {
		return nil, errors.Newf("request error response code (%d): %s", resp.Status, respBytes)
	}

	var eventPersons []EventPerson
	err = json.Unmarshal(respBytes, &eventPersons)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionUnmarshal, logutils.TypeResponseBody, nil, err)
	}

	return eventPersons, nil
}

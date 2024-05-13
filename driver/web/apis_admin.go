// Copyright 2022 Board of Trustees of the University of Illinois.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package web

import (
	"application/core"
	"application/core/model"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/rokwire/core-auth-library-go/v3/authutils"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

// AdminAPIsHandler handles the rest Admin APIs implementation
type AdminAPIsHandler struct {
	app *core.Application
}

func (h AdminAPIsHandler) getConfig(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	config, err := h.app.Admin.GetConfig(id, claims)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeConfig, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(config)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeConfig, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h AdminAPIsHandler) getConfigs(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var configType *string
	typeParam := r.URL.Query().Get("type")
	if len(typeParam) > 0 {
		configType = &typeParam
	}

	configs, err := h.app.Admin.GetConfigs(configType, claims)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeConfig, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(configs)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeConfig, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

type adminUpdateConfigsRequest struct {
	AllApps *bool       `json:"all_apps,omitempty"`
	AllOrgs *bool       `json:"all_orgs,omitempty"`
	Data    interface{} `json:"data"`
	System  bool        `json:"system"`
	Type    string      `json:"type"`
}

func (h AdminAPIsHandler) createConfig(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var requestData adminUpdateConfigsRequest
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	appID := claims.AppID
	if requestData.AllApps != nil && *requestData.AllApps {
		appID = authutils.AllApps
	}
	orgID := claims.OrgID
	if requestData.AllOrgs != nil && *requestData.AllOrgs {
		orgID = authutils.AllOrgs
	}
	config := model.Config{Type: requestData.Type, AppID: appID, OrgID: orgID, System: requestData.System, Data: requestData.Data}

	newConfig, err := h.app.Admin.CreateConfig(config, claims)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionCreate, model.TypeConfig, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(newConfig)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, model.TypeConfig, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h AdminAPIsHandler) updateConfig(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	var requestData adminUpdateConfigsRequest
	err := json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUnmarshal, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	appID := claims.AppID
	if requestData.AllApps != nil && *requestData.AllApps {
		appID = authutils.AllApps
	}
	orgID := claims.OrgID
	if requestData.AllOrgs != nil && *requestData.AllOrgs {
		orgID = authutils.AllOrgs
	}
	config := model.Config{ID: id, Type: requestData.Type, AppID: appID, OrgID: orgID, System: requestData.System, Data: requestData.Data}

	err = h.app.Admin.UpdateConfig(config, claims)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, model.TypeConfig, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h AdminAPIsHandler) deleteConfig(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	params := mux.Vars(r)
	id := params["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	err := h.app.Admin.DeleteConfig(id, claims)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDelete, model.TypeConfig, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h AdminAPIsHandler) getSurvey(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	vars := mux.Vars(r)
	id := vars["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	resData, err := h.app.Admin.GetSurvey(id, claims.OrgID, claims.AppID)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeSurvey, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(resData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h AdminAPIsHandler) getSurveys(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	surveyIDsRaw := r.URL.Query().Get("ids")
	var surveyIDs []string
	if len(surveyIDsRaw) > 0 {
		surveyIDs = strings.Split(surveyIDsRaw, ",")
	}
	surveyTypesRaw := r.URL.Query().Get("types")
	var surveyTypes []string
	if len(surveyTypesRaw) > 0 {
		surveyTypes = strings.Split(surveyTypesRaw, ",")
	}

	calendarEventID := r.URL.Query().Get("calendar_event_id")

	limitRaw := r.URL.Query().Get("limit")
	limit := 20
	if len(limitRaw) > 0 {
		intParsed, err := strconv.Atoi(limitRaw)
		if err != nil {
			return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("limit"), nil, http.StatusBadRequest, false)
		}
		limit = intParsed
	}
	offsetRaw := r.URL.Query().Get("offset")
	offset := 0
	if len(offsetRaw) > 0 {
		intParsed, err := strconv.Atoi(offsetRaw)
		if err != nil {
			return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("offset"), nil, http.StatusBadRequest, false)
		}
		offset = intParsed
	}

	resData, err := h.app.Admin.GetSurveys(claims.OrgID, claims.AppID, nil, surveyIDs, surveyTypes, calendarEventID, &limit, &offset)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeSurvey, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(resData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h AdminAPIsHandler) createSurvey(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var item model.Survey
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	item.OrgID = claims.OrgID
	item.AppID = claims.AppID
	item.CreatorID = claims.Subject
	item.Type = "admin"

	createdItem, err := h.app.Admin.CreateSurvey(item, claims.ExternalIDs)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionCreate, model.TypeSurvey, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(createdItem)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)

}

func (h AdminAPIsHandler) updateSurvey(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	vars := mux.Vars(r)
	id := vars["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	var item model.Survey
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	item.ID = id
	item.OrgID = claims.OrgID
	item.AppID = claims.AppID
	item.Type = "admin"

	err = h.app.Admin.UpdateSurvey(item, claims.Subject, claims.ExternalIDs)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, model.TypeSurvey, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h AdminAPIsHandler) deleteSurvey(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	vars := mux.Vars(r)
	id := vars["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	err := h.app.Admin.DeleteSurvey(id, claims.OrgID, claims.AppID, claims.Subject, claims.ExternalIDs)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDelete, model.TypeSurvey, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h AdminAPIsHandler) getAllSurveyResponses(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	vars := mux.Vars(r)
	id := vars["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	startDateRaw := r.URL.Query().Get("start_date")
	var startDate *time.Time
	if len(startDateRaw) > 0 {
		dateParsed, err := time.Parse(time.RFC3339, startDateRaw)
		if err != nil {
			return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("start_date"), nil, http.StatusBadRequest, false)
		}
		startDate = &dateParsed
	}

	endDateRaw := r.URL.Query().Get("end_date")
	var endDate *time.Time
	if len(endDateRaw) > 0 {
		dateParsed, err := time.Parse(time.RFC3339, endDateRaw)
		if err != nil {
			return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("end_date"), nil, http.StatusBadRequest, false)
		}
		endDate = &dateParsed
	}

	limitRaw := r.URL.Query().Get("limit")
	limit := 20
	if len(limitRaw) > 0 {
		intParsed, err := strconv.Atoi(limitRaw)
		if err != nil {
			return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("limit"), nil, http.StatusBadRequest, false)
		}
		limit = intParsed
	}

	offsetRaw := r.URL.Query().Get("offset")
	offset := 0
	if len(offsetRaw) > 0 {
		intParsed, err := strconv.Atoi(offsetRaw)
		if err != nil {
			return l.HTTPResponseErrorData(logutils.StatusInvalid, logutils.TypeQueryParam, logutils.StringArgs("offset"), nil, http.StatusBadRequest, false)
		}
		offset = intParsed
	}

	resData, err := h.app.Admin.GetAllSurveyResponses(claims.OrgID, claims.AppID, id, claims.Subject, claims.ExternalIDs, startDate, endDate, &limit, &offset)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeSurvey, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(resData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}
	return l.HTTPResponseSuccessJSON(data)
}

func (h AdminAPIsHandler) getAlertContacts(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	resData, err := h.app.Admin.GetAlertContacts(claims.OrgID, claims.AppID)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeAlertContact, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(resData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h AdminAPIsHandler) getAlertContact(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	vars := mux.Vars(r)
	id := vars["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	resData, err := h.app.Admin.GetAlertContact(id, claims.OrgID, claims.AppID)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeAlertContact, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(resData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h AdminAPIsHandler) createAlertContact(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var item model.AlertContact
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	item.OrgID = claims.OrgID
	item.AppID = claims.AppID

	createdItem, err := h.app.Admin.CreateAlertContact(item)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionCreate, model.TypeAlertContact, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(createdItem)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h AdminAPIsHandler) updateAlertContact(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	vars := mux.Vars(r)
	id := vars["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	var item model.AlertContact
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	item.ID = id
	item.OrgID = claims.OrgID
	item.AppID = claims.AppID

	err = h.app.Admin.UpdateAlertContact(item)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, model.TypeAlertContact, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h AdminAPIsHandler) deleteAlertContact(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	vars := mux.Vars(r)
	id := vars["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	err := h.app.Admin.DeleteAlertContact(id, claims.OrgID, claims.AppID)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDelete, model.TypeAlertContact, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

// NewAdminAPIsHandler creates new rest Handler instance
func NewAdminAPIsHandler(app *core.Application) AdminAPIsHandler {
	return AdminAPIsHandler{app: app}
}

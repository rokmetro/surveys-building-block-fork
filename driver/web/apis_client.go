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
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

// ClientAPIsHandler handles the client rest APIs implementation
type ClientAPIsHandler struct {
	app *core.Application
}

func (h ClientAPIsHandler) getSurvey(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	vars := mux.Vars(r)
	id := vars["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	resData, err := h.app.Client.GetSurvey(id, claims.OrgID, claims.AppID)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeSurvey, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(resData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h ClientAPIsHandler) getSurveys(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
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
	limit := 0
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

	publicStr := r.URL.Query().Get("public")

	var public *bool

	if publicStr != "" {
		valuePublic, err := strconv.ParseBool(publicStr)
		if err != nil {
			return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeSurvey, nil, err, http.StatusInternalServerError, true)
		}
		public = &valuePublic
	}

	archivedStr := r.URL.Query().Get("archived")

	var archived *bool

	if archivedStr != "" {
		valueArchived, err := strconv.ParseBool(archivedStr)
		if err != nil {
			return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeSurvey, nil, err, http.StatusInternalServerError, true)
		}
		archived = &valueArchived
	}

	completedStr := r.URL.Query().Get("completed")

	var completed *bool

	if completedStr != "" {
		valueCompleted, err := strconv.ParseBool(completedStr)
		if err != nil {
			return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeSurvey, nil, err, http.StatusInternalServerError, true)
		}
		completed = &valueCompleted
	}

	var timeFilterItems model.SurveyTimeFilterRequest
	startsBeforeRaw := r.URL.Query().Get("starts_before")
	if startsBeforeRaw != "" {
		timeFilterItems.StartTimeBefore = &startsBeforeRaw
	}
	startsAfterRaw := r.URL.Query().Get("starts_after")
	if startsAfterRaw != "" {
		timeFilterItems.StartTimeAfter = &startsAfterRaw
	}
	endsBeforeRaw := r.URL.Query().Get("ends_before")
	if endsBeforeRaw != "" {
		timeFilterItems.EndTimeBefore = &endsBeforeRaw
	}
	endsAfterRaw := r.URL.Query().Get("ends_after")
	if endsAfterRaw != "" {
		timeFilterItems.EndTimeAfter = &endsAfterRaw
	}
	filter := surveyTimeFilter(&timeFilterItems)

	surveys, surverysRsponse, err := h.app.Client.GetSurveys(claims.OrgID, claims.AppID, &claims.Subject, nil, surveyIDs, surveyTypes, calendarEventID,
		&limit, &offset, filter, public, archived, completed)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeSurvey, nil, err, http.StatusInternalServerError, true)
	}

	list := getSurveysResData(surveys, surverysRsponse, completed)
	respData := sortIfpublicIsTrue(list, public)

	// Set response to nil to indicate last page and no more results should be loaded
	if len(surveys) == 0 {
		respData = nil
	} else if respData == nil {
		respData = []model.SurveysResponseData{}
	}

	rdata, err := json.Marshal(respData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(rdata)
}

func (h ClientAPIsHandler) createSurvey(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var items model.SurveyRequest
	err := json.NewDecoder(r.Body).Decode(&items)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}
	items.CreatorID = claims.Subject
	items.OrgID = claims.OrgID
	items.AppID = claims.AppID
	items.Type = "user"
	item := surveyRequestToSurvey(items)

	createdItem, err := h.app.Client.CreateSurvey(item, claims.ExternalIDs)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionCreate, model.TypeSurvey, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(createdItem)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)

}

func (h ClientAPIsHandler) updateSurvey(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	vars := mux.Vars(r)
	id := vars["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	var items model.SurveyRequest
	err := json.NewDecoder(r.Body).Decode(&items)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}
	items.CreatorID = claims.Subject
	items.OrgID = claims.OrgID
	items.AppID = claims.AppID
	items.Type = "user"

	item := updateSurveyRequestToSurvey(items, id)

	err = h.app.Client.UpdateSurvey(item, claims.Subject, claims.ExternalIDs)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, model.TypeSurvey, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h ClientAPIsHandler) deleteSurvey(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	vars := mux.Vars(r)
	id := vars["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	err := h.app.Client.DeleteSurvey(id, claims.OrgID, claims.AppID, claims.Subject, claims.ExternalIDs)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDelete, model.TypeSurvey, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h ClientAPIsHandler) getAllSurveyResponses(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	vars := mux.Vars(r)
	surveyID := vars["id"]

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

	resData, err := h.app.Client.GetAllSurveyResponses(claims.OrgID, claims.AppID, claims.Subject, surveyID, startDate, endDate, &limit, &offset, claims.ExternalIDs)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeSurvey, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(resData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h ClientAPIsHandler) getUserSurveyResponses(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	surveyIDsRaw := r.URL.Query().Get("survey_ids")
	var surveyIDs []string
	if len(surveyIDsRaw) > 0 {
		surveyIDs = strings.Split(surveyIDsRaw, ",")
	}
	surveyTypesRaw := r.URL.Query().Get("survey_types")
	var surveyTypes []string
	if len(surveyTypesRaw) > 0 {
		surveyTypes = strings.Split(surveyTypesRaw, ",")
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

	resData, err := h.app.Client.GetUserSurveyResponses(claims.OrgID, claims.AppID, claims.Subject, surveyIDs, surveyTypes, startDate, endDate, &limit, &offset)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeSurveyResponse, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(resData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h ClientAPIsHandler) getSurveyResponse(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	vars := mux.Vars(r)
	id := vars["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	resData, err := h.app.Client.GetSurveyResponse(id, claims.OrgID, claims.AppID, claims.Subject)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeSurveyResponse, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(resData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h ClientAPIsHandler) createSurveyResponse(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var item model.Survey
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	item.OrgID = claims.OrgID
	item.AppID = claims.AppID
	item.CreatorID = claims.Subject

	createdItem, err := h.app.Client.CreateSurveyResponse(model.SurveyResponse{UserID: claims.Subject, AppID: claims.AppID, OrgID: claims.OrgID, Survey: item}, claims.ExternalIDs)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionCreate, model.TypeSurveyResponse, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(createdItem)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h ClientAPIsHandler) updateSurveyResponse(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
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
	item.CreatorID = claims.Subject

	err = h.app.Client.UpdateSurveyResponse(model.SurveyResponse{UserID: claims.Subject, AppID: claims.AppID, OrgID: claims.OrgID, Survey: item})
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionUpdate, model.TypeSurveyResponse, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h ClientAPIsHandler) deleteSurveyResponse(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	vars := mux.Vars(r)
	id := vars["id"]
	if len(id) <= 0 {
		return l.HTTPResponseErrorData(logutils.StatusMissing, logutils.TypePathParam, logutils.StringArgs("id"), nil, http.StatusBadRequest, false)
	}

	err := h.app.Client.DeleteSurveyResponse(id, claims.OrgID, claims.AppID, claims.Subject)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDelete, model.TypeSurveyResponse, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h ClientAPIsHandler) deleteSurveyResponses(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	surveyIDsRaw := r.URL.Query().Get("survey_ids")
	var surveyIDs []string
	if len(surveyIDsRaw) > 0 {
		surveyIDs = strings.Split(surveyIDsRaw, ",")
	}
	surveyTypesRaw := r.URL.Query().Get("survey_types")
	var surveyTypes []string
	if len(surveyTypesRaw) > 0 {
		surveyTypes = strings.Split(surveyTypesRaw, ",")
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

	err := h.app.Client.DeleteSurveyResponses(claims.OrgID, claims.AppID, claims.Subject, surveyIDs, surveyTypes, startDate, endDate)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDelete, model.TypeSurveyResponse, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h ClientAPIsHandler) createSurveyAlert(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	var item model.SurveyAlert
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionDecode, logutils.TypeRequestBody, nil, err, http.StatusBadRequest, true)
	}

	item.OrgID = claims.OrgID
	item.AppID = claims.AppID

	err = h.app.Client.CreateSurveyAlert(item)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionCreate, model.TypeSurveyAlert, nil, err, http.StatusInternalServerError, true)
	}

	return l.HTTPResponseSuccess()
}

func (h ClientAPIsHandler) getCreatorSurveys(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
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

	publicStr := r.URL.Query().Get("public")

	var public *bool

	if publicStr != "" {
		valuePublic, err := strconv.ParseBool(publicStr)
		if err != nil {
			return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeSurvey, nil, err, http.StatusInternalServerError, true)
		}
		public = &valuePublic
	}

	archivedStr := r.URL.Query().Get("archived")

	var archived *bool

	if archivedStr != "" {
		valueArchived, err := strconv.ParseBool(archivedStr)
		if err != nil {
			return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeSurvey, nil, err, http.StatusInternalServerError, true)
		}
		archived = &valueArchived
	}

	completedStr := r.URL.Query().Get("completed")

	var completed *bool

	if completedStr != "" {
		valueCompleted, err := strconv.ParseBool(completedStr)
		if err != nil {
			return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeSurvey, nil, err, http.StatusInternalServerError, true)
		}
		completed = &valueCompleted
	}

	var timeFilterItems model.SurveyTimeFilterRequest
	startsBeforeRaw := r.URL.Query().Get("starts_before")
	if startsBeforeRaw != "" {
		timeFilterItems.StartTimeBefore = &startsBeforeRaw
	}
	startsAfterRaw := r.URL.Query().Get("starts_after")
	if startsAfterRaw != "" {
		timeFilterItems.StartTimeAfter = &startsAfterRaw
	}
	endsBeforeRaw := r.URL.Query().Get("ends_before")
	if endsBeforeRaw != "" {
		timeFilterItems.EndTimeBefore = &endsBeforeRaw
	}
	endsAfterRaw := r.URL.Query().Get("ends_after")
	if endsAfterRaw != "" {
		timeFilterItems.EndTimeAfter = &endsAfterRaw
	}
	filter := surveyTimeFilter(&timeFilterItems)

	resData, _, err := h.app.Client.GetSurveys(claims.OrgID, claims.AppID, &claims.Subject, &claims.Subject, surveyIDs, surveyTypes, "", &limit, &offset, filter, public, archived, completed)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeSurvey, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(resData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

func (h ClientAPIsHandler) getUserData(l *logs.Log, r *http.Request, claims *tokenauth.Claims) logs.HTTPResponse {
	resData, err := h.app.Client.GetUserData(claims.OrgID, claims.AppID, &claims.Subject)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionGet, model.TypeSurvey, nil, err, http.StatusInternalServerError, true)
	}

	data, err := json.Marshal(resData)
	if err != nil {
		return l.HTTPResponseErrorAction(logutils.ActionMarshal, logutils.TypeResponseBody, nil, err, http.StatusInternalServerError, false)
	}

	return l.HTTPResponseSuccessJSON(data)
}

// NewClientAPIsHandler creates new client API handler instance
func NewClientAPIsHandler(app *core.Application) ClientAPIsHandler {
	return ClientAPIsHandler{app: app}
}

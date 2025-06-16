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

package model

import (
	"time"

	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
)

const (
	//TypeSurveyAlert example type
	TypeSurveyAlert logutils.MessageDataType = "survey alert"
	//TypeAlertContact example type
	TypeAlertContact logutils.MessageDataType = "alert contact"
)

// SurveyAlert is a survey alert to be sent to notifications BB
type SurveyAlert struct {
	OrgID      string                 `json:"org_id" bson:"org_id"`
	AppID      string                 `json:"app_id" bson:"app_id"`
	ContactKey string                 `json:"contact_key" bson:"contact_key"`
	Content    map[string]interface{} `json:"content" bson:"content"`
}

// AlertContact is what will be used to identify where to send survey alerts
type AlertContact struct {
	ID          string                 `json:"id" bson:"_id"`
	OrgID       string                 `json:"org_id" bson:"org_id"`
	AppID       string                 `json:"app_id" bson:"app_id"`
	Key         string                 `json:"key" bson:"key"`
	Type        string                 `json:"type" bson:"type"`
	Address     string                 `json:"address" bson:"address"`
	Params      map[string]interface{} `json:"params" bson:"params"`
	DateCreated time.Time              `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time             `json:"date_updated" bson:"date_updated"`
}

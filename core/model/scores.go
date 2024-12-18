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
)

// Score object maintains data of accumulated scores from surveys
type Score struct {
	UserID            string        `json:"user_id" bson:"user_id"`
	OrgID             string        `json:"org_id" bson:"org_id"`
	AppID             string        `json:"app_id" bson:"app_id"`
	ExternalProfileID string        `json:"external_profile_id" bson:"external_profile_id"`
	TotalScore        uint32        `json:"total_score" bson:"total_score"`
	Scores            []SurveyScore `json:"scores" bson:"scores"`
}

// SurveyScore maintains individual score for specific survey
type SurveyScore struct {
	SurveyID    string    `json:"survey_id" bson:"survey_id"`
	Score       uint32    `json:"score" bson:"score"`
	DateCreated time.Time `json:"date_created" bson:"date_created"`
}

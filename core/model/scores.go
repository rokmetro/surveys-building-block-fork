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

	"github.com/rokwire/logging-library-go/v2/logutils"
)

const (
	TypeScore logutils.MessageDataType = "score"
)

// Score object maintains data of accumulated scores from surveys
type Score struct {
	ID                     string    `json:"id" bson:"_id"`
	OrgID                  string    `json:"org_id" bson:"org_id"`
	AppID                  string    `json:"app_id" bson:"app_id"`
	UserID                 string    `json:"user_id" bson:"user_id"`
	ExternalProfileID      string    `json:"external_profile_id" bson:"external_profile_id"`
	Score                  uint32    `json:"score" bson:"score"`
	ResponseCount          uint32    `json:"response_count" bson:"response_count"`
	PrevSurveyResponseDate time.Time `json:"prev_survey_response_date" bson:"prev_survey_response_date"`
	CurrentStreak          uint32    `json:"current_streak" bson:"current_streak"`
	AnswerCount            uint32    `json:"answer_count" bson:"answer_count"`
	CorrectAnswerCount     uint32    `json:"correct_answer_count" bson:"correct_answer_count"`
}

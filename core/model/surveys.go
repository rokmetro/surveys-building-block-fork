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
	//TypeSurvey example type
	TypeSurvey logutils.MessageDataType = "survey"
	//TypeSurveyResponse example type
	TypeSurveyResponse logutils.MessageDataType = "survey response"
)

// SurveyResponse wraps the entire survey response
type SurveyResponse struct {
	ID          string     `json:"id" bson:"_id"`
	UserID      string     `json:"user_id" bson:"user_id"`
	OrgID       string     `json:"org_id" bson:"org_id"`
	AppID       string     `json:"app_id" bson:"app_id"`
	Survey      Survey     `json:"survey" bson:"survey"`
	DateCreated time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time `json:"date_updated" bson:"date_updated"`
}

// Survey wraps the entire record
type Survey struct {
	ID                 string                 `json:"id" bson:"_id"`
	CreatorID          string                 `json:"creator_id" bson:"creator_id"`
	OrgID              string                 `json:"org_id" bson:"org_id"`
	AppID              string                 `json:"app_id" bson:"app_id"`
	Title              string                 `json:"title" bson:"title"`
	MoreInfo           *string                `json:"more_info" bson:"more_info"`
	Data               map[string]SurveyData  `json:"data" bson:"data"`
	Scored             bool                   `json:"scored" bson:"scored"`
	ResultRules        string                 `json:"result_rules" bson:"result_rules"`
	ResultJSON         string                 `json:"result_json" bson:"result_json"`
	Type               string                 `json:"type" bson:"type"`
	SurveyStats        *SurveyStats           `json:"stats" bson:"stats"`
	Sensitive          bool                   `json:"sensitive" bson:"sensitive"`
	Anonymous          bool                   `json:"anonymous" bson:"anonymous"`
	DefaultDataKey     *string                `json:"default_data_key" bson:"default_data_key"`
	DefaultDataKeyRule *string                `json:"default_data_key_rule" bson:"default_data_key_rule"`
	Constants          map[string]interface{} `json:"constants" bson:"constants"`
	Strings            map[string]interface{} `json:"strings" bson:"strings"`
	SubRules           map[string]interface{} `json:"sub_rules" bson:"sub_rules"`
	ResponseKeys       []string               `json:"response_keys" bson:"response_keys"`
	DateCreated        time.Time              `json:"date_created" bson:"date_created"`
	DateUpdated        *time.Time             `json:"date_updated" bson:"date_updated"`
	CalendarEventID    string                 `json:"calendar_event_id" bson:"calendar_event_id"`
}

// SurveyResponseAnonymous represents an anonymized survey response
type SurveyResponseAnonymous struct {
	ID          string       `json:"id"`
	CreatorID   string       `json:"creator_id"`
	OrgID       string       `json:"org_id"`
	AppID       string       `json:"app_id"`
	Title       string       `json:"title"`
	Type        string       `json:"type"`
	SurveyStats *SurveyStats `json:"stats"`
	DateCreated time.Time    `json:"date_created"`
	DateUpdated *time.Time   `json:"date_updated,omitempty"`
}

// SurveyStats are stats of a Survey
type SurveyStats struct {
	Total         int                    `json:"total" bson:"total"`
	Complete      int                    `json:"complete" bson:"complete"`
	Scored        int                    `json:"scored" bson:"scored"`
	Scores        map[string]float64     `json:"scores" bson:"scores"`
	MaximumScores map[string]float64     `json:"maximum_scores" bson:"maximum_scores"`
	ResponseData  map[string]interface{} `json:"response_data" bson:"response_data"`
}

// SurveyData is data stored for a Survey
type SurveyData struct {
	Section             *string     `json:"section,omitempty" bson:"section,omitempty"`
	Sections            []string    `json:"sections,omitempty" bson:"sections,omitempty"`
	AllowSkip           bool        `json:"allow_skip" bson:"allow_skip"`
	Text                string      `json:"text" bson:"text"`
	MoreInfo            string      `json:"more_info" bson:"more_info"`
	DefaultFollowUpKey  *string     `json:"default_follow_up_key" bson:"default_follow_up_key"`
	DefaultResponseRule *string     `json:"default_response_rule" bson:"default_response_rule"`
	FollowUpRule        *string     `json:"follow_up_rule" bson:"follow_up_rule"`
	ScoreRule           *string     `json:"score_rule" bson:"score_rule"`
	Replace             bool        `json:"replace" bson:"replace"`
	Response            interface{} `json:"response" bson:"response"`

	Type string `json:"type" bson:"type"`

	// Shared
	CorrectAnswer  interface{}   `json:"correct_answer,omitempty" bson:"correct_answer,omitempty"`
	CorrectAnswers []interface{} `json:"correct_answers,omitempty" bson:"correct_answers,omitempty"`
	Options        []OptionData  `json:"options,omitempty" bson:"options,omitempty"`
	Actions        *[]ActionData `json:"actions,omitempty" bson:"actions,omitempty"`
	SelfScore      *bool         `json:"self_score,omitempty" bson:"self_score,omitempty"`
	MaximumScore   *float64      `json:"maximum_score,omitempty" bson:"maximum_score,omitempty"`
	Style          *string       `json:"style,omitempty" bson:"style,omitempty"`

	// Multiple Choice
	AllowMultiple *bool `json:"allow_multiple,omitempty" bson:"allow_multiple,omitempty"`

	// DateTime
	StartTime *time.Time `json:"start_time,omitempty" bson:"start_time,omitempty"`
	EndTime   *time.Time `json:"end_time,omitempty" bson:"end_time,omitempty"`
	AskTime   *bool      `json:"ask_time,omitempty" bson:"ask_time,omitempty"`

	// Numeric
	Minimum  *float64 `json:"minimum,omitempty" bson:"minimum,omitempty"`
	Maximum  *float64 `json:"maximum,omitempty" bson:"maximum,omitempty"`
	WholeNum *bool    `json:"whole_num,omitempty" bson:"whole_num,omitempty"`

	// Text
	MinLength *int `json:"min_length,omitempty" bson:"min_length,omitempty"`
	MaxLength *int `json:"max_length,omitempty" bson:"max_length,omitempty"`

	// DataEntry
	DataFormat map[string]string `json:"data_format,omitempty" bson:"data_format,omitempty"`

	// Page
	DataKeys []string `json:"data_keys,omitempty" bson:"data_keys,omitempty"`
}

// ActionData is the wrapped within SurveyData
type ActionData struct {
	Type   string                 `json:"type" bson:"type"`
	Label  *string                `json:"label" bson:"label"`
	Data   string                 `json:"data" bson:"data"`
	Params map[string]interface{} `json:"params" bson:"params"`
}

// OptionData is the wrapped within SurveyData
type OptionData struct {
	Title    string      `json:"title" bson:"title"`
	Value    interface{} `json:"value" bson:"value"`
	Score    *float64    `json:"score" bson:"score"`
	Selected bool        `json:"selected" bson:"selected"`
}

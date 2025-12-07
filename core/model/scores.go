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
	// TypeScore is a message type for score
	TypeScore logutils.MessageDataType = "score"
	// TypeRank is a message type for rank
	TypeRank logutils.MessageDataType = "rank"
	// SurveyTypeFashionQuiz describes type of survey for fashion quizzes
	SurveyTypeFashionQuiz string = "fashion_quiz"
	// ScoreStreakMultiplier multiplies score if streak is true
	ScoreStreakMultiplier float64 = 2.0
	// ScoreStreakMinDays specifies minimum number of days for a streak
	ScoreStreakMinDays uint32 = 2
)

// Score object maintains data of accumulated scores from surveys
type Score struct {
	ID                     string       `json:"id" bson:"_id"`
	OrgID                  string       `json:"org_id" bson:"org_id"`
	AppID                  string       `json:"app_id" bson:"app_id"`
	UserID                 string       `json:"user_id" bson:"user_id"`
	SurveyType             string       `json:"survey_type" bson:"survey_type"`
	ExternalProfileID      string       `json:"external_profile_id" bson:"external_profile_id"`     // Mastodon ID (legacy)
	ExternalUserID         string       `json:"external_user_id,omitempty" bson:"external_user_id"` // AmgUUID (new)
	Score                  float64      `json:"score" bson:"score"`
	ResponseCount          uint32       `json:"response_count" bson:"response_count"`
	PrevSurveyResponseDate time.Time    `json:"prev_survey_response_date" bson:"prev_survey_response_date"`
	CurrentStreak          uint32       `json:"current_streak" bson:"current_streak"`
	StreakMultiplier       float64      `json:"streak_multiplier" bson:"streak_multiplier"`
	AnswerCount            uint32       `json:"answer_count" bson:"answer_count"`
	CorrectAnswerCount     uint32       `json:"correct_answer_count" bson:"correct_answer_count"`
	Rank                   uint32       `json:"rank" bson:"rank"`
	Leaderboard            *Leaderboard `json:"leaderboard" bson:"leaderboard,omitempty"`
}

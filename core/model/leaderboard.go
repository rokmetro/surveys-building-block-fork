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
	// TypeLeaderboard is a message type for leaderboard
	TypeLeaderboard logutils.MessageDataType = "leaderboard"

	// TypeLeaderboardEntry is a message type for leaderboard entry
	TypeLeaderboardEntry logutils.MessageDataType = "leaderboard entry"
)

// Leaderboard represents a custom leaderboard.
type Leaderboard struct {
	ID      string `json:"id" bson:"_id"`            // corresponds to "id"
	OrgID   string `json:"org_id" bson:"org_id"`     // organization ID
	AppID   string `json:"app_id" bson:"app_id"`     // application ID
	Name    string `json:"name" bson:"name"`         // required "name"
	IsAdmin bool   `json:"is_admin" bson:"is_admin"` // indicates if the user is an admin of the leaderboard

	LastQuizTime *time.Time `json:"last_quiz_time" bson:"last_quiz_time"`

	DateCreated time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated *time.Time `json:"date_updated" bson:"date_updated"`
}

// LeaderboardEntry represents an entry in a leaderboard.
type LeaderboardEntry struct {
	ID            string     `json:"id" bson:"_id"`
	LeaderboardID string     `json:"leaderboard_id" bson:"leaderboard_id"`
	OrgID         string     `json:"org_id" bson:"org_id"`
	AppID         string     `json:"app_id" bson:"app_id"`
	UserID        string     `json:"user_id" bson:"user_id"`
	IsAdmin       bool       `json:"is_admin" bson:"is_admin"`
	DateCreated   time.Time  `json:"date_created" bson:"date_created"`
	DateUpdated   *time.Time `json:"date_updated" bson:"date_updated"`
}

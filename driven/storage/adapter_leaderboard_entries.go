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

package storage

import (
	"application/core/model"

	"github.com/rokwire/rokwire-building-block-sdk-go/utils/errors"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetLeaderboardEntry retrieves a leaderboard entry
func (a *Adapter) GetLeaderboardEntry(leaderboardID string, orgID string, appID string, userID string) (*model.LeaderboardEntry, error) {
	filter := bson.M{
		"leaderboard_id": leaderboardID,
		"org_id":         orgID,
		"app_id":         appID,
		"user_id":        userID,
	}

	var entry model.LeaderboardEntry
	err := a.db.leaderboardEntries.FindOne(a.context, filter, &entry, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeSurvey, filterArgs(filter), err)
	}

	return &entry, nil
}

// CreateLeaderboardEntry creates a new leaderboardEntry object
func (a *Adapter) CreateLeaderboardEntry(leaderboardEntry model.LeaderboardEntry) error {
	_, err := a.db.leaderboardEntries.InsertOne(a.context, leaderboardEntry)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionCreate, model.TypeLeaderboardEntry, nil, err)
	}
	return nil
}

// DeleteLeaderboardEntries deletes a leaderboard entries according to users in leavingUserIDs
func (a *Adapter) DeleteLeaderboardEntries(leaderboardID string, orgID string, appID string, leavingUserIDs []string) error {
	filter := bson.D{
		primitive.E{Key: "leaderboard_id", Value: leaderboardID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "user_id", Value: bson.M{"$in": leavingUserIDs}},
	}

	result, err := a.db.leaderboardEntries.DeleteMany(a.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeLeaderboardEntry, &logutils.FieldArgs{"user_ids": leavingUserIDs}, err)
	}
	if result.DeletedCount == 0 {
		return errors.ErrorData(logutils.StatusMissing, model.TypeLeaderboardEntry, &logutils.FieldArgs{"user_ids": leavingUserIDs})
	}

	return nil
}

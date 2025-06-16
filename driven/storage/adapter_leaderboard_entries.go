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
	"go.mongodb.org/mongo-driver/mongo"
)

// CreateLeaderboardEntry creates a new leaderboardEntry object
func (a *Adapter) CreateLeaderboardEntry(leaderboardEntry model.LeaderboardEntry) error {
	_, err := a.db.leaderboardEntries.InsertOne(a.context, leaderboardEntry)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionCreate, model.TypeLeaderboardEntry, nil, err)
	}
	return nil
}

func (a *Adapter) DeleteLeaderboardEntry(id string, orgID string, appID string, userID string) error {
	filter := bson.M{
		"_id":            id,
		"org_id":        orgID,
		"app_id":        appID,
		"user_id":       userID,
	}

	result, err := a.db.leaderboardEntries.DeleteOne(a.context, filter)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeLeaderboardEntry, &logutils.FieldArgs{"id": id}, err)
	}
	if result.DeletedCount == 0 {
		return errors.ErrorData(logutils.StatusMissing, model.TypeLeaderboardEntry, &logutils.FieldArgs{"id": id})
	}
	return nil
}

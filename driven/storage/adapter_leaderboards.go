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
	"application/core/interfaces"
	"application/core/model"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/errors"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// GetLeaderboards gets all leaderboards for a user
func (a *Adapter) GetLeaderboards(orgID string, appID string, userID string) ([]model.Leaderboard, error) {
	leaderboardEntryFilter := bson.M{
		"org_id":  orgID,
		"app_id":  appID,
		"user_id": userID,
	}

	pipeline := mongo.Pipeline{
		// 1) filter leadboard entries for this user
		bson.D{{Key: "$match", Value: leaderboardEntryFilter}},
		// 2) join each entry to its leaderboard
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "leaderboards"},
			{Key: "localField", Value: "leaderboard_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "leaderboard"},
		}}},
		// 3) unwind the resulting array so we get a single doc per leaderboard
		bson.D{{Key: "$unwind", Value: "$leaderboard"}},
		// 4) replace the root with a merged object:
		//   - all fields from the leaderboard
		//   - plus an "is_admin" field from the entry
		bson.D{{Key: "$replaceRoot", Value: bson.D{
			{Key: "newRoot", Value: bson.D{
				{Key: "$mergeObjects", Value: bson.A{
					"$leaderboard",
					bson.D{{Key: "is_admin", Value: "$is_admin"}},
				}},
			}},
		}}},
	}

	var leaderboards []model.Leaderboard
	err := a.db.leaderboards.Aggregate(a.context, pipeline, &leaderboards, nil)

	return leaderboards, err
}

// CreateLeaderboardAndEntry creates a new leaderboard and corresponding leaderboard entry
func (a *Adapter) CreateLeaderboardAndEntry(leaderboard model.Leaderboard, leaderboardEntry model.LeaderboardEntry) (*model.Leaderboard, error) {
	transaction := func(storage interfaces.Storage) error {
		//1. Create leaderboard
		_, err := a.db.leaderboards.InsertOne(a.context, leaderboard)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionCreate, model.TypeLeaderboard, nil, err)
		}

		//2. Create corresponding leaderboard entry
		_, err = a.db.leaderboardEntries.InsertOne(a.context, leaderboardEntry)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionCreate, model.TypeLeaderboardEntry, nil, err)
		}

		return nil
	}

	err := a.PerformTransaction(transaction)
	if err != nil {
		return nil, err
	}
	return &leaderboard, nil
}

// UpdateLeaderboard updates an existing leaderboard
func (a *Adapter) UpdateLeaderboard(leaderboard model.Leaderboard) error {
	filter := bson.M{
		"_id":    leaderboard.ID,
		"org_id": leaderboard.OrgID,
		"app_id": leaderboard.AppID,
	}
	update := bson.M{
		"$set": bson.M{
			"name": leaderboard.Name,
		},
	}

	_, err := a.db.leaderboards.UpdateOne(a.context, filter, update, nil)
	return err
}

// DeleteLeaderboard deletes a leaderboard by ID along with corresponding leaderboard entries
func (a *Adapter) DeleteLeaderboard(leaderboardID, orgID, appID, userID string) error {
	transaction := func(storage interfaces.Storage) error {
		//1. Delete leaderboard
		filter := bson.M{
			"_id":    leaderboardID,
			"org_id": orgID,
			"app_id": appID,
		}
		_, err := a.db.leaderboards.DeleteOne(a.context, filter, nil)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionDelete, model.TypeLeaderboard, filterArgs(filter), err)
		}

		//2. Delete all leaderboard entries corresponding to leaderboardID
		filter = bson.M{
			"leaderboard_id": leaderboardID,
			"org_id":         orgID,
			"app_id":         appID,
		}
		_, err = a.db.leaderboardEntries.DeleteMany(a.context, filter, nil)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionDelete, model.TypeLeaderboardEntry, filterArgs(filter), err)
		}

		return nil
	}

	return a.PerformTransaction(transaction)
}

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

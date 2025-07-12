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
	"go.mongodb.org/mongo-driver/mongo"
)

// GetLeaderboard gets a leaderboard by ID
func (a *Adapter) GetLeaderboard(leaderboardID string, orgID string, appID string) (*model.Leaderboard, error) {
	filter := bson.M{
		"_id":    leaderboardID,
		"org_id": orgID,
		"app_id": appID,
	}

	var leaderboard model.Leaderboard
	err := a.db.leaderboards.FindOne(a.context, filter, &leaderboard, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeLeaderboard, filterArgs(filter), err)
	}

	return &leaderboard, nil
}

// GetLeaderboardWithUserContext gets a leaderboard by ID and populates the is_admin field for a specific user
func (a *Adapter) GetLeaderboardWithUserContext(leaderboardID string, orgID string, appID string, userID string) (*model.Leaderboard, error) {
	leaderboardFilter := bson.M{
		"_id":    leaderboardID,
		"org_id": orgID,
		"app_id": appID,
	}

	pipeline := mongo.Pipeline{
		// 1) filter leaderboards by ID, org, and app
		bson.D{{Key: "$match", Value: leaderboardFilter}},
		// 2) lookup the leaderboard entry for this specific user
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "leaderboard_entries"},
			{Key: "let", Value: bson.D{{Key: "leaderboard_id", Value: "$_id"}}},
			{Key: "pipeline", Value: mongo.Pipeline{
				bson.D{{Key: "$match", Value: bson.D{
					{Key: "$expr", Value: bson.D{
						{Key: "$and", Value: bson.A{
							bson.D{{Key: "$eq", Value: bson.A{"$leaderboard_id", "$$leaderboard_id"}}},
							bson.D{{Key: "$eq", Value: bson.A{"$user_id", userID}}},
							bson.D{{Key: "$eq", Value: bson.A{"$org_id", orgID}}},
							bson.D{{Key: "$eq", Value: bson.A{"$app_id", appID}}},
						}},
					}},
				}}},
			}},
			{Key: "as", Value: "user_entry"},
		}}},
		// 3) add the is_admin field from the user's entry (if it exists)
		bson.D{{Key: "$addFields", Value: bson.D{
			{Key: "is_admin", Value: bson.D{
				{Key: "$cond", Value: bson.D{
					{Key: "if", Value: bson.D{{Key: "$gt", Value: bson.A{bson.D{{Key: "$size", Value: "$user_entry"}}, 0}}}},
					{Key: "then", Value: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$user_entry.is_admin", 0}}}},
					{Key: "else", Value: nil},
				}},
			}},
		}}},
		// 4) remove the temporary user_entry field
		bson.D{{Key: "$unset", Value: "user_entry"}},
	}

	var leaderboards []model.Leaderboard
	err := a.db.leaderboards.Aggregate(a.context, pipeline, &leaderboards, nil)

	if len(leaderboards) == 0 {
		return nil, errors.ErrorData(logutils.StatusMissing, model.TypeLeaderboard, &logutils.FieldArgs{"leaderboard_id": leaderboardID})
	}

	return &leaderboards[0], err
}

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
		// 2) sort by date_created in descending order (most recent first)
		bson.D{{Key: "$sort", Value: bson.D{{Key: "date_created", Value: -1}}}},
		// 3) join each entry to its leaderboard
		bson.D{{Key: "$lookup", Value: bson.D{
			{Key: "from", Value: "leaderboards"},
			{Key: "localField", Value: "leaderboard_id"},
			{Key: "foreignField", Value: "_id"},
			{Key: "as", Value: "leaderboard"},
		}}},
		// 4) unwind the resulting array so we get a single doc per leaderboard
		bson.D{{Key: "$unwind", Value: "$leaderboard"}},
		// 5) replace the root with a merged object:
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
	err := a.db.leaderboardEntries.Aggregate(a.context, pipeline, &leaderboards, nil)

	return leaderboards, err
}

// CreateLeaderboard creates a new leaderboard
func (a *Adapter) CreateLeaderboard(leaderboard model.Leaderboard) (*model.Leaderboard, error) {
	// Create the leaderboard
	_, err := a.db.leaderboards.InsertOne(a.context, leaderboard)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, model.TypeLeaderboard, nil, err)
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

	setUpdate := bson.M{
		"name":         leaderboard.Name,
		"date_updated": leaderboard.DateUpdated,
	}
	if leaderboard.LastQuizTime != nil {
		setUpdate["last_quiz_time"] = leaderboard.LastQuizTime
	}
	update := bson.M{
		"$set": setUpdate,
	}

	_, err := a.db.leaderboards.UpdateOne(a.context, filter, update, nil)
	return err
}

// DeleteLeaderboard deletes a leaderboard by ID along with corresponding leaderboard entries
func (a *Adapter) DeleteLeaderboard(leaderboardID, orgID, appID, userID string) error {
	filter := bson.M{
		"_id":    leaderboardID,
		"org_id": orgID,
		"app_id": appID,
	}
	_, err := a.db.leaderboards.DeleteOne(a.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeLeaderboard, filterArgs(filter), err)
	}

	return err
}

// GetLeaderboardEntries retrieves a list of leaderboard entry
func (a *Adapter) GetLeaderboardEntries(leaderboardID string, orgID string, appID string, userID *string) ([]model.LeaderboardEntry, error) {
	filter := bson.M{
		"leaderboard_id": leaderboardID,
		"org_id":         orgID,
		"app_id":         appID,
	}

	if userID != nil && *userID != "" {
		filter["user_id"] = *userID
	}

	var results []model.LeaderboardEntry
	err := a.db.leaderboardEntries.Find(a.context, filter, &results, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeSurvey, filterArgs(filter), err)
	}

	return results, nil
}

// CreateLeaderboardEntry creates a new leaderboardEntry object
func (a *Adapter) CreateLeaderboardEntry(leaderboardEntry model.LeaderboardEntry) error {
	_, err := a.db.leaderboardEntries.InsertOne(a.context, leaderboardEntry)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionCreate, model.TypeLeaderboardEntry, nil, err)
	}
	return nil
}

// DeleteLeaderboardEntries deletes non-admin leaderboard entries according to users in leavingUserIDs
func (a *Adapter) DeleteLeaderboardEntries(leaderboardID string, orgID string, appID string, leavingUserIDs []string) error {
	filter := bson.D{
		primitive.E{Key: "leaderboard_id", Value: leaderboardID},
		primitive.E{Key: "org_id", Value: orgID},
		primitive.E{Key: "app_id", Value: appID},
		primitive.E{Key: "user_id", Value: bson.M{"$in": leavingUserIDs}},
		primitive.E{Key: "is_admin", Value: false}, // Only delete non-admin entries
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

// DeleteAllLeaderboardEntries deletes all leaderboard entries for a given leaderboardID
func (a *Adapter) DeleteAllLeaderboardEntries(leaderboardID string, orgID string, appID string) error {
	filter := bson.M{
		"leaderboard_id": leaderboardID,
		"org_id":         orgID,
		"app_id":         appID,
	}
	_, err := a.db.leaderboardEntries.DeleteMany(a.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeLeaderboardEntry, filterArgs(filter), err)
	}

	return err
}

// GetLeaderboardScores retrieves paginated scores for a specific leaderboard
func (a *Adapter) GetLeaderboardScores(leaderboardID string, orgID string, appID string, limit *int, offset *int) ([]model.Score, error) {
	pipeline := mongo.Pipeline{}

	// 1) match only this leaderboard
	leaderboardEntryFilter := bson.D{{Key: "$match", Value: bson.M{
		"leaderboard_id": leaderboardID,
		"org_id":         orgID,
		"app_id":         appID,
	}}}
	pipeline = append(pipeline, leaderboardEntryFilter)

	// 2) lookup into scores by user_id + org/app
	lookup := bson.D{{Key: "$lookup", Value: bson.M{
		"from": "scores",
		"let":  bson.M{"uid": "$user_id"},
		"pipeline": mongo.Pipeline{
			bson.D{{Key: "$match", Value: bson.M{
				"$expr": bson.M{"$and": bson.A{
					bson.M{"$eq": bson.A{"$user_id", "$$uid"}},
					bson.M{"$eq": bson.A{"$org_id", orgID}},
					bson.M{"$eq": bson.A{"$app_id", appID}},
				}},
			}}},
		},
		"as": "scores",
	}}}
	pipeline = append(pipeline, lookup)

	// 3) unwind the single-element scoreDoc array
	unwind := bson.D{{Key: "$unwind", Value: "$scores"}}
	pipeline = append(pipeline, unwind)

	// 4) replace root with the embedded scoreDoc
	replaceRoot := bson.D{{Key: "$replaceRoot", Value: bson.M{
		"newRoot": "$scores",
	}}}
	pipeline = append(pipeline, replaceRoot)

	// 5) add a dense‐rank over score descending
	rankWindow := bson.D{{Key: "$setWindowFields", Value: bson.M{
		"sortBy": bson.M{"score": -1},
		"output": bson.M{
			"rank": bson.M{"$denseRank": bson.M{}},
		},
	}}}
	pipeline = append(pipeline, rankWindow)

	if offset != nil {
		pipeline = append(pipeline, bson.D{{Key: "$skip", Value: *offset}})
	}

	if limit != nil {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: *limit}})
	}

	var scores []model.Score
	err := a.db.leaderboardEntries.Aggregate(a.context, pipeline, &scores, nil)

	return scores, err
}

// GetLeaderboardUserScores retrieves scores for a specific user in each leaderboard they're in
func (a *Adapter) GetLeaderboardUserScores(orgID string, appID string, userID string, limit *int, offset *int) ([]model.Score, error) {
	pipeline := mongo.Pipeline{}

	leaderboardEntryFilter := bson.D{{Key: "$match", Value: bson.M{
		"org_id": orgID,
		"app_id": appID,
	}}}
	pipeline = append(pipeline, leaderboardEntryFilter)

	lookup := bson.D{{Key: "$lookup", Value: bson.M{
		"from": "scores",
		"let":  bson.M{"uid": "$user_id"},
		"pipeline": mongo.Pipeline{
			// match the same user/org/app
			{{Key: "$match", Value: bson.M{
				"$expr": bson.M{"$and": bson.A{
					bson.M{"$eq": bson.A{"$user_id", "$$uid"}},
					bson.M{"$eq": bson.A{"$org_id", orgID}},
					bson.M{"$eq": bson.A{"$app_id", appID}},
				}},
			}}},
		},
		"as": "scores",
	}}}
	pipeline = append(pipeline, lookup)

	unwind := bson.D{{Key: "$unwind", Value: "$scores"}}
	pipeline = append(pipeline, unwind)

	replaceRoot := bson.D{{Key: "$replaceRoot", Value: bson.M{
		"newRoot": bson.M{"$mergeObjects": bson.A{"$scores", "$$ROOT"}},
	}}}
	pipeline = append(pipeline, replaceRoot)

	// lookup leaderboards for each entry
	addLeaderboardField := bson.D{{
		Key: "$lookup", Value: bson.M{
			"from":         "leaderboards",
			"localField":   "leaderboard_id",
			"foreignField": "_id",
			"as":           "leaderboard",
		},
	}}
	pipeline = append(pipeline, addLeaderboardField)

	unwindLeaderboard := bson.D{{
		Key: "$unwind", Value: bson.M{
			"path": "$leaderboard",
		},
	}}
	pipeline = append(pipeline, unwindLeaderboard)

	rankWindow := bson.D{{Key: "$setWindowFields", Value: bson.M{
		"partitionBy": "$leaderboard_id",
		"sortBy":      bson.M{"scores.score": -1},
		"output": bson.M{
			"rank": bson.M{"$denseRank": bson.M{}},
		},
	}}}
	pipeline = append(pipeline, rankWindow)

	userIDFilter := bson.D{{Key: "$match", Value: bson.M{"user_id": userID}}}
	pipeline = append(pipeline, userIDFilter)

	sortByRank := bson.D{{Key: "$sort", Value: bson.M{"rank": 1}}}
	pipeline = append(pipeline, sortByRank)

	if offset != nil {
		pipeline = append(pipeline, bson.D{{Key: "$skip", Value: *offset}})
	}

	if limit != nil {
		pipeline = append(pipeline, bson.D{{Key: "$limit", Value: *limit}})
	}

	var scores []model.Score
	err := a.db.leaderboardEntries.Aggregate(a.context, pipeline, &scores, nil)

	return scores, err
}

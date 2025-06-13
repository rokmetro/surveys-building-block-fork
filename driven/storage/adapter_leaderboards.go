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
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetLeaderboardsForUser gets all leaderboards for a user
func (a *Adapter) GetLeaderboardsForUser(userID, orgID, appID string) ([]model.Leaderboard, error) {
	filter := bson.M{
		"$and": []bson.M{
			{"org_id": orgID},
			{"app_id": appID},
			{"$or": []bson.M{
				{"admin_user_ids": userID},
				{"user_ids": userID},
			}},
		},
	}

	var leaderboards []model.Leaderboard
	err := a.db.leaderboards.Find(context.Background(), filter, &leaderboards, nil)
	if err != nil {
		return nil, err
	}

	return leaderboards, nil
}

// CreateLeaderboard creates a new leaderboard
func (a *Adapter) CreateLeaderboard(lb model.Leaderboard) (*model.Leaderboard, error) {
	if len(lb.ID) == 0 {
		lb.ID = primitive.NewObjectID().Hex()
	}

	_, err := a.db.leaderboards.InsertOne(context.Background(), lb)
	if err != nil {
		return nil, err
	}

	return &lb, nil
}

// UpdateLeaderboard updates an existing leaderboard
func (a *Adapter) UpdateLeaderboard(lb model.Leaderboard) error {
	filter := bson.M{
		"_id":    lb.ID,
		"org_id": lb.OrgID,
		"app_id": lb.AppID,
	}
	update := bson.M{
		"$set": bson.M{
			"name":           lb.Name,
			"admin_user_ids": lb.AdminUserIDs,
			"user_ids":       lb.UserIDs,
		},
	}

	_, err := a.db.leaderboards.UpdateOne(context.Background(), filter, update, nil)
	return err
}

// DeleteLeaderboard deletes a leaderboard by ID, scoped to org_id and app_id
func (a *Adapter) DeleteLeaderboard(id, orgID, appID string) error {
	filter := bson.M{
		"_id":    id,
		"org_id": orgID,
		"app_id": appID,
	}
	_, err := a.db.leaderboards.DeleteOne(context.Background(), filter, nil)
	return err
}

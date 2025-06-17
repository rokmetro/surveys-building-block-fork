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
	"application/driven/storage"
	"context"

	"github.com/rokwire/rokwire-building-block-sdk-go/utils/errors"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// // validateLeaderboard performs basic validation on a leaderboard
// func (a *Adapter) validateLeaderboard(lb model.Leaderboard) error {
// 	if lb.Name == "" {
// 		return errors.ErrorData(logutils.StatusMissing, "name", nil)
// 	}
// 	if len(lb.AdminUserIDs) == 0 {
// 		return errors.ErrorData(logutils.StatusMissing, "admin_user_ids", nil)
// 	}
// 	if len(lb.UserIDs) == 0 {
// 		return errors.ErrorData(logutils.StatusMissing, "user_ids", nil)
// 	}
// 	return nil
// }

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

	// Ensure the creator is in the admin list
	if len(lb.AdminUserIDs) == 0 {
		// If no admins specified, use the first user ID as admin
		if len(lb.UserIDs) > 0 {
			lb.AdminUserIDs = []string{lb.UserIDs[0]}
		} else {
			return nil, errors.ErrorData(logutils.StatusMissing, "admin_user_ids", nil)
		}
	}

	// Validate the leaderboard
	if err := a.validateLeaderboard(lb); err != nil {
		return nil, err
	}

	_, err := a.db.leaderboards.InsertOne(context.Background(), lb)
	if err != nil {
		return nil, err
	}

	return &lb, nil
}

// UpdateLeaderboard updates an existing leaderboard
func (a *Adapter) UpdateLeaderboard(lb model.Leaderboard, userID string) error {
	// First get the current leaderboard to verify admin status and preserve admin list
	var currentLb model.Leaderboard
	err := a.db.leaderboards.FindOne(context.Background(), bson.M{
		"_id":    lb.ID,
		"org_id": lb.OrgID,
		"app_id": lb.AppID,
	}, &currentLb, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionFind, model.TypeLeaderboard, nil, err)
	}

	// Verify the requesting user is an admin
	isAdmin := false
	for _, adminID := range currentLb.AdminUserIDs {
		if adminID == userID {
			isAdmin = true
			break
		}
	}
	if !isAdmin {
		return errors.ErrorData(logutils.StatusInvalid, "user", &logutils.FieldArgs{"admin": false})
	}

	// Preserve the admin list from the current leaderboard
	lb.AdminUserIDs = currentLb.AdminUserIDs

	// Validate the updated leaderboard
	if err := a.validateLeaderboard(lb); err != nil {
		return err
	}

	// Only allow updating name and user IDs, preserve admin list
	filter := bson.M{
		"_id":    lb.ID,
		"org_id": lb.OrgID,
		"app_id": lb.AppID,
	}
	update := bson.M{
		"$set": bson.M{
			"name":     lb.Name,
			"user_ids": lb.UserIDs,
		},
	}

	_, err = a.db.leaderboards.UpdateOne(context.Background(), filter, update, nil)
	return err
}

// DeleteLeaderboard deletes a leaderboard by ID alnog with corresponding leaderboard entries
func (a *Adapter) DeleteLeaderboard(leaderboardID, orgID, appID, userID string) error {
	transaction := func(storage interfaces.Storage) error {
		//1. Delete leaderboard
		filter := bson.M{
			"_id":            leaderboardID,
			"org_id":         orgID,
			"app_id":         appID,
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

	return storage.PerformTransaction(transaction)
}

func (a *Adapter) JoinLeaderboard(id string, orgID string, appID string, userID string) error {

}

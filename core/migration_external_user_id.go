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

package core

import (
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
)

// MigrateScoreExternalUserIDs migrates all score records to populate external_user_id with AmgUUID from Core BB
func (app *Application) MigrateScoreExternalUserIDs(orgID string, appID string, batchSize int) error {
	log.Println("Starting Score External User ID migration...")
	log.Printf("Parameters: orgID=%s, appID=%s, batchSize=%d", orgID, appID, batchSize)

	// Validate batch size
	if batchSize <= 0 {
		batchSize = 100 // Default to 100 if invalid
		log.Printf("Invalid batch size, using default: %d", batchSize)
	}

	// Get all scores that don't have external_user_id populated
	filter := bson.M{
		"org_id": orgID,
		"app_id": appID,
		"$or": []bson.M{
			{"external_user_id": bson.M{"$exists": false}},
			{"external_user_id": ""},
		},
	}

	scores, err := app.storage.FindScores(filter, nil, nil)
	if err != nil {
		return fmt.Errorf("error fetching scores: %v", err)
	}

	total := len(scores)
	log.Printf("Found %d scores to migrate", total)

	if total == 0 {
		log.Println("No scores to migrate")
		return nil
	}

	successCount := 0
	errorCount := 0
	skippedCount := 0

	// Process in batches to avoid overwhelming the Core BB API
	for i := 0; i < total; i += batchSize {
		end := i + batchSize
		if end > total {
			end = total
		}

		batch := scores[i:end]
		log.Printf("Processing batch %d-%d of %d", i+1, end, total)

		for _, score := range batch {
			// Skip if already has external_user_id
			if score.ExternalUserID != "" {
				skippedCount++
				continue
			}

			// Skip if no external_profile_id (Mastodon ID)
			if score.ExternalProfileID == "" {
				log.Printf("Score %s has no external_profile_id, skipping", score.ID)
				skippedCount++
				continue
			}

			// Look up the AmgUUID from Core BB using the Mastodon ID
			// Core BB uses identifiers array with code/identifier pairs
			accountCriteria := map[string]interface{}{
				"identifiers": map[string]interface{}{
					"$elemMatch": map[string]interface{}{
						"code":       "mastodon_id",
						"identifier": score.ExternalProfileID,
					},
				},
			}

			coreAccounts, err := app.corebb.RetrieveCoreUserAccountByCriteria(accountCriteria, &appID, &orgID)
			if err != nil {
				log.Printf("Error retrieving Core account for score %s (mastodon_id: %s): %v",
					score.ID, score.ExternalProfileID, err)
				errorCount++
				continue
			}

			if len(coreAccounts) == 0 {
				log.Printf("No Core account found for score %s (mastodon_id: %s)",
					score.ID, score.ExternalProfileID)
				errorCount++
				continue
			}

			if len(coreAccounts) > 1 {
				log.Printf("WARNING: Multiple Core accounts found for mastodon_id %s, using first one",
					score.ExternalProfileID)
			}

			// Get the AmgUUID from the first matching account
			amgUUID := coreAccounts[0].ID
			if amgUUID == "" {
				log.Printf("Core account has empty ID for score %s", score.ID)
				errorCount++
				continue
			}

			// Update the score with the AmgUUID
			updateFilter := bson.M{"_id": score.ID}
			update := bson.M{
				"$set": bson.M{
					"external_user_id": amgUUID,
				},
			}

			err = app.storage.UpdateScoreByFilter(updateFilter, update)
			if err != nil {
				log.Printf("Error updating score %s: %v", score.ID, err)
				errorCount++
				continue
			}

			log.Printf("✓ Score %s: external_profile_id=%s → external_user_id=%s",
				score.ID, score.ExternalProfileID, amgUUID)
			successCount++
		}

		// // Sleep between batches to avoid rate limiting
		// if end < total {
		// 	log.Println("Sleeping 2 seconds before next batch...")
		// 	time.Sleep(2 * time.Second)
		// }
	}

	log.Println("Migration complete!")
	log.Printf("Summary - Total: %d, Success: %d, Errors: %d, Skipped: %d",
		total, successCount, errorCount, skippedCount)

	// Return error if we had too many failures
	if errorCount > 0 && successCount == 0 {
		return fmt.Errorf("migration failed: no scores were successfully migrated (%d errors)", errorCount)
	}

	if errorCount > total/2 {
		log.Printf("WARNING: High error rate - %d/%d scores failed to migrate", errorCount, total)
	}

	return nil
}

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
	scores, err := app.storage.FindScoresNoRanks(orgID, appID, nil, true, nil, nil)
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

		// Collect unique account IDs from scores in this batch
		accountIDs := []string{}
		for _, score := range batch {
			if score.ExternalUserID != "" {
				skippedCount++
				continue
			}
			if score.UserID == "" {
				log.Printf("Score %s has no UserID (Core BB account ID), skipping", score.ID)
				skippedCount++
				continue
			}
			accountIDs = append(accountIDs, score.UserID)
		}

		if len(accountIDs) == 0 {
			log.Printf("No valid account IDs in this batch, skipping")
			continue
		}

		log.Printf("Fetching Core BB accounts for %d account IDs", len(accountIDs))

		// Look up all AmgUUIDs from Core BB in one call
		coreAccounts, err := app.corebb.RetrieveCoreUserAccountByCriteria(accountIDs, &appID, &orgID)
		if err != nil {
			log.Printf("Error retrieving Core accounts for batch: %v", err)
			errorCount += len(accountIDs)
			continue
		}

		log.Printf("Retrieved %d Core BB accounts", len(coreAccounts))

		// Build a map of account ID -> amg_uuid for quick lookup
		accountIDToAmgUUID := make(map[string]string)
		for _, account := range coreAccounts {
			amgUUID := account.GetAmgUUID()
			if account.ID != "" && amgUUID != "" {
				accountIDToAmgUUID[account.ID] = amgUUID
			}
		}

		// Update scores with their corresponding AmgUUIDs
		for _, score := range batch {
			if score.ExternalUserID != "" || score.UserID == "" {
				continue
			}

			amgUUID, found := accountIDToAmgUUID[score.UserID]
			if !found {
				log.Printf("No Core account found for account ID: %s", score.UserID)
				errorCount++
				continue
			}

			err = app.storage.UpdateScoreExternalUserID(score.ID, amgUUID)
			if err != nil {
				log.Printf("Error updating score %s: %v", score.ID, err)
				errorCount++
				continue
			}

			log.Printf("✓ Score %s: UserID=%s → external_user_id=%s",
				score.ID, score.UserID, amgUUID)
			successCount++
		}
	}

	log.Println("Migration complete!")
	log.Printf("Summary - Total: %d, Success: %d, Errors: %d, Skipped: %d",
		total, successCount, errorCount, skippedCount)

	// Return error if we had too many failures
	if errorCount > 0 && successCount == 0 {
		return fmt.Errorf("migration failed: no scores were successfully migrated (%d errors)", errorCount)
	}

	return nil
}

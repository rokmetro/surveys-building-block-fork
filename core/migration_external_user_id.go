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

const (
	// DefaultExternalUserIDMigrationBatchSize is the default batch size used when migrating
	// external user IDs for score records. This constant is used when an invalid batch size
	// is provided to the migration function.
	DefaultExternalUserIDMigrationBatchSize int = 100
)

// MigrateScoreExternalUserIDs migrates all score records to populate external_user_id with AmgUUID from Core BB
func (app *Application) MigrateScoreExternalUserIDs(orgID string, appID string, batchSize int) error {
	log.Println("Starting Score External User ID migration...")
	log.Printf("Parameters: orgID=%s, appID=%s, batchSize=%d", orgID, appID, batchSize)

	// Validate batch size
	if batchSize <= 0 {
		batchSize = DefaultExternalUserIDMigrationBatchSize
		log.Printf("Invalid batch size, using default: %d", batchSize)
	}

	successCount := 0
	errorCount := 0
	skippedCount := 0
	total := 0
	offset := 0

	// Process in batches to avoid overwhelming the Core BB API and memory
	for {
		// Fetch a batch of scores using limit and offset
		batch, err := app.storage.FindScoresWithoutExternalUserID(orgID, appID, &batchSize, &offset)
		if err != nil {
			return fmt.Errorf("error fetching scores: %v", err)
		}

		// If no more scores, we're done
		if len(batch) == 0 {
			break
		}

		total += len(batch)
		log.Printf("Processing batch %d-%d (fetched %d scores)", offset+1, offset+len(batch), len(batch))

		// Collect unique account IDs from scores in this batch
		accountIDs := []string{}
		for _, score := range batch {
			if score.UserID == "" {
				log.Printf("Score %s has no UserID (Core BB account ID), skipping", score.ID)
				skippedCount++
				continue
			}
			accountIDs = append(accountIDs, score.UserID)
		}

		if len(accountIDs) == 0 {
			log.Printf("No valid account IDs in this batch, skipping")
		} else {
			log.Printf("Fetching Core BB accounts for %d account IDs", len(accountIDs))

			// Look up all AmgUUIDs from Core BB in one call
			coreAccounts, err := app.corebb.RetrieveCoreUserAccountByCriteria(accountIDs, &appID, &orgID)
			if err != nil {
				log.Printf("Error retrieving Core accounts for batch: %v", err)
				errorCount += len(accountIDs)
			} else {
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
					if score.UserID == "" {
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
		}

		// Move to next batch
		offset += batchSize

		// If we got fewer scores than the batch size, we've reached the end
		if len(batch) < batchSize {
			break
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

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
	"application/core/model"
	"application/driven/corebb"
	"log"
)

// SyncAmgUUIDForScore fetches the AmgUUID from Core BB and populates the external_user_id field
func (app *Application) SyncAmgUUIDForScore(score *model.Score, externalIDs map[string]string) error {
	// Get Mastodon ID from externalIDs if provided, otherwise use the one already in score
	mastodonID := score.ExternalProfileID
	if externalIDs != nil {
		if mastodonIDFromMap, ok := externalIDs["mastodon_id"]; ok && mastodonIDFromMap != "" {
			mastodonID = mastodonIDFromMap
		}
	}

	if mastodonID == "" {
		log.Printf("SyncAmgUUIDForScore: no mastodon_id provided for score, skipping AmgUUID sync")
		return nil
	}

	// Look up the AmgUUID from Core BB
	accountCriteria := corebb.BuildCoreAccountCriteriaByMastodonID(mastodonID)

	coreAccounts, err := app.corebb.RetrieveCoreUserAccountByCriteria(accountCriteria, &score.AppID, &score.OrgID)
	if err != nil {
		log.Printf("SyncAmgUUIDForScore: error retrieving Core account: %v", err)
		// Don't fail the score creation if Core BB lookup fails, just log it
		return nil
	}

	if len(coreAccounts) == 0 {
		log.Printf("SyncAmgUUIDForScore: no Core account found for mastodon_id %s", mastodonID)
		return nil
	}

	if len(coreAccounts) > 1 {
		log.Printf("SyncAmgUUIDForScore: WARNING - multiple Core accounts found for mastodon_id %s, using first one", mastodonID)
	}

	// Get the AmgUUID from the identifiers
	amgUUID := coreAccounts[0].GetAmgUUID()
	if amgUUID == "" {
		log.Printf("SyncAmgUUIDForScore: no amg_uuid identifier found for mastodon_id %s", mastodonID)
		return nil
	}

	// Set the AmgUUID on the score
	score.ExternalUserID = amgUUID
	log.Printf("SyncAmgUUIDForScore: synced AmgUUID %s for mastodon_id %s", score.ExternalUserID, mastodonID)

	return nil
}

// PrepareScoresForResponse prepares multiple scores for API response based on feature flag
func (app *Application) PrepareScoresForResponse(scores []model.Score) {
	if !app.shouldUseExternalUserID() {
		return
	}

	for i := range scores {
		if scores[i].ExternalUserID != "" {
			// When flag is enabled and we have an AmgUUID, use it in the external_profile_id field
			scores[i].ExternalProfileID = scores[i].ExternalUserID
		}
	}
}

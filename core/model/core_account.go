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

package model

// CoreAccount represents an account from the Core BB
// This matches the actual Core BB Account structure
type CoreAccount struct {
	ID          string                  `json:"id"`          // This is the AmgUUID
	Identifiers []CoreAccountIdentifier `json:"identifiers"` // Contains all identifiers including mastodon_id
	Profile     CoreProfile             `json:"profile"`     // Basic profile info (optional)
}

// CoreAccountIdentifier represents an identifier in a Core BB account
type CoreAccountIdentifier struct {
	ID         string `json:"id"`
	Code       string `json:"code"`       // e.g., "mastodon_id", "email", etc.
	Identifier string `json:"identifier"` // The actual identifier value
	Verified   bool   `json:"verified"`
}

// CoreProfile represents basic profile info from Core BB
type CoreProfile struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

// GetExternalID returns the identifier value for a given code (e.g., "mastodon_id")
func (ca CoreAccount) GetExternalID(code string) string {
	for _, identifier := range ca.Identifiers {
		if identifier.Code == code {
			return identifier.Identifier
		}
	}
	return ""
}

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

// NotificationMessage wrapper for Notifications BB message
type NotificationMessage struct {
	OrgID string `json:"org_id"`
	AppID string `json:"app_id"`

	Priority int               `json:"priority"`
	Subject  string            `json:"subject"`
	Body     string            `json:"body"`
	Data     map[string]string `json:"data"`

	//recipients related
	Recipients               []NotificationMessageRecipient  `json:"recipients"`
	RecipientsCriteriaList   []NotificationRecipientCriteria `json:"recipients_criteria_list"`
	RecipientAccountCriteria map[string]interface{}          `json:"recipient_account_criteria"`
	Topic                    *string                         `json:"topic"`
}

// GetLoggingParams returns logging parameters for the notification message
func (nm *NotificationMessage) GetLoggingParams() map[string]any {
	recipients := make([]string, len(nm.Recipients))
	for i, r := range nm.Recipients {
		recipients[i] = r.UserID
	}
	topic := ""
	if nm.Topic != nil {
		topic = *nm.Topic
	}

	return map[string]any{
		"recipients":                 recipients,
		"recipient_account_criteria": nm.RecipientAccountCriteria,
		"topic":                      topic,
	}
}

// NotificationMessageRecipient represents a recipient of a Notifications BB message
type NotificationMessageRecipient struct {
	UserID string `json:"user_id"`
	Mute   bool   `json:"mute"`
}

// NotificationRecipientCriteria represents criteria for recipients of a Notifications BB message
type NotificationRecipientCriteria struct {
	AppVersion  *string `json:"app_version"`
	AppPlatform *string `json:"app_platform"`
}

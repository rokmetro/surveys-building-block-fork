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
	"time"

	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logutils"
	"go.mongodb.org/mongo-driver/bson"
)

// GetAlertContacts retrieves all alert contacts
func (a *Adapter) GetAlertContacts(orgID string, appID string) ([]model.AlertContact, error) {
	filter := bson.M{"org_id": orgID, "app_id": appID}
	var entry []model.AlertContact
	err := a.db.alertContacts.Find(a.context, filter, &entry, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeAlertContact, filterArgs(filter), err)
	}
	return entry, nil
}

// GetAlertContact retrieves a single alert contact
func (a *Adapter) GetAlertContact(id string, orgID string, appID string) (*model.AlertContact, error) {
	filter := bson.M{"_id": id, "org_id": orgID, "app_id": appID}
	var entry model.AlertContact
	err := a.db.alertContacts.FindOne(a.context, filter, &entry, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeAlertContact, filterArgs(filter), err)
	}

	return &entry, nil
}

// GetAlertContactsByKey gets all alert contacts that share the key in the filter
func (a *Adapter) GetAlertContactsByKey(key string, orgID string, appID string) ([]model.AlertContact, error) {
	filter := bson.M{"key": key, "org_id": orgID, "app_id": appID}
	var results []model.AlertContact
	err := a.db.alertContacts.Find(a.context, filter, &results, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeAlertContact, filterArgs(filter), err)
	}
	if len(results) == 0 {
		return nil, errors.WrapErrorData(logutils.StatusMissing, model.TypeAlertContact, filterArgs(filter), err)
	}
	return results, nil
}

// CreateAlertContact creates an alert contact
func (a *Adapter) CreateAlertContact(alertContact model.AlertContact) (*model.AlertContact, error) {
	_, err := a.db.alertContacts.InsertOne(a.context, alertContact)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, model.TypeAlertContact, nil, err)
	}

	return &alertContact, nil
}

// UpdateAlertContact updates an alert contact
func (a *Adapter) UpdateAlertContact(alertContact model.AlertContact) error {
	now := time.Now().UTC()
	filter := bson.M{"_id": alertContact.ID, "org_id": alertContact.OrgID, "app_id": alertContact.AppID}
	update := bson.M{"$set": bson.M{
		"key":          alertContact.Key,
		"type":         alertContact.Type,
		"address":      alertContact.Address,
		"params":       alertContact.Params,
		"date_updated": now,
	}}

	res, err := a.db.alertContacts.UpdateOne(a.context, filter, update, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeAlertContact, filterArgs(filter), err)
	}
	if res.ModifiedCount != 1 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeAlertContact, filterArgs(filter), err)
	}

	return nil
}

// DeleteAlertContact deletes an alert contact
func (a *Adapter) DeleteAlertContact(id string, orgID string, appID string) error {
	filter := bson.M{"_id": id, "org_id": orgID, "app_id": appID}
	res, err := a.db.alertContacts.DeleteOne(a.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeAlertContact, filterArgs(filter), err)
	}
	if res.DeletedCount != 1 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeAlertContact, filterArgs(filter), err)
	}
	return nil
}

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
	"context"
	"time"

	"github.com/rokwire/logging-library-go/v2/logs"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type database struct {
	mongoDBAuth  string
	mongoDBName  string
	mongoTimeout time.Duration

	db       *mongo.Database
	dbClient *mongo.Client
	logger   *logs.Logger

	configs         *collectionWrapper
	surveys         *collectionWrapper
	surveyResponses *collectionWrapper
	alertContacts   *collectionWrapper
	scores			*collectionWrapper

	listeners []interfaces.StorageListener
}

func (d *database) start() error {

	d.logger.Info("database -> start")

	//connect to the database
	clientOptions := options.Client().ApplyURI(d.mongoDBAuth)
	connectContext, cancel := context.WithTimeout(context.Background(), d.mongoTimeout)
	client, err := mongo.Connect(connectContext, clientOptions)
	cancel()
	if err != nil {
		return err
	}

	//ping the database
	pingContext, cancel := context.WithTimeout(context.Background(), d.mongoTimeout)
	err = client.Ping(pingContext, nil)
	cancel()
	if err != nil {
		return err
	}

	//apply checks
	db := client.Database(d.mongoDBName)

	configs := &collectionWrapper{database: d, coll: db.Collection("configs")}
	err = d.applyConfigsChecks(configs)
	if err != nil {
		return err
	}

	surveys := &collectionWrapper{database: d, coll: db.Collection("surveys")}
	err = d.applySurveysChecks(surveys)
	if err != nil {
		return err
	}

	surveyResponses := &collectionWrapper{database: d, coll: db.Collection("survey_responses")}
	err = d.applySurveyResponsesChecks(surveyResponses)
	if err != nil {
		return err
	}

	alertContacts := &collectionWrapper{database: d, coll: db.Collection("alert_contacts")}
	err = d.applyAlertContactsChecks(alertContacts)
	if err != nil {
		return err
	}

	scores := &collectionWrapper{database: d, coll: db.Collection("scores")}
	err = d.applyScoresChecks(scores)
	if err != nil {
		return err
	}

	//assign the db, db client and the collections
	d.db = db
	d.dbClient = client

	d.configs = configs
	d.surveys = surveys
	d.surveyResponses = surveyResponses
	d.alertContacts = alertContacts
	d.scores = scores

	go d.configs.Watch(nil, d.logger)

	return nil
}

func (d *database) applyConfigsChecks(configs *collectionWrapper) error {
	d.logger.Info("apply configs checks.....")

	err := configs.AddIndex(nil, bson.D{primitive.E{Key: "type", Value: 1}, primitive.E{Key: "app_id", Value: 1}, primitive.E{Key: "org_id", Value: 1}}, true, nil)
	if err != nil {
		return err
	}

	d.logger.Info("apply configs passed")
	return nil
}

func (d *database) applySurveysChecks(surveys *collectionWrapper) error {
	d.logger.Info("apply surveys checks.....")

	err := surveys.AddIndex(nil, bson.D{primitive.E{Key: "org_id", Value: 1}, primitive.E{Key: "app_id", Value: 1}, primitive.E{Key: "creator_id", Value: 1}}, false, nil)
	if err != nil {
		return err
	}

	err = surveys.AddIndex(nil, bson.D{primitive.E{Key: "calendar_event_id", Value: 1}}, true, bson.D{primitive.E{Key: "calendar_event_id", Value: bson.M{"$gt": ""}}})
	if err != nil {
		return err
	}

	d.logger.Info("surveys passed")
	return nil
}

func (d *database) applySurveyResponsesChecks(surveyResponses *collectionWrapper) error {
	d.logger.Info("apply survey responses checks.....")

	err := surveyResponses.AddIndex(nil, bson.D{primitive.E{Key: "org_id", Value: 1}, primitive.E{Key: "app_id", Value: 1}, primitive.E{Key: "user_id", Value: 1}}, false, nil)
	if err != nil {
		return err
	}

	err = surveyResponses.AddIndex(nil, bson.D{primitive.E{Key: "survey._id", Value: 1}}, false, nil)
	if err != nil {
		return err
	}

	d.logger.Info("survey responses passed")
	return nil
}

func (d *database) applyAlertContactsChecks(alertContacts *collectionWrapper) error {
	d.logger.Info("apply alert contacts checks.....")

	err := alertContacts.AddIndex(nil, bson.D{primitive.E{Key: "org_id", Value: 1}, primitive.E{Key: "app_id", Value: 1}, primitive.E{Key: "key", Value: 1}}, false, nil)
	if err != nil {
		return err
	}

	d.logger.Info("survey alert contacts passed")
	return nil
}

func (d *database) applyScoresChecks(scores *collectionWrapper) error {
	d.logger.Info("apply scores checks.....")

	d.logger.Info("scores passed")
	return nil
}

func (d *database) onDataChanged(changeDoc map[string]interface{}) {
	if changeDoc == nil {
		return
	}
	d.logger.Infof("onDataChanged: %+v\n", changeDoc)
	ns := changeDoc["ns"]
	if ns == nil {
		return
	}
	nsMap := ns.(map[string]interface{})
	coll := nsMap["coll"]

	switch coll {
	case "configs":
		d.logger.Info("configs collection changed")

		for _, listener := range d.listeners {
			go listener.OnConfigsUpdated()
		}
	}
}

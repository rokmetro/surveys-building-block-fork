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

	"github.com/rokwire/rokwire-building-block-sdk-go/utils/errors"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetSurvey retrieves a single survey
func (a *Adapter) GetSurvey(id string, orgID string, appID string) (*model.Survey, error) {
	filter := bson.M{"_id": id, "org_id": orgID, "app_id": appID}
	var entry model.Survey
	err := a.db.surveys.FindOne(a.context, filter, &entry, nil)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionFind, model.TypeSurvey, filterArgs(filter), err)
	}
	return &entry, nil
}

// GetSurveys gets matching surveys
func (a *Adapter) GetSurveys(orgID string, appID string, creatorID *string, surveyIDs []string, surveyTypes []string, calendarEventID string, limit *int, offset *int, timeFilter *model.SurveyTimeFilter, public *bool, archived *bool, completed *bool) ([]model.Survey, error) {
	filter := bson.D{
		{Key: "org_id", Value: orgID},
		{Key: "app_id", Value: appID},
	}

	if creatorID != nil {
		filter = append(filter, bson.E{Key: "creator_id", Value: *creatorID})
	}
	if len(surveyIDs) > 0 {
		filter = append(filter, bson.E{Key: "_id", Value: bson.M{"$in": surveyIDs}})
	}
	if len(surveyTypes) > 0 {
		filter = append(filter, bson.E{Key: "type", Value: bson.M{"$in": surveyTypes}})
	}
	if calendarEventID != "" {
		filter = append(filter, bson.E{Key: "calendar_event_id", Value: calendarEventID})
	}

	if timeFilter.StartTimeAfter != nil {
		filter = append(filter, primitive.E{Key: "$or", Value: bson.A{
			bson.M{"start_date": nil},
			bson.M{"start_date": primitive.M{"$gte": *timeFilter.StartTimeAfter}},
		}})
	}
	if timeFilter.StartTimeBefore != nil {
		filter = append(filter, primitive.E{Key: "$or", Value: bson.A{
			bson.M{"start_date": nil},
			bson.M{"start_date": primitive.M{"$lte": *timeFilter.StartTimeBefore}},
		}})
	}

	if timeFilter.EndTimeAfter != nil {
		filter = append(filter, primitive.E{Key: "$or", Value: bson.A{
			bson.M{"end_date": nil},
			bson.M{"end_date": primitive.M{"$gte": *timeFilter.EndTimeAfter}},
		}})
	}
	if timeFilter.EndTimeBefore != nil {
		filter = append(filter, primitive.E{Key: "$or", Value: bson.A{
			bson.M{"end_date": nil},
			bson.M{"end_date": primitive.M{"$lte": *timeFilter.EndTimeBefore}},
		}})
	}

	if public != nil {
		if *public == true {
			filter = append(filter, bson.E{Key: "public", Value: true})
		} else {
			filter = append(filter, bson.E{Key: "$or", Value: bson.A{
				bson.M{"public": false},
				bson.M{"public": bson.M{"$exists": false}},
				bson.M{"public": nil},
			}})
		}
	}

	if archived != nil {
		if *archived == true {
			filter = append(filter, bson.E{Key: "archived", Value: true})
		} else {
			filter = append(filter, bson.E{Key: "$or", Value: bson.A{
				bson.M{"archived": false},
				bson.M{"archived": bson.M{"$exists": false}},
				bson.M{"archived": nil},
			}})
		}
	}

	opts := options.Find()
	if limit != nil {
		opts.SetLimit(int64(*limit))
	}
	if offset != nil {
		opts.SetSkip(int64(*offset))
	}
	if timeFilter.StartTimeBefore != nil {
		opts.SetSort(bson.D{{Key: "start_date", Value: -1}})
	} else if timeFilter.StartTimeAfter != nil {
		opts.SetSort(bson.D{{Key: "start_date", Value: 1}})
	}

	if timeFilter.EndTimeBefore != nil {
		opts.SetSort(bson.D{{Key: "end_date", Value: -1}})
	} else if timeFilter.EndTimeAfter != nil {
		opts.SetSort(bson.D{{Key: "end_date", Value: 1}})

	}

	var results []model.Survey
	err := a.db.surveys.Find(a.context, filter, &results, opts)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// GetSurveysWithResponses gets surveys with optional responses
func (a *Adapter) GetSurveysWithResponses(orgID string, appID string, userID *string, creatorID *string, surveyIDs []string, surveyTypes []string, calendarEventID string, limit *int, offset *int, timeFilter *model.SurveyTimeFilter, public *bool, archived *bool, completed *bool, includeResponses *bool, sortByDateCreated *bool, unstructuredProperties map[string]interface{}, query *string) ([]model.Survey, error) {
	surveyFilter := bson.D{
		{Key: "org_id", Value: orgID},
		{Key: "app_id", Value: appID},
	}

	if timeFilter == nil {
		timeFilter = &model.SurveyTimeFilter{}
	}

	if creatorID != nil {
		surveyFilter = append(surveyFilter, bson.E{Key: "creator_id", Value: *creatorID})
	}
	if len(surveyIDs) > 0 {
		surveyFilter = append(surveyFilter, bson.E{Key: "_id", Value: bson.M{"$in": surveyIDs}})
	}
	if len(surveyTypes) > 0 {
		surveyFilter = append(surveyFilter, bson.E{Key: "type", Value: bson.M{"$in": surveyTypes}})
	}
	if calendarEventID != "" {
		surveyFilter = append(surveyFilter, bson.E{Key: "calendar_event_id", Value: calendarEventID})
	}
	if timeFilter.StartTimeAfter != nil {
		surveyFilter = append(surveyFilter, primitive.E{Key: "$or", Value: bson.A{
			bson.M{"start_date": nil},
			bson.M{"start_date": primitive.M{"$gte": *timeFilter.StartTimeAfter}},
		}})
	}
	if timeFilter.StartTimeBefore != nil {
		surveyFilter = append(surveyFilter, primitive.E{Key: "$or", Value: bson.A{
			bson.M{"start_date": nil},
			bson.M{"start_date": primitive.M{"$lte": *timeFilter.StartTimeBefore}},
		}})
	}
	if timeFilter.EndTimeAfter != nil {
		surveyFilter = append(surveyFilter, primitive.E{Key: "$or", Value: bson.A{
			bson.M{"end_date": nil},
			bson.M{"end_date": primitive.M{"$gte": *timeFilter.EndTimeAfter}},
		}})
	}
	if timeFilter.EndTimeBefore != nil {
		surveyFilter = append(surveyFilter, primitive.E{Key: "$or", Value: bson.A{
			bson.M{"end_date": nil},
			bson.M{"end_date": primitive.M{"$lte": *timeFilter.EndTimeBefore}},
		}})
	}
	if public != nil {
		if *public {
			surveyFilter = append(surveyFilter, bson.E{Key: "public", Value: true})
		} else {
			surveyFilter = append(surveyFilter, bson.E{Key: "$or", Value: bson.A{
				bson.M{"public": false},
				bson.M{"public": bson.M{"$exists": false}},
				bson.M{"public": nil},
			}})
		}
	}
	if archived != nil {
		if *archived {
			surveyFilter = append(surveyFilter, bson.E{Key: "archived", Value: true})
		} else {
			surveyFilter = append(surveyFilter, bson.E{Key: "$or", Value: bson.A{
				bson.M{"archived": false},
				bson.M{"archived": bson.M{"$exists": false}},
				bson.M{"archived": nil},
			}})
		}
	}

	for key, value := range unstructuredProperties {
		if sliceValue, ok := value.([]interface{}); ok {
			surveyFilter = append(surveyFilter, bson.E{Key: "unstructured_properties." + key, Value: bson.M{"$in": sliceValue}})
		} else {
			surveyFilter = append(surveyFilter, bson.E{Key: "unstructured_properties." + key, Value: bson.M{"$eq": value}})
		}
	}

	if query != nil && *query != "" {
		surveyFilter = append(surveyFilter, bson.E{Key: "$text", Value: bson.M{"$search": query}})
	}

	// Find all surveys matching surveyFilter
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: surveyFilter}},
	}

	sortByCompletion := (public != nil && *public)
	// sort and paginate before responses lookup if not sorting public surveys by completion status
	if !sortByCompletion {
		// Consolidate to a single $sort stage
		var sortFields bson.D
		if sortByDateCreated == nil || !*sortByDateCreated {
			if timeFilter.StartTimeBefore != nil {
				sortFields = append(sortFields, bson.E{Key: "start_date", Value: -1})
			} else if timeFilter.StartTimeAfter != nil {
				sortFields = append(sortFields, bson.E{Key: "start_date", Value: 1})
			}

			if timeFilter.EndTimeBefore != nil {
				sortFields = append(sortFields, bson.E{Key: "end_date", Value: -1})
			} else if timeFilter.EndTimeAfter != nil {
				sortFields = append(sortFields, bson.E{Key: "end_date", Value: 1})
			}
		} else {
			sortFields = append(sortFields, bson.E{Key: "start_date", Value: -1})
		}
		if len(sortFields) > 0 {
			pipeline = append(pipeline, bson.D{{Key: "$sort", Value: sortFields}})
		}

		// Add pagination stages
		if offset != nil && *offset > 0 {
			pipeline = append(pipeline, bson.D{{Key: "$skip", Value: *offset}})
		}
		if limit != nil && *limit > 0 {
			pipeline = append(pipeline, bson.D{{Key: "$limit", Value: *limit}})
		}
	}

	userIDStr := ""
	if userID != nil {
		userIDStr = *userID
	}

	// Conditionally include lookup
	keepResponses := includeResponses == nil || *includeResponses
	filterByCompleted := (completed != nil)
	needLookup := userIDStr != "" && (sortByCompletion || keepResponses || filterByCompleted)
	if needLookup {
		// Build lookup pipeline and add $limit: 1 only when includeResponses is false
		responseLookupPipeline := bson.A{
			bson.D{{Key: "$match", Value: bson.M{
				"$expr": bson.M{
					"$and": bson.A{
						bson.M{"$eq": bson.A{"$survey._id", "$$surveyIdVar"}},
						bson.M{"$eq": bson.A{"$org_id", orgID}},
						bson.M{"$eq": bson.A{"$app_id", appID}},
						bson.M{"$eq": bson.A{"$user_id", userIDStr}},
					},
				},
			}}},
		}
		if !keepResponses {
			responseLookupPipeline = append(responseLookupPipeline, bson.D{{Key: "$limit", Value: 1}})
		}
		responseLookupPipeline = append(responseLookupPipeline, bson.D{{Key: "$project", Value: bson.D{
			{Key: "_id", Value: 1},
			{Key: "user_id", Value: 1},
			{Key: "date_created", Value: 1},
			{Key: "survey", Value: 1},
		}}})

		// Lookup survey_responses for each survey matching the survey_id and user_id.
		lookupStage := bson.D{
			{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "survey_responses"},
				// `surveyIdVar` holds the value of the survey's _id
				{Key: "let", Value: bson.D{{Key: "surveyIdVar", Value: "$_id"}}},
				{Key: "pipeline", Value: responseLookupPipeline},
				{Key: "as", Value: "survey_responses"},
			}},
		}
		pipeline = append(pipeline, lookupStage)

		// Compute a boolean field "Completed" based on whether a response exists
		addCompletedFieldStage := bson.D{
			{Key: "$addFields", Value: bson.D{
				{Key: "completed", Value: bson.D{
					{Key: "$cond", Value: bson.A{
						bson.D{{Key: "$gt", Value: bson.A{
							bson.D{{Key: "$size", Value: "$survey_responses"}},
							0,
						}}},
						true,
						false,
					}},
				}},
			}},
		}
		pipeline = append(pipeline, addCompletedFieldStage)

		// Optionally remove the joined survey responses from the output.
		if !keepResponses {
			projectStage := bson.D{
				{Key: "$project", Value: bson.D{
					{Key: "survey_responses", Value: 0},
				}},
			}
			pipeline = append(pipeline, projectStage)
		}
	}

	// Apply completed filter if specified
	if filterByCompleted {
		matchCriteria := bson.D{
			{Key: "completed", Value: *completed},
		}
		pipeline = append(pipeline, bson.D{{Key: "$match", Value: matchCriteria}})
	}

	// Sort survey results. Branch based on whether public sorting is desired.
	if sortByCompletion {
		// Facet the pipeline into three sections:
		//   - incompleteSurveys: not completed and have a non-null endDate; sorted by endDate ASC.
		//   - noEndDateSurveys: not completed and endDate is null; sorted by startDate DESC or dateCreated DESC.
		//   - completedSurveys: completed surveys; sorted by estimatedCompletionTime DESC, then dateCreated DESC.
		facetStage := bson.D{{Key: "$facet", Value: bson.D{
			{Key: "incompleteSurveys", Value: bson.A{
				bson.D{{Key: "$match", Value: bson.D{
					{Key: "completed", Value: false},
					{Key: "end_date", Value: bson.D{{Key: "$ne", Value: nil}}},
				}}},
				bson.D{{Key: "$sort", Value: bson.D{{Key: "end_date", Value: 1}}}},
			}},
			{Key: "noEndDateSurveys", Value: bson.A{
				bson.D{{Key: "$match", Value: bson.D{
					{Key: "completed", Value: false},
					{Key: "end_date", Value: nil},
				}}},
				bson.D{{Key: "$sort", Value: bson.D{
					{Key: "start_date", Value: -1},
					{Key: "date_created", Value: -1},
				}}},
			}},
			{Key: "completedSurveys", Value: bson.A{
				bson.D{{Key: "$match", Value: bson.D{
					{Key: "completed", Value: true},
				}}},
				bson.D{{Key: "$sort", Value: bson.D{
					{Key: "estimated_completion_time", Value: -1},
					{Key: "date_created", Value: -1},
				}}},
			}},
		}}}
		pipeline = append(pipeline, facetStage)

		// Combine the three facets into one sorted array.
		projectFacetStage := bson.D{{Key: "$project", Value: bson.D{
			{Key: "sortedResults", Value: bson.D{
				{Key: "$concatArrays", Value: bson.A{
					"$incompleteSurveys",
					"$noEndDateSurveys",
					"$completedSurveys",
				}},
			}},
		}}}
		pipeline = append(pipeline, projectFacetStage)

		// Unwind the concatenated array and set each element as the new root.
		unwindStage := bson.D{{Key: "$unwind", Value: "$sortedResults"}}
		replaceRootStage := bson.D{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$sortedResults"}}}}
		pipeline = append(pipeline, unwindStage, replaceRootStage)

		// Add pagination stages
		if offset != nil && *offset > 0 {
			pipeline = append(pipeline, bson.D{{Key: "$skip", Value: *offset}})
		}
		if limit != nil && *limit > 0 {
			pipeline = append(pipeline, bson.D{{Key: "$limit", Value: *limit}})
		}
	}

	var surveys []model.Survey
	err := a.db.surveys.Aggregate(a.context, pipeline, &surveys, nil)

	return surveys, err
}

// CreateSurvey creates a poll
func (a *Adapter) CreateSurvey(survey model.Survey) (*model.Survey, error) {
	_, err := a.db.surveys.InsertOne(a.context, survey)
	if err != nil {
		return nil, errors.WrapErrorAction(logutils.ActionCreate, model.TypeSurvey, nil, err)
	}

	return &survey, nil
}

// UpdateSurvey updates a survey
func (a *Adapter) UpdateSurvey(survey model.Survey, admin bool) error {
	if len(survey.ID) > 0 {
		now := time.Now().UTC()
		filter := bson.M{"_id": survey.ID, "org_id": survey.OrgID, "app_id": survey.AppID}
		if !admin {
			filter["creator_id"] = survey.CreatorID
		}
		update := bson.M{"$set": bson.M{
			"title":                     survey.Title,
			"more_info":                 survey.MoreInfo,
			"data":                      survey.Data,
			"scored":                    survey.Scored,
			"result_rules":              survey.ResultRules,
			"type":                      survey.Type,
			"stats":                     survey.SurveyStats,
			"default_data_key":          survey.DefaultDataKey,
			"default_data_key_rule":     survey.DefaultDataKeyRule,
			"constants":                 survey.Constants,
			"strings":                   survey.Strings,
			"sub_rules":                 survey.SubRules,
			"start_date":                survey.StartDate,
			"end_date":                  survey.EndDate,
			"public":                    survey.Public,
			"archived":                  survey.Archived,
			"estimated_completion_time": survey.EstimatedCompletionTime,
			"date_updated":              now,
			"unstructured_properties":   survey.UnstructuredProperties,
		}}

		res, err := a.db.surveys.UpdateOne(a.context, filter, update, nil)
		if err != nil {
			return errors.WrapErrorAction(logutils.ActionUpdate, model.TypeSurvey, filterArgs(filter), err)
		}
		if res.ModifiedCount != 1 {
			return errors.WrapErrorData(logutils.StatusMissing, model.TypeSurvey, filterArgs(filter), err)
		}
	}

	return nil
}

// DeleteSurvey deletes a survey
func (a *Adapter) DeleteSurvey(id string, orgID string, appID string, creatorID string, admin bool) error {
	filter := bson.M{"_id": id, "org_id": orgID, "app_id": appID}
	if !admin {
		filter["creator_id"] = creatorID
	}
	res, err := a.db.surveys.DeleteOne(a.context, filter, nil)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionDelete, model.TypeSurvey, filterArgs(filter), err)
	}
	if res.DeletedCount != 1 {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeSurvey, filterArgs(filter), err)
	}

	return nil
}

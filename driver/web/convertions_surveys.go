package web

import (
	"application/core/model"
	"sort"
	"time"
)

func surveyRequestToSurvey(item model.SurveyRequest) model.Survey {
	//start
	var startValue *time.Time
	if item.StartDate != nil {
		startValueValue, _ := time.Parse(time.RFC3339, *item.StartDate)
		startValue = &startValueValue
	}
	//end
	var endValue *time.Time
	if item.EndDate != nil {
		endValueTime, _ := time.Parse(time.RFC3339, *item.EndDate)
		endValue = &endValueTime
	}

	return model.Survey{CreatorID: item.CreatorID, OrgID: item.OrgID, AppID: item.AppID, Type: item.Type, Title: item.Title,
		MoreInfo: item.MoreInfo, Data: item.Data, Scored: item.Scored, ResultRules: item.ResultRules, ResultJSON: item.ResultJSON,
		SurveyStats: item.SurveyStats, Sensitive: item.Sensitive, Anonymous: item.Anonymous, DefaultDataKey: item.DefaultDataKey,
		DefaultDataKeyRule: item.DefaultDataKeyRule, Constants: item.Constants, Strings: item.Strings, SubRules: item.SubRules,
		ResponseKeys: item.ResponseKeys, CalendarEventID: item.CalendarEventID, StartDate: startValue, EndDate: endValue,
		Public: item.Public, Archived: item.Archived, EstimatedCompletionTime: item.EstimatedCompletionTime}
}

func getSurvey(item model.Survey) model.Survey {

	return model.Survey{CreatorID: item.CreatorID, OrgID: item.OrgID, AppID: item.AppID, Type: item.Type, Title: item.Title,
		MoreInfo: item.MoreInfo, Data: item.Data, Scored: item.Scored, ResultRules: item.ResultRules, ResultJSON: item.ResultJSON,
		SurveyStats: item.SurveyStats, Sensitive: item.Sensitive, Anonymous: item.Anonymous, DefaultDataKey: item.DefaultDataKey,
		DefaultDataKeyRule: item.DefaultDataKeyRule, Constants: item.Constants, Strings: item.Strings, SubRules: item.SubRules,
		ResponseKeys: item.ResponseKeys, CalendarEventID: item.CalendarEventID, StartDate: item.StartDate, EndDate: item.EndDate,
		Public: item.Public, Archived: item.Archived, EstimatedCompletionTime: item.EstimatedCompletionTime}
}

func getSurveys(items []model.Survey) []model.Survey {
	list := make([]model.Survey, len(items))
	for index := range items {
		list[index] = getSurvey(items[index])
	}
	return list
}

func updateSurveyRequestToSurvey(item model.SurveyRequest, id string) model.Survey {

	// start
	var startValue *time.Time
	if item.StartDate != nil {
		startValueValue, _ := time.Parse(time.RFC3339, *item.StartDate)
		startValue = &startValueValue
	}
	// end
	var endValue *time.Time
	if item.EndDate != nil {
		endValueTime, _ := time.Parse(time.RFC3339, *item.EndDate)
		endValue = &endValueTime
	}

	return model.Survey{ID: id, CreatorID: item.CreatorID, OrgID: item.OrgID, AppID: item.AppID, Type: item.Type, Title: item.Title,
		MoreInfo: item.MoreInfo, Data: item.Data, Scored: item.Scored, ResultRules: item.ResultRules, ResultJSON: item.ResultJSON,
		SurveyStats: item.SurveyStats, Sensitive: item.Sensitive, Anonymous: item.Anonymous, DefaultDataKey: item.DefaultDataKey,
		DefaultDataKeyRule: item.DefaultDataKeyRule, Constants: item.Constants, Strings: item.Strings, SubRules: item.SubRules,
		ResponseKeys: item.ResponseKeys, CalendarEventID: item.CalendarEventID, StartDate: startValue, EndDate: endValue,
		Public: item.Public, Archived: item.Archived, EstimatedCompletionTime: item.EstimatedCompletionTime}
}

func getSurveysResData(items []model.Survey, surveyResponses []model.SurveyResponse, completed *bool) []model.SurveysResponseData {
	var list []model.SurveysResponseData

	for _, item := range items {
		var isCompleted bool

		for _, surveyResponse := range surveyResponses {
			if item.ID == surveyResponse.Survey.ID {
				isCompleted = true
				break
			}
		}

		if completed == nil || *completed == isCompleted {
			list = append(list, model.SurveysResponseData{
				ID:                      item.ID,
				CreatorID:               item.CreatorID,
				OrgID:                   item.OrgID,
				AppID:                   item.AppID,
				Type:                    item.Type,
				Title:                   item.Title,
				MoreInfo:                item.MoreInfo,
				Data:                    item.Data,
				Scored:                  item.Scored,
				ResultRules:             item.ResultRules,
				ResultJSON:              item.ResultJSON,
				SurveyStats:             item.SurveyStats,
				Sensitive:               item.Sensitive,
				Anonymous:               item.Anonymous,
				DefaultDataKey:          item.DefaultDataKey,
				DefaultDataKeyRule:      item.DefaultDataKeyRule,
				Constants:               item.Constants,
				Strings:                 item.Strings,
				SubRules:                item.SubRules,
				ResponseKeys:            item.ResponseKeys,
				CalendarEventID:         item.CalendarEventID,
				StartDate:               item.StartDate,
				EndDate:                 item.EndDate,
				Public:                  item.Public,
				Archived:                item.Archived,
				EstimatedCompletionTime: item.EstimatedCompletionTime,
				Completed:               &isCompleted,
				DateCreated:             item.DateCreated,
			})
		}
	}
	// sort.Slice(list, func(i, j int) bool {
	// 	return list[i].DateCreated.After(list[j].DateCreated)
	// })

	return list
}

func sortIfpublicIsTrue(list []model.SurveysResponseData, public *bool) []model.SurveysResponseData {

	if public == nil || !*public {
		// If public is nil or false, just return the list as-is
		return list
	}

	var incompleteSurveys, noEndDateSurveys, completedSurveys []model.SurveysResponseData

	// Split surveys into categories based on completion status and end date
	for _, survey := range list {
		if survey.Completed != nil && *survey.Completed {
			completedSurveys = append(completedSurveys, survey)
		} else if survey.EndDate != nil {
			incompleteSurveys = append(incompleteSurveys, survey)
		} else {
			noEndDateSurveys = append(noEndDateSurveys, survey)
		}
	}

	// Sort incomplete surveys by end date (ascending)
	sort.Slice(incompleteSurveys, func(i, j int) bool {
		return incompleteSurveys[i].EndDate.Before(*incompleteSurveys[j].EndDate)
	})

	// Sort no-end-date surveys first by start date (descending) or by creation date if start date is missing
	sort.Slice(noEndDateSurveys, func(i, j int) bool {
		if noEndDateSurveys[i].StartDate != nil && noEndDateSurveys[j].StartDate != nil {
			return noEndDateSurveys[i].StartDate.After(*noEndDateSurveys[j].StartDate)
		}
		if noEndDateSurveys[i].StartDate != nil {
			return true
		}
		if noEndDateSurveys[j].StartDate != nil {
			return false
		}
		return noEndDateSurveys[i].DateCreated.After(noEndDateSurveys[j].DateCreated)
	})

	// Sort completed surveys by estimated completion time (descending)
	sort.Slice(completedSurveys, func(i, j int) bool {
		if completedSurveys[i].EstimatedCompletionTime != nil && completedSurveys[j].EstimatedCompletionTime != nil {
			return *completedSurveys[i].EstimatedCompletionTime > *completedSurveys[j].EstimatedCompletionTime
		}
		return completedSurveys[i].DateCreated.After(completedSurveys[j].DateCreated)
	})

	// Combine all sorted slices
	result := append(incompleteSurveys, noEndDateSurveys...)
	result = append(result, completedSurveys...)

	return result
}

func surveyTimeFilter(item *model.SurveyTimeFilterRequest) *model.SurveyTimeFilter {

	filter := model.SurveyTimeFilter{}

	if item.StartTimeBefore != nil {
		beforeStartTime, _ := time.Parse(time.RFC3339, *item.StartTimeBefore)
		filter.StartTimeBefore = &beforeStartTime
	}
	if item.StartTimeAfter != nil {
		afterStartTime, _ := time.Parse(time.RFC3339, *item.StartTimeAfter)
		filter.StartTimeAfter = &afterStartTime
	}

	if item.EndTimeBefore != nil {
		beforeEndTime, _ := time.Parse(time.RFC3339, *item.EndTimeBefore)
		filter.EndTimeBefore = &beforeEndTime
	}
	if item.EndTimeAfter != nil {
		afterEndTime, _ := time.Parse(time.RFC3339, *item.EndTimeAfter)
		filter.EndTimeAfter = &afterEndTime
	}

	return &model.SurveyTimeFilter{
		StartTimeAfter:  filter.StartTimeAfter,
		StartTimeBefore: filter.StartTimeBefore,
		EndTimeAfter:    filter.EndTimeAfter,
		EndTimeBefore:   filter.EndTimeBefore}
}

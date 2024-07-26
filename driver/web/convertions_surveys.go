package web

import (
	"application/core/model"
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
			})
		}
	}

	return list
}

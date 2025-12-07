package web

import (
	"application/core/model"
	Def "application/driver/web/docs/gen"
	"time"
)

func scoreToDef(item *model.Score) *Def.Score {
	if item == nil {
		return nil
	}

	answerCount := float32(item.AnswerCount)
	correctAnswerCount := float32(item.CorrectAnswerCount)
	currentStreak := float32(item.CurrentStreak)
	rank := float32(item.Rank)
	responseCount := float32(item.ResponseCount)
	score := float32(item.Score)
	streakMultiplier := float32(item.StreakMultiplier)
	prevSurveyResponseDate := item.PrevSurveyResponseDate.Format(time.RFC3339)

	return &Def.Score{
		AnswerCount:            &answerCount,
		AppId:                  &item.AppID,
		CorrectAnswerCount:     &correctAnswerCount,
		CurrentStreak:          &currentStreak,
		ExternalProfileId:      &item.ExternalProfileID,
		ExternalUserId:         &item.ExternalUserID,
		Id:                     &item.ID,
		OrgId:                  &item.OrgID,
		PrevSurveyResponseDate: &prevSurveyResponseDate,
		Rank:                   &rank,
		ResponseCount:          &responseCount,
		Score:                  &score,
		StreakMultiplier:       &streakMultiplier,
		SurveyType:             &item.SurveyType,
		UserId:                 &item.UserID,
	}
}

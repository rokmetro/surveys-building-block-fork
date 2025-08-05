package main

import (
	"application/core/model"
	"application/utils"
	"io/ioutil"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func main() {
	calculateScore()
}

func calculateScore() {
	var surveyResponses []model.SurveyResponse

	// Read the JSON file from responses folder
	data, err := ioutil.ReadFile("responses/surveys.survey_responses_1.json")
	if err != nil {
		log.Fatalf("Error reading file: %v", err)
	}

	// Unmarshal JSON into surveyResponses
	err = bson.UnmarshalExtJSON(data, false, &surveyResponses)
	if err != nil {
		log.Fatalf("Error unmarshaling BSON: %v", err)
	}

	log.Printf("Number of survey responses: %d", len(surveyResponses))

	score := model.Score{
		ID:                 "",
		OrgID:              "",
		AppID:              "",
		UserID:             "",
		ExternalProfileID:  "",
		Score:              0,
		ResponseCount:      0,
		CurrentStreak:      0,
		AnswerCount:        0,
		CorrectAnswerCount: 0,
		SurveyType:         model.SurveyTypeFashionQuiz,
	}

	for i := len(surveyResponses) - 1; i >= 0; i-- {
		updateScore(&score, surveyResponses[i])
	}

	log.Printf("Score: %v", score.Score)
	log.Printf("Streak: %v", score.CurrentStreak)
}

// updateScore updates the score model passed in
// Assumes that surveyResponse is a fashion quiz
func updateScore(score *model.Score, surveyResponse model.SurveyResponse) {
	survey := surveyResponse.Survey
	score.ResponseCount++
	score.AnswerCount += uint32(survey.SurveyStats.Total)
	pointsForResponse := float64(survey.SurveyStats.Scores[""])

	if survey.SurveyStats.CorrectAnswerCount == 0 && pointsForResponse >= 0 {
		// Handle clients that don't send correct answer count by assuming
		// pointsForResponse == correct answer count
		score.CorrectAnswerCount += uint32(pointsForResponse)
	} else {
		score.CorrectAnswerCount += uint32(survey.SurveyStats.CorrectAnswerCount)
	}

	responseTime := surveyResponse.DateCreated
	unstructProps := survey.UnstructuredProperties
	if unstructProps == nil {
		unstructProps = survey.UnstructuredPropertiesDep
	}
	if unstructProps != nil {
		externalProfileIDRaw, exists := unstructProps["external_profile_id"]
		if exists {
			externalProfileIDStr, isString := externalProfileIDRaw.(string)
			if isString {
				score.ExternalProfileID = externalProfileIDStr
			}
		}

		localResponseTimeRaw, exists := unstructProps["local_time"]
		if exists {
			localResponseTimeStr, isString := localResponseTimeRaw.(string)
			if isString {
				localResponseTime, err := time.Parse(time.DateTime, localResponseTimeStr)
				// Ensure that local time (from user's device clock) is current (within 24 hours of survey response date)
				if err == nil && surveyResponse.DateCreated.Sub(localResponseTime).Abs().Hours() < 24 {
					responseTime = localResponseTime
				} else {
					log.Printf("Error parsing local time: %v", err)
				}
			}
		}
	}

	if utils.IsPrevOrSameDay(score.PrevSurveyResponseDate, responseTime) {
		// log.Printf("Not updating streak (%d) on %v (%s)", score.CurrentStreak, responseTime, survey.ID)
		// Don't update streak if day is same or previous
	} else if utils.IsNextDay(score.PrevSurveyResponseDate, responseTime) {
		// Update streak
		score.CurrentStreak++
	} else {
		// Reset streak to day 1
		log.Printf("Resetting streak (%d) on %v (%s)", score.CurrentStreak, responseTime, survey.ID)
		score.CurrentStreak = 1
	}
	score.PrevSurveyResponseDate = responseTime

	if score.CurrentStreak >= model.ScoreStreakMinDays {
		score.Score += pointsForResponse * model.ScoreStreakMultiplier
	} else {
		// log.Printf("Not applying streak (%d) on %v (%s)", score.CurrentStreak, responseTime, survey.ID)
		score.Score += pointsForResponse
	}
}

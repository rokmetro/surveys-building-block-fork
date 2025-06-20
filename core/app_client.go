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
	"application/core/interfaces"
	"application/core/model"
	"application/utils"
	"time"

	"github.com/google/uuid"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/errors"
	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logutils"
)

// appClient contains client implementations
type appClient struct {
	app *Application
}

// Surveys
// GetSurvey returns the survey with the provided ID
func (a appClient) GetSurvey(id string, orgID string, appID string) (*model.Survey, error) {
	return a.app.shared.getSurvey(id, orgID, appID)
}

// GetSurvey returns surveys matching the provided query
func (a appClient) GetSurveys(orgID string, appID string, userID *string, creatorID *string, surveyIDs []string, surveyTypes []string, calendarEventID string,
	limit *int, offset *int, filter *model.SurveyTimeFilter, public *bool, archived *bool, completed *bool, includeResponses *bool) ([]model.Survey, error) {
	return a.app.shared.getSurveys(orgID, appID, userID, creatorID, surveyIDs, surveyTypes, calendarEventID, limit, offset, filter, public, archived, completed, includeResponses, nil)
}

// CreateSurvey creates a new survey
func (a appClient) CreateSurvey(survey model.Survey, externalIDs map[string]string) (*model.Survey, error) {
	return a.app.shared.createSurvey(survey, externalIDs)
}

// UpdateSurvey updates the provided survey
func (a appClient) UpdateSurvey(survey model.Survey, userID string, externalIDs map[string]string) error {
	return a.app.shared.updateSurvey(survey, userID, externalIDs, false)
}

// DeleteSurvey deletes the survey with the specified ID
func (a appClient) DeleteSurvey(id string, orgID string, appID string, userID string, externalIDs map[string]string) error {
	return a.app.shared.deleteSurvey(id, orgID, appID, userID, externalIDs, false)
}

// Survey Response
// GetSurveyResponse returns the survey response with the provided ID
func (a appClient) GetSurveyResponse(id string, orgID string, appID string, userID string) (*model.SurveyResponse, error) {
	return a.app.storage.GetSurveyResponse(id, orgID, appID, userID)
}

// GetUserSurveyResponses returns the survey responses matching the provided filters for a specific user
func (a appClient) GetUserSurveyResponses(orgID string, appID string, userID string, surveyIDs []string, surveyTypes []string, startDate *time.Time, endDate *time.Time, limit *int, offset *int) ([]model.SurveyResponse, error) {
	return a.app.storage.GetSurveyResponses(&orgID, &appID, &userID, surveyIDs, surveyTypes, startDate, endDate, limit, offset)
}

// GetAllSurveyResponses returns the survey responses matching the provided filters
func (a appClient) GetAllSurveyResponses(orgID string, appID string, userID string, surveyID string, startDate *time.Time, endDate *time.Time, limit *int, offset *int, externalIDs map[string]string) ([]model.SurveyResponse, error) {
	var allResponses []model.SurveyResponse
	var err error

	survey, err := a.app.shared.getSurvey(surveyID, orgID, appID)
	if err != nil {
		return nil, err
	}

	// Check if survey is sensitive
	if survey.Sensitive {
		return nil, errors.Newf("Survey is sensitive and responses are not available")
	}

	// If no calendar event is associated then user should not have access to responses
	if survey.CalendarEventID == "" {
		return nil, errors.Newf("Survey responses are not available. No calendar event associated")
	}

	// Check if user is admin of calendar event
	admin, err := a.app.shared.isEventAdmin(survey.OrgID, survey.AppID, survey.CalendarEventID, userID, externalIDs)
	if err != nil {
		return nil, errors.WrapErrorAction("checking", "event admin", nil, err)
	}
	if !admin {
		return nil, errors.ErrorData(logutils.StatusInvalid, "user", &logutils.FieldArgs{"calendar_event_id": survey.CalendarEventID, "admin": false})
	}

	// Get responses
	allResponses, err = a.app.storage.GetSurveyResponses(&orgID, &appID, nil, []string{surveyID}, nil, startDate, endDate, limit, offset)
	if err != nil {
		return nil, err
	}

	// If survey is anonymous strip userIDs
	if survey.Anonymous {
		for i := range allResponses {
			allResponses[i].UserID = ""
		}
	}

	return allResponses, nil
}

// CreateSurveyResponse creates a new survey response
func (a appClient) CreateSurveyResponse(surveyResponse model.SurveyResponse, externalIDs map[string]string) (*model.SurveyResponse, error) {
	surveyResponse.ID = uuid.NewString()
	surveyResponse.DateCreated = time.Now().UTC()
	surveyResponse.DateUpdated = nil

	// Get survey from storage
	survey, err := a.app.storage.GetSurvey(surveyResponse.Survey.ID, surveyResponse.OrgID, surveyResponse.AppID)
	if err != nil {
		return nil, err
	}
	// Populate survey with data from client request
	survey.Data = surveyResponse.Survey.Data
	survey.SurveyStats = surveyResponse.Survey.SurveyStats
	survey.ResultJSON = surveyResponse.Survey.ResultJSON
	survey.UnstructuredProperties = surveyResponse.Survey.UnstructuredProperties
	surveyResponse.Survey = *survey

	if survey.CalendarEventID != "" {
		// check if user attended calendar event
		attended, err := a.app.shared.hasAttendedEvent(surveyResponse.OrgID, surveyResponse.AppID, survey.CalendarEventID, surveyResponse.UserID, externalIDs)
		if err != nil {
			return nil, errors.WrapErrorAction("checking", "event attendance", nil, err)
		}
		if !attended {
			return nil, errors.Newf("user has not attended calendar event")
		}
	}

	surveyResponsePtr, err := a.app.storage.CreateSurveyResponse(surveyResponse)

	// If the user completed a fashion quiz, update the score
	if err == nil && survey.Type == model.SurveyTypeFashionQuiz {
		var score *model.Score
		score, err = a.app.storage.GetScore(surveyResponse.OrgID, surveyResponse.AppID, surveyResponse.UserID)

		// Create a new score if not present
		// Otherwise update score
		if score == nil || err != nil {
			_, err = a.CreateScore(surveyResponse.OrgID, surveyResponse.AppID, surveyResponse.UserID, "")
		} else {
			a.UpdateScore(score, surveyResponse)
			err = a.app.storage.UpdateScore(*score)
		}
	}

	return surveyResponsePtr, err
}

// UpdateSurveyResponse updates the provided survey response
func (a appClient) UpdateSurveyResponse(surveyResponse model.SurveyResponse) error {
	return a.app.storage.UpdateSurveyResponse(surveyResponse)
}

// DeleteSurveyResponse deletes the survey with the specified ID
func (a appClient) DeleteSurveyResponse(id string, orgID string, appID string, userID string) error {
	return a.app.storage.DeleteSurveyResponse(id, orgID, appID, userID)
}

// DeleteSurveyResponses deletes the survey responses matching the provided filters
func (a appClient) DeleteSurveyResponses(orgID string, appID string, userID string, surveyIDs []string, surveyTypes []string, startDate *time.Time, endDate *time.Time) error {
	return a.app.storage.DeleteSurveyResponses(orgID, appID, userID, surveyIDs, surveyTypes, startDate, endDate)
}

// Survey Alerts
// CreateSurveyAlert creates a new survey alert
func (a appClient) CreateSurveyAlert(surveyAlert model.SurveyAlert) error {
	contacts, err := a.app.storage.GetAlertContactsByKey(surveyAlert.ContactKey, surveyAlert.OrgID, surveyAlert.AppID)
	if err != nil {
		return err
	}

	for i := 0; i < len(contacts); i++ {
		if contacts[i].Type == "email" {
			subject, ok := surveyAlert.Content["subject"].(string)
			if !ok {
				return errors.ErrorData(logutils.StatusMissing, "subject", nil)
			}
			body, ok := surveyAlert.Content["body"].(string)
			if !ok {
				return errors.ErrorData(logutils.StatusMissing, "body", nil)
			}
			a.app.notifications.SendMail(contacts[i].Address, subject, body)
		}
	}

	return nil
}

// GetScore gets scores and creates one if it doesn't exist
func (a appClient) GetScore(orgID string, appID string, userID string, externalProfileID string) (*model.Score, error) {
	score, err := a.app.storage.GetScore(orgID, appID, userID)
	if score == nil {
		score, err = a.CreateScore(orgID, appID, userID, externalProfileID)
	}
	if err != nil || score == nil {
		return nil, err
	}

	// If score object doesn't have externalID, store the one that's client-provided
	if score.ExternalProfileID == "" && externalProfileID != "" {
		score.ExternalProfileID = externalProfileID
		err = a.app.storage.UpdateScore(*score)
	}

	score.StreakMultiplier = model.ScoreStreakMultiplier
	return score, err
}

// GetScores returns scores in descending order and removes scores with empty external IDs
func (a appClient) GetScores(orgID string, appID string, limit *int, offset *int) ([]model.Score, error) {
	return a.app.storage.GetScores(orgID, appID, limit, offset)
}

// GetScoresWithPivot retrieves scores closest to the user's score
func (a appClient) GetTopAndLocalScores(orgID string, appID string, userID string, limit *int, offset *int, localLimit *int, abovePivotLimit *int, belowPivotLimit *int) ([]model.Score, error) {
	// We replace the local limit with the equal limit when getting scores from database
	scores, err := a.app.storage.GetTopAndLocalScores(orgID, appID, userID, limit, offset, abovePivotLimit, localLimit, belowPivotLimit)
	if err != nil {
		return nil, err
	}

	userScore := scores[0]
	scores = scores[1:]

	userInTopScoresIdx := -1
	for i, score := range scores {
		if score.UserID == userID {
			userInTopScoresIdx = i
			break
		}
	}

	if len(scores) > *limit {
		if userInTopScoresIdx == -1 {
			// Insert user's score into section with surrounding local scores if user is not in top scores
			localScores := a.insertUserScoreIntoLocalScores(scores[*limit:], userScore, *localLimit)
			scores = scores[:*limit]
			scores = append(scores, localScores...)
		} else if userInTopScoresIdx >= *limit-*belowPivotLimit {
			// If user is in top scores, but close to the bottom, we append scores according to the local and belowPivot limits
			scoresLength := userInTopScoresIdx + *belowPivotLimit + 1
			if scoresLength > len(scores) {
				scoresLength = len(scores)
			}
			scores = scores[:scoresLength]
		} else {
			// If user is in top scores, but not close to the bottom, we trim the scores to the limit
			scores = scores[:*limit]
		}
	}

	return scores, nil
}

func (a appClient) insertUserScoreIntoLocalScores(scores []model.Score, userScore model.Score, maxNumScores int) []model.Score {
	insertionIndex := (len(scores) + 1) / 2

	// We take the first and last occurrences of equal scores to the user's
	firstEqual, lastEqual := -1, -1
	for i, s := range scores {
		if s.Score == userScore.Score {
			firstEqual = i
			break
		}
	}
	for i := len(scores) - 1; i >= 0; i-- {
		if scores[i].Score == userScore.Score {
			lastEqual = i
			break
		}
	}

	if firstEqual == -1 {
		// We insert the user's score in its obvious place if no other
		// scores are equal
		insertionIndex = len(scores)
		for i, s := range scores {
			if s.Score < userScore.Score {
				insertionIndex = i
				break
			}
		}
	}

	// if there were equal scores, clamp insertionIndex so it
	// sits beside them
	if firstEqual != -1 {
		if insertionIndex < firstEqual {
			insertionIndex = firstEqual
		} else if insertionIndex > lastEqual+1 {
			insertionIndex = lastEqual + 1
		}
	}

	// actually insert userScore
	buf := make([]model.Score, 0, len(scores)+1)
	buf = append(buf, scores[:insertionIndex]...)
	buf = append(buf, userScore)
	buf = append(buf, scores[insertionIndex:]...)
	scores = buf

	// Take a centered window of length maxNumScores
	half := int(maxNumScores / 2)
	start := insertionIndex - half
	if start < 0 {
		start = 0
	}
	end := start + maxNumScores
	if end > len(scores) {
		end = len(scores)
	}

	return scores[start:end]
}

// CreateScore Creates a score object by iterating over all previous survey responses
func (a appClient) CreateScore(orgID string, appID string, userID string, externalProfileID string) (*model.Score, error) {
	surveyResponses, err := a.app.storage.GetSurveyResponses(&orgID, &appID, &userID, nil, []string{model.SurveyTypeFashionQuiz}, nil, nil, nil, nil)
	if err != nil {
		return nil, err
	}

	score := model.Score{
		ID:                 uuid.NewString(),
		OrgID:              orgID,
		AppID:              appID,
		UserID:             userID,
		ExternalProfileID:  externalProfileID,
		Score:              0,
		ResponseCount:      0,
		CurrentStreak:      0,
		AnswerCount:        0,
		CorrectAnswerCount: 0,
		SurveyType:         model.SurveyTypeFashionQuiz,
	}

	for i := 0; i < len(surveyResponses); i++ {
		a.UpdateScore(&score, surveyResponses[i])
	}
	return &score, a.app.storage.CreateScore(score)
}

// UpdateScore updates the score model passed in
// Assumes that surveyResponse is a fashion quiz
func (a appClient) UpdateScore(score *model.Score, surveyResponse model.SurveyResponse) error {
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

	externalProfileIDRaw, exists := survey.UnstructuredProperties["external_profile_id"]
	if exists {
		externalProfileIDStr, isString := externalProfileIDRaw.(string)
		if isString {
			score.ExternalProfileID = externalProfileIDStr
		}
	}

	localResponseTimeRaw, exists := survey.UnstructuredProperties["local_time"]
	responseTime := surveyResponse.DateCreated
	if exists {
		localResponseTimeStr, isString := localResponseTimeRaw.(string)
		if isString {
			localResponseTime, err := time.Parse(time.DateTime, localResponseTimeStr)

			if err == nil && time.Since(localResponseTime).Abs().Hours() < 24 {
				responseTime = localResponseTime
			}
		}
	}

	if utils.IsPrevOrSameDay(score.PrevSurveyResponseDate, responseTime) {
		// Don't update streak if day is same or previous
	} else if utils.IsNextDay(score.PrevSurveyResponseDate, responseTime) {
		// Update streak
		score.CurrentStreak++
	} else {
		// Reset streak to day 1
		score.CurrentStreak = 1
	}
	score.PrevSurveyResponseDate = responseTime

	if score.CurrentStreak >= model.ScoreStreakMinDays {
		score.Score += pointsForResponse * model.ScoreStreakMultiplier
	} else {
		score.Score += pointsForResponse
	}

	return nil
}

// GetLeaderboards gets all leaderboards for a user
func (a appClient) GetLeaderboards(orgID string, appID string, userID string) ([]model.Leaderboard, error) {
	return a.app.storage.GetLeaderboards(orgID, appID, userID)
}

// GetLeaderboardScores returns the paginated scores in the leaderboard with the provided ID
func (a appClient) GetLeaderboardScores(leaderboardID string, orgID string, appID string, limit *int, offset *int) ([]model.Score, error) {
	return a.app.storage.GetLeaderboardScores(leaderboardID, orgID, appID, limit, offset)
}

// GetLeaderboardUserScores returns the scores of a user in each leaderboard
func (a appClient) GetLeaderboardUserScores(orgID string, appID string, userID string, limit *int, offset *int) ([]model.Score, error) {
	return a.app.storage.GetLeaderboardUserScores(orgID, appID, userID, limit, offset)
}

// CreateLeaderboard creates a new leaderboard
func (a appClient) CreateLeaderboard(leaderboard model.Leaderboard, userID string) (*model.Leaderboard, error) {
	leaderboard.ID = uuid.NewString()
	leaderboard.DateCreated = time.Now().UTC()
	leaderboard.DateUpdated = nil

	leaderboardEntry := a.createLeaderboardEntry(leaderboard.ID, leaderboard.OrgID, leaderboard.AppID, userID, true)

	transaction := func(storage interfaces.Storage) error {
		//1. Create leaderboard
		_, err := storage.CreateLeaderboard(leaderboard)
		if err != nil {
			return err
		}

		//2. Create corresponding leaderboard entry
		err = storage.CreateLeaderboardEntry(leaderboardEntry)
		if err != nil {
			return err
		}

		return nil
	}

	err := a.app.storage.PerformTransaction(transaction)
	if err != nil {
		return nil, err
	}
	leaderboard.IsAdmin = true
	return &leaderboard, nil
}

// UpdateLeaderboard updates an existing leaderboard
func (a appClient) UpdateLeaderboard(leaderboard model.Leaderboard, userID string) error {
	// Check if user is an admin of the leaderboard
	err := a.requireLeaderboardAdmin(leaderboard.ID, leaderboard.OrgID, leaderboard.AppID, userID)
	if err != nil {
		return err
	}

	time := time.Now().UTC()
	leaderboard.DateUpdated = &time

	return a.app.storage.UpdateLeaderboard(leaderboard)
}

// DeleteLeaderboard deletes a leaderboard by ID
func (a appClient) DeleteLeaderboard(leaderboardID string, orgID string, appID string, userID string) error {
	// Check if user is an admin of the leaderboard
	err := a.requireLeaderboardAdmin(leaderboardID, orgID, appID, userID)
	if err != nil {
		return err
	}

	transaction := func(storage interfaces.Storage) error {
		//1. Delete leaderboard
		err := storage.DeleteLeaderboard(leaderboardID, orgID, appID, userID)
		if err != nil {
			return err
		}

		//2. Delete all leaderboard entries corresponding to leaderboardID
		err = storage.DeleteAllLeaderboardEntries(leaderboardID, orgID, appID)
		if err != nil {
			return err
		}

		return nil
	}

	return a.app.storage.PerformTransaction(transaction)
}

func (a appClient) JoinLeaderboard(leaderboardID string, orgID string, appID string, userID string) error {
	return a.app.storage.CreateLeaderboardEntry(a.createLeaderboardEntry(leaderboardID, orgID, appID, userID, false))
}

// createLeaderboardEntry creates a model.LeaderboardEntry with the provided parameters
func (a appClient) createLeaderboardEntry(leaderboardID string, orgID string, appID string, userID string, isAdmin bool) model.LeaderboardEntry {
	return model.LeaderboardEntry{
		ID:            uuid.NewString(),
		LeaderboardID: leaderboardID,
		OrgID:         orgID,
		AppID:         appID,
		UserID:        userID,
		IsAdmin:       isAdmin,
		DateCreated:   time.Now().UTC(),
		DateUpdated:   nil,
	}
}

func (a appClient) LeaveLeaderboard(leaderboardID string, orgID string, appID string, userID string, leavingUserIDs []string) error {
	if len(leavingUserIDs) == 0 {
		// If leavingUserIDs aren't provided, remove the current user from the leaderboard
		leavingUserIDs = append(leavingUserIDs, userID)
	} else if !(len(leavingUserIDs) == 1 && leavingUserIDs[0] == userID) {
		// Only remove specified users other than the user if the current user is an admin
		err := a.requireLeaderboardAdmin(leaderboardID, orgID, appID, userID)
		if err != nil {
			return err
		}
	}

	return a.app.storage.DeleteLeaderboardEntries(leaderboardID, orgID, appID, leavingUserIDs)
}

func (a appClient) requireLeaderboardAdmin(leaderboardID string, orgID string, appID string, userID string) error {
	leaderboardEntry, err := a.app.storage.GetLeaderboardEntry(leaderboardID, orgID, appID, userID)
	if err != nil {
		return errors.WrapErrorData(logutils.StatusMissing, model.TypeLeaderboardEntry, &logutils.FieldArgs{"leaderboard_id": leaderboardID, "org_id": orgID, "app_id": appID, "user_id": userID}, err)
	}

	if !leaderboardEntry.IsAdmin {
		return errors.Newf("User %s is not an admin of leaderboard %s", userID, leaderboardID)
	}

	return nil
}

// newAppClient creates new appClient
func newAppClient(app *Application) appClient {
	return appClient{app: app}
}

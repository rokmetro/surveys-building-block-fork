/*
 *   Copyright (c) 2025 Board of Trustees of the University of Illinois.
 *   All rights reserved.

 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package core

import (
	"application/core/interfaces"
	"application/core/model"
	"application/driven/notifications"
	"application/utils"
	"fmt"
	"time"

	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
)

type streakNotifications struct {
	application *Application
	logger      *logs.Logger

	notifications interfaces.Notifications

	storage interfaces.Storage

	//streak notifications timer
	streakNotificationsTimer     *time.Timer
	streakNotificationsTimerDone chan bool
}

func (n streakNotifications) start() {
	//setup notifications timer
	go n.setupStreakNotificationsTimer()
}

func (n streakNotifications) setupStreakNotificationsTimer() {
	// default to 10AM EST/EDT
	desiredMoment := 36000            // default desired moment of the day in seconds (beginning of the day)
	locationStr := "America/New_York" // default location

	envConfig, err := n.application.GetEnvConfigs()
	if envConfig == nil || err != nil {
		n.logger.Infof("setupStreakNotificationsTimer -> failed to get env config (%v) - applying defaults", err)
	} else {
		desiredMoment = envConfig.StreakNotificationsTimerMoment
		locationStr = envConfig.StreakNotificationsTimezone
	}

	location, err := time.LoadLocation(locationStr)
	if err != nil {
		n.logger.Warnf("setupStreakNotificationsTimer -> error getting location: %v", err)
	}
	now := time.Now().In(location)
	nowSecondsInDay := utils.SecondsInHour*now.Hour() + utils.SecondsInMinute*now.Minute() + now.Second()

	var durationInSeconds int
	n.logger.Infof("setupStreakNotificationsTimer -> nowSecondsInDay:%d", nowSecondsInDay)
	if nowSecondsInDay <= desiredMoment {
		n.logger.Info("setupStreakNotificationsTimer -> notifications not yet processed today")
		durationInSeconds = desiredMoment - nowSecondsInDay
	} else {
		n.logger.Info("setupStreakNotificationsTimer -> notifications have already been processed today")
		durationInSeconds = (utils.SecondsInDay - nowSecondsInDay) + desiredMoment // the time which left today + desired moment from next day
	}

	initialDuration := time.Second * time.Duration(durationInSeconds)
	utils.StartTimer(n.streakNotificationsTimer, n.streakNotificationsTimerDone, &initialDuration, time.Duration(utils.HoursInDay)*time.Hour, n.processNotifications, "processStreakNotifications", n.logger)
}

func (n streakNotifications) processNotifications() {
	// Streak reminder notification: search scores collection for users who have not submitted a response yet today
	nowDay := time.Now().UTC().Truncate(time.Hour)
	prevDay := nowDay.Add(-time.Duration(utils.HoursInDay) * time.Hour)

	timersData, err := n.application.CheckTimersConfig("streak_notifications_last_process_time", prevDay, nowDay)
	if err != nil || timersData == nil {
		n.logger.Warnf("processNotifications -> error finding timers config: %v", err)
		return
	}

	scores, err := n.storage.GetScores(nil, nil, nil, nil, &prevDay, &nowDay)
	if err != nil {
		n.logger.Errorf("processNotifications -> error finding scores: %v", err)
		return
	}

	for _, score := range scores {
		body := fmt.Sprintf("You're on a %d-day Runway Genius streak! Play now to keep it going.", score.CurrentStreak)
		topic := notifications.TopicQuizAll

		data := map[string]string{
			"url": fmt.Sprintf("%s/quiz/landing", notifications.BaseURLVogue),
		}

		message := model.NotificationMessage{
			OrgID: score.OrgID,
			AppID: score.AppID,

			Subject:    notifications.SubjectVogue,
			Body:       body,
			Data:       data,
			Recipients: []model.NotificationMessageRecipient{{UserID: score.UserID}},
			Topic:      &topic,
		}
		n.notifications.SendNotification(message)
	}
}

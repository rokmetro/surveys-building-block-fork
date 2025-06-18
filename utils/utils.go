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

package utils

import (
	"crypto/sha256"
	"time"

	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
)

const (
	// HoursInDay is the number of hours in one day
	HoursInDay int = 24
	// SecondsInDay is the number of seconds in one 24-hour day
	SecondsInDay int = HoursInDay * SecondsInHour
	// SecondsInHour is the number of seconds in one hour
	SecondsInHour int = 60 * SecondsInMinute
	// SecondsInMinute is the number of seconds in one minute
	SecondsInMinute int = 60
)

// GetInt gives the value which this pointer points. Gives 0 if the pointer is nil
func GetInt(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

// GetString gives the value which this pointer points. Gives empty string if the pointer is nil
func GetString(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}

// GetTime gives the value which this pointer points. Gives empty string if the pointer is nil
func GetTime(time *time.Time) string {
	if time == nil {
		return ""
	}
	return time.String()
}

// SHA256Hash computes the SHA256 hash of a byte slice
func SHA256Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// IsNextDay checks if compare is exactly one day after current
func IsNextDay(current time.Time, compare time.Time) bool {
	// Normalize both dates to midnight
	current = time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, current.Location())
	compare = time.Date(compare.Year(), compare.Month(), compare.Day(), 0, 0, 0, 0, compare.Location())

	return compare.Equal(current.AddDate(0, 0, 1))
}

// IsPrevOrSameDay checks if compare is exactly current or before
func IsPrevOrSameDay(current time.Time, compare time.Time) bool {
	// Normalize both dates to midnight
	current = time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, current.Location())
	compare = time.Date(compare.Year(), compare.Month(), compare.Day(), 0, 0, 0, 0, compare.Location())

	return compare.Equal(current) || compare.Before(current)
}

// StartTimer starts a timer with the given name, period, and function to call when the timer goes off
func StartTimer(timer *time.Timer, timerDone chan bool, initialDuration *time.Duration, period time.Duration, periodicFunc func(), name string, logger *logs.Logger) {
	if logger != nil {
		logger.Info("start timer for " + name)
	}

	//cancel if active
	if timer != nil {
		if logger != nil {
			logger.Info(name + " -> there is active timer, so cancel it")
		}

		timerDone <- true
		timer.Stop()
	}

	onTimer(timer, timerDone, initialDuration, period, periodicFunc, name, logger)
}

func onTimer(timer *time.Timer, timerDone chan bool, initialDuration *time.Duration, period time.Duration, periodicFunc func(), name string, logger *logs.Logger) {
	hasLogger := (logger != nil)
	if hasLogger {
		logger.Info(name)
	}

	duration := period
	if initialDuration != nil {
		duration = *initialDuration
	} else {
		periodicFunc()
	}
	timer = time.NewTimer(duration)

	if hasLogger {
		logger.Infof(name+" -> next call after %s", duration)
	}

	select {
	case <-timer.C:
		// timer expired
		if hasLogger {
			logger.Info(name + " -> timer expired")
		}
		timer = nil

		onTimer(timer, timerDone, nil, period, periodicFunc, name, logger)
	case <-timerDone:
		// timer aborted
		if hasLogger {
			logger.Info(name + " -> timer aborted")
		}
		timer = nil
	}
}

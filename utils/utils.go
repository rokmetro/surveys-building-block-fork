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

func IsNextDay(current time.Time, compare time.Time) bool {
	// Normalize both dates to midnight
	current = time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, current.Location())
	compare = time.Date(compare.Year(), compare.Month(), compare.Day(), 0, 0, 0, 0, compare.Location())

	// Check if compare is exactly one day after current
	return compare.Equal(current.AddDate(0, 0, 1))
}

func IsPrevOrSameDay(current time.Time, compare time.Time) bool {
	// Normalize both dates to midnight
	current = time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, current.Location())
	compare = time.Date(compare.Year(), compare.Month(), compare.Day(), 0, 0, 0, 0, compare.Location())

	// Check if compare is exactly current or before
	return compare.Equal(current) || compare.Before(current)
}

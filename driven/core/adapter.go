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

package corebb

import (
	"application/core/model"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/rokwire/core-auth-library-go/v3/authservice"
	"github.com/rokwire/logging-library-go/v2/logs"
)

// Adapter implements the Core interface
type Adapter struct {
	logger                logs.Logger
	coreURL               string
	serviceAccountManager *authservice.ServiceAccountManager
}

// NewCoreAdapter creates a new adapter for Core API
func NewCoreAdapter(coreURL string, serviceAccountManager *authservice.ServiceAccountManager) *Adapter {
	return &Adapter{coreURL: coreURL, serviceAccountManager: serviceAccountManager}
}

// LoadDeletedMemberships loads deleted memberships
func (a *Adapter) LoadDeletedMemberships() ([]model.DeletedUserData, error) {

	if a.serviceAccountManager == nil {
		log.Println("LoadDeletedMemberships: service account manager is nil")
		return nil, errors.New("service account manager is nil")
	}

	url := fmt.Sprintf("%s/bbs/deleted-memberships?service_id=%s", a.coreURL, a.serviceAccountManager.AuthService.ServiceID)

	// Create a new HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("delete membership: error creating request - %s", err)
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := a.serviceAccountManager.MakeRequest(req, "all", "all")
	if err != nil {
		log.Printf("LoadDeletedMemberships: error sending request - %s", err)
		return nil, err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Printf("LoadDeletedMemberships: error with response code - %d", resp.StatusCode)
		return nil, fmt.Errorf("LoadDeletedMemberships: error with response code != 200")
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("LoadDeletedMemberships: unable to read json: %s", err)
		return nil, fmt.Errorf("LoadDeletedMemberships: unable to parse json: %s", err)
	}

	var deletedMemberships []model.DeletedUserData
	err = json.Unmarshal(data, &deletedMemberships)
	if err != nil {
		log.Printf("LoadDeletedMemberships: unable to parse json: %s", err)
		return nil, fmt.Errorf("LoadDeletedMemberships: unable to parse json: %s", err)
	}

	return deletedMemberships, nil
}

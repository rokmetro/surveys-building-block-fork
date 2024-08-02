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

package main

import (
	"application/core"
	"application/core/model"
	"application/driven/calendar"
	corebb "application/driven/core"
	"application/driven/notifications"
	"application/driven/storage"
	"application/driver/web"
	"strings"

	"github.com/rokwire/core-auth-library-go/v3/authutils"
	"github.com/rokwire/core-auth-library-go/v3/keys"

	"github.com/rokwire/core-auth-library-go/v3/authservice"
	"github.com/rokwire/core-auth-library-go/v3/envloader"
	"github.com/rokwire/core-auth-library-go/v3/sigauth"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

var (
	// Version : version of this executable
	Version string
	// Build : build date of this executable
	Build string
)

func main() {
	if len(Version) == 0 {
		Version = "dev"
	}

	serviceID := "surveys"

	loggerOpts := logs.LoggerOpts{SuppressRequests: logs.NewStandardHealthCheckHTTPRequestProperties(serviceID + "/version")}
	logger := logs.NewLogger(serviceID, &loggerOpts)
	envLoader := envloader.NewEnvLoader(Version, logger)

	envPrefix := strings.ReplaceAll(strings.ToUpper(serviceID), "-", "_") + "_"
	port := envLoader.GetAndLogEnvVar(envPrefix+"PORT", false, false)
	if len(port) == 0 {
		port = "80"
	}

	// mongoDB adapter
	mongoDBAuth := envLoader.GetAndLogEnvVar(envPrefix+"MONGO_AUTH", true, true)
	mongoDBName := envLoader.GetAndLogEnvVar(envPrefix+"MONGO_DATABASE", true, false)
	mongoTimeout := envLoader.GetAndLogEnvVar(envPrefix+"MONGO_TIMEOUT", false, false)
	storageAdapter := storage.NewStorageAdapter(mongoDBAuth, mongoDBName, mongoTimeout, logger)
	err := storageAdapter.Start()
	if err != nil {
		logger.Fatalf("Cannot start the mongoDB adapter: %v", err)
	}

	// Service registration
	baseURL := envLoader.GetAndLogEnvVar(envPrefix+"BASE_URL", true, false)
	coreBBBaseURL := envLoader.GetAndLogEnvVar(envPrefix+"CORE_BB_BASE_URL", true, false)

	authService := authservice.AuthService{
		ServiceID:   serviceID,
		ServiceHost: baseURL,
		FirstParty:  true,
		AuthBaseURL: coreBBBaseURL,
	}

	serviceRegLoader, err := authservice.NewRemoteServiceRegLoader(&authService, []string{"notifications", "calendar"})
	if err != nil {
		logger.Fatalf("Error initializing remote service registration loader: %v", err)
	}

	serviceRegManager, err := authservice.NewServiceRegManager(&authService, serviceRegLoader, !strings.HasPrefix(baseURL, "http://localhost"))
	if err != nil {
		logger.Fatalf("Error initializing service registration manager: %v", err)
	}

	// Service account
	var serviceAccountManager *authservice.ServiceAccountManager

	serviceAccountID := envLoader.GetAndLogEnvVar(envPrefix+"SERVICE_ACCOUNT_ID", false, false)
	privKeyRaw := envLoader.GetAndLogEnvVar(envPrefix+"PRIV_KEY", false, true)
	privKeyRaw = strings.ReplaceAll(privKeyRaw, "\\n", "\n")
	privKey, err := keys.NewPrivKey(keys.PS256, privKeyRaw)
	if err != nil {
		logger.Errorf("Error parsing priv key: %v", err)
	} else if serviceAccountID == "" {
		logger.Errorf("Missing service account id")
	} else {
		signatureAuth, err := sigauth.NewSignatureAuth(privKey, serviceRegManager, false, false)
		if err != nil {
			logger.Fatalf("Error initializing signature auth: %v", err)
		}

		serviceAccountLoader, err := authservice.NewRemoteServiceAccountLoader(&authService, serviceAccountID, signatureAuth)
		if err != nil {
			logger.Fatalf("Error initializing remote service account loader: %v", err)
		}

		serviceAccountManager, err = authservice.NewServiceAccountManager(&authService, serviceAccountLoader)
		if err != nil {
			logger.Fatalf("Error initializing service account manager: %v", err)
		}
	}

	// Notifications adapter
	notificationHost := ""
	notificationsReg, err := serviceRegManager.GetServiceReg("notifications")
	if err != nil {
		logger.Errorf("error finding notifications service reg: %s", err)
	} else {
		notificationHost = notificationsReg.Host
	}
	notificationsAdapter, err := notifications.NewNotificationsAdapter(notificationHost, serviceAccountManager, logger)
	if err != nil {
		logger.Fatalf("Error initializing notifications adapter: %v", err)
	}

	// Calendar adapter
	calendarHost := ""
	calendarReg, err := serviceRegManager.GetServiceReg("calendar")
	if err != nil {
		logger.Errorf("error finding calendar service reg: %s", err)
	} else {
		calendarHost = calendarReg.Host
	}
	calendarAdapter, err := calendar.NewCalendarAdapter(calendarHost, serviceAccountManager, logger)
	if err != nil {
		logger.Fatalf("Error initializing calendar adapter: %v", err)
	}

	//core adapter
	coreAdapter := corebb.NewCoreAdapter(coreBBBaseURL, serviceAccountManager)

	// Application
	application := core.NewApplication(Version, Build, storageAdapter, notificationsAdapter,
		calendarAdapter, coreAdapter, serviceID, logger)
	application.Start()

	// read CORS parameters from stored env config
	var envData *model.EnvConfigData
	var corsAllowedHeaders []string
	var corsAllowedOrigins []string
	config, err := storageAdapter.FindConfig(model.ConfigTypeEnv, authutils.AllApps, authutils.AllOrgs)
	if err != nil {
		logger.Fatal(errors.WrapErrorAction(logutils.ActionFind, model.TypeConfig, nil, err).Error())
	}
	if config != nil {
		envData, err = model.GetConfigData[model.EnvConfigData](*config)
		if err != nil {
			logger.Fatal(errors.WrapErrorAction(logutils.ActionCast, model.TypeEnvConfigData, nil, err).Error())
		}

		corsAllowedHeaders = envData.CORSAllowedHeaders
		corsAllowedOrigins = envData.CORSAllowedOrigins
	}

	// Web adapter
	webAdapter := web.NewWebAdapter(baseURL, port, serviceID, application, serviceRegManager, corsAllowedOrigins, corsAllowedHeaders, logger)
	webAdapter.Start()
}

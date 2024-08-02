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

package web

import (
	"application/core"
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/gorilla/mux"
	"github.com/rokwire/core-auth-library-go/v3/authservice"
	"github.com/rokwire/core-auth-library-go/v3/tokenauth"
	"github.com/rokwire/core-auth-library-go/v3/webauth"

	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"

	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// Adapter entity
type Adapter struct {
	baseURL   string
	port      string
	serviceID string

	auth *Auth

	cachedYamlDoc []byte

	defaultAPIsHandler   DefaultAPIsHandler
	clientAPIsHandler    ClientAPIsHandler
	adminAPIsHandler     AdminAPIsHandler
	analyticsAPIsHandler AnalyticsAPIsHandler
	bbsAPIsHandler       BBsAPIsHandler
	tpsAPIsHandler       TPSAPIsHandler
	systemAPIsHandler    SystemAPIsHandler

	app *core.Application

	corsAllowedOrigins []string
	corsAllowedHeaders []string

	logger *logs.Logger
}

type handlerFunc = func(*logs.Log, *http.Request, *tokenauth.Claims) logs.HTTPResponse

// Start starts the module
func (a Adapter) Start() {

	router := mux.NewRouter().StrictSlash(true)

	// handle apis
	baseRouter := router.PathPrefix("/" + a.serviceID).Subrouter()
	baseRouter.PathPrefix("/doc/ui").Handler(a.serveDocUI())
	baseRouter.HandleFunc("/doc", a.serveDoc)
	baseRouter.HandleFunc("/version", a.wrapFunc(a.defaultAPIsHandler.version, nil)).Methods("GET")

	mainRouter := baseRouter.PathPrefix("/api").Subrouter()

	// Client APIs
	mainRouter.HandleFunc("/surveys", a.wrapFunc(a.clientAPIsHandler.getSurveys, a.auth.client.User)).Methods("GET")
	mainRouter.HandleFunc("/surveys/{id}", a.wrapFunc(a.clientAPIsHandler.getSurvey, a.auth.client.User)).Methods("GET")
	mainRouter.HandleFunc("/surveys", a.wrapFunc(a.clientAPIsHandler.createSurvey, a.auth.client.User)).Methods("POST")
	mainRouter.HandleFunc("/surveys/{id}", a.wrapFunc(a.clientAPIsHandler.updateSurvey, a.auth.client.User)).Methods("PUT")
	mainRouter.HandleFunc("/surveys/{id}", a.wrapFunc(a.clientAPIsHandler.deleteSurvey, a.auth.client.User)).Methods("DELETE")
	mainRouter.HandleFunc("/surveys/{id}/responses", a.wrapFunc(a.clientAPIsHandler.getAllSurveyResponses, a.auth.client.User)).Methods("GET")
	mainRouter.HandleFunc("/survey-responses/{id}", a.wrapFunc(a.clientAPIsHandler.getSurveyResponse, a.auth.client.User)).Methods("GET")
	mainRouter.HandleFunc("/survey-responses", a.wrapFunc(a.clientAPIsHandler.getUserSurveyResponses, a.auth.client.User)).Methods("GET")
	mainRouter.HandleFunc("/survey-responses", a.wrapFunc(a.clientAPIsHandler.createSurveyResponse, a.auth.client.User)).Methods("POST")
	mainRouter.HandleFunc("/survey-responses/{id}", a.wrapFunc(a.clientAPIsHandler.updateSurveyResponse, a.auth.client.User)).Methods("PUT")
	mainRouter.HandleFunc("/survey-responses/{id}", a.wrapFunc(a.clientAPIsHandler.deleteSurveyResponse, a.auth.client.User)).Methods("DELETE")
	mainRouter.HandleFunc("/survey-responses", a.wrapFunc(a.clientAPIsHandler.deleteSurveyResponses, a.auth.client.User)).Methods("DELETE")
	mainRouter.HandleFunc("/survey-alerts", a.wrapFunc(a.clientAPIsHandler.createSurveyAlert, a.auth.client.User)).Methods("POST")
	mainRouter.HandleFunc("/creator/surveys", a.wrapFunc(a.clientAPIsHandler.getCreatorSurveys, a.auth.client.User)).Methods("GET")

	// Admin APIs
	adminRouter := mainRouter.PathPrefix("/admin").Subrouter()
	adminRouter.HandleFunc("/configs/{id}", a.wrapFunc(a.adminAPIsHandler.getConfig, a.auth.admin.Permissions)).Methods("GET")
	adminRouter.HandleFunc("/configs", a.wrapFunc(a.adminAPIsHandler.getConfigs, a.auth.admin.Permissions)).Methods("GET")
	adminRouter.HandleFunc("/configs", a.wrapFunc(a.adminAPIsHandler.createConfig, a.auth.admin.Permissions)).Methods("POST")
	adminRouter.HandleFunc("/configs/{id}", a.wrapFunc(a.adminAPIsHandler.updateConfig, a.auth.admin.Permissions)).Methods("PUT")
	adminRouter.HandleFunc("/configs/{id}", a.wrapFunc(a.adminAPIsHandler.deleteConfig, a.auth.admin.Permissions)).Methods("DELETE")

	adminRouter.HandleFunc("/surveys", a.wrapFunc(a.adminAPIsHandler.getSurveys, a.auth.admin.Permissions)).Methods("GET")
	adminRouter.HandleFunc("/surveys/{id}", a.wrapFunc(a.adminAPIsHandler.getSurvey, a.auth.admin.Permissions)).Methods("GET")
	adminRouter.HandleFunc("/surveys", a.wrapFunc(a.adminAPIsHandler.createSurvey, a.auth.admin.Permissions)).Methods("POST")
	adminRouter.HandleFunc("/surveys/{id}", a.wrapFunc(a.adminAPIsHandler.updateSurvey, a.auth.admin.Permissions)).Methods("PUT")
	adminRouter.HandleFunc("/surveys/{id}", a.wrapFunc(a.adminAPIsHandler.deleteSurvey, a.auth.admin.Permissions)).Methods("DELETE")
	adminRouter.HandleFunc("/surveys/{id}/responses", a.wrapFunc(a.adminAPIsHandler.getAllSurveyResponses, a.auth.admin.User)).Methods("GET")
	adminRouter.HandleFunc("/surveys/{id}/response", a.wrapFunc(a.adminAPIsHandler.getAllSurveysResponses, a.auth.admin.Permissions)).Methods("GET")

	adminRouter.HandleFunc("/alert-contacts", a.wrapFunc(a.adminAPIsHandler.getAlertContacts, a.auth.admin.Permissions)).Methods("GET")
	adminRouter.HandleFunc("/alert-contacts/{id}", a.wrapFunc(a.adminAPIsHandler.getAlertContact, a.auth.admin.Permissions)).Methods("GET")
	adminRouter.HandleFunc("/alert-contacts", a.wrapFunc(a.adminAPIsHandler.createAlertContact, a.auth.admin.Permissions)).Methods("POST")
	adminRouter.HandleFunc("/alert-contacts/{id}", a.wrapFunc(a.adminAPIsHandler.updateAlertContact, a.auth.admin.Permissions)).Methods("PUT")
	adminRouter.HandleFunc("/alert-contacts/{id}", a.wrapFunc(a.adminAPIsHandler.deleteAlertContact, a.auth.admin.Permissions)).Methods("DELETE")

	// Analytics APIs
	analyticsRouter := mainRouter.PathPrefix("/analytics").Subrouter()
	analyticsRouter.HandleFunc("/survey-responses", a.wrapFunc(a.analyticsAPIsHandler.getAnonymousSurveyResponses, a.auth.analytics)).Methods("GET")

	// BB APIs
	// bbsRouter := mainRouter.PathPrefix("/bbs").Subrouter()

	// TPS APIs
	// tpsRouter := mainRouter.PathPrefix("/tps").Subrouter()

	// System APIs
	// systemRouter := mainRouter.PathPrefix("/system").Subrouter()

	var handler http.Handler = router
	if len(a.corsAllowedOrigins) > 0 {
		handler = webauth.SetupCORS(a.corsAllowedOrigins, a.corsAllowedHeaders, router)
	}
	a.logger.Fatalf("Error serving: %v", http.ListenAndServe(":"+a.port, handler))
}

func (a Adapter) serveDoc(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("access-control-allow-origin", "*")

	if a.cachedYamlDoc != nil {
		http.ServeContent(w, r, "", time.Now(), bytes.NewReader([]byte(a.cachedYamlDoc)))
	} else {
		http.ServeFile(w, r, "./driver/web/docs/gen/def.yaml")
	}
}

func (a Adapter) serveDocUI() http.Handler {
	url := fmt.Sprintf("%s/doc", a.baseURL)
	return httpSwagger.Handler(httpSwagger.URL(url))
}

func loadDocsYAML(baseServerURL string) ([]byte, error) {
	data, _ := os.ReadFile("./driver/web/docs/gen/def.yaml")
	yamlMap := yaml.MapSlice{}
	err := yaml.Unmarshal(data, &yamlMap)
	if err != nil {
		return nil, err
	}

	for index, item := range yamlMap {
		if item.Key == "servers" {
			var serverList []interface{}
			if baseServerURL != "" {
				serverList = []interface{}{yaml.MapSlice{yaml.MapItem{Key: "url", Value: baseServerURL}}}
			}

			item.Value = serverList
			yamlMap[index] = item
			break
		}
	}

	yamlDoc, err := yaml.Marshal(&yamlMap)
	if err != nil {
		return nil, err
	}

	return yamlDoc, nil
}

func (a Adapter) wrapFunc(handler handlerFunc, authorization tokenauth.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logObj := a.logger.NewRequestLog(req)

		logObj.RequestReceived()

		var response logs.HTTPResponse
		if authorization != nil {
			responseStatus, claims, err := authorization.Check(req)
			if err != nil {
				logObj.SendHTTPResponse(w, logObj.HTTPResponseErrorAction(logutils.ActionValidate, logutils.TypeRequest, nil, err, responseStatus, true))
				return
			}

			if claims != nil {
				logObj.SetContext("account_id", claims.Subject)
			}
			response = handler(logObj, req, claims)
		} else {
			response = handler(logObj, req, nil)
		}

		logObj.SendHTTPResponse(w, response)
		logObj.RequestComplete()
	}
}

// NewWebAdapter creates new WebAdapter instance
func NewWebAdapter(baseURL string, port string, serviceID string, app *core.Application, serviceRegManager *authservice.ServiceRegManager,
	corsAllowedOrigins []string, corsAllowedHeaders []string, logger *logs.Logger) Adapter {
	yamlDoc, err := loadDocsYAML(baseURL)
	if err != nil {
		logger.Fatalf("error parsing docs yaml - %s", err.Error())
	}

	auth, err := NewAuth(serviceRegManager, app)
	if err != nil {
		logger.Fatalf("error creating auth - %s", err.Error())
	}

	defaultAPIsHandler := NewDefaultAPIsHandler(app)
	clientAPIsHandler := NewClientAPIsHandler(app)
	adminAPIsHandler := NewAdminAPIsHandler(app)
	analyticsAPIsHandler := NewAnalyticsAPIsHandler(app)
	systemAPIsHandler := NewSystemAPIsHandler(app)
	return Adapter{baseURL: baseURL, port: port, serviceID: serviceID, cachedYamlDoc: yamlDoc, auth: auth, defaultAPIsHandler: defaultAPIsHandler,
		clientAPIsHandler: clientAPIsHandler, adminAPIsHandler: adminAPIsHandler, analyticsAPIsHandler: analyticsAPIsHandler, systemAPIsHandler: systemAPIsHandler,
		app: app, corsAllowedOrigins: corsAllowedOrigins, corsAllowedHeaders: corsAllowedHeaders, logger: logger}
}

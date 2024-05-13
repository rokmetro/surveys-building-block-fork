package notifications

import (
	"application/core/model"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/rokwire/core-auth-library-go/v3/authservice"
	"github.com/rokwire/logging-library-go/v2/errors"
	"github.com/rokwire/logging-library-go/v2/logs"
	"github.com/rokwire/logging-library-go/v2/logutils"
)

// Adapter implements the Notifications interface
type Adapter struct {
	baseURL               string
	serviceAccountManager *authservice.ServiceAccountManager

	logger *logs.Logger
}

// NewNotificationsAdapter creates a new Notifications BB adapter instance
func NewNotificationsAdapter(notificationHost string, serviceAccountManager *authservice.ServiceAccountManager, logger *logs.Logger) (*Adapter, error) {
	return &Adapter{baseURL: notificationHost, serviceAccountManager: serviceAccountManager, logger: logger}, nil
}

// SendNotification Sends a direct notification trough Notifications BB
func (a *Adapter) SendNotification(notification model.NotificationMessage) {
	if a.serviceAccountManager == nil {
		a.logger.Error("service account manager is nil")
		return
	}

	go a.logger.Error(fmt.Sprint(a.sendNotification(notification)))
}

// SendNotification sends notification to a user
func (a *Adapter) sendNotification(message model.NotificationMessage) error {
	if len(message.Recipients) == 0 && len(message.RecipientsCriteriaList) == 0 && len(message.RecipientAccountCriteria) == 0 {
		return nil
	}
	url := fmt.Sprintf("%s/api/bbs/message", a.baseURL)

	async := true
	bodyData := map[string]interface{}{
		"async":   async,
		"message": message,
	}
	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionMarshal, logutils.TypeRequestBody, nil, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionCreate, logutils.TypeRequest, nil, err)
	}

	resp, err := a.serviceAccountManager.MakeRequest(req, message.AppID, message.OrgID)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionSend, logutils.TypeRequest, nil, err)
	}

	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Newf("request error response code (%d)", resp.Status)
	}
	if resp.StatusCode != 200 {
		return errors.Newf("request error response code (%d): %s", resp.Status, respBytes)
	}
	return nil
}

// SendMail sends email to a user
func (a *Adapter) SendMail(toEmail string, subject string, body string) {
	if a.serviceAccountManager == nil {
		a.logger.Error("service account manager is nil")
		return
	}

	go a.logger.Error(fmt.Sprint(a.sendMail(toEmail, subject, body)))
}

func (a *Adapter) sendMail(toEmail string, subject string, body string) error {
	if len(toEmail) == 0 {
		return nil
	}
	url := fmt.Sprintf("%s/api/bbs/mail", a.baseURL)

	bodyData := map[string]interface{}{
		"to_mail": toEmail,
		"subject": subject,
		"body":    body,
	}
	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionMarshal, logutils.TypeRequestBody, nil, err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionCreate, logutils.TypeRequest, nil, err)
	}

	resp, err := a.serviceAccountManager.MakeRequest(req, "all", "all")
	if err != nil {
		return errors.WrapErrorAction(logutils.ActionSend, logutils.TypeRequest, nil, err)
	}

	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Newf("request error response code (%d)", resp.Status)
	}
	if resp.StatusCode != 200 {
		return errors.Newf("request error response code (%d): %s", resp.Status, respBytes)
	}
	return nil
}

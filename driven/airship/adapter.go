package airship

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rokwire/rokwire-building-block-sdk-go/utils/logging/logs"
)

const (
	vogueDefaultTagGroup string = "vogue-notifications"

	// SubjectVogue is a subject type for vogue
	SubjectVogue string = "Vogue"
	// TagQuizAll is the tag to use for all quiz notifications
	TagQuizAll string = "quiz.all.all"
)

type m map[string]any

// Adapter is the Airship adapter
type Adapter struct {
	host        string
	bearerToken string

	tagGroup string

	logger *logs.Logger
}

// NewAirshipAdapter creates a new Airship adapter instance
func NewAirshipAdapter(host string, bearerToken string, tagGroup string, logger *logs.Logger) *Adapter {
	if tagGroup == "" {
		tagGroup = vogueDefaultTagGroup
	}
	return &Adapter{host: host, bearerToken: bearerToken, tagGroup: tagGroup, logger: logger}
}

// SendNotification sends a notification to multiple Airship users with the given tags
func (a *Adapter) SendNotification(orgID string, appID string, userIDs []string, title string, body string, data map[string]string, tags []string, excludeTags []string) error {
	url := fmt.Sprintf("%s/api/push", a.host)

	client := &http.Client{
		Timeout: 120 * time.Second,
	}

	//TODO check body for additional urls, localization and additional parameters for notification

	ios := m{
		"alert": m{
			"title": title,
			"body":  body,
		},
	}
	android := m{
		"title": title,
		"alert": body,
	}

	if val, ok := data["url"]; ok {
		actions := m{
			"open": m{
				"type":         "deep_link",
				"content":      val,
				"fallback_url": val,
			},
		}
		ios["actions"] = actions
		android["actions"] = actions
	}

	bodyData := m{
		"device_types": []string{"ios", "android"},
		"audience":     a.getAudience(userIDs, tags, excludeTags),
		"notification": m{
			"ios":     ios,
			"android": android,
		},
	}

	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		a.logger.Errorf("error marshalling airship notification request - %s", err)
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		a.logger.Errorf("error creating airship notification request - %s", err)
		return err
	}

	bearerToken := a.bearerToken
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearerToken))

	a.logger.Infof("sending airship push notification request to audience %v", bodyData["audience"])
	resp, err := client.Do(req)
	if err != nil {
		a.logger.Errorf("error loading airship response data - %s", err)
		return err
	}

	defer resp.Body.Close()

	//TODO save response?
	if resp.StatusCode != 202 {
		a.logger.Errorf("error with airship response code - %d", resp.StatusCode)
		return fmt.Errorf("error with airship response code != 200")
	}

	a.logger.Infof("successfully sent airship push notification request to audience %v", bodyData["audience"])
	return nil
}

// Helpers

func (a *Adapter) getAudience(userIDs []string, tags []string, excludeTags []string) m {
	// Build the audience using the "or" helper for multiple named_user
	var userSelectors []m
	for _, id := range userIDs {
		userSelectors = append(userSelectors, m{"named_user": id})
	}
	audience := a.or(userSelectors...)

	if len(tags) > 0 {
		includeSelectors := make([]m, 0, len(tags))
		for _, t := range tags {
			includeSelectors = append(includeSelectors, m{
				"group": a.tagGroup,
				"tag":   t,
			})
		}
		audience = a.and(
			audience,
			a.or(includeSelectors...),
		)
	}

	if len(excludeTags) > 0 {
		excludeSelectors := make([]m, 0, len(excludeTags))
		for _, t := range excludeTags {
			excludeSelectors = append(excludeSelectors, m{
				"group": a.tagGroup,
				"tag":   t,
			})
		}
		return a.and(
			audience,
			a.not(a.or(excludeSelectors...)),
		)
	}

	return audience
}

func (a *Adapter) or(children ...m) m {
	return m{
		"or": children,
	}
}

func (a *Adapter) and(children ...m) m {
	return m{
		"and": children,
	}
}

func (a *Adapter) not(child m) m {
	return m{
		"not": child,
	}
}

package airship

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

const (
	vogueDefaultTagGroup string = "vogue-notifications"
)

type m map[string]any

// Adapter is the Airship adapter
type Adapter struct {
	host        string
	bearerToken string

	tagGroup string
}

// NewAirshipAdapter creates a new Airship adapter instance
func NewAirshipAdapter(host string, bearerToken string, tagGroup string) *Adapter {
	if tagGroup == "" {
		tagGroup = vogueDefaultTagGroup
	}
	return &Adapter{host: host, bearerToken: bearerToken, tagGroup: tagGroup}
}

// SendNotification sends a notification to an Airship user with the given tags
func (a *Adapter) SendNotification(orgID string, appID string, userID string, title string, body string, data m, tags []string, excludeTags []string) error {
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
		"audience":     a.getAudience(userID, tags, excludeTags),
		"notification": m{
			"ios":     ios,
			"android": android,
		},
	}

	bodyBytes, err := json.Marshal(bodyData)
	if err != nil {
		log.Printf("error marshalling airship notification request - %s", err)
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		log.Printf("error creating airship notification request - %s", err)
		return err
	}

	bearerToken := a.bearerToken
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", bearerToken))

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("error loading airship response data - %s", err)
		return err
	}

	defer resp.Body.Close()

	//TODO save response?
	if resp.StatusCode != 202 {
		log.Printf("error with airship response code - %d", resp.StatusCode)
		return fmt.Errorf("error with airship response code != 200")
	}
	return nil
}

// Helpers

func (a *Adapter) getAudience(userID string, tags []string, excludeTags []string) m {
	audience := m{
		"named_user": userID,
	}

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

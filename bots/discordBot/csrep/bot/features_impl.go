package bot

import (
	"fmt"
)

// NavigateFeature handles navigation actions
type NavigateFeature struct{}

func (f *NavigateFeature) Name() string {
	return "navigate"
}

func (f *NavigateFeature) Execute(session *SessionChrome, args map[string]interface{}) (interface{}, error) {
	url, ok := args["url"].(string)
	if !ok || url == "" {
		return nil, fmt.Errorf("url is required")
	}

	if err := session.NavigateTo(url); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status": "success",
		"url":    url,
	}, nil
}

// ScreenshotFeature captures screenshots
type ScreenshotFeature struct{}

func (f *ScreenshotFeature) Name() string {
	return "screenshot"
}

func (f *ScreenshotFeature) Execute(session *SessionChrome, args map[string]interface{}) (interface{}, error) {
	screenshot, err := session.GetScreenshot()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status":     "success",
		"screenshot": screenshot,
	}, nil
}

// OpenProfileFeature opens Steam profile
type OpenProfileFeature struct{}

func (f *OpenProfileFeature) Name() string {
	return "open-profile"
}

func (f *OpenProfileFeature) Execute(session *SessionChrome, args map[string]interface{}) (interface{}, error) {
	url := fmt.Sprintf("https://steamcommunity.com/id/%s", session.UserID)
	if err := session.NavigateTo(url); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status": "success",
		"url":    url,
	}, nil
}

// StatusFeature returns session status
type StatusFeature struct{}

func (f *StatusFeature) Name() string {
	return "status"
}

func (f *StatusFeature) Execute(session *SessionChrome, args map[string]interface{}) (interface{}, error) {
	return map[string]interface{}{
		"sessionId":  session.ID,
		"userId":     session.UserID,
		"port":       session.Port,
		"active":     session.isActive,
		"isLoggedIn": session.IsLoggedIn,
		"currentUrl": session.GetCurrentURL(),
		"createdAt":  session.CreatedAt,
	}, nil
}

package api

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/shelfarr/shelfarr/internal/db"
)

// GeneralSettingsResponse represents the general application settings
type GeneralSettingsResponse struct {
	InstanceName       string   `json:"instanceName"`
	DefaultLanguage    string   `json:"defaultLanguage"`
	PreferredLanguages []string `json:"preferredLanguages"`
	StartPage          string   `json:"startPage"`
	DateFormat         string   `json:"dateFormat"`
}

// GeneralSettingsRequest represents the request body for updating general settings
type GeneralSettingsRequest struct {
	InstanceName       *string  `json:"instanceName,omitempty"`
	DefaultLanguage    *string  `json:"defaultLanguage,omitempty"`
	PreferredLanguages []string `json:"preferredLanguages,omitempty"`
	StartPage          *string  `json:"startPage,omitempty"`
	DateFormat         *string  `json:"dateFormat,omitempty"`
}

// LanguageOption represents a selectable language
type LanguageOption struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// getGeneralSettings returns the current general settings
func (s *Server) getGeneralSettings(c echo.Context) error {
	settings := GeneralSettingsResponse{
		// Defaults
		InstanceName:       "Bookarr",
		DefaultLanguage:    "en",
		PreferredLanguages: []string{"en"},
		StartPage:          "library",
		DateFormat:         "MMMM d, yyyy",
	}

	// Load settings from database
	var dbSettings []db.Setting
	s.db.Where("key LIKE ?", "general_%").Find(&dbSettings)

	for _, setting := range dbSettings {
		switch setting.Key {
		case "general_instance_name":
			settings.InstanceName = setting.Value
		case "general_default_language":
			settings.DefaultLanguage = setting.Value
		case "general_preferred_languages":
			if setting.Value != "" {
				settings.PreferredLanguages = strings.Split(setting.Value, ",")
			}
		case "general_start_page":
			settings.StartPage = setting.Value
		case "general_date_format":
			settings.DateFormat = setting.Value
		}
	}

	return c.JSON(http.StatusOK, settings)
}

// updateGeneralSettings updates general settings
func (s *Server) updateGeneralSettings(c echo.Context) error {
	var req GeneralSettingsRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
	}

	// Update string settings that are provided
	updates := map[string]*string{
		"general_instance_name":    req.InstanceName,
		"general_default_language": req.DefaultLanguage,
		"general_start_page":       req.StartPage,
		"general_date_format":      req.DateFormat,
	}

	for key, valuePtr := range updates {
		if valuePtr != nil {
			setting := db.Setting{Key: key, Value: *valuePtr}
			s.db.Where("key = ?", key).Assign(setting).FirstOrCreate(&setting)
		}
	}

	// Handle preferred languages array (stored as comma-separated)
	if req.PreferredLanguages != nil {
		value := strings.Join(req.PreferredLanguages, ",")
		setting := db.Setting{Key: "general_preferred_languages", Value: value}
		s.db.Where("key = ?", "general_preferred_languages").Assign(setting).FirstOrCreate(&setting)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Settings updated"})
}

// getAvailableLanguages returns all available language options
func (s *Server) getAvailableLanguages(c echo.Context) error {
	languages := []LanguageOption{
		{Code: "en", Name: "English"},
		{Code: "es", Name: "Spanish"},
		{Code: "fr", Name: "French"},
		{Code: "de", Name: "German"},
		{Code: "it", Name: "Italian"},
		{Code: "pt", Name: "Portuguese"},
		{Code: "nl", Name: "Dutch"},
		{Code: "ru", Name: "Russian"},
		{Code: "ja", Name: "Japanese"},
		{Code: "zh", Name: "Chinese"},
		{Code: "ko", Name: "Korean"},
		{Code: "ar", Name: "Arabic"},
		{Code: "hi", Name: "Hindi"},
		{Code: "pl", Name: "Polish"},
		{Code: "sv", Name: "Swedish"},
		{Code: "da", Name: "Danish"},
		{Code: "no", Name: "Norwegian"},
		{Code: "fi", Name: "Finnish"},
		{Code: "tr", Name: "Turkish"},
		{Code: "cs", Name: "Czech"},
		{Code: "hu", Name: "Hungarian"},
		{Code: "el", Name: "Greek"},
		{Code: "he", Name: "Hebrew"},
		{Code: "th", Name: "Thai"},
		{Code: "vi", Name: "Vietnamese"},
	}

	return c.JSON(http.StatusOK, languages)
}

// GetPreferredLanguages is a helper function to get the user's preferred languages
// Can be called from other handlers that need language filtering
func (s *Server) GetPreferredLanguages() []string {
	var setting db.Setting
	if err := s.db.Where("key = ?", "general_preferred_languages").First(&setting).Error; err != nil {
		return []string{"en"} // Default to English
	}
	if setting.Value == "" {
		return []string{"en"}
	}
	return strings.Split(setting.Value, ",")
}

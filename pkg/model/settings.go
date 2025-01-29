package model

type SettingsContextKey struct{}

type Settings struct {
	LightMode bool `json:"light_mode"`
}

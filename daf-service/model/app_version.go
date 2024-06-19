package model

import "gorm.io/gorm"

type AppVersion struct {
	gorm.Model
	LatestVersion string `json:"latest_version"`
	AndroidLink   string `json:"android_link"`
	IosLink       string `json:"ios_link"`
}

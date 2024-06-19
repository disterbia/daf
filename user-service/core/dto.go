package core

import "time"

type LoginRequest struct {
	IdToken  string `json:"id_token"`
	DeviceID string `json:"device_id"`
	FCMToken string `json:"fcm_token"`
}

type AutoLoginRequest struct {
	Email    string `json:"-"`
	FcmToken string `json:"fcm_token"`
	DeviceId string `json:"device_id"`
}

type LoginResponse struct {
	Jwt string `json:"jwt,omitempty"`
	Err string `json:"err,omitempty"`
}

type UserResponse struct {
	Name         string        `json:"name"`
	Nickname     string        `json:"nickname"`
	Email        string        `json:"email"`
	Birthday     time.Time     `json:"birthday"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	Phone        string        `json:"phone"`
	Gender       uint          `json:"gender"` // true:남 false: 여
	SnsType      uint          `json:"sns_type"`
	ProfileImage ImageResponse `json:"profile_image"`
}

type UserRequest struct {
	ID           uint   `json:"-"`
	Nickname     string `json:"nickname"`
	ProfileImage string `json:"profile_image" example:"base64string"`
}

type ImageResponse struct {
	Url          string `json:"url"`
	ThumbnailUrl string `json:"thumbnail_url"`
}

type AppVersionResponse struct {
	LatestVersion string `json:"latest_version"`
	AndroidLink   string `json:"android_link"`
	IosLink       string `json:"ios_link"`
}

type BasicResponse struct {
	Code string `json:"code"`
}

// // for swagger ////
type SuccessResponse struct {
	Jwt string `json:"jwt"`
}
type ErrorResponse struct {
	Err string `json:"err"`
}

package model

import "go-common/library/time"

// Privilege info.
type Privilege struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Title       string    `json:"title"`
	Explain     string    `json:"explain"`
	Type        int8      `json:"type"`
	Operator    string    `json:"operator"`
	State       int8      `json:"state"`
	Deleted     int8      `json:"deleted"`
	IconURL     string    `json:"icon_url"`
	IconGrayURL string    `json:"icon_gray_url"`
	Order       int64     `json:"order"`
	Ctime       time.Time `json:"ctime"`
	Mtime       time.Time `json:"mtime"`
}

// PrivilegeResources privilege resources.
type PrivilegeResources struct {
	ID       int64     `json:"id"`
	PID      int64     `json:"pid"`
	Link     string    `json:"link"`
	ImageURL string    `json:"image_url"`
	Type     int8      `json:"type"`
	Ctime    time.Time `json:"ctime"`
	Mtime    time.Time `json:"mtime"`
}

// PrivilegeResp  resp.
type PrivilegeResp struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Explain     string `json:"explain"`
	Type        int8   `json:"type"`
	Operator    string `json:"operator"`
	State       int8   `json:"state"`
	IconURL     string `json:"icon_url"`
	IconGrayURL string `json:"icon_gray_url"`
	Order       int64  `json:"order"`
	WebLink     string `json:"web_link"`
	WebImageURL string `json:"web_image_url"`
	AppLink     string `json:"app_link"`
	AppImageURL string `json:"app_image_url"`
}

package domain

import "time"

type Settings struct {
	SiteName          string `json:"site_name"`
	PrimaryColor      string `json:"primary_color"`
	CopyrightName     string `json:"copyright_name"`
	TopBarOpacity     int    `json:"top_bar_opacity"`
	NavigationOpacity int    `json:"navigation_opacity"`
	CardOpacity       int    `json:"card_opacity"`
	MaxUploadKB       int    `json:"max_upload_kb"`
	CaptchaEnabled    bool   `json:"captcha_enabled"`
	CaptchaAlgorithm  string `json:"captcha_algorithm"`
	CaptchaCost       int    `json:"captcha_cost"`
}

// CaptchaSettings is managed independently from general site settings so
// editing the site's appearance cannot overwrite CAPTCHA protection.
type CaptchaSettings struct {
	Enabled   bool   `json:"captcha_enabled"`
	Algorithm string `json:"captcha_algorithm"`
	Cost      int    `json:"captcha_cost"`
}

type Question struct {
	ID            string `json:"id"`
	Nickname      string `json:"nickname"`
	Content       string `json:"content"`
	Answer        string `json:"answer,omitempty"`
	Status        string `json:"status"`
	CreatedAt     string `json:"created_at"`
	AnsweredAt    string `json:"answered_at,omitempty"`
	ImageFilename string `json:"image_filename,omitempty"`
}

// Session is the server-side record for one administrator login. TokenHash is
// deliberately kept only in the data file and is never returned by the API.
type Session struct {
	ID        string    `json:"id"`
	TokenHash string    `json:"token_hash"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

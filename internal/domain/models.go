package domain

type Settings struct {
	SiteName          string `json:"site_name"`
	PrimaryColor      string `json:"primary_color"`
	CopyrightName     string `json:"copyright_name"`
	TopBarOpacity     int    `json:"top_bar_opacity"`
	NavigationOpacity int    `json:"navigation_opacity"`
	CardOpacity       int    `json:"card_opacity"`
	MaxUploadKB       int    `json:"max_upload_kb"`
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

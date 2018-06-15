package main

type activationResp struct {
	DeviceCode string `json:"device_code"`
	UserCode   string `json:"user_code"`
}

type authorizationResp struct {
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	Scope            string `json:"scope"`
	TokenType        string `json:"token_type"`
}

// Base contains common fields for every response.
type Base struct {
	Error        string      `json:"error"`
	ErrorCode    interface{} `json:"error_code"`
	ErrorMessage string      `json:"error_message"`
	FormErrors   struct{}    `json:"form_errors"`
	FromCache    interface{} `json:"from_cache"`
	StatusCode   int64       `json:"status_code"`
}

// Pagination contains paging information for some responses.
type Pagination struct {
	Count       int64 `json:"count"`
	End         int64 `json:"end"`
	HasNext     bool  `json:"has_next"`
	HasPrevious bool  `json:"has_previous"`
	Page        int64 `json:"page"`
	Pages       int64 `json:"pages"`
	PerPage     int64 `json:"per_page"`
	Start       int64 `json:"start"`
}

// Bookmarks contains favorites response.
type Bookmarks struct {
	Base
	Data struct {
		Bookmarks  []Child    `json:"bookmarks"`
		Pagination Pagination `json:"pagination"`
	} `json:"data"`
}

// Folder describes favorites folder.
type Folder struct {
	Created    string `json:"created"`
	ID         int64  `json:"id"`
	ItemsCount int64  `json:"items_count"`
	Title      string `json:"title"`
}

// Folders response.
type Folders struct {
	Base
	Data struct {
		Folders    []Folder   `json:"folders"`
		Pagination Pagination `json:"pagination"`
	} `json:"data"`
}

// Child is a object or container.
type Child struct {
	Channel struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"channel"`
	ChildrenCount int64  `json:"children_count"`
	Country       string `json:"country"`
	Description   string `json:"description"`
	Duration      int64  `json:"duration"`
	Files         []struct {
		Bitrate int64  `json:"bitrate"`
		Format  string `json:"format"`
	} `json:"files"`
	ID   int64 `json:"id"`
	IsHd bool  `json:"is_hd"`
	Mark struct {
		Count int64 `json:"count"`
		Total int64 `json:"total"`
	} `json:"mark"`
	Name         string      `json:"name"`
	OnAir        string      `json:"on_air"`
	Parent       int64       `json:"parent"`
	Rating       int64       `json:"rating"`
	SeriesNum    int64       `json:"series_num"`
	ShortName    string      `json:"short_name"`
	ShortNameEng string      `json:"short_name_eng"`
	Tag          string      `json:"tag"`
	Thumb        string      `json:"thumb"`
	Type         string      `json:"type"`
	WatchStatus  int64       `json:"watch_status"`
	Year         interface{} `json:"year"`
}

// Children is a generic response.
type Children struct {
	Data struct {
		Children   []Child    `json:"children"`
		Pagination Pagination `json:"pagination"`
	} `json:"data"`
	Error        string      `json:"error"`
	ErrorCode    interface{} `json:"error_code"`
	ErrorMessage string      `json:"error_message"`
	FormErrors   struct{}    `json:"form_errors"`
	FromCache    interface{} `json:"from_cache"`
	StatusCode   int64       `json:"status_code"`
}

// StreamURL is a response containing media URL.
type StreamURL struct {
	Base
	Data struct {
		URL string `json:"url"`
	} `json:"data"`
}

// Media response.
type Media struct {
	Base
	Data struct {
		Media      []Child    `json:"media"`
		Pagination Pagination `json:"pagination"`
	} `json:"data"`
}

type Channel struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// Channels response.
type Channels struct {
	Base
	Data []Channel `json:"data"`
}

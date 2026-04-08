package models

type TimelineResponse struct {
	Data struct {
		SearchByRawQuery struct {
			SearchTimeline struct {
				Timeline struct {
					Instructions []Instruction `json:"instructions"`
				} `json:"timeline"`
			} `json:"search_timeline"`
		} `json:"search_by_raw_query"`
	} `json:"data"`
}


type Instruction struct {
	Type string `json:"type"`
	Entries []TimelineEntry `json:"entries"`
	Entry *TimelineEntry `json:"entry"` 
}

type TimelineEntry struct {
	EntryID string `json:"entryId"`
	Content struct {
		CursorType string `json:"cursorType"`
		Value string `json:"value"`

		ItemContent struct {
			TweetResults struct {
				Result TweetData `json:"result"`
			} `json:"tweet_results"`
		} `json:"itemContent"`
	} `json:"content"`
}


type TweetData struct {
	RestID string `json:"rest_id"`
	Core struct {
		UserResults struct {
			Result struct {
				Core struct {
					ScreenName string `json:"screen_name"`
					Name string `json:"name"`
				} `json:"core"`
			} `json:"result"`
		} `json:"user_results"`	
	} `json:"core"`
	Legacy struct {
		FullText string `json:"full_text"`
		CreatedAt string `json:"created_at"`
		RetweetCount int `json:"retweet_count"`
		ReplyCount int `json:"reply_count"`
		FavoriteCount int `json:"favorite_count"`
	} `json:"legacy"`
	Views struct {
		ViewCount int `json:"view_count"`
	} `json:"views"`
}
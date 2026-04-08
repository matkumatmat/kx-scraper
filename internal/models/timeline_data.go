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
	Type    string          `json:"type"`
	Entries []TimelineEntry `json:"entries"`
	Entry   *TimelineEntry  `json:"entry"`
}

type TimelineEntry struct {
	EntryID string `json:"entryId"`
	Content struct {
		CursorType string `json:"cursorType"`
		Value      string `json:"value"`

		ItemContent struct {
			TweetResults struct {
				Result TweetData `json:"result"`
			} `json:"tweet_results"`
		} `json:"itemContent"`
	} `json:"content"`
}

type TweetData struct {
	RestID string `json:"rest_id"`
	Source string `json:"source"`
	Core   struct {
		UserResults struct {
			Result struct {
				RestID         string `json:"rest_id"`
				IsBlueVerified bool   `json:"is_blue_verified"`
				Avatar         struct {
					ImageURL string `json:"image_url"`
				} `json:"avatar"`
				Core struct {
					ScreenName string `json:"screen_name"`
					Name       string `json:"name"`
					CreatedAt  string `json:"created_at"`
				} `json:"core"`
				Legacy struct {
					CreatedAt           string `json:"created_at"`
					Description         string `json:"description"`
					FollowersCount      int    `json:"followers_count"`
					FriendsCount        int    `json:"friends_count"`
					StatusesCount       int    `json:"statuses_count"`
					FavouritesCount     int    `json:"favourites_count"`
					ListedCount         int    `json:"listed_count"`
					MediaCount          int    `json:"media_count"`
					DefaultProfile      bool   `json:"default_profile"`
					DefaultProfileImage bool   `json:"default_profile_image"`
					ProfileBannerURL    string `json:"profile_banner_url"`
				} `json:"legacy"`
				Location struct {
					Location string `json:"location"`
				} `json:"location"`
			} `json:"result"`
		} `json:"user_results"`
	} `json:"core"`
	Legacy struct {
		CreatedAt           string `json:"created_at"`
		FullText            string `json:"full_text"`
		RetweetCount        int    `json:"retweet_count"`
		ReplyCount          int    `json:"reply_count"`
		FavoriteCount       int    `json:"favorite_count"`
		QuoteCount          int    `json:"quote_count"`
		BookmarkCount       int    `json:"bookmark_count"`
		Lang                string `json:"lang"`
		InReplyToScreenName string `json:"in_reply_to_screen_name"`
		InReplyToStatusID   string `json:"in_reply_to_status_id_str"`
		ConversationID      string `json:"conversation_id_str"`
	} `json:"legacy"`
	Views struct {
		Count string `json:"count"`
	} `json:"views"`
}

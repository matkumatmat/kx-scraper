package models

import "fmt"

type TweetRecord struct {
	PostID              string
	UserID              string
	AccCreatedAt        string
	Name                string
	ScreenName          string
	IsVerified          bool
	Description         string
	Location            string
	FollowersCount      int
	FriendsCount        int
	StatusesCount       int
	FavoritesCount      int
	ListedCount         int
	MediaCount          int
	DefaultProfile      bool
	DefaultProfileImage bool
	AvatarURL           string
	BannerURL           string
	FullPost            string
	PostingDate         string
	Language            string
	RetweetCount        int
	ReplyCount          int
	FavoritePostCount   int
	QuoteCount          int
	BookmarkCount       int
	ViewsCount          string
	IsReply             bool
	ReplyToAccount      string
	ReplyToTweetID      string
	ConversationID      string
	SourceDevice        string
	URL                 string
}

func (t TweetRecord) GetHeaders() []string {
	return []string{
		"Post ID", "User ID", "Akun Dibuat", "Nama", "Username", "Verified",
		"Bio_Deskripsi", "Lokasi", "Followers", "Following_Friends", "Total Status",
		"Total Fav Akun", "Listed", "Total Media", "Default Profile",
		"Default Avatar", "Avatar URL", "Banner URL", "Teks Tweet", "Tanggal Posting",
		"Bahasa", "Retweets", "Replies", "Likes", "Quotes", "Bookmarks", "Views",
		"Is Reply", "Reply To", "Reply ID", "Conversation ID", "Source Device", "URL Tweet",
	}
}

func (t TweetRecord) ToRow() []string {
	return []string{
		t.PostID, t.UserID, t.AccCreatedAt, t.Name, t.ScreenName,
		fmt.Sprintf("%t", t.IsVerified), t.Description, t.Location,
		fmt.Sprintf("%d", t.FollowersCount), fmt.Sprintf("%d", t.FriendsCount),
		fmt.Sprintf("%d", t.StatusesCount), fmt.Sprintf("%d", t.FavoritesCount),
		fmt.Sprintf("%d", t.ListedCount), fmt.Sprintf("%d", t.MediaCount),
		fmt.Sprintf("%t", t.DefaultProfile), fmt.Sprintf("%t", t.DefaultProfileImage),
		t.AvatarURL, t.BannerURL, t.FullPost, t.PostingDate, t.Language,
		fmt.Sprintf("%d", t.RetweetCount), fmt.Sprintf("%d", t.ReplyCount),
		fmt.Sprintf("%d", t.FavoritePostCount), fmt.Sprintf("%d", t.QuoteCount),
		fmt.Sprintf("%d", t.BookmarkCount), t.ViewsCount,
		fmt.Sprintf("%t", t.IsReply), t.ReplyToAccount, t.ReplyToTweetID,
		t.ConversationID, t.SourceDevice, t.URL,
	}
}

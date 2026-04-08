package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp" // Wajib buat bersihin HTML tag
	"strings"
	"time"

	"kx-scraper/internal/exporter"
	"kx-scraper/internal/infra/parser/curl"
	"kx-scraper/internal/models"
)

// Perhatikan: parameter keempat sekarang nangkep isRaw
func FetchData(authPool []*curl.ParsedReq, maxPages int, exp exporter.Exporter, isRaw bool) error {
	slog.Info("Memulai engine scraper multi-auth", "max_pages", maxPages, "total_auth", len(authPool), "mode_raw", isRaw)

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
		},
		Timeout: 15 * time.Second,
	}

	totalTweets := 0
	var nextCursor string
	currentAuthIdx := 0

	// Senjata buat hapus HTML tag (misal <a> href... </a>)
	htmlRegex := regexp.MustCompile(`<[^>]*>`)

	for page := 1; page <= maxPages; page++ {
		p := authPool[currentAuthIdx]
		slog.Info("Eksekusi halaman", "page", page, "akun_index", currentAuthIdx)

		if nextCursor != "" {
			p.Variables["cursor"] = nextCursor
		}

		varsBytes, _ := json.Marshal(p.Variables)
		varsEncoded := strings.ReplaceAll(url.QueryEscape(string(varsBytes)), "+", "%20")

		req, err := http.NewRequest("GET", p.BaseURL, nil)
		if err != nil {
			slog.Error("Gagal bikin request", "error", err)
			return err
		}
		req.URL.RawQuery = "variables=" + varsEncoded
		if p.Features != "" {
			req.URL.RawQuery += "&features=" + strings.ReplaceAll(url.QueryEscape(p.Features), "+", "%20")
		}

		for k, v := range p.Headers {
			if strings.ToLower(k) != "accept-encoding" {
				req.Header.Set(k, v)
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			slog.Warn("Koneksi gagal, coba rotasi akun...", "error", err)
			currentAuthIdx = (currentAuthIdx + 1) % len(authPool)
			page--
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != 200 {
			slog.Warn("Akun kena limit atau error, rotasi sekarang", "status", resp.StatusCode, "akun_index", currentAuthIdx)
			currentAuthIdx = (currentAuthIdx + 1) % len(authPool)
			time.Sleep(2 * time.Second)
			page--
			continue
		}

		var xData models.TimelineResponse
		if err := json.Unmarshal(body, &xData); err != nil {
			slog.Error("JSON berantakan", "error", err)
			return err
		}

		pageTweetCount := 0
		newCursorFound := false
		instructions := xData.Data.SearchByRawQuery.SearchTimeline.Timeline.Instructions

		for _, inst := range instructions {
			for _, entry := range inst.Entries {
				tweet := entry.Content.ItemContent.TweetResults.Result
				if tweet.RestID != "" {
					userCore := tweet.Core.UserResults.Result
					userLegacy := userCore.Legacy
					postLegacy := tweet.Legacy

					// LOGIC RAW vs SANITIZED
					finalText := postLegacy.FullText
					finalDesc := userLegacy.Description
					finalSource := tweet.Source

					if !isRaw {
						finalText = strings.ReplaceAll(finalText, "\n", " ")
						finalDesc = strings.ReplaceAll(finalDesc, "\n", " ")
						finalSource = htmlRegex.ReplaceAllString(finalSource, "")
					}

					isReply := postLegacy.InReplyToScreenName != ""

					if exp != nil {
						tweetData := models.TweetRecord{
							PostID:              tweet.RestID,
							UserID:              userCore.RestID,
							AccCreatedAt:        userCore.Core.CreatedAt,
							Name:                userCore.Core.Name,
							ScreenName:          userCore.Core.ScreenName,
							IsVerified:          userCore.IsBlueVerified,
							Description:         finalDesc,
							Location:            userCore.Location.Location,
							FollowersCount:      userLegacy.FollowersCount,
							FriendsCount:        userLegacy.FriendsCount,
							StatusesCount:       userLegacy.StatusesCount,
							FavoritesCount:      userLegacy.FavouritesCount,
							ListedCount:         userLegacy.ListedCount,
							MediaCount:          userLegacy.MediaCount,
							DefaultProfile:      userLegacy.DefaultProfile,
							DefaultProfileImage: userLegacy.DefaultProfileImage,
							AvatarURL:           userCore.Avatar.ImageURL,
							BannerURL:           userLegacy.ProfileBannerURL,
							FullPost:            finalText,
							PostingDate:         postLegacy.CreatedAt,
							Language:            postLegacy.Lang,
							RetweetCount:        postLegacy.RetweetCount,
							ReplyCount:          postLegacy.ReplyCount,
							FavoritePostCount:   postLegacy.FavoriteCount,
							QuoteCount:          postLegacy.QuoteCount,
							BookmarkCount:       postLegacy.BookmarkCount,
							ViewsCount:          tweet.Views.Count,
							IsReply:             isReply,
							ReplyToAccount:      postLegacy.InReplyToScreenName,
							ReplyToTweetID:      postLegacy.InReplyToStatusID,
							ConversationID:      postLegacy.ConversationID,
							SourceDevice:        finalSource,
							URL:                 fmt.Sprintf("https://x.com/%s/status/%s", userCore.Core.ScreenName, tweet.RestID),
						}
						exp.Write(tweetData)
					}

					pageTweetCount++
					totalTweets++
				}

				if entry.Content.CursorType == "Bottom" {
					nextCursor = entry.Content.Value
					newCursorFound = true
				}
			}

			if inst.Entry != nil && inst.Entry.Content.CursorType == "Bottom" {
				nextCursor = inst.Entry.Content.Value
				newCursorFound = true
			}
		}

		slog.Info("Halaman selesai", "page", page, "tweets", pageTweetCount)

		// DEBUG: Lacak cursor biar nggak duplikat
		slog.Info("Cursor debug",
			"page", page,
			"nextCursor", nextCursor,
			"found", newCursorFound)

		if !newCursorFound || nextCursor == "" {
			slog.Warn("Cursor abis, scraping stop")
			break
		}

		currentAuthIdx = (currentAuthIdx + 1) % len(authPool)

		if page < maxPages {
			safeRestTime := 50
			totalAccounts := len(authPool)

			// Ceiling division: pastiin total cycle >= 50 detik
			delay := (safeRestTime + totalAccounts - 1) / totalAccounts

			if delay < 2 {
				delay = 2
			}

			totalCycleTime := delay * totalAccounts
			slog.Info("Jeda antar request (ceiling)", "detik_per_akun", delay, "total_cycle_time", totalCycleTime)
			time.Sleep(time.Duration(delay) * time.Second)
		}
	}

	slog.Info("Scraping kelar bre", "total_tweets", totalTweets)
	return nil
}

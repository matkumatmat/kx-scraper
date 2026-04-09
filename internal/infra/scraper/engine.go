package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"kx-scraper/internal/exporter"
	"kx-scraper/internal/infra/parser/curl"
	"kx-scraper/internal/infra/parser/cursor" // IMPORT SIHIR CTF KITA (Top & Latest)
	"kx-scraper/internal/models"
)

// buildVariables bikin salinan bersih variables per-request.
// JANGAN mutate p.Variables langsung karena itu di-share.
func buildVariables(p *curl.ParsedReq, nextCursor string) map[string]interface{} {
	vars := make(map[string]interface{}, len(p.Variables))
	for k, v := range p.Variables {
		vars[k] = v
	}
	if nextCursor != "" {
		vars["cursor"] = nextCursor
	} else {
		delete(vars, "cursor")
	}
	return vars
}

func FetchData(authPool []*curl.ParsedReq, maxPages int, exp exporter.Exporter, isRaw bool) error {
	// Deteksi otomatis kita lagi di Tab apa
	tabMode := authPool[0].Variables["product"].(string)
	isLatest := tabMode == "Latest"

	slog.Info("Memulai engine scraper multi-auth (HYBRID GOD CURSOR MODE)",
		"max_pages", maxPages,
		"total_auth", len(authPool),
		"mode_raw", isRaw,
		"tab", tabMode,
	)

	client := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     30 * time.Second,
		},
		Timeout: 15 * time.Second,
	}

	totalTweets := 0
	var nextCursor string // Ini bakal diisi forged cursor
	currentAuthIdx := 0

	// Senjata buat hapus HTML tag
	htmlRegex := regexp.MustCompile(`<[^>]*>`)

	for page := 1; page <= maxPages; page++ {
		p := authPool[currentAuthIdx]
		slog.Info("Eksekusi halaman", "page", page, "akun_index", currentAuthIdx)

		// Bikin vars baru pake cursor (kalo ada)
		vars := buildVariables(p, nextCursor)

		varsBytes, _ := json.Marshal(vars)
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
		var candidateCursor string

		instructions := xData.Data.SearchByRawQuery.SearchTimeline.Timeline.Instructions

		for _, inst := range instructions {
			for _, entry := range inst.Entries {
				// Ambil cursor di level entry
				if entry.Content.CursorType == "Bottom" {
					candidateCursor = entry.Content.Value
					newCursorFound = true
				}

				tweet := entry.Content.ItemContent.TweetResults.Result
				if tweet.RestID != "" {
					userCore := tweet.Core.UserResults.Result
					userLegacy := userCore.Legacy
					postLegacy := tweet.Legacy

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
			}

			// Ambil cursor di level instruction
			if inst.Entry != nil && inst.Entry.Content.CursorType == "Bottom" {
				candidateCursor = inst.Entry.Content.Value
				newCursorFound = true
			}
		}

		slog.Info("Halaman selesai", "page", page, "tweets", pageTweetCount)

		if !newCursorFound || candidateCursor == "" {
			slog.Warn("Cursor abis, scraping stop")
			break
		}

		// ---------------------------------------------------------
		// MAGIC BYPASS: FORGE GOD CURSOR SESUAI TAB
		// ---------------------------------------------------------
		var forgedCursor string
		var forgeErr error

		if isLatest {
			forgedCursor, forgeErr = cursor.ForgeGodCursorLatest(candidateCursor)
		} else {
			forgedCursor, forgeErr = cursor.ForgeGodCursorTop(candidateCursor)
		}

		if forgeErr != nil {
			slog.Error("Gagal forge cursor! Jatuh ke cursor asli",
				"error", forgeErr,
				"raw_cursor_failing", candidateCursor)
			nextCursor = candidateCursor // Fallback ke asli kalau gagal ngehack
		} else {
			nextCursor = forgedCursor
			slog.Debug("Berhasil forge GOD CURSOR!", "mode", tabMode)
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

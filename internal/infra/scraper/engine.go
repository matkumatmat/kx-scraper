package scraper

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"kx-scraper/internal/infra/parser/curl"
	"kx-scraper/internal/models"
)

// FetchData mengeksekusi scraping berdasarkan data dari cURL
func FetchData(p *curl.ParsedReq, maxPages int, exportFile string) error {
	slog.Info("Memulai engine scraper", "max_pages", maxPages, "target_export", exportFile)

	// SETUP CSV WRITER
	var csvWriter *csv.Writer
	if exportFile != "" {
		file, err := os.Create(exportFile)
		if err != nil {
			slog.Error("Gagal membuat file CSV", "error", err)
			return fmt.Errorf("gagal bikin file csv: %v", err)
		}
		defer file.Close()

		csvWriter = csv.NewWriter(file)
		defer csvWriter.Flush()

		csvWriter.Write([]string{"Username", "Teks", "Retweets", "Likes", "Views", "URL Tweet"})
		slog.Info("File CSV siap digunakan", "file", exportFile)
	}

	// SETUP HTTP CONNECTION POOLING
	// Ini bikin request lu gak perlu buka-tutup jalur TLS dari nol tiap ganti halaman
	customTransport := &http.Transport{
		MaxIdleConns:        10,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
		DisableKeepAlives:   false, // Wajib false biar koneksi di-reuse
	}
	client := &http.Client{
		Transport: customTransport,
		Timeout:   15 * time.Second,
	}

	totalTweets := 0

	// LOOPING PAGINATION
	for page := 1; page <= maxPages; page++ {
		slog.Info("Mengeksekusi halaman", "page", page)

		varsBytes, err := json.Marshal(p.Variables)
		if err != nil {
			slog.Error("Gagal re-marshal variables", "error", err)
			return err
		}

		varsEncoded := url.QueryEscape(string(varsBytes))
		varsEncoded = strings.ReplaceAll(varsEncoded, "+", "%20")

		featEncoded := ""
		if p.Features != "" {
			featEncoded = url.QueryEscape(p.Features)
			featEncoded = strings.ReplaceAll(featEncoded, "+", "%20")
		}

		req, err := http.NewRequest("GET", p.BaseURL, nil)
		if err != nil {
			slog.Error("Gagal membuat HTTP request", "error", err)
			return err
		}

		rawQ := "variables=" + varsEncoded
		if featEncoded != "" {
			rawQ += "&features=" + featEncoded
		}
		req.URL.RawQuery = rawQ

		slog.Debug("Generated URL", "page", page, "url", req.URL.String())

		for key, val := range p.Headers {
			if strings.ToLower(key) == "accept-encoding" {
				continue
			}
			req.Header.Set(key, val)
		}

		// TEMBAK HTTP PAKE POOLED CLIENT
		resp, err := client.Do(req)
		if err != nil {
			slog.Error("Request gagal", "page", page, "error", err)
			return err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()

		if err != nil {
			slog.Error("Gagal membaca response body", "page", page, "error", err)
			return err
		}

		if resp.StatusCode != 200 {
			slog.Error("HTTP Error dari server X", "status_code", resp.StatusCode, "page", page)
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
		}

		var xData models.TimelineResponse
		err = json.Unmarshal(body, &xData)
		if err != nil {
			slog.Error("Gagal parsing JSON response", "page", page, "error", err)
			return err
		}

		var nextCursor string
		pageTweetCount := 0

		instructions := xData.Data.SearchByRawQuery.SearchTimeline.Timeline.Instructions
		for _, inst := range instructions {
			
			// Skenario 1: Cek di dalam array Entries
			if len(inst.Entries) > 0 {
				for _, entry := range inst.Entries {
					tweet := entry.Content.ItemContent.TweetResults.Result
					if tweet.RestID != "" {
						username := tweet.Core.UserResults.Result.Core.ScreenName
						text := tweet.Legacy.FullText

						cleanFullText := strings.ReplaceAll(text, "\n", " ")

						if csvWriter != nil {
							rt := fmt.Sprintf("%d", tweet.Legacy.RetweetCount)
							fav := fmt.Sprintf("%d", tweet.Legacy.FavoriteCount)
							views := fmt.Sprintf("%v", tweet.Views.ViewCount)
							urlTweet := fmt.Sprintf("https://x.com/%s/status/%s", username, tweet.RestID)
							
							csvWriter.Write([]string{username, cleanFullText, rt, fav, views, urlTweet})
						}

						terminalText := cleanFullText
						if len(terminalText) > 60 {
							terminalText = terminalText[:60] + "..."
						}
						// Cetak progress ke terminal tanpa format slog biar gampang dibaca user
						fmt.Printf("   -> [@%s]: %s\n", username, terminalText)

						pageTweetCount++
						totalTweets++
					}

					if entry.Content.CursorType == "Bottom" {
						nextCursor = entry.Content.Value
						slog.Debug("Bottom cursor ditemukan (Array)", "cursor", nextCursor[:30]+"...")
					}
				}
			}

			// Skenario 2: Cek di luar array
			if inst.Entry != nil {
				if inst.Entry.Content.CursorType == "Bottom" {
					nextCursor = inst.Entry.Content.Value
					slog.Debug("Bottom cursor ditemukan (Luar)", "cursor", nextCursor[:30]+"...")
				}
			}
		}
		
		slog.Info("Selesai memproses halaman", "page", page, "tweet_ditemukan", pageTweetCount)

		if nextCursor == "" {
			slog.Warn("Scraping berhenti: Tidak ada cursor lanjutan dari server")
			break
		}

		p.Variables["cursor"] = nextCursor

		if page < maxPages {
			slog.Info("Menjalankan jeda anti-bot", "duration_sec", 50)
			time.Sleep(50 * time.Second)
		}
	}

	slog.Info("Scraping total selesai", "total_tweet", totalTweets)
	
	if exportFile != "" {
		slog.Info("Data berhasil diexport", "file", exportFile)
	}
	return nil
}
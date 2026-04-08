package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"kx-scraper/internal/infra/parser/curl"
	"kx-scraper/internal/infra/scraper"
	"kx-scraper/internal/logger"
	"kx-scraper/internal/models"
)

// params nyimpen inputan user dari CLI (diambil dari package models)
var params models.PostsSearchParams

// globalFlags nyimpen flag umum
var maxPages int
var exportFile string
var logFormat string

var rootCmd = &cobra.Command{
	Use:   "kx",
	Short: "KX Scraper - CLI tools sakti buat ngeruk data X",
	Long:  `KX Scraper adalah tools CLI super cepat untuk mengambil data dari X (Twitter) secara spesifik menggunakan file cURL intercept.`,
	Example: `  # Cari postingan terbaru soal mbg
  kx posts latest --main "mbg" --pages 2

  # Cari postingan terpopuler dari akun tertentu dengan minimal likes
  kx posts top --from "prabowo" --min-faves 1000 --export hasil.csv`,
	Run: func(cmd *cobra.Command, args []string) {
		// Kalau user cuma ngetik "kx" doang, langsung sodorin menu help lengkap!
		cmd.Help()
	},
}

var postsCmd = &cobra.Command{
	Use:   "posts",
	Short: "Cari berdasarkan postingan (tweets)",
}

var topCmd = &cobra.Command{
	Use:   "top",
	Short: "Scrape postingan tab Top",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScraper("Top")
	},
}

var latestCmd = &cobra.Command{
	Use:   "latest",
	Short: "Scrape postingan tab Latest",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runScraper("Latest")
	},
}

// runScraper bertugas jadi "Controller" yang ngehubungin CLI sama Engine
func runScraper(tab string) error {
	// Setup global logger
	logger.Setup(logFormat)

	// Rakit raw query pake method dari models
	rawQueryStr := params.BuildQuery()
	
	if rawQueryStr == "" {
		fmt.Println("[ERROR] Lu harus masukin minimal 1 parameter pencarian! (misal: --main mbg)")
		os.Exit(1)
	}

	fmt.Printf("Mulai Scraping Tab [%s]\n", tab)
	fmt.Printf("Raw Query: %s\n", rawQueryStr)
	fmt.Printf("Maksimal Halaman: %d\n", maxPages)
	fmt.Println("-------------------------------------------------")

	// Panggil layer Infra buat nge-parse curl.txt
	parsedData, err := curl.ParseFile("curl.txt")
	if err != nil {
		return fmt.Errorf("error parsing curl.txt: %v", err)
	}

	// Inject parameter ke dalem template curl
	parsedData.Variables["rawQuery"] = rawQueryStr
	parsedData.Variables["product"] = tab
	delete(parsedData.Variables, "cursor") // Mulai dari halaman 1

	// Panggil layer Infra buat ngeksekusi nembak API
	err = scraper.FetchData(parsedData, maxPages, exportFile)
	if err != nil {
		return fmt.Errorf("scraping gagal: %v", err)
	}

	return nil
}

func init() {
	// Setup Global Flags
	rootCmd.PersistentFlags().IntVar(&maxPages, "pages", 3, "Maksimal halaman yang mau di-scrape")
	rootCmd.PersistentFlags().StringVar(&exportFile, "export", "", "Nama file CSV untuk export (opsional)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log", "text", "Format log: text atau json")

	// Setup Search Parameters Flags
	postsCmd.PersistentFlags().StringVar(&params.MainPhrase, "main", "", "Main keyword")
	postsCmd.PersistentFlags().StringVar(&params.ExactPhrase, "exact", "", "Exact phrase (tanpa kutip)")
	postsCmd.PersistentFlags().StringVar(&params.AnyPhrase, "any", "", "Any phrase (pisahkan dgn spasi)")
	postsCmd.PersistentFlags().StringVar(&params.ExcludePhrase, "exclude", "", "Exclude phrase (tanpa minus)")
	postsCmd.PersistentFlags().StringVar(&params.HashtagPhrase, "hashtag", "", "Hashtag (tanpa #)")
	postsCmd.PersistentFlags().StringVar(&params.FromAccount, "from", "", "Dari username (tanpa @)")
	postsCmd.PersistentFlags().StringVar(&params.ToAccount, "to", "", "Ke username (tanpa @)")
	postsCmd.PersistentFlags().StringVar(&params.Mentioning, "mention", "", "Mention username (tanpa @)")
	postsCmd.PersistentFlags().IntVar(&params.MinReplies, "min-replies", 0, "Minimal jumlah replies")
	postsCmd.PersistentFlags().IntVar(&params.MinFaves, "min-faves", 0, "Minimal jumlah likes/faves")
	postsCmd.PersistentFlags().IntVar(&params.MinRetweets, "min-retweets", 0, "Minimal jumlah retweets")
	postsCmd.PersistentFlags().StringVar(&params.Lang, "lang", "", "Kode bahasa (contoh: id, en)")
	postsCmd.PersistentFlags().StringVar(&params.Until, "until", "", "Batas akhir tanggal (yyyy-mm-dd)")
	postsCmd.PersistentFlags().StringVar(&params.Since, "since", "", "Batas awal tanggal (yyyy-mm-dd)")

	// Susun hirarki command
	postsCmd.AddCommand(topCmd, latestCmd)
	rootCmd.AddCommand(postsCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
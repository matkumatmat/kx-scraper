package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings" // TAMBAHAN: Ini yang bikin error undefined strings tadi

	"github.com/spf13/cobra"

	"kx-scraper/internal/exporter"
	"kx-scraper/internal/infra/parser/curl"
	"kx-scraper/internal/infra/scraper"
	"kx-scraper/internal/logger"
	"kx-scraper/internal/models"
)

// params nyimpen inputan user dari CLI (diambil dari package models)
var params models.PostsSearchParams

// globalFlags nyimpen flag umum
var maxPages int
var logFormat string
var authDir string
var exportFormats string // Buat nangkep "csv,json,sqlite"
var exportName string    // Buat nangkep nama "mbg-januari"
var rawOutput bool

var rootCmd = &cobra.Command{
	Use:   "kx",
	Short: "KX Scraper - CLI tools sakti buat ngeruk data X",
	Long:  `KX Scraper adalah tools CLI super cepat untuk mengambil data dari X (Twitter) secara spesifik menggunakan file cURL intercept.`,
	Example: `  # Cari postingan terbaru soal mbg
  kx posts latest --main "mbg" --pages 2

  # Cari postingan terpopuler dengan rotasi akun ke 3 format sekaligus
  kx posts top --from "prabowo" --export "csv,json,sqlite" --export-name "hasil_prabowo"`,
	Run: func(cmd *cobra.Command, args []string) {
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
	logger.Setup(logFormat)

	rawQueryStr := params.BuildQuery()
	if rawQueryStr == "" {
		fmt.Println("[ERROR] Lu harus masukin minimal 1 parameter pencarian! (misal: --main mbg)")
		os.Exit(1)
	}

	fmt.Printf("Mulai Scraping Tab [%s]\n", tab)
	fmt.Printf("Raw Query: %s\n", rawQueryStr)
	fmt.Printf("Maksimal Halaman: %d\n", maxPages)
	fmt.Printf("Folder Auth: %s\n", authDir)
	fmt.Println("-------------------------------------------------")

	// BACA SEMUA CURL DI DALEM FOLDER AUTH
	var authPool []*curl.ParsedReq

	entries, err := os.ReadDir(authDir)
	if err != nil {
		return fmt.Errorf("gagal ngebaca folder auth '%s': %v", authDir, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			curlPath := filepath.Join(authDir, entry.Name(), "curl.txt")

			if _, err := os.Stat(curlPath); err == nil {
				parsedData, err := curl.ParseFile(curlPath)
				if err != nil {
					fmt.Printf("[WARNING] Gagal parse %s: %v\n", curlPath, err)
					continue
				}

				parsedData.Variables["rawQuery"] = rawQueryStr
				parsedData.Variables["product"] = tab
				delete(parsedData.Variables, "cursor")

				authPool = append(authPool, parsedData)
				fmt.Printf("[+] Berhasil load KTP dari: %s\n", entry.Name())
			}
		}
	}

	if len(authPool) == 0 {
		fmt.Printf("[ERROR] Gak ada satupun auth curl.txt yang valid di folder %s\n", authDir)
		os.Exit(1)
	}

	fmt.Printf("Total KTP siap tempur: %d akun\n", len(authPool))
	fmt.Println("-------------------------------------------------")

	// =========================================================
	// SETUP MULTI-EXPORTER
	// =========================================================
	var activeExporters []exporter.Exporter

	if exportFormats != "" {
		formats := strings.Split(exportFormats, ",")
		baseName := exportName
		if baseName == "" {
			baseName = "hasil_scrape" // Default kalau ga diisi
		}

		for _, format := range formats {
			format = strings.TrimSpace(strings.ToLower(format))
			fileName := fmt.Sprintf("%s.%s", baseName, format)

			switch format {
			case "csv":
				if exp, err := exporter.NewCSV(fileName); err == nil {
					activeExporters = append(activeExporters, exp)
					fmt.Println("[+] Exporter Aktif: CSV ->", fileName)
				}
			case "json":
				if exp, err := exporter.NewJSON(fileName); err == nil {
					activeExporters = append(activeExporters, exp)
					fmt.Println("[+] Exporter Aktif: JSON ->", fileName)
				}
			case "sqlite":
				if exp, err := exporter.NewSQLite(fileName); err == nil {
					activeExporters = append(activeExporters, exp)
					fmt.Println("[+] Exporter Aktif: SQLite ->", fileName)
				}
			default:
				fmt.Printf("[!] Format '%s' ga didukung, di-skip.\n", format)
			}
		}
	}

	// Bungkus semua exporter aktif pake Si Mandor
	var finalExporter exporter.Exporter
	if len(activeExporters) > 0 {
		mandor := exporter.NewMulti(activeExporters...)
		defer mandor.Close()
		finalExporter = mandor
	}

	// Lempar Si Mandor ke Engine, kasih tau dia minta RAW atau Nggak
	err = scraper.FetchData(authPool, maxPages, finalExporter, rawOutput)
	if err != nil {
		return fmt.Errorf("scraping gagal: %v", err)
	}

	return nil
}

func init() {
	// Setup Global Flags (udah dibersihin biar ga ada flag duplicate)
	rootCmd.PersistentFlags().IntVar(&maxPages, "pages", 3, "Maksimal halaman yang mau di-scrape")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log", "text", "Format log: text atau json")
	rootCmd.PersistentFlags().StringVar(&authDir, "auth-dir", "./auth", "Folder tempat nyimpen multi-account (default: ./auth)")
	rootCmd.PersistentFlags().StringVar(&exportFormats, "export", "", "Format export dipisah koma (contoh: csv,json,sqlite)")
	rootCmd.PersistentFlags().StringVar(&exportName, "export-name", "hasil_scrape", "Nama awalan file (tanpa ekstensi)")
	rootCmd.PersistentFlags().BoolVar(&rawOutput, "raw", false, "Output mentah tanpa dibersihkan (pertahankan newline & HTML tag)")

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

	postsCmd.AddCommand(topCmd, latestCmd)
	rootCmd.AddCommand(postsCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

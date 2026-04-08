# 🛠 CLI Search Tool Structure (Cobra)

## 📂 Command Tree
```
mycli
├── posts
│   ├── top      # Pencarian postingan terpopuler
│   └── latest   # Pencarian postingan terbaru
└── user         # Pencarian berdasarkan user
```

## 📦 Parameter Mapping (`SearchParams` Struct)
Semua parameter di bawah ini sudah di-definisi sebagai **flag** dan otomatis ter-bind ke struct berikut:

```go
type SearchParams struct {
    MainPhrase    string // Kata kunci utama
    ExactPhrase   string // Exact phrase (double quote)
    AnyPhrase     string // Any phrase (apit dengan ())
    ExcludePhrase string // Exclude (prefix -)
    HashtagPhrase string // Hashtag (apit dengan ())
    FromAccount   string // Filter pengirim (from:xxx)
    ToAccount     string // Filter penerima (to:yyy)
    Mentioning    string // Filter mention (@aaa)
    MinReplies    int    // Minimum replies
    MinFaves      int    // Minimum faves
    MinRetweets   int    // Minimum retweets
    Lang          string // Kode bahasa
    Until         string // Batas akhir (yyyy-mm-dd)
    Since         string // Batas awal (yyyy-mm-dd)
}
```

## 🐍 Implementasi Cobra (Go)
```go
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var params SearchParams

// Helper buat nge-bind flag ke command
func addSearchFlags(cmd *cobra.Command) {
	f := cmd.Flags()
	f.StringVar(&params.MainPhrase, "main", "", "Main phrase")
	f.StringVar(&params.ExactPhrase, "exact", "", `Exact phrase (contoh: "hour work")`)
	f.StringVar(&params.AnyPhrase, "any", "", `Any phrase (contoh: (osarah))`)
	f.StringVar(&params.ExcludePhrase, "exclude", "", `Exclude phrase (prefix -, contoh: -tidak)`)
	f.StringVar(&params.HashtagPhrase, "hashtag", "", `Hashtag (contoh: (#mbg))`)
	f.StringVar(&params.FromAccount, "from", "", "Filter username pengirim")
	f.StringVar(&params.ToAccount, "to", "", "Filter username penerima")
	f.StringVar(&params.Mentioning, "mention", "", "Filter username mention")
	f.IntVar(&params.MinReplies, "min-replies", 0, "Minimum replies")
	f.IntVar(&params.MinFaves, "min-faves", 0, "Minimum faves")
	f.IntVar(&params.MinRetweets, "min-retweets", 0, "Minimum retweets")
	f.StringVar(&params.Lang, "lang", "", "Kode bahasa (id, en, jp)")
	f.StringVar(&params.Until, "until", "", "Until date (yyyy-mm-dd)")
	f.StringVar(&params.Since, "since", "", "Since date (yyyy-mm-dd)")
}

var rootCmd = &cobra.Command{
	Use:   "mycli",
	Short: "CLI tool untuk pencarian posts & user",
}

var postsCmd = &cobra.Command{
	Use:   "posts",
	Short: "Cari berdasarkan postingan",
}

var topCmd = &cobra.Command{
	Use:   "top",
	Short: "Cari postingan terpopuler",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSearch("top")
	},
}

var latestCmd = &cobra.Command{
	Use:   "latest",
	Short: "Cari postingan terbaru",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSearch("latest")
	},
}

var userCmd = &cobra.Command{
	Use:   "user",
	Short: "Cari berdasarkan user",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSearch("user")
	},
}

func runSearch(mode string) error {
	// 🔧 Di sini logika build query / request API kamu
	fmt.Printf("🚀 Running [%s] search with params:\n", mode)
	fmt.Printf("%+v\n", params)
	return nil
}

func init() {
	// Bind flag ke posts (otomatis diwariskan ke subcommand top & latest)
	// Pake PersistentFlags() biar subcommand ikut dapet
	postsCmd.PersistentFlags().AddFlagSet(userCmd.Flags()) // Trick: kita pake flags dari helper
	// Biar rapi, langsung bind manual ke postsCmd
	postsCmd.PersistentFlags().StringVar(&params.MainPhrase, "main", "", "Main phrase")
	postsCmd.PersistentFlags().StringVar(&params.ExactPhrase, "exact", "", "Exact phrase")
	postsCmd.PersistentFlags().StringVar(&params.AnyPhrase, "any", "", "Any phrase")
	postsCmd.PersistentFlags().StringVar(&params.ExcludePhrase, "exclude", "", "Exclude phrase")
	postsCmd.PersistentFlags().StringVar(&params.HashtagPhrase, "hashtag", "", "Hashtag")
	postsCmd.PersistentFlags().StringVar(&params.FromAccount, "from", "", "From account")
	postsCmd.PersistentFlags().StringVar(&params.ToAccount, "to", "", "To account")
	postsCmd.PersistentFlags().StringVar(&params.Mentioning, "mention", "", "Mentioning")
	postsCmd.PersistentFlags().IntVar(&params.MinReplies, "min-replies", 0, "Min replies")
	postsCmd.PersistentFlags().IntVar(&params.MinFaves, "min-faves", 0, "Min faves")
	postsCmd.PersistentFlags().IntVar(&params.MinRetweets, "min-retweets", 0, "Min retweets")
	postsCmd.PersistentFlags().StringVar(&params.Lang, "lang", "", "Language")
	postsCmd.PersistentFlags().StringVar(&params.Until, "until", "", "Until date")
	postsCmd.PersistentFlags().StringVar(&params.Since, "since", "", "Since date")

	// Bind flag ke user command (standalone)
	userCmd.Flags().AddFlagSet(postsCmd.PersistentFlags())

	// Susun tree command
	postsCmd.AddCommand(topCmd, latestCmd)
	rootCmd.AddCommand(postsCmd, userCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
```

## 🖥️ Contoh Penggunaan di Terminal
```bash
# 1️⃣ Posts Top
mycli posts top \
  --main sarah \
  --exact "saraho" \
  --any "(osarah)" \
  --exclude "-tidak" \
  --hashtag "(#mbg)" \
  --from xxx \
  --to yyy \
  --mention aaa \
  --min-replies 10 \
  --min-faves 11 \
  --min-retweets 12 \
  --lang id \
  --until 2025-02-02 \
  --since 2025-01-01

# 2️⃣ Posts Latest
mycli posts latest --main sarah --lang id --min-faves 50

# 3️⃣ User Search
mycli user --from xxx --lang en --since 2025-01-01
```

## ⚠️ Catatan Penting
1. **Shell Escaping**: Tanda `()` di bash bisa ditafsirkan sebagai subshell. Selalu **apit dengan quote** saat run CLI: `"()"` atau `''`
2. **Flag Inheritance**: `postsCmd.PersistentFlags()` otomatis tersedia di `top` & `latest`. `userCmd` pake `Flags()` biasa karena nggak punya subcommand.
3. **Build Query**: Di dalam `runSearch()`, kamu tinggal rakit string `params` jadi format raw query awal kamu (`sarah "saraho" ...`) atau langsung pass ke HTTP/API client.

Mau aku buatin helper function buat **auto-rakit string query** persis kayak format mentah kamu? (`sarah "saraho" (osarah) ...`) Biar tinggal `fmt.Sprintf()` di `runSearch()`? 🛠️
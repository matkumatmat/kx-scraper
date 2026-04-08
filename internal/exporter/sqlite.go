package exporter

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type SQLite struct {
	db          *sql.DB
	tableWrote  bool
	insertQuery string
}

func NewSQLite(filename string) (*SQLite, error) {
	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		return nil, err
	}
	return &SQLite{db: db}, nil
}

func (s *SQLite) Write(record Exportable) error {
	headers := record.GetHeaders()

	// Bikin tabel otomatis di baris pertama
	if !s.tableWrote {
		var cols []string
		for _, h := range headers {
			// Bersihin spasi jadi underscore
			cleanName := strings.ReplaceAll(strings.ToLower(h), " ", "_")
			// Bersihin garis miring (/) jadi underscore biar SQLite gak crash!
			cleanName = strings.ReplaceAll(cleanName, "/", "_")

			// Bungkus nama kolom pake kutip ganda ("nama_kolom") biar super aman
			cols = append(cols, fmt.Sprintf(`"%s" TEXT`, cleanName))
		}

		createTableSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS scraped_data (%s);", strings.Join(cols, ", "))
		if _, err := s.db.Exec(createTableSQL); err != nil {
			return fmt.Errorf("gagal bikin tabel sqlite: %v", err)
		}

		// Siapin query insert
		placeholders := make([]string, len(headers))
		for i := range placeholders {
			placeholders[i] = "?"
		}
		s.insertQuery = fmt.Sprintf("INSERT INTO scraped_data VALUES (%s)", strings.Join(placeholders, ", "))

		s.tableWrote = true
	}

	// Masukin datanya
	rows := record.ToRow()
	args := make([]interface{}, len(rows))
	for i, v := range rows {
		args[i] = v
	}

	_, err := s.db.Exec(s.insertQuery, args...)
	return err
}

func (s *SQLite) Close() error {
	return s.db.Close()
}

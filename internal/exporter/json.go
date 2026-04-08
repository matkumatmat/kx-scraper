package exporter

import (
	"encoding/json"
	"fmt"
	"os"
)

// ==========================================
// JSON EXPORTER (Streaming Array Mode)
// ==========================================

type JSON struct {
	file     *os.File
	isFirst  bool
	filePath string
}

// NewJSON bikin exporter JSON yang nulis langsung ke file (streaming)
// Format: JSON array [{...}, {...}, ...] yang di-append per record
// ⚠️  Kalau crash di tengah, file belum valid JSON (kurang ]) — pakai RepairJSONFile
func NewJSON(filePath string) (*JSON, error) {
	// Buka file: create if not exist, write-only, append to end
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("buka file JSON '%s': %w", filePath, err)
	}

	// Cek apakah file baru/empty
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("cek info file JSON '%s': %w", filePath, err)
	}

	isFirst := info.Size() == 0

	// Kalau file baru, tulis opening bracket "["
	if isFirst {
		if _, err := file.Write([]byte("[")); err != nil {
			file.Close()
			return nil, fmt.Errorf("tulis opening bracket JSON: %w", err)
		}
	}

	return &JSON{
		file:     file,
		isFirst:  isFirst,
		filePath: filePath,
	}, nil
}

// Write nulis satu record ke file JSON sebagai element array
// Format: {"field":"value"} atau ,{"field":"value"} tergantung posisi
func (j *JSON) Write(record Exportable) error {
	// Encode record ke JSON bytes
	data, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal record ke JSON: %w", err)
	}

	// Kalau bukan record pertama, tambah koma pemisah di depan
	if !j.isFirst {
		if _, err := j.file.Write([]byte(",")); err != nil {
			return fmt.Errorf("tulis koma pemisah JSON: %w", err)
		}
	}
	j.isFirst = false

	// Tulis record + newline biar readable (opsional, tapi enak dilihat)
	if _, err := j.file.Write(data); err != nil {
		return fmt.Errorf("tulis record JSON: %w", err)
	}
	if _, err := j.file.Write([]byte("\n")); err != nil {
		return fmt.Errorf("tulis newline JSON: %w", err)
	}

	// Flush ke disk biar aman kalau crash
	if err := j.file.Sync(); err != nil {
		return fmt.Errorf("flush JSON ke disk: %w", err)
	}

	return nil
}

// Close nutup array JSON dan file
// ⚠️  Panggil ini setelah semua record selesai ditulis biar file valid JSON
func (j *JSON) Close() error {
	// Tulis closing bracket "]"
	if _, err := j.file.Write([]byte("]")); err != nil {
		return fmt.Errorf("tulis closing bracket JSON: %w", err)
	}

	// Flush dan tutup file
	if err := j.file.Sync(); err != nil {
		return fmt.Errorf("flush final JSON: %w", err)
	}
	return j.file.Close()
}

// ==========================================
// REPAIR UTILITY (Kalau Crash di Tengah)
// ==========================================

// RepairJSONFile nambahin closing bracket "]" ke file JSON yang belum selesai
// Pakai ini kalau scraper crash dan file JSON-nya belum valid
func RepairJSONFile(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("buka file JSON '%s' untuk repair: %w", filePath, err)
	}
	defer file.Close()

	// Tambah closing bracket
	if _, err := file.Write([]byte("]")); err != nil {
		return fmt.Errorf("tulis closing bracket repair: %w", err)
	}

	return nil
}

package exporter

import (
	"encoding/csv"
	"os"
)

type CSV struct {
	file        *os.File
	writer      *csv.Writer
	headerWrote bool
}

// NewCSV bikin instance exporter CSV
func NewCSV(filename string) (*CSV, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	return &CSV{
		file:   file,
		writer: csv.NewWriter(file),
	}, nil
}

// Write otomatis nulis header di baris pertama, lanjut isinya
func (c *CSV) Write(record Exportable) error {
	if !c.headerWrote {
		c.writer.Write(record.GetHeaders())
		c.headerWrote = true
	}
	return c.writer.Write(record.ToRow())
}

// Close wajib dipanggil biar file kesimpen sempurna
func (c *CSV) Close() error {
	c.writer.Flush()
	return c.file.Close()
}
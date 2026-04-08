package exporter

import (
	"encoding/json"
	"os"
)

type JSON struct {
	file    *os.File
	records []Exportable // Kita tampung dulu di memory
}

func NewJSON(filename string) (*JSON, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	return &JSON{
		file:    file,
		records: []Exportable{},
	}, nil
}

func (j *JSON) Write(record Exportable) error {
	// Masukin ke array dulu biar formatnya jadi JSON Array yang valid [{},{}]
	j.records = append(j.records, record)
	return nil
}

func (j *JSON) Close() error {
	// Pas program beres, baru kita dump semua array-nya ke file
	encoder := json.NewEncoder(j.file)
	encoder.SetIndent("", "  ") // Biar JSON-nya rapi (pretty print)
	err := encoder.Encode(j.records)
	if err != nil {
		return err
	}
	return j.file.Close()
}
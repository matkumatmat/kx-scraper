package exporter

// Exportable adalah syarat mutlak buat semua data
type Exportable interface {
	GetHeaders() []string
	ToRow() []string
}

// Exporter adalah mesin penyimpannya
type Exporter interface {
	Write(record Exportable) error
	Close() error
}

// ==========================================
// MULTI EXPORTER (Si Mandor)
// ==========================================

type Multi struct {
	exporters []Exporter
}

// NewMulti bikin mandor yang ngawasin banyak exporter sekaligus
func NewMulti(exps ...Exporter) *Multi {
	return &Multi{exporters: exps}
}

// Write nyuruh semua exporter buat nulis data yang sama
func (m *Multi) Write(record Exportable) error {
	for _, exp := range m.exporters {
		if err := exp.Write(record); err != nil {
			return err
		}
	}
	return nil
}

// Close nyuruh semua exporter nutup filenya
func (m *Multi) Close() error {
	for _, exp := range m.exporters {
		exp.Close()
	}
	return nil
}
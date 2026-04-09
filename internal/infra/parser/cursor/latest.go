package cursor

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"
)

// CursorDataLatest menyimpan struktur ramping untuk Tab Latest
type CursorDataLatest struct {
	F1_1 int64 // Max Tweet ID
	F1_2 int64 // Anchor Tweet ID (Pointer Scroll)
	F2   int32 // Direction
	F3   int32 // Unknown Flag
	F4   int32 // PAGE COUNTER
	F5   int64 // Timestamp Query
	F6   int64 // Offset
}

// DecodeLatest membaca cursor Tab Latest
func DecodeLatest(b64Str string) (*CursorDataLatest, error) {
	s := strings.ReplaceAll(b64Str, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	if pad := (4 - len(s)%4) % 4; pad != 4 {
		s += strings.Repeat("=", pad)
	}

	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("decode base64: %w", err)
	}

	pos := 0
	rb := func() (byte, bool) {
		if pos >= len(raw) {
			return 0, false
		}
		v := raw[pos]
		pos++
		return v, true
	}
	ri32 := func() int32 {
		v := int32(binary.BigEndian.Uint32(raw[pos : pos+4]))
		pos += 4
		return v
	}
	ri64 := func() int64 {
		v := int64(binary.BigEndian.Uint64(raw[pos : pos+8]))
		pos += 8
		return v
	}

	var ps func() map[uint16]interface{}
	ps = func() map[uint16]interface{} {
		fields := make(map[uint16]interface{})
	loop:
		for pos < len(raw) {
			ft, ok := rb()
			if !ok || ft == 0 {
				break
			}
			fid := binary.BigEndian.Uint16(raw[pos : pos+2])
			pos += 2
			switch ft {
			case 8:
				fields[fid] = ri32()
			case 10:
				fields[fid] = ri64()
			case 12:
				fields[fid] = ps()
			default:
				// Skip string/binary logic karena di latest ga butuh
				break loop
			}
		}
		return fields
	}

	outer := ps()

	// Di Latest, struct utamanya ngumpet di Field 3 (atau kadang 2, jadi kita cari dinamis aja)
	var inner map[uint16]interface{}
	for _, v := range outer {
		if m, isMap := v.(map[uint16]interface{}); isMap {
			// Ciri khas: Punya Field 4 (int32) dan Field 5 (int64)
			if _, has4 := m[4].(int32); has4 {
				if _, has5 := m[5].(int64); has5 {
					inner = m
					break
				}
			}
		}
	}

	if inner == nil {
		return nil, fmt.Errorf("format outer struct Latest tidak sesuai")
	}

	f1, ok := inner[1].(map[uint16]interface{})
	if !ok {
		return nil, fmt.Errorf("field 1 (anchor pointers) tidak ditemukan")
	}

	return &CursorDataLatest{
		F1_1: f1[1].(int64),
		F1_2: f1[2].(int64),
		F2:   inner[2].(int32),
		F3:   inner[3].(int32),
		F4:   inner[4].(int32),
		F5:   inner[5].(int64),
		F6:   inner[6].(int64),
	}, nil
}

// EncodeLatest membungkus ulang Cursor Tab Latest
func EncodeLatest(d *CursorDataLatest) string {
	wi32 := func(v int32) []byte { buf := make([]byte, 4); binary.BigEndian.PutUint32(buf, uint32(v)); return buf }
	wi64 := func(v int64) []byte { buf := make([]byte, 8); binary.BigEndian.PutUint64(buf, uint64(v)); return buf }
	wf := func(t byte, fid uint16, val []byte) []byte {
		buf := make([]byte, 1+2+len(val))
		buf[0] = t
		binary.BigEndian.PutUint16(buf[1:3], fid)
		copy(buf[3:], val)
		return buf
	}
	wstop := []byte{0x00}

	f1Inner := append(wf(10, 1, wi64(d.F1_1)), wf(10, 2, wi64(d.F1_2))...)
	f1Inner = append(f1Inner, wstop...)

	inner := append(wf(12, 1, f1Inner), wf(8, 2, wi32(d.F2))...)
	inner = append(inner, wf(8, 3, wi32(d.F3))...)
	inner = append(inner, wf(8, 4, wi32(d.F4))...)
	inner = append(inner, wf(10, 5, wi64(d.F5))...)
	inner = append(inner, wf(10, 6, wi64(d.F6))...)
	inner = append(inner, wstop...)

	// Bungkus ke Outer Field 3
	outer := append(wf(12, 3, inner), wstop...)

	result := base64.StdEncoding.EncodeToString(outer)
	return strings.TrimRight(strings.ReplaceAll(strings.ReplaceAll(result, "+", "-"), "/", "_"), "=")
}

// ForgeGodCursorLatest memanipulasi Page Counter Latest agar bypass limit
func ForgeGodCursorLatest(serverCursor string) (string, error) {
	d, err := DecodeLatest(serverCursor)
	if err != nil {
		return "", err
	}
	// d.F4 = 0 // Exploit: Amnesia Page Counter untuk Tab Latest
	return EncodeLatest(d), nil
}

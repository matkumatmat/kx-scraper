package cursor

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"
)

const twitterEpoch = 1288834974657

// CursorData nyimpen hasil bedah cursor dari server
type CursorData struct {
	F2      int64
	F3      int64
	F4      int32
	Magic   []byte
	CurType uint16
	MaxCnt  uint32
	SessID  []byte
	Flags   []byte
	Records []uint64
	F6      int32
	F7      int32
	Anchor  int64
}

// Decode ngebongkar URL-safe base64 cursor dari Twitter
func Decode(b64Str string) (*CursorData, error) {
	// Benerin format base64 url-safe
	s := strings.ReplaceAll(b64Str, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	pad := (4 - len(s)%4) % 4
	if pad != 4 {
		s += strings.Repeat("=", pad)
	}

	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, fmt.Errorf("gagal decode base64: %w", err)
	}

	pos := 0

	// Helper buat baca byte
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
	rstring := func() []byte {
		l := int(binary.BigEndian.Uint32(raw[pos : pos+4]))
		pos += 4
		v := raw[pos : pos+l]
		pos += l
		return v
	}

	// Recursive parser Thrift
	var ps func() map[uint16]interface{}
	// ...existing code...
	ps = func() map[uint16]interface{} {
		fields := make(map[uint16]interface{})
	Loop:
		for pos < len(raw) {
			ft, ok := rb()
			if !ok || ft == 0 {
				break
			}
			fid := binary.BigEndian.Uint16(raw[pos : pos+2])
			pos += 2

			switch ft {
			case 8: // i32
				fields[fid] = ri32()
			case 10: // i64
				fields[fid] = ri64()
			case 11: // string/binary
				fields[fid] = rstring()
			case 12: // struct
				fields[fid] = ps()
			default:
				// Kalau nemu tipe yang gak didukung, stop aja biar gak panic
				break Loop
			}
		}
		return fields
	}
	// ...existing code...

	outer := ps()

	// Cek apakah parsing berhasil nangkep Field#2
	innerRaw, ok := outer[2].(map[uint16]interface{})
	if !ok {
		return nil, fmt.Errorf("format outer struct gak sesuai")
	}

	f5bRaw, ok := innerRaw[5].([]byte)
	if !ok {
		return nil, fmt.Errorf("field 5 (nested base64) gak ketemu")
	}

	// Decode f5 (nested binary)
	f5bStr := string(f5bRaw)
	f5pad := (4 - len(f5bStr)%4) % 4
	if f5pad != 4 {
		f5bStr += strings.Repeat("=", f5pad)
	}

	f5, err := base64.StdEncoding.DecodeString(f5bStr)
	if err != nil {
		return nil, fmt.Errorf("gagal decode nested base64: %w", err)
	}

	rc := binary.BigEndian.Uint16(f5[18:20])
	recs := make([]uint64, rc)
	for i := 0; i < int(rc); i++ {
		offset := 20 + i*8
		recs[i] = binary.BigEndian.Uint64(f5[offset : offset+8])
	}

	// Ambil field 8 (Anchor)
	field8, ok := innerRaw[8].(map[uint16]interface{})
	var anchor int64
	if ok {
		if val, exists := field8[1].(int64); exists {
			anchor = val
		}
	}

	return &CursorData{
		F2:      innerRaw[2].(int64),
		F3:      innerRaw[3].(int64),
		F4:      innerRaw[4].(int32),
		Magic:   f5[0:2],
		CurType: binary.BigEndian.Uint16(f5[2:4]),
		MaxCnt:  binary.BigEndian.Uint32(f5[4:8]),
		SessID:  f5[8:16],
		Flags:   f5[16:18],
		Records: recs,
		F6:      innerRaw[6].(int32),
		F7:      innerRaw[7].(int32),
		Anchor:  anchor,
	}, nil
}

// Encode ngebungkus balik CursorData jadi string URL-safe Base64
func Encode(d *CursorData) string {
	f5 := make([]byte, 0)
	f5 = append(f5, d.Magic...)

	ct := make([]byte, 2)
	binary.BigEndian.PutUint16(ct, d.CurType)
	f5 = append(f5, ct...)

	mc := make([]byte, 4)
	binary.BigEndian.PutUint32(mc, d.MaxCnt)
	f5 = append(f5, mc...)

	f5 = append(f5, d.SessID...)
	f5 = append(f5, d.Flags...)

	rc := make([]byte, 2)
	binary.BigEndian.PutUint16(rc, uint16(len(d.Records)))
	f5 = append(f5, rc...)

	for _, r := range d.Records {
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, r)
		f5 = append(f5, buf...)
	}

	f5b := base64.StdEncoding.EncodeToString(f5)
	f5bBytes := []byte(f5b)

	// Helper build Thrift byte array
	wi32 := func(v int32) []byte {
		buf := make([]byte, 4)
		binary.BigEndian.PutUint32(buf, uint32(v))
		return buf
	}
	wi64 := func(v int64) []byte {
		buf := make([]byte, 8)
		binary.BigEndian.PutUint64(buf, uint64(v))
		return buf
	}
	wf := func(t byte, fid uint16, val []byte) []byte {
		buf := make([]byte, 1+2+len(val))
		buf[0] = t
		binary.BigEndian.PutUint16(buf[1:3], fid)
		copy(buf[3:], val)
		return buf
	}

	wstop := []byte{0x00}

	// Bikin length prefix untuk string (f5b)
	f5bLen := make([]byte, 4)
	binary.BigEndian.PutUint32(f5bLen, uint32(len(f5bBytes)))
	f5bPayload := append(f5bLen, f5bBytes...)

	inner := make([]byte, 0)
	inner = append(inner, wf(10, 2, wi64(d.F2))...)
	inner = append(inner, wf(10, 3, wi64(d.F3))...)
	inner = append(inner, wf(8, 4, wi32(d.F4))...)
	inner = append(inner, wf(11, 5, f5bPayload)...)
	inner = append(inner, wf(8, 6, wi32(d.F6))...)
	inner = append(inner, wf(8, 7, wi32(d.F7))...)

	// Field 8 (Nested Struct untuk Anchor)
	f8Inner := wf(10, 1, wi64(d.Anchor))
	f8Inner = append(f8Inner, wstop...)
	inner = append(inner, wf(12, 8, f8Inner)...)
	inner = append(inner, wstop...)

	outer := wf(12, 2, inner)
	outer = append(outer, wstop...)

	// Encode to Base64
	result := base64.StdEncoding.EncodeToString(outer)

	// Convert ke URL-safe Base64
	result = strings.ReplaceAll(result, "+", "-")
	result = strings.ReplaceAll(result, "/", "_")
	result = strings.TrimRight(result, "=")

	return result
}

// GenerateNextCursor adalah *The Holy Grail* buatan lu.
// Fungsi ini ngambil cursor Bottom asli dari halaman sebelumnya,
// ngekstrak Anchor paling lama, ngereset records, dan ngeluarin Cursor baru (B')
func GenerateNextCursor(oldCursorStr string) (string, error) {
	d, err := Decode(oldCursorStr)
	if err != nil {
		return "", fmt.Errorf("gagal decode cursor lama: %w", err)
	}

	if len(d.Records) == 0 {
		return "", fmt.Errorf("cursor gak punya records, gak bisa dapet anchor")
	}

	// Cari Snowflake ID terkecil (Tweet paling lama) di halaman ini
	var anchorNew uint64 = d.Records[0]
	for _, r := range d.Records {
		if r < anchorNew {
			anchorNew = r
		}
	}

	// Menerapkan Algoritma Forgery (CTF Logic)
	d.Records = []uint64{anchorNew} // Reset ke 1 record
	d.Anchor = int64(anchorNew)     // Set Anchor baru
	d.F3 = int64(anchorNew)         // Batas bawah geser ke anchor baru

	// UPDATE PENTING DARI TEMUAN LU: f7 itu Page Counter!
	// Kita tambahin 1 biar server gak curiga
	d.F7 += 1

	return Encode(d), nil
}

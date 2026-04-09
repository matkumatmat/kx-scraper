package cursor

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"
)

const twitterEpoch = 1288834974657

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

func Decode(b64Str string) (*CursorData, error) {
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

	var ps func() map[uint16]interface{}
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
			case 8:
				fields[fid] = ri32()
			case 10:
				fields[fid] = ri64()
			case 11:
				fields[fid] = rstring()
			case 12:
				fields[fid] = ps()
			default:
				break Loop
			}
		}
		return fields
	}

	outer := ps()

	// FIX 1: DYNAMIC STRUCT SEARCH YANG LEBIH PINTER
	var innerRaw map[uint16]interface{}
	for _, v := range outer {
		if m, isMap := v.(map[uint16]interface{}); isMap {
			// Cek apakah field 5 beneran ada DAN tipenya []byte (bukan integer nyasar)
			if _, has5 := m[5].([]byte); has5 {
				// Cek apakah field 2 (max_id) juga ada. Kalo dua ini ada, fix ini struct target kita!
				if _, has2 := m[2].(int64); has2 {
					innerRaw = m
					break
				}
			}
		}
	}

	if innerRaw == nil {
		return nil, fmt.Errorf("format outer struct gak sesuai, struct target gak ketemu")
	}

	// Langsung sikat f5bRaw, gak usah ngecek !ok lagi karena udah pasti aman
	f5bRaw := innerRaw[5].([]byte)

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

	var anchor int64
	if field8, ok := innerRaw[8].(map[uint16]interface{}); ok {
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

	f8Inner := wf(10, 1, wi64(d.Anchor))
	f8Inner = append(f8Inner, wstop...)
	inner = append(inner, wf(12, 8, f8Inner)...)
	inner = append(inner, wstop...)

	outer := wf(12, 2, inner)
	outer = append(outer, wstop...)

	result := base64.StdEncoding.EncodeToString(outer)
	result = strings.ReplaceAll(result, "+", "-")
	result = strings.ReplaceAll(result, "/", "_")
	result = strings.TrimRight(result, "=")
	return result
}

// FIX 2: THE REAL FORGE LOGIC
// Kita BIKIN cursor baru, TAPI tetep jaga array Records biar ga ada duplikat.
func ForgeGodCursor(serverCursor string) (string, error) {
	d, err := Decode(serverCursor)
	if err != nil {
		return "", fmt.Errorf("gagal decode cursor: %w", err)
	}

	// 1. EXPLOIT: Reset Page Counter (f7) buat bypass proteksi deep pagination X
	d.F7 = 1

	// 2. EXPLOIT: Kita bisa nyoba nembus limit count server
	d.MaxCnt = 1000

	// 3. JANGAN SENTUH d.Records. Biarkan blacklist server tetep ada biar duplikat 0%.

	return Encode(d), nil
}

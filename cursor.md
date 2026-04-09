**tab Top**
[Outer Thrift Struct]
  Field#2 (struct):
    Field#2  i64  = max_tweet_snowflake_id   ← batas atas timeline
    Field#3  i64  = min_tweet_snowflake_id   ← batas bawah timeline  
    Field#4  i32  = 2                        ← page type / direction
    Field#5  str  = base64(nested_binary)    ← isi cursor utama
    Field#6  i32  = 0
    Field#7  i32  = 1
    Field#8 (struct):
      Field#1 i64 = anchor_tweet_id          ← posisi cursor saat ini

[nested_binary = 324 bytes]
  [0:2]   = 0x1263   magic header
  [2:4]   = 0xc2eb   cursor type/variant
  [4:8]   = 500      max count per page
  [8:16]  = session ID (8 bytes)
  [16:18] = 0x0000   flags
  [18:20] = 38       record count
  [20:N]  = 38 × 8 bytes = tweet snowflake IDs array
             └─ record[37] == anchor_id (selalu sama!)


**tab Latest**
0c 00 03 0c 00 01 0a 00 01 1c 57 8d 47 73 56 11 d6 0a 00 02 1c 57 8a 94 76 d7 31 82 00 08 00 02 00 00 00 02 08 00 03 00 00 00 00 08 00 04 00 00 00 00 0a 00 05 1c 57 8d 66 50 80 27 10 0a 00 06 1c 57 8d 66 50 7f d8 f0 00 00
Outer Struct [Field 3]:
  ├── Field 1 (Struct) -> The Pointers
  │     ├── 1.1 (i64): 2042256294053024214  [Konstanta: Max Tweet ID / Awal lu nge-search]
  │     └── 1.2 (i64): 2042253326289416578  [BERUBAH: Ini ANCHOR lu! (Tweet terlama di page 1)]
  │
  ├── Field 2 (i32): 2                      [Konstanta: Direction = Next Page]
  ├── Field 3 (i32): 0                      [Konstanta: Unknown flag]
  ├── Field 4 (i32): 0                      [BERUBAH: INI PAGE COUNTER LU! (0, 1, 2, 3...)]
  ├── Field 5 (i64): 2042256426612565776    [Konstanta: Timestamp query]
  └── Field 6 (i64): 2042256426612545776    [BERUBAH: Offset internal X]             
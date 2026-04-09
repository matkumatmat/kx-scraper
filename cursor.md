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
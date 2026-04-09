import json

def get_tweet_ids(filepath):
    ids = []
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            data = json.load(f)
        
        instructions = data.get('data', {}).get('search_by_raw_query', {}).get('search_timeline', {}).get('timeline', {}).get('instructions', [])
        for inst in instructions:
            entries = inst.get('entries', [])
            for entry in entries:
                content = entry.get('content', {})
                if content.get('__typename') == 'TimelineTimelineItem':
                    tweet = content.get('itemContent', {}).get('tweet_results', {}).get('result', {})
                    if tweet and 'rest_id' in tweet:
                        ids.append(tweet['rest_id'])
    except FileNotFoundError:
        print(f"[!] File {filepath} belum ada.")
    return set(ids)

# Ambil data dari 3 halaman beruntun
set1 = get_tweet_ids('data1.json')
set2 = get_tweet_ids('data_maxid.json')
set3 = get_tweet_ids('data3.json')

if set1 and set2 and set3:
    overlap_1_2 = set1.intersection(set2)
    overlap_2_3 = set2.intersection(set3)
    overlap_1_3 = set1.intersection(set3)
    
    total_overlap = len(overlap_1_2) + len(overlap_2_3) + len(overlap_1_3)

    print("="*55)
    print("📊 HASIL VALIDASI 3 HALAMAN BERUNTUN (MAX_ID)")
    print("="*55)
    print(f"Total Tweet Halaman 1 (data1.json)      : {len(set1)}")
    print(f"Total Tweet Halaman 2 (data_maxid.json) : {len(set2)}")
    print(f"Total Tweet Halaman 3 (data3.json)      : {len(set3)}")
    print("-" * 55)
    print(f"Overlap Hal 1 & 2 : {len(overlap_1_2)}")
    print(f"Overlap Hal 2 & 3 : {len(overlap_2_3)}")
    print(f"Overlap Hal 1 & 3 : {len(overlap_1_3)}")
    print("="*55)
    
    if total_overlap == 0:
        print("✅ SUKSES PERFECT 3x COMBO! 0 Duplikat!")
        print("🚀 Silakan pamerkan arsitektur ini ke seluruh dunia wkwk!")
    else:
        print("❌ Masih ada yang bocor bre!")
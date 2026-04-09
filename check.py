import json

def analyze_duplicates(filepath):
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            data = json.load(f)
    except FileNotFoundError:
        print(f"❌ File '{filepath}' gak ketemu bre!")
        return
    except json.JSONDecodeError:
        print(f"❌ File '{filepath}' JSON-nya belum valid (mungkin kurang bracket tutup ']').")
        return

    total_tweets = len(data)
    if total_tweets == 0:
        print("Kosong bre datanya.")
        return

    seen_ids = set()
    duplicates = []
    page_size = 20  # Standar 1 halaman X = 20 tweet

    print("="*65)
    print(f"📊 DATA SCRAPING ANALYSIS: {filepath}")
    print("="*65)

    for i, tweet in enumerate(data):
        post_id = tweet.get("PostID")
        
        # Kalkulasi index dan page
        page_num = (i // page_size) + 1
        tweet_idx_in_page = (i % page_size) + 1

        if post_id in seen_ids:
            duplicates.append({
                "id": post_id,
                "global_idx": i + 1,
                "page": page_num,
                "page_idx": tweet_idx_in_page
            })
        else:
            seen_ids.add(post_id)

    unique_count = len(seen_ids)
    duplicate_count = len(duplicates)
    duplicate_percentage = (duplicate_count / total_tweets) * 100

    print(f"Total Tweets Ditarik : {total_tweets}")
    print(f"Total Tweets Unik    : {unique_count}")
    print(f"Total Duplikat       : {duplicate_count}")
    print(f"Persentase Duplikat  : {duplicate_percentage:.2f} %")
    print("-" * 65)

    if duplicate_count == 0:
        print("✅ PERFECT SCROLL! 0% Duplikat.")
        print("🚀 Algoritma Forged Cursor BEKERJA SEMPURNA LINTAS AKUN!")
    else:
        print("⚠️ Ditemukan Duplikat. Detail penyebaran:")
        
        # Kelompokkin duplikat berdasarkan Halaman
        dup_by_page = {}
        for dup in duplicates:
            p = dup["page"]
            dup_by_page[p] = dup_by_page.get(p, 0) + 1

        for p, count in sorted(dup_by_page.items()):
            print(f" - Halaman {p} nyumbang {count} duplikat")

        print("\n🔍 5 Duplikat pertama (berdasarkan urutan masuk):")
        for dup in duplicates[:5]:
            print(f"   -> PostID: {dup['id']} nongol lagi di Halaman {dup['page']}, urutan ke-{dup['page_idx']}")
            
    print("="*65)

if __name__ == "__main__":
    # Sesuaiin sama nama file hasil export dari terminal lu
    analyze_duplicates("btc4.json")
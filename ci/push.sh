#!/usr/bin/env bash
set -e

echo "🚀 Preparing & Pushing to GitHub..."
echo "===================================="

# 1. Pastikan .gitignore ada
if [ ! -f .gitignore ]; then
    echo "❌ Error: .gitignore tidak ditemukan. Simpan dulu di root proyek!"
    exit 1
fi

# 2. Init git jika belum
if [ ! -d .git ]; then
    echo "📦 Initializing git repository..."
    git init
    BRANCH=$(git branch --show-current)
else
    BRANCH=$(git branch --show-current)
fi

# 3. Setup remote origin
if git remote | grep -q origin; then
    echo "✅ Remote 'origin' sudah terkonfigurasi."
else
    read -p "🔗 Paste GitHub repo URL (https atau git@github): " REPO_URL
    git remote add origin "$REPO_URL"
fi

# 4. Stage & Review
echo "📝 Staging files..."
git add .
echo ""
echo "📋 Files yang akan di-commit:"
git status --short
echo ""

# 5. Konfirmasi sebelum commit
read -p "✅ Lanjut commit & push ke branch '$BRANCH'? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    git commit -m "🚀 Initial commit: kx-scraper structure, CLI, engine, exporters & fixtures"
    echo "⬆️  Pushing to origin/$BRANCH..."
    git push -u origin "$BRANCH"
    echo "✨ Berhasil! Repo siap di-review."
else
    echo "🛑 Dibatalkan. Cek kembali file yang ter-stage sebelum menjalankan ulang."
fi
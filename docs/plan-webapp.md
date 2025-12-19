# htmlfetch テスト用Webアプリ計画

> **ステータス**: 未実装（計画のみ）
>
> **理由**: nz-html-fetchはライブラリとして依存を最小限に保つため、
> Webアプリは別リポジトリで実装すべき。

## コンセプト

htmlfetchライブラリをGUIでテストできるシンプルなWebアプリ。
lgn-site-check-goの `/sites/{id}` ページを参考に、
**URLを直接入力→取得→履歴管理→プレビュー**に特化。

サイト一覧管理やニュース抽出などの特化機能は不要。

## 画面構成

### メイン画面（1画面のみ）

```
┌─────────────────────────────────────────────────────────────┐
│  [URL入力欄........................] [取得ボタン]           │
│  □広告ブロック □画像ブロック □CSS埋め込み □スクリプト除去  │
├──────────────────────┬──────────────────────────────────────┤
│  履歴リスト          │  プレビュー                          │
│  ──────────────────  │                                      │
│  example.com         │  ┌────────────────────────────────┐  │
│  2024-12-19 16:30    │  │                                │  │
│  1.2s / 45KB         │  │    (iframe)                    │  │
│  [削除]              │  │                                │  │
│  ──────────────────  │  │                                │  │
│  yahoo.co.jp         │  │                                │  │
│  2024-12-19 16:25    │  └────────────────────────────────┘  │
│  0.8s / 251KB        │                                      │
│  [削除]              │  HTML: 251KB / Network: 757KB in     │
├──────────────────────┴──────────────────────────────────────┤
│  Status: 取得完了 (1.2秒)                                    │
└─────────────────────────────────────────────────────────────┘
```

## 機能

| 機能 | 説明 |
|------|------|
| URL取得 | 入力URLをhtmlfetchで取得、圧縮保存 |
| 履歴一覧 | 取得履歴をリスト表示（新しい順） |
| プレビュー | 選択した履歴のHTMLをiframeで表示 |
| 履歴削除 | 個別削除 |
| オプション | ブロッキング、CSS埋め込み、スクリプト除去 |

## DBスキーマ（SQLite）

```sql
CREATE TABLE snapshots (
    id INTEGER PRIMARY KEY,
    url TEXT NOT NULL,
    final_url TEXT NOT NULL,
    html_compressed BLOB NOT NULL,
    original_size INTEGER NOT NULL,
    compressed_size INTEGER NOT NULL,
    network_bytes_in INTEGER DEFAULT 0,
    network_bytes_out INTEGER DEFAULT 0,
    request_count INTEGER DEFAULT 0,
    fetch_duration_ms INTEGER NOT NULL,
    options_json TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

## ファイル構成（別リポジトリ想定）

```
nz-html-fetch-app/
├── go.mod               # nz-html-fetchをrequire
├── main.go
├── handlers.go
├── db/
│   ├── schema.sql
│   └── queries.sql
├── templates/
│   └── index.html
└── static/
    └── split.min.js
```

## 技術スタック

- **Web**: Echo + html/template
- **フロント**: htmx + Tailwind CDN + Split.js
- **DB**: SQLite（単一ファイル）
- **圧縮**: zstd

## APIエンドポイント

| Method | Path | 説明 |
|--------|------|------|
| GET | `/` | メインページ |
| POST | `/fetch` | URL取得（htmx） |
| GET | `/snapshots` | 履歴リスト（htmx partial） |
| GET | `/snapshots/:id` | スナップショット詳細JSON |
| GET | `/snapshots/:id/html` | 生HTML（iframe用） |
| DELETE | `/snapshots/:id` | 削除（htmx） |

## 依存関係の考慮

nz-html-fetchライブラリの依存を最小限に保つため、
このWebアプリは**別リポジトリ**として実装する。

```
nz-html-fetch/        ← ライブラリのみ
  依存: go-rod/rod, go-rod/stealth

nz-html-fetch-app/    ← Webアプリ（別リポジトリ）
  依存: nz-html-fetch, echo, sqlite, zstd, ...
```

## 参考

- lgn-site-check-go `/sites/{id}` ページ
  - `internal/handlers/business_sites.go`
  - `web/components/site_detail_v2.templ`

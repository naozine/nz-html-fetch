# nz-html-fetch

Chromiumブラウザを使用してWebページのHTMLを取得するGoライブラリ。JavaScript実行後の動的コンテンツも取得可能。

## 特徴

- **動的コンテンツ対応**: Chromium (Rod) によるJavaScript実行
- **Bot検出回避**: Stealthモードでブロック回避
- **リソースブロッキング**: 広告、画像、CSS、フォント等を個別にブロック
- **Markdown変換**: Readabilityでコンテンツ抽出し、Markdownに変換
- **ネットワーク統計**: リソース別の通信量を計測
- **高速モード**: `Start()/Close()`でブラウザを再利用

## インストール

```bash
go get github.com/naozine/nz-html-fetch
```

## 使い方

### ライブラリとして使用

```go
package main

import (
    "context"
    "fmt"
    "nz-html-fetch/pkg/htmlfetch"
)

func main() {
    fetcher := htmlfetch.New()

    result, err := fetcher.Fetch(context.Background(), "https://example.com")
    if err != nil {
        panic(err)
    }

    fmt.Println(result.HTML)
    fmt.Printf("Duration: %v\n", result.Duration)
    fmt.Printf("Bytes In: %d\n", result.Stats.TotalBytesIn)
}
```

### Markdown変換

```go
result, err := fetcher.Fetch(context.Background(), "https://example.com/article",
    htmlfetch.WithMarkdown(),
)
if err != nil {
    panic(err)
}

fmt.Println(result.Markdown) // コンテンツ部分のみMarkdownで出力
```

### 高速モード（ブラウザ再利用）

```go
fetcher := htmlfetch.New()

// ブラウザを起動して維持
if err := fetcher.Start(); err != nil {
    panic(err)
}
defer fetcher.Close()

// 複数URLを高速に取得
urls := []string{"https://example.com/1", "https://example.com/2"}
for _, url := range urls {
    result, _ := fetcher.Fetch(context.Background(), url)
    fmt.Println(result.FinalURL)
}
```

### オプション

```go
// Fetcher作成オプション
fetcher := htmlfetch.New(
    htmlfetch.WithStealth(true),           // Bot検出回避（デフォルト: true）
    htmlfetch.WithProxy("http://proxy:8080"), // プロキシ設定
    htmlfetch.WithBrowserPath("/path/to/chrome"), // ブラウザパス
)

// Fetch実行オプション
result, err := fetcher.Fetch(ctx, url,
    htmlfetch.WithWaitStrategy(htmlfetch.WaitNetworkIdle), // 待機戦略
    htmlfetch.WithSelector("#content", 10*time.Second),    // 要素待機
    htmlfetch.WithViewport(1920, 1080),                    // ビューポート
    htmlfetch.WithBlocking(htmlfetch.BlockingOptions{      // リソースブロック
        Ads:   true,
        Image: true,
    }),
    htmlfetch.WithEmbedCSS(),    // 外部CSSを埋め込み
    htmlfetch.WithStripScripts(), // スクリプト除去
    htmlfetch.WithMarkdown(),     // Markdown変換
)
```

## CLI

```bash
# ビルド
go build -o htmlfetch ./cmd/htmlfetch

# 基本的な使用
./htmlfetch https://example.com

# Markdown出力
./htmlfetch -output=markdown https://example.com/article

# 広告ブロック + 統計表示
./htmlfetch -block-ads -output=stats https://example.com

# 要素待機
./htmlfetch -wait=networkidle -selector="#content" https://example.com

# JSON出力（Markdown含む）
./htmlfetch -markdown -output=json https://example.com
```

### CLIオプション

| オプション | 説明 | デフォルト |
|-----------|------|-----------|
| `-wait` | 待機戦略 (load/networkidle/domstable) | load |
| `-selector` | 待機するCSSセレクタ | - |
| `-selector-timeout` | セレクタ待機タイムアウト（秒） | 30 |
| `-viewport` | ビューポートサイズ (WxH) | 1920x1080 |
| `-proxy` | プロキシアドレス | - |
| `-stealth` | Bot検出回避 | true |
| `-block-ads` | 広告ブロック | false |
| `-block-images` | 画像ブロック | false |
| `-block-css` | CSSブロック | false |
| `-block-fonts` | フォントブロック | false |
| `-embed-css` | 外部CSSを埋め込み | false |
| `-strip-scripts` | スクリプト除去 | false |
| `-markdown` | Markdown変換を有効化 | false |
| `-output` | 出力形式 (html/json/stats/markdown) | html |

## 出力形式

### html
取得したHTMLをそのまま出力

### markdown
Readabilityで抽出したコンテンツをMarkdown形式で出力

### json
```json
{
  "final_url": "https://example.com/",
  "duration_ms": 1234,
  "html_length": 52010,
  "markdown_length": 15607,
  "markdown": "# Title\n\nContent...",
  "stats": {
    "total_bytes_in": 567406,
    "total_bytes_out": 20544,
    "request_count": 36,
    "by_resource_type": {
      "Document": { "count": 2, "bytes_in": 14396, "bytes_out": 1121 },
      "Script": { "count": 12, "bytes_in": 359814, "bytes_out": 6421 }
    }
  }
}
```

### stats
```
URL: https://example.com/
Duration: 1.234s
HTML: 50.8 KB
Network: 554.1 KB in / 20.1 KB out (36 requests)

リソース別:
  Document: 2件, 14.1 KB in
  Script: 12件, 351.4 KB in
  Image: 16件, 35.8 KB in
```

## 依存ライブラリ

- [go-rod/rod](https://github.com/go-rod/rod) - ブラウザ自動化
- [go-rod/stealth](https://github.com/go-rod/stealth) - Bot検出回避
- [go-shiori/go-readability](https://github.com/go-shiori/go-readability) - コンテンツ抽出
- [JohannesKaufmann/html-to-markdown](https://github.com/JohannesKaufmann/html-to-markdown) - Markdown変換

## ライセンス

MIT

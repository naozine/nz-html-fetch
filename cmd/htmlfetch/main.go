package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"nz-html-fetch/pkg/htmlfetch"
)

func main() {
	// フラグ定義
	wait := flag.String("wait", "load", "待機戦略 (load/networkidle/domstable)")
	selector := flag.String("selector", "", "待機するCSSセレクタ")
	selectorTimeout := flag.Int("selector-timeout", 30, "セレクタ待機タイムアウト（秒）")
	viewport := flag.String("viewport", "1920x1080", "ビューポートサイズ (WxH)")
	proxy := flag.String("proxy", "", "プロキシアドレス")
	stealth := flag.Bool("stealth", true, "bot検出回避を有効化")
	blockAds := flag.Bool("block-ads", false, "広告ブロック")
	blockImages := flag.Bool("block-images", false, "画像ブロック")
	blockCSS := flag.Bool("block-css", false, "CSSブロック")
	blockFonts := flag.Bool("block-fonts", false, "フォントブロック")
	embedCSS := flag.Bool("embed-css", false, "外部CSSを埋め込み")
	stripScripts := flag.Bool("strip-scripts", false, "スクリプト除去")
	markdown := flag.Bool("markdown", false, "マークダウン変換を有効化")
	output := flag.String("output", "html", "出力形式 (html/json/stats/markdown)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "使用法: %s [オプション] URL\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "オプション:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n例:\n")
		fmt.Fprintf(os.Stderr, "  %s https://example.com\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -block-ads -output=stats https://example.com\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -wait=networkidle -selector=\"#content\" https://example.com\n", os.Args[0])
	}

	flag.Parse()

	// URLを取得
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "エラー: URLを指定してください")
		flag.Usage()
		os.Exit(1)
	}
	url := flag.Arg(0)

	// ビューポートをパース
	vpWidth, vpHeight := parseViewport(*viewport)

	// Fetcherオプションを構築
	var fetcherOpts []htmlfetch.Option
	fetcherOpts = append(fetcherOpts, htmlfetch.WithStealth(*stealth))
	if *proxy != "" {
		fetcherOpts = append(fetcherOpts, htmlfetch.WithProxy(*proxy))
	}

	// Fetchオプションを構築
	var fetchOpts []htmlfetch.FetchOption
	fetchOpts = append(fetchOpts, htmlfetch.WithWaitStrategy(parseWaitStrategy(*wait)))
	fetchOpts = append(fetchOpts, htmlfetch.WithViewport(vpWidth, vpHeight))

	if *selector != "" {
		fetchOpts = append(fetchOpts, htmlfetch.WithSelector(*selector, time.Duration(*selectorTimeout)*time.Second))
	}

	// ブロッキングオプション
	blocking := htmlfetch.BlockingOptions{
		Ads:        *blockAds,
		Image:      *blockImages,
		Stylesheet: *blockCSS,
		Font:       *blockFonts,
	}
	fetchOpts = append(fetchOpts, htmlfetch.WithBlocking(blocking))

	if *embedCSS {
		fetchOpts = append(fetchOpts, htmlfetch.WithEmbedCSS())
	}
	if *stripScripts {
		fetchOpts = append(fetchOpts, htmlfetch.WithStripScripts())
	}
	if *markdown || *output == "markdown" {
		fetchOpts = append(fetchOpts, htmlfetch.WithMarkdown())
	}

	// フェッチ実行
	fetcher := htmlfetch.New(fetcherOpts...)
	result, err := fetcher.Fetch(context.Background(), url, fetchOpts...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "エラー: %v\n", err)
		os.Exit(1)
	}

	// 結果を出力
	switch *output {
	case "html":
		fmt.Print(result.HTML)
	case "markdown":
		fmt.Print(result.Markdown)
	case "json":
		outputJSON(result, *markdown)
	case "stats":
		outputStats(result)
	default:
		fmt.Fprintf(os.Stderr, "エラー: 不明な出力形式: %s\n", *output)
		os.Exit(1)
	}
}

// parseViewport はビューポート文字列をパース
func parseViewport(s string) (int, int) {
	parts := strings.Split(s, "x")
	if len(parts) != 2 {
		return 1920, 1080
	}
	w, err1 := strconv.Atoi(parts[0])
	h, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil {
		return 1920, 1080
	}
	return w, h
}

// parseWaitStrategy は待機戦略文字列をパース
func parseWaitStrategy(s string) htmlfetch.WaitStrategy {
	switch s {
	case "networkidle":
		return htmlfetch.WaitNetworkIdle
	case "domstable":
		return htmlfetch.WaitDOMStable
	default:
		return htmlfetch.WaitLoad
	}
}

// outputJSON はJSON形式で出力
func outputJSON(result *htmlfetch.Result, includeMarkdown bool) {
	type jsonOutput struct {
		FinalURL       string `json:"final_url"`
		DurationMs     int64  `json:"duration_ms"`
		HTMLLength     int    `json:"html_length"`
		MarkdownLength int    `json:"markdown_length,omitempty"`
		Markdown       string `json:"markdown,omitempty"`
		Stats          struct {
			TotalBytesIn   int64 `json:"total_bytes_in"`
			TotalBytesOut  int64 `json:"total_bytes_out"`
			RequestCount   int   `json:"request_count"`
			ByResourceType map[string]struct {
				Count    int   `json:"count"`
				BytesIn  int64 `json:"bytes_in"`
				BytesOut int64 `json:"bytes_out"`
			} `json:"by_resource_type,omitempty"`
		} `json:"stats"`
	}

	out := jsonOutput{
		FinalURL:   result.FinalURL,
		DurationMs: result.Duration.Milliseconds(),
		HTMLLength: len(result.HTML),
	}
	if includeMarkdown && result.Markdown != "" {
		out.MarkdownLength = len(result.Markdown)
		out.Markdown = result.Markdown
	}
	out.Stats.TotalBytesIn = result.Stats.TotalBytesIn
	out.Stats.TotalBytesOut = result.Stats.TotalBytesOut
	out.Stats.RequestCount = result.Stats.RequestCount

	if len(result.Stats.ByResourceType) > 0 {
		out.Stats.ByResourceType = make(map[string]struct {
			Count    int   `json:"count"`
			BytesIn  int64 `json:"bytes_in"`
			BytesOut int64 `json:"bytes_out"`
		})
		for k, v := range result.Stats.ByResourceType {
			out.Stats.ByResourceType[k] = struct {
				Count    int   `json:"count"`
				BytesIn  int64 `json:"bytes_in"`
				BytesOut int64 `json:"bytes_out"`
			}{
				Count:    v.Count,
				BytesIn:  v.BytesIn,
				BytesOut: v.BytesOut,
			}
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(out)
}

// outputStats は統計情報を人間が読みやすい形式で出力
func outputStats(result *htmlfetch.Result) {
	fmt.Printf("URL: %s\n", result.FinalURL)
	fmt.Printf("Duration: %v\n", result.Duration.Round(time.Millisecond))
	fmt.Printf("HTML: %s\n", formatBytes(int64(len(result.HTML))))
	fmt.Printf("Network: %s in / %s out (%d requests)\n",
		formatBytes(result.Stats.TotalBytesIn),
		formatBytes(result.Stats.TotalBytesOut),
		result.Stats.RequestCount,
	)

	if len(result.Stats.ByResourceType) > 0 {
		fmt.Println("\nリソース別:")
		for rtype, stat := range result.Stats.ByResourceType {
			fmt.Printf("  %s: %d件, %s in\n", rtype, stat.Count, formatBytes(stat.BytesIn))
		}
	}
}

// formatBytes はバイト数を読みやすい形式に変換
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d bytes", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

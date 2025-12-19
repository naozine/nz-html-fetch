package htmlfetch

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
)

// Fetcher はrod/Chromiumを使ったHTMLフェッチャー
type Fetcher struct {
	config fetcherConfig
}

// New は新しいFetcherを作成
func New(opts ...Option) *Fetcher {
	cfg := fetcherConfig{
		stealth: true, // デフォルトでstealthを有効
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Fetcher{config: cfg}
}

// Fetch はURLからHTMLを取得
// 毎回ブラウザを起動→取得→終了する
func (f *Fetcher) Fetch(ctx context.Context, url string, opts ...FetchOption) (*Result, error) {
	startTime := time.Now()

	// Fetch設定を構築
	cfg := &fetchConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	applyDefaults(cfg)

	// ブラウザパスを決定
	browserPath := f.config.browserPath
	if browserPath == "" {
		browserPath = detectBrowserPath()
	}

	// ブラウザを起動
	l := launcher.New().Bin(browserPath).Headless(true).NoSandbox(true).
		Leakless(false).
		Set("disable-dev-shm-usage").
		Set("disable-gpu").
		Set("disable-software-rasterizer").
		Set("disable-extensions").
		Set("no-first-run").
		Set("no-default-browser-check").
		Set("disable-background-networking").
		Set("disable-sync").
		Set("disable-breakpad")

	// プロキシが指定されている場合は設定
	if f.config.proxy != "" {
		l = l.Proxy(f.config.proxy)
	}

	launchURL, err := l.Launch()
	if err != nil {
		return nil, &FetchError{
			Code:    ErrBrowserLaunchFailed,
			Message: "ブラウザの起動に失敗しました",
			Cause:   err,
		}
	}

	browser := rod.New().ControlURL(launchURL).MustConnect()
	defer browser.MustClose()

	// ページを作成（stealthモードの場合はstealth経由）
	var page *rod.Page
	if f.config.stealth {
		page = stealth.MustPage(browser)
	} else {
		page = browser.MustPage()
	}
	defer page.MustClose()

	// ブロッキングセットを作成
	blockSet := newBlockingSet(cfg.blocking)

	// ネットワーク統計収集を設定
	collector := newStatsCollector(blockSet)
	collector.setupNetworkStats(page)

	// リソースブロッキングを設定
	setupFetchBlocking(page, blockSet, cfg.blocking.Ads)

	// ビューポートを設定
	page.MustSetViewport(cfg.viewportWidth, cfg.viewportHeight, 1, false)

	// タイムアウトを設定（最大60秒）
	page = page.Timeout(60 * time.Second)

	// コンテキストを適用
	if ctx != nil {
		page = page.Context(ctx)
	}

	// ページに遷移
	if err := page.Navigate(url); err != nil {
		return nil, &FetchError{
			Code:    ErrNavigationFailed,
			Message: "ページへのナビゲーションに失敗しました",
			Cause:   err,
		}
	}

	// 待機戦略に応じて待機
	if err := waitForPage(page, cfg.waitStrategy); err != nil {
		return nil, &FetchError{
			Code:    ErrFetchTimeout,
			Message: "ページの読み込みがタイムアウトしました",
			Cause:   err,
		}
	}

	// セレクタ待機（オプション）
	if cfg.selector != "" {
		if err := waitForSelector(page, cfg.selector, cfg.selectorTimeout); err != nil {
			return nil, &FetchError{
				Code:    ErrSelectorNotFound,
				Message: fmt.Sprintf("セレクタ '%s' が見つかりませんでした", cfg.selector),
				Cause:   err,
			}
		}
	}

	// 少し待ってイベントを確実に収集
	time.Sleep(100 * time.Millisecond)

	// CSS埋め込み（オプション）
	if cfg.embedCSS {
		_ = embedCSS(page)
	}

	// スクリプト除去（オプション）
	if cfg.stripScripts {
		_ = stripScripts(page)
	}

	// HTMLを取得
	html, err := page.HTML()
	if err != nil {
		return nil, &FetchError{
			Code:    ErrInternalError,
			Message: "HTMLの取得に失敗しました",
			Cause:   err,
		}
	}

	// 最終URLを取得
	finalURL := url
	if info, err := page.Info(); err == nil {
		finalURL = info.URL
	}

	return &Result{
		HTML:     html,
		FinalURL: finalURL,
		Stats:    collector.getStats(),
		Duration: time.Since(startTime),
	}, nil
}

// detectBrowserPath はブラウザのパスを自動検出
func detectBrowserPath() string {
	// headless-shellイメージを優先
	headlessPath := "/headless-shell/headless-shell"
	if _, err := os.Stat(headlessPath); err == nil {
		return headlessPath
	}
	// launcherに任せる
	path, _ := launcher.LookPath()
	return path
}

// waitForPage は待機戦略に応じてページを待機
func waitForPage(page *rod.Page, strategy WaitStrategy) error {
	switch strategy {
	case WaitNetworkIdle:
		return page.WaitIdle(2 * time.Second)
	case WaitDOMStable:
		return page.WaitStable(500 * time.Millisecond)
	case WaitLoad:
		fallthrough
	default:
		return page.WaitLoad()
	}
}

// waitForSelector はセレクタが表示されるまで待機
func waitForSelector(page *rod.Page, selector string, timeout time.Duration) error {
	return page.Timeout(timeout).MustElement(selector).WaitVisible()
}

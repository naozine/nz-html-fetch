package htmlfetch

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/stealth"
)

// Fetcher はrod/Chromiumを使ったHTMLフェッチャー
type Fetcher struct {
	config  fetcherConfig
	browser *rod.Browser
	mu      sync.Mutex
	started bool
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

// Start はブラウザを起動して維持する（高速モード）
// Start()を呼ぶと、Fetch()はタブの作成/破棄のみ行う
// 使用後は必ずClose()を呼ぶこと
func (f *Fetcher) Start() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.started {
		return nil
	}

	browser, err := f.launchBrowser()
	if err != nil {
		return err
	}

	f.browser = browser
	f.started = true
	return nil
}

// Close はブラウザを終了する
// Start()を呼んだ場合は必ずClose()を呼ぶこと
func (f *Fetcher) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if !f.started {
		return nil
	}

	err := f.browser.Close()
	f.browser = nil
	f.started = false
	return err
}

// IsStarted はブラウザが起動中かを返す
func (f *Fetcher) IsStarted() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.started
}

// Fetch はURLからHTMLを取得
// Start()が呼ばれていない場合: 毎回ブラウザを起動→取得→終了（従来動作）
// Start()が呼ばれている場合: タブを作成→取得→閉じる（高速モード）
func (f *Fetcher) Fetch(ctx context.Context, url string, opts ...FetchOption) (*Result, error) {
	startTime := time.Now()

	// Fetch設定を構築
	cfg := &fetchConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	applyDefaults(cfg)

	// ブラウザとページを取得
	_, page, cleanup, err := f.getBrowserAndPage()
	if err != nil {
		return nil, err
	}
	defer cleanup()

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

// getBrowserAndPage はブラウザとページを取得し、クリーンアップ関数を返す
func (f *Fetcher) getBrowserAndPage() (*rod.Browser, *rod.Page, func(), error) {
	f.mu.Lock()
	started := f.started
	browser := f.browser
	f.mu.Unlock()

	if started {
		// 高速モード: 既存ブラウザでタブを作成
		page, err := f.createPage(browser)
		if err != nil {
			return nil, nil, nil, err
		}
		cleanup := func() {
			page.MustClose()
		}
		return browser, page, cleanup, nil
	}

	// 従来モード: ブラウザを起動
	newBrowser, err := f.launchBrowser()
	if err != nil {
		return nil, nil, nil, err
	}

	page, err := f.createPage(newBrowser)
	if err != nil {
		newBrowser.MustClose()
		return nil, nil, nil, err
	}

	cleanup := func() {
		page.MustClose()
		newBrowser.MustClose()
	}
	return newBrowser, page, cleanup, nil
}

// launchBrowser はブラウザを起動
func (f *Fetcher) launchBrowser() (*rod.Browser, error) {
	browserPath := f.config.browserPath
	if browserPath == "" {
		browserPath = detectBrowserPath()
	}

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
	return browser, nil
}

// createPage はページを作成
func (f *Fetcher) createPage(browser *rod.Browser) (*rod.Page, error) {
	if f.config.stealth {
		return stealth.MustPage(browser), nil
	}
	return browser.MustPage(), nil
}

// detectBrowserPath はブラウザのパスを自動検出
func detectBrowserPath() string {
	headlessPath := "/headless-shell/headless-shell"
	if _, err := os.Stat(headlessPath); err == nil {
		return headlessPath
	}
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

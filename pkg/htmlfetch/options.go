package htmlfetch

import "time"

// Fetcher設定（内部状態）
type fetcherConfig struct {
	browserPath string
	stealth     bool
	proxy       string
}

// Option はFetcher作成時のオプション
type Option func(*fetcherConfig)

// WithBrowserPath はブラウザパスを指定
func WithBrowserPath(path string) Option {
	return func(c *fetcherConfig) {
		c.browserPath = path
	}
}

// WithProxy はプロキシアドレスを指定
func WithProxy(addr string) Option {
	return func(c *fetcherConfig) {
		c.proxy = addr
	}
}

// WithStealth はstealth（bot検出回避）を有効化
func WithStealth(enabled bool) Option {
	return func(c *fetcherConfig) {
		c.stealth = enabled
	}
}

// Fetch実行時設定（内部状態）
type fetchConfig struct {
	waitStrategy    WaitStrategy
	selector        string
	selectorTimeout time.Duration
	viewportWidth   int
	viewportHeight  int
	blocking        BlockingOptions
	embedCSS        bool
	stripScripts    bool
	markdown        bool
}

// FetchOption はFetch実行時のオプション
type FetchOption func(*fetchConfig)

// WithWaitStrategy は待機戦略を指定
func WithWaitStrategy(strategy WaitStrategy) FetchOption {
	return func(c *fetchConfig) {
		c.waitStrategy = strategy
	}
}

// WithSelector は要素待機を指定
func WithSelector(selector string, timeout time.Duration) FetchOption {
	return func(c *fetchConfig) {
		c.selector = selector
		c.selectorTimeout = timeout
	}
}

// WithViewport はビューポートサイズを指定
func WithViewport(width, height int) FetchOption {
	return func(c *fetchConfig) {
		c.viewportWidth = width
		c.viewportHeight = height
	}
}

// WithBlocking はリソースブロッキングを指定
func WithBlocking(opts BlockingOptions) FetchOption {
	return func(c *fetchConfig) {
		c.blocking = opts
	}
}

// WithEmbedCSS は外部CSSの埋め込みを有効化
func WithEmbedCSS() FetchOption {
	return func(c *fetchConfig) {
		c.embedCSS = true
	}
}

// WithStripScripts はスクリプト除去を有効化
func WithStripScripts() FetchOption {
	return func(c *fetchConfig) {
		c.stripScripts = true
	}
}

// WithMarkdown はマークダウン変換を有効化
func WithMarkdown() FetchOption {
	return func(c *fetchConfig) {
		c.markdown = true
	}
}

// デフォルト値を適用
func applyDefaults(c *fetchConfig) {
	if c.waitStrategy == "" {
		c.waitStrategy = WaitLoad
	}
	if c.viewportWidth == 0 {
		c.viewportWidth = 1920
	}
	if c.viewportHeight == 0 {
		c.viewportHeight = 1080
	}
	if c.selectorTimeout == 0 {
		c.selectorTimeout = 30 * time.Second
	}
}

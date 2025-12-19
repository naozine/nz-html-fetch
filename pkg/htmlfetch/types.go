package htmlfetch

import "time"

// Result はフェッチ結果
type Result struct {
	HTML     string
	FinalURL string
	Stats    NetworkStats
	Duration time.Duration
}

// NetworkStats はネットワーク通信統計
type NetworkStats struct {
	TotalBytesIn   int64
	TotalBytesOut  int64
	RequestCount   int
	ByResourceType map[string]*ResourceStat
}

// ResourceStat はリソースタイプ別統計
type ResourceStat struct {
	Count    int
	BytesIn  int64
	BytesOut int64
}

// BlockingOptions はブロック設定
type BlockingOptions struct {
	Ads        bool
	Image      bool
	Stylesheet bool
	Font       bool
	Media      bool
	Ping       bool
	Script     bool
	XHR        bool
	Fetch      bool
}

// WaitStrategy は待機戦略
type WaitStrategy string

const (
	WaitLoad        WaitStrategy = "load"
	WaitNetworkIdle WaitStrategy = "networkidle"
	WaitDOMStable   WaitStrategy = "domstable"
)

// FetchError は構造化エラー
type FetchError struct {
	Code    string
	Message string
	Cause   error
}

func (e *FetchError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *FetchError) Unwrap() error {
	return e.Cause
}

// エラーコード定数
const (
	ErrBrowserLaunchFailed = "BROWSER_LAUNCH_FAILED"
	ErrNavigationFailed    = "NAVIGATION_FAILED"
	ErrFetchTimeout        = "FETCH_TIMEOUT"
	ErrSelectorNotFound    = "SELECTOR_NOT_FOUND"
	ErrInternalError       = "INTERNAL_ERROR"
)

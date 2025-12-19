package htmlfetch

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// blockingSet はブロックするリソースタイプのセット
type blockingSet map[string]bool

// newBlockingSet はBlockingOptionsからブロッキングセットを作成
func newBlockingSet(opts BlockingOptions) blockingSet {
	set := make(blockingSet)
	if opts.Image {
		set["Image"] = true
	}
	if opts.Stylesheet {
		set["Stylesheet"] = true
	}
	if opts.Font {
		set["Font"] = true
	}
	if opts.Media {
		set["Media"] = true
	}
	if opts.Ping {
		set["Ping"] = true
	}
	if opts.Script {
		set["Script"] = true
	}
	if opts.XHR {
		set["XHR"] = true
	}
	if opts.Fetch {
		set["Fetch"] = true
	}
	return set
}

// shouldBlock はリソースタイプがブロック対象かを判定
func (b blockingSet) shouldBlock(resourceType string) bool {
	return b[resourceType]
}

// isEmpty はブロック対象がないかを判定
func (b blockingSet) isEmpty() bool {
	return len(b) == 0
}

// setupFetchBlocking はCDP Fetchドメインでリソースブロッキングを設定
func setupFetchBlocking(page *rod.Page, blockSet blockingSet, blockAds bool) {
	if blockSet.isEmpty() && !blockAds {
		return
	}

	// Fetchドメインを有効化してリクエストをインターセプト
	patterns := []*proto.FetchRequestPattern{
		{URLPattern: "*", RequestStage: proto.FetchRequestStageRequest},
	}
	_ = proto.FetchEnable{Patterns: patterns}.Call(page)

	// リクエストを監視
	go page.EachEvent(func(e *proto.FetchRequestPaused) {
		reqURL := e.Request.URL
		resourceType := string(e.ResourceType)

		// リソースタイプによるブロック
		if blockSet.shouldBlock(resourceType) {
			_ = proto.FetchFailRequest{
				RequestID:   e.RequestID,
				ErrorReason: proto.NetworkErrorReasonBlockedByClient,
			}.Call(page)
			return
		}

		// 広告ドメインによるブロック
		if blockAds && isAdURL(reqURL) {
			_ = proto.FetchFailRequest{
				RequestID:   e.RequestID,
				ErrorReason: proto.NetworkErrorReasonBlockedByClient,
			}.Call(page)
			return
		}

		// リクエストを続行
		_ = proto.FetchContinueRequest{RequestID: e.RequestID}.Call(page)
	})()
}

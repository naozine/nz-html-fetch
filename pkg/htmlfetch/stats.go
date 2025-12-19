package htmlfetch

import (
	"sync"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// requestInfo はリクエスト情報を一時保存する構造体
type requestInfo struct {
	resourceType string
	requestSize  int64
}

// statsCollector はネットワーク統計を収集する
type statsCollector struct {
	stats    *NetworkStats
	requests map[proto.NetworkRequestID]*requestInfo
	blockSet blockingSet
	mu       sync.Mutex
}

// newStatsCollector は新しいstatsCollectorを作成
func newStatsCollector(blockSet blockingSet) *statsCollector {
	return &statsCollector{
		stats: &NetworkStats{
			ByResourceType: make(map[string]*ResourceStat),
		},
		requests: make(map[proto.NetworkRequestID]*requestInfo),
		blockSet: blockSet,
	}
}

// setupNetworkStats はネットワーク統計収集を設定
func (sc *statsCollector) setupNetworkStats(page *rod.Page) {
	// Networkドメインを有効化
	_ = proto.NetworkEnable{}.Call(page)

	// リクエスト送信時とレスポンス完了時のイベントを監視
	go page.EachEvent(func(e *proto.NetworkRequestWillBeSent) {
		sc.mu.Lock()
		defer sc.mu.Unlock()

		resourceType := string(e.Type)
		if resourceType == "" {
			resourceType = "Other"
		}

		// ブロックされたリソースはカウントしない
		if sc.blockSet.shouldBlock(resourceType) {
			return
		}

		// リクエストヘッダーのサイズを概算
		requestSize := int64(len(e.Request.URL))
		if e.Request.PostData != "" {
			requestSize += int64(len(e.Request.PostData))
		}
		// ヘッダーサイズの概算を追加
		requestSize += 500

		sc.requests[e.RequestID] = &requestInfo{
			resourceType: resourceType,
			requestSize:  requestSize,
		}

		sc.stats.TotalBytesOut += requestSize
		sc.stats.RequestCount++

		if sc.stats.ByResourceType[resourceType] == nil {
			sc.stats.ByResourceType[resourceType] = &ResourceStat{}
		}
		sc.stats.ByResourceType[resourceType].Count++
		sc.stats.ByResourceType[resourceType].BytesOut += requestSize
	}, func(e *proto.NetworkLoadingFinished) {
		sc.mu.Lock()
		defer sc.mu.Unlock()

		// レスポンスサイズを加算
		bytesIn := int64(e.EncodedDataLength)
		sc.stats.TotalBytesIn += bytesIn

		if info, ok := sc.requests[e.RequestID]; ok {
			if sc.stats.ByResourceType[info.resourceType] != nil {
				sc.stats.ByResourceType[info.resourceType].BytesIn += bytesIn
			}
		}
	})()
}

// getStats は収集した統計を返す
func (sc *statsCollector) getStats() NetworkStats {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return *sc.stats
}

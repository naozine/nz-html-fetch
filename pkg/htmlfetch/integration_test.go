package htmlfetch

import (
	"context"
	"strings"
	"testing"
)

// TestFetchDynamicContent は動的コンテンツの取得を各待機戦略でテストする。
// Chromiumブラウザが必要。go test -short でスキップされる。
func TestFetchDynamicContent(t *testing.T) {
	if testing.Short() {
		t.Skip("統合テストをスキップ（-short指定）")
	}

	ts := newTestServer(t)
	defer ts.Close()

	fetcher := New(WithStealth(false))
	if err := fetcher.Start(); err != nil {
		t.Fatalf("ブラウザの起動に失敗: %v", err)
	}
	defer fetcher.Close()

	t.Run("Static_WaitLoad", func(t *testing.T) {
		result, err := fetcher.Fetch(context.Background(), ts.URL+"/",
			WithWaitStrategy(WaitLoad))
		if err != nil {
			t.Fatalf("Fetchに失敗: %v", err)
		}
		assertContains(t, result.HTML, "STATIC_CONTENT_MARKER")
	})

	t.Run("Handlebars_WaitAuto", func(t *testing.T) {
		result, err := fetcher.Fetch(context.Background(), ts.URL+"/handlebars",
			WithWaitStrategy(WaitAuto))
		if err != nil {
			t.Fatalf("Fetchに失敗: %v", err)
		}
		assertContains(t, result.HTML, "DYNAMIC_HANDLEBARS_TITLE_1")
		assertContains(t, result.HTML, "DYNAMIC_HANDLEBARS_TITLE_2")
		assertContains(t, result.HTML, "DYNAMIC_HANDLEBARS_TITLE_3")
		assertContains(t, result.HTML, "DYNAMIC_HANDLEBARS_BODY_1")
		assertContains(t, result.HTML, `class="news-item"`)
	})

	t.Run("RiotJS_WaitAuto", func(t *testing.T) {
		result, err := fetcher.Fetch(context.Background(), ts.URL+"/riotjs",
			WithWaitStrategy(WaitAuto))
		if err != nil {
			t.Fatalf("Fetchに失敗: %v", err)
		}
		assertContains(t, result.HTML, "DYNAMIC_RIOT_TOPIC_1")
		assertContains(t, result.HTML, "DYNAMIC_RIOT_TOPIC_2")
		assertContains(t, result.HTML, "DYNAMIC_RIOT_TOPIC_3")
		assertContains(t, result.HTML, `class="topic-item"`)
	})

	// WaitNetworkIdle（WaitIdle）単独では非同期JSの描画完了を保証しない。
	// これは期待通りの動作であり、autoが必要な理由でもある。
	t.Run("Handlebars_WaitNetworkIdle_MayMissDynamic", func(t *testing.T) {
		result, err := fetcher.Fetch(context.Background(), ts.URL+"/handlebars",
			WithWaitStrategy(WaitNetworkIdle))
		if err != nil {
			t.Fatalf("Fetchに失敗: %v", err)
		}
		// WaitNetworkIdleでは動的コンテンツが取れない場合がある
		if strings.Contains(result.HTML, "DYNAMIC_HANDLEBARS_TITLE_1") {
			t.Log("WaitNetworkIdleでも動的コンテンツが取れた（タイミング依存）")
		} else {
			t.Log("WaitNetworkIdleでは動的コンテンツが取れなかった（期待通り）")
		}
	})

	t.Run("Handlebars_WaitDOMStable", func(t *testing.T) {
		result, err := fetcher.Fetch(context.Background(), ts.URL+"/handlebars",
			WithWaitStrategy(WaitDOMStable))
		if err != nil {
			t.Fatalf("Fetchに失敗: %v", err)
		}
		assertContains(t, result.HTML, "DYNAMIC_HANDLEBARS_TITLE_1")
	})
}

func assertContains(t *testing.T, html, marker string) {
	t.Helper()
	if !strings.Contains(html, marker) {
		t.Errorf("HTMLに %q が含まれるべきですが、含まれていません (HTML長: %d)", marker, len(html))
	}
}

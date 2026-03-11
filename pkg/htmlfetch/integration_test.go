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

// TestAcceptLanguage_DefaultRedirectsToJa はデフォルトのAccept-Language設定で
// 日本語ページにリダイレクトされることを検証する。
func TestAcceptLanguage_DefaultRedirectsToJa(t *testing.T) {
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

	result, err := fetcher.Fetch(context.Background(), ts.URL+"/lang-redirect",
		WithWaitStrategy(WaitLoad))
	if err != nil {
		t.Fatalf("Fetchに失敗: %v", err)
	}

	// デフォルトで日本語ページにリダイレクトされるべき
	assertContains(t, result.HTML, "LANG_JA_MARKER")

	if strings.Contains(result.HTML, "LANG_EN_MARKER") {
		t.Error("英語ページにリダイレクトされました（Accept-Languageが日本語になっていない）")
	}
}

// TestUserAgent_NoHeadlessChrome はUserAgentにHeadlessChromeが含まれないことを検証する。
func TestUserAgent_NoHeadlessChrome(t *testing.T) {
	if testing.Short() {
		t.Skip("統合テストをスキップ（-short指定）")
	}

	ts := newTestServer(t)
	defer ts.Close()

	tests := []struct {
		name    string
		stealth bool
	}{
		{"Stealth", true},
		{"NonStealth", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := New(WithStealth(tt.stealth))
			if err := fetcher.Start(); err != nil {
				t.Fatalf("ブラウザの起動に失敗: %v", err)
			}
			defer fetcher.Close()

			result, err := fetcher.Fetch(context.Background(), ts.URL+"/echo-headers",
				WithWaitStrategy(WaitLoad))
			if err != nil {
				t.Fatalf("Fetchに失敗: %v", err)
			}

			if strings.Contains(result.HTML, "HeadlessChrome") {
				t.Error("UserAgentにHeadlessChromeが含まれています")
			}
			assertContains(t, result.HTML, "Chrome/")
		})
	}
}

func assertContains(t *testing.T, html, marker string) {
	t.Helper()
	if !strings.Contains(html, marker) {
		t.Errorf("HTMLに %q が含まれるべきですが、含まれていません (HTML長: %d)", marker, len(html))
	}
}

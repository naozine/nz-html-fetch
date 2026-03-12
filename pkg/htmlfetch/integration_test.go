package htmlfetch

import (
	"context"
	"encoding/json"
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

// TestStealth_BotDetection はstealth有効時にbot検出チェックをパスすることを検証する。
// 各チェック項目の詳細はtestserver_test.goのbotDetectPageコメントを参照。
func TestStealth_BotDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("統合テストをスキップ（-short指定）")
	}

	ts := newTestServer(t)
	defer ts.Close()

	fetcher := New(WithStealth(true))
	if err := fetcher.Start(); err != nil {
		t.Fatalf("ブラウザの起動に失敗: %v", err)
	}
	defer fetcher.Close()

	result, err := fetcher.Fetch(context.Background(), ts.URL+"/bot-detect",
		WithWaitStrategy(WaitDOMStable))
	if err != nil {
		t.Fatalf("Fetchに失敗: %v", err)
	}

	// <pre id="results">...</pre> からJSONを抽出
	start := strings.Index(result.HTML, `<pre id="results">`)
	end := strings.Index(result.HTML, `</pre>`)
	if start == -1 || end == -1 {
		t.Fatalf("結果のHTMLからJSONを抽出できません (HTML長: %d)", len(result.HTML))
	}
	jsonStr := result.HTML[start+len(`<pre id="results">`):end]

	var results map[string]string
	if err := json.Unmarshal([]byte(jsonStr), &results); err != nil {
		t.Fatalf("結果JSONのパースに失敗: %v\nJSON: %s", err, jsonStr)
	}

	// stealth対策済みのチェック項目（PASSが期待される）
	mustPass := []struct {
		key  string
		desc string
	}{
		{"webdriver", "navigator.webdriver が false/undefined"},
		{"chrome_exists", "window.chrome が存在する"},
		{"chrome_app", "window.chrome.app が存在する"},
		{"chrome_csi", "window.chrome.csi が関数として存在する"},
		{"chrome_loadTimes", "window.chrome.loadTimes が関数として存在する"},
		{"plugins", "navigator.plugins が空でない"},
		{"languages", "navigator.languages が空でない"},
		{"outer_dimensions", "window.outerWidth/Height が 0 でない"},
		{"user_agent", "User-Agentに HeadlessChrome が含まれない"},
	}

	for _, tc := range mustPass {
		t.Run(tc.key, func(t *testing.T) {
			v, ok := results[tc.key]
			if !ok {
				t.Errorf("チェック項目 %q が結果に含まれていません", tc.key)
				return
			}
			if v != "PASS" {
				t.Errorf("%s: %s (got %s)", tc.key, tc.desc, v)
			}
		})
	}

	// SKIPを許容するチェック項目（環境依存）
	maySkip := []struct {
		key  string
		desc string
	}{
		{"notification_permission", "Notification.permission が denied でない"},
		{"webgl", "WebGL renderer に SwiftShader が含まれない"},
		{"fn_toString", "Permissions.prototype.query.toString() が [native code] を返す"},
	}

	for _, tc := range maySkip {
		t.Run(tc.key, func(t *testing.T) {
			v, ok := results[tc.key]
			if !ok {
				t.Skipf("チェック項目 %q が結果に含まれていません", tc.key)
				return
			}
			switch v {
			case "PASS":
				// OK
			case "SKIP":
				t.Skipf("%s: 環境依存でスキップ", tc.key)
			default:
				t.Errorf("%s: %s (got %s)", tc.key, tc.desc, v)
			}
		})
	}

	// 全結果をログに出力（デバッグ用）
	t.Logf("bot検出結果: %s", jsonStr)
}

// TestIgnoreCertErrors は自己署名証明書のHTTPSサーバーに対する動作を検証する。
func TestIgnoreCertErrors(t *testing.T) {
	if testing.Short() {
		t.Skip("統合テストをスキップ（-short指定）")
	}

	ts := newTLSTestServer(t)
	defer ts.Close()

	t.Run("WithoutOption_Fails", func(t *testing.T) {
		fetcher := New(WithStealth(false))
		if err := fetcher.Start(); err != nil {
			t.Fatalf("ブラウザの起動に失敗: %v", err)
		}
		defer fetcher.Close()

		_, err := fetcher.Fetch(context.Background(), ts.URL+"/",
			WithWaitStrategy(WaitLoad))
		if err == nil {
			t.Error("証明書エラーが発生するはずですが、成功しました")
		}
	})

	t.Run("WithOption_Succeeds", func(t *testing.T) {
		fetcher := New(WithStealth(false), WithIgnoreCertErrors(true))
		if err := fetcher.Start(); err != nil {
			t.Fatalf("ブラウザの起動に失敗: %v", err)
		}
		defer fetcher.Close()

		result, err := fetcher.Fetch(context.Background(), ts.URL+"/",
			WithWaitStrategy(WaitLoad))
		if err != nil {
			t.Fatalf("Fetchに失敗: %v", err)
		}
		assertContains(t, result.HTML, "STATIC_CONTENT_MARKER")
	})
}

func assertContains(t *testing.T, html, marker string) {
	t.Helper()
	if !strings.Contains(html, marker) {
		t.Errorf("HTMLに %q が含まれるべきですが、含まれていません (HTML長: %d)", marker, len(html))
	}
}

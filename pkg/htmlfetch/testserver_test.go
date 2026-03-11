package htmlfetch

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestServer は動的コンテンツのテスト用HTTPサーバーを作成する。
// Handlebars風・Riot.js風の動的ページと、それぞれのJSON APIを提供する。
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleStatic)
	mux.HandleFunc("/handlebars", handleHandlebars)
	mux.HandleFunc("/api/handlebars-data", handleHandlebarsAPI)
	mux.HandleFunc("/riotjs", handleRiotJS)
	mux.HandleFunc("/api/riotjs-data", handleRiotJSAPI)
	mux.HandleFunc("/lang-redirect", handleLangRedirect)
	mux.HandleFunc("/lang-redirect/ja", handleLangJa)
	mux.HandleFunc("/lang-redirect/en", handleLangEn)
	mux.HandleFunc("/echo-headers", handleEchoHeaders)
	mux.HandleFunc("/bot-detect", handleBotDetect)
	return httptest.NewServer(mux)
}

// handleStatic は静的ページを返す（ベースラインテスト用）
func handleStatic(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(staticPage))
}

// handleHandlebars はHandlebars風の動的ページを返す。
// テンプレートが<script>タグ内にあり、JSがAPIからデータを取得してDOMに描画する。
func handleHandlebars(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(handlebarsPage))
}

// handleHandlebarsAPI はHandlebarsページ用のJSONデータを返す。
// 200msの遅延でネットワーク遅延をシミュレートする。
func handleHandlebarsAPI(w http.ResponseWriter, r *http.Request) {
	time.Sleep(200 * time.Millisecond)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(handlebarsData))
}

// handleRiotJS はRiot.js風の動的ページを返す。
// カスタムタグがJSによってDOMに描画される。
func handleRiotJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(riotjsPage))
}

// handleRiotJSAPI はRiot.jsページ用のJSONデータを返す。
// 200msの遅延でネットワーク遅延をシミュレートする。
func handleRiotJSAPI(w http.ResponseWriter, r *http.Request) {
	time.Sleep(200 * time.Millisecond)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(riotjsData))
}

// handleLangRedirect はAccept-Languageヘッダーを見て言語別ページにリダイレクトする。
// 早稲田大学サイトのように、ブラウザの言語設定でリダイレクト先が変わる動作を再現する。
func handleLangRedirect(w http.ResponseWriter, r *http.Request) {
	lang := r.Header.Get("Accept-Language")
	if strings.Contains(lang, "ja") {
		http.Redirect(w, r, "/lang-redirect/ja", http.StatusFound)
		return
	}
	http.Redirect(w, r, "/lang-redirect/en", http.StatusFound)
}

func handleLangJa(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(jaPage))
}

func handleLangEn(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(enPage))
}

const jaPage = `<!DOCTYPE html>
<html lang="ja">
<head><title>日本語ページ</title></head>
<body><p>LANG_JA_MARKER</p></body>
</html>`

const enPage = `<!DOCTYPE html>
<html lang="en">
<head><title>English Page</title></head>
<body><p>LANG_EN_MARKER</p></body>
</html>`

// handleEchoHeaders はリクエストのUser-AgentとAccept-LanguageをHTML内に埋め込んで返す。
func handleEchoHeaders(w http.ResponseWriter, r *http.Request) {
	ua := r.Header.Get("User-Agent")
	lang := r.Header.Get("Accept-Language")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html><body>
<p id="ua">%s</p>
<p id="lang">%s</p>
</body></html>`, ua, lang)
}

// handleBotDetect はブラウザのbot検出チェックを実行し、結果をJSON形式でHTMLに埋め込むページを返す。
// 各チェックはbot検出サイト（bot.sannysoft.com, CreepJS等）で使われている手法を再現している。
func handleBotDetect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(botDetectPage))
}

// --- bot検出ページ ---
// 各チェックはクライアントサイドJSで実行し、結果をJSON形式で <pre id="results"> に出力する。
// go-rod/stealth が対策済みの項目に絞り、ヘッドレスChromeで検出可能な代表的手法をカバーする。
//
// チェック項目と出典:
//
// 1. navigator.webdriver
//    - 自動化ブラウザでは true、通常ブラウザでは undefined
//    - 出典: W3C WebDriver spec (https://www.w3.org/TR/webdriver/#dom-navigatorautomationinformation-webdriver)
//    - 参考: bot.sannysoft.com, puppeteer-extra-plugin-stealth "navigator.webdriver" evasion
//
// 2. window.chrome 存在チェック
//    - ヘッドレスChromeでは window.chrome が未定義
//    - 出典: bot.sannysoft.com "Chrome" test
//    - 参考: puppeteer-extra-plugin-stealth "chrome.app", "chrome.csi", "chrome.loadTimes", "chrome.runtime" evasions
//
// 3. window.chrome.app
//    - 通常Chromeでは isInstalled, getDetails 等のプロパティを持つオブジェクト
//    - 出典: puppeteer-extra-plugin-stealth "chrome.app" evasion
//    - 参考: https://nickchrysler.com/are-you-chrome/
//
// 4. window.chrome.csi
//    - 通常Chromeでは performance.timing 相当のデータを返す関数
//    - 出典: puppeteer-extra-plugin-stealth "chrome.csi" evasion
//    - 参考: bot.sannysoft.com "Chrome (chrome.csi)" test
//
// 5. window.chrome.loadTimes
//    - 通常Chromeではページ読み込みタイミング情報を返す関数
//    - 出典: puppeteer-extra-plugin-stealth "chrome.loadTimes" evasion
//    - 参考: bot.sannysoft.com "Chrome (chrome.loadTimes)" test
//
// 6. navigator.plugins
//    - ヘッドレスChromeでは空配列、通常Chromeでは "Chrome PDF Plugin" 等が存在
//    - 出典: bot.sannysoft.com "Plugins" test
//    - 参考: puppeteer-extra-plugin-stealth "navigator.plugins" evasion
//
// 7. navigator.languages
//    - ヘッドレスChromeでは空または未定義になる場合がある
//    - 出典: bot.sannysoft.com "Languages" test
//    - 参考: puppeteer-extra-plugin-stealth "navigator.languages" evasion
//
// 8. window.outerWidth / window.outerHeight
//    - ヘッドレスChromeでは物理ウィンドウがないため 0 になる
//    - 出典: bot.sannysoft.com "Outer dimensions" test
//    - 参考: puppeteer-extra-plugin-stealth "window.outerdimensions" evasion
//
// 9. Notification.permission
//    - ヘッドレスChromeでは "denied" を返すが、通常Chromeでは "default"
//    - 出典: puppeteer-extra-plugin-stealth "navigator.permissions" evasion
//    - 参考: CreepJS permissions check
//
// 10. User-Agent文字列
//     - ヘッドレスChromeでは "HeadlessChrome" を含む
//     - 出典: bot.sannysoft.com "User Agent" test
//     - 参考: puppeteer-extra-plugin-stealth "user-agent-override" evasion
//
// 11. WebGL vendor/renderer
//     - ヘッドレスChromeではソフトウェアレンダリング "Google SwiftShader" が返る
//     - 出典: CreepJS WebGL fingerprint, bot.sannysoft.com "WebGL Vendor" test
//     - 参考: puppeteer-extra-plugin-stealth "webgl.vendor" evasion
//
// 12. Function.prototype.toString 整合性
//     - stealth等でProxy書き換えした関数が toString() で [native code] を返すか検証
//     - 出典: CreepJS "lies" detection, intoli.com/blog/not-possible-to-block-chrome-headless
//     - 参考: puppeteer-extra-plugin-stealth "Function.prototype.toString" evasion
const botDetectPage = `<!DOCTYPE html>
<html>
<head><title>Bot Detection Test</title></head>
<body>
<pre id="results"></pre>
<script>
(function() {
  var results = {};

  // 1. navigator.webdriver
  results.webdriver = (navigator.webdriver === true) ? "FAIL" : "PASS";

  // 2. window.chrome
  results.chrome_exists = (typeof window.chrome !== "undefined") ? "PASS" : "FAIL";

  // 3. window.chrome.app
  results.chrome_app = (window.chrome && typeof window.chrome.app !== "undefined") ? "PASS" : "FAIL";

  // 4. window.chrome.csi
  results.chrome_csi = (window.chrome && typeof window.chrome.csi === "function") ? "PASS" : "FAIL";

  // 5. window.chrome.loadTimes
  results.chrome_loadTimes = (window.chrome && typeof window.chrome.loadTimes === "function") ? "PASS" : "FAIL";

  // 6. navigator.plugins
  results.plugins = (navigator.plugins && navigator.plugins.length > 0) ? "PASS" : "FAIL";

  // 7. navigator.languages
  results.languages = (navigator.languages && navigator.languages.length > 0) ? "PASS" : "FAIL";

  // 8. window.outerWidth / outerHeight
  results.outer_dimensions = (window.outerWidth > 0 && window.outerHeight > 0) ? "PASS" : "FAIL";

  // 9. Notification.permission
  try {
    results.notification_permission = (Notification.permission === "denied") ? "FAIL" : "PASS";
  } catch(e) {
    results.notification_permission = "SKIP";
  }

  // 10. User-Agent
  results.user_agent = (navigator.userAgent.indexOf("HeadlessChrome") === -1) ? "PASS" : "FAIL";

  // 11. WebGL vendor/renderer
  try {
    var canvas = document.createElement("canvas");
    var gl = canvas.getContext("webgl");
    if (gl) {
      var debugInfo = gl.getExtension("WEBGL_debug_renderer_info");
      if (debugInfo) {
        var vendor = gl.getParameter(debugInfo.UNMASKED_VENDOR_WEBGL);
        var renderer = gl.getParameter(debugInfo.UNMASKED_RENDERER_WEBGL);
        results.webgl_vendor = vendor;
        results.webgl_renderer = renderer;
        results.webgl = (renderer.indexOf("SwiftShader") === -1) ? "PASS" : "FAIL";
      } else {
        results.webgl = "SKIP";
      }
    } else {
      results.webgl = "SKIP";
    }
  } catch(e) {
    results.webgl = "SKIP";
  }

  // 12. Function.prototype.toString 整合性
  // navigator.permissionsのqueryがnative codeとして見えるか検証
  try {
    var fnStr = Permissions.prototype.query.toString();
    results.fn_toString = (fnStr.indexOf("[native code]") !== -1) ? "PASS" : "FAIL";
  } catch(e) {
    results.fn_toString = "SKIP";
  }

  document.getElementById("results").textContent = JSON.stringify(results);
})();
</script>
</body>
</html>`

// --- 静的ページ ---

const staticPage = `<!DOCTYPE html>
<html>
<head><title>Static Test</title></head>
<body>
  <div id="content">
    <p>STATIC_CONTENT_MARKER</p>
  </div>
</body>
</html>`

// --- Handlebars風ページ ---
// テンプレートが<script>タグ内にあり、JSがfetchでデータ取得→DOMに描画する。
// 初期HTMLには DYNAMIC_HANDLEBARS_TITLE_* は含まれない。

const handlebarsPage = `<!DOCTYPE html>
<html>
<head><title>Handlebars Test</title></head>
<body>
  <div id="news-container">
    <!-- JSが描画するまで空 -->
  </div>

  <script type="text/x-handlebars-template" id="news-template">
    <ul>
      {{#each items}}
      <li class="news-item"><a href="{{url}}">{{title}}</a><p>{{body}}</p></li>
      {{/each}}
    </ul>
  </script>

  <script>
    (async function() {
      var resp = await fetch('/api/handlebars-data');
      var data = await resp.json();
      var container = document.getElementById('news-container');
      var html = '<ul>';
      data.items.forEach(function(item) {
        html += '<li class="news-item"><a href="' + item.url + '">' + item.title + '</a><p>' + item.body + '</p></li>';
      });
      html += '</ul>';
      container.innerHTML = html;
    })();
  </script>
</body>
</html>`

const handlebarsData = `{
  "items": [
    {"title": "DYNAMIC_HANDLEBARS_TITLE_1", "body": "DYNAMIC_HANDLEBARS_BODY_1", "url": "/news/1"},
    {"title": "DYNAMIC_HANDLEBARS_TITLE_2", "body": "DYNAMIC_HANDLEBARS_BODY_2", "url": "/news/2"},
    {"title": "DYNAMIC_HANDLEBARS_TITLE_3", "body": "DYNAMIC_HANDLEBARS_BODY_3", "url": "/news/3"}
  ]
}`

// --- Riot.js風ページ ---
// カスタムタグ<news-list>がJSによってDOMに描画される。
// 初期HTMLには DYNAMIC_RIOT_TOPIC_* は含まれない。

const riotjsPage = `<!DOCTYPE html>
<html>
<head><title>Riot.js Test</title></head>
<body>
  <div id="app">
    <news-list></news-list>
  </div>

  <script>
    (async function() {
      var resp = await fetch('/api/riotjs-data');
      var data = await resp.json();
      var app = document.getElementById('app');
      var html = '<ul class="news-list">';
      data.topics.forEach(function(topic) {
        html += '<li class="topic-item">' +
                '<a href="' + topic.url + '">' + topic.title + '</a>' +
                '<span class="date">' + topic.date + '</span>' +
                '</li>';
      });
      html += '</ul>';
      app.innerHTML = html;
    })();
  </script>
</body>
</html>`

const riotjsData = `{
  "topics": [
    {"title": "DYNAMIC_RIOT_TOPIC_1", "url": "/topics/1", "date": "2025-01-15"},
    {"title": "DYNAMIC_RIOT_TOPIC_2", "url": "/topics/2", "date": "2025-01-10"},
    {"title": "DYNAMIC_RIOT_TOPIC_3", "url": "/topics/3", "date": "2025-01-05"}
  ]
}`

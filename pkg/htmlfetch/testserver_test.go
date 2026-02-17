package htmlfetch

import (
	"net/http"
	"net/http/httptest"
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

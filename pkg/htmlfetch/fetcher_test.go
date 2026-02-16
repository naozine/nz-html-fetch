package htmlfetch

import (
	"testing"
)

// TestCreatePage_ClosedBrowser_Stealth はブラウザ接続が切れた状態で
// createPage(stealth=true)がパニックせずエラーを返すことを検証する。
// 修正前: MustPage がパニックしてテストがクラッシュする
// 修正後: エラーが返されてテストがパスする
func TestCreatePage_ClosedBrowser_Stealth(t *testing.T) {
	f := New(WithStealth(true))

	browser, err := f.launchBrowser()
	if err != nil {
		t.Fatalf("ブラウザの起動に失敗: %v", err)
	}

	// ブラウザ接続を閉じてから createPage を呼ぶ
	browser.MustClose()

	_, err = f.createPage(browser)
	if err == nil {
		t.Fatal("閉じたブラウザでのページ作成でエラーが返されるべき")
	}

	t.Logf("期待通りエラーが返された: %v", err)
}

// TestCreatePage_ClosedBrowser_NonStealth はブラウザ接続が切れた状態で
// createPage(stealth=false)がパニックせずエラーを返すことを検証する。
func TestCreatePage_ClosedBrowser_NonStealth(t *testing.T) {
	f := New(WithStealth(false))

	browser, err := f.launchBrowser()
	if err != nil {
		t.Fatalf("ブラウザの起動に失敗: %v", err)
	}

	// ブラウザ接続を閉じてから createPage を呼ぶ
	browser.MustClose()

	_, err = f.createPage(browser)
	if err == nil {
		t.Fatal("閉じたブラウザでのページ作成でエラーが返されるべき")
	}

	t.Logf("期待通りエラーが返された: %v", err)
}

// TestMustConnect_ClosedConnection は無効なURLでの MustConnect が
// パニックせずエラーを返すことを検証する。
func TestMustConnect_ClosedConnection(t *testing.T) {
	f := New()

	browser, err := f.launchBrowser()
	if err != nil {
		t.Fatalf("ブラウザの起動に失敗: %v", err)
	}
	browser.MustClose()

	// 閉じたブラウザで Fetch を呼ぶと、高速モードでは createPage が呼ばれる
	f.browser = browser
	f.started = true

	_, err = f.Fetch(nil, "https://example.com")
	if err == nil {
		t.Fatal("閉じたブラウザでの Fetch でエラーが返されるべき")
	}

	t.Logf("期待通りエラーが返された: %v", err)
}

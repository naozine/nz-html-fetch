package htmlfetch

import "strings"

// adDomains は広告・トラッキング関連のドメインリスト
var adDomains = []string{
	// Google広告
	"googleadservices.com",
	"googlesyndication.com",
	"doubleclick.net",
	"google-analytics.com",
	"googletagmanager.com",
	"googletagservices.com",
	"adservice.google.com",
	"pagead2.googlesyndication.com",
	// Facebook/Meta
	"facebook.com/tr",
	"connect.facebook.net",
	"facebook.net",
	"fbcdn.net",
	// Amazon
	"amazon-adsystem.com",
	"assoc-amazon.com",
	// Microsoft/Bing
	"bat.bing.com",
	"ads.microsoft.com",
	// Twitter/X
	"ads-twitter.com",
	"ads.twitter.com",
	"analytics.twitter.com",
	// Yahoo
	"ads.yahoo.com",
	"analytics.yahoo.com",
	"yimg.com/cv/",
	// 汎用広告ネットワーク
	"adnxs.com",
	"adsrvr.org",
	"adform.net",
	"criteo.com",
	"criteo.net",
	"outbrain.com",
	"taboola.com",
	"pubmatic.com",
	"rubiconproject.com",
	"openx.net",
	"casalemedia.com",
	"advertising.com",
	"adcolony.com",
	"unity3d.com/ads",
	"applovin.com",
	"mopub.com",
	"inmobi.com",
	"smaato.com",
	"chartboost.com",
	"vungle.com",
	"ironsource.com",
	// トラッキング・アナリティクス
	"hotjar.com",
	"fullstory.com",
	"mouseflow.com",
	"crazyegg.com",
	"clicktale.com",
	"mixpanel.com",
	"amplitude.com",
	"segment.com",
	"segment.io",
	"optimizely.com",
	"newrelic.com",
	"nr-data.net",
	"sentry.io",
	"bugsnag.com",
	// ソーシャルウィジェット・シェアボタン
	"addthis.com",
	"addtoany.com",
	"sharethis.com",
	// ポップアップ・プッシュ通知
	"onesignal.com",
	"pushwoosh.com",
	"pushnami.com",
	// その他広告関連
	"2mdn.net",
	"serving-sys.com",
	"moatads.com",
	"adzerk.net",
	"adversal.com",
	"adblade.com",
	"revcontent.com",
	"mgid.com",
	"contentad.net",
	"adroll.com",
	"perfectaudience.com",
	"retargeter.com",
	// 日本の広告ネットワーク
	"i-mobile.co.jp",
	"microad.jp",
	"ad-stir.com",
	"geniee.co.jp",
	"logly.co.jp",
	"fluct.jp",
	"adingo.jp",
	"yads.yahoo.co.jp",
	"yads.c.yimg.jp",
}

// isAdURL は広告ドメインのURLかどうかをチェックする
func isAdURL(rawURL string) bool {
	for _, domain := range adDomains {
		if strings.Contains(rawURL, domain) {
			return true
		}
	}
	return false
}

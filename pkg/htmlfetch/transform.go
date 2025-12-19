package htmlfetch

import "github.com/go-rod/rod"

// embedCSS は外部CSSをインラインスタイルに埋め込む
func embedCSS(page *rod.Page) error {
	_, err := page.Evaluate(&rod.EvalOptions{
		JS: `(async () => {
			const links = document.querySelectorAll('link[rel="stylesheet"]');
			for (const link of links) {
				try {
					const href = link.href;
					if (!href) continue;
					const response = await fetch(href);
					if (response.ok) {
						const css = await response.text();
						const style = document.createElement('style');
						style.textContent = css;
						style.setAttribute('data-original-href', href);
						link.parentNode.replaceChild(style, link);
					}
				} catch (e) {
					console.warn('Failed to embed CSS:', e);
				}
			}
		})()`,
		AwaitPromise: true,
	})
	return err
}

// stripScripts はスクリプトを除去する
func stripScripts(page *rod.Page) error {
	_, err := page.Evaluate(&rod.EvalOptions{
		JS: `(() => {
			// <script>タグを削除
			document.querySelectorAll('script').forEach(el => el.remove());

			// イベントハンドラ属性を削除
			const eventAttrs = ['onclick', 'onload', 'onerror', 'onmouseover', 'onmouseout',
				'onkeydown', 'onkeyup', 'onsubmit', 'onchange', 'onfocus', 'onblur'];
			document.querySelectorAll('*').forEach(el => {
				eventAttrs.forEach(attr => el.removeAttribute(attr));
			});

			// javascript: URLを無効化
			document.querySelectorAll('a[href^="javascript:"]').forEach(el => {
				el.removeAttribute('href');
			});
		})()`,
	})
	return err
}

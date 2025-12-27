package htmlfetch

import (
	"net/url"
	"strings"

	htmltomarkdown "github.com/JohannesKaufmann/html-to-markdown/v2"
	"github.com/go-shiori/go-readability"
)

// convertToMarkdown はHTMLからコンテンツを抽出しMarkdownに変換する
func convertToMarkdown(html string, pageURL string) (string, error) {
	// URLをパース
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		return "", &FetchError{
			Code:    ErrInternalError,
			Message: "URLのパースに失敗しました",
			Cause:   err,
		}
	}

	// Readabilityでコンテンツ抽出
	article, err := readability.FromReader(strings.NewReader(html), parsedURL)
	if err != nil {
		return "", &FetchError{
			Code:    ErrInternalError,
			Message: "コンテンツ抽出に失敗しました",
			Cause:   err,
		}
	}

	// html-to-markdownで変換
	markdown, err := htmltomarkdown.ConvertString(article.Content)
	if err != nil {
		return "", &FetchError{
			Code:    ErrInternalError,
			Message: "Markdown変換に失敗しました",
			Cause:   err,
		}
	}

	// タイトルがあれば先頭に追加
	if article.Title != "" {
		markdown = "# " + article.Title + "\n\n" + markdown
	}

	return markdown, nil
}

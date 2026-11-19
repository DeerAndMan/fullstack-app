package service

import (
	"context"
	"encoding/json"
	"html"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const sinaFlashURL = "https://finance.sina.com.cn/7x24/"
const eastmoneyFlashURL = "https://kuaixun.eastmoney.com/"
const kaipanlaNewsURL = "https://www.kaipanla.com/page"
const thsStockNewsURL = "https://stock.10jqka.com.cn/"

var flashMarkup = regexp.MustCompile(`<[^>]+>`)

func flashTitle(content string) string {
	text := strings.Join(strings.Fields(html.UnescapeString(flashMarkup.ReplaceAllString(content, " "))), " ")
	if rest, ok := strings.CutPrefix(text, "【"); ok {
		if title, _, found := strings.Cut(rest, "】"); found && strings.TrimSpace(title) != "" {
			text = strings.TrimSpace(title)
		}
	}
	runes := []rune(text)
	if len(runes) > 76 {
		return string(runes[:76]) + "…"
	}
	return text
}

func sinaFlashArticle(raw string) string {
	location, err := url.Parse(raw)
	if err == nil && location.Scheme == "https" && location.User == nil {
		switch location.Host {
		case "finance.sina.cn", "finance.sina.com.cn", "wap.cj.sina.cn":
			return location.String()
		}
	}
	return sinaFlashURL
}

func (s *MarketService) sinaFlash(ctx context.Context, _ string) (MarketDataset, error) {
	type reply struct {
		Result struct {
			Status struct{ Code int }
			Data   struct {
				Feed struct {
					List []struct {
						ID        int64
						Text      string `json:"rich_text"`
						Published string `json:"create_time"`
						DocURL    string `json:"docurl"`
					}
				}
			}
		}
	}

	data, err := marketJSON[reply](ctx, s, "https://zhibo.sina.com.cn/api/zhibo/feed", url.Values{
		"page": {"1"}, "page_size": {"24"}, "zhibo_id": {"152"}, "type": {"0"},
	})
	if err != nil || data.Result.Status.Code != 0 || data.Result.Data.Feed.List == nil {
		return MarketDataset{}, errMarketData
	}

	out := MarketDataset{Source: "新浪财经", SourceURL: sinaFlashURL}
	for _, item := range data.Result.Data.Feed.List {
		title := flashTitle(item.Text)
		if _, err := time.ParseInLocation(time.DateTime, item.Published, marketZone); err != nil || title == "" || item.ID <= 0 {
			continue
		}
		out.Rows = append(out.Rows, MarketRow{
			Code: strconv.FormatInt(item.ID, 10), Name: title, Date: item.Published, URL: sinaFlashArticle(item.DocURL),
		})
	}
	return out, nil
}

func (s *MarketService) eastmoneyFlash(ctx context.Context, _ string) (MarketDataset, error) {
	type reply struct {
		Code json.RawMessage
		Data struct {
			Items []struct {
				Code, Title, Summary string
				Published            string `json:"showTime"`
			} `json:"fastNewsList"`
		}
	}

	data, err := marketJSON[reply](ctx, s, "https://np-weblist.eastmoney.com/comm/web/getFastNewsList", url.Values{
		"client": {"web"}, "biz": {"web_724"}, "fastColumn": {"101"},
		"pageSize": {"24"}, "pageIndex": {"0"}, "sortEnd": {""}, "req_trace": {"web"},
	})
	code := marketJSONNumber(data.Code)
	if err != nil || code == nil || *code != 1 || data.Data.Items == nil {
		return MarketDataset{}, errMarketData
	}

	out := MarketDataset{Source: "东方财富", SourceURL: eastmoneyFlashURL}
	for _, item := range data.Data.Items {
		title := flashTitle(item.Title)
		if title == "" {
			title = flashTitle(item.Summary)
		}
		if _, err := time.ParseInLocation(time.DateTime, item.Published, marketZone); err != nil || title == "" {
			continue
		}
		if _, err := strconv.ParseUint(item.Code, 10, 64); err != nil {
			continue
		}
		out.Rows = append(out.Rows, MarketRow{
			Code: item.Code, Name: title, Date: item.Published, URL: "https://finance.eastmoney.com/a/" + item.Code + ".html",
		})
	}
	sort.SliceStable(out.Rows, func(i, j int) bool { return out.Rows[i].Date > out.Rows[j].Date })
	return out, nil
}

func (s *MarketService) kaipanlaFlash(ctx context.Context, _ string) (MarketDataset, error) {
	type reply struct {
		Errcode string
		List    []struct {
			CID, Time, Title, Content, Source string
		}
	}
	data, err := marketJSON[reply](ctx, s, "https://apparticle.longhuvip.com/w1/api/index.php", url.Values{
		"c": {"PCNewsFlash"}, "a": {"GetList"}, "Index": {"0"}, "st": {"24"},
	})
	if err != nil || data.Errcode != "0" || data.List == nil {
		return MarketDataset{}, errMarketData
	}
	out := MarketDataset{Source: "开盘啦", SourceURL: kaipanlaNewsURL}
	for _, item := range data.List {
		ts, err := strconv.ParseInt(item.Time, 10, 64)
		title := flashTitle(item.Title)
		if title == "" {
			title = flashTitle(item.Content)
		}
		if err != nil || ts <= 0 || item.CID == "" || title == "" {
			continue
		}
		out.Rows = append(out.Rows, MarketRow{
			Code: item.CID, Name: title, Date: time.Unix(ts, 0).In(marketZone).Format(time.DateTime),
			URL: kaipanlaNewsURL, Detail: item.Source,
		})
	}
	return out, nil
}

var thsStockArticle = regexp.MustCompile(`(?is)<a\b[^>]*\bhref=["']https?://((?:stock|news|fund)\.10jqka\.com\.cn/(\d{8})/c(\d+)\.shtml)["'][^>]*>(.*?)</a>`)

func (s *MarketService) thsStockNews(ctx context.Context, _ string) (MarketDataset, error) {
	raw, err := s.request(ctx, thsStockNewsURL, nil)
	if err != nil {
		return MarketDataset{}, err
	}
	page, err := decodeTHS(raw)
	if err != nil {
		return MarketDataset{}, err
	}
	out := MarketDataset{
		Source: "同花顺", SourceURL: thsStockNewsURL,
		Message: "同花顺股票频道要闻，按发布日期展示。",
	}
	seen := make(map[string]bool)
	for _, match := range thsStockArticle.FindAllStringSubmatch(page, -1) {
		title := flashTitle(match[4])
		date, err := time.Parse("20060102", match[2])
		if err != nil || title == "" || seen[match[3]] {
			continue
		}
		seen[match[3]] = true
		out.Rows = append(out.Rows, MarketRow{
			Code: match[3], Name: title, Date: date.Format(time.DateOnly), URL: "https://" + match[1],
		})
	}
	if len(out.Rows) == 0 {
		return MarketDataset{}, errMarketData
	}
	sort.SliceStable(out.Rows, func(i, j int) bool { return out.Rows[i].Date > out.Rows[j].Date })
	return out, nil
}

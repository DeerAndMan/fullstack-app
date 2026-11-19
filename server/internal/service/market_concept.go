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

const sinaMarketCenter = "https://vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php/Market_Center."

func (s *MarketService) conceptMembers(ctx context.Context, code string) (MarketDataset, error) {
	return s.loadSinaBoardMembers(ctx, code, false)
}

func (s *MarketService) loadSinaBoardMembers(ctx context.Context, code string, funds bool) (MarketDataset, error) {
	count, err := marketJSON[string](ctx, s, sinaMarketCenter+"getHQNodeStockCount", url.Values{"node": {code}})
	if err != nil {
		return MarketDataset{}, err
	}
	total, err := strconv.Atoi(strings.TrimSpace(count))
	if err != nil || total < 0 {
		return MarketDataset{}, errMarketData
	}

	type member struct {
		Code, Name, Symbol           string
		Trade, Changepercent, Amount json.RawMessage
		Changeratio, Netamount       json.RawMessage
	}

	out := MarketDataset{Source: "新浪财经", Total: &total}
	const pageSize = 80
	seen := make(map[string]bool)
	endpoint, node := sinaMarketCenter+"getHQNodeData", "node"
	if funds {
		endpoint, node = sinaMoneyFlow+"ssl_bkzj_ssggzj", "bankuai"
	}
	for page := 1; (page-1)*pageSize < total && page <= 10; page++ {
		list, err := marketJSON[[]member](ctx, s, endpoint, url.Values{
			"page": {strconv.Itoa(page)}, "num": {strconv.Itoa(pageSize)}, "sort": {"symbol"}, "asc": {"1"}, node: {code},
		})
		if err != nil {
			return MarketDataset{}, err
		}
		for _, row := range list {
			if funds && marketStockCode.MatchString(row.Symbol) {
				row.Code = row.Symbol[2:]
			}
			if row.Code == "" || row.Name == "" || seen[row.Code] {
				continue
			}
			seen[row.Code] = true
			price := marketJSONNumber(row.Trade)
			change := marketJSONNumber(row.Changepercent)
			if funds {
				change = marketJSONNumber(row.Changeratio)
				if change != nil {
					*change *= 100
				}
			}
			if price != nil && *price <= 0 {
				price, change = nil, nil
			}
			if !marketStockCode.MatchString(row.Symbol) {
				row.Symbol = stockSymbol(row.Code)
			}
			out.Rows = append(out.Rows, MarketRow{
				Code: row.Code, Name: row.Name, Price: price, Change: change, Amount: marketJSONNumber(row.Amount),
				Symbol: row.Symbol, Net: marketJSONNumber(row.Netamount),
			})
		}
		if len(list) < pageSize {
			break
		}
	}
	if len(out.Rows) < total {
		out.Message = "行情源只返回了部分成分股。"
	}
	return out, nil
}

type conceptSearchItem struct {
	Title   string
	Text    string `json:"content_text"`
	Snippet string `json:"content_short"`
	URL     string `json:"uri"`
	Time    int64  `json:"display_time"`
}

var searchHighlight = regexp.MustCompile(`(?i)</?em>`)

func conceptSearchText(value string) string {
	return strings.Join(strings.Fields(html.UnescapeString(searchHighlight.ReplaceAllString(value, ""))), " ")
}

func conceptSearchURL(value string) string {
	link, err := url.Parse(value)
	if err != nil || link.Scheme != "https" || link.Host != "wallstreetcn.com" || link.User != nil {
		return ""
	}
	if !strings.HasPrefix(link.Path, "/articles/") && !strings.HasPrefix(link.Path, "/livenews/") &&
		!strings.HasPrefix(link.Path, "/premium/articles/") {
		return ""
	}
	link.RawQuery = ""
	link.ForceQuery = false
	link.Fragment = ""
	return link.String()
}

func searchHeadline(item conceptSearchItem, keyword string) string {
	title := conceptSearchText(item.Title)
	body := conceptSearchText(item.Text + " " + item.Snippet)
	if !strings.Contains(title, keyword) && !strings.Contains(body, keyword) {
		return ""
	}
	if title != "" {
		return title
	}
	// 无独立标题的快讯只截取包含题材词的附近文本。
	loc := strings.Index(body, keyword)
	before := []rune(body[:loc])
	if len(before) > 16 {
		before = append([]rune{'…'}, before[len(before)-16:]...)
	}
	runes := append(before, []rune(body[loc:])...)
	if len(runes) > 75 {
		runes = append(runes[:75], '…')
	}
	return string(runes)
}

func (s *MarketService) conceptNews(ctx context.Context, name string) (MarketDataset, error) {
	type result struct {
		Code int
		Data struct {
			Items []conceptSearchItem
		}
	}

	keyword := strings.TrimSuffix(strings.TrimSuffix(strings.TrimSpace(name), "概念股"), "概念")
	if keyword == "" {
		keyword = name
	}
	searchURL := "https://wallstreetcn.com/search?" + url.Values{"q": {keyword}}.Encode()
	out := MarketDataset{Source: "华尔街见闻", SourceURL: searchURL}
	seenTitles, seenURLs := map[string]bool{}, map[string]bool{}
	failed := 0
	for _, category := range []string{"live", "article"} {
		searchCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		response, err := marketJSON[result](searchCtx, s, wallstreetAPI+"/search/"+category,
			url.Values{"query": {keyword}, "limit": {"12"}})
		cancel()
		if err != nil || response.Code != 20000 || response.Data.Items == nil {
			failed++
			continue
		}

		for _, item := range response.Data.Items {
			link := conceptSearchURL(item.URL)
			title := searchHeadline(item, keyword)
			if link == "" || title == "" || item.Time <= 0 || seenURLs[link] || seenTitles[title] {
				continue
			}
			seenTitles[title] = true
			seenURLs[link] = true
			out.Rows = append(out.Rows, MarketRow{
				Code: link, URL: link, Name: title, Date: time.Unix(item.Time, 0).In(marketZone).Format("2006-01-02 15:04:05"),
			})
		}
	}
	if failed == 2 {
		return MarketDataset{}, errMarketData
	}
	if failed > 0 {
		out.Message = "部分消息检索暂不可用，以下为本次取得的结果。"
	}
	sort.SliceStable(out.Rows, func(i, j int) bool { return out.Rows[i].Date > out.Rows[j].Date })
	return out, nil
}

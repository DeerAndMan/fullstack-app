package service

import (
	"bytes"
	"context"
	"html"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
)

const thsCatalogURL = "https://q.10jqka.com.cn/gn/"
const sinaCatalogURL = "https://vip.stock.finance.sina.com.cn/mkt/"

var thsConceptLink = regexp.MustCompile(`(?i)<a\s[^>]*href=["'](?:https?:)?//q\.10jqka\.com\.cn/gn/detail/code/(\d{6})/?["'][^>]*>\s*([^<]+)\s*</a>`)
var thsStockLink = regexp.MustCompile(`(?i)<a\s[^>]*href=["'](?:https?:)?//stockpage\.10jqka\.com\.cn/(\d{6})/?["'][^>]*>\s*([^<]+)\s*</a>`)
var researchStockCode = regexp.MustCompile(`^\d{6}$`)
var researchQuoteID = regexp.MustCompile(`^\d+\.\d{6}$`)

func (s *MarketService) researchQuote(ctx context.Context, code string) (string, string) {
	_, name, link := s.researchNamedQuote(ctx, code, "")
	return name, link
}

func (s *MarketService) researchNamedQuote(ctx context.Context, code, name string) (string, string, string) {
	type search struct {
		QuotationCodeTable struct {
			Data []struct{ Code, Name, Classify, QuoteID string }
		}
	}
	input := name
	if researchStockCode.MatchString(code) {
		input = code
	} else {
		code = ""
	}
	if input != "" {
		data, err := marketJSON[search](ctx, s, "https://searchapi.eastmoney.com/api/suggest/get",
			url.Values{"input": {input}, "type": {"14"}, "count": {"8"}})
		if err == nil {
			for _, row := range data.QuotationCodeTable.Data {
				if row.Classify == "AStock" && researchStockCode.MatchString(row.Code) &&
					(code == "" || row.Code == code) && (name == "" || row.Name == name) &&
					researchQuoteID.MatchString(row.QuoteID) && strings.HasSuffix(row.QuoteID, "."+row.Code) {
					return row.Code, row.Name, "https://quote.eastmoney.com/unify/r/" + row.QuoteID
				}
			}
		}
	}
	if input != name && name != "" {
		return s.researchNamedQuote(ctx, "", name)
	}
	return "", "", ""
}

func decodeTHS(body []byte) (string, error) {
	if bytes.Contains(bytes.ToLower(body[:min(len(body), 1800)]), []byte("charset=\"utf-8\"")) {
		return string(body), nil
	}
	text, err := simplifiedchinese.GB18030.NewDecoder().Bytes(body)
	return string(text), err
}

func (s *MarketService) thsCatalog(ctx context.Context, _ string) (MarketDataset, error) {
	body, err := s.request(ctx, thsCatalogURL, nil)
	if err != nil {
		return MarketDataset{}, err
	}
	text, err := decodeTHS(body)
	if err != nil {
		return MarketDataset{}, err
	}
	out := MarketDataset{Source: "同花顺", SourceURL: thsCatalogURL}
	seen := map[string]bool{}
	for _, match := range thsConceptLink.FindAllStringSubmatch(text, -1) {
		code := match[1]
		name := strings.TrimSpace(html.UnescapeString(match[2]))
		if name == "" || seen[code] {
			continue
		}
		seen[code] = true
		out.Rows = append(out.Rows, MarketRow{Code: code, Name: name})
	}
	if len(out.Rows) == 0 {
		return MarketDataset{}, errMarketData
	}
	return out, nil
}

func (s *MarketService) thsMembers(ctx context.Context, code string) (MarketDataset, error) {
	endpoint := thsCatalogURL + "detail/code/" + code + "/"
	body, err := s.request(ctx, endpoint, nil)
	if err != nil {
		return MarketDataset{}, err
	}
	text, err := decodeTHS(body)
	if err != nil {
		return MarketDataset{}, err
	}
	from := strings.Index(text, `id="maincont"`)
	if from < 0 {
		return MarketDataset{}, errMarketData
	}
	table, _, ok := strings.Cut(text[from:], "</table>")
	if !ok {
		return MarketDataset{}, errMarketData
	}
	out := MarketDataset{Source: "同花顺", SourceURL: endpoint, Message: "只核验公开成分股列表首屏，缺席不代表未收录。"}
	seen := map[string]bool{}
	for _, match := range thsStockLink.FindAllStringSubmatch(table, -1) {
		stock := match[1]
		name := strings.TrimSpace(html.UnescapeString(match[2]))
		if name == "" || name == stock || seen[stock] {
			continue
		}
		seen[stock] = true
		out.Rows = append(out.Rows, MarketRow{Code: stock, Name: name})
	}
	if len(out.Rows) == 0 {
		return MarketDataset{}, errMarketData
	}
	return out, nil
}

func normalizeResearchTopic(text string) string {
	text = strings.ToLower(strings.Join(strings.Fields(text), ""))
	for _, suffix := range []string{"概念股", "供应链", "产业链", "概念", "板块", "链"} {
		text = strings.TrimSuffix(text, suffix)
	}
	return text
}

type topicSelection struct {
	Name string
	Key  string
}

func researchTopics(question string, available map[string]topicSelection) []topicSelection {
	compact := strings.ToLower(strings.Join(strings.Fields(question), ""))
	type match struct {
		key string
		at  int
	}
	var matches []match
	for key := range available {
		if len([]rune(key)) >= 2 {
			if at := strings.Index(compact, key); at >= 0 {
				matches = append(matches, match{key, at})
			}
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		if matches[i].at != matches[j].at {
			return matches[i].at < matches[j].at
		}
		return len(matches[i].key) > len(matches[j].key)
	})
	var selected []topicSelection
	end := 0
	for _, match := range matches {
		if match.at >= end && len(selected) < 4 {
			selected = append(selected, available[match.key])
			end = match.at + len(match.key)
		}
	}
	return selected
}

type researchBoard struct {
	Provider, Topic, Actual, URL, Note string
	Rows                               []MarketRow
}

func (s *TopicResearchService) researchBoards(ctx context.Context, topics []topicSelection, sina, ths map[string]MarketRow) []researchBoard {
	type job struct {
		provider, topic, key string
		row                  MarketRow
	}
	var jobs []job
	for _, topic := range topics {
		jobs = append(jobs,
			job{"新浪财经", topic.Name, topic.Key, sina[topic.Key]},
			job{"同花顺", topic.Name, topic.Key, ths[topic.Key]},
		)
	}
	completed := make(chan struct {
		index int
		board researchBoard
	}, len(jobs))
	for i, job := range jobs {
		go func() {
			board := researchBoard{Provider: job.provider, Topic: job.topic, Actual: job.row.Name}
			if job.provider == "新浪财经" {
				board.URL = sinaCatalogURL
				if job.row.Code != "" {
					board.URL += "#" + url.PathEscape(job.row.Code)
				}
			} else {
				board.URL = thsCatalogURL
			}
			if job.row.Code == "" {
				board.Note = "未找到同口径的题材板块。"
			} else {
				var result MarketDataset
				var err error
				if job.provider == "新浪财经" {
					result, err = s.market.Feed(ctx, "concept-members", job.row.Code, "")
				} else {
					board.URL = thsCatalogURL + "detail/code/" + job.row.Code + "/"
					result, err = s.market.cached(ctx, "research-ths:"+job.row.Code, time.Minute, s.market.thsMembers, job.row.Code)
				}
				if err != nil || result.Status == "unavailable" {
					board.Note = "成分股来源此次不可用。"
				} else {
					board.Rows = result.Rows
					board.Note = result.Message
					if result.Status == "stale" {
						board.Note = "显示最近一次可用快照。 " + board.Note
					}
				}
			}
			completed <- struct {
				index int
				board researchBoard
			}{i, board}
		}()
	}
	out := make([]researchBoard, len(jobs))
	for range jobs {
		result := <-completed
		out[result.index] = result.board
	}
	return out
}

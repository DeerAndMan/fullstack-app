package service

import (
	"context"
	"fmt"
	"html"
	"math"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const thsIndustryURL = "https://q.10jqka.com.cn/thshy/"
const thsConceptQuotesURL = "https://data.10jqka.com.cn/funds/gnzjl/"
const thsPoolURL = "https://data.10jqka.com.cn/limit_up/"

var thsTableRow = regexp.MustCompile(`(?is)<tr\b[^>]*>(.*?)</tr>`)
var thsTableCell = regexp.MustCompile(`(?is)<td\b[^>]*>(.*?)</td>`)
var thsIndustryPath = regexp.MustCompile(`/thshy/detail/code/(\d{6})/`)
var thsConceptPath = regexp.MustCompile(`/gn/detail/code/(\d{6})/`)
var thsLeaderPath = regexp.MustCompile(`stockpage\.10jqka\.com\.cn/(\d{6})/`)
var thsIndustryPagination = regexp.MustCompile(`class="page_info"[^>]*>\d+/(\d+)`)

func marketYi(value string) *float64 {
	amount := marketNumber(value)
	if amount != nil {
		*amount = math.Round(*amount * 1e8)
	}
	return amount
}

func (s *MarketService) thsIndustry(ctx context.Context, _ string) (MarketDataset, error) {
	return s.thsBoardTable(ctx, false)
}

func (s *MarketService) thsConceptQuotes(ctx context.Context, _ string) (MarketDataset, error) {
	return s.thsBoardTable(ctx, true)
}

func (s *MarketService) thsBoardTable(ctx context.Context, concept bool) (MarketDataset, error) {
	base, pagePath, boardPath := thsIndustryURL, "index/field/199112/order/desc/page/%d/", thsIndustryPath
	boardURL, changeCell, netCell, leaderCell := thsIndustryURL+"detail/code/", 2, 5, 9
	if concept {
		base, pagePath, boardPath = thsConceptQuotesURL, "field/tradezdf/order/asc/page/%d/", thsConceptPath
		boardURL, changeCell, netCell, leaderCell = thsCatalogURL+"detail/code/", 3, 6, 8
	}
	out := MarketDataset{Source: "同花顺", SourceURL: base}
	seen := make(map[string]bool)
	totalPages := 1
	const maxPages = 10
	for page := 1; page <= totalPages && page <= maxPages; page++ {
		endpoint := base
		if page > 1 {
			endpoint += fmt.Sprintf(pagePath, page)
		}
		raw, err := s.request(ctx, endpoint, nil)
		if err != nil || len(raw) == 0 {
			out.Message = "同花顺只返回了部分板块。"
			break
		}
		markup, err := decodeTHS(raw)
		if err != nil {
			out.Message = "同花顺只返回了部分板块。"
			break
		}
		if match := thsIndustryPagination.FindStringSubmatch(markup); page == 1 && len(match) > 0 {
			totalPages, _ = strconv.Atoi(match[1])
		}
		before := len(out.Rows)
		for _, tr := range thsTableRow.FindAllStringSubmatch(markup, -1) {
			cells := thsTableCell.FindAllStringSubmatch(tr[1], -1)
			if len(cells) < 10 {
				continue
			}
			match := boardPath.FindStringSubmatch(cells[1][1])
			if len(match) == 0 || seen[match[1]] {
				continue
			}
			seen[match[1]] = true
			plain := func(i int) string {
				return strings.TrimSpace(html.UnescapeString(flashMarkup.ReplaceAllString(cells[i][1], "")))
			}
			row := MarketRow{
				Code: match[1], Name: plain(1), Change: marketNumber(strings.TrimSuffix(plain(changeCell), "%")),
				Net: marketYi(plain(netCell)), URL: boardURL + match[1] + "/",
			}
			if !concept {
				row.Amount = marketYi(plain(4))
			}
			if leader := thsLeaderPath.FindStringSubmatch(cells[leaderCell][1]); len(leader) > 0 {
				row.Detail, row.Symbol = plain(leaderCell), stockSymbol(leader[1])
			}
			out.Rows = append(out.Rows, row)
		}
		if len(out.Rows) == before || (page == maxPages && page < totalPages) {
			out.Message = "同花顺只返回了部分板块。"
			break
		}
	}
	if len(out.Rows) == 0 {
		return MarketDataset{}, errMarketData
	}
	if concept {
		sort.SliceStable(out.Rows, func(i, j int) bool {
			a, b := out.Rows[i].Change, out.Rows[j].Change
			return a != nil && (b == nil || *a > *b)
		})
	}
	return out, nil
}

func (s *MarketService) thsFundFlow(ctx context.Context, _ string) (MarketDataset, error) {
	out, err := s.Feed(ctx, "industry-ths", "", "")
	if err != nil || out.Status != "ready" {
		return MarketDataset{}, errMarketData
	}
	rows := make([]MarketRow, 0, len(out.Rows))
	for _, row := range out.Rows {
		if row.Net != nil {
			rows = append(rows, row)
		}
	}
	if len(rows) == 0 {
		return MarketDataset{}, errMarketData
	}
	sort.SliceStable(rows, func(i, j int) bool { return *rows[i].Net > *rows[j].Net })
	out.Rows = rows
	return out, nil
}

func (s *MarketService) thsLimitPool(path string) marketLoader {
	return func(ctx context.Context, date string) (MarketDataset, error) {
		type reply struct {
			Status int `json:"status_code"`
			Data   *struct {
				Date string
				Page struct{ Total int }
				Info []struct {
					Code, Name string
					Count      *int   `json:"high_days_value"`
					Tag        string `json:"change_tag"`
				}
			}
		}
		compact := strings.ReplaceAll(date, "-", "")
		out := MarketDataset{Source: "同花顺", SourceURL: thsPoolURL}
		query := url.Values{"page": {"1"}, "limit": {"100"}, "date": {compact}}
		for page := 1; page <= 5; page++ {
			query.Set("page", strconv.Itoa(page))
			data, err := marketJSON[reply](ctx, s, "https://data.10jqka.com.cn/dataapi/"+path, query)
			if err != nil || data.Status != 0 || data.Data == nil {
				if page == 1 {
					return MarketDataset{}, errMarketData
				}
				break
			}
			if data.Data.Date != compact {
				return MarketDataset{}, errMarketSession
			}
			out.Total = &data.Data.Page.Total
			for _, item := range data.Data.Info {
				count := item.Count
				if count == nil && item.Tag == "FIRST_LIMIT" {
					one := 1
					count = &one
				}
				out.Rows = append(out.Rows, MarketRow{
					Code: item.Code, Symbol: stockSymbol(item.Code), Name: item.Name, Date: date, Count: count,
				})
			}
			if len(out.Rows) >= *out.Total || len(data.Data.Info) == 0 {
				break
			}
		}
		if len(out.Rows) < *out.Total {
			out.Message = "同花顺只返回了部分股池记录。"
		}
		return out, nil
	}
}

package service

import (
	"context"
	"encoding/json"
	"net/url"
	"sort"
	"strconv"
	"sync"
	"time"
)

const eastmoneyRegionPage = "https://data.eastmoney.com/bkzj/dy.html"
const eastmoneyBoardPageURL = "https://data.eastmoney.com/bkzj/"

var eastmoneyBoardHosts = []string{"push2test.eastmoney.com", "20.push2.eastmoney.com"}

type eastmoneyBoardList struct {
	Total int
	Diff  []struct {
		Code   string          `json:"f12"`
		Name   string          `json:"f14"`
		Price  json.RawMessage `json:"f2"`
		Change json.RawMessage `json:"f3"`
		Amount json.RawMessage `json:"f6"`
		Net    json.RawMessage `json:"f62"`
		Leader string          `json:"f128"`
		Ticker string          `json:"f140"`
	}
}

func (s *MarketService) eastmoneyBoardPage(ctx context.Context, query url.Values, hosts []string) (*eastmoneyBoardList, error) {
	type reply struct {
		RC   int
		Data *eastmoneyBoardList
	}
	for _, host := range hosts {
		attempt, cancel := context.WithTimeout(ctx, 2500*time.Millisecond)
		data, err := marketJSON[reply](attempt, s, "https://"+host+"/api/qt/clist/get", query)
		cancel()
		if err == nil && data.RC == 0 && data.Data != nil {
			return data.Data, nil
		}
	}
	return nil, errMarketData
}

func (s *MarketService) regionFundFlow(ctx context.Context, _ string) (MarketDataset, error) {
	return s.eastmoneyBoardRows(ctx, "m:90 t:1", eastmoneyRegionPage, eastmoneyBoardHosts, false)
}

const sinaRegionPage = "https://vip.stock.finance.sina.com.cn/mkt/#diyu"

func (s *MarketService) sinaRegionFundFlow(ctx context.Context, _ string) (MarketDataset, error) {
	catalog, err := s.sinaSectorFeed("newFLJK.php", url.Values{"param": {"area"}})(ctx, "")
	if err != nil {
		return MarketDataset{}, err
	}
	const workers = 10
	semaphore := make(chan struct{}, workers)
	var pending sync.WaitGroup
	for i := range catalog.Rows {
		pending.Add(1)
		go func(row *MarketRow) {
			defer pending.Done()
			if !sinaRegionCode.MatchString(row.Code) {
				return
			}
			select {
			case semaphore <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-semaphore }()
			data, err := s.loadSinaBoardMembers(ctx, row.Code, true)
			if err != nil || data.Total == nil || *data.Total == 0 || len(data.Rows) != *data.Total {
				return
			}
			var net float64
			for _, stock := range data.Rows {
				if stock.Net == nil {
					return
				}
				net += *stock.Net
			}
			row.Net = &net
			row.Count = data.Total
		}(&catalog.Rows[i])
	}
	pending.Wait()

	out := MarketDataset{Source: "新浪财经", SourceURL: sinaRegionPage}
	for _, row := range catalog.Rows {
		if row.Net != nil {
			out.Rows = append(out.Rows, row)
		}
	}
	if len(out.Rows) == 0 {
		return MarketDataset{}, errMarketData
	}
	if len(out.Rows) < len(catalog.Rows) {
		out.Message = "部分地区的成分股资金数据不全，暂未列出这些地区。"
	}
	sort.SliceStable(out.Rows, func(i, j int) bool { return *out.Rows[i].Net > *out.Rows[j].Net })
	return out, nil
}

func (s *MarketService) eastmoneyFundFlow(ctx context.Context) (MarketDataset, error) {
	return s.eastmoneyBoardRows(ctx, "m:90 s:4", eastmoneyBoardPageURL, []string{"push2.eastmoney.com"}, false)
}

func (s *MarketService) sourceFundFlow(ctx context.Context, _ string) (MarketDataset, error) {
	return s.eastmoneyBoardRows(ctx, "m:90 s:4", eastmoneyBoardPageURL, eastmoneyBoardHosts, false)
}

func (s *MarketService) eastmoneySectorFeed(filter, sourceURL string) marketLoader {
	return func(ctx context.Context, _ string) (MarketDataset, error) {
		query := url.Values{
			"pz": {"100"}, "po": {"1"}, "np": {"1"}, "fltt": {"2"}, "fid": {"f3"},
			"fs": {filter}, "fields": {"f12,f14,f3,f6,f128,f140"},
		}
		out := MarketDataset{Source: "东方财富", SourceURL: sourceURL}
		seen := make(map[string]bool)
		for page := 1; page <= 15; page++ {
			query.Set("pn", strconv.Itoa(page))
			data, err := s.eastmoneyBoardPage(ctx, query, eastmoneyBoardHosts)
			if err != nil {
				break
			}
			out.Total = &data.Total
			before := len(seen)
			for _, row := range data.Diff {
				if !eastmoneyBoardCode.MatchString(row.Code) || row.Name == "" || seen[row.Code] {
					continue
				}
				seen[row.Code] = true
				out.Rows = append(out.Rows, MarketRow{
					Code: row.Code, Name: row.Name, Change: marketJSONNumber(row.Change), Amount: marketJSONNumber(row.Amount),
					Detail: row.Leader, Symbol: stockSymbol(row.Ticker), URL: "https://quote.eastmoney.com/bk/90." + row.Code + ".html",
				})
			}
			if len(seen) >= data.Total || len(seen) == before {
				break
			}
		}
		if len(out.Rows) == 0 {
			return MarketDataset{}, errMarketData
		}
		if out.Total != nil && len(out.Rows) < *out.Total {
			out.Message = "东方财富只返回了部分板块。"
		}
		return out, nil
	}
}

func (s *MarketService) boardMembers(ctx context.Context, code string) (MarketDataset, error) {
	if eastmoneyBoardCode.MatchString(code) {
		return s.eastmoneyBoardRows(ctx, "b:"+code+" f:!2", eastmoneyBoardPageURL+code+".html", eastmoneyBoardHosts, true)
	}

	flowCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	flow, err := s.loadSinaBoardMembers(flowCtx, code, true)
	cancel()
	if err == nil && flow.Message == "" {
		if len(flow.Rows) > 0 {
			flow.SourceURL = "https://vip.stock.finance.sina.com.cn/moneyflow/"
		}
		return flow, nil
	}
	quotes, quoteErr := s.loadSinaBoardMembers(ctx, code, false)
	if quoteErr != nil && err == nil && len(flow.Rows) > 0 {
		return flow, nil
	}
	return quotes, quoteErr
}

func (s *MarketService) eastmoneyBoardRows(ctx context.Context, filter, sourceURL string, hosts []string, members bool) (MarketDataset, error) {
	sortBy, fields := "f62", "f12,f14,f3,f62"
	if members {
		sortBy, fields = "f12", "f12,f14,f2,f3,f6,f62"
	}
	query := url.Values{
		"pz": {"100"}, "po": {"1"}, "np": {"1"}, "fltt": {"2"}, "fid": {sortBy},
		"fs": {filter}, "fields": {fields}, "ut": {"8dec03ba335b81bf4ebdf7b29ec27d15"},
	}
	out := MarketDataset{Source: "东方财富", SourceURL: sourceURL}
	seen := make(map[string]bool)
	for page := 1; page <= 30; page++ {
		query.Set("pn", strconv.Itoa(page))
		data, err := s.eastmoneyBoardPage(ctx, query, hosts)
		if err != nil {
			if page == 1 {
				return MarketDataset{}, err
			}
			break
		}
		out.Total = &data.Total
		before := len(seen)
		for _, row := range data.Diff {
			if row.Code == "" || row.Name == "" || seen[row.Code] {
				continue
			}
			seen[row.Code] = true
			item := MarketRow{Code: row.Code, Name: row.Name, Change: marketJSONNumber(row.Change), Net: marketJSONNumber(row.Net)}
			if members {
				item.Symbol = stockSymbol(row.Code)
				item.Price = marketJSONNumber(row.Price)
				item.Amount = marketJSONNumber(row.Amount)
			}
			out.Rows = append(out.Rows, item)
		}
		if len(seen) >= data.Total || len(seen) == before {
			break
		}
	}
	if len(out.Rows) < *out.Total {
		out.Message = "行情源只返回了部分板块数据。"
	}
	if !members && len(out.Rows) == 0 {
		return MarketDataset{}, errMarketData
	}
	return out, nil
}

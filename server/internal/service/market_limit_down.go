package service

import (
	"context"
	"net/url"
	"strings"
	"time"
)

const xuanguLimitDownURL = "https://xuangubao.com.cn/"

func (s *MarketService) xuanguLimitDown(ctx context.Context, date string) (MarketDataset, error) {
	type reply struct {
		Code int
		Data []struct {
			Symbol string
			Name   string `json:"stock_chi_name"`
			Price  *float64
			Change *float64 `json:"change_percent"`
			Last   int64    `json:"last_limit_down"`
			Count  *int     `json:"limit_down_days"`
		}
	}
	data, err := marketJSON[reply](ctx, s, "https://flash-api.xuangubao.com.cn/api/pool/detail", url.Values{
		"pool_name": {"limit_down"}, "date": {date},
	})
	if err != nil || data.Code != 20000 {
		return MarketDataset{}, errMarketData
	}
	out := MarketDataset{Source: "选股宝", SourceURL: xuanguLimitDownURL}
	if data.Data == nil {
		out.Message = "选股宝未提供所选日期的跌停记录。"
		return out, nil
	}
	for _, stock := range data.Data {
		if stock.Last <= 0 || time.Unix(stock.Last, 0).In(marketZone).Format(time.DateOnly) != date {
			return MarketDataset{}, errMarketSession
		}
		code, _, _ := strings.Cut(stock.Symbol, ".")
		symbol := stockSymbol(code)
		if symbol == "" || stock.Name == "" {
			continue
		}
		var change *float64
		if stock.Change != nil {
			pct := *stock.Change * 100
			change = &pct
		}
		out.Rows = append(out.Rows, MarketRow{
			Code: code, Symbol: symbol, Name: stock.Name, Price: stock.Price,
			Change: change, Date: date, Count: stock.Count,
		})
	}
	total := len(out.Rows)
	out.Total = &total
	return out, nil
}

package service

import (
	"context"
	"encoding/json"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

func (s *MarketService) stockQuote(ctx context.Context, symbol string) (MarketDataset, error) {
	return s.quotes(symbol)(ctx, "")
}

func stockVolumeShares(volume *float64, symbol string) *float64 {
	if volume == nil || *volume < 0 {
		return nil
	}
	// 腾讯的科创板和沪市存托凭证行情按股计量，其余 A 股行情按手计量。
	if strings.HasPrefix(symbol, "sh688") || strings.HasPrefix(symbol, "sh689") {
		return volume
	}
	shares := *volume * 100
	return &shares
}

func (s *MarketService) kline(period string) marketLoader {
	return func(ctx context.Context, symbol string) (MarketDataset, error) {
		type reply struct {
			Code int
			Data map[string]map[string]json.RawMessage
		}
		endpoint := tencentApp + "/fqkline/get"
		if strings.HasPrefix(symbol, "bj") {
			endpoint = "https://proxy.finance.qq.com/ifzqgtimg/appstock/app/newfqkline/get"
		}
		data, err := marketJSON[reply](ctx, s, endpoint, url.Values{
			"param": {symbol + "," + period + ",,,120,qfq"},
		})
		if err != nil {
			return MarketDataset{}, err
		}
		if data.Code != 0 {
			return MarketDataset{}, errMarketData
		}

		series := data.Data[symbol]
		var rows [][]json.RawMessage
		out := MarketDataset{Source: "腾讯财经", Adjustment: "qfq"}
		if adjusted, ok := series["qfq"+period]; ok {
			if err := json.Unmarshal(adjusted, &rows); err != nil {
				return MarketDataset{}, err
			}
		}
		if len(rows) == 0 {
			out.Adjustment = "none"
			if err := json.Unmarshal(series[period], &rows); err != nil {
				return MarketDataset{}, err
			}
		}
		for _, fields := range rows {
			if len(fields) < 6 {
				continue
			}
			date := strings.Trim(string(fields[0]), `"`)
			if _, err := time.Parse(time.DateOnly, date); err != nil {
				continue
			}
			open, close := marketJSONNumber(fields[1]), marketJSONNumber(fields[2])
			high, low := marketJSONNumber(fields[3]), marketJSONNumber(fields[4])
			if open == nil || close == nil || high == nil || low == nil || *high < *low {
				continue
			}
			out.Rows = append(out.Rows, MarketRow{
				Code: date, Date: date, Open: open, Price: close, High: high, Low: low,
				Volume: stockVolumeShares(marketJSONNumber(fields[5]), symbol),
			})
		}
		if len(out.Rows) == 0 {
			return MarketDataset{}, errMarketData
		}
		sort.SliceStable(out.Rows, func(i, j int) bool { return out.Rows[i].Date < out.Rows[j].Date })
		return out, nil
	}
}

type stockTrendDay struct {
	Date string
	Data []string
}

func stockTrendRows(symbol string, days []stockTrendDay) []MarketRow {
	var rows []MarketRow
	for _, day := range days {
		date, err := time.Parse("20060102", day.Date)
		if err != nil {
			continue
		}
		var previousVolume float64
		for _, point := range day.Data {
			fields := strings.Fields(point)
			if len(fields) < 2 || len(fields[0]) != 4 {
				continue
			}
			clock, err := strconv.Atoi(fields[0])
			price := marketNumber(fields[1])
			if err != nil || clock%100 >= 60 || !(clock >= 930 && clock <= 1130 || clock >= 1300 && clock <= 1500) ||
				price == nil || *price <= 0 {
				continue
			}
			at := date.Format(time.DateOnly) + " " + fields[0][:2] + ":" + fields[0][2:]
			row := MarketRow{Code: at, Date: at, Price: price}
			if len(fields) > 2 {
				if cumulative := stockVolumeShares(marketNumber(fields[2]), symbol); cumulative != nil {
					if delta := *cumulative - previousVolume; delta >= 0 {
						row.Volume = &delta
					}
					previousVolume = *cumulative
					if len(fields) > 3 && *cumulative > 0 {
						if amount := marketNumber(fields[3]); amount != nil && *amount > 0 {
							average := *amount / *cumulative
							row.Average = &average
						}
					}
				}
			}
			rows = append(rows, row)
		}
	}
	sort.SliceStable(rows, func(i, j int) bool { return rows[i].Date < rows[j].Date })
	return rows
}

func (s *MarketService) stockMinute(ctx context.Context, symbol string) (MarketDataset, error) {
	type reply struct {
		Code int
		Data map[string]struct{ Data stockTrendDay }
	}
	data, err := marketJSON[reply](ctx, s, tencentApp+"/minute/query", url.Values{"code": {symbol}})
	if err != nil {
		return MarketDataset{}, err
	}
	rows := stockTrendRows(symbol, []stockTrendDay{data.Data[symbol].Data})
	if data.Code != 0 || len(rows) == 0 {
		return MarketDataset{}, errMarketData
	}
	return MarketDataset{Source: "腾讯财经", Rows: rows}, nil
}

func (s *MarketService) stockFiveDay(ctx context.Context, symbol string) (MarketDataset, error) {
	type reply struct {
		Code int
		Data map[string]struct{ Data []stockTrendDay }
	}
	data, err := marketJSON[reply](ctx, s, tencentApp+"/day/query", url.Values{"code": {symbol}})
	if err != nil {
		return MarketDataset{}, err
	}
	rows := stockTrendRows(symbol, data.Data[symbol].Data)
	if data.Code != 0 || len(rows) == 0 {
		return MarketDataset{}, errMarketData
	}
	return MarketDataset{Source: "腾讯财经", Rows: rows}, nil
}

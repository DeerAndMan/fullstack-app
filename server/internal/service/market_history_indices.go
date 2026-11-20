package service

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const sinaHistoryURL = "https://finance.sina.com.cn/realstock/company/sh000001/nc.shtml"
const eastmoneyHistoryURL = "https://quote.eastmoney.com/sh000001.html"

// The archive records the same two EastMoney indices through 2026-11-19.
//
//go:embed market_history_eastmoney_archive.json
var eastmoneyHistoryArchive []byte

func (s *MarketService) eastmoneyIndexAmounts(ctx context.Context, secID string, query url.Values) (map[string]float64, error) {
	type reply struct {
		RC   int
		Data *struct {
			Code   string
			Market int
			Klines []string
		}
	}
	query.Set("secid", secID)
	for _, host := range []string{"8.push2his.eastmoney.com", "push2his.eastmoney.com"} {
		attempt, cancel := context.WithTimeout(ctx, 1900*time.Millisecond)
		data, err := marketJSON[reply](attempt, s, "https://"+host+"/api/qt/stock/kline/get", query)
		cancel()
		if err != nil || data.RC != 0 || data.Data == nil ||
			strconv.Itoa(data.Data.Market)+"."+data.Data.Code != secID || data.Data.Klines == nil {
			continue
		}
		amounts := make(map[string]float64)
		for _, line := range data.Data.Klines {
			fields := strings.Split(line, ",")
			if len(fields) < 7 {
				continue
			}
			amount := marketNumber(fields[6])
			if _, err := time.Parse(time.DateOnly, fields[0]); err == nil && amount != nil && *amount > 0 {
				amounts[fields[0]] = *amount
			}
		}
		return amounts, nil
	}
	return nil, errMarketData
}

func (s *MarketService) eastmoneyMarketHistory(ctx context.Context, period string) (MarketDataset, error) {
	now := time.Now().In(marketZone)
	start, end, _ := historyPeriod(period, now)
	today := now.Format(time.DateOnly)
	if end >= today {
		end = now.AddDate(0, 0, -1).Format(time.DateOnly)
	}
	if start == "" {
		start = now.AddDate(0, 0, -60).Format(time.DateOnly)
	}
	out := MarketDataset{Source: "东方财富", SourceURL: eastmoneyHistoryURL}
	if start > end {
		return out, nil
	}
	query := url.Values{
		"fields1": {"f1,f2,f3,f4,f5"}, "fields2": {"f51,f52,f53,f54,f55,f56,f57"},
		"klt": {"101"}, "fqt": {"0"}, "lmt": {"1023"}, "ut": {"fa5fd1943c7b386f172d6893dbfba10b"},
		"beg": {strings.ReplaceAll(start, "-", "")}, "end": {strings.ReplaceAll(end, "-", "")},
	}
	sh, err := s.eastmoneyIndexAmounts(ctx, "1.000001", query)
	var sz map[string]float64
	if err == nil {
		sz, err = s.eastmoneyIndexAmounts(ctx, "0.399106", query)
	}
	var archived map[string]float64
	_ = json.Unmarshal(eastmoneyHistoryArchive, &archived)
	cacheDir := historyCachePath()
	if cacheDir != "" {
		cacheDir = filepath.Join(cacheDir, "eastmoney-indices")
		readHistorySnapshots(cacheDir, start, end, func(date string, raw []byte) {
			var amount float64
			if json.Unmarshal(raw, &amount) == nil && amount > 0 {
				archived[date] = amount
			}
		})
	}
	seen := make(map[string]bool)
	for date, shAmount := range sh {
		szAmount, ok := sz[date]
		if !ok || date < start || date > end {
			continue
		}
		amount := shAmount + szAmount
		seen[date] = true
		out.Rows = append(out.Rows, MarketRow{Code: date, Name: date, Date: date, Amount: &amount, QuoteSource: out.Source, URL: out.SourceURL})
		if cacheDir != "" && archived[date] != amount {
			_ = saveHistorySnapshot(cacheDir, date, amount)
		}
	}
	liveRows := len(out.Rows)
	for date, amount := range archived {
		if date >= start && date <= end && !seen[date] {
			out.Rows = append(out.Rows, MarketRow{
				Code: date, Name: date, Date: date, Amount: &amount, QuoteSource: "东方财富（归档）", URL: out.SourceURL,
			})
		}
	}
	if err != nil && len(out.Rows) == 0 {
		return MarketDataset{}, err
	}
	sort.Slice(out.Rows, func(i, j int) bool { return out.Rows[i].Date < out.Rows[j].Date })
	if len(out.Rows) > liveRows {
		out.Message = "包含东方财富已归档成交额，逐日明细标明来源。"
		if err != nil {
			out.Message = "东方财富实时查询暂不可用，展示所选区间内已归档的交易日。"
		}
	} else if liveRows < len(sh) || liveRows < len(sz) || len(sh) >= 1023 || len(sz) >= 1023 {
		out.Message = "东方财富只返回了部分区间的沪深指数成交额。"
	}
	if period == "" && len(out.Rows) > 20 {
		out.Rows = out.Rows[len(out.Rows)-20:]
	}
	return out, nil
}

func (s *MarketService) sinaIndexChanges(ctx context.Context, symbol string) (map[string]float64, error) {
	type result struct {
		Result struct {
			Status struct{ Code int }
			Data   []struct{ Day, Close string }
		}
	}
	data, err := marketJSON[result](ctx, s, "https://quotes.sina.cn/cn/api/openapi.php/CN_MarketDataService.getKLineData", url.Values{
		"symbol": {symbol}, "scale": {"240"}, "ma": {"no"}, "datalen": {"1023"},
	})
	if err != nil || data.Result.Status.Code != 0 || len(data.Result.Data) < 2 {
		return nil, errMarketData
	}
	rows := data.Result.Data
	sort.Slice(rows, func(i, j int) bool { return rows[i].Day < rows[j].Day })
	changes := make(map[string]float64, len(rows)-1)
	var previous *float64
	var previousDay string
	for _, row := range rows {
		if _, err := time.Parse(time.DateOnly, row.Day); err != nil || row.Day == previousDay {
			return nil, errMarketData
		}
		close := marketNumber(row.Close)
		if close != nil && *close > 0 && previous != nil && *previous > 0 {
			changes[row.Day] = (*close / *previous - 1) * 100
		}
		previous, previousDay = close, row.Day
	}
	return changes, nil
}

func (s *MarketService) sinaMarketHistory(ctx context.Context, period string) (MarketDataset, error) {
	sh, err := s.sinaIndexChanges(ctx, "sh000001")
	if err != nil {
		return MarketDataset{}, err
	}
	sz, err := s.sinaIndexChanges(ctx, "sz399106")
	if err != nil {
		return MarketDataset{}, err
	}
	now := time.Now().In(marketZone)
	start, end, _ := historyPeriod(period, now)
	today := now.Format(time.DateOnly)
	out := MarketDataset{Source: "新浪财经", SourceURL: sinaHistoryURL}
	var earliest string
	for date, shChange := range sh {
		szChange, ok := sz[date]
		if !ok {
			continue
		}
		if earliest == "" || date < earliest {
			earliest = date
		}
		if date >= start && date <= end && date < today {
			out.Rows = append(out.Rows, MarketRow{
				Code: date, Name: date, Date: date,
				ShanghaiChange: &shChange, ShenzhenChange: &szChange,
			})
		}
	}
	if earliest == "" {
		return MarketDataset{}, errMarketData
	}
	if start != "" && start < earliest {
		out.Message = "新浪财经指数日线可计算日涨跌幅的最早日期为 " + earliest + "。"
	}
	sort.Slice(out.Rows, func(i, j int) bool { return out.Rows[i].Date < out.Rows[j].Date })
	if period == "" && len(out.Rows) > 20 {
		out.Rows = out.Rows[len(out.Rows)-20:]
	}
	return out, nil
}

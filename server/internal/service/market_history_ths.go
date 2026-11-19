package service

import (
	"context"
	"sort"
	"strings"
	"time"
)

const thsMarketHistoryURL = "https://stockpage.10jqka.com.cn/1A0001/"

func (s *MarketService) thsMarketHistory(ctx context.Context, period string) (MarketDataset, error) {
	now := time.Now().In(marketZone)
	start, end, _ := historyPeriod(period, now)
	last, _ := time.Parse(time.DateOnly, end)
	first := last.AddDate(0, 0, -60)
	if start != "" {
		first, _ = time.Parse(time.DateOnly, start)
	}
	out := MarketDataset{Source: "同花顺", SourceURL: thsMarketHistoryURL}
	today := now.Format(time.DateOnly)
	for year := last.Year(); year >= first.Year(); year-- {
		sh, errSH := s.thsIndexKlines(ctx, "1A0001", year)
		sz, errSZ := s.thsIndexKlines(ctx, "399106", year)
		if errSH != nil || errSZ != nil {
			out.Message = "同花顺只返回了部分交易日。"
			continue
		}
		shAmounts, szAmounts := thsDailyAmounts(sh), thsDailyAmounts(sz)
		for date, a := range shAmounts {
			b, ok := szAmounts[date]
			if !ok || date > end || date >= today || date < first.Format(time.DateOnly) {
				continue
			}
			amount := a + b
			out.Rows = append(out.Rows, MarketRow{
				Code: date, Name: date, Date: date, Amount: &amount, QuoteSource: "同花顺",
				URL: thsMarketHistoryURL,
			})
		}
	}
	if len(out.Rows) == 0 && out.Message != "" {
		return MarketDataset{}, errMarketData
	}
	sort.Slice(out.Rows, func(i, j int) bool { return out.Rows[i].Date < out.Rows[j].Date })
	if period == "" && len(out.Rows) > 20 {
		out.Rows = out.Rows[len(out.Rows)-20:]
	}
	return out, nil
}

func thsDailyAmounts(records []string) map[string]float64 {
	amounts := make(map[string]float64, len(records))
	for _, record := range records {
		fields := strings.Split(record, ",")
		if len(fields) < 7 {
			continue
		}
		date, err := time.Parse("20060102", fields[0])
		amount := marketNumber(fields[6])
		if err == nil && amount != nil && *amount > 0 {
			amounts[date.Format(time.DateOnly)] = *amount
		}
	}
	return amounts
}

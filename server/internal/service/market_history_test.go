package service

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHistoryClosingSourceAndRestart(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("MARKET_HISTORY_CACHE_DIR", dir)
	const day = "2025-02-07"
	service := NewMarketService()
	service.client.Transport = marketTestTransport(func(r *http.Request) (*http.Response, error) {
		switch {
		case r.URL.Host == "d.10jqka.com.cn":
			return marketTestResponse(r, 200, `quotebridge_v6_line_hs_1A0001_01_2025({"data":"20250207,1,1"})`), nil
		case r.URL.Query().Get("a") == "ChangeStatistics":
			return marketTestResponse(r, 200, `{"errcode":"0","info":[{"Day":"2025-02-07","strong":"0"}]}`), nil
		case r.URL.Query().Get("Day") == day:
			return marketTestResponse(r, 200, `{"errcode":"0","date":"2025-02-07","info":{"SZJS":"0","XDJS":"5012","qscln":34600}}`), nil
		default:
			t.Errorf("unexpected request %s", r.URL)
			return marketTestResponse(r, 404, ""), nil
		}
	})
	verify := func(s *MarketService) {
		t.Helper()
		data, err := s.Feed(context.Background(), "market-history", "", "2025-02")
		if err != nil || len(data.Rows) != 1 {
			t.Fatalf("response %+v err %v", data, err)
		}
		row := data.Rows[0]
		if row.Date != day || row.QuoteSource != "开盘啦" || *row.Sentiment != 0 || *row.Advancers != 0 || *row.Decliners != 5012 || *row.Amount != 346000000 {
			t.Fatalf("unexpected row %+v", row)
		}
	}
	verify(service)
	if _, err := os.Stat(filepath.Join(dir, day+".json")); err != nil {
		t.Fatal(err)
	}
	restarted := NewMarketService()
	restarted.client.Transport = marketTestTransport(func(r *http.Request) (*http.Response, error) { return marketTestResponse(r, 503, ""), nil })
	verify(restarted)
}

func TestHistoryArchiveCoversMarketSessionsAndWinsOverBackup(t *testing.T) {
	t.Setenv("MARKET_HISTORY_CACHE_DIR", t.TempDir())
	s := NewMarketService()
	s.client.Transport = marketTestTransport(func(r *http.Request) (*http.Response, error) { return marketTestResponse(r, 503, ""), nil })
	data, err := s.marketHistory(context.Background(), "2026-10-19..2026-11-02")
	scores := []float64{45, 64, 49, 31, 78, 52, 36, 70, 33, 36, 87}
	if err != nil || len(data.Rows) != len(scores) {
		t.Fatalf("rows %d err %v", len(data.Rows), err)
	}
	for i, row := range data.Rows {
		if row.Detail != "开盘啦情绪指标" || row.QuoteSource != "开盘啦" || *row.Sentiment != scores[i] {
			t.Fatalf("%s %v", row.Date, row)
		}
	}
	last := data.Rows[10]
	if last.Date != "2026-11-02" || *last.Amount != 1783690410000 || *last.Advancers != 4681 || *last.Decliners != 817 {
		t.Fatalf("%+v", last)
	}
}

func TestHistoryFillsOnlyMissingMetrics(t *testing.T) {
	t.Setenv("MARKET_HISTORY_CACHE_DIR", t.TempDir())
	for _, primaryMissing := range []string{"score", "quote", "both"} {
		t.Run(primaryMissing, func(t *testing.T) {
			s := NewMarketService()
			s.client.Transport = marketTestTransport(func(r *http.Request) (*http.Response, error) {
				switch {
				case r.URL.Host == "d.10jqka.com.cn":
					return marketTestResponse(r, 200, `quotebridge_v6_line_hs_1A0001_01_2025({"data":"20250207,1"})`), nil
				case r.URL.Query().Get("a") == "ChangeStatistics":
					if primaryMissing == "quote" {
						return marketTestResponse(r, 200, `{"errcode":0,"info":[{"Day":"2025-02-07","strong":55}]}`), nil
					}
				case r.URL.Query().Get("a") == "HisZhangFuDetail":
					date := "2025-02-06"
					if primaryMissing == "score" {
						date = "2025-02-07"
					}
					return marketTestResponse(r, 200, `{"errcode":0,"date":"`+date+`","info":{"SZJS":"7","XDJS":"2","qscln":123}}`), nil
				case r.URL.Host == "raw.githubusercontent.com":
					return marketTestResponse(r, 200, `{"summary":{"trade_date":"2025-02-07","advancers":9,"decliners":1,"emotion_score":86},"indices":[{"code":"000001","amount_yi":1},{"code":"399001","amount_yi":2}]}`), nil
				}
				return marketTestResponse(r, 503, ""), nil
			})
			data, err := s.marketHistory(context.Background(), "2025-02")
			if err != nil || len(data.Rows) != 1 || !strings.Contains(data.Message, "备用来源") {
				t.Fatalf("%+v %v", data, err)
			}
			row := data.Rows[0]
			if primaryMissing == "quote" && (*row.Sentiment != 55 || row.Detail != "开盘啦情绪指标") || primaryMissing != "quote" && (*row.Sentiment != 86 || row.Detail != reportName) {
				t.Fatalf("wrong score source %+v", row)
			}
			if primaryMissing == "score" && (*row.Amount != 1230000 || row.QuoteSource != "开盘啦") || primaryMissing != "score" && (*row.Advancers != 9 || row.QuoteSource != reportName) {
				t.Fatalf("wrong quote source %+v", row)
			}
		})
	}
}

func TestHistoryZhishiRequiresDatedClosingFigures(t *testing.T) {
	s := NewMarketService()
	s.client.Transport = marketTestTransport(func(r *http.Request) (*http.Response, error) {
		return marketTestResponse(r, 200, `<section aria-label="市场整体情绪"><p>2025-02-07 · 收盘统计</p><p>综合情绪因子 <b>53.8</b> / 100 · v1</p><p>上涨 2,274 · 下跌 3129 · 平盘 160 家 · 成交额 14,375.1 亿元</p></section>`), nil
	})
	d := s.zhishiBackup(context.Background(), "2025-02-07")
	if !d.valid() || d.Up != 2274 || d.AmountYi != 14375.1 || d.Score != 53.8 {
		t.Fatal(d)
	}
	if s.zhishiBackup(context.Background(), "2025-02-06").valid() {
		t.Fatal("accepted wrong session")
	}
	var seed []historyDay
	if err := json.Unmarshal(historyArchive, &seed); err != nil {
		t.Fatal(err)
	}
	for i, d := range seed {
		if !d.complete() || (i > 0 && d.Date <= seed[i-1].Date) {
			t.Fatal(d)
		}
	}
}

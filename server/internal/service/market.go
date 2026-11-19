package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type MarketRow struct {
	PETTM       *float64 `json:"peTtm,omitempty"`
	PBMRQ       *float64 `json:"pbMrq,omitempty"`
	PSTTM       *float64 `json:"psTtm,omitempty"`
	PCFOCFTTM   *float64 `json:"pcfOcfTtm,omitempty"`
	QuoteSource string   `json:"quoteSource,omitempty"`
	ScoreURL    string   `json:"scoreUrl,omitempty"`
	Advancers   *int     `json:"advancers,omitempty"`
	Decliners   *int     `json:"decliners,omitempty"`
	Sentiment   *float64 `json:"sentiment,omitempty"`
	Code        string   `json:"code"`
	Symbol      string   `json:"symbol,omitempty"`
	Name        string   `json:"name"`
	Date        string   `json:"date,omitempty"`
	Detail      string   `json:"detail,omitempty"`
	URL         string   `json:"url,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Change      *float64 `json:"change,omitempty"`
	Amount      *float64 `json:"amount,omitempty"`
	Net         *float64 `json:"net,omitempty"`
	Count       *int     `json:"count,omitempty"`
	Priority    *int     `json:"priority,omitempty"`
	Open        *float64 `json:"open,omitempty"`
	High        *float64 `json:"high,omitempty"`
	Low         *float64 `json:"low,omitempty"`
	Volume      *float64 `json:"volume,omitempty"`
	Average     *float64 `json:"average,omitempty"`
	Previous    string   `json:"previous,omitempty"`
	Forecast    string   `json:"forecast,omitempty"`
	Actual      string   `json:"actual,omitempty"`
}

type MarketDataset struct {
	Source     string      `json:"source"`
	SourceURL  string      `json:"sourceUrl,omitempty"`
	FetchedAt  time.Time   `json:"fetchedAt"`
	Status     string      `json:"status"`
	Message    string      `json:"message,omitempty"`
	Rows       []MarketRow `json:"rows"`
	Total      *int        `json:"total,omitempty"`
	Adjustment string      `json:"adjustment,omitempty"`
}

type marketLoader func(context.Context, string) (MarketDataset, error)

type marketCacheEntry struct {
	data    MarketDataset
	expires time.Time
	pending chan struct{}
}

type MarketService struct {
	client *http.Client
	mu     sync.Mutex
	cache  map[string]*marketCacheEntry
}

var marketZone = time.FixedZone("CST", 8*60*60)
var ErrMarketQuery = errors.New("无效的市场数据查询")
var errMarketSession = errors.New("provider did not confirm the requested session")
var conceptCode = regexp.MustCompile(`^gn_[a-z0-9_]{1,40}$`)
var marketStockCode = regexp.MustCompile(`^(?:sh6[0-9]{5}|sz(?:00|30)[0-9]{4}|bj[489][0-9]{5})$`)
var eastmoneyBoardCode = regexp.MustCompile(`^BK[0-9]{4,6}$`)
var sinaIndustryCode = regexp.MustCompile(`^(?:new_[a-z0-9]+|hangye_[A-Za-z0-9]+)$`)

const (
	indexSymbols  = "sh000001,sz399001,sz399006,sh000300,sh000016,sh000905,sh000852,sh000688"
	globalSymbols = "hkHSI,hkHSCEI,usIXIC,usDJI,usINX"
	etfSymbols    = "sh510300,sh510050,sh510500,sh512100,sh588000,sz159915,sh563300"
)

func NewMarketService() *MarketService {
	return &MarketService{
		client: &http.Client{Timeout: 6 * time.Second},
		cache:  make(map[string]*marketCacheEntry),
	}
}

func (s *MarketService) Feed(ctx context.Context, feed, symbol, date string) (MarketDataset, error) {
	loaders := map[string]marketLoader{
		"indices": s.quotes(indexSymbols), "global": s.quotes(globalSymbols), "etfs": s.quotes(etfSymbols),
		"industry": s.sectors(false), "concept": s.sectors(true), "fund-flow": s.fundFlow,
		"turnover": s.turnover, "limit-up": s.limitPool("getTopicZTPool"), "limit-down": s.limitPool("getTopicDTPool"),
		"opened-limit": s.limitPool("getTopicZBPool"), "flash": s.flash, "calendar": s.calendar,
		"flash-sina": s.sinaFlash, "flash-eastmoney": s.eastmoneyFlash,
		"flash-kaipanla": s.kaipanlaFlash, "flash-ths": s.thsStockNews,
		"concept-eastmoney":  s.eastmoneySectorFeed("m:90 t:3", "https://quote.eastmoney.com/center/boardlist.html#concept_board"),
		"industry-eastmoney": s.eastmoneySectorFeed("m:90 t:2", "https://quote.eastmoney.com/center/boardlist.html#industry_board"),
		"industry-ths":       s.thsIndustry, "fund-flow-ths": s.thsFundFlow,
		"concept-ths":          s.thsConceptQuotes,
		"fund-flow-eastmoney":  s.sourceFundFlow,
		"fund-flow-sina":       func(ctx context.Context, _ string) (MarketDataset, error) { return s.sinaFundFlow(ctx) },
		"limit-up-ths":         s.thsLimitPool("limit_up/limit_up_pool"),
		"opened-limit-ths":     s.thsLimitPool("limit_up/open_limit_pool"),
		"leaderboard-kaipanla": s.kaipanlaLeaderboard, "calendar-eastmoney": s.eastmoneyCalendar,
		"leaderboard": s.leaderboard, "margin": s.margin,
		"daily": s.daily, "minute": s.minute,
		"stock-quote": s.stockQuote, "stock-minute": s.stockMinute, "stock-five-day": s.stockFiveDay,
		"stock-daily": s.kline("day"), "stock-weekly": s.kline("week"), "stock-monthly": s.kline("month"),
		"stock-valuation":    s.stockValuation,
		"market-history":     s.marketHistory,
		"market-history-ths": s.thsMarketHistory,
		"concept-members":    s.conceptMembers,
		"concept-news":       s.conceptNews,
		"region-fund-flow":   s.regionFundFlow,
		"board-members":      s.boardMembers,
	}
	loader, ok := loaders[feed]
	if !ok {
		return MarketDataset{}, ErrMarketQuery
	}

	now := time.Now().In(marketZone)
	arg := ""
	switch feed {
	case "board-members":
		if !eastmoneyBoardCode.MatchString(symbol) && !sinaIndustryCode.MatchString(symbol) {
			return MarketDataset{}, ErrMarketQuery
		}
		arg = symbol
	case "market-history", "market-history-ths":
		if _, _, ok := historyPeriod(date, now); !ok {
			return MarketDataset{}, ErrMarketQuery
		}
		arg = date
	case "concept-members", "concept-news":
		if !conceptCode.MatchString(symbol) {
			return MarketDataset{}, ErrMarketQuery
		}
		catalog, err := s.Feed(ctx, "concept", "", "")
		if err != nil {
			return MarketDataset{}, err
		}
		if catalog.Status == "unavailable" {
			return MarketDataset{Status: "unavailable", Rows: []MarketRow{}, Message: "题材目录暂不可用。"}, nil
		}
		var name string
		for _, row := range catalog.Rows {
			if row.Code == symbol {
				name = row.Name
				break
			}
		}
		if name == "" {
			return MarketDataset{}, ErrMarketQuery
		}
		arg = symbol
		if feed == "concept-news" {
			arg = name
		}
	case "daily", "minute":
		allSymbols := "," + indexSymbols + "," + globalSymbols + "," + etfSymbols + ","
		if symbol == "" || strings.Contains(symbol, ",") || !strings.Contains(allSymbols, ","+symbol+",") {
			return MarketDataset{}, ErrMarketQuery
		}
		arg = symbol
	case "stock-quote", "stock-minute", "stock-five-day", "stock-daily", "stock-weekly", "stock-monthly", "stock-valuation":
		if !marketStockCode.MatchString(symbol) {
			return MarketDataset{}, ErrMarketQuery
		}
		arg = symbol
	case "limit-up", "limit-down", "opened-limit", "limit-up-ths", "opened-limit-ths", "leaderboard", "leaderboard-kaipanla", "calendar", "calendar-eastmoney":
		arg = date
		if arg == "" {
			arg = now.Format(time.DateOnly)
		}
		parsed, err := time.ParseInLocation(time.DateOnly, arg, marketZone)
		if err != nil || parsed.Format(time.DateOnly) != arg || arg < now.AddDate(0, 0, -120).Format(time.DateOnly) || arg > now.Format(time.DateOnly) {
			return MarketDataset{}, ErrMarketQuery
		}
		if (feed == "leaderboard" || feed == "leaderboard-kaipanla") && date == "" {
			arg = ""
		}
	}

	ttl := time.Minute
	switch feed {
	case "flash", "flash-sina", "flash-eastmoney", "flash-kaipanla", "flash-ths":
		ttl = 20 * time.Second
	case "stock-valuation":
		ttl = 30 * time.Minute
	case "market-history", "market-history-ths":
		ttl = 15 * time.Minute
	case "daily", "margin", "calendar", "calendar-eastmoney", "leaderboard", "leaderboard-kaipanla", "concept-news":
		ttl = 5 * time.Minute
	}
	return s.cached(ctx, feed+":"+arg, ttl, loader, arg)
}

func stockSymbol(code string) string {
	for _, exchange := range []string{"sh", "sz", "bj"} {
		if symbol := exchange + code; marketStockCode.MatchString(symbol) {
			return symbol
		}
	}
	return ""
}

func (s *MarketService) cached(ctx context.Context, key string, ttl time.Duration, load marketLoader, arg string) (MarketDataset, error) {
	for {
		now := time.Now()
		s.mu.Lock()
		for name, entry := range s.cache {
			if entry.pending == nil && now.Sub(entry.expires) > 24*time.Hour {
				delete(s.cache, name)
			}
		}
		entry := s.cache[key]
		if entry == nil {
			entry = &marketCacheEntry{}
			s.cache[key] = entry
		}
		if entry.pending != nil {
			wait := entry.pending
			s.mu.Unlock()
			select {
			case <-ctx.Done():
				return MarketDataset{}, ctx.Err()
			case <-wait:
				continue
			}
		}
		if now.Before(entry.expires) {
			data := entry.data
			s.mu.Unlock()
			return data, nil
		}
		entry.pending = make(chan struct{})
		s.mu.Unlock()

		loadCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 8*time.Second)
		fresh, err := load(loadCtx, arg)
		cancel()
		s.mu.Lock()
		if err != nil {
			entry.expires = time.Now().Add(15 * time.Second)
			if entry.data.FetchedAt.IsZero() || time.Since(entry.data.FetchedAt) > 24*time.Hour {
				entry.data = MarketDataset{Status: "unavailable", Rows: []MarketRow{}}
				if errors.Is(err, errMarketSession) {
					entry.data.Message = "数据源未返回所选交易日的股池，请切换日期或查看来源页面。"
				}
			} else {
				entry.data.Status = "stale"
			}
		} else {
			fresh.FetchedAt = time.Now()
			fresh.Status = "ready"
			if fresh.Rows == nil {
				fresh.Rows = []MarketRow{}
			}
			entry.data = fresh
			if strings.HasPrefix(key, "market-history:") && fresh.Message != "" {
				ttl = 15 * time.Second
			}
			entry.expires = fresh.FetchedAt.Add(ttl)
		}
		close(entry.pending)
		entry.pending = nil
		result := entry.data
		s.mu.Unlock()
		return result, nil
	}
}

func (s *MarketService) request(ctx context.Context, endpoint string, query url.Values) ([]byte, error) {
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	switch req.URL.Host {
	case "push2test.eastmoney.com", "20.push2.eastmoney.com":
		req.Header.Set("Referer", "https://data.eastmoney.com/")
		req.Header.Set("User-Agent", "Mozilla/5.0")
	case "q.10jqka.com.cn", "data.10jqka.com.cn":
		req.Header.Set("Referer", "https://"+req.URL.Host+"/")
		req.Header.Set("User-Agent", "Mozilla/5.0")
	case "datacenter.eastmoney.com", "datacenter-web.eastmoney.com":
		if req.URL.Query().Get("reportName") == "RPT_CPH_FECALENDAR" {
			req.Header.Set("Referer", "https://emdata.eastmoney.com/")
			req.Header.Set("User-Agent", "Mozilla/5.0")
		}
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("market provider: HTTP %d", resp.StatusCode)
	}
	const max = 2 << 20
	body, err := io.ReadAll(io.LimitReader(resp.Body, max+1))
	if err == nil && len(body) > max {
		return nil, errors.New("market provider response too large")
	}
	return body, err
}

func marketNumber(value string) *float64 {
	number, err := strconv.ParseFloat(value, 64)
	if err != nil || strings.ContainsAny(value, "INin") {
		return nil
	}
	return &number
}

func marketDate(value string) string {
	date, _, _ := strings.Cut(value, " ")
	return date
}

var tencentTimestamp = regexp.MustCompile(`^\d{14}$`)

func quoteTimestamp(value string) string {
	if tencentTimestamp.MatchString(value) {
		parsed, err := time.ParseInLocation("20060102150405", value, marketZone)
		if err == nil {
			return parsed.Format("2006-01-02 15:04:05")
		}
	}
	return strings.ReplaceAll(value, "/", "-")
}

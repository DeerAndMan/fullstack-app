package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
)

const valuationEndpoint = "https://datacenter.eastmoney.com/api/data/v1/get"
const valuationPageSize = 1000

type valuationRecord struct {
	Symbol string   `json:"SECUCODE"`
	Date   string   `json:"TRADE_DATE"`
	PE     *float64 `json:"PE_TTM"`
	PB     *float64 `json:"PB_MRQ"`
	PS     *float64 `json:"PS_TTM"`
	PCF    *float64 `json:"PCF_OCF_TTM"`
}

type valuationPage struct {
	Pages int
	Count int
	Data  []valuationRecord
}

func (s *MarketService) valuationPage(ctx context.Context, filter string, number int) (valuationPage, error) {
	body, err := s.request(ctx, valuationEndpoint, url.Values{
		"reportName": {"RPT_VALUEANALYSIS_DET"},
		"columns":    {"SECUCODE,TRADE_DATE,PE_TTM,PB_MRQ,PS_TTM,PCF_OCF_TTM"},
		"filter":     {filter}, "pageNumber": {strconv.Itoa(number)}, "pageSize": {strconv.Itoa(valuationPageSize)},
		"sortColumns": {"TRADE_DATE"}, "sortTypes": {"-1"},
		"source": {"WEB"}, "client": {"WEB"},
	})
	if err != nil {
		return valuationPage{}, err
	}
	var reply struct {
		Success bool
		Code    int
		Result  *valuationPage
	}
	if err := json.Unmarshal(body, &reply); err != nil {
		return valuationPage{}, err
	}
	if number == 1 && reply.Code == 9201 && reply.Result == nil {
		return valuationPage{}, nil
	}
	if !reply.Success || reply.Result == nil {
		return valuationPage{}, errMarketData
	}
	return *reply.Result, nil
}

func (s *MarketService) stockValuation(ctx context.Context, symbol string) (MarketDataset, error) {
	code := symbol[2:]
	tag := code + "." + strings.ToUpper(symbol[:2])
	out := MarketDataset{Source: "东方财富", SourceURL: "https://data.eastmoney.com/gzfx/detail/" + code + ".html"}
	now := time.Now().In(marketZone)
	start, end := now.AddDate(-10, 0, -14).Format(time.DateOnly), now.Format(time.DateOnly)
	filter := fmt.Sprintf("(SECUCODE=%q)(TRADE_DATE>='%s')(TRADE_DATE<='%s')", tag, start, end)
	first, err := s.valuationPage(ctx, filter, 1)
	if err != nil || first.Count == 0 {
		return out, err
	}
	if first.Count < 0 || first.Pages < 1 || first.Pages > 4 || first.Pages != (first.Count+valuationPageSize-1)/valuationPageSize {
		return out, errMarketData
	}

	pages := make([]valuationPage, first.Pages)
	pageErrors := make([]error, first.Pages)
	pages[0] = first
	var pending sync.WaitGroup
	for i := 1; i < len(pages); i++ {
		pending.Add(1)
		go func() {
			defer pending.Done()
			pages[i], pageErrors[i] = s.valuationPage(ctx, filter, i+1)
		}()
	}
	pending.Wait()

	seen := make(map[string]bool, first.Count)
	for i, page := range pages {
		if pageErrors[i] != nil {
			return out, pageErrors[i]
		}
		want := min(valuationPageSize, first.Count-i*valuationPageSize)
		if page.Count != first.Count || page.Pages != first.Pages || len(page.Data) != want {
			return out, errMarketData
		}
		for _, row := range page.Data {
			date := marketDate(row.Date)
			if _, err := time.Parse(time.DateOnly, date); err != nil || row.Symbol != tag || date < start || date > end || seen[date] {
				return out, errMarketData
			}
			seen[date] = true
			out.Rows = append(out.Rows, MarketRow{
				Code: date, Name: date, Symbol: symbol, Date: date,
				PETTM: row.PE, PBMRQ: row.PB, PSTTM: row.PS, PCFOCFTTM: row.PCF,
			})
		}
	}
	slices.SortFunc(out.Rows, func(a, b MarketRow) int { return strings.Compare(a.Date, b.Date) })
	return out, nil
}

package service

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const eastmoneyCalendarURL = "https://emdata.eastmoney.com/cjrl/index.html"
const kaipanlaLeaderboardURL = "https://www.kaipanla.com/omain/"

func (s *MarketService) eastmoneyCalendar(ctx context.Context, date string) (MarketDataset, error) {
	type reply struct {
		Success bool
		Result  *struct {
			Pages int
			Data  []struct {
				Code  string `json:"FE_CODE"`
				Name  string `json:"FE_NAME"`
				Type  string `json:"FE_TYPE"`
				City  string `json:"CITY"`
				Start string `json:"START_DATE"`
				End   string `json:"END_DATE"`
			}
		}
	}
	start, _ := time.Parse(time.DateOnly, date)
	end := start.AddDate(0, 0, 7).Format(time.DateOnly)
	query := url.Values{
		"reportName": {"RPT_CPH_FECALENDAR"}, "columns": {"ALL"}, "pageSize": {"100"}, "pageNumber": {"1"},
		"filter": {fmt.Sprintf("(END_DATE>='%s')(START_DATE<'%s')", date, end)},
	}
	out := MarketDataset{Source: "东方财富", SourceURL: eastmoneyCalendarURL}
	for page := 1; page <= 4; page++ {
		query.Set("pageNumber", strconv.Itoa(page))
		attempt, cancel := context.WithTimeout(ctx, 3*time.Second)
		data, err := marketJSON[reply](attempt, s, "https://datacenter.eastmoney.com/api/data/v1/get", query)
		cancel()
		if err != nil || !data.Success {
			data, err = marketJSON[reply](ctx, s, "https://datacenter-web.eastmoney.com/api/data/v1/get", query)
		}
		if err != nil || !data.Success {
			if page == 1 {
				return MarketDataset{}, errMarketData
			}
			out.Message = "东方财富只返回了部分事件。"
			break
		}
		if data.Result == nil {
			break
		}
		for _, item := range data.Result.Data {
			from, _, _ := strings.Cut(item.Start, " ")
			to, _, _ := strings.Cut(item.End, " ")
			if _, err := time.Parse(time.DateOnly, from); err != nil || from >= end || item.Name == "" || item.Code == "" {
				continue
			}
			if _, err := time.Parse(time.DateOnly, to); err != nil || to < date {
				continue
			}
			span := from
			if to != "" && to != from {
				span += " 至 " + to
			}
			out.Rows = append(out.Rows, MarketRow{
				Code: item.Code + ":" + from, Name: item.Name, Date: span,
				Detail: item.Type, Actual: item.City,
			})
		}
		if page >= data.Result.Pages {
			break
		}
		if page == 4 {
			out.Message = "东方财富只返回了部分事件。"
		}
	}
	sort.SliceStable(out.Rows, func(i, j int) bool { return out.Rows[i].Date < out.Rows[j].Date })
	return out, nil
}

func (s *MarketService) kaipanlaLeaderboard(ctx context.Context, date string) (MarketDataset, error) {
	type reply struct {
		Time, Errcode string
		Total         int
		List          []struct {
			ID, Name, IncreaseAmount, BuyIn string
			JoinNum                         *int
		}
	}
	query := url.Values{
		"c": {"LongHuBang"}, "a": {"GetStockList"}, "Type": {"2"}, "Index": {"0"}, "st": {"100"},
		"Time": {date}, "apiv": {"w47"}, "PhoneOSNew": {"1"}, "VerSion": {"6.2.20.2"},
	}
	data, err := marketJSON[reply](ctx, s, "https://applhb.longhuvip.com/w1/api/index.php", query)
	if err != nil || data.Errcode != "0" || data.List == nil {
		return MarketDataset{}, errMarketData
	}
	if _, err := time.Parse(time.DateOnly, data.Time); err != nil || (date != "" && data.Time != date) {
		return MarketDataset{}, errMarketSession
	}
	out := MarketDataset{Source: "开盘啦", SourceURL: kaipanlaLeaderboardURL, Total: &data.Total}
	for _, item := range data.List {
		if item.ID == "" || item.Name == "" {
			continue
		}
		out.Rows = append(out.Rows, MarketRow{
			Code: item.ID, Symbol: stockSymbol(item.ID), Name: item.Name, Date: data.Time,
			Change: marketNumber(strings.TrimSuffix(item.IncreaseAmount, "%")), Net: marketNumber(item.BuyIn), Count: item.JoinNum,
		})
	}
	sort.SliceStable(out.Rows, func(i, j int) bool {
		a, b := out.Rows[i].Net, out.Rows[j].Net
		return a != nil && (b == nil || *a > *b)
	})
	return out, nil
}

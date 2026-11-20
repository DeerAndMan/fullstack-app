package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
)

const eastmoneyCalendarURL = "https://emdata.eastmoney.com/cjrl/index.html"
const kaipanlaLeaderboardURL = "https://www.kaipanla.com/omain/"
const thsCalendarURL = "https://stock.10jqka.com.cn/fincalendar.shtml"
const sinaCalendarURL = "https://finance.sina.com.cn/stock/usstock/"

func decodeMarketText(raw []byte) (string, error) {
	if utf8.Valid(raw) {
		return string(raw), nil
	}
	text, err := simplifiedchinese.GB18030.NewDecoder().Bytes(raw)
	return string(text), err
}

func (s *MarketService) sinaCalendar(ctx context.Context, date string) (MarketDataset, error) {
	type reply struct {
		Result struct {
			Status struct{ Code int }
			Data   struct {
				Data []struct {
					ID, Desc, Country, Importance string
					Published                     string `json:"publish_time"`
					Previous                      string `json:"pre_value"`
					Forecast                      string `json:"fur_value"`
					Actual                        string `json:"real_value"`
				}
			}
		}
	}
	start, _ := time.Parse(time.DateOnly, date)
	end := start.AddDate(0, 0, 6).Format(time.DateOnly)
	data, err := marketJSON[reply](ctx, s, "https://rl.cj.sina.com.cn/client/api/calendar_v2/get_economic_lists", url.Values{
		"page": {"1"}, "num": {"500"}, "platform": {"pc"}, "start_time": {date}, "end_time": {end},
	})
	if err != nil || data.Result.Status.Code != 0 || data.Result.Data.Data == nil {
		return MarketDataset{}, errMarketData
	}
	out := MarketDataset{Source: "新浪财经", SourceURL: sinaCalendarURL}
	for _, event := range data.Result.Data.Data {
		published, err := time.ParseInLocation(time.DateTime, event.Published, marketZone)
		if err != nil || event.ID == "" || event.Desc == "" {
			continue
		}
		if day := published.Format(time.DateOnly); day < date || day > end {
			continue
		}
		importance, _ := strconv.Atoi(event.Importance)
		out.Rows = append(out.Rows, MarketRow{
			Code: event.ID, Name: event.Desc, Detail: event.Country, Date: event.Published, Count: &importance,
			Previous: event.Previous, Forecast: event.Forecast, Actual: event.Actual,
		})
	}
	sort.SliceStable(out.Rows, func(i, j int) bool { return out.Rows[i].Date < out.Rows[j].Date })
	return out, nil
}

func (s *MarketService) thsCalendar(ctx context.Context, date string) (MarketDataset, error) {
	type reply struct {
		Stat string
		Data []struct {
			Date   string
			Events [][]string
		}
	}
	start, _ := time.Parse(time.DateOnly, date)
	end := start.AddDate(0, 0, 6)
	out := MarketDataset{Source: "同花顺", SourceURL: thsCalendarURL}
	seen := make(map[string]bool)
	failures := 0
	for month := start.AddDate(0, 0, 1-start.Day()); !month.After(end); month = month.AddDate(0, 1, 0) {
		raw, err := s.request(ctx, "https://comment.10jqka.com.cn/tzrl/getTzrlData.php", url.Values{
			"type": {"data"}, "date": {month.Format("200601")},
		})
		if err != nil {
			failures++
			continue
		}
		body, err := decodeMarketText(raw)
		var data reply
		if err != nil || json.Unmarshal([]byte(body), &data) != nil || data.Stat != "ok" || data.Data == nil {
			failures++
			continue
		}
		for _, day := range data.Data {
			eventDate, err := time.Parse(time.DateOnly, day.Date)
			if err != nil || eventDate.Before(start) || eventDate.After(end) || eventDate.Month() != month.Month() {
				continue
			}
			for _, event := range day.Events {
				if len(event) == 0 || strings.TrimSpace(event[0]) == "" {
					continue
				}
				title := strings.TrimSpace(event[0])
				key := day.Date + ":" + title
				if !seen[key] {
					seen[key] = true
					out.Rows = append(out.Rows, MarketRow{Code: key, Name: title, Date: day.Date})
				}
			}
		}
	}
	if failures > 0 {
		if len(out.Rows) == 0 {
			return MarketDataset{}, errMarketData
		}
		out.Message = "部分日期的同花顺投资日历暂不可用。"
	}
	sort.SliceStable(out.Rows, func(i, j int) bool { return out.Rows[i].Date < out.Rows[j].Date })
	return out, nil
}

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

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/encoding/simplifiedchinese"
)

var errMarketData = errors.New("market provider returned invalid data")

func marketJSON[T any](ctx context.Context, s *MarketService, endpoint string, query url.Values) (T, error) {
	var data T
	body, err := s.request(ctx, endpoint, query)
	if err == nil {
		err = json.Unmarshal(body, &data)
	}
	return data, err
}

func marketJSONNumber(raw json.RawMessage) *float64 {
	return marketNumber(strings.Trim(string(raw), `"`))
}

const tencentApp = "https://web.ifzq.gtimg.cn/appstock/app"
const eastmoneyToken = "7eea3edcaed734bea9cbfc24409ed989"

func (s *MarketService) quotes(symbols string) marketLoader {
	return func(ctx context.Context, _ string) (MarketDataset, error) {
		body, err := s.request(ctx, "https://qt.gtimg.cn/q="+symbols, nil)
		if err != nil {
			return MarketDataset{}, err
		}
		decoded, err := simplifiedchinese.GB18030.NewDecoder().Bytes(body)
		if err != nil {
			return MarketDataset{}, err
		}
		out := MarketDataset{Source: "腾讯财经"}
		for _, line := range strings.Split(string(decoded), "\n") {
			key, value, found := strings.Cut(line, "=\"")
			fields := strings.Split(value, "~")
			if !found || len(fields) < 35 {
				continue
			}
			price := marketNumber(fields[3])
			if price == nil || *price <= 0 {
				continue
			}
			row := MarketRow{
				Code: strings.TrimPrefix(key, "v_"), Name: fields[1], Date: quoteTimestamp(fields[30]),
				Price: price, Change: marketNumber(fields[32]), Previous: fields[4],
				Open: marketNumber(fields[5]), High: marketNumber(fields[33]), Low: marketNumber(fields[34]),
			}
			if marketStockCode.MatchString(row.Code) {
				row.Volume = stockVolumeShares(marketNumber(fields[6]), row.Code)
				if len(fields) > 35 {
					parts := strings.Split(fields[35], "/")
					if len(parts) == 3 {
						row.Amount = marketNumber(parts[2])
					}
				}
			}
			out.Rows = append(out.Rows, row)
		}
		if len(out.Rows) == 0 {
			return MarketDataset{}, errMarketData
		}
		return out, nil
	}
}

func (s *MarketService) daily(ctx context.Context, symbol string) (MarketDataset, error) {
	return s.kline("day")(ctx, symbol)
}

func (s *MarketService) minute(ctx context.Context, symbol string) (MarketDataset, error) {
	type reply struct {
		Code int
		Data map[string]struct {
			Data struct {
				Date string
				Data []string
			}
		}
	}
	data, err := marketJSON[reply](ctx, s, tencentApp+"/minute/query", url.Values{"code": {symbol}})
	if err != nil {
		return MarketDataset{}, err
	}
	out := MarketDataset{Source: "腾讯财经"}
	points := data.Data[symbol].Data
	date := points.Date
	if parsed, err := time.Parse("20060102", date); err == nil {
		date = parsed.Format(time.DateOnly)
	}
	for _, point := range points.Data {
		fields := strings.Fields(point)
		if len(fields) < 2 || len(fields[0]) != 4 || marketNumber(fields[1]) == nil {
			continue
		}
		clock := fields[0][:2] + ":" + fields[0][2:]
		out.Rows = append(out.Rows, MarketRow{Code: clock, Date: date, Price: marketNumber(fields[1])})
	}
	if data.Code != 0 || len(out.Rows) == 0 {
		return MarketDataset{}, errMarketData
	}
	return out, nil
}

const sinaBoards = "https://vip.stock.finance.sina.com.cn/q/view/"

func (s *MarketService) sectors(concept bool) marketLoader {
	if concept {
		return s.sinaSectorFeed("newFLJK.php", url.Values{"param": {"class"}})
	}
	return s.sinaSectorFeed("newSinaHy.php", nil)
}

func (s *MarketService) sinaSectorFeed(path string, query url.Values) marketLoader {
	return func(ctx context.Context, _ string) (MarketDataset, error) {
		endpoint := sinaBoards + path
		body, err := s.request(ctx, endpoint, query)
		if err != nil {
			return MarketDataset{}, err
		}
		decoded, err := simplifiedchinese.GB18030.NewDecoder().Bytes(body)
		if err != nil {
			return MarketDataset{}, err
		}
		_, payload, ok := strings.Cut(string(decoded), "=")
		var boards map[string]string
		if !ok || json.Unmarshal([]byte(strings.TrimSuffix(strings.TrimSpace(payload), ";")), &boards) != nil {
			return MarketDataset{}, errMarketData
		}
		out := MarketDataset{Source: "新浪财经"}
		for code, row := range boards {
			fields := strings.Split(row, ",")
			if len(fields) < 13 || marketNumber(fields[5]) == nil {
				continue
			}
			out.Rows = append(out.Rows, MarketRow{
				Code: code, Name: fields[1], Change: marketNumber(fields[5]), Amount: marketNumber(fields[7]),
				Detail: fields[12], Symbol: fields[8],
			})
		}
		if len(out.Rows) == 0 {
			return MarketDataset{}, errMarketData
		}
		sort.Slice(out.Rows, func(i, j int) bool {
			if *out.Rows[i].Change == *out.Rows[j].Change {
				return out.Rows[i].Code < out.Rows[j].Code
			}
			return *out.Rows[i].Change > *out.Rows[j].Change
		})
		return out, nil
	}
}

func (s *MarketService) turnover(ctx context.Context, _ string) (MarketDataset, error) {
	type quote struct {
		Symbol, Name, Trade, Ticktime string
		Changepercent, Amount         *float64
	}
	const endpoint = "https://vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php/Market_Center.getHQNodeData"
	data, err := marketJSON[[]quote](ctx, s, endpoint, url.Values{
		"page": {"1"}, "num": {"20"}, "sort": {"amount"}, "asc": {"0"}, "node": {"hs_a"},
	})
	if err != nil {
		return MarketDataset{}, err
	}
	out := MarketDataset{Source: "新浪财经"}
	for _, row := range data {
		out.Rows = append(out.Rows, MarketRow{
			Code: row.Symbol, Name: row.Name, Price: marketNumber(row.Trade),
			Change: row.Changepercent, Amount: row.Amount, Date: row.Ticktime,
		})
	}
	if len(out.Rows) == 0 {
		return MarketDataset{}, errMarketData
	}
	return out, nil
}

func (s *MarketService) fundFlow(ctx context.Context, _ string) (MarketDataset, error) {
	primaryCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	out, err := s.eastmoneyFundFlow(primaryCtx)
	cancel()
	if err == nil {
		return out, nil
	}
	return s.sinaFundFlow(ctx)
}

const sinaMoneyFlow = "https://vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php/MoneyFlow."

func (s *MarketService) sinaFundFlow(ctx context.Context) (MarketDataset, error) {
	type item struct {
		Code string `json:"category"`
		Name string `json:"name"`
		Rate string `json:"avg_changeratio"`
		Net  string `json:"netamount"`
	}
	const site = "https://vip.stock.finance.sina.com.cn"
	data, err := marketJSON[[]item](ctx, s, sinaMoneyFlow+"ssl_bkzj_bk",
		url.Values{"fenlei": {"zjh"}, "page": {"1"}, "num": {"100"}, "sort": {"netamount"}, "asc": {"0"}})
	if err != nil {
		return MarketDataset{}, err
	}
	out := MarketDataset{Source: "新浪财经", SourceURL: site + "/moneyflow/#blocktol_zjh"}
	for _, row := range data {
		net := marketNumber(row.Net)
		if row.Code == "" || row.Name == "" || net == nil {
			continue
		}
		change := marketNumber(row.Rate)
		if change != nil {
			*change *= 100
		}
		out.Rows = append(out.Rows, MarketRow{Code: row.Code, Name: row.Name, Change: change, Net: net})
	}
	if len(out.Rows) == 0 {
		return MarketDataset{}, errMarketData
	}
	sort.Slice(out.Rows, func(i, j int) bool { return *out.Rows[i].Net > *out.Rows[j].Net })
	return out, nil
}

func (s *MarketService) limitPool(pool string) marketLoader {
	type reply struct {
		RC   int
		Data *struct {
			Date  int `json:"qdate"`
			Total int `json:"tc"`
			Pool  []struct {
				Code   string   `json:"c"`
				Name   string   `json:"n"`
				Group  string   `json:"hybk"`
				Price  *float64 `json:"p"`
				Change *float64 `json:"zdp"`
				Amount *float64 `json:"amount"`
				Fund   *float64 `json:"fund"`
				Streak *int     `json:"lbc"`
			}
		}
	}
	return func(ctx context.Context, date string) (MarketDataset, error) {
		compact := strings.ReplaceAll(date, "-", "")
		const pageSize = 100
		sortBy := "fbt:asc"
		if pool == "getTopicDTPool" {
			sortBy = "fund:asc"
		}
		out := MarketDataset{Source: "东方财富"}
		for page := 0; page < 6; page++ {
			data, err := marketJSON[reply](ctx, s, "https://push2ex.eastmoney.com/"+pool, url.Values{
				"ut": {eastmoneyToken}, "dpt": {"wz.ztzt"}, "Pageindex": {strconv.Itoa(page)},
				"pagesize": {strconv.Itoa(pageSize)}, "sort": {sortBy}, "date": {compact},
			})
			if err != nil || data.RC != 0 {
				return MarketDataset{}, errMarketData
			}
			if data.Data == nil || strconv.Itoa(data.Data.Date) != compact {
				return MarketDataset{}, errMarketSession
			}
			out.Total = &data.Data.Total
			for _, row := range data.Data.Pool {
				var price *float64
				if row.Price != nil && *row.Price > 0 {
					value := *row.Price / 1000
					price = &value
				}
				out.Rows = append(out.Rows, MarketRow{
					Code: row.Code, Symbol: stockSymbol(row.Code), Name: row.Name, Detail: row.Group, Date: date,
					Price: price, Change: row.Change, Amount: row.Amount, Net: row.Fund, Count: row.Streak,
				})
			}
			if len(out.Rows) >= data.Data.Total || len(data.Data.Pool) < pageSize {
				return out, nil
			}
		}
		return out, nil
	}
}

const wallstreetAPI = "https://api-one.wallstcn.com/apiv1"

func (s *MarketService) flash(ctx context.Context, _ string) (MarketDataset, error) {
	type reply struct {
		Code int
		Data struct {
			Items []struct {
				ID     int64
				Title  string
				Text   string                   `json:"content_text"`
				Time   int64                    `json:"display_time"`
				Score  *int                     `json:"score"`
				Themes []struct{ Title string } `json:"related_themes"`
			}
		}
	}
	data, err := marketJSON[reply](ctx, s, wallstreetAPI+"/content/lives", url.Values{"channel": {"global-channel"}, "limit": {"24"}})
	if err != nil || data.Code != 20000 || data.Data.Items == nil {
		return MarketDataset{}, errMarketData
	}
	out := MarketDataset{Source: "华尔街见闻"}
	for _, row := range data.Data.Items {
		title := strings.TrimSpace(row.Title)
		if title == "" {
			runes := []rune(strings.TrimSpace(row.Text))
			title = string(runes[:min(70, len(runes))])
		}
		if title == "" {
			continue
		}
		topic := ""
		for _, theme := range row.Themes {
			if theme.Title != "" {
				topic = theme.Title
				if topic != "A股板块异动播报" {
					break
				}
			}
		}
		out.Rows = append(out.Rows, MarketRow{
			Code: strconv.FormatInt(row.ID, 10), Name: title, Date: time.Unix(row.Time, 0).In(marketZone).Format("2006-01-02 15:04:05"),
			URL: fmt.Sprintf("https://wallstreetcn.com/livenews/%d", row.ID), Priority: row.Score, Detail: topic,
		})
	}
	return out, nil
}

func (s *MarketService) calendar(ctx context.Context, date string) (MarketDataset, error) {
	type reply struct {
		Code int
		Data struct {
			Items []struct {
				ID                         int64
				Time                       int64 `json:"public_date"`
				Title, Country             string
				Previous, Forecast, Actual string
				Importance                 int
			}
		}
	}
	start, _ := time.ParseInLocation(time.DateOnly, date, marketZone)
	end := start.AddDate(0, 0, 7)
	data, err := marketJSON[reply](ctx, s, wallstreetAPI+"/finance/macrodatas", url.Values{
		"start": {strconv.FormatInt(start.Unix(), 10)}, "end": {strconv.FormatInt(end.Unix(), 10)},
	})
	if err != nil || data.Code != 20000 || data.Data.Items == nil {
		return MarketDataset{}, errMarketData
	}
	out := MarketDataset{Source: "华尔街见闻"}
	for _, row := range data.Data.Items {
		eventTime := time.Unix(row.Time, 0).In(marketZone)
		if eventTime.Before(start) || !eventTime.Before(end) || row.Title == "" {
			continue
		}
		out.Rows = append(out.Rows, MarketRow{
			Code: strconv.FormatInt(row.ID, 10), Name: row.Title, Detail: row.Country,
			Date: eventTime.Format("2006-01-02 15:04:05"), Count: &row.Importance,
			Previous: row.Previous, Forecast: row.Forecast, Actual: row.Actual,
		})
	}
	sort.Slice(out.Rows, func(i, j int) bool { return out.Rows[i].Date < out.Rows[j].Date })
	return out, nil
}

func eastmoneyData[T any](ctx context.Context, s *MarketService, query url.Values) ([]T, error) {
	type reply struct {
		Success bool
		Result  *struct{ Data []T }
	}
	data, err := marketJSON[reply](ctx, s, "https://datacenter-web.eastmoney.com/api/data/v1/get", query)
	if err != nil || !data.Success {
		return nil, errMarketData
	}
	if data.Result == nil {
		return nil, nil
	}
	return data.Result.Data, nil
}

func (s *MarketService) leaderboard(ctx context.Context, date string) (MarketDataset, error) {
	type item struct {
		Date   string   `json:"TRADE_DATE"`
		Code   string   `json:"SECURITY_CODE"`
		Name   string   `json:"SECURITY_NAME_ABBR"`
		Change *float64 `json:"CHANGE_RATE"`
		Net    *float64 `json:"BILLBOARD_NET_AMT"`
		Reason string   `json:"EXPLANATION"`
	}
	query := url.Values{
		"reportName": {"RPT_DAILYBILLBOARD_DETAILSNEW"},
		"columns":    {"TRADE_DATE,SECURITY_CODE,SECURITY_NAME_ABBR,CHANGE_RATE,BILLBOARD_NET_AMT,EXPLANATION"},
		"pageSize":   {"100"}, "pageNumber": {"1"}, "sortColumns": {"TRADE_DATE,BILLBOARD_NET_AMT"}, "sortTypes": {"-1,-1"},
	}
	if date != "" {
		query.Set("filter", fmt.Sprintf("(TRADE_DATE='%s')", date))
	}
	filterCtx, cancel := context.WithTimeout(ctx, 4*time.Second)
	data, err := eastmoneyData[item](filterCtx, s, query)
	cancel()
	if err != nil && date != "" {
		query.Set("filter", fmt.Sprintf("(TRADE_DATE='%s 00:00:00')", date))
		data, err = eastmoneyData[item](ctx, s, query)
	}
	if err != nil {
		return MarketDataset{}, err
	}
	out := MarketDataset{Source: "东方财富数据中心"}
	if date == "" && len(data) > 0 {
		date = marketDate(data[0].Date)
	}
	for i, row := range data {
		if marketDate(row.Date) == date {
			out.Rows = append(out.Rows, MarketRow{
				Code: fmt.Sprintf("%s-%d", row.Code, i), Symbol: stockSymbol(row.Code), Name: row.Name, Date: date, Detail: row.Reason,
				Change: row.Change, Net: row.Net,
			})
		}
	}
	return out, nil
}

func (s *MarketService) margin(ctx context.Context, _ string) (MarketDataset, error) {
	type item struct {
		Date    string   `json:"DIM_DATE"`
		Balance *float64 `json:"RZRQYE"`
		Loan    *float64 `json:"RZYE"`
		Net     *float64 `json:"RZJME"`
	}
	data, err := eastmoneyData[item](ctx, s, url.Values{
		"reportName": {"RPTA_RZRQ_LSHJ"}, "columns": {"DIM_DATE,RZRQYE,RZYE,RZJME"},
		"pageSize": {"60"}, "pageNumber": {"1"}, "sortColumns": {"DIM_DATE"}, "sortTypes": {"-1"},
	})
	if err != nil {
		return MarketDataset{}, err
	}
	out := MarketDataset{Source: "东方财富数据中心"}
	for _, row := range data {
		date := marketDate(row.Date)
		if row.Balance == nil {
			continue
		}
		out.Rows = append(out.Rows, MarketRow{Code: date, Date: date, Price: row.Loan, Amount: row.Balance, Net: row.Net})
	}
	if len(out.Rows) == 0 {
		return MarketDataset{}, errMarketData
	}
	return out, nil
}

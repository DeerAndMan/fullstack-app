package service

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type marketTestTransport func(*http.Request) (*http.Response, error)

func (fn marketTestTransport) RoundTrip(req *http.Request) (*http.Response, error) { return fn(req) }

func marketTestResponse(req *http.Request, status int, body string) *http.Response {
	return &http.Response{Request: req, StatusCode: status, Body: io.NopCloser(strings.NewReader(body))}
}

func TestXuanguLimitDownRejectsAnotherSessionsStocks(t *testing.T) {
	t.Parallel()
	svc := NewMarketService()
	svc.client.Transport = marketTestTransport(func(req *http.Request) (*http.Response, error) {
		return marketTestResponse(req, http.StatusOK, `{"code":20000,"data":[{"symbol":"600031.SH","stock_chi_name":"示例","last_limit_down":1795068495}]}`), nil
	})
	// The timestamp is on November 19 in Beijing.
	if _, err := svc.xuanguLimitDown(context.Background(), "2026-11-18"); err != errMarketSession {
		t.Fatalf("wrong session returned %v", err)
	}
	got, err := svc.xuanguLimitDown(context.Background(), "2026-11-19")
	if err != nil || len(got.Rows) != 1 || got.Rows[0].Symbol != "sh600031" {
		t.Fatalf("expected same-session stock, got %+v, %v", got, err)
	}
}

func TestSinaRegionFlowOmitsIncompleteRegions(t *testing.T) {
	t.Parallel()
	svc := NewMarketService()
	svc.client.Transport = marketTestTransport(func(req *http.Request) (*http.Response, error) {
		switch {
		case strings.HasSuffix(req.URL.Path, "/newFLJK.php"):
			return marketTestResponse(req, http.StatusOK, `var S_Finance_bankuai_area = {
				"diyu_110000":"diyu_110000,Beijing,2,1,1,1,1,1,sh600001,1,1,1,Leader",
				"diyu_120000":"diyu_120000,Tianjin,2,1,1,1,1,1,sh600001,1,1,1,Leader"
			};`), nil
		case strings.HasSuffix(req.URL.Path, "getHQNodeStockCount"):
			return marketTestResponse(req, http.StatusOK, `"2"`), nil
		case req.URL.Query().Get("bankuai") == "diyu_110000":
			return marketTestResponse(req, http.StatusOK, `[
				{"symbol":"sh600031","name":"First","netamount":"100"},
				{"symbol":"sz000001","name":"Second","netamount":"-20"}
			]`), nil
		default:
			return marketTestResponse(req, http.StatusOK, `[{"symbol":"sz000001","name":"First","netamount":"-20"}]`), nil
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	got, err := svc.sinaRegionFundFlow(ctx, "")
	if err != nil || len(got.Rows) != 1 || got.Rows[0].Code != "diyu_110000" || *got.Rows[0].Net != 80 || got.Message == "" {
		t.Fatalf("expected complete regional total, got %+v, %v", got, err)
	}
}

func TestTHSRegionFundFlowSkipsUnpublishedTotals(t *testing.T) {
	t.Parallel()
	svc := NewMarketService()
	svc.client.Transport = marketTestTransport(func(req *http.Request) (*http.Response, error) {
		return marketTestResponse(req, http.StatusOK, `<table><tr><td>1</td>
		<td><a href="http://q.10jqka.com.cn/dy/detail/code/882033/">浙江</a></td>
		<td>0.00</td><td>0.00</td><td>0.00</td><td>0.00</td><td>82</td><td>672</td><td>0</td><td>--</td>
		</tr></table>`), nil
	})
	got, err := svc.thsRegionFundFlow(context.Background(), "")
	if err != nil || len(got.Rows) != 0 || got.SourceURL != thsRegionURL || got.Message == "" {
		t.Fatalf("expected unpublished regional flow, got %+v, %v", got, err)
	}
}

func TestTHSStockPoolsRejectMismatchedDate(t *testing.T) {
	t.Parallel()
	svc := NewMarketService()
	svc.client.Transport = marketTestTransport(func(req *http.Request) (*http.Response, error) {
		return marketTestResponse(req, http.StatusOK, `{"status_code":0,"data":{"date":"19990101","page":{"total":1},"info":[{"code":"000001","name":"样本"}]}}`), nil
	})
	for _, feed := range []string{"limit-up-ths", "opened-limit-ths", "limit-down-ths"} {
		got, err := svc.Feed(context.Background(), feed, "", "")
		if err != nil || got.Status != "unavailable" || len(got.Rows) != 0 {
			t.Errorf("%s returned a different date: %+v, err = %v", feed, got, err)
		}
	}
}

func TestTHSConceptQuotesUseTheConceptTableColumns(t *testing.T) {
	t.Parallel()
	svc := NewMarketService()
	svc.client.Transport = marketTestTransport(func(req *http.Request) (*http.Response, error) {
		return marketTestResponse(req, http.StatusOK, `<meta charset="utf-8">
		<table><tr><td>1</td><td><a href="http://q.10jqka.com.cn/gn/detail/code/307816/">概念</a></td>
		<td>1000</td><td>0.35%</td><td>10.52</td><td>8.3</td><td>2.22</td><td>13</td>
		<td><a href="http://stockpage.10jqka.com.cn/688172/">样本</a></td><td>1.9%</td><td>12.20</td>
		</tr></table><span class="page_info">1/1</span>`), nil
	})
	got, err := svc.Feed(context.Background(), "concept-ths", "", "")
	if err != nil || got.Status != "ready" || len(got.Rows) != 1 {
		t.Fatalf("concept dataset = %+v, err = %v", got, err)
	}
	row := got.Rows[0]
	if row.Code != "307816" || row.Name != "概念" || row.Symbol != "sh688172" || row.Amount != nil ||
		row.Change == nil || *row.Change != 0.35 || row.Net == nil || *row.Net != 222000000 {
		t.Fatalf("unexpected concept row: %+v", row)
	}
}

func TestMarketFundFlowFallsBackToSina(t *testing.T) {
	t.Parallel()
	svc := NewMarketService()
	svc.client.Transport = marketTestTransport(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Host {
		case "push2.eastmoney.com":
			return marketTestResponse(req, http.StatusBadGateway, ""), nil
		case "vip.stock.finance.sina.com.cn":
			return marketTestResponse(req, http.StatusOK, `[
				{"category":"energy","name":"能源","avg_changeratio":"-0.02","netamount":"-150000"},
				{"category":"chips","name":"半导体","avg_changeratio":"0.01","netamount":"200000"}
			]`), nil
		default:
			t.Errorf("unexpected market provider %s", req.URL.Host)
			return marketTestResponse(req, http.StatusBadRequest, ""), nil
		}
	})

	got, err := svc.Feed(context.Background(), "fund-flow", "", "")
	if err != nil || got.Status != "ready" || got.Source != "新浪财经" || len(got.Rows) != 2 {
		t.Fatalf("fallback dataset = %+v, err = %v", got, err)
	}
	first := got.Rows[0]
	if first.Code != "chips" || first.Change == nil || *first.Change != 1 || first.Net == nil || *first.Net != 200000 {
		t.Fatalf("fallback row = %+v", first)
	}
}

func TestMarketRegionFundFlowLoadsMembersAcrossPages(t *testing.T) {
	t.Parallel()
	svc := NewMarketService()
	svc.client.Transport = marketTestTransport(func(req *http.Request) (*http.Response, error) {
		query := req.URL.Query()
		if req.URL.Host == "push2test.eastmoney.com" {
			return marketTestResponse(req, http.StatusBadGateway, ""), nil
		}
		switch query.Get("fs") {
		case "m:90 t:1":
			return marketTestResponse(req, http.StatusOK, `{"rc":0,"data":{"total":1,"diff":[
				{"f12":"BK0152","f14":"甘肃板块","f3":-1.2,"f62":200000}
			]}}`), nil
		case "b:BK0152 f:!2":
			if query.Get("pn") == "1" {
				return marketTestResponse(req, http.StatusOK, `{"rc":0,"data":{"total":2,"diff":[
					{"f12":"603999","f14":"读者传媒","f2":5.53,"f3":-1.04,"f62":-2000}
				]}}`), nil
			}
			return marketTestResponse(req, http.StatusOK, `{"rc":0,"data":{"total":2,"diff":[
				{"f12":"000779","f14":"甘咨询","f2":9.1,"f3":0.08,"f62":"-"}
			]}}`), nil
		default:
			t.Fatalf("unexpected market filter %q", query.Get("fs"))
			return nil, nil
		}
	})

	catalog, err := svc.Feed(context.Background(), "region-fund-flow", "", "")
	if err != nil || catalog.Status != "ready" || len(catalog.Rows) != 1 || catalog.Rows[0].Code != "BK0152" {
		t.Fatalf("region catalog = %+v, err = %v", catalog, err)
	}
	members, err := svc.Feed(context.Background(), "board-members", catalog.Rows[0].Code, "")
	if err != nil || members.Status != "ready" || len(members.Rows) != 2 || members.Message != "" {
		t.Fatalf("region members = %+v, err = %v", members, err)
	}
	if first, second := members.Rows[0], members.Rows[1]; first.Symbol != "sh603999" ||
		first.Net == nil || *first.Net != -2000 || second.Symbol != "sz000779" || second.Net != nil {
		t.Fatalf("unexpected region quotes: %+v", members.Rows)
	}
}

func TestMarketSinaIndustryMembersIncludeFundFlow(t *testing.T) {
	t.Parallel()
	svc := NewMarketService()
	svc.client.Transport = marketTestTransport(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Path {
		case "/quotes_service/api/json_v2.php/Market_Center.getHQNodeStockCount":
			return marketTestResponse(req, http.StatusOK, `"1"`), nil
		case "/quotes_service/api/json_v2.php/MoneyFlow.ssl_bkzj_ssggzj":
			if req.URL.Query().Get("bankuai") != "new_jxhy" {
				t.Errorf("unexpected board: %s", req.URL)
			}
			return marketTestResponse(req, http.StatusOK, `[
				{"symbol":"sh600031","name":"三一重工","trade":"21.85","changeratio":"0.012","netamount":"20000"}
			]`), nil
		default:
			t.Fatalf("unexpected market request: %s", req.URL)
			return nil, nil
		}
	})

	got, err := svc.Feed(context.Background(), "board-members", "new_jxhy", "")
	if err != nil || got.Status != "ready" || len(got.Rows) != 1 {
		t.Fatalf("industry members = %+v, err = %v", got, err)
	}
	if row := got.Rows[0]; row.Code != "600031" || row.Symbol != "sh600031" ||
		row.Change == nil || *row.Change != 1.2 || row.Net == nil || *row.Net != 20000 {
		t.Fatalf("industry stock = %+v", row)
	}
}

func TestMarketRejectsLimitPoolForWrongDate(t *testing.T) {
	t.Parallel()
	svc := NewMarketService()
	svc.client.Transport = marketTestTransport(func(req *http.Request) (*http.Response, error) {
		return marketTestResponse(req, http.StatusOK, `{"rc":0,"data":{"qdate":19990101,"tc":1,"pool":[{"c":"000001","n":"样本"}]}}`), nil
	})
	got, err := svc.Feed(context.Background(), "limit-up", "", "")
	if err != nil || got.Status != "unavailable" || got.Message == "" || len(got.Rows) != 0 {
		t.Fatalf("wrong-session dataset = %+v, err = %v", got, err)
	}
}

func TestStockFiveDayOrdersSessionsAndDecodesMinuteVolume(t *testing.T) {
	t.Parallel()
	svc := NewMarketService()
	svc.client.Transport = marketTestTransport(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path != "/appstock/app/day/query" || req.URL.Query().Get("code") != "sh601155" {
			t.Errorf("unexpected provider request: %s", req.URL)
		}
		return marketTestResponse(req, http.StatusOK, `{"code":0,"data":{"sh601155":{"data":[
			{"date":"20261119","data":["0930 12.00 3 3600","0931 12.10 5 6020","1530 12.10 7 8440"]},
			{"date":"20261118","data":["1500 11.80 10 11800"]}
		]}}}`), nil
	})
	got, err := svc.Feed(context.Background(), "stock-five-day", "sh601155", "")
	if err != nil || got.Status != "ready" || len(got.Rows) != 3 {
		t.Fatalf("five-day dataset = %+v, err = %v", got, err)
	}
	if got.Rows[0].Date != "2026-11-18 15:00" || got.Rows[1].Date != "2026-11-19 09:30" ||
		got.Rows[1].Volume == nil || *got.Rows[1].Volume != 300 {
		t.Fatalf("sessions were not ordered or volume did not reset: %+v", got.Rows)
	}
	last := got.Rows[2]
	if last.Volume == nil || *last.Volume != 200 || last.Average == nil || *last.Average != 12.04 {
		t.Fatalf("unexpected minute volume or average: %+v", last)
	}
}

func TestStockKlineUsesRequestedPeriodAndRecordsAdjustment(t *testing.T) {
	t.Parallel()
	svc := NewMarketService()
	svc.client.Transport = marketTestTransport(func(req *http.Request) (*http.Response, error) {
		var series string
		switch req.URL.Query().Get("param") {
		case "sh601155,week,,,120,qfq":
			series = "qfqweek"
		case "sh601155,month,,,120,qfq":
			series = "month"
		default:
			t.Errorf("unexpected provider request: %s", req.URL)
		}
		return marketTestResponse(req, http.StatusOK,
			`{"code":0,"data":{"sh601155":{"`+series+`":[["2026-11-13","12.54","11.72","12.63","11.69","490235",{"nd":"2025"}]]}}}`), nil
	})
	for feed, adjustment := range map[string]string{"stock-weekly": "qfq", "stock-monthly": "none"} {
		got, err := svc.Feed(context.Background(), feed, "sh601155", "")
		if err != nil || got.Status != "ready" || got.Adjustment != adjustment || len(got.Rows) != 1 {
			t.Fatalf("%s dataset = %+v, err = %v", feed, got, err)
		}
		row := got.Rows[0]
		if row.High == nil || *row.High != 12.63 || row.Volume == nil || *row.Volume != 49023500 {
			t.Fatalf("%s candle = %+v", feed, row)
		}
	}
}

func TestMarketLeaderboardRetriesHistoricalFilter(t *testing.T) {
	t.Parallel()
	svc := NewMarketService()
	const date = "2025-01-09"
	calls := 0
	svc.client.Transport = marketTestTransport(func(req *http.Request) (*http.Response, error) {
		calls++
		switch req.URL.Query().Get("filter") {
		case "(TRADE_DATE='2025-01-09')":
			return marketTestResponse(req, http.StatusBadGateway, ""), nil
		case "(TRADE_DATE='2025-01-09 00:00:00')":
			return marketTestResponse(req, http.StatusOK, `{"success":true,"result":{"data":[{"TRADE_DATE":"2025-01-09 00:00:00","SECURITY_CODE":"000001","SECURITY_NAME_ABBR":"样本"}]}}`), nil
		default:
			t.Errorf("unexpected filter: %s", req.URL.Query().Get("filter"))
			return marketTestResponse(req, http.StatusBadRequest, ""), nil
		}
	})
	got, err := svc.leaderboard(context.Background(), date)
	if err != nil || calls != 2 || len(got.Rows) != 1 || got.Rows[0].Date != date || got.Rows[0].Symbol != "sz000001" {
		t.Fatalf("historical dataset = %+v, calls = %d, err = %v", got, calls, err)
	}
}

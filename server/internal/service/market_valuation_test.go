package service

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestStockValuationReadsEveryPageAndKeepsMissingFields(t *testing.T) {
	now := time.Now().In(marketZone)
	latest := now.Format(time.DateOnly)
	earliest := now.AddDate(0, 0, -valuationPageSize).Format(time.DateOnly)
	value := 7.5
	negative := -4.13
	s := NewMarketService()
	s.client.Transport = marketTestTransport(func(req *http.Request) (*http.Response, error) {
		query := req.URL.Query()
		if query.Get("reportName") != "RPT_VALUEANALYSIS_DET" || !strings.Contains(query.Get("filter"), `(SECUCODE="600519.SH")`) {
			t.Errorf("unexpected valuation query: %s", req.URL)
		}
		page := valuationPage{Pages: 2, Count: valuationPageSize + 1}
		switch query.Get("pageNumber") {
		case "1":
			for i := range valuationPageSize {
				page.Data = append(page.Data, valuationRecord{
					Symbol: "600519.SH", Date: now.AddDate(0, 0, -i).Format(time.DateOnly) + " 00:00:00", PB: &value,
				})
			}
		case "2":
			page.Data = []valuationRecord{{Symbol: "600519.SH", Date: earliest, PCF: &negative}}
		default:
			t.Errorf("unexpected page: %s", req.URL)
		}
		body, _ := json.Marshal(map[string]any{"success": true, "result": page})
		return marketTestResponse(req, http.StatusOK, string(body)), nil
	})

	got, err := s.Feed(context.Background(), "stock-valuation", "sh600519", "")
	if err != nil || got.Status != "ready" || len(got.Rows) != valuationPageSize+1 {
		t.Fatalf("status %s, rows %d, error %v", got.Status, len(got.Rows), err)
	}
	first, last := got.Rows[0], got.Rows[len(got.Rows)-1]
	if first.Date != earliest || first.PBMRQ != nil || first.PCFOCFTTM == nil || *first.PCFOCFTTM != negative {
		t.Fatalf("unexpected earliest row: %+v", first)
	}
	if last.Date != latest || last.PBMRQ == nil || *last.PBMRQ != value || last.PETTM != nil {
		t.Fatalf("unexpected latest row: %+v", last)
	}
}

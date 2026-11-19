package service

import (
	"context"
	"net/http"
	"testing"
)

func TestConceptMembersDecodeMixedMarketNumbers(t *testing.T) {
	t.Parallel()
	market := NewMarketService()
	market.client.Transport = marketTestTransport(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/quotes_service/api/json_v2.php/Market_Center.getHQNodeStockCount" {
			return marketTestResponse(req, http.StatusOK, `"5"`), nil
		}
		return marketTestResponse(req, http.StatusOK, `[
			{"code":"300024","name":"机器人","changepercent":1.3},
			{"code":"000609","name":"样本","changepercent":"0.000"},
			{"code":"002050","name":"三花智控","changepercent":null},
			{"code":"000913","name":"下跌样本","trade":"10.83","changepercent":"-0.37"},
			{"code":"000001","name":"无报价样本","trade":"0.00","changepercent":"0.000"}
		]`), nil
	})
	data, err := market.conceptMembers(context.Background(), "gn_zjqrgn")
	if err != nil || len(data.Rows) != 5 || data.Message != "" {
		t.Fatalf("members = %+v, err = %v", data, err)
	}
	if data.Rows[0].Change == nil || *data.Rows[0].Change != 1.3 ||
		data.Rows[1].Change == nil || *data.Rows[1].Change != 0 || data.Rows[2].Change != nil ||
		data.Rows[3].Change == nil || *data.Rows[3].Change != -0.37 || data.Rows[4].Change != nil {
		t.Fatalf("unexpected market values: %+v", data.Rows)
	}
}

func TestResearchIntersectionDistinguishesFullAndSplitCoverage(t *testing.T) {
	t.Parallel()
	a := MarketRow{Code: "002050", Name: "三花智控"}
	b := MarketRow{Code: "601689", Name: "拓普集团"}
	c := MarketRow{Code: "300024", Name: "机器人"}
	d := MarketRow{Code: "300001", Name: "测试"}
	boards := []researchBoard{
		{Provider: "新浪财经", Topic: "机器人", Rows: []MarketRow{a, b, c, d}},
		{Provider: "新浪财经", Topic: "特斯拉", Rows: []MarketRow{a, c}},
		{Provider: "同花顺", Topic: "机器人", Rows: []MarketRow{a}},
		{Provider: "同花顺", Topic: "特斯拉", Rows: []MarketRow{a, b}},
	}
	got, sources := compareResearchBoards(boards, []string{"机器人", "特斯拉"}, "all")
	if len(got) != 3 || len(sources) != 4 {
		t.Fatalf("candidates = %+v, sources = %+v", got, sources)
	}
	if got[0].Code != a.Code || got[0].Status != "double" ||
		got[1].Code != b.Code || got[1].Status != "combined" ||
		got[2].Code != c.Code || got[2].Status != "single" {
		t.Fatalf("unexpected evidence tiers: %+v", got)
	}
}

func TestResearchRejectsObservationWithoutMatchingEvidence(t *testing.T) {
	t.Parallel()
	market := NewMarketService()
	market.client.Transport = marketTestTransport(func(req *http.Request) (*http.Response, error) {
		return marketTestResponse(req, http.StatusOK, `{"QuotationCodeTable":{"Data":[
			{"Code":"601689","Name":"拓普集团","Classify":"AStock","QuoteID":"1.601689"}
		]}}`), nil
	})
	s := NewTopicResearchService(nil, market, false)
	found := s.verifyResearchObservations(context.Background(), []ResearchObservation{
		{Code: "601689", Name: "拓普集团", Note: "查阅报道原文", SourceIDs: []string{"S1", "bad"}},
		{Code: "123456", Name: "杜撰名称", Note: "无相关来源", SourceIDs: []string{"S1"}},
	}, nil, []ResearchSource{{ID: "S1", Kind: "search", Title: "报道提及拓普集团"}})
	if len(found) != 1 || found[0].QuoteURL != "https://quote.eastmoney.com/unify/r/1.601689" || len(found[0].SourceIDs) != 1 {
		t.Fatalf("validated observations = %+v", found)
	}
}

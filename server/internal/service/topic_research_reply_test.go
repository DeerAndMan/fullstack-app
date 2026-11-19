package service

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestResearchReplyCarriesHistoryWithoutSkills(t *testing.T) {
	s := NewTopicResearchService(nil, nil, false)
	s.client.Transport = marketTestTransport(func(req *http.Request) (*http.Response, error) {
		var body struct {
			Stream   bool
			Messages []struct{ Content string }
		}
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if !body.Stream || len(body.Messages) != 2 || !strings.Contains(body.Messages[1].Content, "先前的问题") {
			t.Fatal("missing conversation context or streaming")
		}
		return marketTestResponse(req, http.StatusOK, "data: {\"choices\":[{\"delta\":{\"content\":\"回答\"}}]}\n\ndata: {\"choices\":[{\"delta\":{\"content\":\"正文\"},\"finish_reason\":\"stop\"}]}\n\n"), nil
	})
	reply := s.streamReply(context.Background(), 7, "example", "请继续", nil, &ResearchRunView{History: []string{"先前的问题"}}, ResearchReport{})
	if reply != "回答正文" || s.Draft(7) != reply || s.Draft(8) != "" {
		t.Fatalf("unexpected streamed reply %q", reply)
	}
}

func TestSupplyReplyLimitsMissingEvidenceToCurrentSearch(t *testing.T) {
	s := NewTopicResearchService(nil, nil, false)
	s.client.Transport = marketTestTransport(func(req *http.Request) (*http.Response, error) {
		return marketTestResponse(req, http.StatusOK, "data: {\"choices\":[{\"delta\":{\"content\":\"没有公\"}}]}\n\ndata: {\"choices\":[{\"delta\":{\"content\":\"司供货。样本公司披露送样。\"},\"finish_reason\":\"stop\"}]}\n\n"), nil
	})
	report := ResearchReport{Optimus: &OptimusReport{}}
	reply := s.streamReply(context.Background(), 9, "example", "哪些公司有公告确认供货", nil, nil, report)
	if reply != noSupplyAnnouncement+"\n\n样本公司披露送样。" || s.Draft(9) != reply {
		t.Fatalf("unexpected reply %q", reply)
	}
	evidence := OptimusEvidence{Status: "confirmed", SourceID: "S1", FullText: true}
	report.Optimus.Stocks = []OptimusStock{{Evidence: []OptimusEvidence{evidence}}}
	report.Sources = []ResearchSource{{ID: "S1", Provider: "cninfo.com.cn", URL: "https://static.cninfo.com.cn/finalpage/example.PDF"}}
	if needsSupplyQualification("哪些公司有公告确认供货", report) {
		t.Fatal("qualified a verified announcement as missing")
	}
}

func TestCombinedSkillReplyUsesVerifiedNotesWithoutRewriting(t *testing.T) {
	s := NewTopicResearchService(nil, nil, false)
	s.client.Transport = marketTestTransport(func(req *http.Request) (*http.Response, error) {
		t.Fatal("combined evidence should not be rewritten by another completion")
		return nil, nil
	})
	report := ResearchReport{SkillChecks: []ResearchSkillCheck{
		{Method: "risks"},
		{Method: "orders", Findings: []ResearchSkillFinding{{
			Subject: "样本公司", Note: "说明会披露送样。",
			Evidence: []ResearchSkillEvidence{{SourceID: "E1"}, {SourceID: "E1"}, {SourceID: "E2"}},
		}}},
		{Method: "industry"},
	}}
	got := s.streamReply(context.Background(), 15, "example", "核验订单兑现、行业景气和风险反证", nil, nil, report)
	want := "订单兑现\n样本公司 · 说明会披露送样。（E1、E2）\n\n行业景气\n本轮未取得足以支持结论的原文。\n\n风险反证\n本轮未取得足以支持结论的原文。"
	if got != want || s.Draft(15) != want {
		t.Fatalf("combined reply = %q, want %q", got, want)
	}
}

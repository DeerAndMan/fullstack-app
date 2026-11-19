package service

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestConceptReasonUsesCuratedEvidence(t *testing.T) {
	svc := NewMarketService()
	svc.client.Transport = marketTestTransport(func(req *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected upstream request: %s", req.URL)
		return nil, nil
	})
	reason, err := svc.ConceptReason(context.Background(), "gn_hwhs", "sh688037")
	if err != nil || reason.Status != "ready" || !strings.Contains(reason.Text, "单片式湿法设备") || len(reason.Sources) == 0 {
		t.Fatalf("reason = %+v, err = %v", reason, err)
	}
}

func TestConceptReasonChecksTopicAndStockForProfileEvidence(t *testing.T) {
	svc := NewMarketService()
	calls := 0
	svc.client.Transport = marketTestTransport(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Host {
		case "vip.stock.finance.sina.com.cn":
			return marketTestResponse(req, http.StatusOK, `var data = {"gn_a":"gn_a,AI,,0,0,1,0,0,0,0,0,0,","gn_demo":"gn_demo,\u6837\u672c\u9898\u6750,,0,0,1,0,0,0,0,0,0,"};`), nil
		case "emweb.eastmoney.com":
			calls++
			if req.URL.Query().Get("code") != "SZ000001" {
				t.Errorf("wrong stock requested: %s", req.URL)
			}
			return marketTestResponse(req, http.StatusOK, `{"hxtc":[
				{"SECURITY_CODE":"000002","KEYWORD":"AI","MAINPOINT_CONTENT":"这是其他股票的材料"},
				{"SECURITY_CODE":"000001","KEYWORD":"样本题材","MAINPOINT_CONTENT":"样本题材所在行业增长。","KEY_CLASSIF":"行业背景"},
				{"SECURITY_CODE":"000001","KEYWORD":"业务说明","MAINPOINT_CONTENT":"公司对样本题材有研发，尚未形成订单。","KEY_CLASSIF":"主营业务"}
			]}`), nil
		default:
			t.Errorf("unexpected provider: %s", req.URL)
			return marketTestResponse(req, http.StatusBadRequest, ""), nil
		}
	})

	ctx := context.Background()
	reason, err := svc.ConceptReason(ctx, "gn_demo", "sz000001")
	if err != nil || reason.Status != "ready" || !strings.Contains(reason.Text, "尚未形成订单") ||
		len(reason.Sources) != 1 || !strings.Contains(reason.Sources[0].URL, "code=SZ000001") {
		t.Fatalf("matched reason = %+v, err = %v", reason, err)
	}
	missing, err := svc.ConceptReason(ctx, "gn_a", "sz000001")
	if err != nil || missing.Status != "missing" || missing.Text != "" || calls != 1 {
		t.Fatalf("unmatched reason = %+v, err = %v, profile requests = %d", missing, err, calls)
	}
	if _, err := svc.ConceptReason(ctx, "gn_hwhs", "../../foo"); !errors.Is(err, ErrMarketQuery) {
		t.Fatalf("invalid symbol error = %v", err)
	}
}

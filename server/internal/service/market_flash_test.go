package service

import (
	"context"
	"net/http"
	"testing"
)

func TestFlashSourcesUsePublisherHeadlinesAndLinks(t *testing.T) {
	t.Parallel()
	svc := NewMarketService()
	svc.client.Transport = marketTestTransport(func(req *http.Request) (*http.Response, error) {
		switch req.URL.Host {
		case "zhibo.sina.com.cn":
			return marketTestResponse(req, http.StatusOK, `{"result":{"status":{"code":0},"data":{"feed":{"list":[
				{"id":1,"rich_text":"【<b>样本</b>公司公告】正文","create_time":"2026-11-19 15:00:00","docurl":"https://finance.sina.cn/7x24/detail.d.html"},
				{"id":2,"rich_text":"另一条快讯","create_time":"2026-11-19 14:59:00","docurl":"javascript:alert(1)"}
			]}}}}`), nil
		case "np-weblist.eastmoney.com":
			if req.URL.Query().Get("fastColumn") != "101" {
				t.Errorf("expected the focus channel, got %s", req.URL)
			}
			return marketTestResponse(req, http.StatusOK, `{"code":"1","data":{"fastNewsList":[
				{"code":"202611193935001528","title":"样本财经动态","showTime":"2026-11-19 15:00:00"}
			]}}`), nil
		default:
			t.Errorf("unexpected provider: %s", req.URL)
			return marketTestResponse(req, http.StatusBadRequest, ""), nil
		}
	})

	ctx := context.Background()
	sina, err := svc.Feed(ctx, "flash-sina", "", "")
	if err != nil || sina.Status != "ready" || len(sina.Rows) != 2 {
		t.Fatalf("Sina = %+v, err = %v", sina, err)
	}
	if first := sina.Rows[0]; first.Name != "样本 公司公告" || first.URL != "https://finance.sina.cn/7x24/detail.d.html" || sina.Rows[1].URL != sinaFlashURL {
		t.Fatalf("unexpected Sina rows: %+v", sina.Rows)
	}
	eastmoney, err := svc.Feed(ctx, "flash-eastmoney", "", "")
	if err != nil || eastmoney.Status != "ready" || len(eastmoney.Rows) != 1 ||
		eastmoney.Rows[0].URL != "https://finance.eastmoney.com/a/202611193935001528.html" {
		t.Fatalf("Eastmoney = %+v, err = %v", eastmoney, err)
	}
}

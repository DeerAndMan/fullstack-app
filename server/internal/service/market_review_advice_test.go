package service

import (
	"encoding/json"
	"testing"
)

func TestReviewStyleLabelsUseAverageIndexAndDailySpread(t *testing.T) {
	raw := json.RawMessage(`{"rows":[{"code":"880865","close":12,"ema5":9,"ema13":8},{"code":"880003","close":9,"ema5":10,"ema13":9}],"spread":-1.14}`)
	avg, style, err := reviewStyleLabels(raw)
	if err != nil || avg != "低于EMA5、等于EMA13" || style != "新高弱于平均股价" {
		t.Fatalf("%s %s %v", avg, style, err)
	}
	avg, style, err = reviewStyleLabels(nil)
	if err != nil || avg != "" || style != "" {
		t.Fatal("missing style was classified")
	}
}

package service

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//go:embed market_history_backup.json
var historyBackupArchive []byte

const reportRaw = "https://raw.githubusercontent.com/ztangz/daily-a-share-market-report/main/data/processed/"
const reportName = "A股每日复盘"
const zhishiName = "知市综合情绪（第一版）"

type backupHistoryDay struct {
	Date            string
	Up, Down        int
	Score, AmountYi float64
	Source          string
}

func (d backupHistoryDay) valid() bool {
	return d.Source != "" && d.Up >= 0 && d.Down >= 0 && d.Up+d.Down > 0 && d.AmountYi > 0 && d.Score >= 0 && d.Score <= 100
}

func (d backupHistoryDay) link() string {
	if d.Source == zhishiName {
		return "https://zhishi.ltd/replay?date=" + d.Date + "&historyDays=60&tab=sentiment"
	}
	return reportRaw + d.Date + "/market.json"
}

// The Shanghai index supplies trading dates only. A non-trading weekday must never become an observation.
func (s *MarketService) historyCalendar(ctx context.Context, start, end string) ([]string, error) {
	var dates []string
	last, _ := time.Parse(time.DateOnly, end)
	first := last.AddDate(0, 0, -60)
	if start != "" {
		first, _ = time.Parse(time.DateOnly, start)
	}
	today := time.Now().In(marketZone).Format(time.DateOnly)
	for year := last.Year(); year >= first.Year(); year-- {
		records, err := s.thsIndexKlines(ctx, "1A0001", year)
		if err != nil {
			return dates, err
		}
		for _, rec := range records {
			s, _, _ := strings.Cut(rec, ",")
			t, err := time.Parse("20060102", s)
			if err != nil {
				continue
			}
			d := t.Format(time.DateOnly)
			if !t.Before(first) && d <= end && d < today {
				dates = append(dates, d)
			}
		}
	}
	return dates, nil
}

func (s *MarketService) thsIndexKlines(ctx context.Context, code string, year int) ([]string, error) {
	endpoint := fmt.Sprintf("https://d.10jqka.com.cn/v6/line/hs_%s/01/%d.js", code, year)
	raw, err := s.request(ctx, endpoint, nil)
	if err != nil {
		return nil, err
	}
	prefix := fmt.Sprintf("quotebridge_v6_line_hs_%s_01_%d(", code, year)
	body, ok := strings.CutPrefix(strings.TrimSpace(string(raw)), prefix)
	body = strings.TrimSuffix(strings.TrimSuffix(body, ";"), ")")
	var kline struct{ Data string }
	if !ok || json.Unmarshal([]byte(body), &kline) != nil || kline.Data == "" {
		return nil, errMarketData
	}
	return strings.Split(kline.Data, ";"), nil
}

func (s *MarketService) reportBackup(ctx context.Context, date string) backupHistoryDay {
	var p struct {
		Summary struct {
			Date  string   `json:"trade_date"`
			Up    *int     `json:"advancers"`
			Down  *int     `json:"decliners"`
			Score *float64 `json:"emotion_score"`
		}
		Indices []struct {
			Code   string
			Amount float64 `json:"amount_yi"`
		}
	}
	body, err := s.request(ctx, reportRaw+date+"/market.json", nil)
	if err != nil || json.Unmarshal(body, &p) != nil {
		return backupHistoryDay{}
	}
	v := p.Summary
	if v.Date != date || v.Up == nil || v.Down == nil || v.Score == nil {
		return backupHistoryDay{}
	}
	amounts := map[string]float64{}
	for _, i := range p.Indices {
		amounts[i.Code] = i.Amount
	}
	if amounts["000001"] <= 0 || amounts["399001"] <= 0 {
		return backupHistoryDay{}
	}
	return backupHistoryDay{date, *v.Up, *v.Down, *v.Score, amounts["000001"] + amounts["399001"], reportName}
}

var historyHTMLTags = regexp.MustCompile(`<[^>]+>`)
var zhishiScore = regexp.MustCompile(`([0-9-]{10}) · 收盘统计 综合情绪因子 ([0-9.]+) / 100 · v1\b`)
var zhishiBreadth = regexp.MustCompile(`上涨 ([0-9,]+) · 下跌 ([0-9,]+) · 平盘 [0-9,]+ 家 · 成交额 ([0-9,.]+) 亿元`)

func (s *MarketService) zhishiBackup(ctx context.Context, date string) backupHistoryDay {
	body, err := s.request(ctx, "https://zhishi.ltd/replay", url.Values{"date": {date}, "historyDays": {"60"}, "tab": {"sentiment"}})
	if err != nil {
		return backupHistoryDay{}
	}
	_, section, ok := strings.Cut(string(body), `<section aria-label="市场整体情绪"`)
	if !ok {
		return backupHistoryDay{}
	}
	section, _, _ = strings.Cut(section, "</section>")
	visible := strings.Join(strings.Fields(historyHTMLTags.ReplaceAllString(section, " ")), " ")
	score := zhishiScore.FindStringSubmatch(visible)
	counts := zhishiBreadth.FindStringSubmatch(strings.ReplaceAll(visible, ",", ""))
	if score == nil || score[1] != date || counts == nil {
		return backupHistoryDay{}
	}
	n, e1 := strconv.ParseFloat(score[2], 64)
	up, e2 := strconv.Atoi(counts[1])
	down, e3 := strconv.Atoi(counts[2])
	amount, e4 := strconv.ParseFloat(counts[3], 64)
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil || math.IsInf(amount, 0) {
		return backupHistoryDay{}
	}
	return backupHistoryDay{date, up, down, n, amount, zhishiName}
}

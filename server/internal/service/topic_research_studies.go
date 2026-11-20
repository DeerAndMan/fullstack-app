package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"fullstack-app/server/internal/model"
	"gorm.io/gorm"
)

func studyKind(q string) string {
	for _, p := range []string{"回踩", "小平台", "大平台", "上涨中继", "形态", "触发", "候选", "大阳"} {
		if strings.Contains(q, p) {
			return "patterns"
		}
	}
	for _, p := range []string{"板块", "行业容量", "行业强度", "百日新高", "行业排名"} {
		if strings.Contains(q, p) {
			return "sectors"
		}
	}
	for _, p := range []string{"风格", "平均股价", "近期新高指数", "880865", "880003"} {
		if strings.Contains(q, p) {
			return "style"
		}
	}
	return ""
}

var reviewSettings = regexp.MustCompile(`(大阳|贴线|平台宽度|十字实体|振幅比)(?:实体涨幅|或支撑距离|占振幅)?(?:上限|下限|容差)?(?:[（(]%[）)])?[\s为:：=<>≥≤到调至设]*(\-?\d+(?:\.\d+)?)`)

func (s *TopicResearchService) studyReply(ctx context.Context, day model.MarketReviewDay, q string, r *ResearchReport) {
	kind := studyKind(q)
	if kind == "" {
		return
	}
	payload, err := s.Reviews.Study(ctx, day.ID, kind)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		r.Conclusion += " 该日此项尚未归档。"
		return
	}
	if err != nil {
		r.Conclusion += " 读取该项归档失败。"
		return
	}
	switch kind {
	case "style":
		var data struct {
			Spread float64
			Rows   []struct {
				Code                       string
				Close, Change, EMA5, EMA13 float64
			}
		}
		if json.Unmarshal(payload, &data) != nil {
			return
		}
		for _, a := range data.Rows {
			r.Conclusion += fmt.Sprintf(" 指数 %s 收盘 %.2f，涨跌 %.2f%%，EMA5 %.2f，EMA13 %.2f。", a.Code, a.Close, a.Change, a.EMA5, a.EMA13)
		}
		r.Conclusion += fmt.Sprintf("近期新高相对平均股价的涨跌幅差为 %.2f 个百分点。", data.Spread)
	case "sectors":
		var data struct {
			Rows []struct {
				Name, Code        string
				Hits              []string
				Strength          *float64
				Eligible, Members int
			}
		}
		if json.Unmarshal(payload, &data) != nil {
			return
		}
		if strings.Contains(q, "强度") {
			sort.SliceStable(data.Rows, func(i, j int) bool {
				a, b := data.Rows[i].Strength, data.Rows[j].Strength
				return a != nil && (b == nil || *a > *b)
			})
		}
		named := false
		for _, v := range data.Rows {
			if strings.Contains(q, v.Name) || strings.Contains(q, v.Code) {
				named = true
				break
			}
		}
		n := 0
		for _, v := range data.Rows {
			if named && !strings.Contains(q, v.Name) && !strings.Contains(q, v.Code) {
				continue
			}
			strength := "缺少指数收盘"
			if v.Strength != nil {
				strength = fmt.Sprintf("%.2f%%", *v.Strength)
			}
			r.Conclusion += fmt.Sprintf(" %s，百日新高 %d 家，合格 %d/%d，20 日强度 %s。", v.Name, len(v.Hits), v.Eligible, v.Members, strength)
			if named && len(v.Hits) > 0 {
				r.Conclusion += "命中 " + strings.Join(v.Hits, "、") + "。"
			}
			n++
			if n == 5 {
				break
			}
		}
	case "patterns":
		t := DefaultReviewThresholds()
		for _, m := range reviewSettings.FindAllStringSubmatch(q, -1) {
			v, _ := strconv.ParseFloat(m[2], 64)
			switch m[1] {
			case "大阳":
				t.Big = v
			case "贴线":
				t.Near = v
			case "平台宽度":
				t.Width = v
			case "十字实体":
				t.Doji = v
			case "振幅比":
				t.Shrink = v
			}
		}
		stage := "signal"
		if strings.Contains(q, "触发") {
			stage = "trigger"
		}
		pattern := ""
		for _, name := range []string{"上涨中继", "EMA5 小平台", "EMA5 大平台", "EMA13 回踩", "超跌回踩"} {
			if strings.Contains(strings.ReplaceAll(strings.ToLower(q), " ", ""), strings.ReplaceAll(strings.ToLower(name), " ", "")) {
				pattern = name
				break
			}
		}
		code := ""
		if v := reviewStockCode.FindStringSubmatch(q); v != nil {
			code = v[1]
		}
		rows, e := FilterReviewPatterns(payload, t, pattern, code, stage)
		if e != nil {
			r.Conclusion += " 筛选阈值超出可用范围。"
			return
		}
		r.Conclusion += fmt.Sprintf(" 大阳下限 %.2f%%，贴线容差 %.2f%%，平台宽度 %.2f%%，十字实体上限 %.2f%%，振幅比上限 %.2f，符合 %d 个形态信号。", t.Big, t.Near, t.Width, t.Doji, t.Shrink, len(rows))
		params := url.Values{"date": {day.Date}, "stage": {stage}, "pattern": {pattern}, "search": {code}}
		for key, value := range map[string]float64{"big": t.Big, "near": t.Near, "width": t.Width, "doji": t.Doji, "shrink": t.Shrink} {
			params.Set(key, strconv.FormatFloat(value, 'f', -1, 64))
		}
		r.Sources[0].URL = "/hotspots/review?" + params.Encode()
		r.Sources[0].Title = "查看同阈值的完整形态名单"
		if len(rows) > 10 {
			r.Conclusion += " 以下列出前 10 个，来源链接可打开完整名单。"
		}
		for _, v := range rows[:min(10, len(rows))] {
			r.Conclusion += fmt.Sprintf(" %s %s，%s，大阳日 %s，候选日 %s。", v.Symbol, v.Name, v.Kind, v.Anchor, v.SignalDate)
			if v.TriggerDate != "" {
				r.Conclusion += "触发日 " + v.TriggerDate + "。"
			}
		}
	}
}

package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"fullstack-app/server/internal/model"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

var reviewDatePattern = regexp.MustCompile(`(20[0-9]{2})[-年/.]([0-9]{1,2})[-月/.]([0-9]{1,2})日?`)
var reviewCompactDate = regexp.MustCompile(`\b(20[0-9]{2})([01][0-9])([0-3][0-9])\b`)
var reviewMonthDay = regexp.MustCompile(`([0-9]{1,2})月([0-9]{1,2})日?`)
var reviewStockCode = regexp.MustCompile(`(?:^|[^0-9])((?:00|30|60|68)[0-9]{4})(?:$|[^0-9])`)

func asksReviewDates(question string) bool {
	for _, phrase := range []string{"哪些日期可查", "哪些交易日可查", "哪些交易日可以复盘", "哪些交易日已归档", "已归档的交易日有哪些"} {
		if strings.Contains(question, phrase) {
			return true
		}
	}
	return false
}

func useMarketReview(question string, skills []ResearchSkill) bool {
	if !slices.ContainsFunc(skills, func(s ResearchSkill) bool { return s.seedKey == skillMarketReview }) {
		return false
	}
	if studyKind(question) != "" || asksReviewDates(question) || reviewDatePattern.MatchString(question) || reviewCompactDate.MatchString(question) || reviewMonthDay.MatchString(question) || reviewStockCode.MatchString(question) {
		return true
	}
	if len(skills) == 1 && (strings.Contains(question, "建议") || strings.Contains(question, "买入")) {
		return true
	}
	for _, term := range []string{"复盘", "上涨家数", "新高", "沪深300", "沪深 300", "均线", "ma50", "50日", "50 日", "名单", "分母", "占比", "背离", "交易日"} {
		if strings.Contains(strings.ToLower(question), term) {
			return true
		}
	}
	return false
}

func (s *TopicResearchService) reviewEvidence(ctx context.Context, question string, reference *ResearchRunView) ResearchReport {
	r := ResearchReport{Topics: []string{"强势股复盘"}, Match: "all", Candidates: []ResearchCandidate{}, Sources: []ResearchSource{}, Notes: []string{}}
	if asksReviewDates(question) {
		dates, err := s.Reviews.Dates(ctx)
		if err != nil {
			r.Conclusion = "读取复盘归档失败。"
			return r
		}
		if len(dates) == 0 {
			r.Conclusion = "当前还没有已发布的复盘交易日。"
			return r
		}
		r.Conclusion = fmt.Sprintf("归档共 %d 个交易日，覆盖 %s 至 %s。", len(dates), dates[len(dates)-1], dates[0])
		n := min(31, len(dates))
		r.Conclusion += fmt.Sprintf("最近 %d 个已归档交易日为 %s。", n, strings.Join(dates[:n], "、"))
		r.Sources = []ResearchSource{{ID: "R1", Provider: "复盘归档", Title: "已归档交易日", URL: "/hotspots/review", Kind: "original", Date: dates[0]}}
		return r
	}
	date := ""
	if reference != nil && reference.Result != nil && !strings.Contains(question, "最新") {
		for _, src := range reference.Result.Sources {
			if src.Provider == "复盘归档" {
				date = src.Date
				break
			}
		}
	}
	now := time.Now().In(marketZone)
	match := reviewDatePattern.FindStringSubmatch(question)
	if match == nil {
		match = reviewCompactDate.FindStringSubmatch(question)
	}
	if match != nil {
		month, _ := strconv.Atoi(match[2])
		day, _ := strconv.Atoi(match[3])
		date = fmt.Sprintf("%s-%02d-%02d", match[1], month, day)
	} else if md := reviewMonthDay.FindStringSubmatch(question); md != nil {
		year := now.Year()
		if t, e := time.Parse(time.DateOnly, date); e == nil {
			year = t.Year()
		}
		month, _ := strconv.Atoi(md[1])
		day, _ := strconv.Atoi(md[2])
		date = fmt.Sprintf("%04d-%02d-%02d", year, month, day)
	} else if strings.Contains(question, "今天") || strings.Contains(question, "今日") || strings.Contains(question, "昨天") {
		if strings.Contains(question, "昨天") {
			now = now.AddDate(0, 0, -1)
		}
		date = now.Format(time.DateOnly)
	} else if strings.Contains(question, "上一交易日") || strings.Contains(question, "前一交易日") {
		anchor, e := s.Reviews.Day(ctx, date, "")
		if e != nil {
			r.Conclusion = "缺少基准交易日归档。"
			return r
		}
		date = anchor.Previous
	}
	day, err := s.Reviews.Day(ctx, date, "")
	if err != nil {
		r.Conclusion = "读取复盘归档失败。"
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.Conclusion = date + " 暂无已发布复盘。可在强势股复盘页选择已归档交易日。"
		}
		return r
	}
	ratio := "无有效样本"
	if day.Eligible > 0 {
		ratio = fmt.Sprintf("%.2f%%", 100*float64(day.Above)/float64(day.Eligible))
	}
	r.Conclusion = fmt.Sprintf("%s 沪深 A 股上涨 %d 家，250 日新高 %d 家（合格 %d 家），沪深 300 的 50 日占比为 %s（%d/%d，成份 %d 家）。", day.Date, day.Up, day.Highs, day.HighEligible, ratio, day.Above, day.Eligible, day.CSI)
	r.Sources = []ResearchSource{{ID: "R1", Provider: "复盘归档", Title: "收盘统计与逐股明细", URL: "/hotspots/review?date=" + day.Date, Kind: "original", Date: day.Date}}
	r.Notes = append(r.Notes, fmt.Sprintf("沪深当日股票池 %d 家，有当日成交日线 %d 家。新高窗口不足 %d 家，沪深 300 的 MA50 排除 %d 家。", day.Total, day.Traded, day.Traded-day.HighEligible, day.CSI-day.Eligible))
	r.Notes = append(r.Notes, "按日涨跌幅计上涨；复权最高价严格突破此前 249 根有效日线；收盘严格高于含当日的 MA50。规则 "+day.Rule+"，批次 "+day.Batch)
	if prev, e := s.Reviews.Day(ctx, day.Previous, ""); e == nil && prev.Rule == day.Rule {
		delta := day.Highs - prev.Highs
		r.Notes = append(r.Notes, fmt.Sprintf("较 %s，一年新高家数变化 %+d。", prev.Date, delta))
	}
	if strings.Contains(question, "建议") || strings.Contains(question, "买入") {
		raw, e := s.Reviews.Study(ctx, day.ID, model.ReviewAdviceKind)
		var advice model.ReviewAdvice
		if e == nil && json.Unmarshal(raw, &advice) == nil {
			r.Conclusion += " 已归档 AI 研判：" + advice.Market + " 买入条件：" + advice.Buy
		} else if errors.Is(e, gorm.ErrRecordNotFound) {
			r.Notes = append(r.Notes, "当日尚无已归档 AI 建议。")
		} else {
			r.Notes = append(r.Notes, "读取已归档 AI 建议失败。")
		}
	}
	if studyKind(question) != "" {
		s.studyReply(ctx, day, question, &r)
		return r
	}
	metric := "high"
	if strings.Contains(question, "新高") {
		metric = "high"
	} else if strings.Contains(question, "上涨") {
		metric = "up"
	} else if strings.Contains(question, "300") {
		metric = "csi"
	}
	search := ""
	if code := reviewStockCode.FindStringSubmatch(question); code != nil {
		search = code[1]
	}
	if search != "" {
		metric = "all"
	}
	if search != "" || strings.Contains(question, "哪些") || strings.Contains(question, "名单") {
		rows, n, e := s.Reviews.Members(ctx, day.ID, metric, search, 1)
		if e == nil {
			codes := []string{}
			for _, m := range rows {
				label := m.Symbol + " " + m.Name
				if search != "" {
					if m.Change != nil {
						label += fmt.Sprintf("，涨跌幅 %.2f%%", *m.Change)
					}
					if m.NewHigh != nil {
						label += fmt.Sprintf("，当日最高 %.2f，前 249 根最高 %.4f", *m.High, *m.PreviousHigh)
					}
					if m.MA50 != nil {
						label += fmt.Sprintf("，收盘 %.2f，MA50 %.4f", *m.Close, *m.MA50)
					}
					if m.Reason != "" {
						label += "，" + m.Reason
					}
				}
				codes = append(codes, label)
			}
			r.Notes = append(r.Notes, fmt.Sprintf("符合条件 %d 家，按代码列出前 %d 家：%s", n, len(codes), strings.Join(codes, "、")))
			r.Conclusion += " " + r.Notes[len(r.Notes)-1]
		}
	}
	return r
}

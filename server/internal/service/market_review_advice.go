package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"fullstack-app/server/internal/model"
	"fullstack-app/server/pkg/errcode"
	"gorm.io/gorm"
)

const reviewAdvicePrompt = `你是A股收盘复盘研究员，只能依据输入交易日及此前的归档，写条件性复盘建议。禁止引用事后表现、虚构个股、价格、成交额或北交所统计。
原文方法：一年新高不足20家通常暂停强势股常规开仓，从20及以下重返20以上恢复找机会。上涨少于1000家且新高增加，可在连续普跌背景下观察逆势品种。沪深300收盘在MA50上方占比低于20%或高于70%是ETF冷热参考，不能直接推出个股买卖。
用行业百日新高家数衡量容量，二十日涨幅衡量强度，观察方向是否持续。比较近期新高880865与平均股价880003和EMA5、EMA13。强股分歧时看大阳后的十字星、EMA5平台、首次EMA13回踩或旧支撑企稳、板块呼应与失效条件。缺个股确认时建议观望或候选评估。候选信号数不是去重股票数。
day.highs 和 history.highs 都是按250根有效日线统计的全市场一年新高家数。只用「一年新高」描述这两个字段，用「百日新高」描述 industries 中的行业容量。
rules 是后端已经计算的真值。严禁重新推导或改变阈值口径。若回到20以上为false，不要提“重返、重回、恢复、重新、脱离”。ETF占比位于中间区间时，不允许称冷、热、冰、沸。只引用明确给出的真实数据。
返回JSON字符串字段 market（结合当日真实数值说明总体建议，80字内），buy（说明常规新开仓、观察或条件性买入评估，以及需要的确认和失效条件，100字内），averageState（平均股价相对归档均线或当日涨跌的短语，12字内），styleBias（近期新高相对平均股价的强弱，12字内，不能推断题材/权重），sectorState（仅归纳有证据的行业表现，16字内）。历史的leaders为容量前三，同数按代码排序。缺证据的标签返回空字符串。不要描述为原作者人工标注。`

// Ground the two index comparisons in numerical evidence before the model interprets them.
func reviewStyleLabels(raw json.RawMessage) (string, string, error) {
	if len(raw) == 0 {
		return "", "", nil
	}
	var data model.ReviewStyleData
	if err := json.Unmarshal(raw, &data); err != nil {
		return "", "", err
	}
	average, bias := data.Labels()
	return average, bias, nil
}

func (s *TopicResearchService) GenerateReviewAdvice(ctx context.Context, date, batch string) (*model.ReviewAdvice, error) {
	day, err := s.Reviews.Day(ctx, date, batch)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if _, busy := s.active.LoadOrStore("review-advice:"+batch, true); busy {
		return nil, errResearchBusy
	}
	defer s.active.Delete("review-advice:" + batch)
	var result model.ReviewAdvice
	raw, err := s.Reviews.Study(ctx, day.ID, model.ReviewAdviceKind)
	if err == nil {
		err = json.Unmarshal(raw, &result)
		return &result, err
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	key := researchAPIKey(s.localKey)
	if key == "" {
		return nil, errResearchKey
	}
	select {
	case s.slots <- struct{}{}:
		defer func() { <-s.slots }()
	default:
		return nil, errResearchBusy
	}
	history, err := s.Reviews.Timeline(ctx, date)
	if err != nil {
		return nil, err
	}
	if len(history) == 0 || history[len(history)-1].Batch != batch {
		return nil, errcode.ErrBadRequest
	}
	for i := range history {
		history[i].Advice = nil
	}
	input := map[string]any{"day": day, "history": history}
	today := history[len(history)-1]
	if today.AdvicePending {
		return nil, errcode.New(http.StatusConflict, "当日行业和形态尚在归档。", http.StatusConflict)
	}
	recovered := today.HighDelta != nil && day.Highs > 20 && day.Highs-*today.HighDelta <= 20
	temperature := "无有效成份"
	if day.Eligible > 0 {
		switch {
		case day.Above*100 < 20*day.Eligible:
			temperature = "偏冷"
		case day.Above*100 > 70*day.Eligible:
			temperature = "偏热"
		default:
			temperature = "中间区间"
		}
	}
	input["rules"] = map[string]any{"ETF占比区间": temperature, "新高重返20以上": recovered, "新高不足20家": day.Highs < 20, "上涨不足1000家": day.Up < 1000}
	style, err := s.Reviews.Study(ctx, day.ID, "style")
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err == nil {
		input["style"] = style
	}
	average, bias, err := reviewStyleLabels(style)
	if err != nil {
		return nil, err
	}
	input["已核对风格"] = map[string]string{"平均股价状态": average, "当日相对涨跌": bias}
	sectors, err := s.Reviews.Study(ctx, day.ID, "sectors")
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	var board struct {
		Rows []struct {
			Name     string
			Hits     []string
			Strength *float64
		}
	}
	if err == nil {
		if err = json.Unmarshal(sectors, &board); err != nil {
			return nil, err
		}
		rows := make([]map[string]any, 0, len(board.Rows))
		for _, b := range board.Rows {
			rows = append(rows, map[string]any{"name": b.Name, "capacity": len(b.Hits), "strengthPct": b.Strength})
		}
		input["industries"] = rows
	}
	patterns, err := s.Reviews.Study(ctx, day.ID, "patterns")
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err == nil {
		for _, stage := range []string{"signal", "trigger"} {
			list, e := FilterReviewPatterns(patterns, DefaultReviewThresholds(), "", "", stage)
			if e != nil {
				return nil, e
			}
			input[stage+"Count"] = len(list)
		}
	}
	for attempt := 0; attempt < 2; attempt++ {
		result, err = researchChat[model.ReviewAdvice](ctx, s.client, key, reviewAdvicePrompt, input, 900)
		if err != nil {
			return nil, err
		}
		copyText := result.Market + result.Buy
		invalidTemperature := temperature == "中间区间" && strings.ContainsAny(copyText, "冷热冰沸")
		invalidRecovery := !recovered && (strings.Contains(result.Market, "重返") || strings.Contains(result.Market, "重回"))
		if !invalidTemperature && !invalidRecovery {
			break
		}
		if attempt == 1 {
			return nil, errors.New("AI advice conflicts with archived market rules")
		}
		input["修订要求"] = "只分析新高家数与行业方向，不评论ETF温度，不使用重返、重回。"
	}
	for _, field := range []*string{&result.Market, &result.Buy, &result.AverageState, &result.StyleBias, &result.SectorState} {
		*field = strings.TrimSpace(*field)
		if utf8.RuneCountInString(*field) > 500 {
			return nil, errors.New("AI response exceeds review field limit")
		}
	}
	if result.Market == "" || result.Buy == "" {
		return nil, errors.New("AI review is missing advice")
	}
	result.AverageState = average
	result.StyleBias = bias
	if len(board.Rows) == 0 {
		result.SectorState = ""
	}
	result.Model = researchModel
	result.Rule = model.ReviewAdviceKind
	result.GeneratedAt = time.Now().UTC().Format(time.RFC3339)
	raw, err = json.Marshal(result)
	if err != nil {
		return nil, err
	}
	if err = s.Reviews.SaveStudy(ctx, date, batch, model.ReviewAdviceKind, raw); err != nil {
		return nil, err
	}
	raw, err = s.Reviews.Study(ctx, day.ID, model.ReviewAdviceKind)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(raw, &result)
	return &result, err
}

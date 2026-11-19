package service

import (
	"context"
	"net/url"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

type MarketReasonSource struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type MarketConceptReason struct {
	Status  string               `json:"status"`
	Kind    string               `json:"kind,omitempty"`
	Text    string               `json:"text,omitempty"`
	Sources []MarketReasonSource `json:"sources,omitempty"`
}

const eastmoneyConceptProfile = "https://emweb.eastmoney.com/PC_HSF10/CoreConception/"

func (s *MarketService) ConceptReason(ctx context.Context, concept, symbol string) (MarketConceptReason, error) {
	if !conceptCode.MatchString(concept) || !marketStockCode.MatchString(symbol) {
		return MarketConceptReason{}, ErrMarketQuery
	}
	if reason, ok := conceptReasonSources[concept][symbol]; ok {
		reason.Status = "ready"
		return reason, nil
	}

	catalog, err := s.Feed(ctx, "concept", "", "")
	if err != nil {
		return MarketConceptReason{}, err
	}
	if catalog.Status == "unavailable" {
		return MarketConceptReason{Status: "unavailable"}, nil
	}
	var topic string
	for _, row := range catalog.Rows {
		if row.Code == concept {
			topic = row.Name
			break
		}
	}
	if topic == "" {
		return MarketConceptReason{}, ErrMarketQuery
	}

	profile, err := s.cached(ctx, "stock-concept-profile:"+symbol, 12*time.Hour, s.stockConceptProfile, symbol)
	if err != nil {
		return MarketConceptReason{}, err
	}
	if profile.Status == "unavailable" {
		return MarketConceptReason{Status: "unavailable"}, nil
	}
	if excerpt := conceptProfileExcerpt(concept, topic, profile.Rows); excerpt != "" {
		return MarketConceptReason{
			Status: "ready", Kind: "资料摘录", Text: excerpt,
			Sources: []MarketReasonSource{{Title: "东方财富 · 核心题材", URL: profile.SourceURL}},
		}, nil
	}
	return MarketConceptReason{Status: "missing"}, nil
}

func (s *MarketService) stockConceptProfile(ctx context.Context, symbol string) (MarketDataset, error) {
	type point struct {
		Code           string `json:"SECURITY_CODE"`
		Keyword        string `json:"KEYWORD"`
		Content        string `json:"MAINPOINT_CONTENT"`
		Classification string `json:"KEY_CLASSIF"`
	}
	params := url.Values{"code": {strings.ToUpper(symbol)}}
	data, err := marketJSON[struct {
		Points []point `json:"hxtc"`
	}](ctx, s, eastmoneyConceptProfile+"PageAjax", params)
	if err != nil {
		return MarketDataset{}, err
	}
	if data.Points == nil {
		return MarketDataset{}, errMarketData
	}
	params.Set("type", "web")
	out := MarketDataset{Source: "东方财富", SourceURL: eastmoneyConceptProfile + "Index?" + params.Encode()}
	var other, scope []MarketRow
	for _, point := range data.Points {
		if point.Code != symbol[2:] || strings.TrimSpace(point.Content) == "" || point.Classification == "行业背景" {
			continue
		}
		row := MarketRow{Name: point.Keyword, Detail: strings.Join(strings.Fields(point.Content), " ")}
		switch point.Classification {
		case "主营业务":
			out.Rows = append(out.Rows, row)
		case "经营范围":
			scope = append(scope, row)
		default:
			other = append(other, row)
		}
	}
	out.Rows = append(append(out.Rows, other...), scope...)
	return out, nil
}

func conceptProfileExcerpt(concept, topic string, rows []MarketRow) string {
	topic = normalizeResearchTopic(topic)
	if utf8.RuneCountInString(topic) < 2 {
		return ""
	}
	for _, row := range rows {
		if normalizeResearchTopic(row.Name) == topic && utf8.RuneCountInString(row.Detail) <= 220 {
			return row.Detail
		}
	}
	terms := []string{topic}
	if concept == "gn_hwhs" {
		terms = append(terms, "海思半导体", "海思光电子")
	}
	for _, row := range rows {
		// Include the full disclosure paragraph to retain qualifiers and negations.
		if utf8.RuneCountInString(row.Detail) > 220 {
			continue
		}
		for _, term := range terms {
			if strings.ContainsFunc(term, func(r rune) bool { return unicode.Is(unicode.Han, r) }) &&
				strings.Contains(strings.ToLower(row.Detail), term) {
				return row.Detail
			}
		}
	}
	return ""
}

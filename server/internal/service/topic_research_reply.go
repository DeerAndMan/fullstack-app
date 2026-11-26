package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"
)

const noSupplyAnnouncement = "本轮检索到的原文里未找到明确确认供货的公告。"
const researchScopeRule = "涉及供货关系时，按本轮原文支持的范围表述。未查到确认供货的公告时，写“" + noSupplyAnnouncement + "” 不得据此断言没有公司供货。公司明确否认时注明公司、项目和来源。"

func needsSupplyQualification(question string, report ResearchReport) bool {
	if !strings.Contains(question, "公告") || !strings.Contains(question, "供货") && !strings.Contains(question, "供应") && !strings.Contains(question, "交付") {
		return false
	}
	if report.Optimus == nil {
		return true // General reports have no verified announcement text.
	}
	announcements := map[string]bool{}
	for _, source := range report.Sources {
		link, err := url.Parse(source.URL)
		if err != nil || !slices.Contains(optimusPrimary, source.Provider) {
			continue
		}
		if strings.HasSuffix(strings.ToLower(link.Path), ".pdf") || strings.Contains(link.Path, "/disclosure/") {
			announcements[source.ID] = true
		}
	}
	for _, stock := range report.Optimus.Stocks {
		for _, ev := range stock.Evidence {
			if ev.FullText && ev.Status == "confirmed" && announcements[ev.SourceID] {
				return false
			}
		}
	}
	return true
}

var blanketSupplyClaim = regexp.MustCompile(`(?:没有|并无|暂无|并没有|无任何|无一家|未有)(?:任何|一家|一家公司|任何一家|上市|相关|上述|一家上市|任何上市|A股)*(?:公司|企业|厂商|供应商)|(?:所有|全部)(?:公司|企业|厂商).{0,12}(?:均未|都未|没有)|(?:公司|企业|厂商)(?:均未|都未|均无)`)

// Reply streams the conversational answer after the structured evidence has been checked.
func (s *TopicResearchService) streamReply(ctx context.Context, id uint, key, question string, skills []ResearchSkill, reference *ResearchRunView, report ResearchReport) string {
	minesweeper := usesMinesweeper(skills)
	client := s.client
	timeout := 25 * time.Second
	maxTokens := 800
	if minesweeper {
		// A complete checklist takes longer to stream than the ordinary research summary.
		timeout = 75 * time.Second
		maxTokens = 4200
		client = &http.Client{Transport: s.client.Transport, CheckRedirect: s.client.CheckRedirect, Timeout: timeout}
	} else {
		if text := combinedSkillReply(question, report.SkillChecks); text != "" {
			s.replies.Store(id, text)
			return text
		}
	}
	scope, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	materials := map[string]any{"question": question, "skills": skills, "topics": report.Topics, "conclusion": report.Conclusion, "notes": report.Notes, "observations": report.Observations, "candidates": report.Candidates[:min(50, len(report.Candidates))]}
	if len(report.SkillChecks) > 0 {
		materials["originalEvidenceChecks"] = report.SkillChecks
		materials["originalSources"] = slices.DeleteFunc(slices.Clone(report.Sources), func(source ResearchSource) bool {
			return source.Kind != "original"
		})
	}
	if report.Optimus != nil {
		materials["stocks"] = report.Optimus.Stocks
	}
	if reference != nil {
		materials["previousTurns"] = reference.History
	}
	lead := ""
	instruction := researchScopeRule
	if len(report.SkillChecks) > 0 {
		instruction = " 优先报告 originalEvidenceChecks 的已核对结论，在相关事实后标注相应的原文编号（如 E1）。空 findings 表示该项未取得可支持结论的原文。沿用结论中的公司、期间和数字，勿补充引句未支持的其他数字或客户。搜索摘要仅作线索；不能以公司个例替代行业统计，或将工业机器人指标当作人形机器人指标。成分股缺席仅限本轮取数范围。"
	}
	hasOrderEvidence := slices.ContainsFunc(report.SkillChecks, func(check ResearchSkillCheck) bool {
		return check.Method == "orders" && len(check.Findings) > 0
	})
	if (report.Optimus != nil || !hasOrderEvidence) && needsSupplyQualification(question, report) {
		lead = noSupplyAnnouncement
		s.replies.Store(id, lead)
		instruction += " 系统已在回复开头显示上述范围说明。直接补充当前资料中的具体公司、证据和出处，不重复总括结论。"
	}
	if minesweeper {
		passages := map[string]string{}
		for _, source := range report.Sources {
			if source.content != "" {
				passages[source.ID] = skillSourcePassages(source.content, []string{"审计意见", "主要会计数据", "经营活动产生", "投资活动产生", "应收账款", "存货", "运输费", "在建工程", "商誉", "前五名", "非经营性占用", "研发投入", "财务总监", "独立董事", "立案"})
			}
		}
		materials["disclosurePassages"] = passages
		materials["sources"] = report.Sources
		instruction += " 按所选财报排雷指令给出逐项判定和评分。disclosurePassages 为以来源编号标记的披露节选，核对公司、年份、单位及上下文后再引用。缺少全文或数据时列出对应的 SKIP，不得把节选未提及的问题判为 PASS。"
	}
	input, _ := json.Marshal(materials)
	body, _ := json.Marshal(map[string]any{"model": researchModel, "stream": true, "thinking": map[string]string{"type": "disabled"}, "max_tokens": maxTokens, "messages": []map[string]string{
		{"role": "system", "content": "你是中文题材研究助手。直接回答用户本轮问题，结合历轮对话和本次已核验的资料。区分披露与传闻，仅使用所给材料，不编造业务或链接。完整表格与原文由界面展示，回答用简洁纯文本。" + instruction},
		{"role": "user", "content": string(input)},
	}})
	req, err := http.NewRequestWithContext(scope, http.MethodPost, researchEndpoint+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return lead
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)
	res, err := client.Do(req)
	if err != nil {
		return lead
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return lead
	}
	scan := bufio.NewScanner(res.Body)
	var reply strings.Builder
	reply.WriteString(lead)
	pending := ""
	publish := func(sentence string) {
		if lead != "" {
			if strings.TrimSpace(sentence) == "" || blanketSupplyClaim.MatchString(sentence) || strings.Contains(sentence, strings.TrimSuffix(lead, "。")) {
				return
			}
			if reply.Len() == len(lead) {
				reply.WriteString("\n\n")
			}
		}
		reply.WriteString(sentence)
		s.replies.Store(id, reply.String())
	}
	finish := func() string { publish(pending); return strings.TrimSpace(reply.String()) }
	for scan.Scan() {
		data, ok := strings.CutPrefix(scan.Text(), "data:")
		if !ok {
			continue
		}
		data = strings.TrimSpace(data)
		if data == "[DONE]" {
			return finish()
		}
		var chunk struct {
			Choices []struct {
				Delta  struct{ Content string }
				Finish string `json:"finish_reason"`
			}
		}
		if json.Unmarshal([]byte(data), &chunk) != nil || len(chunk.Choices) == 0 {
			continue
		}
		choice := chunk.Choices[0]
		if choice.Delta.Content != "" {
			if lead == "" {
				publish(choice.Delta.Content)
			} else {
				// Check whole sentences before publishing, including negations split across model tokens.
				pending += choice.Delta.Content
				for {
					cut := strings.IndexAny(pending, "。！？\n")
					if cut < 0 {
						break
					}
					_, width := utf8.DecodeRuneInString(pending[cut:])
					end := cut + width
					publish(pending[:end])
					pending = pending[end:]
				}
			}
		}
		if choice.Finish == "stop" {
			return finish()
		}
		if minesweeper && choice.Finish == "length" {
			publish("\n本轮清单未输出完整，请复核剩余项目和评分。")
			return finish()
		}
	}
	if lead != "" {
		return strings.TrimSpace(reply.String())
	}
	return ""
}

func combinedSkillReply(question string, checks []ResearchSkillCheck) string {
	// When the question explicitly asks for multiple verified methods, use the
	// cited findings directly so a second completion cannot mix their sources.
	methods := []struct{ id, name string }{
		{"orders", "订单兑现"}, {"industry", "行业景气"}, {"risks", "风险反证"},
	}
	var sections []string
	hasFindings := false
	for _, method := range methods {
		if !strings.Contains(question, method.name) {
			continue
		}
		for _, check := range checks {
			if check.Method != method.id {
				continue
			}
			lines := []string{method.name}
			if len(check.Findings) == 0 {
				lines = append(lines, "本轮未取得足以支持结论的原文。")
			}
			for _, finding := range check.Findings {
				hasFindings = true
				var sourceIDs []string
				for _, evidence := range finding.Evidence {
					if !slices.Contains(sourceIDs, evidence.SourceID) {
						sourceIDs = append(sourceIDs, evidence.SourceID)
					}
				}
				line := finding.Note + "（" + strings.Join(sourceIDs, "、") + "）"
				if finding.Subject != "" && !strings.HasPrefix(line, finding.Subject) {
					line = finding.Subject + " · " + line
				}
				lines = append(lines, line)
			}
			sections = append(sections, strings.Join(lines, "\n"))
		}
	}
	if len(sections) < 2 || !hasFindings {
		return ""
	}
	return strings.Join(sections, "\n\n")
}

func (s *TopicResearchService) Draft(id uint) string {
	if value, ok := s.replies.Load(id); ok {
		return value.(string)
	}
	return ""
}

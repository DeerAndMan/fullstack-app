package service

import "testing"

func TestDisclosureIssuerUsesHeaderAndRejectsDifferentCompany(t *testing.T) {
	for _, tc := range []struct {
		header, requested, expected string
	}{
		{"证券代码：000887 证券简称：中鼎股份 关于安徽中鼎密封件股份有限公司的问询回复", "领益智造", ""},
		{"证券代码：002600 证券简称：领益智造 债券简称：领益转债 广东领益智造股份有限公司", "领益智造", "领益智造"},
		{"## 鸣志电器2026年半年度业绩说明会召开实况", "鸣志电器", "鸣志电器"},
		{"证券代码：603728 证券简称：鸣志电器 公告编号：2025-044", "603728", "603728"},
		{"广东领益智造股份有限公司 2026 年半年度报告全文", "领益智造", "领益智造"},
	} {
		if got := disclosureIssuer(tc.header, tc.requested); got != tc.expected {
			t.Errorf("issuer for %q = %q, want %q", tc.header, got, tc.expected)
		}
	}
}

func TestVerifySkillChecksRequiresContinuousOriginalQuoteForMethod(t *testing.T) {
	const quote = "公司已向相关整机厂商送样，并持续获得批量订单。"
	documents := []researchSkillDocument{{
		ResearchSource: ResearchSource{ID: "E1", Kind: "original"},
		Methods:        []string{"orders"},
		Text:           "答： " + quote + "\n[…段落间隔…]\n交付收入还待后续定期报告确认。",
		Issuer:         "样本公司",
	}}
	finding := func(quote string) ResearchSkillFinding {
		return ResearchSkillFinding{Subject: "样本公司", Note: "核对样本", Evidence: []ResearchSkillEvidence{{SourceID: "E1", Quote: quote}}}
	}
	result := researchSkillResult{checks: []ResearchSkillCheck{{Method: "orders"}, {Method: "industry"}}}
	verifySkillChecks(&result, []ResearchSkillCheck{
		{Method: "orders", Findings: []ResearchSkillFinding{
			finding("公司已完成整机批量订单交付并确认收入。"),
			finding(quote + "交付收入还待后续定期报告确认。"),
			finding("公司已向相关整机厂商送样， 并持续获得批量订单。"),
			{Subject: "另一家", Note: "其他公司被误归", Evidence: []ResearchSkillEvidence{{SourceID: "E1", Quote: quote}}},
			{Subject: "样本公司", Note: "把原文误归给另一家", Evidence: []ResearchSkillEvidence{{SourceID: "E1", Quote: quote}}},
		}},
		{Method: "industry", Findings: []ResearchSkillFinding{finding(quote)}},
	}, documents, []string{"样本公司", "另一家"})
	if len(result.checks[0].Findings) != 1 || len(result.checks[1].Findings) != 0 || len(result.sources) != 1 {
		t.Fatalf("unexpected supported findings: %+v", result)
	}
}

func TestIndustryQuantitiesMustAppearInCitedQuote(t *testing.T) {
	finding := ResearchSkillFinding{
		Note:     "10月工业机器人产量82,846套，同比增长26%。",
		Evidence: []ResearchSkillEvidence{{Quote: "工业机器人（套） 82846 26.0 874141 32.3"}},
	}
	if !industryQuantitiesSupported(finding) {
		t.Fatal("quoted measurements should allow comma and decimal formatting")
	}
	finding.Note += "服务机器人同比下降4.9%。"
	if industryQuantitiesSupported(finding) {
		t.Fatal("uncited measurements should be rejected")
	}
}

package service

import (
	"regexp"
	"strings"
)

var (
	disclosureCompany = regexp.MustCompile(`[\p{Han}A-Za-z0-9·（）()]{3,38}(?:股份有限公司|有限责任公司)`)
	disclosureShort   = regexp.MustCompile(`(?:证券|股票)简称[:：]([*\p{Han}A-Za-z0-9·]{2,20})`)
	disclosureCode    = regexp.MustCompile(`(?:证券|股票)代码[:：](\d{6})`)
	disclosureBrief   = regexp.MustCompile(`^#+([*\p{Han}A-Za-z·]{2,12})20\d{2}年[^。]{0,20}业绩说明会`)
	disclosureDate    = regexp.MustCompile(`/(?:t)?(20\d{2})-?(\d{2})-?(\d{2})(?:/|_)`)
)

func disclosureIssuer(header, want string) string {
	header = optimusCompact(header)
	header = truncateResearchText(header, 300)
	company := disclosureCompany.FindString(header)
	short := ""
	if match := disclosureShort.FindStringSubmatch(header); len(match) > 1 {
		short = match[1]
		for _, suffix := range []string{"债券简称", "债券代码", "公告编号", "股票代码", "证券代码", "公告", "关于"} {
			short, _, _ = strings.Cut(short, suffix)
		}
	}
	if short == "" {
		if match := disclosureBrief.FindStringSubmatch(header); len(match) > 1 {
			short = match[1]
		}
	}
	if want == "" {
		if short != "" {
			return short
		}
		return company
	}
	if short == want || strings.Contains(company, want) {
		return want
	}
	if match := disclosureCode.FindStringSubmatch(header); len(match) > 1 && match[1] == want {
		return want
	}
	return ""
}

func disclosureLabel(issuer, body string) string {
	header := truncateResearchText(body, 240)
	for _, kind := range []string{"投资者关系活动记录表", "业绩说明会", "半年度报告", "年度报告", "季度报告", "审核问询函", "募集说明书"} {
		if strings.Contains(header, kind) {
			return issuer + " · " + kind
		}
	}
	return issuer + " · 公司披露"
}

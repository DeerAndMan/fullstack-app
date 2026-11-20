package service

import (
	"context"
	"html"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const sinaLeaderboardURL = "https://vip.stock.finance.sina.com.cn/q/go.php/vInvestConsult/kind/lhb/index.phtml"

var sinaLeaderboardDetail = regexp.MustCompile(`showDetail\('\d{2}','(\d{6})','(\d{4}-\d{2}-\d{2})',this\)`)

func (s *MarketService) sinaLeaderboard(ctx context.Context, date string) (MarketDataset, error) {
	query := url.Values{}
	if date != "" {
		query.Set("tradedate", date)
	}
	raw, err := s.request(ctx, sinaLeaderboardURL, query)
	if err != nil {
		return MarketDataset{}, err
	}
	markup, err := decodeMarketText(raw)
	if err != nil || !strings.Contains(markup, "龙虎榜") {
		return MarketDataset{}, errMarketData
	}
	out := MarketDataset{Source: "新浪财经", SourceURL: sinaLeaderboardURL}
	if date != "" {
		out.SourceURL += "?" + query.Encode()
	}
	var group string
	for _, tr := range marketHTMLRow.FindAllStringSubmatch(markup, -1) {
		cells := marketHTMLCell.FindAllStringSubmatch(tr[0], -1)
		if strings.HasPrefix(tr[0], `<tr class="head">`) && len(cells) == 1 {
			group = marketCellText(cells[0][1])
			continue
		}
		if len(cells) != 8 {
			continue
		}
		detail := sinaLeaderboardDetail.FindStringSubmatch(cells[7][1])
		if len(detail) == 0 || marketCellText(cells[1][1]) != detail[1] {
			continue
		}
		if _, err := time.Parse(time.DateOnly, detail[2]); err != nil || (date != "" && detail[2] != date) {
			return MarketDataset{}, errMarketSession
		}
		date = detail[2]
		amount := marketNumber(marketCellText(cells[6][1]))
		if amount != nil {
			*amount *= 1e4
		}
		out.Rows = append(out.Rows, MarketRow{
			Code: detail[1], Symbol: stockSymbol(detail[1]), Name: marketCellText(cells[2][1]),
			Date: date, Price: marketNumber(marketCellText(cells[3][1])), Amount: amount, Detail: group,
		})
	}
	return out, nil
}

func marketCellText(markup string) string {
	return strings.TrimSpace(html.UnescapeString(flashMarkup.ReplaceAllString(markup, "")))
}

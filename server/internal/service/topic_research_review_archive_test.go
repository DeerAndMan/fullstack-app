package service

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"fullstack-app/server/internal/config"
	"fullstack-app/server/internal/database"
	"fullstack-app/server/internal/repository"
)

// This opt-in, read-only test exercises the research question path and the timeline
// used by the review page against the same published MySQL revision.
func TestPublishedMarketReview(t *testing.T) {
	day := os.Getenv("REVIEW_AUDIT_DATE")
	if day == "" {
		t.Skip("set REVIEW_AUDIT_DATE to check a published trading day")
	}
	cfg, err := config.Load("../../config/config.yaml")
	if err != nil {
		t.Fatal(err)
	}
	db, err := database.NewMySQL(&cfg.MySQL, cfg.Server.Mode)
	if err != nil {
		t.Fatal(err)
	}
	repo := repository.NewMarketReviewRepository(db)
	ctx := context.Background()
	rows, err := repo.Timeline(ctx, day)
	if err != nil {
		t.Fatal(err)
	}
	index := slices.IndexFunc(rows, func(row repository.ReviewTimelineRow) bool { return row.Date == day })
	if index < 0 {
		t.Fatalf("%s has no published timeline row", day)
	}
	row := rows[index]
	d, err := repo.Day(ctx, day, row.Batch)
	if err != nil {
		t.Fatal(err)
	}
	if row.Up != d.Up || row.Highs != d.Highs || row.Above != d.Above || row.Eligible != d.Eligible {
		t.Fatal("daily and timeline totals differ")
	}
	for metric, want := range map[string]int{"up": d.Up, "high": d.Highs, "csi": d.CSI, "all": d.Total} {
		seen := map[string]bool{}
		above, eligible := 0, 0
		for page := 1; ; page++ {
			members, total, err := repo.Members(ctx, d.ID, metric, "", page)
			if err != nil {
				t.Fatal(err)
			}
			if int(total) != want {
				t.Fatalf("%s total %d, want %d", metric, total, want)
			}
			for _, m := range members {
				if seen[m.Symbol] {
					t.Fatalf("%s duplicate: %s", metric, m.Symbol)
				}
				seen[m.Symbol] = true
				if metric == "csi" && m.Above != nil {
					eligible++
					if *m.Above {
						above++
					}
				}
			}
			if len(seen) == want {
				break
			}
			if len(members) == 0 {
				t.Fatalf("%s incomplete after %d pages", metric, page)
			}
		}
		if metric == "csi" && (above != d.Above || eligible != d.Eligible) {
			t.Fatalf("MA50 details %d/%d", above, eligible)
		}
	}
	if row.Advice == nil || row.Advice.Market == "" || row.Advice.Buy == "" || row.Advice.Model == "" || row.Advice.GeneratedAt == "" {
		t.Fatal("missing published AI advice")
	}
	svc := TopicResearchService{Reviews: repo}
	answer := svc.investigate(ctx, "", day+" 复盘建议", []ResearchSkill{{seedKey: skillMarketReview}}, nil)
	for _, s := range []string{fmt.Sprintf("上涨 %d 家", d.Up), fmt.Sprintf("250 日新高 %d 家", d.Highs), fmt.Sprintf("%d/%d", d.Above, d.Eligible), row.Advice.Market, row.Advice.Buy} {
		if !strings.Contains(answer.Conclusion, s) {
			t.Fatal("research result differs from timeline")
		}
	}
	if !strings.Contains(strings.Join(answer.Notes, " "), d.Batch) {
		t.Fatal("research batch differs from timeline")
	}
	t.Logf("date=%s batch=%s traded=%d up=%d high=%d/%d MA50=%d/%d CSI=%d model=%s generated=%s", day, d.Batch, d.Traded, d.Up, d.Highs, d.HighEligible, d.Above, d.Eligible, d.CSI, row.Advice.Model, row.Advice.GeneratedAt)
}

// Import a checked day and its evidence in one MySQL transaction.
package main

import (
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"gorm.io/gorm"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fullstack-app/server/internal/config"
	"fullstack-app/server/internal/database"
	"fullstack-app/server/internal/model"
	"fullstack-app/server/internal/repository"
	"fullstack-app/server/internal/service"
)

type artifact struct {
	Date, Previous, Rule, Release, Source, WindowStart string
	Rows                                               []model.MarketReviewMember
}

func validateMember(m model.MarketReviewMember) error {
	hasQuote := m.Bars > 0
	if m.Bars < 0 || m.Bars > 250 || hasQuote != (m.Change != nil) || hasQuote != (m.Close != nil) {
		return fmt.Errorf("%s: inconsistent quote window", m.Symbol)
	}
	if hasQuote && (m.High == nil || m.Factor == nil || *m.Factor <= 0 || *m.High < *m.Close || *m.Close <= 0) {
		return fmt.Errorf("%s: invalid prices or adjustment factor", m.Symbol)
	}
	eligibleHigh := m.Bars == 250
	if eligibleHigh != (m.NewHigh != nil) || eligibleHigh != (m.PreviousHigh != nil) {
		return fmt.Errorf("%s: missing or ineligible high comparison", m.Symbol)
	}
	eligibleMA := m.CSI && m.Bars >= 50
	if eligibleMA != (m.Above != nil) || eligibleMA != (m.MA50 != nil) {
		return fmt.Errorf("%s: missing or ineligible MA50 comparison", m.Symbol)
	}
	return nil
}

func importDay(repo *repository.MarketReviewRepository, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer z.Close()
	raw, err := io.ReadAll(z)
	if err != nil {
		return err
	}
	var a artifact
	if err = json.Unmarshal(raw, &a); err != nil {
		return err
	}
	for _, s := range []string{a.Date, a.Previous, a.WindowStart, a.Release} {
		if _, err = time.Parse(time.DateOnly, s); err != nil {
			return err
		}
	}
	if a.Rule != "shsz-a.high250.ma50.v1" || a.Previous >= a.Date || a.WindowStart > a.Previous || a.Date > a.Release || len(a.Rows) == 0 {
		return fmt.Errorf("invalid review metadata")
	}
	d := model.MarketReviewDay{Date: a.Date, Previous: a.Previous, Rule: a.Rule, Release: a.Release, Source: a.Source, WindowStart: a.WindowStart, Batch: fmt.Sprintf("%x", sha256.Sum256(raw)), Total: len(a.Rows)}
	seen := map[string]bool{}
	for _, m := range a.Rows {
		if m.Symbol == "" || seen[m.Symbol] {
			return fmt.Errorf("invalid or duplicate symbol %q", m.Symbol)
		}
		if err := validateMember(m); err != nil {
			return err
		}
		seen[m.Symbol] = true
		if m.Change != nil {
			d.Traded++
			if *m.Change > 0 {
				d.Up++
			}
		}
		if m.NewHigh != nil {
			d.HighEligible++
			if *m.NewHigh {
				d.Highs++
			}
		}
		if m.CSI {
			d.CSI++
		}
		if m.Above != nil {
			d.Eligible++
			if *m.Above {
				d.Above++
			}
		}
	}
	if d.CSI != 300 {
		return fmt.Errorf("unexpected CSI 300 size: %d", d.CSI)
	}
	if err = repo.Publish(context.Background(), &d, a.Rows); err != nil {
		return err
	}
	fmt.Printf("%s 股票池 %d 成交 %d 上涨 %d 新高 %d/%d 沪深300 %d/%d 批次 %.12s\n", d.Date, d.Total, d.Traded, d.Up, d.Highs, d.HighEligible, d.Above, d.Eligible, d.Batch)
	return nil
}

type studyIdentity struct{ Date, Batch, Kind string }

func importStudy(repo *repository.MarketReviewRepository, path string, expected *studyIdentity) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	z, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer z.Close()
	var a struct {
		Date, Batch, Kind string
		Data              json.RawMessage
	}
	if err = json.NewDecoder(z).Decode(&a); err != nil {
		return err
	}
	if expected != nil && (a.Date != expected.Date || a.Batch != expected.Batch || a.Kind != expected.Kind) {
		return fmt.Errorf("study does not match refreshed daily batch")
	}
	if len(a.Batch) != 64 {
		return fmt.Errorf("missing batch")
	}
	return repo.SaveStudy(context.Background(), a.Date, a.Batch, a.Kind, a.Data)
}

func run() error {
	defaultConfig := os.Getenv("REVIEW_IMPORT_CONFIG")
	if defaultConfig == "" {
		defaultConfig = "config/config.yaml"
	}
	path := flag.String("config", defaultConfig, "project config")
	refresh := flag.Bool("refresh", false, "atomically publish one prepared daily archive, its studies and fresh AI advice")
	core := flag.Bool("core", false, "atomically publish one verified daily archive and its style study")
	list := flag.Bool("dates", false, "list archived dates as JSON")
	readStyle := flag.Bool("read-style", false, "read the archived style payload for a date and batch")
	coverage := flag.Bool("studies", false, "list study coverage by current review batch")
	advice := flag.Bool("advice", false, "generate or reuse archived AI advice for the supplied trading dates")
	study := flag.Bool("study", false, "import gzipped review study")
	cacheRead := flag.String("read-quote-cache", "", "export rolling BaoStock windows to a temporary directory")
	cacheWrite := flag.Bool("write-quote-cache", false, "store verified rolling BaoStock windows")
	styleBars := flag.String("style-bars", "", "read dated daily bars for the two style indices")
	beijingBars := flag.String("beijing-bars", "", "read dated Beijing stock bars and corporate actions")
	flag.Parse()
	if *beijingBars != "" {
		return writeBeijingBars(*beijingBars, flag.Args())
	}
	if *styleBars != "" {
		return writeStyleBars(*styleBars)
	}
	if *core && *refresh {
		return fmt.Errorf("-core and -refresh are mutually exclusive")
	}
	if *cacheRead == "" && !*list && !*coverage && flag.NArg() == 0 {
		return fmt.Errorf("supply one or more daily .json.gz artifacts")
	}
	cfg, err := config.Load(*path)
	if err != nil {
		return err
	}
	db, err := database.NewMySQL(&cfg.MySQL, cfg.Server.Mode)
	if err != nil {
		return err
	}
	if *cacheRead != "" || *cacheWrite {
		return quoteCache(db, *cacheRead, flag.Args())
	}
	repo := repository.NewMarketReviewRepository(db)
	if *readStyle {
		if flag.NArg() != 2 {
			return fmt.Errorf("expected date and batch")
		}
		ctx := context.Background()
		day, err := repo.Day(ctx, flag.Arg(0), flag.Arg(1))
		if err != nil {
			return err
		}
		raw, err := repo.Study(ctx, day.ID, "style")
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(raw)
	}
	if *refresh || *core {
		if flag.NArg() != 1 {
			return fmt.Errorf("expected a single daily artifact")
		}
		archive := flag.Arg(0)
		date := strings.TrimSuffix(filepath.Base(archive), ".json.gz")
		if _, err = time.Parse(time.DateOnly, date); err != nil {
			return err
		}
		return db.Transaction(func(tx *gorm.DB) error {
			reviews := repository.NewMarketReviewRepository(tx)
			if err := importDay(reviews, archive); err != nil {
				return err
			}
			day, err := reviews.Day(context.Background(), date, "")
			if err != nil {
				return err
			}
			kinds := []string{"style"}
			if *refresh {
				kinds = append(kinds, "sectors", "patterns")
			}
			for _, kind := range kinds {
				file := filepath.Join(filepath.Dir(archive), date+"-"+kind+".json.gz")
				if err := importStudy(reviews, file, &studyIdentity{date, day.Batch, kind}); err != nil {
					return err
				}
			}
			if *core {
				return nil
			}
			if err := tx.Where("day_id=? AND kind=?", day.ID, model.ReviewAdviceKind).Delete(&model.MarketReviewStudy{}).Error; err != nil {
				return err
			}
			svc := service.NewTopicResearchService(nil, nil, cfg.Server.Mode == "debug")
			svc.Reviews = reviews
			_, err = svc.GenerateReviewAdvice(context.Background(), date, day.Batch)
			return err
		})
	}
	if *advice {
		svc := service.NewTopicResearchService(nil, nil, cfg.Server.Mode == "debug")
		svc.Reviews = repo
		for _, date := range flag.Args() {
			if _, err := time.Parse(time.DateOnly, date); err != nil {
				return err
			}
			day, err := repo.Day(context.Background(), date, "")
			if err != nil {
				return err
			}
			if _, err = svc.GenerateReviewAdvice(context.Background(), date, day.Batch); err != nil {
				return fmt.Errorf("%s AI review: %w", date, err)
			}
			fmt.Println(date, "AI 建议已归档")
		}
		return nil
	}
	if *list {
		if !db.Migrator().HasTable(&model.MarketReviewDay{}) {
			return json.NewEncoder(os.Stdout).Encode([]string{})
		}
		dates, err := repo.Dates(context.Background())
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(dates)
	}
	if err = db.AutoMigrate(&model.MarketReviewDay{}, &model.MarketReviewMember{}); err != nil {
		return err
	}
	if *coverage || *study {
		if err = db.AutoMigrate(&model.MarketReviewStudy{}); err != nil {
			return err
		}
		if *coverage {
			rows, err := repo.StudyCoverage(context.Background())
			if err != nil {
				return err
			}
			return json.NewEncoder(os.Stdout).Encode(rows)
		}
		for _, path := range flag.Args() {
			if err := importStudy(repo, path, nil); err != nil {
				return fmt.Errorf("%s: %w", path, err)
			}
		}
		return nil
	}
	for _, f := range flag.Args() {
		if err = importDay(repo, f); err != nil {
			return fmt.Errorf("%s: %w", f, err)
		}
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

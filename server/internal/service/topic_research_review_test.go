package service

import (
	"context"
	"strings"
	"testing"

	"fullstack-app/server/internal/repository"
	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestReviewQuestionUsesRequestedDayAndMatchingRoster(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	s := TopicResearchService{Reviews: repository.NewMarketReviewRepository(db)}
	skills := []ResearchSkill{{ID: "custom:8", seedKey: skillMarketReview}}
	question := "2026年10月30日有哪些一年新高？"
	if !useMarketReview(question, skills) || useMarketReview(question, nil) {
		t.Fatal("incorrect skill routing")
	}
	columns := []string{"id", "date", "previous", "rule", "up", "highs", "eligible", "above", "csi"}
	mock.ExpectQuery("SELECT .*market_review_days.*date = ").WithArgs("2026-10-30", 1).
		WillReturnRows(sqlmock.NewRows(columns).AddRow(9, "2026-10-30", "2026-10-29", "v1", 100, 1, 300, 75, 300))
	mock.ExpectQuery("SELECT .*market_review_days.*date = ").WithArgs("2026-10-29", 1).
		WillReturnRows(sqlmock.NewRows(columns))
	mock.ExpectQuery("SELECT count.*new_high = ").WithArgs(9, true).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery("SELECT .*market_review_members.*new_high = ").WithArgs(9, true, 50).
		WillReturnRows(sqlmock.NewRows([]string{"symbol", "name"}).AddRow("600000.SH", "样本"))
	report := s.reviewEvidence(context.Background(), question, nil)
	if !strings.Contains(report.Conclusion, "25.00%") || !strings.Contains(report.Conclusion, "600000.SH") || report.Sources[0].Date != "2026-10-30" {
		t.Fatal(report)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

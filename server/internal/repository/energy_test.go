package repository

import (
	"encoding/json"
	"errors"
	"regexp"
	"testing"

	"fullstack-app/server/internal/model"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func mockEnergyRepository(t *testing.T) (*EnergyRepository, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Error(err)
		}
		_ = sqlDB.Close()
	})
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return NewEnergyRepository(db), mock
}

func TestCreateSnapshotsStoresOnlyPositionIDsInSummary(t *testing.T) {
	t.Parallel()
	repo, mock := mockEnergyRepository(t)
	const date = "2026-11-23 15:01:00"
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `energy`")).
		WillReturnResult(sqlmock.NewResult(71, 2))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `summary`")).
		WithArgs(date, "", "", "", "", "", 0, "", "", "", "",
			`{"energyIds":[71,72]}`, 0, 0, 0, 0, "", "").
		WillReturnResult(sqlmock.NewResult(100, 1))
	mock.ExpectCommit()
	err := repo.CreateSnapshots([]AssetSnapshot{{
		Summary: model.Summary{Date: date},
		Positions: []model.Energy{
			{DATETIME: date, Zqdm: "600000"},
			{DATETIME: date, Zqdm: "000001"},
		},
	}})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateSnapshotsRollsBackWhenSummaryFails(t *testing.T) {
	t.Parallel()
	repo, mock := mockEnergyRepository(t)
	writeFailure := errors.New("summary unavailable")
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `energy`")).
		WillReturnResult(sqlmock.NewResult(71, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO `summary`")).WillReturnError(writeFailure)
	mock.ExpectRollback()
	err := repo.CreateSnapshots([]AssetSnapshot{{Positions: []model.Energy{{Zqdm: "600000"}}}})
	if !errors.Is(err, writeFailure) {
		t.Fatalf("got %v, want summary failure", err)
	}
}

func TestListSummaryReturnsLegacyArraysAndLoadsReferencedPositions(t *testing.T) {
	t.Parallel()
	repo, mock := mockEnergyRepository(t)
	const legacy = `[{"id":10,"Zqdm":"600019","legacy":true}]`
	const date = "2026-11-23 15:01:00"
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `summary` WHERE date >= ? AND date < ?")).
		WithArgs("start", "end").
		WillReturnRows(sqlmock.NewRows([]string{"id", "positions"}).
			AddRow(1, legacy).
			AddRow(2, `{"energyIds":[72,71]}`).
			AddRow(3, `{"energyIds":[71]}`))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `energy` WHERE id IN (?,?)")).
		WithArgs(72, 71).
		WillReturnRows(sqlmock.NewRows([]string{"id", "Zqdm", "DATETIME"}).
			AddRow(71, "600000", date).
			AddRow(72, "000001", date))

	summaries, err := repo.ListSummaryByDateRange("start", "end")
	if err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 3 || summaries[0].Positions != legacy {
		t.Fatalf("legacy positions changed: %+v", summaries)
	}
	var positions []model.Energy
	if err := json.Unmarshal([]byte(summaries[1].Positions), &positions); err != nil {
		t.Fatal(err)
	}
	if len(positions) != 2 || positions[0].ID != 72 || positions[1].ID != 71 ||
		positions[0].Zqdm != "000001" || positions[0].DATETIME != date {
		t.Fatalf("incorrect hydrated data or order: %+v", positions)
	}
}

package service

import (
	"context"
	"database/sql/driver"
	"regexp"
	"testing"

	"fullstack-app/server/internal/repository"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestResearchSkillsSeedNewMethodsForExistingAccount(t *testing.T) {
	t.Parallel()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	svc := NewTopicResearchService(repository.NewTopicResearchRepository(db), nil, false)

	const userID = 42
	columns := []string{"id", "user_id", "seed_key", "name", "instructions"}
	existingRows := func() *sqlmock.Rows {
		return sqlmock.NewRows(columns).
			AddRow(1, userID, "builtin:verification", "自改交叉核验", "保留用户修改的正文").
			AddRow(2, userID, "builtin:chain", "产业链拆解", "用户版本").
			AddRow(3, userID, "builtin:disclosures", "公告优先", "用户版本").
			AddRow(4, userID, "builtin:changes", "复核历史结论", "用户版本").
			AddRow(5, userID, nil, "用户自建方法", "用户自建正文")
	}

	seeded := map[string]bool{
		"builtin:verification": true, "builtin:chain": true,
		"builtin:disclosures": true, "builtin:changes": true,
	}
	var added []ResearchSkill
	var args []driver.Value
	for _, skill := range initialResearchSkills() {
		if seeded[skill.ID] {
			continue
		}
		seeded[skill.ID] = true
		added = append(added, skill)
		args = append(args, sqlmock.AnyArg(), sqlmock.AnyArg(), nil, userID, skill.ID, skill.Name, skill.Instructions)
	}
	if len(added) < 8 {
		t.Fatalf("expected the expanded research catalog, got %d new methods", len(added))
	}
	allRows := func() *sqlmock.Rows {
		rows := existingRows()
		for i, skill := range added {
			rows.AddRow(6+i, userID, skill.ID, skill.Name, skill.Instructions)
		}
		return rows
	}

	read := regexp.QuoteMeta("SELECT * FROM `topic_research_skills` WHERE user_id = ? AND `topic_research_skills`.`deleted_at` IS NULL ORDER BY id ASC")
	mock.ExpectQuery(read).WithArgs(userID).WillReturnRows(existingRows())
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `topic_research_skills` .* ON DUPLICATE KEY UPDATE").
		WithArgs(args...).WillReturnResult(sqlmock.NewResult(6, int64(len(added))))
	mock.ExpectCommit()
	mock.ExpectQuery(read).WithArgs(userID).WillReturnRows(allRows())
	mock.ExpectQuery(read).WithArgs(userID).WillReturnRows(allRows())

	for range 2 {
		got, err := svc.Skills(context.Background(), userID)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != 5+len(added) || got[0].Name != "自改交叉核验" ||
			got[0].Instructions != "保留用户修改的正文" || got[4].Name != "用户自建方法" {
			t.Fatalf("unexpected skill catalog after seeding: %+v", got)
		}
		if !got[0].DefaultSelected || !got[2].DefaultSelected || got[4].DefaultSelected {
			t.Fatal("defaults must use the seed key and exclude custom methods")
		}
		for i, skill := range added {
			row := got[5+i]
			if row.Name != skill.Name || row.Instructions != skill.Instructions || row.ID == skill.ID ||
				row.DefaultSelected != (skill.ID == skillCounterEvidence) {
				t.Fatalf("unexpected selectable skill at index %d: %+v", i, row)
			}
		}
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

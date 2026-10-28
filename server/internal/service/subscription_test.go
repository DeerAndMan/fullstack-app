package service

import (
	"errors"
	"regexp"
	"testing"

	"fullstack-app/server/internal/repository"
	"fullstack-app/server/pkg/errcode"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func subscriptionTestService(t *testing.T) (*SubscriptionService, sqlmock.Sqlmock) {
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
	db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	return NewSubscriptionService(repository.NewSubscriptionRepository(db), nil), mock
}

func TestSubscriptionPreferencesRejectInvalidInput(t *testing.T) {
	negative, tooLarge := int32(-1), int32(1000000)
	starred := true
	cases := []struct {
		name       string
		id, userID int64
		req        *SubscriptionPreferencesRequest
	}{
		{"nil", 1, 2, nil},
		{"empty", 1, 2, &SubscriptionPreferencesRequest{}},
		{"negative order", 1, 2, &SubscriptionPreferencesRequest{SortOrder: &negative}},
		{"excessive order", 1, 2, &SubscriptionPreferencesRequest{SortOrder: &tooLarge}},
		{"invalid ID", 0, 2, &SubscriptionPreferencesRequest{IsStarred: &starred}},
		{"invalid user ID", 1, -2, &SubscriptionPreferencesRequest{IsStarred: &starred}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc := NewSubscriptionService(nil, nil)
			if err := svc.UpdatePreferences(tc.id, tc.userID, tc.req); err != errcode.ErrBadRequest {
				t.Fatalf("got %v, want bad request", err)
			}
		})
	}
}

func TestSubscriptionPreferencesCanClearStarAndOrder(t *testing.T) {
	svc, mock := subscriptionTestService(t)
	starred, order := false, int32(0)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `xq_subscription` SET `is_starred`=?,`sort_order`=? WHERE id = ? AND user_id = ?")).
		WithArgs(false, 0, int64(373676297131397120), int64(6571686717)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := svc.UpdatePreferences(373676297131397120, 6571686717, &SubscriptionPreferencesRequest{IsStarred: &starred, SortOrder: &order}); err != nil {
		t.Fatal(err)
	}
}

func TestSubscriptionPreferencesPartialUpdateAndNoOp(t *testing.T) {
	for _, exists := range []bool{true, false} {
		t.Run(map[bool]string{true: "unchanged record", false: "missing record"}[exists], func(t *testing.T) {
			svc, mock := subscriptionTestService(t)
			starred := true
			mock.ExpectBegin()
			mock.ExpectExec(regexp.QuoteMeta("UPDATE `xq_subscription` SET `is_starred`=? WHERE id = ? AND user_id = ?")).
				WithArgs(true, 1, 2).WillReturnResult(sqlmock.NewResult(0, 0))
			mock.ExpectCommit()
			count := 0
			if exists {
				count = 1
			}
			mock.ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `xq_subscription` WHERE id = ? AND user_id = ?")).
				WithArgs(1, 2).WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(count))
			err := svc.UpdatePreferences(1, 2, &SubscriptionPreferencesRequest{IsStarred: &starred})
			if exists && err != nil {
				t.Fatal(err)
			}
			if !exists && err != errcode.ErrSubscriptionNotFound {
				t.Fatalf("got %v, want not found", err)
			}
		})
	}
}

func TestSubscriptionPreferencesDatabaseFailure(t *testing.T) {
	svc, mock := subscriptionTestService(t)
	order := int32(10)
	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta("UPDATE `xq_subscription` SET `sort_order`=? WHERE id = ? AND user_id = ?")).
		WithArgs(10, 1, 2).WillReturnError(errors.New("database unavailable"))
	mock.ExpectRollback()
	if err := svc.UpdatePreferences(1, 2, &SubscriptionPreferencesRequest{SortOrder: &order}); err != errcode.ErrInternal {
		t.Fatalf("got %v, want internal error", err)
	}
}

func TestSubscriptionsReadPersistedPreferencesAndExactIDs(t *testing.T) {
	svc, mock := subscriptionTestService(t)
	mock.ExpectQuery(regexp.QuoteMeta("SELECT * FROM `xq_subscription` ORDER BY is_starred DESC,sort_order DESC,id ASC,user_id ASC")).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "description", "enabled", "is_starred", "sort_order"}).
			AddRow(int64(373676297131397120), int64(6571686717), "重点订阅", true, true, 20).
			AddRow(1, 2, "普通订阅", false, false, 0))
	items, err := svc.GetAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].ID != "373676297131397120" || items[0].UserID != "6571686717" || !items[0].IsStarred || items[0].SortOrder != 20 {
		t.Fatalf("unexpected subscriptions: %#v", items)
	}
	if items[1].IsStarred || items[1].SortOrder != 0 {
		t.Fatalf("unexpected defaults: %#v", items[1])
	}
}

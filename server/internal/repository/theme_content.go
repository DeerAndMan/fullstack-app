package repository

import (
	"fullstack-app/server/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ThemeContentRepository struct {
	db *gorm.DB
}

func NewThemeContentRepository(db *gorm.DB) *ThemeContentRepository {
	return &ThemeContentRepository{db: db}
}

func (r *ThemeContentRepository) Create(tc *model.XqThemeContent) error {
	return r.db.Create(tc).Error
}

func (r *ThemeContentRepository) BatchCreateIgnore(tcs []model.XqThemeContent) error {
	return r.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&tcs).Error
}

func (r *ThemeContentRepository) Update(id, userID int64, updates map[string]any) error {
	return r.db.Model(&model.XqThemeContent{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(updates).Error
}

func (r *ThemeContentRepository) Delete(id, userID int64) (int64, error) {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.XqThemeContent{})
	return result.RowsAffected, result.Error
}

func (r *ThemeContentRepository) GetByIDAndUserID(id, userID int64) (*model.XqThemeContent, error) {
	var tc model.XqThemeContent
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&tc).Error
	return &tc, err
}

func (r *ThemeContentRepository) GetByUserID(userID int64, limit, offset int) ([]model.XqThemeContent, int64, error) {
	var total int64
	r.db.Model(&model.XqThemeContent{}).Where("user_id = ?", userID).Count(&total)

	var list []model.XqThemeContent
	q := r.db.Where("user_id = ?", userID).Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	err := q.Find(&list).Error
	return list, total, err
}

func (r *ThemeContentRepository) GetAll(limit, offset int) ([]model.XqThemeContent, int64, error) {
	var total int64
	r.db.Model(&model.XqThemeContent{}).Count(&total)

	var list []model.XqThemeContent
	q := r.db.Order("created_at DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	if offset > 0 {
		q = q.Offset(offset)
	}
	err := q.Find(&list).Error
	return list, total, err
}

func (r *ThemeContentRepository) Exists(id, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.XqThemeContent{}).
		Where("id = ? AND user_id = ?", id, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *ThemeContentRepository) Search(keyword string, userID int64, limit, offset int) ([]model.XqThemeContent, int64, error) {
	var total int64
	q := r.db.Model(&model.XqThemeContent{})

	if userID > 0 {
		q = q.Where("user_id = ?", userID)
	}
	if keyword != "" {
		like := "%" + keyword + "%"
		q = q.Where("text LIKE ? OR screen_name LIKE ? OR meta_keywords LIKE ?", like, like, like)
	}

	q.Count(&total)

	var list []model.XqThemeContent
	q2 := q.Session(&gorm.Session{}).Order("created_at DESC")
	if limit > 0 {
		q2 = q2.Limit(limit)
	}
	if offset > 0 {
		q2 = q2.Offset(offset)
	}
	err := q2.Find(&list).Error
	return list, total, err
}

func (r *ThemeContentRepository) SubscriptionExistsByUserID(userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.XqSubscription{}).
		Where("user_id = ?", userID).
		Count(&count).Error
	return count > 0, err
}

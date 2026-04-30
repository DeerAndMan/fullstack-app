package repository

import (
	"fullstack-app/server/internal/model"

	"gorm.io/gorm"
)

type SubscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) *SubscriptionRepository {
	return &SubscriptionRepository{db: db}
}

func (r *SubscriptionRepository) Create(sub *model.XqSubscription) error {
	return r.db.Create(sub).Error
}

func (r *SubscriptionRepository) Delete(id, userID int64) (int64, error) {
	result := r.db.Where("id = ? AND user_id = ?", id, userID).Delete(&model.XqSubscription{})
	return result.RowsAffected, result.Error
}

func (r *SubscriptionRepository) GetByID(id, userID int64) (*model.XqSubscription, error) {
	var sub model.XqSubscription
	err := r.db.Where("id = ? AND user_id = ?", id, userID).First(&sub).Error
	return &sub, err
}

func (r *SubscriptionRepository) GetByUserID(userID int64) ([]model.XqSubscription, error) {
	var subs []model.XqSubscription
	err := r.db.Where("user_id = ?", userID).Find(&subs).Error
	return subs, err
}

func (r *SubscriptionRepository) GetAll() ([]model.XqSubscription, error) {
	var subs []model.XqSubscription
	err := r.db.Find(&subs).Error
	return subs, err
}

func (r *SubscriptionRepository) UpdateEnabled(userID int64, enabled bool) (int64, error) {
	result := r.db.Model(&model.XqSubscription{}).
		Where("user_id = ?", userID).
		Update("enabled", enabled)
	return result.RowsAffected, result.Error
}

func (r *SubscriptionRepository) Exists(id, userID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.XqSubscription{}).
		Where("id = ? AND user_id = ?", id, userID).
		Count(&count).Error
	return count > 0, err
}

func (r *SubscriptionRepository) UpdateDescription(id, userID int64, desc string) error {
	return r.db.Model(&model.XqSubscription{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("description", desc).Error
}

func (r *SubscriptionRepository) GetAllEnabled() ([]model.XqSubscription, error) {
	var subs []model.XqSubscription
	err := r.db.Where("enabled = ?", true).Find(&subs).Error
	return subs, err
}

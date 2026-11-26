package repository

import (
	"context"
	"errors"

	"fullstack-app/server/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrSkillGroupOccupied = errors.New("skill group is occupied")

func (r *TopicResearchRepository) SkillGroups(ctx context.Context, userID uint) ([]model.TopicResearchSkillGroup, error) {
	groups := []model.TopicResearchSkillGroup{}
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("id ASC").Find(&groups).Error
	return groups, err
}

func (r *TopicResearchRepository) SaveSkillGroup(ctx context.Context, userID, id uint, name string) (model.TopicResearchSkillGroup, error) {
	group := model.TopicResearchSkillGroup{ID: id, UserID: userID, Name: name}
	db := r.db.WithContext(ctx)
	if id == 0 {
		err := db.Create(&group).Error
		return group, err
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&group, id).Error; err != nil {
			return err
		}
		return tx.Model(&group).Update("name", name).Error
	})
	return group, err
}

func (r *TopicResearchRepository) DeleteSkillGroup(ctx context.Context, userID, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var group model.TopicResearchSkillGroup
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&group, id).Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&model.TopicResearchSkill{}).Where("user_id = ? AND group_id = ?", userID, id).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return ErrSkillGroupOccupied
		}
		return tx.Delete(&group).Error
	})
}

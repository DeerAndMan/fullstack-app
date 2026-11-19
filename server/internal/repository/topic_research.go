package repository

import (
	"context"
	"time"

	"fullstack-app/server/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type TopicResearchRepository struct{ db *gorm.DB }

func NewTopicResearchRepository(db *gorm.DB) *TopicResearchRepository {
	return &TopicResearchRepository{db: db}
}

func (r *TopicResearchRepository) Skills(ctx context.Context, userID uint) ([]model.TopicResearchSkill, error) {
	rows := []model.TopicResearchSkill{}
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("id ASC").Find(&rows).Error
	return rows, err
}

func (r *TopicResearchRepository) EnsureSkills(ctx context.Context, rows []model.TopicResearchSkill) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
}

func (r *TopicResearchRepository) SaveSkill(ctx context.Context, userID, id uint, name, instructions string) (model.TopicResearchSkill, error) {
	row := model.TopicResearchSkill{UserID: userID, Name: name, Instructions: instructions}
	db := r.db.WithContext(ctx)
	if id == 0 {
		err := db.Create(&row).Error
		return row, err
	}
	if err := db.Where("user_id = ?", userID).First(&row, id).Error; err != nil {
		return row, err
	}
	row.Name, row.Instructions = name, instructions
	err := db.Model(&row).Where("user_id = ?", userID).Select("name", "instructions").Updates(&row).Error
	return row, err
}

func (r *TopicResearchRepository) AddRun(ctx context.Context, row *model.TopicResearchRun) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *TopicResearchRepository) FinishRun(ctx context.Context, userID, id uint, status, result, errorText string) error {
	return r.db.WithContext(ctx).Model(&model.TopicResearchRun{}).
		Where("id = ? AND user_id = ? AND status = ?", id, userID, "running").
		Updates(map[string]any{"status": status, "result_text": result, "error_text": errorText}).Error
}

func (r *TopicResearchRepository) ExpireRuns(ctx context.Context, userID uint, cutoff time.Time) error {
	return r.db.WithContext(ctx).Model(&model.TopicResearchRun{}).
		Where("user_id = ? AND status = ? AND created_at < ?", userID, "running", cutoff).
		Updates(map[string]any{"status": "failed", "error_text": "查询任务已中断，可以用原问题重新查询。"}).Error
}

func (r *TopicResearchRepository) Runs(ctx context.Context, userID, beforeID uint, limit int) ([]model.TopicResearchRun, error) {
	rows := []model.TopicResearchRun{}
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if beforeID != 0 {
		query = query.Where("id < ?", beforeID)
	}
	err := query.Omit("skill_text", "result_text").Order("id DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *TopicResearchRepository) Run(ctx context.Context, userID, id uint) (model.TopicResearchRun, error) {
	var row model.TopicResearchRun
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&row, id).Error
	return row, err
}

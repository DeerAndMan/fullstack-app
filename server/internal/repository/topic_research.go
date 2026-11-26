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
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("sort_order DESC, id ASC").Find(&rows).Error
	return rows, err
}

func (r *TopicResearchRepository) EnsureSkills(ctx context.Context, rows []model.TopicResearchSkill) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
}

func (r *TopicResearchRepository) SaveSkill(ctx context.Context, userID, id uint, name, instructions string, markdown bool) (model.TopicResearchSkill, error) {
	row := model.TopicResearchSkill{UserID: userID, Name: name, Instructions: instructions, Markdown: markdown}
	db := r.db.WithContext(ctx)
	if id == 0 {
		err := db.Create(&row).Error
		return row, err
	}
	if err := db.Where("user_id = ?", userID).First(&row, id).Error; err != nil {
		return row, err
	}
	row.Name, row.Instructions, row.Markdown = name, instructions, markdown
	err := db.Model(&row).Where("user_id = ?", userID).Select("name", "instructions", "markdown").Updates(&row).Error
	return row, err
}

type SkillGroupMove struct {
	SkillID uint `json:"skillId"`
	GroupID uint `json:"groupId"`
}

// OrderSkills updates the entire account list atomically.
func (r *TopicResearchRepository) OrderSkills(ctx context.Context, userID uint, ids []uint, move *SkillGroupMove) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if move != nil && move.GroupID != 0 {
			var target model.TopicResearchSkillGroup
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&target, move.GroupID).Error; err != nil {
				return err
			}
		}
		var rows []model.TopicResearchSkill
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id", "sort_order").
			Where("user_id = ?", userID).Order("id ASC").Find(&rows).Error; err != nil {
			return err
		}
		if len(rows) != len(ids) {
			return gorm.ErrRecordNotFound
		}
		ranks := make(map[uint]int, len(rows))
		for _, row := range rows {
			ranks[row.ID] = row.SortOrder
		}
		for index, id := range ids {
			old, ok := ranks[id]
			if !ok {
				return gorm.ErrRecordNotFound
			}
			delete(ranks, id)
			rank := len(ids) - index
			updates := map[string]any{}
			if move != nil && move.SkillID == id {
				updates["group_id"] = move.GroupID
			}
			if rank != old {
				updates["sort_order"] = rank
			}
			if len(updates) > 0 {
				if err := tx.Model(&model.TopicResearchSkill{}).Where("user_id = ? AND id = ?", userID, id).
					UpdateColumns(updates).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
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

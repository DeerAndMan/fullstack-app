package service

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"fullstack-app/server/internal/model"
	"fullstack-app/server/internal/repository"
	"fullstack-app/server/pkg/errcode"
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

func (s *TopicResearchService) SkillGroups(ctx context.Context, userID uint) ([]model.TopicResearchSkillGroup, error) {
	return s.repo.SkillGroups(ctx, userID)
}

func (s *TopicResearchService) SaveSkillGroup(ctx context.Context, userID, id uint, name string) (model.TopicResearchSkillGroup, error) {
	name = strings.TrimSpace(name)
	if name == "" || name == "未分组" || len([]rune(name)) > 60 {
		return model.TopicResearchSkillGroup{}, errcode.New(400, "请输入 1–60 字的分组名称，“未分组”为保留名称。", http.StatusBadRequest)
	}
	group, err := s.repo.SaveSkillGroup(ctx, userID, id, name)
	var dbErr *mysql.MySQLError
	if errors.As(err, &dbErr) && dbErr.Number == 1062 {
		return group, errcode.New(409, "分组名称已存在。", http.StatusConflict)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return group, errcode.ErrNotFound
	}
	return group, err
}

func (s *TopicResearchService) DeleteSkillGroup(ctx context.Context, userID, id uint) error {
	err := s.repo.DeleteSkillGroup(ctx, userID, id)
	if errors.Is(err, repository.ErrSkillGroupOccupied) {
		return errcode.New(409, "请先移出组内 Skill。", http.StatusConflict)
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.ErrNotFound
	}
	return err
}

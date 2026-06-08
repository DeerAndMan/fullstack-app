package repository

import (
	"fullstack-app/server/internal/model"

	"gorm.io/gorm"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

func (r *RoleRepository) Create(role *model.Role) error {
	return r.db.Create(role).Error
}

func (r *RoleRepository) GetByID(id uint) (*model.Role, error) {
	var role model.Role
	err := r.db.Where("role_id = ? AND del_flag = ?", id, 0).First(&role).Error
	return &role, err
}

func (r *RoleRepository) Update(role *model.Role) error {
	return r.db.Save(role).Error
}

func (r *RoleRepository) Delete(id uint) error {
	return r.db.Model(&model.Role{}).Where("role_id = ?", id).Update("del_flag", 1).Error
}

func (r *RoleRepository) List(page, pageSize int, keyword string) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64

	query := r.db.Model(&model.Role{}).Where("del_flag = ?", 0)
	if keyword != "" {
		query = query.Where("role_name LIKE ? OR role_key LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("role_id DESC").Find(&roles).Error
	return roles, total, err
}

func (r *RoleRepository) ExistsByName(name string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Role{}).Where("role_name = ? AND del_flag = ?", name, 0).Count(&count).Error
	return count > 0, err
}

func (r *RoleRepository) ExistsByCode(code string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Role{}).Where("role_key = ? AND del_flag = ?", code, 0).Count(&count).Error
	return count > 0, err
}

func (r *RoleRepository) GetAll() ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Where("del_flag = ?", 0).Order("sort ASC, role_id ASC").Find(&roles).Error
	return roles, err
}

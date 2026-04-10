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
	err := r.db.First(&role, id).Error
	return &role, err
}

func (r *RoleRepository) Update(role *model.Role) error {
	return r.db.Save(role).Error
}

func (r *RoleRepository) Delete(id uint) error {
	return r.db.Delete(&model.Role{}, id).Error
}

func (r *RoleRepository) List(page, pageSize int, keyword string) ([]model.Role, int64, error) {
	var roles []model.Role
	var total int64

	query := r.db.Model(&model.Role{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR code LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&roles).Error
	return roles, total, err
}

func (r *RoleRepository) ExistsByName(name string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Role{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

func (r *RoleRepository) ExistsByCode(code string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Role{}).Where("code = ?", code).Count(&count).Error
	return count > 0, err
}

func (r *RoleRepository) GetAll() ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Where("status = 1").Find(&roles).Error
	return roles, err
}

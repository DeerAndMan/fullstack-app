package repository

import (
	"encoding/base64"

	"fullstack-app/server/internal/model"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *UserRepository) GetByName(name string) (*model.User, error) {
	var user model.User
	err := r.db.Where("name = ?", name).First(&user).Error
	return &user, err
}

func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}

func (r *UserRepository) List(page, pageSize int, keyword string) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.db.Model(&model.User{})
	if keyword != "" {
		query = query.Where("name LIKE ? OR email LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err := query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&users).Error
	return users, total, err
}

func (r *UserRepository) ExistsByName(name string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) GetLatestAvatar(userID uint) (string, error) {
	var image model.Image
	err := r.db.Where("relevance_id = ?", userID).Order("uploadTime DESC").First(&image).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return base64.StdEncoding.EncodeToString(image.Image), nil
}

func (r *UserRepository) GetRoleByUserID(userID uint) (*model.Role, error) {
	var userRole model.SysUserRole
	err := r.db.Where("user_id = ?", userID).First(&userRole).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}

	var role model.Role
	err = r.db.Where("role_id = ? AND del_flag = ?", userRole.RoleID, 0).First(&role).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func (r *UserRepository) GetMenusByRoleID(roleID int64) ([]model.Menu, error) {
	var menus []model.Menu
	err := r.db.Select("sys_menu.*").
		Joins("left join sys_menu_role on sys_menu.id = sys_menu_role.menu_id").
		Where("sys_menu_role.role_id = ?", roleID).
		Where("sys_menu_role.is_delete = ?", 0).
		Where("sys_menu.is_delete = ?", 0).
		Order("sys_menu.level ASC, sys_menu.id ASC").
		Find(&menus).Error
	return menus, err
}

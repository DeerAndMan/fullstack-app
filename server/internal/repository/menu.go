package repository

import (
	"fullstack-app/server/internal/model"

	"gorm.io/gorm"
)

type MenuRepository struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

func (r *MenuRepository) ListActive() ([]model.Menu, error) {
	var menus []model.Menu
	err := r.db.Where("is_delete = 0").Order("level ASC, id ASC").Find(&menus).Error
	return menus, err
}

func (r *MenuRepository) GetByLinkURLAndCode(linkURL, menuCode string) (*model.Menu, error) {
	var menu model.Menu
	err := r.db.Where("link_url = ? AND menu_code = ? AND is_delete = 0", linkURL, menuCode).First(&menu).Error
	return &menu, err
}

func (r *MenuRepository) BatchCreate(menus []model.Menu) error {
	return r.db.Create(&menus).Error
}

func (r *MenuRepository) GetMenusByRoleID(roleID int64) ([]model.Menu, error) {
	var menus []model.Menu
	err := r.db.Select("sys_menu.*").
		Joins("LEFT JOIN sys_menu_role ON sys_menu.id = sys_menu_role.menu_id").
		Where("sys_menu_role.role_id = ?", roleID).
		Where("sys_menu_role.is_delete = 0").
		Find(&menus).Error
	return menus, err
}

func (r *MenuRepository) DeleteMenuRoleByRoleAndMenus(roleID int64, menuIDs []int64) error {
	return r.db.Where("role_id = ? AND menu_id IN ?", roleID, menuIDs).
		Delete(&model.SysMenuRole{}).Error
}

func (r *MenuRepository) CreateMenuRole(record *model.SysMenuRole) error {
	return r.db.Create(record).Error
}

func (r *MenuRepository) GetMenusByIDs(ids []int64) ([]model.Menu, error) {
	var menus []model.Menu
	err := r.db.Where("id IN ?", ids).Find(&menus).Error
	return menus, err
}

func (r *MenuRepository) RoleExists(roleID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.Role{}).Where("role_id = ?", roleID).Count(&count).Error
	return count > 0, err
}

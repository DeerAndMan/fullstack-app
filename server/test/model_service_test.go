package test

import (
	"testing"

	"fullstack-app/server/internal/model"
	"fullstack-app/server/internal/service"
)

func TestModelTableNames(t *testing.T) {
	cases := map[string]string{
		"user":             (model.User{}).TableName(),
		"sys_role":         (model.Role{}).TableName(),
		"sys_menu":         (model.Menu{}).TableName(),
		"energy":           (model.Energy{}).TableName(),
		"jy_data":          (model.JyData{}).TableName(),
		"xq_subscription":  (model.XqSubscription{}).TableName(),
		"xq_theme_content": (model.XqThemeContent{}).TableName(),
		"sys_user_role":    (model.SysUserRole{}).TableName(),
		"sys_menu_role":    (model.SysMenuRole{}).TableName(),
	}
	for want, got := range cases {
		if got != want {
			t.Errorf("TableName() = %q, want %q", got, want)
		}
	}
}

func TestUserAndMenuResponseConversion(t *testing.T) {
	if got := service.ToUserResponse(nil); got != nil {
		t.Fatalf("ToUserResponse(nil) = %#v, want nil", got)
	}
	users := []model.User{{ID: 7, Name: "alice", Age: 30, Email: "a@example.test", Status: 1}}
	got := service.ToUserResponses(users)
	if len(got) != 1 || got[0].ID != 7 || got[0].Name != "alice" || got[0].Email != "a@example.test" {
		t.Fatalf("unexpected user conversion: %#v", got)
	}

	menus := []model.Menu{{ID: 12, ParentID: 3, Name: "Dashboard", LinkURL: "/dashboard", MenuCode: "dashboard", NodeType: 1, Level: 2}}
	menuResp := service.ToMenuResponses(menus)
	if len(menuResp) != 1 || menuResp[0].ID != "12" || menuResp[0].ParentID != "3" || menuResp[0].LinkURL != "/dashboard" {
		t.Fatalf("unexpected menu conversion: %#v", menuResp)
	}
}

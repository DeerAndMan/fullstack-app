package errcode

import "net/http"

type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	HTTP    int    `json:"-"`
}

func (e *Error) Error() string {
	return e.Message
}

func New(code int, msg string, httpStatus int) *Error {
	return &Error{Code: code, Message: msg, HTTP: httpStatus}
}

// 通用错误
var (
	Success          = New(0, "操作成功", http.StatusOK)
	ErrBadRequest    = New(400, "请求参数错误", http.StatusBadRequest)
	ErrUnauthorized  = New(401, "未授权", http.StatusUnauthorized)
	ErrForbidden     = New(403, "禁止访问", http.StatusForbidden)
	ErrNotFound      = New(404, "资源不存在", http.StatusNotFound)
	ErrInternal      = New(500, "服务器内部错误", http.StatusInternalServerError)
)

// 认证错误
var (
	ErrInvalidCredentials = New(10001, "用户名或密码错误", http.StatusUnauthorized)
	ErrTokenExpired       = New(10002, "登录已过期", http.StatusUnauthorized)
	ErrTokenInvalid       = New(10003, "无效的令牌", http.StatusUnauthorized)
	ErrUsernameExists     = New(10004, "用户名已存在", http.StatusBadRequest)
)

// 用户错误
var (
	ErrUserNotFound = New(20001, "用户不存在", http.StatusNotFound)
	ErrUserDisabled = New(20002, "用户已被禁用", http.StatusForbidden)
)

// 角色错误
var (
	ErrRoleNotFound   = New(30001, "角色不存在", http.StatusNotFound)
	ErrRoleNameExists = New(30002, "角色名称已存在", http.StatusBadRequest)
	ErrRoleCodeExists = New(30003, "角色编码已存在", http.StatusBadRequest)
)

// 上传错误
var (
	ErrFileTooLarge       = New(40001, "文件过大", http.StatusBadRequest)
	ErrFileTypeNotAllowed = New(40002, "不支持的文件类型", http.StatusBadRequest)
)

// 菜单错误
var (
	ErrMenuNotFound  = New(50001, "菜单不存在", http.StatusNotFound)
	ErrMenuDuplicate = New(50002, "菜单已存在", http.StatusBadRequest)
)

// 订阅错误
var (
	ErrSubscriptionNotFound  = New(60001, "订阅不存在", http.StatusNotFound)
	ErrSubscriptionDuplicate = New(60002, "订阅已存在", http.StatusBadRequest)
)

// 主题内容错误
var (
	ErrThemeContentNotFound  = New(70001, "主题内容不存在", http.StatusNotFound)
	ErrThemeContentDuplicate = New(70002, "主题内容已存在", http.StatusBadRequest)
)

// 用户角色错误
var (
	ErrRoleDisabled    = New(30004, "角色已被禁用", http.StatusBadRequest)
	ErrRoleInvalid     = New(30005, "包含无效的角色", http.StatusBadRequest)
)

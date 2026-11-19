package test

import (
	"net/http"
	"testing"

	"fullstack-app/server/pkg/errcode"
)

func TestErrcodeNewAndError(t *testing.T) {
	err := errcode.New(12345, "测试错误", http.StatusTeapot)
	if err.Code != 12345 {
		t.Fatalf("Code = %d, want %d", err.Code, 12345)
	}
	if err.Message != "测试错误" {
		t.Fatalf("Message = %q, want %q", err.Message, "测试错误")
	}
	if err.HTTP != http.StatusTeapot {
		t.Fatalf("HTTP = %d, want %d", err.HTTP, http.StatusTeapot)
	}
	if err.Error() != "测试错误" {
		t.Fatalf("Error() = %q, want %q", err.Error(), "测试错误")
	}
}

func TestErrcodeCommonErrors(t *testing.T) {
	tests := []struct {
		name string
		err  *errcode.Error
		code int
		http int
	}{
		{name: "success", err: errcode.Success, code: 0, http: http.StatusOK},
		{name: "bad request", err: errcode.ErrBadRequest, code: 400, http: http.StatusBadRequest},
		{name: "unauthorized", err: errcode.ErrUnauthorized, code: 401, http: http.StatusUnauthorized},
		{name: "forbidden", err: errcode.ErrForbidden, code: 403, http: http.StatusForbidden},
		{name: "not found", err: errcode.ErrNotFound, code: 404, http: http.StatusNotFound},
		{name: "internal", err: errcode.ErrInternal, code: 500, http: http.StatusInternalServerError},
		{name: "invalid credentials", err: errcode.ErrInvalidCredentials, code: 10001, http: http.StatusUnauthorized},
		{name: "user not found", err: errcode.ErrUserNotFound, code: 20001, http: http.StatusNotFound},
		{name: "role disabled", err: errcode.ErrRoleDisabled, code: 30004, http: http.StatusBadRequest},
		{name: "file too large", err: errcode.ErrFileTooLarge, code: 40001, http: http.StatusBadRequest},
		{name: "menu not found", err: errcode.ErrMenuNotFound, code: 50001, http: http.StatusNotFound},
		{name: "subscription not found", err: errcode.ErrSubscriptionNotFound, code: 60001, http: http.StatusNotFound},
		{name: "theme content not found", err: errcode.ErrThemeContentNotFound, code: 70001, http: http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatal("error value is nil")
			}
			if tt.err.Code != tt.code {
				t.Errorf("Code = %d, want %d", tt.err.Code, tt.code)
			}
			if tt.err.HTTP != tt.http {
				t.Errorf("HTTP = %d, want %d", tt.err.HTTP, tt.http)
			}
			if tt.err.Error() != tt.err.Message {
				t.Errorf("Error() = %q, want Message %q", tt.err.Error(), tt.err.Message)
			}
		})
	}
}

package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"fullstack-app/server/pkg/errcode"
)

type AiService struct {
	baseURL string
	token   string
}

func NewAiService(baseURL, token string) *AiService {
	return &AiService{baseURL: baseURL, token: token}
}

func (s *AiService) GetConversations(user string) (interface{}, error) {
	url := fmt.Sprintf("%s/v1/conversations?user=%s", s.baseURL, user)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errcode.New(500, "创建请求错误", 500)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errcode.New(500, "请求错误", 500)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errcode.New(500, "读取响应错误", 500)
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, errcode.New(500, "解析响应错误", 500)
	}

	return data, nil
}

package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"fullstack-app/server/pkg/errcode"
)

type SseService struct {
	baseURL string
	token   string
}

func NewSseService(baseURL, token string) *SseService {
	return &SseService{baseURL: baseURL, token: token}
}

type ChatRequest struct {
	Inputs         map[string]interface{} `json:"inputs"`
	Query          string                 `json:"query" vd:"len($)>0"`
	ResponseMode   string                 `json:"response_mode"`
	ConversationID string                 `json:"conversation_id"`
	User           string                 `json:"user" vd:"len($)>0"`
	Files          []ChatFile             `json:"files"`
}

type ChatFile struct {
	Type           string `json:"type"`
	TransferMethod string `json:"transfer_method"`
	URL            string `json:"url"`
}

func (s *SseService) ChatMessages(reqData *ChatRequest) (*http.Response, error) {
	body, err := json.Marshal(reqData)
	if err != nil {
		return nil, errcode.New(500, "请求数据序列化失败", 500)
	}

	req, err := http.NewRequest("POST", s.baseURL+"/v1/chat-messages", bytes.NewBuffer(body))
	if err != nil {
		return nil, errcode.New(500, "创建请求失败", 500)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errcode.New(500, "请求失败", 500)
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, errcode.New(resp.StatusCode, fmt.Sprintf("AI 服务返回错误: %d", resp.StatusCode), resp.StatusCode)
	}

	return resp, nil
}

func (s *SseService) GetChatMessages(conversationID, user string) (interface{}, error) {
	url := fmt.Sprintf("%s/v1/messages?user=%s&conversation_id=%s", s.baseURL, user, conversationID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, errcode.New(500, "创建请求失败", 500)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errcode.New(500, "请求失败", 500)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errcode.New(resp.StatusCode, "请求失败", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errcode.New(500, "读取数据失败", 500)
	}

	var data interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, errcode.New(500, "解析数据失败", 500)
	}

	return data, nil
}

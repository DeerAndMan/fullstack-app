package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
)

const researchModel = "deepseek-flash"
const researchEndpoint = "https://api.deepseek.com"

func researchAPIKey(allowLocal bool) string {
	if env := strings.TrimSpace(os.Getenv("DEEPSEEK_API_KEY")); env != "" {
		return env
	}
	if allowLocal {
		if data, err := os.ReadFile("config/.env.deepseek.local"); err == nil {
			if value, ok := strings.CutPrefix(strings.TrimSpace(string(data)), "DEEPSEEK_API_KEY="); ok {
				return strings.TrimSpace(value)
			}
		}
	}
	return ""
}

func researchJSON[T any](ctx context.Context, client *http.Client, key, path string, input any) (T, error) {
	var result T
	body, err := json.Marshal(input)
	if err != nil {
		return result, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, researchEndpoint+path, bytes.NewReader(body))
	if err != nil {
		return result, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)
	res, err := client.Do(req)
	if err != nil {
		return result, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return result, fmt.Errorf("DeepSeek HTTP %d", res.StatusCode)
	}
	err = json.NewDecoder(io.LimitReader(res.Body, 3<<20)).Decode(&result)
	return result, err
}

func researchChat[T any](ctx context.Context, client *http.Client, key, system string, payload any, maxTokens int) (T, error) {
	type chatResult struct {
		Choices []struct {
			Finish  string `json:"finish_reason"`
			Message struct{ Content string }
		}
	}
	var result T
	user, err := json.Marshal(payload)
	if err != nil {
		return result, err
	}
	output, err := researchJSON[chatResult](ctx, client, key, "/chat/completions", map[string]any{
		"model": researchModel, "thinking": map[string]string{"type": "disabled"},
		"response_format": map[string]string{"type": "json_object"}, "max_tokens": maxTokens,
		"messages": []map[string]string{{"role": "system", "content": system}, {"role": "user", "content": string(user)}},
	})
	if err != nil {
		return result, err
	}
	if len(output.Choices) == 0 || output.Choices[0].Finish != "stop" {
		return result, errors.New("DeepSeek output is incomplete")
	}
	err = json.Unmarshal([]byte(output.Choices[0].Message.Content), &result)
	return result, err
}

type researchSearchHit struct {
	Title, URL, Snippet, Date, Content string
}

var researchMarkup = regexp.MustCompile(`<[^>]+>`)

func researchSearch(ctx context.Context, client *http.Client, key, query string, hosts []string) ([]ResearchSource, error) {
	return researchSearchPages(ctx, client, key, query, hosts, false)
}

func researchSearchPages(ctx context.Context, client *http.Client, key, query string, hosts []string, needContent bool) ([]ResearchSource, error) {
	query = strings.Join(strings.Fields(query), " ")
	query = strings.Map(func(r rune) rune {
		if r == ':' || r == '"' || r == '\'' {
			return ' '
		}
		return r
	}, query)
	sites := make([]string, 0, len(hosts))
	for _, host := range hosts {
		sites = append(sites, "site:"+host)
	}
	scope := strings.Join(sites, " OR ")
	request := map[string]any{"query": truncateResearchText(query, 127-len(scope)) + " " + scope, "count": 8, "need_content": needContent}
	result, err := researchJSON[struct{ Results []researchSearchHit }](ctx, client, key, "/v1/tools/search", request)
	if err != nil {
		return nil, err
	}
	var out []ResearchSource
	seen := map[string]bool{}
	for _, row := range result.Results {
		link, err := url.Parse(row.URL)
		if err != nil || (link.Scheme != "https" && link.Scheme != "http") || link.User != nil || seen[row.URL] {
			continue
		}
		if link.Hostname() == "guba.eastmoney.com" || link.Hostname() == "caifuhao.eastmoney.com" {
			continue
		}
		source := ""
		for _, site := range hosts {
			if link.Hostname() == site || strings.HasSuffix(link.Hostname(), "."+site) {
				source = site
				break
			}
		}
		if source == "" || row.Title == "" {
			continue
		}
		seen[row.URL] = true
		out = append(out, ResearchSource{
			Provider: source, Title: truncateResearchText(researchMarkup.ReplaceAllString(row.Title, ""), 130), URL: row.URL,
			Date: row.Date, Excerpt: truncateResearchText(researchMarkup.ReplaceAllString(row.Snippet, " "), 220), Kind: "search",
			content: truncateResearchText(row.Content, 90000),
		})
	}
	return out, nil
}

func truncateResearchText(text string, max int) string {
	runes := []rune(strings.TrimSpace(text))
	if len(runes) > max {
		return string(runes[:max])
	}
	return string(runes)
}

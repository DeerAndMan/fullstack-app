package v1

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/response"
	"github.com/cloudwego/hertz/pkg/app"
)

type reviewRefresh struct {
	sync.Mutex
	run         reviewRefreshRun
	config      string
	sourcesOnce sync.Once
	sources     []reviewSource
}

type reviewSource struct {
	ID     string `json:"id"`
	Label  string `json:"label"`
	Reason string `json:"reason"`
}

func reviewPython() string {
	if python := os.Getenv("REVIEW_PYTHON"); python != "" {
		return python
	}
	venv := filepath.Join("..", "tmp", "review-python", "bin", "python")
	if _, err := os.Stat(venv); err == nil {
		return venv
	}
	return "python3"
}

func (r *reviewRefresh) providers() []reviewSource {
	r.sourcesOnce.Do(func() {
		out, err := exec.Command(reviewPython(), "-B", filepath.Join("scripts", "sync_market_review.py"), "--list-providers").Output()
		if err != nil || json.Unmarshal(out, &r.sources) != nil || len(r.sources) == 0 {
			r.sources = []reviewSource{}
		}
	})
	return r.sources
}

func (r *reviewRefresh) automatic() bool {
	for _, source := range r.providers() {
		if source.ID == "public" {
			return source.Reason == ""
		}
	}
	return false
}

type reviewRefreshRun struct {
	Provider      string `json:"provider"`
	Date          string `json:"date"`
	State         string `json:"state"`
	Message       string `json:"message"`
	ID            int64  `json:"id"`
	CorePublished bool   `json:"corePublished"`
}

func reviewLocalTime(now time.Time) time.Time {
	return now.In(time.FixedZone("Asia/Shanghai", 8*3600))
}

func (h *MarketReviewHandler) RefreshStatus(ctx context.Context, c *app.RequestContext) {
	local := reviewLocalTime(time.Now())
	today := local.Format(time.DateOnly)
	midnight := time.Date(local.Year(), local.Month(), local.Day()+1, 0, 0, 0, 0, local.Location())
	h.refresh.Lock()
	run := h.refresh.run
	h.refresh.Unlock()
	response.OK(ctx, c, map[string]any{"today": today, "allowed": true, "windowChangeMs": midnight.Sub(local).Milliseconds(), "run": run, "providers": h.refresh.providers(), "auto": h.refresh.automatic()})
}

func (h *MarketReviewHandler) StartRefresh(ctx context.Context, c *app.RequestContext) {
	today := reviewLocalTime(time.Now()).Format(time.DateOnly)
	var input struct{ Date, Provider string }
	if c.BindJSON(&input) != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}
	if input.Date == "" {
		input.Date = today
	}
	if _, err := time.Parse(time.DateOnly, input.Date); err != nil || input.Date > today {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, "请选择今天或更早的日期")
		return
	}
	if input.Provider == "" {
		input.Provider = "public"
		if input.Date != today {
			input.Provider = "baostock"
		}
	} // Previously shipped clients.
	enabled := false
	for _, p := range h.refresh.providers() {
		if p.ID == input.Provider && p.Reason == "" {
			enabled = true
			break
		}
	}
	if !enabled {
		response.FailWithMessage(ctx, c, errcode.ErrBadRequest, "行情来源不可用，请检查服务端配置")
		return
	}
	h.refresh.Lock()
	defer h.refresh.Unlock()
	if h.refresh.run.State == "running" {
		if h.refresh.run.Date != input.Date || h.refresh.run.Provider != input.Provider {
			response.FailWithMessage(ctx, c, errcode.ErrBadRequest, "已有其他日期或来源的更新正在运行")
			return
		}
		response.OK(ctx, c, h.refresh.run)
		return
	}
	h.refresh.run = reviewRefreshRun{Date: input.Date, Provider: input.Provider, ID: time.Now().UnixMilli(), State: "running", Message: "正在检查归档缺口"}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Hour)
		defer cancel()
		h.refresh.execute(ctx, input.Date, input.Provider, false)
	}()
	response.OK(ctx, c, h.refresh.run)
}

func (r *reviewRefresh) execute(ctx context.Context, date, provider string, scheduled bool) {
	args := []string{"-B", "-u", filepath.Join("scripts", "sync_market_review.py"), "--refresh-date", date, "--provider", provider}
	if scheduled {
		args = append(args, "--scheduled")
	}
	cmd := exec.CommandContext(ctx, reviewPython(), args...)
	cmd.Env = append(os.Environ(), "REVIEW_IMPORT_CONFIG="+r.config)
	cmd.Stderr = os.Stderr
	out, err := cmd.StdoutPipe()
	if err == nil {
		err = cmd.Start()
	}
	if err == nil {
		scanner := bufio.NewScanner(out)
		for scanner.Scan() {
			if msg, ok := strings.CutPrefix(scanner.Text(), "REVIEW "); ok {
				r.Lock()
				if core, published := strings.CutPrefix(msg, "CORE "); published {
					r.run.CorePublished = true
					msg = core
				}
				r.run.Message = msg
				r.Unlock()
			}
		}
		err = cmd.Wait()
	}
	r.Lock()
	defer r.Unlock()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 77 {
		r.run.State = "partial"
		return
	}
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 75 {
		r.run.State = "unpublished"
		return
	}
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 76 {
		r.run = reviewRefreshRun{}
		return
	}
	if err != nil {
		r.run.State = "failed"
		if r.run.CorePublished {
			r.run.Message += "。此步骤未完成，可再次更新。"
			return
		}
		r.run.Message += "。更新未完成，请检查行情发布时间、同步服务或 AI 服务后重试。"
		return
	}
	r.run.State = "done"
	r.run.Message = "市场指标、专题和 AI 建议已归档"
}

// RunCloseRefresh warms the historical data and publishes the daily close.
func (h *MarketReviewHandler) RunCloseRefresh(ctx context.Context) {
	r := &h.refresh
	if !r.automatic() {
		return
	}
	ticker := time.NewTicker(90 * time.Second)
	defer ticker.Stop()
	var warmed, finished string
	for {
		now := reviewLocalTime(time.Now())
		day := now.Format(time.DateOnly)
		switch reviewSchedule(now) {
		case "warm":
			if warmed != day {
				warmed = day
				warmCtx, cancel := context.WithTimeout(ctx, 5*time.Hour)
				cmd := exec.CommandContext(warmCtx, reviewPython(), "-B", "-u", filepath.Join("scripts", "sync_market_review.py"),
					"--refresh-date", day, "--provider", "public", "--prewarm", "--scheduled")
				cmd.Env = append(os.Environ(), "REVIEW_IMPORT_CONFIG="+r.config)
				cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
				_ = cmd.Run()
				cancel()
			}
		case "close":
			if finished != day {
				r.Lock()
				busy := r.run.State == "running"
				if !busy {
					r.run = reviewRefreshRun{Date: day, Provider: "public", ID: now.UnixMilli(), State: "running", Message: "正在核对公开收盘行情"}
				}
				r.Unlock()
				if !busy {
					closeCtx, cancel := context.WithTimeout(ctx, 2*time.Hour)
					r.execute(closeCtx, day, "public", true)
					cancel()
					r.Lock()
					if r.run.State == "done" || r.run.State == "partial" || r.run.State == "" {
						finished = day
					}
					r.Unlock()
				}
			}

		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func reviewSchedule(now time.Time) string {
	if now.Weekday() == time.Saturday || now.Weekday() == time.Sunday {
		return ""
	}
	minute := now.Hour()*60 + now.Minute()
	if minute >= 8*60 && minute < 15*60 {
		return "warm"
	}
	if minute >= 15*60+5 && minute < 19*60 {
		return "close"
	}
	return ""
}

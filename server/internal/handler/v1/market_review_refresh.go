package v1

import (
	"bufio"
	"context"
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
	run    reviewRefreshRun
	config string
}

type reviewRefreshRun struct {
	Date    string `json:"date"`
	State   string `json:"state"`
	Message string `json:"message"`
	ID      int64  `json:"id"`
}

func refreshWindow(now time.Time) (string, bool) {
	local := now.In(time.FixedZone("Asia/Shanghai", 8*3600))
	return local.Format(time.DateOnly), local.Hour() >= 15
}

func (h *MarketReviewHandler) RefreshStatus(ctx context.Context, c *app.RequestContext) {
	now := time.Now()
	today, allowed := refreshWindow(now)
	local := now.In(time.FixedZone("Asia/Shanghai", 8*3600))
	change := time.Date(local.Year(), local.Month(), local.Day(), 15, 0, 0, 0, local.Location())
	if allowed {
		change = time.Date(local.Year(), local.Month(), local.Day()+1, 0, 0, 0, 0, local.Location())
	}
	h.refresh.Lock()
	run := h.refresh.run
	h.refresh.Unlock()
	response.OK(ctx, c, map[string]any{"today": today, "allowed": allowed, "windowChangeMs": change.Sub(now).Milliseconds(), "run": run})
}

func (h *MarketReviewHandler) StartRefresh(ctx context.Context, c *app.RequestContext) {
	today, allowed := refreshWindow(time.Now())
	if !allowed {
		response.FailWithMessage(ctx, c, errcode.ErrForbidden, "北京时间 15:00 起可更新复盘数据")
		return
	}
	var input struct{ Date string }
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
	h.refresh.Lock()
	defer h.refresh.Unlock()
	if h.refresh.run.State == "running" {
		if h.refresh.run.Date != input.Date {
			response.FailWithMessage(ctx, c, errcode.ErrBadRequest, "另一交易日正在更新")
			return
		}
		response.OK(ctx, c, h.refresh.run)
		return
	}
	h.refresh.run = reviewRefreshRun{Date: input.Date, ID: time.Now().UnixMilli(), State: "running", Message: "等待收盘同步"}
	go h.refresh.execute(input.Date)
	response.OK(ctx, c, h.refresh.run)
}

func (r *reviewRefresh) execute(date string) {
	cmd := exec.Command("python3", "-u", filepath.Join("scripts", "sync_market_review.py"), "--refresh-date", date)
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
				r.run.Message = msg
				r.Unlock()
			}
		}
		err = cmd.Wait()
	}
	r.Lock()
	defer r.Unlock()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 75 {
		r.run.State = "unpublished"
		return
	}
	if err != nil {
		r.run.State = "failed"
		r.run.Message += "。更新未完成，请检查行情发布时间、同步服务或 AI 服务后重试。"
		return
	}
	r.run.State = "done"
	r.run.Message = "市场指标、专题和新生成的 AI 建议已归档"
}

package v1

import (
	"context"
	"errors"
	"fullstack-app/server/internal/middleware"
	"fullstack-app/server/internal/model"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"fullstack-app/server/internal/repository"
	"fullstack-app/server/internal/service"
	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/response"
	"github.com/cloudwego/hertz/pkg/app"
	"gorm.io/gorm"
)

type MarketReviewHandler struct {
	repo    *repository.MarketReviewRepository
	refresh reviewRefresh
}

func NewMarketReviewHandler(repo *repository.MarketReviewRepository, configPath string) *MarketReviewHandler {
	return &MarketReviewHandler{repo: repo, refresh: reviewRefresh{config: configPath}}
}

func (h *MarketReviewHandler) Dates(ctx context.Context, c *app.RequestContext) {
	dates, err := h.repo.Dates(ctx)
	writeMarketResult(ctx, c, dates, err)
}

func (h *MarketReviewHandler) Timeline(ctx context.Context, c *app.RequestContext) {
	through := c.Query("through")
	if _, err := time.Parse(time.DateOnly, through); err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}
	rows, err := h.repo.Timeline(ctx, through)
	writeMarketResult(ctx, c, rows, err)
}

func (h *MarketReviewHandler) Daily(ctx context.Context, c *app.RequestContext) {
	date := c.Query("date")
	if _, err := time.Parse(time.DateOnly, date); err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}
	day, err := h.repo.Day(ctx, date, "")
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.OK(ctx, c, nil)
		return
	}
	if err != nil {
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	previous, err := h.repo.Day(ctx, day.Previous, "")
	var delta *int
	if err == nil && previous.Rule == day.Rule {
		n := day.Highs - previous.Highs
		delta = &n
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	response.OK(ctx, c, map[string]any{"day": day, "highDelta": delta})
}

func (h *MarketReviewHandler) Members(ctx context.Context, c *app.RequestContext) {
	date, batch, metric := c.Query("date"), c.Query("batch"), c.Query("metric")
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	_, dateErr := time.Parse(time.DateOnly, date)
	knownMetric := slices.Contains([]string{"up", "high", "csi", "excluded", "all"}, metric)
	if dateErr != nil || len(batch) != 64 || err != nil || page < 1 || page > 100000 || !knownMetric {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}
	day, err := h.repo.Day(ctx, date, batch)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.Fail(ctx, c, errcode.ErrNotFound)
		return
	}
	if err != nil {
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	rows, total, err := h.repo.Members(ctx, day.ID, metric, c.Query("search"), page)
	writeMarketResult(ctx, c, map[string]any{"rows": rows, "total": total}, err)
}

func (h *MarketReviewHandler) Study(ctx context.Context, c *app.RequestContext) {
	date, batch, kind := c.Query("date"), c.Query("batch"), c.Query("kind")
	_, dateErr := time.Parse(time.DateOnly, date)
	if dateErr != nil || len(batch) != 64 || !slices.Contains([]string{"style", "sectors", "patterns"}, kind) {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}
	day, err := h.repo.Day(ctx, date, batch)
	if err != nil {
		writeMarketResult(ctx, c, nil, err)
		return
	}
	payload, err := h.repo.Study(ctx, day.ID, kind)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.OK(ctx, c, nil)
		return
	}
	if err != nil {
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	if kind != "patterns" {
		response.OK(ctx, c, payload)
		return
	}
	t := service.DefaultReviewThresholds()
	for key, field := range map[string]*float64{"big": &t.Big, "near": &t.Near, "width": &t.Width, "doji": &t.Doji, "shrink": &t.Shrink} {
		if raw := c.Query(key); raw != "" {
			*field, err = strconv.ParseFloat(raw, 64)
			if err != nil {
				response.Fail(ctx, c, errcode.ErrBadRequest)
				return
			}
		}
	}
	page, e := strconv.Atoi(c.DefaultQuery("page", "1"))
	stage := c.DefaultQuery("stage", "signal")
	if !t.Valid() || e != nil || page < 1 || page > 100000 || (stage != "signal" && stage != "trigger") {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}
	rows, err := service.FilterReviewPatterns(payload, t, c.Query("pattern"), c.Query("search"), stage)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrInternal)
		return
	}
	total := len(rows)
	first := min((page-1)*50, total)
	response.OK(ctx, c, map[string]any{"total": total, "rows": rows[first:min(first+50, total)]})
}

var reviewSymbol = regexp.MustCompile(`^(6[0-9]{5}\.SH|(00|30)[0-9]{4}\.SZ)$`)

func (h *MarketReviewHandler) Note(ctx context.Context, c *app.RequestContext) {
	date, symbol := c.Query("date"), c.Query("symbol")
	if _, e := time.Parse(time.DateOnly, date); e != nil || !reviewSymbol.MatchString(symbol) {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}
	note, err := h.repo.Note(ctx, middleware.GetUserID(c), date, symbol)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.OK(ctx, c, nil)
		return
	}
	writeMarketResult(ctx, c, note, err)
}
func (h *MarketReviewHandler) SaveNote(ctx context.Context, c *app.RequestContext) {
	var input struct{ Date, Symbol, Text string }
	if c.BindJSON(&input) != nil || !reviewSymbol.MatchString(input.Symbol) || utf8.RuneCountInString(input.Text) > 1000 {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}
	if _, e := time.Parse(time.DateOnly, input.Date); e != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}
	day, err := h.repo.Day(ctx, input.Date, "")
	if err != nil {
		writeMarketResult(ctx, c, nil, err)
		return
	}
	members, _, err := h.repo.Members(ctx, day.ID, "all", input.Symbol, 1)
	if err != nil {
		writeMarketResult(ctx, c, nil, err)
		return
	}
	if len(members) != 1 || members[0].Symbol != input.Symbol {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}
	note := model.MarketReviewNote{UserID: middleware.GetUserID(c), Date: input.Date, Symbol: input.Symbol, Text: strings.TrimSpace(input.Text)}
	err = h.repo.SaveNote(ctx, &note)
	writeMarketResult(ctx, c, note, err)
}

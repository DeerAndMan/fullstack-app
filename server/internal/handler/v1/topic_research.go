package v1

import (
	"context"
	"errors"
	"strconv"
	"time"

	"fullstack-app/server/internal/middleware"
	"fullstack-app/server/internal/service"
	"fullstack-app/server/pkg/errcode"
	"fullstack-app/server/pkg/response"

	"github.com/cloudwego/hertz/pkg/app"
)

type TopicResearchHandler struct {
	research *service.TopicResearchService
}

func NewTopicResearchHandler(research *service.TopicResearchService) *TopicResearchHandler {
	return &TopicResearchHandler{research: research}
}

func researchResponse(ctx context.Context, c *app.RequestContext, body any, err error) {
	if err == nil {
		response.OK(ctx, c, body)
		return
	}
	var apiError *errcode.Error
	if errors.As(err, &apiError) {
		response.Fail(ctx, c, apiError)
		return
	}
	response.Fail(ctx, c, errcode.ErrInternal)
}

func (h *TopicResearchHandler) Skills(ctx context.Context, c *app.RequestContext) {
	data, err := h.research.Skills(ctx, middleware.GetUserID(c))
	researchResponse(ctx, c, data, err)
}

func (h *TopicResearchHandler) SaveSkill(ctx context.Context, c *app.RequestContext) {
	var req service.ResearchSkillRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}
	data, err := h.research.SaveSkill(ctx, middleware.GetUserID(c), req)
	researchResponse(ctx, c, data, err)
}

func (h *TopicResearchHandler) Runs(ctx context.Context, c *app.RequestContext) {
	beforeID, err := strconv.ParseUint(c.DefaultQuery("before", "0"), 10, strconv.IntSize)
	if err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}
	data, err := h.research.Runs(ctx, middleware.GetUserID(c), uint(beforeID))
	researchResponse(ctx, c, data, err)
}

func (h *TopicResearchHandler) Run(ctx context.Context, c *app.RequestContext) {
	id, err := strconv.ParseUint(c.Param("id"), 10, strconv.IntSize)
	if err != nil || id == 0 {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}
	data, err := h.research.Run(ctx, middleware.GetUserID(c), uint(id))
	researchResponse(ctx, c, data, err)
}

func (h *TopicResearchHandler) Start(ctx context.Context, c *app.RequestContext) {
	var req service.ResearchRunRequest
	if err := c.BindAndValidate(&req); err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}
	data, err := h.research.Start(ctx, middleware.GetUserID(c), req)
	researchResponse(ctx, c, data, err)
}

func (h *TopicResearchHandler) ReviewAdvice(ctx context.Context, c *app.RequestContext) {
	var input struct{ Date, Batch string }
	if c.BindJSON(&input) != nil || len(input.Batch) != 64 {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}
	if _, err := time.Parse(time.DateOnly, input.Date); err != nil {
		response.Fail(ctx, c, errcode.ErrBadRequest)
		return
	}
	data, err := h.research.GenerateReviewAdvice(ctx, input.Date, input.Batch)
	researchResponse(ctx, c, data, err)
}

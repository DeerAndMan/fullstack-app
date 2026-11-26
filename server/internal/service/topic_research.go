package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"fullstack-app/server/internal/model"
	"fullstack-app/server/internal/repository"
	"fullstack-app/server/pkg/errcode"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	errResearchBusy = errcode.New(429, "AI查询正在进行，请稍后再试。", http.StatusTooManyRequests)
	errResearchKey  = errcode.New(503, "AI查询需要配置服务端 DEEPSEEK_API_KEY。", http.StatusServiceUnavailable)
)

type TopicResearchService struct {
	Reviews  *repository.MarketReviewRepository
	repo     *repository.TopicResearchRepository
	market   *MarketService
	client   *http.Client
	localKey bool
	active   sync.Map
	phases   sync.Map
	replies  sync.Map
	slots    chan struct{}
}

func NewTopicResearchService(repo *repository.TopicResearchRepository, market *MarketService, localKey bool) *TopicResearchService {
	return &TopicResearchService{repo: repo, market: market, localKey: localKey, slots: make(chan struct{}, 4),
		client: &http.Client{Timeout: 32 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

type ResearchSource struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	Title    string `json:"title"`
	URL      string `json:"url"`
	Kind     string `json:"kind"`
	Date     string `json:"date,omitempty"`
	Excerpt  string `json:"excerpt,omitempty"`
	content  string
}

type ResearchCandidate struct {
	Code      string   `json:"code"`
	Name      string   `json:"name"`
	Status    string   `json:"status"`
	Topics    []string `json:"topics"`
	SourceIDs []string `json:"sourceIds"`
}

type ResearchObservation struct {
	Code      string   `json:"code"`
	Name      string   `json:"name,omitempty"`
	Note      string   `json:"note"`
	SourceIDs []string `json:"sourceIds"`
	QuoteURL  string   `json:"quoteUrl,omitempty"`
}

type ResearchReport struct {
	Reply        string                `json:"reply,omitempty"`
	Topics       []string              `json:"topics"`
	Match        string                `json:"match"`
	Candidates   []ResearchCandidate   `json:"candidates"`
	Sources      []ResearchSource      `json:"sources"`
	Notes        []string              `json:"notes"`
	Conclusion   string                `json:"conclusion,omitempty"`
	Observations []ResearchObservation `json:"observations,omitempty"`
	SkillChecks  []ResearchSkillCheck  `json:"skillChecks,omitempty"`
	Optimus      *OptimusReport        `json:"optimus,omitempty"`
}

type ResearchRunRequest struct {
	Question    string   `json:"question"`
	SkillIDs    []string `json:"skillIds"`
	ReferenceID uint     `json:"referenceId"`
}

type ResearchSkillRequest struct {
	Markdown     bool   `json:"markdown"`
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Instructions string `json:"instructions"`
}

type ResearchRunView struct {
	model.TopicResearchRun
	History []string        `json:"-"`
	Skills  []ResearchSkill `json:"skills,omitempty"`
	Result  *ResearchReport `json:"result,omitempty"`
}

func (s *TopicResearchService) Skills(ctx context.Context, userID uint) ([]ResearchSkill, error) {
	records, err := s.repo.Skills(ctx, userID)
	if err != nil {
		return nil, err
	}
	defaults := initialResearchSkills()
	defaultSelection := make(map[string]bool, len(defaults))
	for _, skill := range defaults {
		defaultSelection[skill.ID] = skill.DefaultSelected
	}
	seeded := map[string]bool{}
	for _, row := range records {
		if row.SeedKey != nil {
			seeded[*row.SeedKey] = true
		}
	}
	var missing []model.TopicResearchSkill
	for _, skill := range defaults {
		if !seeded[skill.ID] {
			id := skill.ID
			missing = append(missing, model.TopicResearchSkill{UserID: userID, SeedKey: &id, Name: skill.Name, Instructions: skill.Instructions})
		}
	}
	if len(missing) > 0 {
		if err := s.repo.EnsureSkills(ctx, missing); err != nil {
			return nil, err
		}
		records, err = s.repo.Skills(ctx, userID)
		if err != nil {
			return nil, err
		}
	}
	skills := make([]ResearchSkill, 0, len(records))
	for _, record := range records {
		skill := ResearchSkill{
			ID: fmt.Sprintf("custom:%d", record.ID), Name: record.Name, Instructions: record.Instructions, GroupID: record.GroupID, Markdown: record.Markdown,
		}
		if record.SeedKey != nil {
			skill.seedKey = *record.SeedKey
			skill.DefaultSelected = defaultSelection[skill.seedKey]
		}
		skills = append(skills, skill)
	}
	return skills, nil
}

func (s *TopicResearchService) SaveSkill(ctx context.Context, userID uint, req ResearchSkillRequest) (ResearchSkill, error) {
	name, instructions := strings.TrimSpace(req.Name), strings.TrimSpace(req.Instructions)
	if name == "" || instructions == "" || len([]rune(name)) > 80 || len([]rune(instructions)) > 3000 {
		return ResearchSkill{}, errcode.New(400, "Skill 需要名称和正文。名称最多 80 字，正文最多 3000 字。", http.StatusBadRequest)
	}
	record, err := s.repo.SaveSkill(ctx, userID, req.ID, name, instructions, req.Markdown)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ResearchSkill{}, errcode.ErrNotFound
	}
	return ResearchSkill{ID: fmt.Sprintf("custom:%d", record.ID), Name: record.Name, Instructions: record.Instructions, GroupID: record.GroupID, Markdown: record.Markdown}, err
}

func (s *TopicResearchService) OrderSkills(ctx context.Context, userID uint, ids []uint, move *repository.SkillGroupMove) error {
	if len(ids) == 0 || (move != nil && !slices.Contains(ids, move.SkillID)) {
		return errcode.ErrBadRequest
	}
	err := s.repo.OrderSkills(ctx, userID, ids, move)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errcode.New(409, "Skill 列表已变化，请刷新后重新排序。", http.StatusConflict)
	}
	return err
}

func (s *TopicResearchService) Runs(ctx context.Context, userID, beforeID uint) ([]model.TopicResearchRun, error) {
	if err := s.repo.ExpireRuns(ctx, userID, time.Now().Add(-3*time.Minute)); err != nil {
		return nil, err
	}
	return s.repo.Runs(ctx, userID, beforeID, 20)
}

func (s *TopicResearchService) Run(ctx context.Context, userID, id uint) (ResearchRunView, error) {
	record, err := s.repo.Run(ctx, userID, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ResearchRunView{}, errcode.ErrNotFound
	}
	if err != nil {
		return ResearchRunView{}, err
	}
	if record.Status == "running" && time.Since(record.CreatedAt) > 3*time.Minute {
		if err = s.repo.ExpireRuns(ctx, userID, time.Now().Add(-3*time.Minute)); err != nil {
			return ResearchRunView{}, err
		}
		record, err = s.repo.Run(ctx, userID, id)
		if err != nil {
			return ResearchRunView{}, err
		}
	}
	view := ResearchRunView{TopicResearchRun: record}
	if err = json.Unmarshal([]byte(record.SkillText), &view.Skills); err != nil {
		return ResearchRunView{}, err
	}
	if record.ResultText != "" {
		err = json.Unmarshal([]byte(record.ResultText), &view.Result)
	}
	return view, err
}

func (s *TopicResearchService) Start(ctx context.Context, userID uint, req ResearchRunRequest) (model.TopicResearchRun, error) {
	question := strings.TrimSpace(req.Question)
	if question == "" || len([]rune(question)) > 120 || len(req.SkillIDs) > 8 {
		return model.TopicResearchRun{}, errcode.New(400, "请输入 1–120 字的题材问题，最多附带 8 个可选 Skill。", http.StatusBadRequest)
	}
	key := researchAPIKey(s.localKey)

	skills, err := s.Skills(ctx, userID)
	if err != nil {
		return model.TopicResearchRun{}, err
	}
	known := make(map[string]ResearchSkill)
	selected := []ResearchSkill{}
	for _, skill := range skills {
		known[skill.ID] = skill
	}
	for _, id := range req.SkillIDs {
		if skill, ok := known[id]; ok {
			selected = append(selected, skill)
			delete(known, id)
		} else {
			return model.TopicResearchRun{}, errcode.ErrBadRequest
		}
	}
	if key == "" && !(s.Reviews != nil && useMarketReview(question, selected)) {
		return model.TopicResearchRun{}, errResearchKey
	}
	var reference *ResearchRunView
	if req.ReferenceID != 0 {
		previous, err := s.Run(ctx, userID, req.ReferenceID)
		if err != nil {
			return model.TopicResearchRun{}, err
		}
		if previous.Status == "running" {
			return model.TopicResearchRun{}, errResearchBusy
		}
		var history []string
		for turn := previous; ; {
			text := turn.Question
			if turn.Result != nil {
				if reference == nil {
					completed := turn
					reference = &completed
				}
				text += " → " + truncateResearchText(strings.TrimSpace(turn.Result.Reply+" "+turn.Result.Conclusion), 480)
			}
			history = append(history, text)
			if turn.ReferenceID == 0 || turn.ReferenceID >= turn.ID || len(history) >= 16 {
				break
			}
			turn, err = s.Run(ctx, userID, turn.ReferenceID)
			if err != nil {
				return model.TopicResearchRun{}, err
			}
		}
		if reference != nil {
			slices.Reverse(history)
			reference.History = history
		}
	}

	if _, exists := s.active.LoadOrStore(userID, true); exists {
		return model.TopicResearchRun{}, errResearchBusy
	}
	select {
	case s.slots <- struct{}{}:
	default:
		s.active.Delete(userID)
		return model.TopicResearchRun{}, errResearchBusy
	}
	release := func() { <-s.slots; s.active.Delete(userID) }
	snapshot, _ := json.Marshal(selected)
	row := model.TopicResearchRun{UserID: userID, Question: question, Model: researchModel, Status: "running",
		ReferenceID: req.ReferenceID, SkillText: string(snapshot)}
	if err = s.repo.AddRun(ctx, &row); err != nil {
		release()
		return model.TopicResearchRun{}, err
	}
	go func() {
		defer release()
		defer s.phases.Delete(row.ID)
		defer s.replies.Delete(row.ID)
		job, cancel := context.WithTimeout(context.Background(), 130*time.Second)
		defer cancel()
		job = context.WithValue(job, researchRunKey{}, row.ID)
		s.progress(job, "正在检索题材和公开资料")
		report := s.investigate(job, key, question, selected, reference)
		s.progress(job, "正在撰写回答")
		if key != "" {
			report.Reply = s.streamReply(job, row.ID, key, question, selected, reference, report)
		}
		if report.Reply == "" && s.Reviews != nil && useMarketReview(question, selected) {
			report.Reply = report.Conclusion
		}
		data, err := json.Marshal(report)
		save, done := context.WithTimeout(context.Background(), 5*time.Second)
		defer done()
		if err == nil {
			err = s.repo.FinishRun(save, userID, row.ID, "done", string(data), "")
		}
		if err != nil {
			zap.L().Warn("topic research result could not be saved", zap.Uint("run_id", row.ID), zap.Error(err))
		}
	}()
	return row, nil
}

type researchPlan struct{ Topics []string }

func topicCatalog(data MarketDataset) map[string]MarketRow {
	result := map[string]MarketRow{}
	for _, row := range data.Rows {
		if key := normalizeResearchTopic(row.Name); key != "" {
			result[key] = row
		}
	}
	return result
}

func (s *TopicResearchService) investigate(ctx context.Context, key, question string, selected []ResearchSkill, reference *ResearchRunView) ResearchReport {
	focus := question
	if reference != nil {
		focus = strings.Join(reference.Result.Topics, "、") + " " + question
	}
	if s.Reviews != nil && useMarketReview(question, selected) {
		return s.reviewEvidence(ctx, question, reference)
	}
	plans := researchSkillPlans(focus, selected)
	if len(plans) == 0 {
		return s.investigateTopic(ctx, key, question, selected, reference)
	}
	evidence := make(chan researchSkillResult, 1)
	go func() { evidence <- s.checkResearchSkills(ctx, key, focus, plans) }()
	report := s.investigateTopic(ctx, key, question, selected, reference)
	s.progress(ctx, "正在比对研究结论与原文")
	result := <-evidence
	report.SkillChecks = result.checks
	report.Sources = append(report.Sources, result.sources...)
	return report
}

func (s *TopicResearchService) investigateTopic(ctx context.Context, key, question string, selected []ResearchSkill, reference *ResearchRunView) ResearchReport {
	if isOptimusQuestion(question) || reference != nil && reference.Result != nil && reference.Result.Optimus != nil {
		return s.investigateOptimus(ctx, key, question, selected, reference)
	}
	report := ResearchReport{Topics: []string{}, Candidates: []ResearchCandidate{}, Sources: []ResearchSource{}, Notes: []string{}, Match: "all"}
	if strings.Contains(question, "或") || strings.Contains(question, "|") {
		report.Match = "any"
	}
	catalog, _ := s.market.Feed(ctx, "concept", "", "")
	other, _ := s.market.cached(ctx, "research-ths-catalog", 15*time.Minute, s.market.thsCatalog, "")
	if catalog.Status == "unavailable" {
		report.Notes = append(report.Notes, "新浪财经题材目录本次不可用。")
	}
	if other.Status == "unavailable" {
		report.Notes = append(report.Notes, "同花顺公开题材目录本次不可用，双站收录结果可能不完整。")
	}
	sina, ths := topicCatalog(catalog), topicCatalog(other)
	available := map[string]topicSelection{}
	for _, set := range []map[string]MarketRow{sina, ths} {
		for key, row := range set {
			if _, exists := available[key]; !exists {
				available[key] = topicSelection{Name: row.Name, Key: key}
			}
		}
	}

	searchQuestion := question
	if reference != nil {
		searchQuestion = strings.Join(reference.Result.Topics, "、") + " " + question
	}
	topics := researchTopics(question, available)
	if len(topics) == 0 && reference != nil {
		topics = researchTopics(searchQuestion, available)
	}
	var names []string
	if len(topics) == 0 {
		for _, row := range available {
			names = append(names, row.Name)
		}
		sort.Strings(names)
	}
	// 先独立检索披露与专业网站，模型只能通过实际返回的来源 ID 引用它们。
	scopes := [][]string{{"cninfo.com.cn", "sse.com.cn", "szse.cn"}, {"10jqka.com.cn", "eastmoney.com", "wallstreetcn.com"}}
	if strings.Contains(question, "链") || strings.Contains(question, "供货") {
		scopes = append(scopes, []string{"stcn.com", "cls.cn", "cnstock.com"})
	}
	type searchBatch struct {
		index   int
		sources []ResearchSource
		err     error
	}
	searchResults := make(chan searchBatch, len(scopes))
	for index, hosts := range scopes {
		go func() {
			scope, cancel := context.WithTimeout(ctx, 10*time.Second)
			defer cancel()
			var sources []ResearchSource
			var err error
			if index == 0 && usesMinesweeper(selected) {
				sources, err = researchSearchPages(scope, s.client, key, "年度报告全文 审计意见 财务报表 "+searchQuestion, hosts, true)
			} else {
				sources, err = researchSearch(scope, s.client, key, searchQuestion, hosts)
			}
			searchResults <- searchBatch{index, sources, err}
		}()
	}
	if len(topics) == 0 && len(names) > 0 {
		scope, cancel := context.WithTimeout(ctx, 15*time.Second)
		plan, err := researchChat[researchPlan](scope, s.client, key,
			"根据问题从给定目录选择最贴近的题材。保留市场和语义范围，候选名称必须逐字来自目录。不确定则返回空列表。仅输出 JSON：{\"topics\":[\"目录名称\"]}。",
			map[string]any{"question": question, "topicCatalog": names}, 320)
		cancel()
		if err == nil {
			seen := map[string]bool{}
			for _, name := range plan.Topics {
				selection, ok := available[normalizeResearchTopic(name)]
				if ok && !seen[selection.Key] && len(topics) < 4 {
					topics = append(topics, selection)
					seen[selection.Key] = true
				}
			}
		}
	}

	for _, topic := range topics {
		report.Topics = append(report.Topics, topic.Name)
	}
	s.progress(ctx, "正在核对成分股与披露来源")
	boards := s.researchBoards(ctx, topics, sina, ths)
	report.Candidates, report.Sources = compareResearchBoards(boards, report.Topics, report.Match)
	if len(topics) == 0 {
		report.Notes = append(report.Notes, "行情目录没有匹配到题材。搜索和 AI 解读仅供定位来源。")
	}
	if len(report.Candidates) == 0 && len(topics) > 0 {
		report.Notes = append(report.Notes, "已取得的成分股中没有符合当前匹配条件的标的。")
	}
	report.Notes = append(report.Notes, "网站成分股收录只核验板块归类。供应链关系需查看具体公司披露原文。")

	batches := make([]searchBatch, len(scopes))
	for range scopes {
		batch := <-searchResults
		batches[batch.index] = batch
	}
	seenSources := map[string]bool{}
	for _, batch := range batches {
		if batch.err != nil {
			report.Notes = append(report.Notes, "检索暂不可用："+strings.Join(scopes[batch.index], "、"))
		}
		for _, source := range batch.sources {
			if seenSources[source.URL] {
				continue
			}
			seenSources[source.URL] = true
			source.ID = fmt.Sprintf("S%d", len(report.Sources)+1)
			report.Sources = append(report.Sources, source)
		}
	}

	s.progress(ctx, "正在整理证据与结论")
	interpretationInput := map[string]any{
		"question": question, "skills": selected, "topics": report.Topics, "matching": report.Match,
		"candidates": report.Candidates[:min(32, len(report.Candidates))],
		"sources":    report.Sources, "limitations": report.Notes,
	}
	if reference != nil {
		interpretationInput["conversation"] = reference.History
		interpretationInput["selectedPreviousReport"] = map[string]any{
			"id": reference.ID, "date": reference.CreatedAt, "conclusion": reference.Result.Conclusion,
			"candidates": reference.Result.Candidates[:min(32, len(reference.Result.Candidates))],
		}
	}
	scope, cancel := context.WithTimeout(ctx, 28*time.Second)
	explanation, err := researchChat[struct {
		Conclusion   string
		Observations []ResearchObservation
	}](scope, s.client, key,
		researchScopeRule+" 自定义 Skill 只调整研究角度。当前行情目录和检索摘要是待核验资料。结论最多 350 中文字，明确哪些是概念收录，哪些只有市场线索。最多给 8 条个股线索。公司名称或六位代码必须出现在引用的检索摘要或标题中，若代码不确定则略去。涉及供货时区分披露、媒体观察和投资者提问，保持未确认措辞，勿作投资建议。仅输出 JSON：{\"conclusion\":\"...\",\"observations\":[{\"code\":\"六位代码\",\"name\":\"证券简称\",\"note\":\"证据或待核对点\",\"sourceIds\":[\"来源ID\"]}]}。",
		interpretationInput, 1350)
	cancel()
	if err != nil {
		report.Notes = append(report.Notes, "本次未取得可用的 DeepSeek 解读，已保留实际查询到的资料。")
		return report
	}
	report.Conclusion = truncateResearchText(explanation.Conclusion, 800)
	report.Observations = s.verifyResearchObservations(ctx, explanation.Observations, report.Candidates, report.Sources)
	return report
}

var researchCodePart = regexp.MustCompile(`\b\d{6}\b`)

func (s *TopicResearchService) verifyResearchObservations(ctx context.Context, rows []ResearchObservation, candidates []ResearchCandidate, sources []ResearchSource) []ResearchObservation {
	known := map[string]ResearchCandidate{}
	for _, candidate := range candidates {
		known[candidate.Code] = candidate
	}
	sourceByID := map[string]ResearchSource{}
	for _, source := range sources {
		sourceByID[source.ID] = source
	}
	var possible []ResearchObservation
	seen := map[string]bool{}
	for _, row := range rows {
		candidate, ok := known[row.Code]
		if !researchStockCode.MatchString(row.Code) || seen[row.Code] || strings.TrimSpace(row.Note) == "" {
			continue
		}
		name := strings.TrimSpace(row.Name)
		if ok {
			name = candidate.Name
		}
		ids := []string{}
		for _, id := range row.SourceIDs {
			source, ok := sourceByID[id]
			if !ok {
				continue
			}
			mention := len([]rune(name)) >= 2 && strings.Contains(source.Title+" "+source.Excerpt, name)
			for _, code := range researchCodePart.FindAllString(source.Title+" "+source.Excerpt, -1) {
				mention = mention || code == row.Code
			}
			if source.Kind == "search" && mention && !slices.Contains(ids, id) {
				ids = append(ids, id)
			}
		}
		if len(ids) > 0 {
			seen[row.Code] = true
			possible = append(possible, ResearchObservation{Code: row.Code, Name: name, Note: truncateResearchText(row.Note, 260), SourceIDs: ids})
			if len(possible) == 8 {
				break
			}
		}
	}
	scope, stop := context.WithTimeout(ctx, 5*time.Second)
	defer stop()
	validated := make(chan struct {
		index int
		row   ResearchObservation
	}, len(possible))
	for i, row := range possible {
		go func() {
			if _, found := known[row.Code]; !found {
				name, link := s.market.researchQuote(scope, row.Code)
				if name == "" || row.Name != "" && row.Name != name {
					row = ResearchObservation{}
				} else {
					row.Name, row.QuoteURL = name, link
				}
			}
			validated <- struct {
				index int
				row   ResearchObservation
			}{i, row}
		}()
	}
	output := make([]ResearchObservation, len(possible))
	for range possible {
		row := <-validated
		output[row.index] = row.row
	}
	output = slices.DeleteFunc(output, func(row ResearchObservation) bool { return row.Code == "" })
	return output
}

func compareResearchBoards(boards []researchBoard, topics []string, mode string) ([]ResearchCandidate, []ResearchSource) {
	type evidence struct {
		name  string
		all   map[string]bool
		sites map[string]map[string]bool
		ids   []string
	}
	index := map[string]*evidence{}
	sources := []ResearchSource{}
	for _, board := range boards {
		id := "B" + strconv.Itoa(len(sources)+1)
		title := board.Topic
		if board.Actual != "" {
			title = board.Actual
		}
		sources = append(sources, ResearchSource{ID: id, Provider: board.Provider, Title: title, URL: board.URL,
			Kind: "board", Excerpt: fmt.Sprintf("读取 %d 只成分股。 %s", len(board.Rows), board.Note)})
		for _, row := range board.Rows {
			if !researchStockCode.MatchString(row.Code) || strings.HasSuffix(row.Name, "退") {
				continue
			}
			item := index[row.Code]
			if item == nil {
				item = &evidence{row.Name, map[string]bool{}, map[string]map[string]bool{}, nil}
				index[row.Code] = item
			}
			if item.sites[board.Provider] == nil {
				item.sites[board.Provider] = map[string]bool{}
			}
			item.sites[board.Provider][board.Topic] = true
			item.all[board.Topic] = true
			item.ids = append(item.ids, id)
		}
	}

	require := len(topics)
	if mode == "any" && require > 0 {
		require = 1
	}
	candidates := []ResearchCandidate{}
	for code, item := range index {
		if len(item.all) < require {
			continue
		}
		full := 0
		for _, topicSet := range item.sites {
			if len(topicSet) >= require {
				full++
			}
		}
		status := "combined"
		if full >= 2 {
			status = "double"
		} else if full == 1 {
			status = "single"
		}
		matches := []string{}
		for _, topic := range topics {
			if item.all[topic] {
				matches = append(matches, topic)
			}
		}
		candidates = append(candidates, ResearchCandidate{Code: code, Name: item.name, Status: status, Topics: matches, SourceIDs: item.ids})
	}
	order := map[string]int{"double": 0, "combined": 1, "single": 2}
	sort.Slice(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if a.Status != b.Status {
			return order[a.Status] < order[b.Status]
		}
		return a.Code < b.Code
	})
	return candidates, sources
}

type researchRunKey struct{}

func (s *TopicResearchService) progress(ctx context.Context, message string) {
	if id, ok := ctx.Value(researchRunKey{}).(uint); ok {
		s.phases.Store(id, message)
	}
}

func (s *TopicResearchService) Progress(id uint) string {
	if message, ok := s.phases.Load(id); ok {
		return message.(string)
	}
	return ""
}

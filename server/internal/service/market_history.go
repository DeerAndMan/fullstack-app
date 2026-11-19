package service

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const kplHistory = "https://apphis.longhuvip.com/w1/api/index.php"
const kplWebsite = "https://www.kaipanla.com/omain/"

// Verified closing snapshots from ChangeStatistics and HisZhangFuDetail, joined by trading date.
//
//go:embed market_history_archive.json
var historyArchive []byte

type historyDay struct {
	Date      string `json:"date"`
	Up        int    `json:"up"`
	Down      int    `json:"down"`
	AmountWan int64  `json:"amountWan"`
	Score     int    `json:"score"`
}

func (d historyDay) complete() bool {
	return d.Up >= 0 && d.Down >= 0 && d.Up+d.Down > 0 && d.AmountWan > 0 && d.Score >= 0 && d.Score <= 100
}

func historyDetailQuery(date string) url.Values {
	return url.Values{"a": {"HisZhangFuDetail"}, "c": {"HisHomeDingPan"}, "apiv": {"w42"}, "Day": {date}}
}

func (s *MarketService) loadClosingDay(ctx context.Context, date string, score int) (historyDay, error) {
	var reply struct {
		Code json.Number `json:"errcode"`
		Date string
		Info struct {
			Up     json.Number `json:"SZJS"`
			Down   json.Number `json:"XDJS"`
			Amount json.Number `json:"qscln"`
		}
	}
	body, err := s.request(ctx, kplHistory, historyDetailQuery(date))
	if err != nil {
		return historyDay{}, err
	}
	if json.Unmarshal(body, &reply) != nil || reply.Code.String() != "0" || reply.Date != date {
		return historyDay{}, errMarketSession
	}
	up, e1 := reply.Info.Up.Int64()
	down, e2 := reply.Info.Down.Int64()
	amount, e3 := reply.Info.Amount.Int64()
	d := historyDay{Date: date, Up: int(up), Down: int(down), AmountWan: amount, Score: score}
	if e1 != nil || e2 != nil || e3 != nil || d.Up < 0 || d.Down < 0 || d.Up+d.Down == 0 || d.AmountWan <= 0 {
		return historyDay{}, errMarketData
	}
	return d, nil
}

type historyScore struct {
	Day    string
	Strong json.Number
}

// The list is newest first. Index is a row offset, not a page number.
func (s *MarketService) loadHistoryScores(ctx context.Context, start, end string, scores map[string]int) error {
	const size = 250
	var previous string
	for index := 0; ; index += size {
		q := url.Values{"a": {"ChangeStatistics"}, "c": {"HisHomeDingPan"}, "apiv": {"w44"}, "st": {strconv.Itoa(size)}, "Index": {strconv.Itoa(index)}}
		data, err := s.request(ctx, kplHistory, q)
		if err != nil {
			return err
		}
		var page struct {
			Code json.Number `json:"errcode"`
			Info []historyScore
		}
		if json.Unmarshal(data, &page) != nil || page.Code.String() != "0" || page.Info == nil || (index == 0 && len(page.Info) == 0) {
			return errMarketData
		}
		for _, row := range page.Info {
			score, err := row.Strong.Int64()
			if _, e := time.Parse(time.DateOnly, row.Day); e != nil || err != nil || score < 0 || score > 100 || (previous != "" && row.Day >= previous) {
				return errMarketData
			}
			previous = row.Day
			if row.Day >= start && row.Day <= end {
				scores[row.Day] = int(score)
			}
		}
		if start == "" || previous <= start || len(page.Info) < size {
			return nil
		}
	}
}

// Each completed session has its own file so overlapping range requests and restarts are safe.
func historyCachePath() string {
	if path := os.Getenv("MARKET_HISTORY_CACHE_DIR"); path != "" {
		return path
	}
	if dir, err := os.UserCacheDir(); err == nil {
		return filepath.Join(dir, "fullstack-app", "market-history-kpl-v1")
	}
	return ""
}

func saveHistorySnapshot(dir, date string, snapshot any) error {
	if dir == "" {
		return fmt.Errorf("missing cache directory")
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(dir, ".closing-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	err = json.NewEncoder(file).Encode(snapshot)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(file.Name(), filepath.Join(dir, date+".json"))
}

func readHistorySnapshots(dir, start, end string, load func(string, []byte)) {
	if dir == "" {
		return
	}
	files, _ := os.ReadDir(dir)
	for _, f := range files {
		date, ok := strings.CutSuffix(f.Name(), ".json")
		if !ok || date < start || date > end {
			continue
		}
		if _, err := time.Parse(time.DateOnly, date); err != nil {
			continue
		}
		if b, err := os.ReadFile(filepath.Join(dir, f.Name())); err == nil {
			load(date, b)
		}
	}
}

// Accept an archive month, an inclusive date range, or the latest 20 archived days.
func historyPeriod(value string, now time.Time) (start, end string, ok bool) {
	if value == "" {
		return "", now.Format(time.DateOnly), true
	}
	if from, to, rangeQuery := strings.Cut(value, ".."); rangeQuery {
		_, fromErr := time.Parse(time.DateOnly, from)
		_, toErr := time.Parse(time.DateOnly, to)
		return from, to, fromErr == nil && toErr == nil && from <= to && to <= now.Format(time.DateOnly)
	}
	month, err := time.Parse("2006-01", value)
	return value + "-01", month.AddDate(0, 1, -1).Format(time.DateOnly), err == nil && value <= now.Format("2006-01")
}

func (s *MarketService) marketHistory(ctx context.Context, period string) (MarketDataset, error) {
	out := MarketDataset{Source: "开盘啦", SourceURL: kplWebsite}
	start, end, _ := historyPeriod(period, time.Now().In(marketZone))
	var seed []historyDay
	if err := json.Unmarshal(historyArchive, &seed); err != nil {
		return out, err
	}
	saved := make(map[string]historyDay)
	scores := make(map[string]int)
	remember := func(d historyDay) {
		if d.Date >= start && d.Date <= end && d.complete() {
			saved[d.Date] = d
			scores[d.Date] = d.Score
		}
	}
	for _, d := range seed {
		remember(d)
	}
	dir := historyCachePath()
	readHistorySnapshots(dir, start, end, func(date string, b []byte) {
		var day historyDay
		if json.Unmarshal(b, &day) == nil && day.Date == date {
			remember(day)
		}
	})
	var notices []string
	var backups []backupHistoryDay
	if err := json.Unmarshal(historyBackupArchive, &backups); err != nil {
		return out, err
	}
	backupByDate := make(map[string]backupHistoryDay)
	known := make(map[string]bool)
	for _, b := range backups {
		if b.Date >= start && b.Date <= end && b.valid() {
			backupByDate[b.Date] = b
			known[b.Date] = true
		}
	}
	backupDir := ""
	if dir != "" {
		backupDir = filepath.Join(dir, "secondary")
	}
	readHistorySnapshots(backupDir, start, end, func(date string, b []byte) {
		var day backupHistoryDay
		if json.Unmarshal(b, &day) == nil && day.Date == date && day.valid() {
			backupByDate[date] = day
			known[date] = true
		}
	})
	// A slow directory must leave time for cached days and the secondary providers.
	indexCtx, cancel := context.WithTimeout(ctx, 2500*time.Millisecond)
	defer cancel()
	type calendarResult struct {
		dates []string
		err   error
	}
	calendar := make(chan calendarResult, 1)
	go func() { dates, err := s.historyCalendar(indexCtx, start, end); calendar <- calendarResult{dates, err} }()
	indexErr := s.loadHistoryScores(indexCtx, start, end, scores)
	trading := <-calendar
	if indexErr != nil {
		notices = append(notices, "开盘啦历史目录暂未完整刷新，已优先使用开盘啦收盘快照。")
	}
	if trading.err != nil {
		notices = append(notices, "同花顺交易日历暂未完整刷新，按已取得的归档日期展示。")
	}
	for date := range scores {
		known[date] = true
	}
	for _, date := range trading.dates {
		known[date] = true
	}
	if len(known) == 0 && indexErr != nil && trading.err != nil {
		return out, indexErr
	}
	dates := make([]string, 0, len(known))
	for date := range known {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	if period == "" && len(dates) > 20 {
		dates = dates[len(dates)-20:]
	}
	type job struct {
		index int
		date  string
	}
	jobs := make(chan job, len(dates))
	days := make([]historyDay, len(dates))
	writeFailed := make([]bool, len(dates))
	secondary := make([]backupHistoryDay, len(dates))
	for i, date := range dates {
		secondary[i] = backupByDate[date]
		if d, ok := saved[date]; ok && d.Score == scores[date] {
			days[i] = d
			continue
		}
		jobs <- job{i, date}
	}
	close(jobs)
	var wg sync.WaitGroup
	for range min(4, len(jobs)) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := range jobs {
				if ctx.Err() != nil {
					return
				}
				score, hasScore := scores[j.date]
				if !hasScore {
					score = -1
				}
				primaryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
				d, err := s.loadClosingDay(primaryCtx, j.date, score)
				cancel()
				if err == nil {
					days[j.index] = d
					if hasScore {
						writeFailed[j.index] = saveHistorySnapshot(dir, d.Date, d) != nil
					}
				}
				if (err != nil || !hasScore) && !secondary[j.index].valid() {
					fallbackCtx, stop := context.WithTimeout(ctx, 2*time.Second)
					b := s.reportBackup(fallbackCtx, j.date)
					stop()
					if !b.valid() {
						b = s.zhishiBackup(ctx, j.date)
					}
					secondary[j.index] = b
					if b.valid() {
						writeFailed[j.index] = saveHistorySnapshot(backupDir, b.Date, b) != nil || writeFailed[j.index]
					}
				}
			}
		}()
	}
	wg.Wait()
	missing, supplemented := 0, 0
	cacheWarning := false
	for i, date := range dates {
		row := MarketRow{Code: date, Name: date, Date: date}
		b := secondary[i]
		usedBackup := false
		if n, ok := scores[date]; ok {
			score := float64(n)
			row.Sentiment = &score
			row.Detail = "开盘啦情绪指标"
			row.ScoreURL = kplWebsite
		} else if b.valid() {
			row.Sentiment = &b.Score
			row.Detail = b.Source
			row.ScoreURL = b.link()
			usedBackup = true
		}
		d := days[i]
		if d.AmountWan == 0 {
			d = saved[date]
		}
		if d.AmountWan > 0 {
			amount := float64(d.AmountWan) * 10000
			row.Advancers = &d.Up
			row.Decliners = &d.Down
			row.Amount = &amount
			row.QuoteSource = "开盘啦"
			row.URL = kplHistory + "?" + historyDetailQuery(date).Encode()
		} else if b.valid() {
			amount := math.Round(b.AmountYi * 1e8)
			row.Advancers = &b.Up
			row.Decliners = &b.Down
			row.Amount = &amount
			row.QuoteSource = b.Source
			row.URL = b.link()
			usedBackup = true
		}
		if row.Sentiment == nil || row.Amount == nil {
			missing++
		}
		if usedBackup {
			supplemented++
		}
		cacheWarning = cacheWarning || writeFailed[i]
		out.Rows = append(out.Rows, row)
	}
	if supplemented > 0 {
		notices = append(notices, fmt.Sprintf("%d 个交易日使用了备用来源，详见逐日明细。", supplemented))
	}
	if missing > 0 {
		notices = append(notices, fmt.Sprintf("%d 个交易日存在缺项，空缺保留并重试。", missing))
	}
	if cacheWarning {
		notices = append(notices, "本次取得的数据未能保存到本地缓存。")
	}
	out.Message = strings.Join(notices, " ")
	return out, nil
}

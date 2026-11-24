package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"regexp"
	"time"

	"github.com/injoyai/ios"
	"github.com/injoyai/tdx"
	"github.com/injoyai/tdx/protocol"
)

type styleBar struct {
	Date  string  `json:"date"`
	Close float64 `json:"close"`
}

func dialQuote(host string) (*tdx.Client, error) {
	dial := func(ctx context.Context) (ios.ReadWriteCloser, string, error) {
		d := net.Dialer{Timeout: 4 * time.Second}
		conn, err := d.DialContext(ctx, "tcp", host)
		return conn, host, err
	}
	return tdx.DialWith(dial, tdx.WithDebug(false))
}

var quoteHosts = []string{"116.205.183.150:7709", "116.205.171.132:7709", "111.230.186.52:7709"}

// Read a dated window from a single quote server so both indices use the same feed.
func indexWindows(host, date string) (map[string][]styleBar, error) {
	client, err := dialQuote(host)
	if err != nil {
		return nil, err
	}
	defer client.Close()
	out := make(map[string][]styleBar)
	for _, code := range []string{"880003", "880865"} {
		const page = 500
		var bars []styleBar
		for offset := 0; offset < 32000 && len(bars) < 250; offset += page {
			response, err := client.GetIndexDay("sh"+code, uint16(offset), page)
			if err != nil {
				return nil, err
			}
			part := make([]styleBar, 0, len(response.List))
			for _, bar := range response.List {
				day := bar.Time.Format(time.DateOnly)
				if day <= date {
					part = append(part, styleBar{day, bar.Close.Float64()})
				}
			}
			bars = append(part, bars...)
			if len(response.List) < page {
				break
			}
		}
		if len(bars) < 250 || bars[len(bars)-1].Date != date {
			return nil, fmt.Errorf("%s 缺少 %s 收盘或历史日线", code, date)
		}
		out[code] = bars[len(bars)-250:]
	}
	return out, nil
}

func writeStyleBars(date string) error {
	if _, err := time.Parse(time.DateOnly, date); err != nil {
		return err
	}
	var err error
	for _, host := range quoteHosts {
		var bars map[string][]styleBar
		bars, err = indexWindows(host, date)
		if err == nil {
			return json.NewEncoder(os.Stdout).Encode(bars)
		}
	}
	return err
}

type equityBar struct {
	Date  string  `json:"date"`
	High  float64 `json:"high"`
	Close float64 `json:"close"`
}

type equityWindow struct {
	Bars    []equityBar      `json:"bars"`
	Actions []*protocol.XRXD `json:"actions"`
}

func beijingWindow(client *tdx.Client, symbol, date string) (equityWindow, error) {
	result := equityWindow{}
	code := "bj" + symbol[:6]
	const page = 500
	for offset := 0; offset < 32000 && len(result.Bars) < 250; offset += page {
		resp, err := client.GetKlineDay(code, uint16(offset), page)
		if err != nil {
			return result, err
		}
		var bars []equityBar
		for _, b := range resp.List {
			stamp := b.Time.Format(time.DateOnly)
			if stamp <= date && b.Amount > 0 {
				bars = append(bars, equityBar{stamp, b.High.Float64(), b.Close.Float64()})
			}
		}
		result.Bars = append(bars, result.Bars...)
		if len(resp.List) < page {
			break
		}
	}
	if len(result.Bars) == 0 || result.Bars[len(result.Bars)-1].Date != date {
		return result, fmt.Errorf("%s 缺少 %s 日线", symbol, date)
	}
	if len(result.Bars) < 250 {
		return result, nil
	}
	result.Bars = result.Bars[len(result.Bars)-250:]
	events, err := client.GetGbbq(code)
	if err != nil {
		return result, err
	}
	for _, ev := range events.List {
		stamp := ev.Time.Format(time.DateOnly)
		if ev.IsXRXD() && stamp > result.Bars[0].Date && stamp <= date {
			if ev.Code != code {
				return result, fmt.Errorf("%s 除权代码不符", symbol)
			}
			result.Actions = append(result.Actions, ev.XRXD())
		}
	}
	return result, nil
}

func writeBeijingBars(date string, symbols []string) error {
	if _, err := time.Parse(time.DateOnly, date); err != nil {
		return err
	}
	valid := regexp.MustCompile(`^[0-9]{6}\.BJ$`)
	for _, s := range symbols {
		if !valid.MatchString(s) {
			return fmt.Errorf("无效北交所代码 %q", s)
		}
	}
	windows := make(map[string]equityWindow)
	var lastErr error
	for _, host := range quoteHosts {
		client, err := dialQuote(host)
		if err != nil {
			lastErr = err
			continue
		}
		for _, s := range symbols {
			if _, ok := windows[s]; ok {
				continue
			}
			w, err := beijingWindow(client, s, date)
			if err != nil {
				lastErr = fmt.Errorf("%s: %w", s, err)
				break
			}
			windows[s] = w
		}
		client.Close()
		if len(windows) == len(symbols) {
			return json.NewEncoder(os.Stdout).Encode(windows)
		}
	}
	return lastErr
}

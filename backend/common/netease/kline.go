package netease

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	em "stock_assistant/backend/common/eastmoney"
)

func GetDailyKline(ctx context.Context, code string, limit int) ([]*em.KlineItem, error) {
	if limit <= 0 {
		limit = 5
	}
	ncode := toNeteaseCode(code)
	end := time.Now().Format("20060102")
	start := time.Now().AddDate(0, 0, -limit*3).Format("20060102")
	url := fmt.Sprintf("http://quotes.money.163.com/service/chddata.html?code=%s&start=%s&end=%s", ncode, start, end)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	r := csv.NewReader(resp.Body)
	r.FieldsPerRecord = -1
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	if len(records) <= 1 {
		return nil, fmt.Errorf("no csv data")
	}
	var items []*em.KlineItem
	for i := 1; i < len(records); i++ {
		row := records[i]
		if len(row) < 8 {
			continue
		}
		date := strings.TrimSpace(row[0])
		open, _ := strconv.ParseFloat(row[6], 64)
		high, _ := strconv.ParseFloat(row[4], 64)
		low, _ := strconv.ParseFloat(row[5], 64)
		closep, _ := strconv.ParseFloat(row[3], 64)
		vol, _ := strconv.ParseInt(strings.ReplaceAll(row[11], ",", ""), 10, 64)
		changePct := 0.0
		items = append(items, &em.KlineItem{
			Date:          date,
			Open:          open,
			Close:         closep,
			High:          high,
			Low:           low,
			Volume:        vol,
			ChangePercent: changePct,
		})
	}
	if len(items) > limit {
		items = items[len(items)-limit:]
	}
	return items, nil
}

func toNeteaseCode(code string) string {
	c := strings.ToUpper(strings.TrimSpace(code))
	if strings.HasPrefix(c, "SH") {
		return "0" + c[2:]
	}
	if strings.HasPrefix(c, "SZ") {
		return "1" + c[2:]
	}
	if len(c) == 6 && c[0] == '6' {
		return "0" + c
	}
	return "1" + c
}

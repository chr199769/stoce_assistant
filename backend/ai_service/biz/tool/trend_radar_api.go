package tool

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// --- Toutiao API ---

type ToutiaoHotResponse struct {
	Data []struct {
		Title string `json:"Title"`
		Url   string `json:"Url"` // Sometimes empty, need to construct
	} `json:"data"`
}

func GetToutiaoHotTrends() ([]string, error) {
	url := "https://www.toutiao.com/hot-event/hot-board/?origin=toutiao_pc"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res ToutiaoHotResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	var trends []string
	for i, item := range res.Data {
		if i >= 20 { // Top 20
			break
		}
		trends = append(trends, fmt.Sprintf("%d. %s", i+1, item.Title))
	}
	return trends, nil
}

// --- Baidu API (HTML Scraping) ---

func GetBaiduHotTrends() ([]string, error) {
	url := "https://top.baidu.com/board?tab=realtime"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	content := string(body)
	// Regex to find titles in <div class="c-single-text-ellipsis"> ... </div>
	// This is brittle but works for now.
	re := regexp.MustCompile(`<div class="c-single-text-ellipsis">\s*(.*?)\s*</div>`)
	matches := re.FindAllStringSubmatch(content, -1)

	var trends []string
	count := 0
	for _, match := range matches {
		if len(match) > 1 {
			title := strings.TrimSpace(match[1])
			if title != "" {
				count++
				trends = append(trends, fmt.Sprintf("%d. %s", count, title))
				if count >= 20 {
					break
				}
			}
		}
	}

	if len(trends) == 0 {
		return nil, fmt.Errorf("no trends found (parsing likely failed)")
	}

	return trends, nil
}

// --- WallstreetCN API ---

type WallstreetCNHotResponse struct {
	Data struct {
		Items []struct {
			Title string `json:"title"`
			Uri   string `json:"uri"`
		} `json:"items"`
	} `json:"data"`
}

func GetWallstreetCNHotTrends(period string) ([]string, error) {
	if period == "" {
		period = "24h" // default
	}
	url := fmt.Sprintf("https://api-one-wscn.wallstreetcn.com/apiv1/content/articles/hot?period=%s", period)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res WallstreetCNHotResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	var trends []string
	for i, item := range res.Data.Items {
		if i >= 20 {
			break
		}
		trends = append(trends, fmt.Sprintf("%d. %s", i+1, item.Title))
	}
	return trends, nil
}

// --- The Paper API ---

type ThePaperHotResponse struct {
	Data struct {
		HotNews []struct {
			Name string `json:"name"`
		} `json:"hotNews"`
	} `json:"data"`
}

func GetThePaperHotTrends() ([]string, error) {
	url := "https://cache.thepaper.cn/contentapi/wwwIndex/rightSidebar"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res ThePaperHotResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	var trends []string
	for i, item := range res.Data.HotNews {
		if i >= 20 {
			break
		}
		trends = append(trends, fmt.Sprintf("%d. %s", i+1, item.Name))
	}
	return trends, nil
}

// --- Zhihu API ---

type ZhihuHotResponse struct {
	Data []struct {
		Target struct {
			Title string `json:"title"`
		} `json:"target"`
	} `json:"data"`
}

func GetZhihuHotTrends() ([]string, error) {
	url := "https://api.zhihu.com/topstory/hot-list"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res ZhihuHotResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	var trends []string
	for i, item := range res.Data {
		if i >= 20 {
			break
		}
		trends = append(trends, fmt.Sprintf("%d. %s", i+1, item.Target.Title))
	}
	return trends, nil
}

// --- Cailian Press API ---

type CLSTelegraphResponse struct {
	Data struct {
		RollData []struct {
			Title   string `json:"title"`
			Content string `json:"content"`
			Ctime   int64  `json:"ctime"`
		} `json:"roll_data"`
	} `json:"data"`
}

func GetCailianPressTelegraph() ([]string, error) {
	// Use nodeapi/telegraphList?rn=20
	url := "https://www.cls.cn/nodeapi/telegraphList?rn=20"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var res CLSTelegraphResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}

	var trends []string
	for i, item := range res.Data.RollData {
		if i >= 20 {
			break
		}
		title := item.Title
		if title == "" {
			// Some telegraphs have no title, use content snippet
			title = item.Content
			if len(title) > 50 {
				title = title[:50] + "..."
			}
		}
		t := time.Unix(item.Ctime, 0)
		trends = append(trends, fmt.Sprintf("[%s] %s", t.Format("15:04"), title))
	}
	return trends, nil
}

// --- Aggregator ---

func GetAllTrends() (string, error) {
	var sb strings.Builder

	// Toutiao
	sb.WriteString("=== Toutiao Hot Trends ===\n")
	ttTrends, err := GetToutiaoHotTrends()
	if err != nil {
		sb.WriteString(fmt.Sprintf("Error: %v\n", err))
	} else {
		sb.WriteString(strings.Join(ttTrends, "\n"))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	// Baidu
	sb.WriteString("=== Baidu Hot Trends ===\n")
	bdTrends, err := GetBaiduHotTrends()
	if err != nil {
		sb.WriteString(fmt.Sprintf("Error: %v\n", err))
	} else {
		sb.WriteString(strings.Join(bdTrends, "\n"))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	// WallstreetCN
	sb.WriteString("=== WallstreetCN Hot Trends ===\n")
	wscnTrends, err := GetWallstreetCNHotTrends("24h")
	if err != nil {
		sb.WriteString(fmt.Sprintf("Error: %v\n", err))
	} else {
		sb.WriteString(strings.Join(wscnTrends, "\n"))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	// The Paper
	sb.WriteString("=== The Paper Hot Trends ===\n")
	paperTrends, err := GetThePaperHotTrends()
	if err != nil {
		sb.WriteString(fmt.Sprintf("Error: %v\n", err))
	} else {
		sb.WriteString(strings.Join(paperTrends, "\n"))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	// Zhihu
	sb.WriteString("=== Zhihu Hot Trends ===\n")
	zhTrends, err := GetZhihuHotTrends()
	if err != nil {
		sb.WriteString(fmt.Sprintf("Error: %v\n", err))
	} else {
		sb.WriteString(strings.Join(zhTrends, "\n"))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")

	// Cailian Press
	sb.WriteString("=== Cailian Press Telegraph ===\n")
	clsTrends, err := GetCailianPressTelegraph()
	if err != nil {
		sb.WriteString(fmt.Sprintf("Error: %v\n", err))
	} else {
		sb.WriteString(strings.Join(clsTrends, "\n"))
		sb.WriteString("\n")
	}

	return sb.String(), nil
}

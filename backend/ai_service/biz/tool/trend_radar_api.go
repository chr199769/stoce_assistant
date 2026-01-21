package tool

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
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
	count := 0
	for _, item := range res.Data {
		if isRelevant(item.Title) {
			count++
			trends = append(trends, fmt.Sprintf("%d. %s", count, item.Title))
			if count >= 20 {
				break
			}
		}
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
			if title != "" && isRelevant(title) {
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
		DayItems []struct {
			Title string `json:"title"`
			Uri   string `json:"uri"`
		} `json:"day_items"`
		WeekItems []struct {
			Title string `json:"title"`
			Uri   string `json:"uri"`
		} `json:"week_items"`
	} `json:"data"`
}

func GetWallstreetCNHotTrends(period string) ([]string, error) {
	if period == "" {
		period = "day" // default
	}
	url := fmt.Sprintf("https://api.wallstreetcn.com/apiv1/content/articles/hot?period=%s", period)

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
	items := res.Data.DayItems
	if len(items) == 0 {
		items = res.Data.WeekItems
	}
	for i, item := range items {
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
	count := 0
	for _, item := range res.Data.HotNews {
		if isRelevant(item.Name) {
			count++
			trends = append(trends, fmt.Sprintf("%d. %s", count, item.Name))
			if count >= 20 {
				break
			}
		}
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
	count := 0
	for _, item := range res.Data {
		if isRelevant(item.Target.Title) {
			count++
			trends = append(trends, fmt.Sprintf("%d. %s", count, item.Target.Title))
			if count >= 20 {
				break
			}
		}
	}
	return trends, nil
}

// --- Cailian Press API ---

type CLSTelegraphResponse struct {
	Data struct {
		RollData []struct {
			Title      string `json:"title"`
			Content    string `json:"content"`
			Ctime      int64  `json:"ctime"`
			ReadingNum int    `json:"reading_num"`
		} `json:"roll_data"`
	} `json:"data"`
}

func GetCailianPressTelegraph() ([]string, error) {
	// Use nodeapi/telegraphList?rn=100 to get a larger pool for sorting
	url := "https://www.cls.cn/nodeapi/telegraphList?rn=100"

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

	// Sort by ReadingNum desc
	items := res.Data.RollData
	sort.Slice(items, func(i, j int) bool {
		return items[i].ReadingNum > items[j].ReadingNum
	})

	var trends []string
	for i, item := range items {
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
		trends = append(trends, fmt.Sprintf("[%s] %s (Heat: %d)", t.Format("15:04"), title, item.ReadingNum))
	}
	return trends, nil
}

// --- Filter ---

func isRelevant(title string) bool {
	// Basic keywords for finance, economy, policy, tech industry
	keywords := []string{
		"经济", "金融", "股市", "股票", "A股", "美股", "港股", "证券", "基金",
		"银行", "央行", "货币", "汇率", "外汇", "期货", "大宗", "黄金", "原油",
		"能源", "芯片", "半导体", "AI", "人工智能", "科技", "互联网",
		"房地产", "楼市", "车企", "新能源", "汽车", "消费", "贸易", "出口", "进口",
		"GDP", "CPI", "PPI", "PMI",
		"政策", "会议", "改革", "发展", "规划", "监管",
		"IPO", "上市", "财报", "营收", "利润", "亏损",
		"裁员", "招聘", "就业", "失业", "通胀", "紧缩", "利率", "降息", "加息",
		"公司", "企业", "产业", "市场", "投资", "融资", "收购", "并购",
		"美元", "人民币", "欧元", "日元",
		"拜登", "特朗普", "普京", "欧盟", "北约", // Geopolitics
		"战争", "冲突", "制裁",
	}

	// Blocklist to filter out noise
	blocklist := []string{
		"明星", "绯闻", "出轨", "离婚", "电视剧", "综艺", "网红", "穿搭", "减肥",
		"星座", "八卦", "吃瓜", "搞笑", "段子", "宠物", "猫", "狗",
		"杀人", "砍人", "跳楼", "强奸", "猥亵", "家暴", "吵架", "斗殴",
	}

	for _, block := range blocklist {
		if strings.Contains(title, block) {
			return false
		}
	}

	for _, kw := range keywords {
		if strings.Contains(title, kw) {
			return true
		}
	}

	return false
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
	wscnTrends, err := GetWallstreetCNHotTrends("day")
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

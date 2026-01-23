package crawler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

type NewsItem struct {
	Title   string
	Content string
	Source  string
	Url     string
	Time    string
}

// --- Toutiao API ---

type ToutiaoHotResponse struct {
	Data []struct {
		Title string `json:"Title"`
		Url   string `json:"Url"` // Sometimes empty, need to construct
	} `json:"data"`
}

func GetToutiaoHotTrends() ([]*NewsItem, error) {
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

	var trends []*NewsItem
	count := 0
	for _, item := range res.Data {
		if isRelevant(item.Title) {
			count++
			trends = append(trends, &NewsItem{
				Title:   item.Title,
				Content: item.Title,
				Source:  "今日头条",
				Url:     item.Url,
				Time:    time.Now().Format("2006-01-02 15:04:05"),
			})
			if count >= 20 {
				break
			}
		}
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

func GetWallstreetCNHotTrends(period string) ([]*NewsItem, error) {
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

	var trends []*NewsItem
	items := res.Data.DayItems
	if len(items) == 0 {
		items = res.Data.WeekItems
	}
	for i, item := range items {
		if i >= 20 {
			break
		}
		trends = append(trends, &NewsItem{
			Title:   item.Title,
			Content: item.Title,
			Source:  "华尔街见闻",
			Url:     item.Uri,
			Time:    time.Now().Format("2006-01-02 15:04:05"),
		})
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

func GetThePaperHotTrends() ([]*NewsItem, error) {
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

	var trends []*NewsItem
	count := 0
	for _, item := range res.Data.HotNews {
		if isRelevant(item.Name) {
			count++
			trends = append(trends, &NewsItem{
				Title:   item.Name,
				Content: item.Name,
				Source:  "澎湃新闻",
				Url:     "",
				Time:    time.Now().Format("2006-01-02 15:04:05"),
			})
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
			Url   string `json:"url"`
		} `json:"target"`
	} `json:"data"`
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

func GetCailianPressTelegraph() ([]*NewsItem, error) {
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

	var trends []*NewsItem
	for i, item := range items {
		if i >= 20 {
			break
		}
		title := item.Title
		if title == "" {
			// Some telegraphs have no title, use content snippet
			title = item.Content
			// Fix: Use runes to avoid cutting multi-byte characters
			runes := []rune(title)
			if len(runes) > 50 {
				title = string(runes[:50]) + "..."
			}
		}
		t := time.Unix(item.Ctime, 0)
		trends = append(trends, &NewsItem{
			Title:   title,
			Content: item.Content,
			Source:  "财联社电报",
			Url:     "", // 财联社电报 URL 构造较复杂，暂时留空
			Time:    t.Format("2006-01-02 15:04:05"),
		})
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

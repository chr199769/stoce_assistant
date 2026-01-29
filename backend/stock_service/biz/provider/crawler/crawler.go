package crawler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"stock_assistant/backend/stock_service/config"
)

type NewsItem struct {
	Title   string
	Content string
	Source  string
	Url     string
	Time    string
}

type MacroItem struct {
	Indicator string
	Label     string
	Value     string
	Unit      string
	Time      string
	Source    string
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

func GetAkShareMacroData() ([]*MacroItem, error) {
	cfg := config.Get()
	if cfg != nil && cfg.AkShare != nil && cfg.AkShare.ServiceURL != "" {
		url := cfg.AkShare.ServiceURL
		funcs := "macro_china_cpi,macro_china_ppi,macro_china_pmi,macro_china_gdp,macro_china_fx_reserves,macro_china_money_supply"
		limit := 12
		symbol := ""
		if cfg.AkShare.Macro != nil {
			if len(cfg.AkShare.Macro.Funcs) > 0 {
				funcs = strings.Join(cfg.AkShare.Macro.Funcs, ",")
			}
			if cfg.AkShare.Macro.Limit > 0 {
				limit = cfg.AkShare.Macro.Limit
			}
			symbol = strings.TrimSpace(cfg.AkShare.Macro.Symbol)
		}
		reqURL := fmt.Sprintf("%s/macro?funcs=%s&limit=%d&symbol=%s", strings.TrimRight(url, "/"), funcs, limit, symbol)
		resp, err := http.Get(reqURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			var items []*MacroItem
			if json.Unmarshal(body, &items) == nil {
				return items, nil
			}
		}
	}
	pythonBin := "python3"
	if cfg != nil && cfg.AkShare != nil && cfg.AkShare.PythonBin != "" {
		pythonBin = cfg.AkShare.PythonBin
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	pythonCode := buildAkShareMacroScript(cfg)
	cmd := exec.CommandContext(ctx, pythonBin, "-c", pythonCode)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var items []*MacroItem
	if err := json.Unmarshal(output, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func buildAkShareMacroScript(cfg *config.Config) string {
	funcs := []string{"macro_china_cpi", "macro_china_ppi", "macro_china_pmi", "macro_china_gdp", "macro_china_fx_reserves", "macro_china_money_supply"}
	limit := 12
	symbol := ""
	if cfg != nil && cfg.AkShare != nil && cfg.AkShare.Macro != nil {
		if len(cfg.AkShare.Macro.Funcs) > 0 {
			funcs = cfg.AkShare.Macro.Funcs
		}
		if cfg.AkShare.Macro.Limit > 0 {
			limit = cfg.AkShare.Macro.Limit
		}
		symbol = strings.TrimSpace(cfg.AkShare.Macro.Symbol)
	}
	funcsJSON, _ := json.Marshal(funcs)
	script := strings.Join([]string{
		"import json",
		"import akshare as ak",
		"def pick(row, keys):",
		"    for key in keys:",
		"        if key in row:",
		"            v = row[key]",
		"            if v is None:",
		"                continue",
		"            s = str(v).strip()",
		"            if s != \"\":",
		"                return s",
		"    return \"\"",
		"def pick_time(row):",
		"    return pick(row, [\"日期\",\"时间\",\"date\",\"time\",\"report_date\",\"reportDate\",\"指标日期\"])",
		"def pick_label(row):",
		"    return pick(row, [\"指标\",\"指标名称\",\"name\",\"title\",\"名称\"])",
		"def pick_unit(row):",
		"    return pick(row, [\"单位\",\"unit\"])",
		"def pick_value(row):",
		"    for key in [\"值\",\"value\",\"现值\",\"指标值\",\"同比\",\"环比\"]:",
		"        if key in row and row[key] is not None:",
		"            return str(row[key]).strip()",
		"    for key, val in row.items():",
		"        if isinstance(val, (int, float)) and key not in [\"year\",\"month\",\"day\"]:",
		"            return str(val)",
		"    return \"\"",
		fmt.Sprintf("funcs = %s", funcsJSON),
		fmt.Sprintf("limit = %d", limit),
		fmt.Sprintf("symbol = %q", symbol),
		"items = []",
		"for name in funcs:",
		"    func = getattr(ak, name, None)",
		"    if func is None:",
		"        continue",
		"    try:",
		"        if symbol:",
		"            df = func(symbol=symbol)",
		"        else:",
		"            df = func()",
		"    except Exception:",
		"        continue",
		"    if df is None:",
		"        continue",
		"    if hasattr(df, \"empty\") and df.empty:",
		"        continue",
		"    records = df.to_dict(orient=\"records\")",
		"    tail = records[-limit:] if limit > 0 else records",
		"    for row in tail:",
		"        t = pick_time(row)",
		"        v = pick_value(row)",
		"        if t == \"\" and v == \"\":",
		"            continue",
		"        label = pick_label(row)",
		"        unit = pick_unit(row)",
		"        items.append({",
		"            \"indicator\": name,",
		"            \"label\": label,",
		"            \"value\": v,",
		"            \"unit\": unit,",
		"            \"time\": t,",
		"            \"source\": \"AkShare\",",
		"        })",
		"print(json.dumps(items, ensure_ascii=False))",
	}, "\n")
	return script
}

func GetAkShareNews() ([]*NewsItem, error) {
	return getAkShareNewsFromPython()
}

func getAkShareNewsFromPython() ([]*NewsItem, error) {
	cfg := config.Get()
	if cfg != nil && cfg.AkShare != nil && cfg.AkShare.ServiceURL != "" {
		url := cfg.AkShare.ServiceURL
		funcName := ""
		limit := 50
		symbol := ""
		if cfg.AkShare.News != nil {
			funcName = strings.TrimSpace(cfg.AkShare.News.Func)
			symbol = strings.TrimSpace(cfg.AkShare.News.Symbol)
			if cfg.AkShare.News.Limit > 0 {
				limit = cfg.AkShare.News.Limit
			}
		}
		reqURL := fmt.Sprintf("%s/news?func=%s&symbol=%s&limit=%d", strings.TrimRight(url, "/"), funcName, symbol, limit)
		resp, err := http.Get(reqURL)
		if err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			var items []struct {
				Title   string `json:"title"`
				Content string `json:"content"`
				Source  string `json:"source"`
				Url     string `json:"url"`
				Time    string `json:"time"`
			}
			if json.Unmarshal(body, &items) == nil {
				if len(items) == 0 {
					return []*NewsItem{}, nil
				}
				news := make([]*NewsItem, 0, len(items))
				for _, item := range items {
					title := strings.TrimSpace(item.Title)
					if title == "" {
						continue
					}
					content := strings.TrimSpace(item.Content)
					if content == "" {
						content = title
					}
					source := strings.TrimSpace(item.Source)
					if source == "" {
						source = "AkShare"
					}
					news = append(news, &NewsItem{
						Title:   title,
						Content: content,
						Source:  source,
						Url:     strings.TrimSpace(item.Url),
						Time:    strings.TrimSpace(item.Time),
					})
				}
				return news, nil
			}
		}
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	scriptPath := filepath.Join(cwd, "biz", "provider", "crawler", "akshare_news.py")
	pythonBin := "python3"
	if cfg != nil && cfg.AkShare != nil && cfg.AkShare.PythonBin != "" {
		pythonBin = cfg.AkShare.PythonBin
	}
	configPath := config.GetConfigPath()
	if configPath == "" {
		configPath = "conf/prod.json"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, pythonBin, scriptPath, configPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var items []struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Source  string `json:"source"`
		Url     string `json:"url"`
		Time    string `json:"time"`
	}
	if err := json.Unmarshal(output, &items); err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return []*NewsItem{}, nil
	}
	news := make([]*NewsItem, 0, len(items))
	for _, item := range items {
		title := strings.TrimSpace(item.Title)
		if title == "" {
			continue
		}
		content := strings.TrimSpace(item.Content)
		if content == "" {
			content = title
		}
		source := strings.TrimSpace(item.Source)
		if source == "" {
			source = "AkShare"
		}
		news = append(news, &NewsItem{
			Title:   title,
			Content: content,
			Source:  source,
			Url:     strings.TrimSpace(item.Url),
			Time:    strings.TrimSpace(item.Time),
		})
	}
	return news, nil
}

func extractExternalItems(raw interface{}) []map[string]interface{} {
	if raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case []interface{}:
		return castItemSlice(v)
	case map[string]interface{}:
		if data, ok := v["data"]; ok {
			if list, ok := data.([]interface{}); ok {
				return castItemSlice(list)
			}
		}
		if list, ok := v["items"].([]interface{}); ok {
			return castItemSlice(list)
		}
	}
	return nil
}

func castItemSlice(list []interface{}) []map[string]interface{} {
	items := make([]map[string]interface{}, 0, len(list))
	for _, item := range list {
		if m, ok := item.(map[string]interface{}); ok {
			items = append(items, m)
		}
	}
	return items
}

func pickString(item map[string]interface{}, keys []string) string {
	for _, key := range keys {
		if value, ok := item[key]; ok {
			switch v := value.(type) {
			case string:
				if strings.TrimSpace(v) != "" {
					return strings.TrimSpace(v)
				}
			case float64:
				if v > 0 && (strings.Contains(strings.ToLower(key), "time") || strings.Contains(strings.ToLower(key), "timestamp")) {
					return time.Unix(int64(v), 0).Format("2006-01-02 15:04:05")
				}
			}
		}
	}
	return ""
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

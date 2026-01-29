import json
import sys

import akshare as ak


def pick(row, keys):
    for key in keys:
        if key in row:
            value = str(row[key]).strip()
            if value:
                return value
    return ""


def normalize_items(df, limit):
    records = df.to_dict(orient="records")
    items = []
    for row in records:
        title = pick(row, ["title", "标题", "新闻标题", "headline", "name"])
        if not title:
            continue
        content = pick(row, ["content", "内容", "正文", "摘要", "summary", "detail"])
        if not content:
            content = title
        url = pick(row, ["url", "链接", "网址", "source_url", "news_url"])
        source = pick(row, ["source", "来源", "文章来源", "媒体"])
        time_value = pick(row, ["time", "时间", "发布时间", "日期", "pub_time"])
        items.append(
            {
                "title": title,
                "content": content,
                "url": url,
                "source": source,
                "time": time_value,
            }
        )
        if limit > 0 and len(items) >= limit:
            break
    return items


def load_config():
    config_path = "conf/prod.json"
    if len(sys.argv) > 1:
        config_path = sys.argv[1]
    try:
        with open(config_path, "r", encoding="utf-8") as f:
            raw = json.load(f)
            return raw.get("akshare", {})
    except Exception:
        return {}


def fetch_news():
    cfg = load_config()
    news_cfg = cfg.get("news", {})
    func_name = str(news_cfg.get("func", "")).strip()
    symbol = str(news_cfg.get("symbol", "")).strip()
    limit = int(news_cfg.get("limit", 50))
    candidates = []
    if func_name:
        candidates.append(func_name)
    for name in ["stock_news_em", "news_cctv"]:
        if name not in candidates:
            candidates.append(name)

    for name in candidates:
        func = getattr(ak, name, None)
        if func is None:
            continue
        try:
            if name == "stock_news_em":
                if not symbol:
                    symbol = "A股"
                df = func(symbol=symbol)
            else:
                df = func()
        except Exception:
            continue
        if df is None:
            continue
        if hasattr(df, "empty") and df.empty:
            continue
        return normalize_items(df, limit)
    return []


if __name__ == "__main__":
    items = fetch_news()
    print(json.dumps(items, ensure_ascii=False))

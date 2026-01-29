import sys
import json
import akshare as ak

def pick(row, keys):
    for k in keys:
        if k in row:
            v = row[k]
            if v is None:
                continue
            s = str(v).strip()
            if s:
                return s
    return ""

def pick_time(row):
    return pick(row, ["日期","时间","date","time","report_date","reportDate","指标日期"])

def pick_label(row):
    return pick(row, ["指标","指标名称","name","title","名称"])

def pick_unit(row):
    return pick(row, ["单位","unit"])

def pick_value(row):
    for key in ["值","value","现值","指标值","同比","环比"]:
        if key in row and row[key] is not None:
            return str(row[key]).strip()
    for key, val in row.items():
        if isinstance(val, (int, float)) and key not in ["year","month","day"]:
            return str(val)
    return ""

def main():
    funcs_arg = sys.argv[1] if len(sys.argv) > 1 else ""
    limit_arg = sys.argv[2] if len(sys.argv) > 2 else "12"
    symbol_arg = sys.argv[3] if len(sys.argv) > 3 else ""
    funcs = [s.strip() for s in funcs_arg.split(",") if s.strip()]
    if not funcs:
        funcs = ["macro_china_cpi","macro_china_ppi","macro_china_pmi","macro_china_gdp","macro_china_fx_reserves","macro_china_money_supply"]
    try:
        limit = int(limit_arg)
    except:
        limit = 12
    symbol = symbol_arg.strip()
    items = []
    for name in funcs:
        func = getattr(ak, name, None)
        if func is None:
            continue
        try:
            if symbol:
                df = func(symbol=symbol)
            else:
                df = func()
        except Exception:
            continue
        if df is None:
            continue
        if hasattr(df, "empty") and df.empty:
            continue
        records = df.to_dict(orient="records")
        tail = records[-limit:] if limit > 0 else records
        for row in tail:
            t = pick_time(row)
            v = pick_value(row)
            if not t and not v:
                continue
            label = pick_label(row)
            unit = pick_unit(row)
            items.append({
                "indicator": name,
                "label": label,
                "value": v,
                "unit": unit,
                "time": t,
                "source": "AkShare",
            })
    print(json.dumps(items, ensure_ascii=False))

if __name__ == "__main__":
    main()

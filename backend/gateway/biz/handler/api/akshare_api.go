package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"stock_assistant/backend/gateway/config"
)

func GetAkShareMacro(_ context.Context, ctx *app.RequestContext) {
	cfg := config.Get()
	base := "http://127.0.0.1:8895"
	if cfg != nil && cfg.RPC != nil && cfg.RPC.AkShareServiceAddr != "" {
		base = "http://" + cfg.RPC.AkShareServiceAddr
	}
	funcs := string(ctx.Query("funcs"))
	limit := string(ctx.Query("limit"))
	symbol := string(ctx.Query("symbol"))
	if funcs == "" {
		funcs = "macro_china_cpi,macro_china_ppi,macro_china_pmi,macro_china_gdp,macro_china_fx_reserves,macro_china_money_supply"
	}
	if limit == "" {
		limit = "12"
	}
	url := strings.TrimRight(base, "/") + "/macro?funcs=" + funcs + "&limit=" + limit + "&symbol=" + symbol
	resp, err := http.Get(url)
	if err != nil {
		ctx.SetStatusCode(http.StatusBadGateway)
		ctx.Write([]byte(`{"error":"akshare service unreachable"}`))
		return
	}
	defer resp.Body.Close()
	ctx.SetStatusCode(resp.StatusCode)
	ctx.SetContentType("application/json; charset=utf-8")
	body := make([]byte, 0)
	buf := make([]byte, 4096)
	for {
		n, e := resp.Body.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
		}
		if e != nil {
			break
		}
	}
	ctx.Write(body)
}

func GetAkShareNews(_ context.Context, ctx *app.RequestContext) {
	cfg := config.Get()
	base := "http://127.0.0.1:8895"
	if cfg != nil && cfg.RPC != nil && cfg.RPC.AkShareServiceAddr != "" {
		base = "http://" + cfg.RPC.AkShareServiceAddr
	}
	funcName := string(ctx.Query("func"))
	symbol := string(ctx.Query("symbol"))
	limit := string(ctx.Query("limit"))
	if limit == "" {
		limit = "50"
	}
	url := strings.TrimRight(base, "/") + "/news?func=" + funcName + "&symbol=" + symbol + "&limit=" + limit
	resp, err := http.Get(url)
	if err != nil {
		ctx.SetStatusCode(http.StatusBadGateway)
		ctx.Write([]byte(`{"error":"akshare service unreachable"}`))
		return
	}
	defer resp.Body.Close()
	ctx.SetStatusCode(resp.StatusCode)
	ctx.SetContentType("application/json; charset=utf-8")
	body := make([]byte, 0)
	buf := make([]byte, 4096)
	for {
		n, e := resp.Body.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
		}
		if e != nil {
			break
		}
	}
	ctx.Write(body)
}

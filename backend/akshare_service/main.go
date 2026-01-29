package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

type MacroItem struct {
	Indicator string `json:"indicator"`
	Label     string `json:"label"`
	Value     string `json:"value"`
	Unit      string `json:"unit"`
	Time      string `json:"time"`
	Source    string `json:"source"`
}

type NewsItem struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	Source  string `json:"source"`
	Url     string `json:"url"`
	Time    string `json:"time"`
}

func runPython(args ...string) ([]byte, error) {
	pythonBin := os.Getenv("AK_PYTHON_BIN")
	if pythonBin == "" {
		pythonBin = "python3"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, pythonBin, args...)
	return cmd.Output()
}

func handleMacro(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	funcs := strings.TrimSpace(q.Get("funcs"))
	if funcs == "" {
		funcs = "macro_china_cpi,macro_china_ppi,macro_china_pmi,macro_china_gdp,macro_china_fx_reserves,macro_china_money_supply"
	}
	limitStr := strings.TrimSpace(q.Get("limit"))
	if limitStr == "" {
		limitStr = "12"
	}
	symbol := strings.TrimSpace(q.Get("symbol"))
	script := "scripts/akshare_macro.py"
	out, err := runPython(script, funcs, limitStr, symbol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(out)
}

func handleNews(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	funcName := strings.TrimSpace(q.Get("func"))
	symbol := strings.TrimSpace(q.Get("symbol"))
	limitStr := strings.TrimSpace(q.Get("limit"))
	if limitStr == "" {
		limitStr = "50"
	}
	script := "scripts/akshare_news.py"
	out, err := runPython(script, funcName, symbol, limitStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write(out)
}

func main() {
	addr := os.Getenv("AK_SERVICE_ADDR")
	if addr == "" {
		addr = ":8895"
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/macro", handleMacro)
	mux.HandleFunc("/news", handleNews)
	s := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadTimeout:       20 * time.Second,
		WriteTimeout:      20 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}
	log.Printf("AkShare HTTP service listening at %s", addr)
	log.Fatal(s.ListenAndServe())
}

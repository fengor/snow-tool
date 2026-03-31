package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ---------------------------------------------------------------------------
// Table keyword → ServiceNow table name mapping
// ---------------------------------------------------------------------------

var tableKeywords = map[string]string{
	"incident":  "incident",
	"inc":       "incident",
	"change":    "change_request",
	"chg":       "change_request",
	"problem":   "problem",
	"prb":       "problem",
	"user":      "sys_user",
	"usr":       "sys_user",
	"group":     "sys_user_group",
	"grp":       "sys_user_group",
	"ci":        "cmdb_ci",
	"cmdb":      "cmdb_ci",
	"task":      "task",
	"request":   "sc_request",
	"req":       "sc_request",
	"ritm":      "sc_req_item",
	"catalog":   "sc_cat_item",
	"knowledge": "kb_knowledge",
	"kb":        "kb_knowledge",
}

func resolveTable(keyword string) string {
	if t, ok := tableKeywords[strings.ToLower(keyword)]; ok {
		return t
	}
	return keyword
}

// ---------------------------------------------------------------------------
// Config – read from environment variables
// ---------------------------------------------------------------------------
type Config struct {
	Instance string // SNOW_INSTANCE  e.g. "dev12345"
	User     string // SNOW_USER
	Password string // SNOW_PASSWORD
}

func loadConfig() Config {
	c := Config{
		Instance: os.Getenv("SNOW_INSTANCE"),
		User:     os.Getenv("SNOW_USER"),
		Password: os.Getenv("SNOW_PASSWORD"),
	}
	var missing []string
	if c.Instance == "" {
		missing = append(missing, "SNOW_INSTANCE")
	}
	if c.User == "" {
		missing = append(missing, "SNOW_USER")
	}
	if c.Password == "" {
		missing = append(missing, "SNOW_PASSWORD")
	}
	if len(missing) > 0 {
		fatalf("missing required environment variable(s): %s", strings.Join(missing, ", "))
	}
	return c
}

func (c Config) baseURL() string {
	return fmt.Sprintf("https://%s/api/now/table", c.Instance)
}

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------
var httpClient = &http.Client{Timeout: 30 * time.Second}

func doRequest(method, url, user, password string, body []byte) ([]byte, int, error) {
	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, 0, fmt.Errorf("building request: %w", err)
	}

	req.SetBasicAuth(user, password)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("reading response: %w", err)
	}
	return respBody, resp.StatusCode, nil
}

func prettyJSON(data []byte) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Println(string(data))
		return
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}

// ---------------------------------------------------------------------------
// Verbs
// ---------------------------------------------------------------------------
func cmdList(cfg Config, table string) {
	url := fmt.Sprintf("%s/%s", cfg.baseURL(), table)
	body, status, err := doRequest(http.MethodGet, url, cfg.User, cfg.Password, nil)
	if err != nil {
		fatalf("list: %v", err)
	}
	if status < 200 || status >= 300 {
		fatalf("list HTTP %d: %s", status, string(body))
	}
	fmt.Printf("// Table: %s\n", table)
	prettyJSON(body)
}

func cmdGet(cfg Config, table, sysID string) {
	url := fmt.Sprintf("%s/%s/%s", cfg.baseURL(), table, sysID)
	body, status, err := doRequest(http.MethodGet, url, cfg.User, cfg.Password, nil)
	if err != nil {
		fatalf("get: %v", err)
	}
	if status < 200 || status >= 300 {
		fatalf("get HTTP %d: %s", status, string(body))
	}
	fmt.Printf("// Table: %s  sys_id: %s\n", table, sysID)
	prettyJSON(body)
}

func cmdCreate(cfg Config, table, rawJSON string) {
	if !json.Valid([]byte(rawJSON)) {
		fatalf("create: argument is not valid JSON")
	}
	url := fmt.Sprintf("%s/%s", cfg.baseURL(), table)
	body, status, err := doRequest(http.MethodPost, url, cfg.User, cfg.Password, []byte(rawJSON))
	if err != nil {
		fatalf("create: %v", err)
	}
	if status < 200 || status >= 300 {
		fatalf("create HTTP %d: %s", status, string(body))
	}
	fmt.Printf("// Created in table: %s\n", table)
	prettyJSON(body)
}

func cmdDelete(cfg Config, table, sysID string) {
	url := fmt.Sprintf("%s/%s/%s", cfg.baseURL(), table, sysID)
	body, status, err := doRequest(http.MethodDelete, url, cfg.User, cfg.Password, nil)
	if err != nil {
		fatalf("delete: %v", err)
	}
	if status == http.StatusNoContent {
		fmt.Printf("Deleted sys_id %s from table %s\n", sysID, table)
		return
	}
	if status < 200 || status >= 300 {
		fatalf("delete HTTP %d: %s", status, string(body))
	}
	fmt.Printf("Deleted sys_id %s from table %s\n", sysID, table)
}

// ---------------------------------------------------------------------------
// Usage
// ---------------------------------------------------------------------------
func usage() {
	fmt.Fprintln(os.Stderr, `Usage:
  snow-tool <table> list
  snow-tool <table> get    <sys_id>
  snow-tool <table> create <json>
  snow-tool <table> delete <sys_id>

Environment variables (required):
  SNOW_INSTANCE   ServiceNow instance URL (e.g. dev12345.service-now.com)
  SNOW_USER       Basic-auth username
  SNOW_PASSWORD   Basic-auth password

Table keywords (or use any raw ServiceNow table name directly):
  incident / inc      → incident
  change   / chg      → change_request
  problem  / prb      → problem
  user     / usr      → sys_user
  group    / grp      → sys_user_group
  ci       / cmdb     → cmdb_ci
  task                → task
  request  / req      → sc_request
  ritm                → sc_req_item
  catalog             → sc_cat_item
  knowledge / kb      → kb_knowledge

Examples:
  snow-tool incident list
  snow-tool inc get    1234abcd1234abcd1234abcd1234abcd
  snow-tool inc create '{"short_description":"Disk full","urgency":"1"}'
  snow-tool inc delete 1234abcd1234abcd1234abcd1234abcd
  snow-tool cmdb_ci_server list`)
	os.Exit(1)
}

// ---------------------------------------------------------------------------
// Entry point
// ---------------------------------------------------------------------------
func main() {
	args := os.Args[1:]

	if len(args) < 2 {
		usage()
	}

	tableKeyword := args[0]
	verb := strings.ToLower(args[1])
	table := resolveTable(tableKeyword)

	cfg := loadConfig()

	switch verb {
	case "list":
		if len(args) != 2 {
			fatalf("list takes no additional arguments\n\nUsage: snow-tool <table> list")
		}
		cmdList(cfg, table)

	case "get":
		if len(args) != 3 {
			fatalf("get requires exactly one argument: <sys_id>\n\nUsage: snow-tool <table> get <sys_id>")
		}
		cmdGet(cfg, table, args[2])

	case "create":
		if len(args) != 3 {
			fatalf("create requires exactly one argument: <json>\n\nUsage: snow-tool <table> create '<json>'")
		}
		cmdCreate(cfg, table, args[2])

	case "delete":
		if len(args) != 3 {
			fatalf("delete requires exactly one argument: <sys_id>\n\nUsage: snow-tool <table> delete <sys_id>")
		}
		cmdDelete(cfg, table, args[2])

	default:
		fatalf("unknown verb %q — must be one of: list, get, create, delete", verb)
	}
}

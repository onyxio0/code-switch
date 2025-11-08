package services

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/daodao97/xgo/xdb"
	"github.com/daodao97/xgo/xrequest"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	_ "modernc.org/sqlite"
)

type ProviderRelayService struct {
	providerService *ProviderService
	server          *http.Server
	addr            string
}

func NewProviderRelayService(providerService *ProviderService, addr string) *ProviderRelayService {
	if addr == "" {
		addr = ":18100"
	}

	home, _ := os.UserHomeDir()

	if err := xdb.Inits([]xdb.Config{
		{
			Name:   "default",
			Driver: "sqlite",
			DSN:    filepath.Join(home, ".code-switch", "app.db?cache=shared&mode=rwc"),
		},
	}); err != nil {
		fmt.Printf("初始化数据库失败: %v\n", err)
	} else if err := ensureRequestLogTable(); err != nil {
		fmt.Printf("初始化 request_log 表失败: %v\n", err)
	}

	return &ProviderRelayService{
		providerService: providerService,
		addr:            addr,
	}
}

func (prs *ProviderRelayService) Start() error {
	router := gin.Default()
	prs.registerRoutes(router)

	prs.server = &http.Server{
		Addr:    prs.addr,
		Handler: router,
	}

	fmt.Printf("provider relay server listening on %s\n", prs.addr)

	go func() {
		if err := prs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("provider relay server error: %v\n", err)
		}
	}()
	return nil
}

func (prs *ProviderRelayService) Stop() error {
	if prs.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return prs.server.Shutdown(ctx)
}

func (prs *ProviderRelayService) Addr() string {
	return prs.addr
}

func (prs *ProviderRelayService) registerRoutes(router gin.IRouter) {
	router.POST("/v1/messages", prs.proxyHandler("claude", "/messages"))
	router.POST("/responses", prs.proxyHandler("codex", "/responses"))
}

func (prs *ProviderRelayService) proxyHandler(kind string, endpoint string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var bodyBytes []byte
		if c.Request.Body != nil {
			data, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
				return
			}
			bodyBytes = data
			c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		isStream := gjson.GetBytes(bodyBytes, "stream").Bool()
		model := gjson.GetBytes(bodyBytes, "model").String()

		providers, err := prs.providerService.LoadProviders(kind)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load providers"})
			return
		}

		active := make([]Provider, 0, len(providers))
		for _, provider := range providers {
			if provider.Enabled && provider.APIURL != "" && provider.APIKey != "" {
				active = append(active, provider)
			}
		}

		if len(active) == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "no providers available"})
			return
		}

		query := flattenQuery(c.Request.URL.Query())
		clientHeaders := cloneHeaders(c.Request.Header)

		var lastErr error
		for _, provider := range active {
			ok, err := prs.forwardRequest(c, kind, provider, endpoint, query, clientHeaders, bodyBytes, isStream, model)
			if ok {
				return
			}
			lastErr = err
		}

		message := "all providers failed"
		if lastErr != nil {
			message = lastErr.Error()
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": message})
	}
}

func (prs *ProviderRelayService) forwardRequest(
	c *gin.Context,
	kind string,
	provider Provider,
	endpoint string,
	query map[string]string,
	clientHeaders map[string]string,
	bodyBytes []byte,
	isStream bool,
	model string,
) (bool, error) {
	targetURL := joinURL(provider.APIURL, endpoint)
	headers := cloneMap(clientHeaders)
	headers["Authorization"] = fmt.Sprintf("Bearer %s", provider.APIKey)
	if _, ok := headers["Accept"]; !ok {
		headers["Accept"] = "application/json"
	}

	requestLog := &ReqeustLog{
		Platform: kind,
		Provider: provider.Name,
		Model:    model,
		IsStream: isStream,
	}
	start := time.Now()
	defer func() {
		requestLog.DurationSec = time.Since(start).Seconds()
		if _, err := xdb.New("request_log").Insert(xdb.Record{
			"platform":            requestLog.Platform,
			"model":               requestLog.Model,
			"provider":            requestLog.Provider,
			"http_code":           requestLog.HttpCode,
			"input_tokens":        requestLog.InputTokens,
			"output_tokens":       requestLog.OutputTokens,
			"cache_create_tokens": requestLog.CacheCreateTokens,
			"cache_read_tokens":   requestLog.CacheReadTokens,
			"reasoning_tokens":    requestLog.ReasoningTokens,
			"is_stream":           boolToInt(requestLog.IsStream),
			"duration_sec":        requestLog.DurationSec,
		}); err != nil {
			fmt.Printf("写入 request_log 失败: %v\n", err)
		}
	}()

	req := xrequest.New().
		SetHeaders(headers).
		SetQueryParams(query).
		SetRetry(1, 500*time.Millisecond).
		SetTimeout(60 * time.Second)

	reqBody := bytes.NewReader(bodyBytes)
	req = req.SetBody(reqBody)

	resp, err := req.Post(targetURL)
	if err != nil {
		return false, err
	}

	if resp == nil {
		return false, fmt.Errorf("empty response")
	}

	if resp.Error() != nil {
		return false, resp.Error()
	}

	status := resp.StatusCode()
	requestLog.HttpCode = status

	if status >= http.StatusOK && status < http.StatusMultipleChoices {
		_, copyErr := resp.ToHttpResponseWriter(c.Writer, ReqeustLogHook(c, kind, requestLog))
		return copyErr == nil, copyErr
	}

	return false, fmt.Errorf("upstream status %d", status)
}

func cloneHeaders(header http.Header) map[string]string {
	cloned := make(map[string]string, len(header))
	for key, values := range header {
		if len(values) > 0 {
			cloned[key] = values[len(values)-1]
		}
	}
	return cloned
}

func cloneMap(m map[string]string) map[string]string {
	cloned := make(map[string]string, len(m))
	for k, v := range m {
		cloned[k] = v
	}
	return cloned
}

func flattenQuery(values map[string][]string) map[string]string {
	query := make(map[string]string, len(values))
	for key, items := range values {
		if len(items) > 0 {
			query[key] = items[len(items)-1]
		}
	}
	return query
}

func joinURL(base string, endpoint string) string {
	base = strings.TrimSuffix(base, "/")
	endpoint = "/" + strings.TrimPrefix(endpoint, "/")
	return base + endpoint
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func ensureRequestLogColumn(db *sql.DB, column string, definition string) error {
	query := fmt.Sprintf("SELECT COUNT(*) FROM pragma_table_info('request_log') WHERE name = '%s'", column)
	var count int
	if err := db.QueryRow(query).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		alter := fmt.Sprintf("ALTER TABLE request_log ADD COLUMN %s %s", column, definition)
		if _, err := db.Exec(alter); err != nil {
			return err
		}
	}
	return nil
}

func ensureRequestLogTable() error {
	db, err := xdb.DB("default")
	if err != nil {
		return err
	}
	return ensureRequestLogTableWithDB(db)
}

func ensureRequestLogTableWithDB(db *sql.DB) error {
	const createTableSQL = `CREATE TABLE IF NOT EXISTS request_log (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		platform TEXT,
		model TEXT,
		provider TEXT,
		http_code INTEGER,
		input_tokens INTEGER,
		output_tokens INTEGER,
		cache_create_tokens INTEGER,
		cache_read_tokens INTEGER,
		reasoning_tokens INTEGER,
		is_stream INTEGER DEFAULT 0,
		duration_sec REAL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	if _, err := db.Exec(createTableSQL); err != nil {
		return err
	}

	if err := ensureRequestLogColumn(db, "created_at", "DATETIME DEFAULT CURRENT_TIMESTAMP"); err != nil {
		return err
	}
	if err := ensureRequestLogColumn(db, "is_stream", "INTEGER DEFAULT 0"); err != nil {
		return err
	}
	if err := ensureRequestLogColumn(db, "duration_sec", "REAL DEFAULT 0"); err != nil {
		return err
	}

	return nil
}

func ReqeustLogHook(c *gin.Context, kind string, usage *ReqeustLog) func(data []byte) (bool, []byte) { // SSE 钩子：累计字节和解析 token 用量
	return func(data []byte) (bool, []byte) {
		payload := strings.TrimSpace(string(data))

		parserFn := ClaudeCodeParseTokenUsageFromResponse
		if kind == "codex" {
			parserFn = CodexParseTokenUsageFromResponse
		}
		parseEventPayload(payload, parserFn, usage)

		return true, data
	}
}

func parseEventPayload(payload string, parser func(string, *ReqeustLog), usage *ReqeustLog) {
	lines := strings.Split(payload, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "data:") {
			parser(strings.TrimPrefix(line, "data: "), usage)
		}
	}
}

type ReqeustLog struct {
	ID                int64   `json:"id"`
	Platform          string  `json:"platform"` // claude code or codex
	Model             string  `json:"model"`
	Provider          string  `json:"provider"` // provider name
	HttpCode          int     `json:"http_code"`
	InputTokens       int     `json:"input_tokens"`
	OutputTokens      int     `json:"output_tokens"`
	CacheCreateTokens int     `json:"cache_create_tokens"`
	CacheReadTokens   int     `json:"cache_read_tokens"`
	ReasoningTokens   int     `json:"reasoning_tokens"`
	IsStream          bool    `json:"is_stream"`
	DurationSec       float64 `json:"duration_sec"`
	CreatedAt         string  `json:"created_at"`
	InputCost         float64 `json:"input_cost"`
	OutputCost        float64 `json:"output_cost"`
	CacheCreateCost   float64 `json:"cache_create_cost"`
	CacheReadCost     float64 `json:"cache_read_cost"`
	Ephemeral5mCost   float64 `json:"ephemeral_5m_cost"`
	Ephemeral1hCost   float64 `json:"ephemeral_1h_cost"`
	TotalCost         float64 `json:"total_cost"`
	HasPricing        bool    `json:"has_pricing"`
}

// claude code usage parser
func ClaudeCodeParseTokenUsageFromResponse(data string, usage *ReqeustLog) {
	usage.InputTokens += int(gjson.Get(data, "message.usage.input_tokens").Int())
	usage.OutputTokens += int(gjson.Get(data, "message.usage.output_tokens").Int())
	usage.CacheCreateTokens += int(gjson.Get(data, "message.usage.cache_creation_input_tokens").Int())
	usage.CacheReadTokens += int(gjson.Get(data, "message.usage.cache_read_input_tokens").Int())

	usage.InputTokens += int(gjson.Get(data, "usage.input_tokens").Int())
	usage.OutputTokens += int(gjson.Get(data, "usage.output_tokens").Int())
}

// codex usage parser
func CodexParseTokenUsageFromResponse(data string, usage *ReqeustLog) {
	usage.InputTokens += int(gjson.Get(data, "response.usage.input_tokens").Int())
	usage.OutputTokens += int(gjson.Get(data, "response.usage.output_tokens").Int())
	usage.CacheReadTokens += int(gjson.Get(data, "response.usage.input_tokens_details.cached_tokens").Int())
	usage.ReasoningTokens += int(gjson.Get(data, "response.usage.output_tokens_details.reasoning_tokens").Int())
	fmt.Println("data ---->", data, fmt.Sprintf("%v", usage))
}

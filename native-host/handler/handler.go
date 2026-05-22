package handler

import (
	"browser-bridge/native-host/native"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

type Handler struct {
	bridge *native.Bridge
}

func NewHandler(bridge *native.Bridge) *Handler {
	return &Handler{bridge: bridge}
}

// apiResponse 统一 HTTP 响应
func apiResponse(w http.ResponseWriter, success bool, data interface{}, errMsg string) {
	resp := map[string]interface{}{
		"success": success,
	}
	if data != nil {
		resp["data"] = data
	}
	if errMsg != "" {
		resp["error"] = errMsg
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// sendAndWait 向 Extension 发送消息并等待响应
func (h *Handler) sendAndWait(action string, params map[string]interface{}) (*native.ResponseMsg, error) {
	id := uuid.New().String()
	timeout := make(chan struct{})
	go func() {
		time.Sleep(30 * time.Second)
		close(timeout)
	}()

	return h.bridge.SendAndWait(&native.RequestMsg{
		ID:     id,
		Action: action,
		Params: params,
	}, timeout)
}

// HandleTabs 获取 Tab 列表
func (h *Handler) HandleTabs(w http.ResponseWriter, r *http.Request) {
	resp, err := h.sendAndWait("tab.list", nil)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleTabGet 获取指定 Tab
func (h *Handler) HandleTabGet(w http.ResponseWriter, r *http.Request) {
	tabID := r.PathValue("tabId")
	params := map[string]interface{}{"tabId": toInt(tabID)}
	resp, err := h.sendAndWait("tab.get", params)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleTabCreate 创建 Tab
func (h *Handler) HandleTabCreate(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apiResponse(w, false, nil, "invalid request body")
		return
	}
	resp, err := h.sendAndWait("tab.create", body)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleTabClose 关闭 Tab
func (h *Handler) HandleTabClose(w http.ResponseWriter, r *http.Request) {
	tabID := r.PathValue("tabId")
	params := map[string]interface{}{"tabId": toInt(tabID)}
	resp, err := h.sendAndWait("tab.close", params)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleTabActivate 激活 Tab
func (h *Handler) HandleTabActivate(w http.ResponseWriter, r *http.Request) {
	tabID := r.PathValue("tabId")
	params := map[string]interface{}{"tabId": toInt(tabID)}
	resp, err := h.sendAndWait("tab.activate", params)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleSnapshot 获取页面快照
func (h *Handler) HandleSnapshot(w http.ResponseWriter, r *http.Request) {
	tabID := r.PathValue("tabId")
	params := map[string]interface{}{"tabId": toInt(tabID)}

	refOnly := r.URL.Query().Get("refOnly")
	if refOnly == "true" {
		params["refOnly"] = true
	}

	resp, err := h.sendAndWait("page.snapshot", params)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleContent 获取页面内容
func (h *Handler) HandleContent(w http.ResponseWriter, r *http.Request) {
	tabID := r.PathValue("tabId")
	params := map[string]interface{}{"tabId": toInt(tabID)}
	if format := r.URL.Query().Get("format"); format != "" {
		params["format"] = format
	}
	resp, err := h.sendAndWait("page.content", params)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleScreenshot 页面截图
func (h *Handler) HandleScreenshot(w http.ResponseWriter, r *http.Request) {
	tabID := r.PathValue("tabId")
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		body = map[string]interface{}{}
	}
	body["tabId"] = toInt(tabID)
	resp, err := h.sendAndWait("page.screenshot", body)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleExecute 执行脚本
func (h *Handler) HandleExecute(w http.ResponseWriter, r *http.Request) {
	tabID := r.PathValue("tabId")
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apiResponse(w, false, nil, "invalid request body")
		return
	}
	body["tabId"] = toInt(tabID)
	resp, err := h.sendAndWait("page.execute", body)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleClick 点击元素
func (h *Handler) HandleClick(w http.ResponseWriter, r *http.Request) {
	tabID := r.PathValue("tabId")
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apiResponse(w, false, nil, "invalid request body")
		return
	}
	body["tabId"] = toInt(tabID)
	resp, err := h.sendAndWait("page.click", body)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleType 输入文本
func (h *Handler) HandleType(w http.ResponseWriter, r *http.Request) {
	tabID := r.PathValue("tabId")
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apiResponse(w, false, nil, "invalid request body")
		return
	}
	body["tabId"] = toInt(tabID)
	resp, err := h.sendAndWait("page.type", body)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleSelect 选择下拉框
func (h *Handler) HandleSelect(w http.ResponseWriter, r *http.Request) {
	tabID := r.PathValue("tabId")
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apiResponse(w, false, nil, "invalid request body")
		return
	}
	body["tabId"] = toInt(tabID)
	resp, err := h.sendAndWait("page.select", body)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleScroll 滚动页面
func (h *Handler) HandleScroll(w http.ResponseWriter, r *http.Request) {
	tabID := r.PathValue("tabId")
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		body = map[string]interface{}{}
	}
	body["tabId"] = toInt(tabID)
	resp, err := h.sendAndWait("page.scroll", body)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleQuery 查询 DOM 元素
func (h *Handler) HandleQuery(w http.ResponseWriter, r *http.Request) {
	tabID := r.PathValue("tabId")
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apiResponse(w, false, nil, "invalid request body")
		return
	}
	body["tabId"] = toInt(tabID)
	resp, err := h.sendAndWait("page.query", body)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleWait 等待元素出现
func (h *Handler) HandleWait(w http.ResponseWriter, r *http.Request) {
	tabID := r.PathValue("tabId")
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apiResponse(w, false, nil, "invalid request body")
		return
	}
	body["tabId"] = toInt(tabID)
	resp, err := h.sendAndWait("page.wait", body)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleNavigate 导航到 URL
func (h *Handler) HandleNavigate(w http.ResponseWriter, r *http.Request) {
	tabID := r.PathValue("tabId")
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apiResponse(w, false, nil, "invalid request body")
		return
	}
	body["tabId"] = toInt(tabID)
	resp, err := h.sendAndWait("nav.goto", body)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleBack 后退
func (h *Handler) HandleBack(w http.ResponseWriter, r *http.Request) {
	tabID := r.PathValue("tabId")
	params := map[string]interface{}{"tabId": toInt(tabID)}
	resp, err := h.sendAndWait("nav.back", params)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleForward 前进
func (h *Handler) HandleForward(w http.ResponseWriter, r *http.Request) {
	tabID := r.PathValue("tabId")
	params := map[string]interface{}{"tabId": toInt(tabID)}
	resp, err := h.sendAndWait("nav.forward", params)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleReload 刷新
func (h *Handler) HandleReload(w http.ResponseWriter, r *http.Request) {
	tabID := r.PathValue("tabId")
	params := map[string]interface{}{"tabId": toInt(tabID)}
	resp, err := h.sendAndWait("nav.reload", params)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleWaitLoad 等待页面加载
func (h *Handler) HandleWaitLoad(w http.ResponseWriter, r *http.Request) {
	tabID := r.PathValue("tabId")
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		body = map[string]interface{}{}
	}
	body["tabId"] = toInt(tabID)
	resp, err := h.sendAndWait("nav.waitLoad", body)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleCookieGet 获取 Cookie
func (h *Handler) HandleCookieGet(w http.ResponseWriter, r *http.Request) {
	params := map[string]interface{}{}
	if url := r.URL.Query().Get("url"); url != "" {
		params["url"] = url
	}
	if name := r.URL.Query().Get("name"); name != "" {
		params["name"] = name
	}
	if domain := r.URL.Query().Get("domain"); domain != "" {
		params["domain"] = domain
	}
	resp, err := h.sendAndWait("cookie.get", params)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleCookieSet 设置 Cookie
func (h *Handler) HandleCookieSet(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apiResponse(w, false, nil, "invalid request body")
		return
	}
	resp, err := h.sendAndWait("cookie.set", body)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleCookieDelete 删除 Cookie
func (h *Handler) HandleCookieDelete(w http.ResponseWriter, r *http.Request) {
	params := map[string]interface{}{}
	if url := r.URL.Query().Get("url"); url != "" {
		params["url"] = url
	}
	if name := r.URL.Query().Get("name"); name != "" {
		params["name"] = name
	}
	resp, err := h.sendAndWait("cookie.delete", params)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleSearch 搜索
func (h *Handler) HandleSearch(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apiResponse(w, false, nil, "invalid request body")
		return
	}

	engine, _ := body["engine"].(string)
	if engine == "" {
		engine = "baidu"
	}

	action := "search." + engine
	delete(body, "engine")

	resp, err := h.sendAndWait(action, body)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleFetchContent fetch content from URL
func (h *Handler) HandleFetchContent(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		apiResponse(w, false, nil, "invalid request body")
		return
	}
	resp, err := h.sendAndWait("fetch.content", body)
	if err != nil {
		apiResponse(w, false, nil, "timeout")
		return
	}
	if !resp.Success {
		apiResponse(w, false, nil, resp.Error)
		return
	}
	apiResponse(w, true, resp.Data, "")
}

// HandleHealth 健康检查
func (h *Handler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	apiResponse(w, true, map[string]string{"status": "ok"}, "")
}

func toInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

// RegisterRoutes 注册所有路由
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	prefix := "/api/v1"

	// Health
	mux.HandleFunc(prefix+"/health", h.HandleHealth)

	// Tabs
	mux.HandleFunc(prefix+"/tabs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.HandleTabs(w, r)
		case http.MethodPost:
			h.HandleTabCreate(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc(prefix+"/tabs/{tabId}", h.HandleTabGet)
	mux.HandleFunc(prefix+"/tabs/{tabId}/close", h.HandleTabClose)
	mux.HandleFunc(prefix+"/tabs/{tabId}/activate", h.HandleTabActivate)

	// Page
	mux.HandleFunc(prefix+"/tabs/{tabId}/snapshot", h.HandleSnapshot)
	mux.HandleFunc(prefix+"/tabs/{tabId}/content", h.HandleContent)
	mux.HandleFunc(prefix+"/tabs/{tabId}/screenshot", h.HandleScreenshot)
	mux.HandleFunc(prefix+"/tabs/{tabId}/execute", h.HandleExecute)
	mux.HandleFunc(prefix+"/tabs/{tabId}/click", h.HandleClick)
	mux.HandleFunc(prefix+"/tabs/{tabId}/type", h.HandleType)
	mux.HandleFunc(prefix+"/tabs/{tabId}/select", h.HandleSelect)
	mux.HandleFunc(prefix+"/tabs/{tabId}/scroll", h.HandleScroll)
	mux.HandleFunc(prefix+"/tabs/{tabId}/query", h.HandleQuery)
	mux.HandleFunc(prefix+"/tabs/{tabId}/wait", h.HandleWait)

	// Navigation
	mux.HandleFunc(prefix+"/tabs/{tabId}/navigate", h.HandleNavigate)
	mux.HandleFunc(prefix+"/tabs/{tabId}/back", h.HandleBack)
	mux.HandleFunc(prefix+"/tabs/{tabId}/forward", h.HandleForward)
	mux.HandleFunc(prefix+"/tabs/{tabId}/reload", h.HandleReload)
	mux.HandleFunc(prefix+"/tabs/{tabId}/wait-load", h.HandleWaitLoad)

	// Cookies
	mux.HandleFunc(prefix+"/cookies", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			h.HandleCookieGet(w, r)
		case http.MethodPost:
			h.HandleCookieSet(w, r)
		case http.MethodDelete:
			h.HandleCookieDelete(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Search
	mux.HandleFunc(prefix+"/search", h.HandleSearch)

	// Fetch
	mux.HandleFunc(prefix+"/fetch/content", h.HandleFetchContent)

	// Root health check
	mux.HandleFunc("/", h.HandleHealth)

	_ = fmt.Sprintf("") // avoid unused import
}

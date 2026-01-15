package server

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"o2m/internal/converter"
)

// 嵌入静态资源
//
//go:embed static
var staticFiles embed.FS

// Version 版本号
const Version = "v1.5.0"

// ConvertRequest 转换请求结构
type ConvertRequest struct {
	DDL string `json:"ddl"` // Oracle DDL 内容
}

// ConvertResponse 转换响应结构
type ConvertResponse struct {
	Success bool   `json:"success"` // 是否成功
	Result  string `json:"result"`  // 转换结果
	Error   string `json:"error"`   // 错误信息
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

const (
	AuthCookieName = "o2m_session"
	// AuthCookieValue 基于版本号，每次更新版本后自动失效旧 session
	AuthCookieValue = "auth_" + Version
	HardcodedUser   = "cdfh"
	HardcodedPass   = "cdfh@2026!"
)

// StartWebServer 启动 Web 服务器
func StartWebServer(port int) error {
	// 静态文件路由
	http.HandleFunc("/sql-formatter.js", handleStaticFile)

	// 页面路由
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/login", handleLoginPage)
	http.HandleFunc("/logout", handleLogout)

	// API 路由
	http.HandleFunc("/api/login", handleLoginApi)
	http.HandleFunc("/api/convert", handleConvert)
	http.HandleFunc("/api/upload", handleUpload)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("------------------------------------------")
	log.Printf("Web 服务器启动在 http://localhost%s", addr)
	//log.Printf("账号: %s", HardcodedUser)
	//log.Printf("密码: %s", HardcodedPass)
	log.Printf("------------------------------------------")

	return http.ListenAndServe(addr, nil)
}

// checkAuth 检查是否已登录
func checkAuth(r *http.Request) bool {
	cookie, err := r.Cookie(AuthCookieName)
	if err != nil {
		return false
	}
	return cookie.Value == AuthCookieValue
}

// handleLoginPage 渲染登录页
func handleLoginPage(w http.ResponseWriter, r *http.Request) {
	log.Printf("[访问] %s %s - 登录页", r.Method, r.URL.Path)
	if checkAuth(r) {
		log.Printf("[授权] 用户已登录，重定向到主页")
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	data, err := staticFiles.ReadFile("static/login.html")
	if err != nil {
		log.Printf("[错误] 读取 login.html 失败: %v", err)
		http.Error(w, "无法加载登录页面", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

// handleLogout 登出
func handleLogout(w http.ResponseWriter, r *http.Request) {
	log.Printf("[访问] %s %s - 登出", r.Method, r.URL.Path)
	http.SetCookie(w, &http.Cookie{
		Name:   AuthCookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusFound)
}

// handleLoginApi 处理登录 API
func handleLoginApi(w http.ResponseWriter, r *http.Request) {
	log.Printf("[访问] %s %s - 登录验证接口", r.Method, r.URL.Path)
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[错误] 解析登录请求失败: %v", err)
		json.NewEncoder(w).Encode(ConvertResponse{Success: false, Error: "解析请求失败"})
		return
	}

	if req.Username == HardcodedUser && req.Password == HardcodedPass {
		log.Printf("[成功] 登录验证通过: %s", req.Username)
		// 设置 Cookie
		http.SetCookie(w, &http.Cookie{
			Name:     AuthCookieName,
			Value:    AuthCookieValue,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   3600 * 24, // 24小时
		})
		json.NewEncoder(w).Encode(ConvertResponse{Success: true})
	} else {
		log.Printf("[失败] 登录验证失败，账号: %s", req.Username)
		json.NewEncoder(w).Encode(ConvertResponse{Success: false, Error: "账号或密码不正确"})
	}
}

// handleIndex 处理主页请求
func handleIndex(w http.ResponseWriter, r *http.Request) {
	log.Printf("[访问] %s %s - 主页", r.Method, r.URL.Path)

	// 处理 favicon 等自动请求，避免干扰
	if r.URL.Path != "/" {
		if !checkAuth(r) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	if !checkAuth(r) {
		log.Printf("[授权] 未登录，重定向到 /login")
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	// 从嵌入的文件系统读取 index.html
	data, err := staticFiles.ReadFile("static/index.html")
	if err != nil {
		log.Printf("[错误] 读取 index.html 失败: %v", err)
		http.Error(w, "无法加载主页面", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(data)
}

// handleConvert 处理文本转换请求
func handleConvert(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(r) {
		http.Error(w, "未授权访问", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "只支持 POST 请求", http.StatusMethodNotAllowed)
		return
	}

	// 设置 CORS 头（如果需要）
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 解析请求
	var req ConvertRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		resp := ConvertResponse{
			Success: false,
			Error:   "无效的请求格式",
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	// 执行转换
	result, err := converter.ConvertToMySQL(req.DDL)
	if err != nil {
		resp := ConvertResponse{
			Success: false,
			Error:   fmt.Sprintf("转换失败: %v", err),
		}
		json.NewEncoder(w).Encode(resp)
		return
	}

	// 返回成功结果
	resp := ConvertResponse{
		Success: true,
		Result:  result,
	}
	json.NewEncoder(w).Encode(resp)

	log.Printf("文本转换成功，输入长度: %d, 输出长度: %d", len(req.DDL), len(result))
}

// handleUpload 处理文件上传请求
func handleUpload(w http.ResponseWriter, r *http.Request) {
	if !checkAuth(r) {
		http.Error(w, "未授权访问", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "只支持 POST 请求", http.StatusMethodNotAllowed)
		return
	}

	// 解析 multipart form，限制最大 10MB
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "无法解析上传的文件", http.StatusBadRequest)
		return
	}

	// 获取上传的文件
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "无法获取上传的文件", http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Printf("接收到文件上传: %s, 大小: %d bytes", handler.Filename, handler.Size)

	// 读取文件内容
	content, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "无法读取文件内容", http.StatusInternalServerError)
		return
	}

	// 执行转换
	result, err := converter.ConvertToMySQL(string(content))
	if err != nil {
		http.Error(w, fmt.Sprintf("转换失败: %v", err), http.StatusInternalServerError)
		return
	}

	// 生成输出文件名
	originalName := handler.Filename
	outputName := strings.TrimSuffix(originalName, filepath.Ext(originalName)) + "_mysql.sql"

	// 设置响应头，触发文件下载
	w.Header().Set("Content-Type", "application/sql")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", outputName))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(result)))

	// 写入转换结果
	w.Write([]byte(result))

	log.Printf("文件转换成功: %s -> %s", originalName, outputName)
}

// handleStaticFile 处理静态文件请求
func handleStaticFile(w http.ResponseWriter, r *http.Request) {
	// 从URL路径中获取文件名
	filename := r.URL.Path[1:] // 移除开头的 /

	// 从嵌入的文件系统读取文件
	data, err := staticFiles.ReadFile("static/" + filename)
	if err != nil {
		log.Printf("[错误] 读取静态文件失败 %s: %v", filename, err)
		http.Error(w, "文件未找到", http.StatusNotFound)
		return
	}

	// 根据文件扩展名设置Content-Type
	contentType := "application/javascript"
	if strings.HasSuffix(filename, ".css") {
		contentType = "text/css"
	} else if strings.HasSuffix(filename, ".js") {
		contentType = "application/javascript"
	}

	w.Header().Set("Content-Type", contentType)
	w.Write(data)
}

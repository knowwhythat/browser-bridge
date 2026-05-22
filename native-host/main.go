package main

import (
	"browser-bridge/native-host/handler"
	"browser-bridge/native-host/native"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func main() {
	log.Println("[native-host] starting...")

	// 打印启动参数
	for i, arg := range os.Args {
		log.Printf("[native-host] arg[%d]: %s", i, arg)
	}

	// 初始化 Native Messaging Bridge
	bridge := native.NewBridge()

	// 启动 stdin 读取协程
	go bridge.PumpStdin()

	// 查找可用端口
	port := findAvailablePort(3000, 4000)

	// 写入端口号文件
	if err := writePortFile(port); err != nil {
		log.Fatalf("[native-host] failed to write port file: %v", err)
	}

	// 注册路由
	h := handler.NewHandler(bridge)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)

	// 启动 HTTP 服务
	addr := fmt.Sprintf("127.0.0.1:%d", port)
	server := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 3 * time.Second,
	}

	log.Printf("[native-host] HTTP server listening on %s", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("[native-host] server failed: %v", err)
	}
}

func findAvailablePort(start, end int) int {
	for port := start; port < end; port++ {
		l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			l.Close()
			return port
		}
	}
	log.Fatalf("[native-host] no available port in range %d-%d", start, end)
	return 0
}

func writePortFile(port int) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(homeDir, ".browser-bridge")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	portFile := filepath.Join(dir, "nativehost_port")
	return os.WriteFile(portFile, []byte(strconv.Itoa(port)), 0644)
}

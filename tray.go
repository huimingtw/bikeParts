package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/getlantern/systray"
)

func runTray(server *http.Server, port string) {
	// Ctrl+C / SIGTERM もtray quit と同様に処理
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		systray.Quit()
	}()

	systray.Run(onTrayReady(server, port), onTrayExit(server))
}

func onTrayReady(server *http.Server, port string) func() {
	return func() {
		systray.SetIcon(trayIcon())
		systray.SetTooltip("零件庫存系統（port " + port + "）")

		mOpen := systray.AddMenuItem("開啟瀏覽器", "在瀏覽器中開啟系統")
		systray.AddSeparator()
		mQuit := systray.AddMenuItem("結束", "關閉零件庫存系統")

		go func() {
			for {
				select {
				case <-mOpen.ClickedCh:
					openBrowser("http://localhost:" + port)
				case <-mQuit.ClickedCh:
					systray.Quit()
				}
			}
		}()

		// 啟動時自動開啟瀏覽器，並建立桌面捷徑
		go openBrowser("http://localhost:" + port)
		createDesktopShortcut(port)
	}
}

func onTrayExit(server *http.Server) func() {
	return func() {
		log.Println("正在關閉伺服器...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("伺服器強制關閉：%v", err)
		}
	}
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}

// createDesktopShortcut 在桌面建立捷徑（只建一次）
func createDesktopShortcut(port string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}

	var path, content string
	switch runtime.GOOS {
	case "windows":
		path = filepath.Join(home, "Desktop", "零件庫存系統.url")
		content = fmt.Sprintf("[InternetShortcut]\r\nURL=http://localhost:%s\r\n", port)
	case "darwin":
		path = filepath.Join(home, "Desktop", "零件庫存系統.webloc")
		content = fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict><key>URL</key><string>http://localhost:%s</string></dict></plist>`, port)
	default:
		return
	}

	if _, err := os.Stat(path); err == nil {
		return // 已存在，跳過
	}
	_ = os.WriteFile(path, []byte(content), 0o644)
}

// trayIcon 產生一個簡單的深藍色方形圖示
func trayIcon() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	bg := color.RGBA{26, 58, 92, 255} // --blue-dark
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			img.Set(x, y, bg)
		}
	}
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

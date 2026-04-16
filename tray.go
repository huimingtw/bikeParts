package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
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
	// Handle Ctrl+C / SIGTERM the same way as quitting from the tray menu.
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

		// Auto-open browser on startup and create desktop shortcut.
		go openBrowser("http://localhost:" + port)
		createDesktopShortcut(port)
	}
}

func onTrayExit(server *http.Server) func() {
	return func() {
		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Server forced to shut down: %v", err)
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

// createDesktopShortcut creates a desktop shortcut on first run only.
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
		return // already exists, skip
	}
	_ = os.WriteFile(path, []byte(content), 0o644)
}

// trayIcon returns an ICO-format icon with 16x16 and 32x32 blue circle images.
func trayIcon() []byte {
	sizes := []int{16, 32}
	var pngs [][]byte
	for _, s := range sizes {
		var buf bytes.Buffer
		_ = png.Encode(&buf, circleIcon(s))
		pngs = append(pngs, buf.Bytes())
	}
	return buildICO(pngs, sizes)
}

func circleIcon(size int) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, size, size))
	cx, cy := float64(size)/2, float64(size)/2
	r := float64(size)/2 - 0.5
	blue := color.NRGBA{30, 87, 153, 255}
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) + 0.5 - cx
			dy := float64(y) + 0.5 - cy
			if dx*dx+dy*dy <= r*r {
				img.Set(x, y, blue)
			}
		}
	}
	return img
}

// buildICO assembles an ICO file from PNG-encoded images (modern ICO format).
func buildICO(images [][]byte, sizes []int) []byte {
	n := len(images)
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, uint16(0)) // reserved
	binary.Write(&buf, binary.LittleEndian, uint16(1)) // type: ICO
	binary.Write(&buf, binary.LittleEndian, uint16(n)) // count

	offset := uint32(6 + 16*n)
	for i, img := range images {
		s := uint8(sizes[i])
		buf.WriteByte(s)                                                   // width
		buf.WriteByte(s)                                                   // height
		buf.WriteByte(0)                                                   // color count
		buf.WriteByte(0)                                                   // reserved
		binary.Write(&buf, binary.LittleEndian, uint16(1))                // planes
		binary.Write(&buf, binary.LittleEndian, uint16(32))               // bit count
		binary.Write(&buf, binary.LittleEndian, uint32(len(img)))         // size
		binary.Write(&buf, binary.LittleEndian, offset)                   // offset
		offset += uint32(len(img))
	}
	for _, img := range images {
		buf.Write(img)
	}
	return buf.Bytes()
}

// Keep math imported for potential future use in circleIcon.
var _ = math.Pi

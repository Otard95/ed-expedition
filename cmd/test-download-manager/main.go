package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"ed-expedition/download"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <url> <dest-path>\n", os.Args[0])
		os.Exit(1)
	}

	url := os.Args[1]
	destPath := os.Args[2]

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		fmt.Fprintf(os.Stderr, "Error: URL must start with http:// or https://\n")
		os.Exit(1)
	}

	if destPath == "" || strings.HasPrefix(destPath, "-") {
		fmt.Fprintf(os.Stderr, "Error: dest-path must be a valid file path\n")
		os.Exit(1)
	}

	mgr, err := download.NewManager(url, destPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating download manager: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("URL:        %s\n", url)
	fmt.Printf("Dest:       %s\n", destPath)
	fmt.Printf("Total size: %s\n", formatBytes(mgr.TotalBytes()))
	fmt.Printf("Downloaded: %s (%.1f%%)\n", formatBytes(mgr.DownloadedBytes()), percent(mgr.DownloadedBytes(), mgr.TotalBytes()))
	fmt.Println()

	if mgr.IsComplete() {
		fmt.Println("Download already complete!")
		return
	}

	fmt.Print("Press Enter to start download...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')
	fmt.Println()

	startTime := time.Now()
	lastReport := time.Now()
	lastBytes := mgr.DownloadedBytes()

	err = mgr.Download(func(downloaded int64) {
		now := time.Now()
		if now.Sub(lastReport) >= time.Second {
			elapsed := now.Sub(lastReport).Seconds()
			bytesPerSec := float64(downloaded-lastBytes) / elapsed
			pct := percent(downloaded, mgr.TotalBytes())
			fmt.Printf("\r%.1f%% (%s / %s) - %s/s    ",
				pct,
				formatBytes(downloaded),
				formatBytes(mgr.TotalBytes()),
				formatBytes(int64(bytesPerSec)))
			lastReport = now
			lastBytes = downloaded
		}
	})

	fmt.Println()

	if err != nil {
		mgr.Close()
		fmt.Fprintf(os.Stderr, "Error during download: %v\n", err)
		os.Exit(1)
	}

	elapsed := time.Since(startTime)
	avgSpeed := float64(mgr.TotalBytes()) / elapsed.Seconds()

	fmt.Println()
	fmt.Printf("Download complete!\n")
	fmt.Printf("Time:       %s\n", elapsed.Round(time.Second))
	fmt.Printf("Avg speed:  %s/s\n", formatBytes(int64(avgSpeed)))
}

func percent(current, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(current) / float64(total) * 100
}

func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Icon struct {
	Path string `json:"path"`
	Name string `json:"name"`
}

type IconsResponse struct {
	Total      int    `json:"total"`
	TotalPages int    `json:"totalPages"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	Icons      []Icon `json:"icons"`
}

var allIcons []Icon

func collectIcons() error {
	allIcons = []Icon{}
	iconDir := "icon"

	err := filepath.Walk(iconDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".png" || ext == ".svg" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" {
				allIcons = append(allIcons, Icon{
					Path: filepath.ToSlash(path),
					Name: info.Name(),
				})
			}
		}
		return nil
	})
	return err
}

func iconsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")
	search := r.URL.Query().Get("search")

	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 20
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	// 筛选图标
	var filteredIcons []Icon
	if search == "" {
		filteredIcons = allIcons
	} else {
		searchLower := strings.ToLower(search)
		for _, icon := range allIcons {
			if strings.Contains(strings.ToLower(icon.Name), searchLower) {
				filteredIcons = append(filteredIcons, icon)
			}
		}
	}

	total := len(filteredIcons)
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}

	var icons []Icon
	if start < total {
		icons = filteredIcons[start:end]
	} else {
		icons = []Icon{}
	}

	response := IconsResponse{
		Total:      total,
		TotalPages: totalPages,
		Page:       page,
		PageSize:   pageSize,
		Icons:      icons,
	}

	json.NewEncoder(w).Encode(response)
}

func main() {
	fmt.Println("Collecting icons...")
	if err := collectIcons(); err != nil {
		log.Fatal("Error collecting icons:", err)
	}
	fmt.Printf("Found %d icons\n", len(allIcons))

	http.HandleFunc("/api/icons", iconsHandler)
	// 自定义文件服务器处理函数
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 如果请求根路径，返回 index.html
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "index.html")
			return
		}
		// 尝试解码路径中的特殊字符
		decodedPath, err := url.PathUnescape(r.URL.Path)
		if err != nil {
			decodedPath = r.URL.Path
		}
		// 创建一个新的请求，使用解码后的路径
		newReq := *r
		newReq.URL = &url.URL{
			Path: decodedPath,
		}
		// 使用解码后的路径提供文件
		http.FileServer(http.Dir(".")).ServeHTTP(w, &newReq)
	})

	fmt.Println("Server starting on http://localhost:6024")
	log.Fatal(http.ListenAndServe(":6024", nil))
}

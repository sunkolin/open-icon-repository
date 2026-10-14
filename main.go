package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/mozillazg/go-pinyin"
)

type Icon struct {
	Path          string `json:"path"`
	Name          string `json:"name"`
	nameLower     string
	pinyinFull    string
	pinyinInitial string
}

type IconsResponse struct {
	Total      int    `json:"total"`
	TotalPages int    `json:"totalPages"`
	Page       int    `json:"page"`
	PageSize   int    `json:"pageSize"`
	Icons      []Icon `json:"icons"`
}

var allIcons []Icon

var suffixPattern = regexp.MustCompile(`^(.+?)-(\d+)$`)

var pinyinArgs = pinyin.NewArgs()

type overrideEntry struct {
	cn   string
	full string
}

var pinyinOverrides []overrideEntry

func init() {
	pinyinOverrides = []overrideEntry{
		{"农商行", "nong shang hang"},
		{"银行", "yin hang"},
		{"西藏", "xi zang"},
		{"音乐", "yin yue"},
		{"重庆", "chong qing"},
		{"重生", "chong sheng"},
		{"重返", "chong fan"},
		{"重卡", "zhong ka"},
		{"长安", "chang an"},
		{"长沙", "chang sha"},
		{"长城", "chang cheng"},
		{"长江", "chang jiang"},
		{"长荣", "chang rong"},
		{"长龙", "chang long"},
		{"长佩", "chang pei"},
		{"首都", "shou du"},
	}
	sort.Slice(pinyinOverrides, func(i, j int) bool {
		return len([]rune(pinyinOverrides[i].cn)) > len([]rune(pinyinOverrides[j].cn))
	})
}

func buildPinyin(text string) (full, initial string) {
	runes := []rune(text)
	var fullBuilder, initialBuilder strings.Builder
	i := 0
	for i < len(runes) {
		matched := false
		for _, entry := range pinyinOverrides {
			ocr := []rune(entry.cn)
			if i+len(ocr) <= len(runes) && string(runes[i:i+len(ocr)]) == entry.cn {
				parts := strings.Split(entry.full, " ")
				for _, p := range parts {
					p = strings.ToLower(p)
					fullBuilder.WriteString(p)
					if len(p) > 0 {
						initialBuilder.WriteByte(p[0])
					}
				}
				i += len(ocr)
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		r := runes[i]
		if r >= 0x4E00 && r <= 0x9FFF {
			py := pinyin.SinglePinyin(r, pinyinArgs)
			if len(py) > 0 {
				fullBuilder.WriteString(py[0])
				initialBuilder.WriteString(string(py[0][0]))
			}
		} else {
			fullBuilder.WriteRune(r)
			initialBuilder.WriteRune(r)
		}
		i++
	}
	full = strings.ToLower(strings.ReplaceAll(fullBuilder.String(), " ", ""))
	initial = strings.ToLower(strings.ReplaceAll(initialBuilder.String(), " ", ""))
	return
}

func parseFileName(name string) (baseName string, suffix int, hasSuffix bool) {
	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)

	if m := suffixPattern.FindStringSubmatch(base); m != nil {
		num, err := strconv.Atoi(m[2])
		if err == nil {
			return m[1], num, true
		}
	}
	return base, 0, false
}

func sortIcons(icons []Icon) {
	sort.Slice(icons, func(i, j int) bool {
		iBase, iSuffix, iHasSuffix := parseFileName(icons[i].Name)
		jBase, jSuffix, jHasSuffix := parseFileName(icons[j].Name)

		if iBase != jBase {
			return iBase < jBase
		}

		if !iHasSuffix && jHasSuffix {
			return true
		}
		if iHasSuffix && !jHasSuffix {
			return false
		}
		if !iHasSuffix && !jHasSuffix {
			return icons[i].Name < icons[j].Name
		}

		return iSuffix < jSuffix
	})
}

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
				baseName := strings.TrimSuffix(info.Name(), filepath.Ext(info.Name()))
				full, initial := buildPinyin(baseName)
				allIcons = append(allIcons, Icon{
					Path:          filepath.ToSlash(path),
					Name:          info.Name(),
					nameLower:     strings.ToLower(info.Name()),
					pinyinFull:    full,
					pinyinInitial: initial,
				})
			}
		}
		return nil
	})

	if err == nil {
		sortIcons(allIcons)
	}

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
			if strings.Contains(icon.nameLower, searchLower) ||
				strings.Contains(icon.pinyinFull, searchLower) ||
				strings.Contains(icon.pinyinInitial, searchLower) {
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
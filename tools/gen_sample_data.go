package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// TorrentFile 表示种子中的单个文件
type TorrentFile struct {
	Path   string `json:"path"`
	Length int64  `json:"length"`
}

// TorrentMetadata 表示一个种子的元数据
type TorrentMetadata struct {
	InfoHash   string        `json:"info_hash"`
	Name       string        `json:"name"`
	Files      []TorrentFile `json:"files"`
	TotalSize  int64         `json:"total_size"`
	FileCount  int           `json:"file_count"`
	CreateDate time.Time     `json:"create_date"`
}

var (
	mediaExtensions = []string{
		".mp4", ".mkv", ".avi", ".mov", ".flv", ".wmv", ".m4v",
		".mp3", ".wav", ".flac", ".aac", ".ogg",
	}

	documentExtensions = []string{
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".txt", ".zip", ".rar",
	}

	imageExtensions = []string{
		".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp",
	}

	softwareExtensions = []string{
		".exe", ".dmg", ".apk", ".iso", ".deb", ".rpm",
	}

	// 种子类型和对应的名字模板
	torrentTypes = map[string][]string{
		"电影": {
			"[高清] %s (2020-2025).%s",
			"%s.2025.%s.1080p.BluRay.x264",
			"%s.2024.CHINESE.WEBRip.4K.%s",
			"%s.%s.HDR10Plus.官方中字",
			"【漫游字幕组】%s.v2.%s.简繁内封",
		},
		"电视剧": {
			"%s.S01E01-E10.Complete.%s.1080p",
			"%s.全集.2024.%s.国语中字",
			"【追新番】%s 第一季 全12集 %s",
			"%s (2024) 完整季 [1-12].%s.双语字幕",
			"%s.Season.01.Complete.1080p.%s.中英双字",
		},
		"音乐": {
			"%s - %s全集 (320Kbps)",
			"%s - %s [FLAC+CUE]",
			"[2024] %s - %s [MP3-320K]",
			"%s - %s [无损WAV+CUE]",
			"%s - %s (2025) [FLAC]",
		},
		"软件": {
			"%s.v2025.%s.中文特别版",
			"%s.官方中文版.%s.绿色版",
			"%s 2024 %s 激活版",
			"%s企业版.%s.完整破解版",
			"%s.官方原版.%s.便携版",
		},
		"游戏": {
			"%s.中文豪华版.%s",
			"%s.全DLC终极收藏版.%s.绿色中文版",
			"%s.官方中文.%s.免安装版",
			"%s.全解锁存档.%s.单机版",
			"%s.完全版.%s.v2024",
		},
	}

	// 常见电影名/软件名等
	commonNames = map[string][]string{
		"电影": {
			"流浪地球", "长津湖", "你好，李焕英", "狂飙", "熊出没", "封神第一部", "满江红",
			"独行月球", "四海", "这个杀手不太冷静", "奇迹", "我和我的祖国", "哥斯拉大战金刚",
			"蜘蛛侠", "复仇者联盟", "速度与激情", "碟中谍", "变形金刚", "星球大战", "哈利波特",
			"指环王", "加勒比海盗", "终结者", "异形", "黑客帝国", "头号玩家", "移动迷宫",
		},
		"电视剧": {
			"三体", "赘婿", "山河令", "觉醒年代", "扫黑风暴", "御赐小仵作", "甄嬛传",
			"庆余年", "琅琊榜", "知否知否", "大明风华", "鹿鼎记", "隐秘的角落", "有翡",
			"武林外传", "神探狄仁杰", "白鹿原", "大宋少年志", "天龙八部", "西游记", "大秦帝国",
		},
		"音乐": {
			"周杰伦", "林俊杰", "邓紫棋", "五月天", "薛之谦", "华晨宇", "张杰",
			"Beyond", "陈奕迅", "王菲", "李荣浩", "刘德华", "孙燕姿", "田馥甄",
			"朴树", "TF Boys", "李宇春", "周深", "毛不易", "TFBOYS", "李健",
		},
		"软件": {
			"Adobe Photoshop", "Microsoft Office", "AutoCAD", "Premiere Pro", "Visual Studio",
			"Windows 11", "CorelDRAW", "After Effects", "Illustrator", "Maya", "3ds Max",
			"Final Cut Pro", "Logic Pro", "Pro Tools", "Unity", "Unreal Engine", "Blender",
			"光影魔术手", "会声会影", "WPS", "迅雷", "爱奇艺", "腾讯视频", "优酷",
		},
		"游戏": {
			"王者荣耀", "英雄联盟", "原神", "绝地求生", "我的世界", "战神", "只狼",
			"赛博朋克2077", "艾尔登法环", "地平线", "质量效应", "上古卷轴", "文明", "模拟人生",
			"极品飞车", "使命召唤", "魔兽世界", "穿越火线", "地下城与勇士", "梦幻西游", "天涯明月刀",
			"剑网3", "阴阳师", "第五人格", "和平精英", "荒野行动", "明日方舟", "万国觉醒",
		},
	}
)

// 随机生成种子名和类别
func generateRandomTorrent() TorrentMetadata {
	// 随机选择种子类别
	categories := make([]string, 0, len(torrentTypes))
	for k := range torrentTypes {
		categories = append(categories, k)
	}
	category := categories[rand.Intn(len(categories))]

	// 随机选择名称模板和名称
	templates := torrentTypes[category]
	names := commonNames[category]

	template := templates[rand.Intn(len(templates))]
	name := names[rand.Intn(len(names))]

	// 确定扩展名
	var extensions []string
	switch category {
	case "电影", "电视剧":
		extensions = mediaExtensions[:7] // 视频扩展名
	case "音乐":
		extensions = mediaExtensions[7:] // 音频扩展名
	case "软件":
		extensions = softwareExtensions
	case "游戏":
		extensions = append(softwareExtensions, ".iso")
	default:
		extensions = mediaExtensions
	}

	ext := extensions[rand.Intn(len(extensions))]

	// 生成种子名称 (在模板中可能有两个占位符 %s)
	var torrentName string
	if strings.Count(template, "%s") > 1 {
		// 如果模板有两个%s，第二个用于扩展名或其他标识
		randWord := []string{
			"HDR", "4K", "x264", "官方版", "终极版", "完整版", "精选集",
			"华语版", "国际版", "双语版", "特别版", "珍藏版", "豪华版", "超清版",
		}
		torrentName = fmt.Sprintf(template, name, randWord[rand.Intn(len(randWord))])
	} else {
		torrentName = fmt.Sprintf(template, name)
	}

	// 生成文件列表
	fileCount := rand.Intn(10) + 1 // 1-10个文件
	files := make([]TorrentFile, 0, fileCount)
	totalSize := int64(0)

	for i := 0; i < fileCount; i++ {
		var filename string
		var fileSize int64

		// 根据种子类别生成不同类型的文件
		switch category {
		case "电影":
			if i == 0 {
				filename = fmt.Sprintf("movie%s", ext)
				fileSize = rand.Int63n(10000000000) + 500000000 // 500MB-10.5GB
			} else {
				// 电影可能有额外的样本视频或字幕文件
				if rand.Intn(2) == 0 {
					filename = fmt.Sprintf("sample%s", ext)
					fileSize = rand.Int63n(100000000) + 10000000 // 10MB-110MB
				} else {
					filename = fmt.Sprintf("subtitle_%d.srt", i)
					fileSize = rand.Int63n(1000000) + 10000 // 10KB-1MB
				}
			}
		case "电视剧":
			episodeNum := i + 1
			filename = fmt.Sprintf("Episode.%02d%s", episodeNum, ext)
			fileSize = rand.Int63n(5000000000) + 300000000 // 300MB-5.3GB
		case "音乐":
			filename = fmt.Sprintf("Track_%02d%s", i+1, ext)
			fileSize = rand.Int63n(50000000) + 5000000 // 5MB-55MB
		case "软件":
			if i == 0 {
				filename = fmt.Sprintf("setup%s", ext)
				fileSize = rand.Int63n(5000000000) + 100000000 // 100MB-5.1GB
			} else {
				// 软件可能有readme或crack文件
				if rand.Intn(2) == 0 {
					filename = "readme.txt"
					fileSize = rand.Int63n(100000) + 1000 // 1KB-101KB
				} else {
					filename = "crack.exe"
					fileSize = rand.Int63n(10000000) + 100000 // 100KB-10.1MB
				}
			}
		case "游戏":
			if i == 0 {
				filename = fmt.Sprintf("game%s", ext)
				fileSize = rand.Int63n(50000000000) + 1000000000 // 1GB-51GB
			} else {
				// 游戏可能有不同的文件
				fileTypes := []string{
					"data.bin", "textures.pak", "audio.pak", "video.mp4",
					"readme.txt", "patch.exe", "update.bin", "dlc.pak",
				}
				filename = fileTypes[rand.Intn(len(fileTypes))]
				fileSize = rand.Int63n(2000000000) + 50000000 // 50MB-2.05GB
			}
		default:
			filename = fmt.Sprintf("file_%d%s", i, ext)
			fileSize = rand.Int63n(1000000000) + 10000000 // 10MB-1.01GB
		}

		files = append(files, TorrentFile{
			Path:   filename,
			Length: fileSize,
		})
		totalSize += fileSize
	}

	// 创建日期为过去3年内的随机时间
	daysAgo := rand.Intn(365 * 3) // 0-3年的随机天数
	createDate := time.Now().AddDate(0, 0, -daysAgo)

	// 生成info_hash (使用种子名和当前时间的哈希值)
	hashInput := fmt.Sprintf("%s-%d-%d", torrentName, totalSize, createDate.UnixNano())
	hash := sha1.Sum([]byte(hashInput))
	infoHash := hex.EncodeToString(hash[:])

	return TorrentMetadata{
		InfoHash:   infoHash,
		Name:       torrentName,
		Files:      files,
		TotalSize:  totalSize,
		FileCount:  len(files),
		CreateDate: createDate,
	}
}

func main() {
	// Go 1.20+ 的标准库已默认初始化随机数种子
	// 生成的种子数量
	numTorrents := 500

	fmt.Printf("开始生成 %d 个测试种子数据\n", numTorrents)

	// 批量生成并插入
	batchSize := 50
	for i := 0; i < numTorrents; i += batchSize {
		var buf bytes.Buffer

		// 确定此批次的实际大小
		currentBatchSize := batchSize
		if i+batchSize > numTorrents {
			currentBatchSize = numTorrents - i
		}

		// 生成批量插入命令
		for j := 0; j < currentBatchSize; j++ {
			torrent := generateRandomTorrent()

			// 创建索引命令
			indexCmd := map[string]interface{}{
				"index": map[string]interface{}{
					"_index": "bittorrent_metadata",
					"_id":    torrent.InfoHash,
				},
			}

			indexJSON, _ := json.Marshal(indexCmd)
			buf.Write(indexJSON)
			buf.WriteString("\n")

			// 添加文档内容
			docJSON, _ := json.Marshal(torrent)
			buf.Write(docJSON)
			buf.WriteString("\n")
		}

		// 发送批量请求
		req, err := http.NewRequest("POST", "http://localhost:9200/_bulk", &buf)
		if err != nil {
			fmt.Printf("创建请求失败: %v\n", err)
			continue
		}

		req.Header.Set("Content-Type", "application/x-ndjson")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("发送请求失败: %v\n", err)
			continue
		}

		// 读取并丢弃响应体
		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		fmt.Printf("批次 %d/%d 完成\n", (i/batchSize)+1, (numTorrents+batchSize-1)/batchSize)
	}

	fmt.Println("所有测试数据生成完毕")
}

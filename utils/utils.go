package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"traktshow/trakt"
	"golang.org/x/oauth2"
)

// 令牌存储文件路径
const tokenFileName = ".trakt-access-token.json"

// SaveToken 持久化存储访问令牌（避免重复授权）
func SaveToken(token *oauth2.Token) error {
	tokenPath := getTokenPath()
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化令牌失败：%v", err)
	}
	return os.WriteFile(tokenPath, data, 0600)
}

// LoadToken 加载已保存的访问令牌（存在且未过期则直接使用）
func LoadToken() (*oauth2.Token, error) {
	tokenPath := getTokenPath()
	if _, err := os.Stat(tokenPath); err != nil {
		return nil, fmt.Errorf("令牌文件不存在：%v", err)
	}

	data, err := os.ReadFile(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("读取令牌文件失败：%v", err)
	}

	var token oauth2.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("解析令牌文件失败：%v", err)
	}

	// 检查令牌是否过期（过期则返回错误，需要重新授权）
	if time.Now().After(token.Expiry) {
		return nil, fmt.Errorf("令牌已过期，请重新授权")
	}

	log.Println("令牌加载成功（未过期）")
	return &token, nil
}

// PrintUserInfo 格式化打印用户基本信息
func PrintUserInfo(info *trakt.TraktUserInfo) {
	fmt.Printf("\n===== Trakt 用户基本信息 =====\n")
	fmt.Printf("用户名：%s\n", info.Username)
	fmt.Printf("姓名：%s\n", info.Name)
	fmt.Printf("注册时间：%s\n", info.JoinedAt)
	fmt.Printf("所在地：%s\n", info.Location)
	fmt.Printf("已观看电影：%d 部\n", info.Stats.Movies.Watched)
	fmt.Printf("已观看剧集：%d 部\n", info.Stats.Shows.Watched)
	fmt.Printf("=============================\n")
}

// PrintWatchHistory 格式化打印观看记录
func PrintWatchHistory(history []trakt.TraktWatchHistoryItem) {
	fmt.Printf("\n===== 最近 %d 条观看记录 =====\n", len(history))
	for i, item := range history {
		fmt.Printf("\n【第 %d 条】\n", i+1)
		fmt.Printf("观看时间：%s\n", item.WatchedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("动作：%s\n", getActionChineseName(item.Action))
		fmt.Printf("类型：%s\n", getTypeChineseName(item.Type))

		if item.Type == "episode" && item.Show != nil {
			fmt.Printf("\n===== 剧集信息 =====\n")
			fmt.Printf("中文剧名：%s\n", item.Show.OriginalTitle)
			fmt.Printf("英文剧名：%s\n", item.Show.Title)
			fmt.Printf("发布年份：%d\n", item.Show.Year)
			if item.Show.Overview != "" {
				fmt.Printf("简介：%s\n", item.Show.Overview)
			}
			if len(item.Show.Genres) > 0 {
				fmt.Printf("类型/标签：%s\n", strings.Join(item.Show.Genres, ", "))
			}
			fmt.Printf("状态：%s\n", item.Show.Status)
			fmt.Printf("评分：%.2f (%d票)\n", item.Show.Rating, item.Show.Votes)
			if item.Show.Network != "" {
				fmt.Printf("播出网络：%s\n", item.Show.Network)
			}
			if item.Show.Language != "" {
				fmt.Printf("语言：%s\n", item.Show.Language)
			}
			if item.Show.FirstAired != "" {
				fmt.Printf("首播日期：%s\n", item.Show.FirstAired)
			}
			if item.Show.Runtime > 0 {
				fmt.Printf("单集时长：%d分钟\n", item.Show.Runtime)
			}
			fmt.Printf("总集数：%d集\n", item.Show.AiredEpisodes)
			fmt.Printf("播放时间：每周%s %s (%s)\n", item.Show.Airs.Day, item.Show.Airs.Time, item.Show.Airs.Timezone)
			if item.Show.Homepage != "" {
				fmt.Printf("主页：%s\n", item.Show.Homepage)
			}
			fmt.Printf("认证级别：%s\n", item.Show.Certification)
			fmt.Printf("出品国家：%s\n", item.Show.Country)
			fmt.Printf("Trakt ID：%d\n", item.Show.IDs.Trakt)
			fmt.Printf("IMDB ID：%s\n", item.Show.IDs.IMDB)
			fmt.Printf("TMDb ID：%d\n", item.Show.IDs.TMDb)
			
			if item.Show.Episode != nil {
				fmt.Printf("\n----- 单集信息 -----\n")
				fmt.Printf("第 %d 季 第 %d 集\n", item.Show.Episode.Season, item.Show.Episode.Number)
				fmt.Printf("集数标题：%s\n", item.Show.Episode.Title)
				fmt.Printf("中文集名：%s\n", item.Show.Episode.OriginalTitle)
				fmt.Printf("集数编码：S%02dE%02d\n", item.Show.Episode.Season, item.Show.Episode.Number)
				if item.Show.Episode.Overview != "" {
					fmt.Printf("集数简介：%s\n", item.Show.Episode.Overview)
				}
				fmt.Printf("集数评分：%.2f (%d票)\n", item.Show.Episode.Rating, item.Show.Episode.Votes)
				if item.Show.Episode.Runtime > 0 {
					fmt.Printf("集数时长：%d分钟\n", item.Show.Episode.Runtime)
				}
				if item.Show.Episode.FirstAired != "" {
					fmt.Printf("首播日期：%s\n", item.Show.Episode.FirstAired)
				}
				fmt.Printf("单集Trakt ID：%.0f\n", item.Show.Episode.IDs.Trakt)
				fmt.Printf("单集IMDB ID：%s\n", item.Show.Episode.IDs.IMDB)
				fmt.Printf("单集TMDb ID：%.0f\n", item.Show.Episode.IDs.TMDb)
			}
			fmt.Printf("====================\n")
		} else if item.Type == "movie" && item.Movie != nil {
			fmt.Printf("\n===== 电影信息 =====\n")
			fmt.Printf("标题：\n") // 暂时不处理
			fmt.Printf("====================\n")
		}
	}
	fmt.Printf("=============================\n")
}

// getTypeChineseName 获取类型的中文翻译
func getTypeChineseName(itemType string) string {
	switch itemType {
	case "movie":
		return "电影"
	case "show":
		return "剧集"
	case "episode":
		return "单集"
	default:
		return itemType
	}
}

// getActionChineseName 获取动作类型的中文翻译
func getActionChineseName(action string) string {
	switch action {
	case "watch":
		return "观看"
	case "scrobble":
		return "记录"
	case "checkin":
		return "签到"
	default:
		return action
	}
}

// 获取令牌文件路径（与配置文件同目录）
func getTokenPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Panicf("获取用户主目录失败：%v", err)
	}
	return filepath.Join(homeDir, tokenFileName)
}
package trakt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"traktshow/config"
	"golang.org/x/oauth2"
)

const traktAPIEndpoint = "https://api.trakt.tv"

// TraktUserInfo 用户基本信息结构体
type TraktUserInfo struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	JoinedAt string `json:"joined_at"`
	Location string `json:"location"`
	Stats    struct {
		Movies struct {
			Watched int `json:"watched"`
		} `json:"movies"`
		Shows struct {
			Watched int `json:"watched"`
		} `json:"shows"`
	} `json:"stats"`
}

// TraktWatchHistoryItem 观看记录结构体
type TraktWatchHistoryItem struct {
	WatchedAt time.Time `json:"watched_at"`
	Action    string    `json:"action"`
	Type      string    `json:"type"`
	Show      *struct {
		AiredEpisodes int      `json:"aired_episodes"`
		Airs          struct {
			Day     string `json:"day"`
			Time    string `json:"time"`
			Timezone string `json:"timezone"`
		} `json:"airs"`
		AvailableTranslations []string `json:"available_translations"`
		Certification         string   `json:"certification"`
		CommentCount          int      `json:"comment_count"`
		Country               string   `json:"country"`
		FirstAired            string   `json:"first_aired"`
		Genres                []string `json:"genres"`
		Homepage              string   `json:"homepage"`
		IDs                   struct {
			IMDB  string `json:"imdb"`
			Slug  string `json:"slug"`
			TMDb  int    `json:"tmdb"`
			Trakt int    `json:"trakt"`
			TVDB  int    `json:"tvdb"`
			TVRage int   `json:"tvrage"`
		} `json:"ids"`
		Language      string   `json:"language"`
		Languages     []string `json:"languages"`
		Network       string   `json:"network"`
		OriginalTitle string   `json:"original_title"`
		Overview      string   `json:"overview"`
		Rating        float64  `json:"rating"`
		Runtime       int      `json:"runtime"`
		Status        string   `json:"status"`
		Subgenres     []string `json:"subgenres"`
		Tagline       string   `json:"tagline"`
		Title         string   `json:"title"`
		Trailer       string   `json:"trailer"`
		UpdatedAt     string   `json:"updated_at"`
		Votes         int      `json:"votes"`
		Year          int      `json:"year"`
		Episode       *struct {
			AfterCredits          bool      `json:"after_credits"`
			AvailableTranslations []string  `json:"available_translations"`
			CommentCount          int       `json:"comment_count"`
			DuringCredits         bool      `json:"during_credits"`
			EpisodeType           string    `json:"episode_type"`
			FirstAired            string    `json:"first_aired"`
			IDs                   struct {
				IMDB   string  `json:"imdb"`
				TMDb   float64 `json:"tmdb"`
				Trakt  float64 `json:"trakt"`
				TVDB   float64 `json:"tvdb"`
				TVRage float64 `json:"tvrage"`
			} `json:"ids"`
			Number       int    `json:"number"`
			NumberAbs    interface{} `json:"number_abs"`
			OriginalTitle string `json:"original_title"`
			Overview     string `json:"overview"`
			Rating       float64 `json:"rating"`
			Runtime      int    `json:"runtime"`
			Season       int    `json:"season"`
			Title        string `json:"title"`
			UpdatedAt    string `json:"updated_at"`
			Votes        int    `json:"votes"`
		} `json:"episode"`
	} `json:"show"`
	Movie *struct {
		// 暂时不处理电影，因为您的数据中没有电影
	} `json:"movie"`
}

// GetOAuthConfig 获取OAuth2基础配置（仅用于生成授权URL）
func GetOAuthConfig() *oauth2.Config {
	cfg := config.Get()
	return &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURI,
		Scopes:       []string{"public", "user"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  traktAPIEndpoint + "/oauth/authorize",
			TokenURL: traktAPIEndpoint + "/oauth/token",
		},
	}
}

// ExchangeTokenManual 手动交换令牌（解决Trakt OAuth参数要求）
func ExchangeTokenManual(code string) (*oauth2.Token, error) {
	cfg := config.Get()
	tokenURL := traktAPIEndpoint + "/oauth/token"

	// 构造Trakt要求的完整请求体
	requestBody := map[string]string{
		"code":          code,
		"client_id":     cfg.ClientID,
		"client_secret": cfg.ClientSecret,
		"redirect_uri":  cfg.RedirectURI,
		"grant_type":    "authorization_code",
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("构造请求体失败：%v", err)
	}

	// 发送POST请求交换令牌
	resp, err := http.Post(tokenURL, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("发送令牌请求失败：%v", err)
	}
	defer resp.Body.Close()

	// 读取响应体到字节切片（可重复使用）
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取令牌响应失败：%v", err)
	}

	// 解析响应体（用于调试）
	var respBody map[string]interface{}
	if err := json.Unmarshal(respBodyBytes, &respBody); err != nil {
		return nil, fmt.Errorf("解析令牌响应失败：%v（响应内容：%s）", err, string(respBodyBytes))
	}

	// 打印响应详情（辅助排查）
	fmt.Printf("[调试] 令牌交换响应 - 状态码：%d，内容：%v\n", resp.StatusCode, respBody)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("令牌交换失败：状态码 %d，错误信息：%v", resp.StatusCode, respBody["error"])
	}

	// 构造oauth2.Token对象
	var token oauth2.Token
	token.AccessToken = respBody["access_token"].(string)
	token.TokenType = respBody["token_type"].(string)
	token.Expiry = time.Now().Add(time.Duration(respBody["expires_in"].(float64)) * time.Second)
	if rt, ok := respBody["refresh_token"].(string); ok {
		token.RefreshToken = rt
	}

	return &token, nil
}

// GetUserInfo 获取用户基本信息（带令牌认证）
func GetUserInfo(token *oauth2.Token) (*TraktUserInfo, error) {
	cfg := config.Get()
	url := traktAPIEndpoint + "/users/me"

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败：%v", err)
	}

	// 设置必需请求头（包含令牌认证）
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("trakt-api-version", "2")
	req.Header.Set("trakt-api-key", cfg.ClientID)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败：%v", err)
	}
	defer resp.Body.Close()

	// 读取响应体到字节切片（可重复使用）
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取用户信息响应失败：%v", err)
	}

	// 解析响应体（用于调试）
	var respBody map[string]interface{}
	json.Unmarshal(respBodyBytes, &respBody) // 忽略错误，仅用于调试
	fmt.Printf("[调试] 用户信息响应 - 状态码：%d，内容：%v\n", resp.StatusCode, respBody)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败：状态码 %d，错误信息：%v（响应内容：%s）", resp.StatusCode, respBody["error"], string(respBodyBytes))
	}

	// 解析为结构体
	var info TraktUserInfo
	if err := json.Unmarshal(respBodyBytes, &info); err != nil {
		return nil, fmt.Errorf("解析用户信息响应失败：%v（响应内容：%s）", err, string(respBodyBytes))
	}

	return &info, nil
}

// GetWatchHistory 获取用户观看记录（带令牌认证）
func GetWatchHistory(token *oauth2.Token, limit int) ([]TraktWatchHistoryItem, error) {
	cfg := config.Get()
	// 添加 extended 参数获取完整信息
	url := fmt.Sprintf("%s/users/me/history?limit=%d&extended=full", traktAPIEndpoint, limit)

	// 创建请求
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败：%v", err)
	}

	// 设置必需请求头（包含令牌认证）
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("trakt-api-version", "2")
	req.Header.Set("trakt-api-key", cfg.ClientID)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken))

	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败：%v", err)
	}
	defer resp.Body.Close()

	// 读取响应体到字节切片（可重复使用）
	respBodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取观看记录响应失败：%v", err)
	}

	// 解析响应体（用于调试）
	var respBody interface{}
	json.Unmarshal(respBodyBytes, &respBody) // 忽略错误，仅用于调试
	fmt.Printf("[调试] 观看记录响应 - 状态码：%d，内容：%v\n", resp.StatusCode, respBody)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API请求失败：状态码 %d，错误信息：%v（响应内容：%s）", resp.StatusCode, respBody, string(respBodyBytes))
	}

	// 解析为结构体
	var history []TraktWatchHistoryItem
	if err := json.Unmarshal(respBodyBytes, &history); err != nil {
		return nil, fmt.Errorf("解析观看记录响应失败：%v（响应内容：%s）", err, string(respBodyBytes))
	}

	return history, nil
}
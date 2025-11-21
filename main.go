package main

import (
	"fmt"
	"log"
	"os"

	"traktshow/config"
	"traktshow/trakt"
	"traktshow/utils"
	"golang.org/x/oauth2"
)

var accessToken *oauth2.Token

func main() {
	// 1. 初始化配置（自动使用 8081 端口回调地址）
	if err := config.Init(); err != nil {
		log.Fatalf("配置初始化失败：%v", err)
	}

	// 2. 尝试加载已保存的令牌（有则直接使用）
	token, err := utils.LoadToken()
	if err == nil {
		accessToken = token
		log.Println("使用已保存的令牌直接查询数据...")
		fetchAndPrintData()
		return
	}

	// 3. 手动授权流程（核心：无需本地服务，彻底绕开端口占用）
	log.Println("\n===== Trakt 手动授权流程 =====")
	// 生成授权URL（回调地址已自动为 8081）
	authURL := trakt.GetOAuthConfig().AuthCodeURL("state-random-123", oauth2.AccessTypeOffline)
	fmt.Printf("1. 请复制以下URL到浏览器打开：\n%s\n", authURL)
	fmt.Println("\n2. 浏览器中完成以下操作：")
	fmt.Println("   - 用你的Google账号登录Trakt")
	fmt.Println("   - 登录成功后，点击「Allow」（允许应用访问你的信息）")
	fmt.Println("   - 授权成功后，查看浏览器地址栏的URL")
	fmt.Println("   - 复制 URL 中「code=」后面的字符串（到「&state=」前结束，示例：code=abc123 → 复制 abc123）")

	// 手动输入授权码
	var code string
	fmt.Print("\n3. 请粘贴复制的授权码：")
	if _, err := fmt.Scanln(&code); err != nil {
		log.Fatalf("输入授权码失败：%v", err)
	}

	// 4. 手动交换令牌（核心步骤，无依赖本地服务）
	log.Printf("正在交换访问令牌...（授权码：%s）", code)
	token, err = trakt.ExchangeTokenManual(code)
	if err != nil {
		log.Fatalf("令牌交换失败：%v", err)
	}
	accessToken = token
	log.Println("✅ 令牌交换成功！")

	// 5. 保存令牌（下次无需重复授权）
	if err := utils.SaveToken(token); err != nil {
		log.Printf("⚠️  令牌保存失败：%v（不影响本次使用）", err)
	} else {
		log.Println("✅ 令牌已保存，下次运行直接使用")
	}

	// 6. 查询并打印用户数据
	fetchAndPrintData()
}

// fetchAndPrintData 查询用户信息和观看记录并打印
func fetchAndPrintData() {
	log.Println("\n===== 开始查询数据 =====")

	// 获取用户基本信息
	userInfo, err := trakt.GetUserInfo(accessToken)
	if err != nil {
		log.Fatalf("❌ 获取用户信息失败：%v", err)
	}
	utils.PrintUserInfo(userInfo)

	// 获取最近10条观看记录
	watchHistory, err := trakt.GetWatchHistory(accessToken, 100)
	if err != nil {
		log.Fatalf("❌ 获取观看记录失败：%v", err)
	}
	utils.PrintWatchHistory(watchHistory)

	log.Println("\n🎉 所有数据查询完成！")
	os.Exit(0)
}
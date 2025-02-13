package tests

import (
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	log "github.com/sirupsen/logrus"
	"strings"
	"testing"
	"time"
)

func Test_login(t *testing.T) {
	var err error

	// 启动并连接到浏览器
	browser := rod.New().ControlURL(launcher.New().Headless(false).MustLaunch()).MustConnect()

	// 打开新页面
	page := browser.MustPage("https://sign.cscec8b.com.cn:9280/")

	// 等待页面加载完成
	page.MustWaitLoad()

	time.Sleep(1 * time.Second)

	// 查找用户名输入框并填充文本 "123456"
	usernameInput := page.MustElementX("//*[@name='username']")
	usernameInput.MustInput("13012567125")
	if err != nil {
		log.Printf("failed to execute click via JavaScript: %v", err)
		return
	}

	time.Sleep(1 * time.Second)

	// 查找密码输入框并填充文本 "123456"（如果适用）
	passwordInput := page.MustElementX("//*[@type='password']")
	passwordInput.MustInput("Chint.wmq2024c")
	if err != nil {
		log.Printf("failed to execute click via JavaScript: %v", err)
		return
	}

	// 查找登录按钮并点击
	loginButton := page.MustElementX("//*[@id='app']/div/div[1]/div/div[2]/button")
	//loginButton.MustClick()
	_, err = loginButton.Eval(`(element) => {
		const el = document.querySelector(element.Object.description);
		el.click();
		return true;
	}`, loginButton)
	if err != nil {
		log.Printf("failed to execute click via JavaScript: %v", err)
		return
	}

	time.Sleep(1 * time.Second)

	var body string
	body, err = page.HTML()
	if err != nil {
		log.Printf("failed to get html: %v", err)
		return
	}

	if strings.Contains(body, "class=\"app-main\"") {
		fmt.Println("login success.")
	} else {
		fmt.Println("login failed.")
	}

	time.Sleep(100 * time.Second)

	// 等待一段时间以查看结果（可选）
	browser.Timeout(100)

	// 关闭浏览器
	browser.MustClose()

}

func Test_get_username_input(t *testing.T) {
	var err error

	// 启动并连接到浏览器
	browser := rod.New().ControlURL(launcher.New().Headless(false).Set("ignore-certificate-errors").MustLaunch()).MustConnect()

	// 打开新页面
	//page := browser.MustPage("https://app.jxwdyf.com:8130/")
	page := browser.MustPage("https://mail.baoiem.com/")

	// 等待页面加载完成
	page.MustWaitLoad()

	time.Sleep(5 * time.Second)

	// 查找用户名输入框并填充文本 "123456"
	usernameInput, err := page.ElementX("//*[@id=\"uid\"]")
	if err != nil {
		log.Fatal(err)
		return
	}

	err = usernameInput.Input("admin")
	if err != nil {
		log.Fatal(err)
		return
	}

	time.Sleep(100 * time.Second)

	// 关闭浏览器
	browser.MustClose()

}

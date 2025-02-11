package tests

import (
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	log "github.com/sirupsen/logrus"
	"testing"
	"time"
)

func Test_login(t *testing.T) {
	var err error

	// 启动并连接到浏览器
	browser := rod.New().ControlURL(launcher.New().Headless(false).MustLaunch()).MustConnect()

	// 打开新页面
	page := browser.MustPage("http://119.3.42.40:9081/")

	// 等待页面加载完成
	page.MustWaitLoad()

	time.Sleep(5 * time.Second)

	// 查找用户名输入框并填充文本 "123456"
	usernameInput := page.MustElementX("/html/body/div/div[2]/div/div/div/div/div[2]/div/form/ul/li[2]/div[1]/div/div/div[1]/div/div[2]/div/input")
	//usernameInput.MustInput("admin")
	_, err = usernameInput.Eval(`(element,value) => {
		const el = document.querySelector(element.Object.description);
		el.value = value;
		el.dispatchEvent(new Event('input', { bubbles: true })); 
		el.dispatchEvent(new Event('change', { bubbles: true }));
		return true;
	}`, usernameInput, "limin")
	if err != nil {
		log.Printf("failed to execute click via JavaScript: %v", err)
		return
	}

	// 查找密码输入框并填充文本 "123456"（如果适用）
	passwordInput := page.MustElementX("/html/body/div/div[2]/div/div/div/div/div[2]/div/form/ul/li[2]/div[1]/div/div/div[2]/div/div[2]/div/input")
	//passwordInput.MustInput("123456")
	_, err = passwordInput.Eval(`(element,value) => {
		const el = document.querySelector(element.Object.description);
		el.value = value;
		el.dispatchEvent(new Event('input', { bubbles: true })); 
		el.dispatchEvent(new Event('change', { bubbles: true }));
		return true;
	}`, passwordInput, "123456")
	if err != nil {
		log.Printf("failed to execute click via JavaScript: %v", err)
		return
	}

	// 查找登录按钮并点击
	loginButton := page.MustElementX("/html/body/div/div[2]/div/div/div/div/div[2]/div/form/input[2]")
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

	// 查找登录按钮并点击
	// 执行 JavaScript 设置复选框状态
	//_, err = page.Eval(`() => {
	//	const checkbox = document.querySelector('input[type="checkbox"]');
	//	if (checkbox) {
	//		checkbox.checked = true;
	//		return true;
	//	}
	//	return false;
	//}`)
	//if err != nil {
	//	log.Printf("Error executing script: %v", err)
	//	return
	//}

	time.Sleep(100 * time.Second)
	// 等待一段时间以查看结果（可选）
	browser.Timeout(100)

	// 关闭浏览器
	browser.MustClose()

}

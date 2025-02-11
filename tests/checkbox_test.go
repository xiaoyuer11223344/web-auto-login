package tests

import (
	"fmt"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"testing"
	"time"
)

func Test_checkbox(t *testing.T) {
	// 启动浏览器
	browser := rod.New().ControlURL(launcher.New().Headless(false).MustLaunch()).MustConnect()
	defer browser.MustClose()

	// 创建一个页面并导航到指定的URL
	page := browser.MustPage("https://passport.geely.com/")
	defer page.Close()

	// 等待目标元素加载完成
	page.MustWaitLoad()

	el := page.MustElement(`[type="checkbox"]`)
	// check it if not checked
	if !el.MustProperty("checked").Bool() {
		fmt.Println("321321")
		el.Click(proto.InputMouseButtonLeft, 1)
	} else {
		fmt.Println("123123")
	}

	time.Sleep(10 * time.Second)
}

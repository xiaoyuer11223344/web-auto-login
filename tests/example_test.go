package tests

import (
	"fmt"
	"testing"
	"time"
	"zp-weblogin/pkg/browser"
	"zp-weblogin/pkg/sdk"
)

func Test_sdk_login(t *testing.T) {
	for _, url := range []string{"http://106.14.66.138/"} {
		result, err := sdk.Login(sdk.Config{
			URL:      "http://106.14.66.138/",
			User:     "admin",
			Pass:     "admin123",
			Headless: true,
			Timeout:  30 * time.Second,
		})

		if err != nil {
			fmt.Printf("%s Login failed: %v\\n", url, err)
			continue
		}

		if result.Success {
			fmt.Printf("%s Login successful with user: %s pass: %s \n", result.Url, result.User, result.Pass)
		} else {
			fmt.Printf("%s Login faile with user: %s pass: %s error: %s\n", result.Url, result.User, result.Pass, result.Error)
		}

	}
}

func Test_sdk_login_by_selector(t *testing.T) {
	// 使用自定义选择器的示例
	selector := &browser.Selector{
		UserInput:     "/html/body/div/div/div[2]/form/ul/li[2]/div/input",
		PasswordInput: "/html/body/div/div/div[2]/form/ul/li[3]/div/input",
		LoginBtn:      "/html/body/div/div/div[2]/form/ul/li[4]/button",
	}

	result, err := sdk.LoginWithSelector(sdk.Config{
		URL:      "http://106.14.66.138/",
		User:     "admin",
		Pass:     "admin123",
		Headless: true,
		Timeout:  30 * time.Second,
	}, selector)

	if err != nil {
		fmt.Printf("Login with custom selector failed: %v\\n", err)
		return
	}

	if result.Success {
		fmt.Printf("Login with custom selector successful\\n")
	}

}

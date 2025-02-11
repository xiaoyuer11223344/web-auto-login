package sdk

import (
	"context"
	"time"
	"zp-weblogin/pkg/browser"
)

type Config struct {
	URL      string        // 登录URL
	User     string        // 用户名
	Pass     string        // 密码
	OCRUrl   string        // OCR服务地址
	Headless bool          // 无头模式
	Timeout  time.Duration // 超时时间，默认30秒
}

type Result struct {
	Url     string
	Success bool   // 登录是否成功
	User    string // 成功登录的用户名
	Pass    string // 成功登录的密码
	Error   error  // 如果失败，错误信息
}

// Login 使用默认选择器进行登录
func Login(c Config) (*Result, error) {
	b, err := browser.New(c.Headless, "")
	if err != nil {
		return nil, err
	}
	defer b.Close()

	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	if err = b.Navigate(ctx, c.URL); err != nil {
		return nil, err
	}

	s := &browser.Selector{
		UserInput:     "input[name='username']",
		PasswordInput: "input[type='password']",
		LoginBtn:      "button[type='submit']",
		RememberMe:    "input[type='checkbox']",
	}

	err = b.Login(ctx, s, c.User, c.Pass)
	return &Result{
		Url:     c.URL,
		Success: err == nil,
		User:    c.User,
		Pass:    c.Pass,
		Error:   err,
	}, nil
}

// LoginWithSelector 允许使用自定义选择器进行登录
func LoginWithSelector(c Config, s *browser.Selector) (*Result, error) {
	b, err := browser.New(c.Headless, "")
	if err != nil {
		return nil, err
	}
	defer b.Close()

	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	if err := b.Navigate(ctx, c.URL); err != nil {
		return nil, err
	}

	err = b.Login(ctx, s, c.User, c.Pass)
	return &Result{
		Url:     c.URL,
		Success: err == nil,
		User:    c.User,
		Pass:    c.Pass,
		Error:   err,
	}, nil
}

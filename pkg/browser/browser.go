package browser

import (
	"context"
	"fmt"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/launcher"
	"strings"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	log "github.com/sirupsen/logrus"
)

type Browser struct {
	browser *rod.Browser
	page    *rod.Page
	mu      sync.Mutex
}

var MyDevice = devices.Device{
	Title:          "Chrome computer",
	Capabilities:   []string{"touch", "mobile"},
	UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
	AcceptLanguage: "en",
	Screen: devices.Screen{
		DevicePixelRatio: 2,
		Horizontal: devices.ScreenSize{
			Width:  1500,
			Height: 900,
		},
		Vertical: devices.ScreenSize{
			Width:  1500,
			Height: 900,
		},
	},
}

// New
// @Description: 初始化go-rod browser
// @param headless
// @param proxy
// @return *Browser
// @return error
func New(headless bool, proxy string) (*Browser, error) {
	l := launcher.New().Headless(headless)
	l.Set("ignore-certificate-errors").
		Delete("disable-component-extensions-with-background-pages").
		Set("disable-extensions").
		Append("disable-features", "BlinkGenPropertyTrees").
		Set("hide-scrollbars").
		Set("mute-audio").
		Set("no-default-browser-check").
		Delete("no-startup-window").
		Set("password-store", "basic").
		Set("safebrowsing-disable-auto-update").
		Set("disable-gpu").
		Set("no-default-browser-check").
		Set("disable-images", "true").
		Set("enable-automation", "false").                     // 防止监测 webdriver
		Set("disable-blink-features", "AutomationControlled"). // 禁用 blink 特征，绕过了加速乐检测
		NoSandbox(true)

	if proxy != "" {
		l = l.Proxy(proxy)
	}

	browser := rod.New().ControlURL(l.MustLaunch()).Timeout(30 * time.Second).MustConnect()
	browser.DefaultDevice(MyDevice)
	return &Browser{browser: browser}, nil
}

// Close
// @Description: 资源释放
// @receiver b
// @return error
func (b *Browser) Close() error {
	if b.page != nil {
		if err := b.page.Close(); err != nil {
			return fmt.Errorf("failed to close page: %w", err)
		}
	}
	if err := b.browser.Close(); err != nil {
		return fmt.Errorf("failed to close browser: %w", err)
	}
	return nil
}

func (b *Browser) IsLoggedIn() bool {
	// Check for common login success indicators
	successIndicators := []string{
		".user-info",
		".user-profile",
		".logout-btn",
		"#logout",
		".welcome-message",
	}

	for _, selector := range successIndicators {
		if el, err := b.page.Element(selector); err == nil && el != nil {
			if visible, _ := el.Visible(); visible {
				return true
			}
		}
	}

	// Check URL for login-related paths
	currentURL := b.page.MustInfo().URL
	loginPaths := []string{"/login", "/signin", "/auth"}
	for _, path := range loginPaths {
		if strings.Contains(currentURL, path) {
			return false
		}
	}

	// Check for error messages
	errorIndicators := []string{
		".error-message",
		".alert-error",
		".login-error",
		".colorR",
	}

	for _, selector := range errorIndicators {
		if el, err := b.page.Element(selector); err == nil && el != nil {
			if visible, _ := el.Visible(); visible {
				return false
			}
		}
	}

	return true
}

func (b *Browser) findElement(selector, name string) (*rod.Element, error) {
	logger := log.WithFields(log.Fields{
		"selector": selector,
		"name":     name,
		"attempt":  "0",
	})
	logger.Debug("Finding element")

	var el *rod.Element
	var err error

	// Common CSS selectors for form elements
	cssSelectors := map[string][]string{
		"username input": {
			"#user_login",
			"input[name='user[login]']",
			"input[name='username']",
			"input[type='text']",
			"input[id*='username']",
			"input[placeholder*='用户名']",
		},
		"password input": {
			"#user_password",
			"input[name='user[password]']",
			"input[name='password']",
			"input[type='password']",
			"input[id*='password']",
			"input[placeholder*='密码']",
		},
		"login button": {
			"button[type='submit']",
			"input[type='submit']",

			"button[id*='login-btn']",
			"input[id*='login-btn']",

			"button[id*='commit']",
			"input[name='commit']",

			"div[class='lui_login_button_div_c']",
			"button[class*='login_button']",
			"input[class*='login_button']",

			"button[class*='btn-login']",
			"input[class*='btn-login']",

			"button[value*='登录']",
			"input[value*='登录']",
		},
		"remember checkbox": {
			"`input[type='checkbox']",
		},
	}

	// Wait for element with retry and fallback
	for i := 0; i < 3; i++ {
		// Try XPath first
		el, err = b.page.ElementX(selector)
		// If XPath fails, try CSS selectors
		if (err != nil || el == nil) && cssSelectors[name] != nil {
			for _, cssSelector := range cssSelectors[name] {
				el, err = b.page.Element(cssSelector)
				if err == nil && el != nil {
					if visible, _ := el.Visible(); visible {
						logger.WithField("css_selector", cssSelector).Debug("Element found using CSS selector")
						return el, nil
					}
				}
			}
		}

		if err == nil && el != nil {
			if visible, _ := el.Visible(); visible {
				logger.Info("Element found and ready")
				return el, nil
			}
		}

		if i < 2 {
			time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
			logger.WithField("attempt", i+1).Debug("Element not found, retrying...")
		}
	}

	if name == "remember checkbox" {
		return nil, nil
	}

	return nil, fmt.Errorf("element %s not found or not visible after retries", name)
}

func (b *Browser) performLogin(selector *Selector, username, password string) error {
	start := time.Now()
	logger := log.WithFields(log.Fields{
		"action":   "perform_login",
		"username": username,
		"password": password,
		"url":      b.page.MustInfo().URL,
	})

	logger.Debug("Starting form interaction")

	if selector == nil {
		return fmt.Errorf("selector cannot be nil")
	}

	var err error

	// todo: Find UserInput elements
	var userEL *rod.Element
	if userEL, err = b.findElement(selector.UserInput, "username input"); err != nil {
		return err
	}
	if err = userEL.Input(username); err != nil {
		return fmt.Errorf("failed to input username: %v", err)
	}
	//if _, err = userEL.Eval(`(xpath,value) => {
	//		const xpathExpression = xpath;
	//		const result = document.evaluate(
	//			xpathExpression,
	//			document,
	//			null,
	//			XPathResult.FIRST_ORDERED_NODE_TYPE,
	//			null
	//		);
	//
	//		const element = result.singleNodeValue;
	//		if (element) {
	//			element.value = value
	//			return true;
	//		}
	//		return false;
	//	}`, selector.UserInput, username); err != nil {
	//	return fmt.Errorf("failed to input username: %v", err)
	//}
	time.Sleep(500 * time.Millisecond)

	// todo: Find PasswordInput elements
	var passEl *rod.Element
	if passEl, err = b.findElement(selector.PasswordInput, "password input"); err != nil {
		return err
	}
	if err = passEl.Input(password); err != nil {
		return fmt.Errorf("failed to input password: %v", err)
	}
	//if _, err = passEl.Eval(`(xpath,value) => {
	//		const xpathExpression = xpath;
	//		const result = document.evaluate(
	//			xpathExpression,
	//			document,
	//			null,
	//			XPathResult.FIRST_ORDERED_NODE_TYPE,
	//			null
	//		);
	//
	//		const element = result.singleNodeValue;
	//		if (element) {
	//			element.value = value
	//			return true;
	//		}
	//		return false;
	//	}`, selector.PasswordInput, password); err != nil {
	//	return fmt.Errorf("failed to input password: %v", err)
	//}
	time.Sleep(500 * time.Millisecond)

	// todo: Find CheckBox elements
	if _, err = b.page.Eval(`() => {
		const checkbox = document.querySelector('input[type="checkbox"]');
		console.log(checkbox);
		if (checkbox) {
			checkbox.click();
			return true;
		}
		return false;
	}`); err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)

	// todo: Find LoginBtn elements
	var btnEL *rod.Element
	if btnEL, err = b.findElement(selector.LoginBtn, "login button"); err != nil {
		return err
	}

	//if err = btnEL.Click(proto.InputMouseButtonLeft, 1); err != nil {
	//	return fmt.Errorf("failed to click login button: %v", err)
	//}
	//if _, err = btnEL.Eval(`(element) => {
	//		console.log(element)
	//		const el = document.querySelector(element.Object.description);
	//		el.click();
	//	}`, btnEL); err != nil {
	//	return fmt.Errorf("failed to click login button: %v", err)
	//}
	if _, err = btnEL.Eval(`(xpath) => {
			const xpathExpression = xpath;
			const result = document.evaluate(
				xpathExpression,
				document,
				null,
				XPathResult.FIRST_ORDERED_NODE_TYPE,
				null
			);
	
			const element = result.singleNodeValue; // 获取匹配的节点
			if (element) {
				element.click()
				return true;
			}
			return false;
		}`, selector.LoginBtn); err != nil {
		return fmt.Errorf("failed to click login button: %v", err)
	}
	time.Sleep(500 * time.Millisecond)

	// Brief wait for form submission

	logger.WithField("duration", time.Since(start)).Debug("Login form submitted")

	return nil
}

func (b *Browser) Login(ctx context.Context, selector *Selector, username, password string) error {
	start := time.Now()

	logger := log.WithFields(log.Fields{
		"action":   "login_attempt",
		"username": username,
		"password": password,
		"url":      b.page.MustInfo().URL,
	})

	logger.Debug("Testing credentials")

	if selector == nil {
		var err error
		selector, err = b.DetectFormSelectors()
		if err != nil {
			return fmt.Errorf("failed to detect selectors: %w", err)
		}
	}

	// 每次任务登录的上下文
	loginCtx, cancel := context.WithTimeout(ctx, time.Duration(10)*time.Second)
	defer cancel()

	// 登录操作
	if err := b.performLogin(selector, username, password); err != nil {
		return err
	}

	// 轮询等待结果
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-loginCtx.Done():
			return fmt.Errorf("login timeout after %v", 10)
		case <-ticker.C:
			// Check for error messages first
			if errEL, err := b.page.Element("div[role='alert']"); err == nil && errEL != nil {
				if visible, _ := errEL.Visible(); visible {
					if text, err := errEL.Text(); err == nil && text != "" {
						logger.WithField("error", text).Debug("Found error message")
						return fmt.Errorf("login error: %s", text)
					}
				}
			}

			// Check login status
			if b.IsLoggedIn() {
				logger.WithField("duration", time.Since(start)).Info("Login successful")
				return nil
			}
		}
	}
}

func (b *Browser) Navigate(ctx context.Context, url string) error {
	var err error

	logger := log.WithFields(log.Fields{
		"action": "navigate",
		"url":    url,
	})
	logger.Debug("Starting navigation")

	// Clean up previous session
	if b.page != nil {
		if err = b.page.Close(); err != nil {
			logger.WithError(err).Debug("Error during cleanup")
		}
	}

	// Create new browser page
	var page *rod.Page
	loginURL := strings.TrimRight(url, "/")
	page, err = b.browser.Page(proto.TargetCreateTarget{URL: loginURL})
	if err != nil {
		return fmt.Errorf("page creation failed: %w", err)
	}

	b.page = page.Context(ctx)

	// Create error channel for timeout handling
	errChan := make(chan error, 1)
	go func() {
		defer close(errChan)

		// Configure viewport
		//if err = b.page.SetViewport(&proto.EmulationSetDeviceMetricsOverride{
		//	Width:  1920,
		//	Height: 1080,
		//}); err != nil {
		//	errChan <- fmt.Errorf("viewport setup failed: %w", err)
		//	return
		//}

		// Wait for initial page load
		if err = b.page.WaitLoad(); err != nil {
			errChan <- fmt.Errorf("page load failed: %w", err)
			return
		}

		errChan <- nil
	}()

	// Wait for navigation to complete or timeout
	select {
	case err = <-errChan:
		if err != nil {
			return fmt.Errorf("navigation failed: %w", err)
		}
	case <-ctx.Done():
		return fmt.Errorf("navigate timed out after 15 seconds")
	}

	logger.Debug("Navigation completed successfully")
	return nil
}

// GetHtmlContent
// @Description: 获取当前rod.Page对象的页面信息
// @receiver b
// @return string
// @return error
func (b *Browser) GetHtmlContent() (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.page != nil {
		return b.page.HTML()
	} else {
		return "", fmt.Errorf("browser page is nil")
	}
}

// GetPage
// @Description: 获取当前rod.Page对象
// @receiver b
// @return *rod.Page
func (b *Browser) GetPage() *rod.Page {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.page
}

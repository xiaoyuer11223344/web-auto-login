package tests

import (
	"context"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"os"
	"testing"
	"time"
)

func TestCaptchaScreenshot(t *testing.T) {
	// Create browser instance
	l := launcher.New().Headless(false).NoSandbox(true)
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
		Set("enable-automation", "false"). // 防止监测 webdriver
		Set("disable-blink-features", "AutomationControlled")

	browser := rod.New().ControlURL(l.MustLaunch()).Timeout(30 * time.Second).MustConnect()

	defer browser.MustClose()

	// Create page
	page := browser.MustPage()
	defer page.MustClose()

	// Navigate to target URL
	err := page.Navigate("http://120.46.192.23:90/vul/burteforce/bf_client.php")
	if err != nil {
		t.Fatalf("Failed to navigate: %v", err)
	}

	// Wait for page load
	page.MustWaitLoad()

	// Wait for network idle
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	page.Context(ctx).MustWaitIdle()

	// Find element by XPath
	el, err := page.ElementX("/html/body/div[2]/div[2]/div/div[2]/div/div/form/label[4]/input")
	if err != nil {
		t.Fatalf("Failed to find element: %v", err)
	}

	// Take screenshot
	data, err := el.Screenshot(proto.PageCaptureScreenshotFormat("png"), 1)
	if err != nil {
		t.Fatalf("Failed to take screenshot: %v", err)
	}

	// Save to file
	err = os.WriteFile("3.png", data, 0644)
	if err != nil {
		t.Fatalf("Failed to save screenshot: %v", err)
	}
}

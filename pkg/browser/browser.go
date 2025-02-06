package browser

import (
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/sirupsen/logrus"
)

type Browser struct {
	browser *rod.Browser
	page    *rod.Page
	timeout time.Duration
}

func New(headless bool, proxy string) (*Browser, error) {
	l := launcher.New()
	if headless {
		l = l.Headless(true)
	}
	if proxy != "" {
		l = l.Proxy(proxy)
	}
	
	u := l.MustLaunch()
	browser := rod.New().
		ControlURL(u).
		Timeout(30 * time.Second).
		MustConnect()

	return &Browser{
		browser: browser,
		timeout: 30 * time.Second,
	}, nil
}

func (b *Browser) Close() error {
	if b.page != nil {
		if err := b.page.Close(); err != nil {
			logrus.Warnf("Error closing page: %v", err)
		}
	}
	return b.browser.Close()
}

func (b *Browser) IsLoggedIn() bool {
	if err := b.page.WaitIdle(b.timeout); err != nil {
		logrus.Warnf("Timeout waiting for page idle: %v", err)
		return false
	}

	// Check for successful login by looking for logout link or secure area text
	// Wait for page to settle after login attempt
	time.Sleep(2 * time.Second)

	// Try to find success indicators with timeout
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			logrus.Debug("Login check timed out")
			return false
		case <-ticker.C:
			// Check for logout link
			if el, err := b.page.Element("a#logout"); err == nil {
				if visible, _ := el.Visible(); visible {
					logrus.Debug("Login successful - found logout link")
					return true
				}
			}
			// Check for secure area heading
			if el, err := b.page.Element("h2"); err == nil {
				if text, _ := el.Text(); text == "Secure Area" {
					logrus.Debug("Login successful - found secure area heading")
					return true
				}
			}
		}
	}
}

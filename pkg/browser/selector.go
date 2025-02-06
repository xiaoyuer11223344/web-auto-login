package browser

import (
	"fmt"
	"time"

	"github.com/go-rod/rod/lib/proto"
	"github.com/sirupsen/logrus"
)

type Selector struct {
	UserInput     string `yaml:"userInput"`
	PasswordInput string `yaml:"passwordInput"`
	LoginBtn      string `yaml:"loginBtn"`
	CaptchaInput  string `yaml:"captchaInput"`
	CaptchaImg    string `yaml:"captchaImg"`
}

func (b *Browser) DetectSelectors() (*Selector, error) {
	selector := &Selector{}

	userInputSelectors := []string{
		"#username",
		"input[name='username']",
		"input[type='text']",
		"input[name*='user']",
		"input[id*='user']",
		"input[class*='user']",
		"input[name='uid']",
		"input[id='uid']",
		"input[id='usernameIpt']",
	}

	passInputSelectors := []string{
		"#password",
		"input[name='password']",
		"input[type='password']",
		"input[name*='pass']",
		"input[id*='pass']",
		"input[class*='pass']",
		"input[name='pwd']",
		"input[id='pwd']",
	}

	loginBtnSelectors := []string{
		"button[type='submit']",
		"input[type='submit']",
		".radius",
		".btn-login",
		".login-btn",
		"#login-btn",
		"#loginBtn",
		"input[value='Login']",
		"button.btn-primary",
		"button.submit",
	}

	for _, sel := range userInputSelectors {
		logrus.Debugf("Trying username selector: %s", sel)
		if el := b.page.MustElement(sel); el != nil {
			if visible, _ := el.Visible(); visible {
				selector.UserInput = sel
				logrus.Debugf("Found username input: %s", sel)
				break
			}
		}
	}

	for _, sel := range passInputSelectors {
		logrus.Debugf("Trying password selector: %s", sel)
		if el := b.page.MustElement(sel); el != nil {
			if visible, _ := el.Visible(); visible {
				selector.PasswordInput = sel
				logrus.Debugf("Found password input: %s", sel)
				break
			}
		}
	}

	for _, sel := range loginBtnSelectors {
		logrus.Debugf("Trying login button selector: %s", sel)
		if el := b.page.MustElement(sel); el != nil {
			if visible, _ := el.Visible(); visible {
				selector.LoginBtn = sel
				logrus.Debugf("Found login button: %s", sel)
				break
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"userInput": selector.UserInput,
		"passInput": selector.PasswordInput,
		"loginBtn":  selector.LoginBtn,
	}).Debug("Detected selectors")

	missing := []string{}
	if selector.UserInput == "" {
		missing = append(missing, "username input")
	}
	if selector.PasswordInput == "" {
		missing = append(missing, "password input")
	}
	if selector.LoginBtn == "" {
		missing = append(missing, "login button")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("failed to detect selectors: %v", missing)
	}

	return selector, nil
}

func (b *Browser) Navigate(url string) error {
	logrus.Debugf("Navigating to URL: %s", url)

	page, err := b.browser.Page(proto.TargetCreateTarget{URL: url})
	if err != nil {
		return fmt.Errorf("failed to create page: %v", err)
	}
	b.page = page

	page.MustHandleDialog()

	logrus.Debug("Waiting for page to load...")
	if err := b.page.WaitLoad(); err != nil {
		return fmt.Errorf("failed to wait for page load: %v", err)
	}

	logrus.Debug("Waiting for network idle...")
	if err := b.page.WaitIdle(5 * time.Second); err != nil {
		logrus.Warnf("Timeout waiting for network idle: %v", err)
	}

	logrus.Debug("Additional wait for dynamic content...")
	time.Sleep(2 * time.Second)
	return nil
}

func (b *Browser) Login(selector *Selector, username, password string) error {
	userEl := b.page.MustElementX(selector.UserInput)
	if err := userEl.Input(username); err != nil {
		return fmt.Errorf("failed to input username: %v", err)
	}

	passEl := b.page.MustElementX(selector.PasswordInput)
	if err := passEl.Input(password); err != nil {
		return fmt.Errorf("failed to input password: %v", err)
	}

	btnEl := b.page.MustElementX(selector.LoginBtn)
	if err := btnEl.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("failed to click login button: %v", err)
	}

	// Wait for navigation and check for success elements
	time.Sleep(1 * time.Second)
	if el, err := b.page.Element(".flash.success"); err == nil {
		if visible, _ := el.Visible(); visible {
			logrus.Debug("Login successful - found success message")
			return nil
		}
	}

	return fmt.Errorf("login attempt failed - no success message found")

}

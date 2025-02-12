package browser

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	log "github.com/sirupsen/logrus"
)

type FormScore struct {
	Form      *rod.Element
	Score     int
	HasLogin  bool
	HasPass   bool
	HasSubmit bool
	Position  proto.Point
}

type Selector struct {
	UserInput     string `yaml:"userInput" json:"userInput"`
	PasswordInput string `yaml:"passwordInput" json:"passwordInput"`
	LoginBtn      string `yaml:"loginBtn" json:"loginBtn"`
	RememberMe    string `yaml:"rememberMe" json:"rememberMe"`
	CaptchaInput  string `yaml:"captchaInput" json:"captchaInput"`
	CaptchaImg    string `yaml:"captchaImg" json:"captchaImg"`
	form          *rod.Element
}

func scoreLoginForm(form *rod.Element) (*FormScore, error) {
	score := &FormScore{Form: form}

	// GitLab-style inputs (highest priority)
	if el, _ := form.Element("input[name='user[login]']"); el != nil {
		score.Score += 10
		score.HasLogin = true
	}
	if el, _ := form.Element("input[name='user[password]']"); el != nil {
		score.Score += 10
		score.HasPass = true
	}

	// Standard inputs
	if el, _ := form.Element("input[type='password']"); el != nil {
		score.Score += 5
		score.HasPass = true
	}

	// Form attributes
	if action, _ := form.Attribute("action"); action != nil {
		if strings.Contains(*action, "login") || strings.Contains(*action, "signin") {
			score.Score += 3
		}
	}

	// Submit button detection
	submitSelectors := []string{
		"input[type='submit']",
		"button[type='submit']",
		"input[name='commit'][type='submit']",
		"button:contains('Sign in')",
		"button:contains('Login')",
	}

	for _, sel := range submitSelectors {
		if el, _ := form.Element(sel); el != nil {
			if visible, _ := el.Visible(); visible {
				score.Score += 2
				score.HasSubmit = true
				break
			}
		}
	}

	// Get form position
	if shape, err := form.Shape(); err == nil && len(shape.Quads) > 0 {
		quad := shape.Quads[0]
		score.Position = proto.Point{X: quad[0], Y: quad[1]}
	}

	return score, nil
}

func (b *Browser) tryFindElement(selector string) (*rod.Element, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Wait for dynamic content
	time.Sleep(2 * time.Second)

	// Wait for page to stabilize and JavaScript to execute
	time.Sleep(3 * time.Second)

	logger := log.WithField("action", "find_element")
	logger.WithField("selector", selector).Debug("Attempting to find element")

	var el *rod.Element
	var err error

	if strings.HasPrefix(selector, "//") {
		el, err = b.page.Context(ctx).ElementX(selector)
	} else {
		el, err = b.page.Context(ctx).Element(selector)
	}

	if err != nil {
		return nil, fmt.Errorf("element not found: %v", err)
	}

	if el == nil {
		return nil, fmt.Errorf("element not found with selector %s", selector)
	}

	if err := el.Context(ctx).WaitVisible(); err != nil {
		return nil, fmt.Errorf("element not visible: %v", err)
	}

	return el, nil
}

// findForm
// @Description: 匹配表单元素
// @receiver b
// @return *rod.Element
// @return error
func (b *Browser) findForm() (*rod.Element, error) {
	var form *rod.Element
	var err error

	logger := log.WithField("action", "detect_form")
	logger.Debug("Starting form detection")

	for i := 0; i < MaxRetries; i++ {
		if form, err = b.page.Element("form"); err == nil && form != nil {
			if visible, _ := form.Visible(); visible {
				break
			}
		}
		if i < MaxRetries-1 {
			time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
		}
	}

	if form == nil {
		// 没匹配到表单的话则进行结束
		return nil, fmt.Errorf("no visible form found")
	} else {
		// 匹配到表单的话则进行打印信息
		formDetails := map[string]interface{}{
			"tag":      form.MustEval("() => this.tagName ").String(),
			"id":       form.MustEval("() => this.id ").String(),
			"action":   form.MustEval("() => this.action ").String(),
			"method":   form.MustEval("() => this.method ").String(),
			"position": form.MustShape().Box(),
		}
		logger.WithFields(log.Fields{"formDetails": formDetails}).Info("Form Details")
	}

	return form, nil

}

// findFormElements
// @Description: 匹配表单内元素
// @receiver b
// @param form
// @return *Selector
// @return error
func (b *Browser) findFormElements(form *rod.Element) (*Selector, error) {
	logger := log.WithField("action", "find_form_elements")

	selector := &Selector{form: form}

	// Common selectors for form elements
	userInputSelectors := []string{
		"input[placeholder='用户名']",
		"input[placeholder='账号']",
		"input[name='user[login]']",
		"input[name='username']",
		"#username",
		"input[type='text']",
		"input[name*='user']",
		"input[id*='user']",
		"input[class*='user']",
		"input[name='uid']",
		"input[id='uid']",
		"input[id='usernameIpt']",
		"input[name='account']",
		"input[id='account']",
	}

	passInputSelectors := []string{
		"input[placeholder*='密码']",
		"input[name='user[password]']",
		"input[type='password']",
		"#password",
		"input[name='password']",
		"input[name*='pass']",
		"input[id*='pass']",
		"input[class*='pass']",
		"input[name='pwd']",
		"input[id='pwd']",
	}

	loginBtnSelectors := []string{
		"button[type='button']",
		"button[type='submit']",
		"input[type='submit']",
		"input[value='Login']",
		"input[value*='Sign in']",
		"input[value*='登录']",
		"button[id*='login-btn']",
		"input[id*='login-btn']",
		"button[id*='commit']",
		"input[name='commit']",
		"button[class*='login_button']",
		"div[class*='login_button']",
		"input[class*='login_button']",
		"button[class*='btn-login']",
		"input[class*='btn-login']",
		"button[value*='登录']",

		".radius",
		".btn-login",
		".login-btn",
		"button.btn-primary",
		"button.submit",
		"#login-btn",
		"#loginBtn",
	}

	checkBoxSelectors := []string{
		"input[type='checkbox']",
	}

	captchaInputSelectors := []string{
		"input[placeholder*='验证码']",
		"input[placeholder*='verification']",
		"input[placeholder*='Verification']",
	}

	captchaImageSelectors := []string{
		"img",
		"input[id='checkCode']",
		//"img[id*='captcha']",
		//"img[id*='Captcha']",
		//"img[alt*='验证码']",
		//"img[alt*='captcha']",
		//"img[src*='captcha']",
		//"img[src*='verify']",
		//"img[class*='captcha']",
		//"img[id*='captcha']",
		//"img[title*='验证码']",
		//"img[title*='captcha']",
		".el-image img[src*='captcha']",
		".el-image[alt*='验证码']",
		".el-image[alt*='captcha']",
		".captcha-img",
		".verify-img",
	}

	// Find username input with retry
	for i := 0; i < MaxRetries; i++ {
		for _, sel := range userInputSelectors {
			if el, err := form.Element(sel); err == nil && el != nil {
				if visible, _ := el.Visible(); visible {
					selector.UserInput = el.MustGetXPath(false)
					logger.WithField("xpath", selector.UserInput).Debug("Found username input")
					goto foundPass
				}
			}
		}
		if i < MaxRetries-1 {
			time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
			logger.WithField("attempt", i+1).Debug("Username input not found, retrying...")
		}
	}

foundPass:
	// Find password input with retry
	for i := 0; i < MaxRetries; i++ {
		for _, sel := range passInputSelectors {
			if el, err := form.Element(sel); err == nil && el != nil {
				if visible, _ := el.Visible(); visible {
					selector.PasswordInput = el.MustGetXPath(false)
					logger.WithField("xpath", selector.PasswordInput).Debug("Found password input")
					goto foundButton
				}
			}
		}
		if i < MaxRetries-1 {
			time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
			logger.WithField("attempt", i+1).Debug("Password input not found, retrying...")
		}
	}

foundButton:
	// Find login button with retry
	for i := 0; i < MaxRetries; i++ {
		for _, sel := range loginBtnSelectors {
			if el, err := form.Element(sel); err == nil && el != nil {
				if visible, _ := el.Visible(); visible {
					selector.LoginBtn = el.MustGetXPath(false)
					logger.WithField("xpath", selector.LoginBtn).Debug("Found login button")
					goto foundRememberCheckBox
				}
			}
		}
		if i < MaxRetries-1 {
			time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
			logger.WithField("attempt", i+1).Debug("Login button not found, retrying...")
		}
	}

foundRememberCheckBox:
	for i := 0; i < MaxRetries; i++ {
		for _, sel := range checkBoxSelectors {
			// Find checkboxes (both remember me and agreement types)
			checkboxes, err := form.Elements(sel)
			if err == nil && len(checkboxes) > 0 {
				for _, checkbox := range checkboxes {
					//if visible, _ := checkbox.Visible(); visible {}
					selector.RememberMe = checkbox.MustGetXPath(false)
					logger.WithField("xpath", selector.RememberMe).Debug("Found rememberMe checkbox")
					goto foundCaptchaInput
				}
			}
		}

		if i < MaxRetries-1 {
			time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
			logger.WithField("attempt", i+1).Debug("rememberMe checkbox not found, retrying...")
		}
	}

foundCaptchaInput:
	if b.captchaHandler != nil {
		for i := 0; i < MaxRetries; i++ {
			for _, sel := range captchaInputSelectors {
				if el, err := form.Element(sel); err == nil && el != nil {
					if visible, _ := el.Visible(); visible {
						selector.CaptchaInput = el.MustGetXPath(false)
						logger.WithField("xpath", selector.CaptchaInput).Debug("Found Captcha Input")
						goto foundCaptchaImage
					}
				}
			}

			if i < MaxRetries-1 {
				time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
				logger.WithField("attempt", i+1).Debug("captcha input not found, retrying...")
			}
		}
	}

foundCaptchaImage:
	if b.captchaHandler != nil && selector.CaptchaInput != "" {
		for i := 0; i < MaxRetries; i++ {
			for _, sel := range captchaImageSelectors {
				if el, err := form.Element(sel); err == nil && el != nil {
					if visible, _ := el.Visible(); visible {
						selector.CaptchaImg = el.MustGetXPath(false)
						logger.WithField("xpath", selector.CaptchaImg).Debug("Found Captcha Image")
						goto over
					}
				}
			}

			if i < MaxRetries-1 {
				time.Sleep(BackoffFactor * time.Duration(1<<uint(i)))
				logger.WithField("attempt", i+1).Debug("captcha image not found, retrying...")
			}
		}
	}

over:
	if selector.UserInput != "" && selector.PasswordInput != "" && selector.LoginBtn != "" {
		return selector, nil
	}

	return nil, fmt.Errorf("not form all elements found")
}

// DetectFormSelectors
// @Description: 自动探测Form表单以及内部相关的其他标签元素
// @receiver b
// @return *Selector
// @return error
func (b *Browser) DetectFormSelectors() (*Selector, error) {
	logger := log.WithField("action", "detect_form_and_selectors")

	var err error
	var form *rod.Element
	if form, err = b.findForm(); err != nil {
		// 匹配表单信息
		return nil, err
	}

	logger.Debug("Starting selector detection")
	var selector *Selector
	if selector, err = b.findFormElements(form); err == nil {
		// 匹配表单内标签元素
		return selector, nil
	}

	// 如果直接检测失败，请尝试表单评分方法
	logger.Debug("Direct detection failed, trying form scoring approach")
	var formScores []*FormScore
	forms, err := b.page.Elements("form")
	if err != nil {
		return nil, fmt.Errorf("no forms found: %w", err)
	}

	// 对所有表格进行评分
	for _, form = range forms {
		score, formErr := scoreLoginForm(form)
		if formErr == nil {
			formScores = append(formScores, score)
			formDetails := map[string]interface{}{
				"tag":    form.MustEval("() => this.tagName").String(),
				"id":     form.MustEval("() => this.id").String(),
				"action": form.MustEval("() => this.action").String(),
				"method": form.MustEval("() => this.method").String(),
			}

			logger.WithFields(log.Fields{
				"score":       score.Score,
				"hasLogin":    score.HasLogin,
				"hasPass":     score.HasPass,
				"hasSubmit":   score.HasSubmit,
				"position":    score.Position,
				"formDetails": formDetails,
			}).Debug("Form scored with details")
		}
	}

	// 得分排序
	sort.Slice(formScores, func(i, j int) bool {
		return formScores[i].Score > formScores[j].Score
	})

	// 选择匹配度最高的表单元素
	if len(formScores) > 0 && formScores[0].Score > 5 {
		form = formScores[0].Form
		return b.findFormElements(form)
	}

	return nil, fmt.Errorf("no suitable login form found")
}

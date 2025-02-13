package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"time"
	"xiaoyu/pkg/browser"
	"xiaoyu/pkg/crack"
)

func init() {
	flags := webLoginCmd.Flags()

	// Cracking flags
	flags.BoolVar(&globalOptions.crackAll, "crack-all", false, "crack all user and pass")
	flags.IntVar(&globalOptions.delay, "delay", 1, "delay time between crack")
	flags.IntVar(&globalOptions.maxAttempts, "max-attempts", 3, "max attempts")
	flags.IntVar(&globalOptions.maxCrackNum, "max-crack-num", 0, "max crack num, 0 is no limit")
	flags.IntVar(&globalOptions.maxCrackTime, "max-crack-time", 300, "max crack time in sec")

	// Detection flags
	flags.IntVar(&globalOptions.navigationTimeout, "navigation-timeout", 10, "navigation timeout in seconds")
	flags.IntVar(&globalOptions.elementTimeout, "element-timeout", 5, "element detection timeout in seconds")
	flags.IntVar(&globalOptions.loginTimeout, "login-all-timeout", 15, "login attempt timeout in seconds")

	flags.StringSliceVar(&globalOptions.userList, "user", nil, "user list, split by comma")
	flags.StringVar(&globalOptions.userFile, "user-file", "", "user file")
	flags.StringSliceVar(&globalOptions.passList, "pass", nil, "pass list, split by comma")
	flags.StringVar(&globalOptions.passFile, "pass-file", "", "pass file")
	flags.StringVar(&globalOptions.selectorFile, "selector-file", "", "selector file")

	flags.StringVar(&globalOptions.ocrURL, "ocr-url", "http://120.26.57.12:8000", "OCR service URL for captcha solving")

	rootCmd.AddCommand(webLoginCmd)
}

var webLoginCmd = &cobra.Command{
	Use:   "weblogin",
	Short: "Start web login testing",
	Long:  "Test web login pages with automatic form detection and login attempts. Use --detect-only to only detect form elements without attempting login.",
	Example: `  # Test login with credentials
  ./weblogin weblogin -i http://example.com:9001 --user admin --pass admin123
  
  # Only detect form elements
  ./weblogin weblogin -i http://example.com:9001 --detect-only`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.Info("Program started")

		log.WithFields(log.Fields{
			"count": len(globalOptions.inputs),
		}).Info("Target URLs loaded")

		printConfiguration(globalOptions)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := setupLogging(globalOptions.logLevel); err != nil {
			return fmt.Errorf("failed to setup logging: %w", err)
		}

		cfg, err := loadConfig(globalOptions)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		return run(cfg)
	},
}

func CreateTasks(flags *Options) []crack.Task {
	var tasks []crack.Task
	if flags.crackAll {
		for _, url := range flags.inputs {
			for _, user := range flags.userList {
				for _, pass := range flags.passList {
					tasks = append(tasks, crack.Task{
						URL:      url,
						Username: user,
						Password: pass,
					})
				}
			}
		}
	} else {
		for _, url := range flags.inputs {
			for i := range flags.userList {
				if i < len(flags.passList) {
					tasks = append(tasks, crack.Task{
						URL:      url,
						Username: flags.userList[i],
						Password: flags.passList[i],
					})
				}
			}
		}
	}
	return tasks
}

func GetSelector(ctx context.Context, url string) (s *browser.Selector, err error) {
	var data []byte
	var results []map[string]interface{}

	if globalOptions.selectorFile != "" {
		data, err = os.ReadFile(globalOptions.selectorFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read selector file: %w", err)
		}

		if err = yaml.Unmarshal(data, &s); err != nil {
			return nil, fmt.Errorf("failed to parse selector file: %w", err)
		}
	} else {
		var b *browser.Browser
		b, err = browser.New(globalOptions.headless, globalOptions.proxy, globalOptions.ocrURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create browser: %w", err)
		}

		// 释放资源
		defer b.Close()

		// 创建带有超时的上下文
		navigateCtx, cancel := context.WithTimeout(ctx, time.Duration(globalOptions.navigationTimeout)*time.Second)
		defer cancel()

		// 访问网站
		if err = b.Navigate(navigateCtx, url); err != nil {
			log.WithError(err).Errorf("Failed to navigate to URL: %s", url)
			return nil, err
		}

		// 探测选择器
		s, err = b.DetectFormSelectors()
		if err != nil {
			log.WithError(err).Errorf("Failed to detect_form_and_selectors for URL: %s", url)
			return nil, err
		}
	}

	// 打印探测信息
	result := map[string]interface{}{
		"url": url,
		"selectors": map[string]string{
			"userInput":     s.UserInput,
			"passwordInput": s.PasswordInput,
			"loginBtn":      s.LoginBtn,
			"rememberMe":    s.RememberMe,
		},
	}

	// 保存结果
	results = append(results, result)
	err = saveResults(results, globalOptions.outputFile)
	if err != nil {
		log.WithError(err).Errorf("Failed to save selector result for URL: %s", url)
		return nil, err
	}

	return s, nil
}

func Crack(ctx context.Context, task crack.Task, s *browser.Selector) {
	var b *browser.Browser
	var err error

	b, err = browser.New(globalOptions.headless, globalOptions.proxy, globalOptions.ocrURL)
	if err != nil {
		return
	}

	// 释放资源
	defer b.Close()

	// 超时上下文
	navigateCtx, cancel := context.WithTimeout(ctx, time.Duration(globalOptions.navigationTimeout)*time.Second)
	defer cancel()

	// 访问网站
	if err = b.Navigate(navigateCtx, task.URL); err != nil {
		log.WithError(err).Errorf("Failed to navigate to URL: %s", task.URL)
		return
	}

	// 初始化对象
	cracker := crack.New(
		globalOptions.delay,
		globalOptions.maxAttempts,
		globalOptions.maxCrackNum,
		globalOptions.maxCrackTime,
		globalOptions.threads,
		b,
		s,
	)

	crackCtx, crackCancel := context.WithTimeout(ctx, time.Duration(globalOptions.maxCrackTime)*time.Second)
	defer crackCancel()

	// 登录网站
	results := cracker.SingleTaskCrack(crackCtx, task)

	// 保存记录
	if len(results) > 0 {
		saveResults(results, globalOptions.outputFile)
	}

}

func run(options *Options) error {
	var err error

	for _, url := range options.inputs {
		var s *browser.Selector

		// 获取选择器
		s, err = GetSelector(gCtx, url)
		if err != nil {
			log.Fatal("try to get selector error.")
		}

		if s == nil {
			log.Fatal("get selector is nil.")
		}

		// 输出匹配信息
		selectorJSON, _ := json.Marshal(map[string]interface{}{
			"userInput":     s.UserInput,
			"passwordInput": s.PasswordInput,
			"loginBtn":      s.LoginBtn,
			"rememberMe":    s.RememberMe,
			"captchaInput":  s.CaptchaInput,
			"captchaImg":    s.CaptchaImg,
		})

		log.WithField("selector", string(selectorJSON)).Debug("found selectors successfully")

		// 仅探测
		if options.detectOnly {
			continue
		}

		for _, task := range CreateTasks(options) {
			Crack(gCtx, task, s)
		}

	}

	return nil
}

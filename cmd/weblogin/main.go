package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"zp-weblogin/pkg/browser"
	"zp-weblogin/pkg/config"
	"zp-weblogin/pkg/crack"
)

var rootCmd = &cobra.Command{
	Use:   "auto-web-login",
	Short: "Web login testing tool",
	Long:  "A tool for testing web login pages with automatic detection and manual configuration support",
}

var webLoginCmd = &cobra.Command{
	Use:     "weblogin",
	Short:   "Start web login testing",
	Example: "./auto-web-login weblogin -i http://example.com:9001",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Parse flags
		inputs, _ := cmd.Flags().GetStringSlice("inputs")
		inputsFile, _ := cmd.Flags().GetString("inputs-file")
		logLevel, _ := cmd.Flags().GetString("level")
		outputFile, _ := cmd.Flags().GetString("output-file")
		crackAll, _ := cmd.Flags().GetBool("crack-all")
		delay, _ := cmd.Flags().GetInt("delay")
		headless, _ := cmd.Flags().GetBool("headless")
		maxAttempts, _ := cmd.Flags().GetInt("max-attempts")
		maxCrackNum, _ := cmd.Flags().GetInt("max-crack-num")
		maxCrackTime, _ := cmd.Flags().GetInt("max-crack-time")
		passList, _ := cmd.Flags().GetStringSlice("pass")
		passFile, _ := cmd.Flags().GetString("pass-file")
		proxy, _ := cmd.Flags().GetString("proxy")
		selectorFile, _ := cmd.Flags().GetString("selector-file")
		threads, _ := cmd.Flags().GetInt("threads")
		userList, _ := cmd.Flags().GetStringSlice("user")
		userFile, _ := cmd.Flags().GetString("user-file")

		// Set log level
		level, err := logrus.ParseLevel(logLevel)
		if err != nil {
			return err
		}
		logrus.SetLevel(level)

		// Read inputs from file if specified
		if inputsFile != "" {
			data, err := os.ReadFile(inputsFile)
			if err != nil {
				return err
			}
			inputs = append(inputs, strings.Split(string(data), "\n")...)
		}

		// Read users from file if specified
		if userFile != "" {
			data, err := os.ReadFile(userFile)
			if err != nil {
				return err
			}
			userList = append(userList, strings.Split(string(data), "\n")...)
		}

		// Read passwords from file if specified
		if passFile != "" {
			data, err := os.ReadFile(passFile)
			if err != nil {
				return err
			}
			passList = append(passList, strings.Split(string(data), "\n")...)
		}

		// Initialize browser
		b, err := browser.New(headless, proxy)
		if err != nil {
			return err
		}
		defer b.Close()

		// Navigate to first URL to detect or load selectors
		if len(inputs) == 0 {
			return fmt.Errorf("no input URLs provided")
		}

		err = b.Navigate(inputs[0])
		if err != nil {
			return err
		}

		var selector *browser.Selector
		if selectorFile != "" {
			// Load selectors from file
			data, err := os.ReadFile(selectorFile)
			if err != nil {
				return err
			}
			var cfg config.SelectorConfig
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				return err
			}

			if cfg.UserInput == "" || cfg.PasswordInput == "" {
				return errors.New("cfg.UserInput == \"\" || cfg.PasswordInput == \"\", please check ")
			}

			selector = &browser.Selector{
				UserInput:     cfg.UserInput,
				PasswordInput: cfg.PasswordInput,
				LoginBtn:      cfg.LoginBtn,
			}
		} else {
			// Auto detect selectors
			selector, err = b.DetectSelectors()
			if err != nil {
				return err
			}
		}

		// Prepare tasks
		var tasks []crack.Task
		if crackAll {
			for _, url := range inputs {
				for _, user := range userList {
					for _, pass := range passList {
						tasks = append(tasks, crack.Task{
							URL:      url,
							Username: user,
							Password: pass,
						})
					}
				}
			}
		} else {
			for _, url := range inputs {
				for i := range userList {
					if i < len(passList) {
						tasks = append(tasks, crack.Task{
							URL:      url,
							Username: userList[i],
							Password: passList[i],
						})
					}
				}
			}
		}

		// Start cracking
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(maxCrackTime)*time.Second)
		defer cancel()

		cracker := crack.New(delay, maxAttempts, maxCrackNum, maxCrackTime, threads, b, selector)
		results := cracker.Start(ctx, tasks)

		// Write results
		file, err := os.Create(outputFile)
		if err != nil {
			return err
		}
		defer file.Close()

		return json.NewEncoder(file).Encode(results)
	},
}

func init() {
	rootCmd.AddCommand(webLoginCmd)

	webLoginCmd.Flags().StringSliceP("inputs", "i", nil, "inputs split by comma")
	webLoginCmd.Flags().StringP("inputs-file", "f", "", "inputs file split by line")
	webLoginCmd.Flags().String("level", "debug", "logger level(debug|info|error)")
	webLoginCmd.Flags().StringP("output-file", "o", "output.json", "output file to write results")

	webLoginCmd.Flags().Bool("crack-all", false, "crack all user and pass")
	webLoginCmd.Flags().Int("delay", 1, "delay time between crack")
	webLoginCmd.Flags().Bool("headless", false, "headless mode")
	webLoginCmd.Flags().Int("max-attempts", 3, "max attempts")
	webLoginCmd.Flags().Int("max-crack-num", 0, "max crack num, 0 is no limit")
	webLoginCmd.Flags().Int("max-crack-time", 300, "max crack time in sec")
	webLoginCmd.Flags().StringSlice("pass", nil, "pass list, split by comma")
	webLoginCmd.Flags().String("pass-file", "", "pass file")
	webLoginCmd.Flags().String("proxy", "", "proxy")
	webLoginCmd.Flags().String("selector-file", "", "selector file")
	webLoginCmd.Flags().Int("threads", 3, "scan threads")
	webLoginCmd.Flags().StringSlice("user", nil, "user list, split by comma")
	webLoginCmd.Flags().String("user-file", "", "user file")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

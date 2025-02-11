package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
	"xiaoyu/pkg/browser"
)

var gCtx = context.Background()
var rootCmd = &cobra.Command{
	Use:   "root",
	Short: "Web login testing tool",
	Long:  "A tool for testing web login pages with automatic detection and manual configuration support",
}

type Options struct {
	inputs            []string
	inputsFile        string
	logLevel          string
	outputFile        string
	crackAll          bool
	delay             int
	headless          bool
	detectOnly        bool
	maxAttempts       int
	maxCrackNum       int
	maxCrackTime      int
	loginTimeout      int
	elementTimeout    int
	navigationTimeout int
	passList          []string
	passFile          string
	proxy             string
	selectorFile      string
	threads           int
	userList          []string
	userFile          string
	ocrURL            string
}

var globalOptions = &Options{}

func setupLogging(level string) error {
	parsedLevel, err := log.ParseLevel(level)
	if err != nil {
		return err
	}
	log.SetLevel(parsedLevel)
	return nil
}

func loadConfig(flags *Options) (*Options, error) {
	if flags.inputsFile != "" {
		data, err := os.ReadFile(flags.inputsFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read inputs file: %w", err)
		}
		flags.inputs = append(flags.inputs, strings.Split(string(data), "\n")...)
	}

	if flags.userFile != "" {
		data, err := os.ReadFile(flags.userFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read user file: %w", err)
		}
		flags.userList = append(flags.userList, strings.Split(string(data), "\n")...)
	}

	if flags.passFile != "" {
		data, err := os.ReadFile(flags.passFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read password file: %w", err)
		}
		flags.passList = append(flags.passList, strings.Split(string(data), "\n")...)
	}

	if len(flags.inputs) == 0 {
		return nil, fmt.Errorf("no input URLs provided")
	}

	return flags, nil
}

func printConfiguration(flags *Options) {
	config := map[string]interface{}{
		"headlessOptions": map[string]interface{}{
			"headless":    flags.headless,
			"debug":       false,
			"maxAttempts": flags.maxAttempts,
			"ocrBaseURL":  flags.ocrURL,
		},
		"threads":      flags.threads,
		"crackAll":     flags.crackAll,
		"delay":        flags.delay,
		"maxCrackTime": flags.maxCrackTime,
	}

	jsonBytes, _ := json.Marshal(config)
	log.WithFields(log.Fields{
		"config": string(jsonBytes),
	}).Info("Configuration loaded")

	if flags.selectorFile != "" {
		data, _ := os.ReadFile(flags.selectorFile)
		var s browser.Selector
		_ = yaml.Unmarshal(data, &s)
		selectorJSON, _ := json.MarshalIndent(s, "", "  ")
		log.WithFields(log.Fields{
			"selectors": string(selectorJSON),
		}).Info("Login selectors detected")
	}
}

func saveResults(results interface{}, outputFile string) error {
	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(results); err != nil {
		return fmt.Errorf("failed to write results: %w", err)
	}
	return nil
}

func Execute() {
	rootFlags := rootCmd.PersistentFlags()

	// Browser automation flags
	rootFlags.BoolVar(&globalOptions.headless, "headless", false, "headless mode")
	rootFlags.StringVar(&globalOptions.proxy, "proxy", "", "proxy")
	rootFlags.BoolVar(&globalOptions.detectOnly, "detect-only", false, "only detect form elements without login")
	rootFlags.IntVar(&globalOptions.threads, "threads", 3, "scan threads")

	// Input flags
	rootFlags.StringSliceVarP(&globalOptions.inputs, "inputs", "i", nil, "inputs split by comma")
	rootFlags.StringVarP(&globalOptions.inputsFile, "inputs-file", "f", "", "inputs file split by line")
	rootFlags.StringVarP(&globalOptions.outputFile, "output-file", "o", "output.json", "output file to write results")
	rootFlags.StringVar(&globalOptions.logLevel, "level", "debug", "logger level(debug|info|error)")

	// required
	_ = rootCmd.MarkPersistentFlagRequired("inputs")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

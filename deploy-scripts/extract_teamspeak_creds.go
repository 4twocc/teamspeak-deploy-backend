package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run extract_teamspeak_creds.go <log_file>")
		os.Exit(1)
	}

	logFile := os.Args[1]

	// 检查日志文件是否存在
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		fmt.Printf("Error: Log file %s does not exist\n", logFile)
		os.Exit(1)
	}

	// 提取敏感信息
	creds, err := extractCredentials(logFile)
	if err != nil {
		fmt.Printf("Error extracting credentials: %v\n", err)
		os.Exit(1)
	}

	// 获取项目根目录
	projectRoot, err := getProjectRoot()
	if err != nil {
		fmt.Printf("Error getting project root: %v\n", err)
		os.Exit(1)
	}

	// 更新.env文件
	envFile := filepath.Join(projectRoot, "backend", ".env")
	err = updateEnvFile(envFile, creds)
	if err != nil {
		fmt.Printf("Error updating .env file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully extracted TeamSpeak credentials and updated .env file")
	fmt.Printf("ServerAdmin Password: %s\n", creds.Password)
	if creds.APIKey != "" {
		fmt.Printf("Server Query API Key: %s\n", creds.APIKey)
	}
	if creds.AdminToken != "" {
		fmt.Printf("Server Admin Token: %s\n", creds.AdminToken)
	}
}

type Credentials struct {
	Password   string
	APIKey     string
	AdminToken string
}

func extractCredentials(logFile string) (*Credentials, error) {
	file, err := os.Open(logFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	creds := &Credentials{}
	scanner := bufio.NewScanner(file)

	// 正则表达式匹配敏感信息
	passwordRegex := regexp.MustCompile(`password=\s*(\S+)`)
	apiKeyRegex := regexp.MustCompile(`API key:\s*(\S+)`)
	tokenRegex := regexp.MustCompile(`token=([a-zA-Z0-9+/=]+)`)

	for scanner.Scan() {
		line := scanner.Text()

		// 提取ServerAdmin密码
		if creds.Password == "" {
			if matches := passwordRegex.FindStringSubmatch(line); len(matches) > 1 {
				creds.Password = matches[1]
			}
		}

		// 提取API密钥
		if creds.APIKey == "" {
			if matches := apiKeyRegex.FindStringSubmatch(line); len(matches) > 1 {
				creds.APIKey = matches[1]
			}
		}

		// 提取管理员令牌
		if creds.AdminToken == "" {
			if matches := tokenRegex.FindStringSubmatch(line); len(matches) > 1 {
				creds.AdminToken = matches[1]
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading log file: %w", err)
	}

	if creds.Password == "" {
		return nil, fmt.Errorf("serveradmin password not found in log file")
	}

	return creds, nil
}

func getProjectRoot() (string, error) {
	// 获取当前工作目录
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// 向上查找直到找到go.mod文件或达到文件系统根目录
	current := wd
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			// 已达到文件系统根目录
			break
		}
		current = parent
	}

	// 如果没有找到go.mod，使用当前工作目录的父目录作为项目根目录
	return filepath.Dir(wd), nil
}

func updateEnvFile(envFile string, creds *Credentials) error {
	// 读取现有.env文件内容（如果存在）
	var lines []string
	if _, err := os.Stat(envFile); err == nil {
		file, err := os.Open(envFile)
		if err != nil {
			return fmt.Errorf("failed to open .env file: %w", err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			// 跳过已存在的TeamSpeak凭证行
			if strings.HasPrefix(line, "TEAMSPEAK_PASSWORD=") ||
				strings.HasPrefix(line, "TEAMSPEAK_SERVER_QUERY_APIKEY=") ||
				strings.HasPrefix(line, "TEAMSPEAK_SERVER_ADMIN_TOKEN=") {
				continue
			}
			lines = append(lines, line)
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading .env file: %w", err)
		}
	}

	// 添加新的TeamSpeak凭证
	lines = append(lines, "")
	lines = append(lines, "# TeamSpeak Credentials (auto-generated)")
	lines = append(lines, fmt.Sprintf("TEAMSPEAK_PASSWORD=%s", creds.Password))
	if creds.APIKey != "" {
		lines = append(lines, fmt.Sprintf("TEAMSPEAK_SERVER_QUERY_APIKEY=%s", creds.APIKey))
	}
	if creds.AdminToken != "" {
		lines = append(lines, fmt.Sprintf("TEAMSPEAK_SERVER_ADMIN_TOKEN=%s", creds.AdminToken))
	}
	lines = append(lines, "")

	// 写入更新后的内容
	file, err := os.Create(envFile)
	if err != nil {
		return fmt.Errorf("failed to create .env file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return fmt.Errorf("error writing to .env file: %w", err)
		}
	}

	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("error flushing .env file: %w", err)
	}

	return nil
}

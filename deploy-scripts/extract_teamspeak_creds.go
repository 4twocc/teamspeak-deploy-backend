package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Credentials 敏感信息结构体
type Credentials struct {
	Password   string
	APIKey     string
	AdminToken string
}

func main() {
	// 检查参数数量，如果只有一个参数，则只生成JWT密钥
	if len(os.Args) == 1 {
		// 只生成JWT密钥并输出到stdout
		jwtSecret, err := generateJWTSecret()
		if err != nil {
			fmt.Printf("Error generating JWT secret: %v\n", err)
			os.Exit(1)
		}
		fmt.Print(jwtSecret)
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: go run extract_teamspeak_creds.go <log_file>")
		fmt.Println("       go run extract_teamspeak_creds.go (to generate JWT secret only)")
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

	// 获取项目根目录路径
	projectRoot := filepath.Dir(filepath.Dir(os.Args[0]))
	envFile := filepath.Join(projectRoot, "backend", ".env")

	// 更新.env文件
	err = updateEnvFile(envFile, creds, projectRoot)
	if err != nil {
		fmt.Printf("Error updating .env file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Successfully updated .env file with credentials and generated JWT secret")
}

// generateJWTSecret 使用openssl生成JWT密钥
func generateJWTSecret() (string, error) {
	cmd := exec.Command("openssl", "rand", "-base64", "32")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to generate JWT secret: %w", err)
	}
	
	return strings.TrimSpace(string(output)), nil
}

// extractCredentials 从日志文件中提取敏感信息
func extractCredentials(logFile string) (*Credentials, error) {
	content, err := os.ReadFile(logFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read log file: %w", err)
	}

	creds := &Credentials{}

	// 提取serveradmin密码
	passwordRegex := regexp.MustCompile(`password=\s*(\S+)`)
	passwordMatches := passwordRegex.FindStringSubmatch(string(content))
	if len(passwordMatches) > 1 {
		creds.Password = passwordMatches[1]
	}

	// 提取server query API key
	apiKeyRegex := regexp.MustCompile(`API key:\s*(\S+)`)
	apiKeyMatches := apiKeyRegex.FindStringSubmatch(string(content))
	if len(apiKeyMatches) > 1 {
		creds.APIKey = apiKeyMatches[1]
	}

	// 提取server admin token
	tokenRegex := regexp.MustCompile(`token=([^\s]+)`)
	tokenMatches := tokenRegex.FindStringSubmatch(string(content))
	if len(tokenMatches) > 1 {
		creds.AdminToken = tokenMatches[1]
	}

	return creds, nil
}

// updateEnvFile 更新.env文件
func updateEnvFile(envFile string, creds *Credentials, projectRoot string) error {
	var lines []string
	
	// 检查.env文件是否存在
	if _, err := os.Stat(envFile); os.IsNotExist(err) {
		// 如果.env文件不存在，从.env.example复制
		envExampleFile := filepath.Join(projectRoot, "backend", ".env.example")
		if _, err := os.Stat(envExampleFile); err == nil {
			// .env.example存在，复制其内容
			content, err := os.ReadFile(envExampleFile)
			if err != nil {
				return fmt.Errorf("failed to read .env.example file: %w", err)
			}
			lines = strings.Split(string(content), "\n")
		} else {
			// .env.example也不存在，创建空的lines
			lines = []string{}
		}
	} else {
		// .env文件存在，读取现有内容（但排除敏感配置）
		file, err := os.Open(envFile)
		if err != nil {
			return fmt.Errorf("failed to open .env file: %w", err)
		}
		defer file.Close()
		
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			// 跳过敏感配置行
			if strings.HasPrefix(line, "TEAMSPEAK_PASSWORD=") || 
			   strings.HasPrefix(line, "TEAMSPEAK_SERVER_QUERY_APIKEY=") || 
			   strings.HasPrefix(line, "TEAMSPEAK_SERVER_ADMIN_TOKEN=") ||
			   strings.HasPrefix(line, "JWT_SECRET=") {
				continue
			}
			lines = append(lines, line)
		}
		
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("error reading .env file: %w", err)
		}
	}
	
	// 添加TeamSpeak凭证
	lines = append(lines, "")
	lines = append(lines, "# TeamSpeak Credentials (auto-generated)")
	lines = append(lines, fmt.Sprintf("TEAMSPEAK_SERVER_ADMIN_USERNAME=%s", "serveradmin"))
	lines = append(lines, fmt.Sprintf("TEAMSPEAK_PASSWORD=%s", creds.Password))
	if creds.APIKey != "" {
		lines = append(lines, fmt.Sprintf("TEAMSPEAK_SERVER_QUERY_APIKEY=%s", creds.APIKey))
	}
	if creds.AdminToken != "" {
		lines = append(lines, fmt.Sprintf("TEAMSPEAK_SERVER_ADMIN_TOKEN=%s", creds.AdminToken))
	}
	lines = append(lines, "")

	// 自动生成JWT_SECRET
	lines = append(lines, "# JWT Secret (auto-generated)")
	jwtSecret, err := generateJWTSecret()
	if err != nil {
		return fmt.Errorf("failed to generate JWT secret: %w", err)
	}
	lines = append(lines, fmt.Sprintf("JWT_SECRET=%s", jwtSecret))
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
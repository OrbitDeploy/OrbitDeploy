package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// 辅助函数用于从map中安全获取值
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
		if i, ok := v.(int); ok {
			return float64(i)
		}
	}
	return 0
}

func getMap(m map[string]interface{}, key string) map[string]interface{} {
	if v, ok := m[key]; ok {
		if subMap, ok := v.(map[string]interface{}); ok {
			return subMap
		}
	}
	return make(map[string]interface{})
}

// 辅助函数

func getInputWithDefault(prompt, defaultValue string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("%s [%s]: ", prompt, defaultValue)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	if input == "" {
		return defaultValue
	}
	return input
}

func getOrDefault(value, defaultValue string) string {
	if value == "" {
		return defaultValue
	}
	return value
}

// getProjectFromConfig 从配置文件中获取项目名称
func getProjectFromConfig() string {
	spec, err := loadSpecFromFile("orbitdeploy.toml")
	if err != nil {
		return ""
	}
	return spec.Project
}

func loadSpecFromFile(filename string) (*specTOML, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", filename)
	}

	// 读取并解析文件
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	spec, err := parseSpecTOML(b)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return spec, nil
}

// getAPIBase 获取API基础URL
func getAPIBase() string {
	// Return normalized server base (without trailing /api)
	return serverBase()
}

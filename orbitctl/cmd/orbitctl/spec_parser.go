package main

import (
	"errors"
	"fmt"
	"regexp"

	toml "github.com/pelletier/go-toml/v2"
)

// 根据ORBITDEPLOY_SPEC_TOML_SCHEMA_AND_PARSER.md定义的结构

type imageCfg struct {
	Ref    string `toml:"ref,omitempty"`
	Digest string `toml:"digest,omitempty"`
}

type setRef struct {
	SetRef string `toml:"set_ref,omitempty"`
}

type resourcesCfg struct {
	CPU    string `toml:"cpu,omitempty"`
	Memory string `toml:"memory,omitempty"`
}

type healthcheckCfg struct {
	Kind           string `toml:"kind,omitempty"`
	Path           string `toml:"path,omitempty"`
	TimeoutSeconds int    `toml:"timeout_seconds,omitempty"`
	Retries        int    `toml:"retries,omitempty"`
}

type domainCfg struct {
	Host string `toml:"host"`
	Path string `toml:"path,omitempty"`
}

type volumeCfg struct {
	Source   string `toml:"source"`
	Target   string `toml:"target"`
	ReadOnly bool   `toml:"read_only,omitempty"`
}

type containerTOML struct {
	Name        string            `toml:"name"`
	PublishPort int               `toml:"publish_port,omitempty"`
	Image       *imageCfg         `toml:"image,omitempty"`
	Resources   *resourcesCfg     `toml:"resources,omitempty"`
	Healthcheck *healthcheckCfg   `toml:"healthcheck,omitempty"`
	Domains     []domainCfg       `toml:"domains,omitempty"`
	Volumes     []volumeCfg       `toml:"volumes,omitempty"`
	ExtraEnv    map[string]string `toml:"extra_env,omitempty"`
	Labels      map[string]string `toml:"labels,omitempty"`
	Annotations map[string]string `toml:"annotations,omitempty"`
}

type specTOML struct {
	APIVersion  string          `toml:"api_version"`
	Kind        string          `toml:"kind"`
	AppName     string          `toml:"app_name"`     // 新增：应用名称，优先使用
	Project     string          `toml:"project"`
	Environment string          `toml:"environment"`
	Name        string          `toml:"name"`         // 保留向后兼容，当 app_name 为空时使用
	Strategy    string          `toml:"strategy,omitempty"`
	Replicas    int             `toml:"replicas,omitempty"`
	Image       *imageCfg       `toml:"image,omitempty"`
	Env         *setRef         `toml:"env,omitempty"`
	Secret      *setRef         `toml:"secret,omitempty"`
	Resources   *resourcesCfg   `toml:"resources,omitempty"`
	Healthcheck *healthcheckCfg `toml:"healthcheck,omitempty"`
	Containers  []containerTOML `toml:"containers"`
}

// 验证器正则表达式
var (
	namePattern     = regexp.MustCompile(`^[a-z0-9-]+$`)
	validStrategies = map[string]bool{
		"direct": true, "blue-green": true, "rolling": true, "canary": true,
	}
	validHealthcheckKinds = map[string]bool{
		"http": true, "tcp": true, "cmd": true,
	}
)

func parseSpecTOML(b []byte) (*specTOML, error) {
	var s specTOML
	if err := toml.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("TOML解析失败: %w", err)
	}

	// 基础字段验证
	// if s.APIVersion != "webdeploy.io/v1" {
	// 	return nil, errors.New("api_version 必须为 webdeploy.io/v1")
	// }
	if s.Kind != "DeploymentSpec" {
		return nil, errors.New("kind 必须为 DeploymentSpec")
	}

	// 应用名称处理：优先使用 app_name，否则使用 name
	if s.AppName != "" {
		s.Name = s.AppName // 统一使用 Name 字段存储应用名称
	}
	
	if s.Name == "" {
		return nil, errors.New("app_name 或 name 为必填字段")
	}

	// 对于 CLI 部署，project 和 environment 可以是可选的
	// 如果未提供，使用默认值
	if s.Project == "" {
		s.Project = "default"
	}
	if s.Environment == "" {
		s.Environment = "production"
	}

	// 名称格式验证
	if !namePattern.MatchString(s.Name) {
		return nil, errors.New("app_name/name 格式无效，必须为小写字母数字和连字符")
	}
	if !namePattern.MatchString(s.Project) {
		return nil, errors.New("project 名称格式无效，必须为小写字母数字和连字符")
	}
	if !namePattern.MatchString(s.Environment) {
		return nil, errors.New("environment 名称格式无效，必须为小写字母数字和连字符")
	}

	// 策略验证
	if s.Strategy != "" && !validStrategies[s.Strategy] {
		return nil, fmt.Errorf("strategy 无效: %s，支持的策略: direct, blue-green, rolling, canary", s.Strategy)
	}

	// 副本数验证
	if s.Replicas < 0 {
		return nil, errors.New("replicas 不能为负数")
	}

	// 容器验证
	if len(s.Containers) == 0 {
		return nil, errors.New("至少需要配置一个容器")
	}

	for i, c := range s.Containers {
		if c.Name == "" {
			return nil, fmt.Errorf("containers[%d].name 为必填字段", i)
		}
		if !namePattern.MatchString(c.Name) {
			return nil, fmt.Errorf("containers[%d].name 格式无效，必须为小写字母数字和连字符", i)
		}

		// 端口验证
		if c.PublishPort != 0 && (c.PublishPort < 1 || c.PublishPort > 65535) {
			return nil, fmt.Errorf("containers[%d].publish_port 端口范围无效(1-65535)", i)
		}

		// 健康检查验证
		if c.Healthcheck != nil {
			if c.Healthcheck.Kind != "" && !validHealthcheckKinds[c.Healthcheck.Kind] {
				return nil, fmt.Errorf("containers[%d].healthcheck.kind 无效: %s", i, c.Healthcheck.Kind)
			}
			if c.Healthcheck.TimeoutSeconds < 0 {
				return nil, fmt.Errorf("containers[%d].healthcheck.timeout_seconds 不能为负数", i)
			}
			if c.Healthcheck.Retries < 0 {
				return nil, fmt.Errorf("containers[%d].healthcheck.retries 不能为负数", i)
			}
		}

		// 域名验证
		for j, d := range c.Domains {
			if d.Host == "" {
				return nil, fmt.Errorf("containers[%d].domains[%d].host 为必填字段", i, j)
			}
		}

		// 卷挂载验证
		for j, v := range c.Volumes {
			if v.Source == "" {
				return nil, fmt.Errorf("containers[%d].volumes[%d].source 为必填字段", i, j)
			}
			if v.Target == "" {
				return nil, fmt.Errorf("containers[%d].volumes[%d].target 为必填字段", i, j)
			}
		}
	}

	// 全局健康检查验证
	if s.Healthcheck != nil {
		if s.Healthcheck.Kind != "" && !validHealthcheckKinds[s.Healthcheck.Kind] {
			return nil, fmt.Errorf("healthcheck.kind 无效: %s", s.Healthcheck.Kind)
		}
		if s.Healthcheck.TimeoutSeconds < 0 {
			return nil, errors.New("healthcheck.timeout_seconds 不能为负数")
		}
		if s.Healthcheck.Retries < 0 {
			return nil, errors.New("healthcheck.retries 不能为负数")
		}
	}

	return &s, nil
}

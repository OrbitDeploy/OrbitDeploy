package main

import (
	"fmt"
	"os"
)

func cmdSpecValidate(file string) error {
	// 检查文件是否存在
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", file)
	}
	
	// 读取文件内容
	b, err := os.ReadFile(file)
	if err != nil { 
		return fmt.Errorf("读取文件失败: %w", err)
	}
	
	// 解析TOML
	spec, err := parseSpecTOML(b)
	if err != nil { 
		return fmt.Errorf("配置验证失败: %w", err)
	}
	
	// 进行更深入的业务逻辑验证
	if err := validateBusinessLogic(spec); err != nil {
		return fmt.Errorf("业务逻辑验证失败: %w", err)
	}
	
	fmt.Printf("✅ 配置文件 %s 验证通过\n", file)
	fmt.Printf("   项目: %s, 环境: %s, 应用: %s\n", spec.Project, spec.Environment, spec.Name)
	fmt.Printf("   容器数量: %d\n", len(spec.Containers))
	
	return nil
}

func validateBusinessLogic(spec *specTOML) error {
	// 验证域名配置
	domainSet := make(map[string]bool)
	for i, c := range spec.Containers {
		for j, d := range c.Domains {
			if d.Host == "" { 
				return fmt.Errorf("容器[%d].域名[%d].host 不能为空", i, j) 
			}
			
			// 检查域名重复
			domainKey := d.Host + d.Path
			if domainSet[domainKey] {
				return fmt.Errorf("重复的域名配置: %s%s", d.Host, d.Path)
			}
			domainSet[domainKey] = true
			
			// 基本域名格式检查
			if len(d.Host) > 253 {
				return fmt.Errorf("域名过长: %s", d.Host)
			}
		}
	}
	
	// 验证端口冲突
	portSet := make(map[int]bool)
	for i, c := range spec.Containers {
		if c.PublishPort > 0 {
			if portSet[c.PublishPort] {
				return fmt.Errorf("端口冲突: 容器[%d]的端口 %d 已被使用", i, c.PublishPort)
			}
			portSet[c.PublishPort] = true
		}
	}
	
	// 验证卷挂载路径
	for i, c := range spec.Containers {
		targetSet := make(map[string]bool)
		for j, v := range c.Volumes {
			if targetSet[v.Target] {
				return fmt.Errorf("容器[%d]卷挂载目标路径重复: %s", i, v.Target)
			}
			targetSet[v.Target] = true
			
			// TODO: 在实际环境中可以验证源路径的存在性和权限
			if len(v.Target) == 0 {
				return fmt.Errorf("容器[%d].卷[%d].target 不能为空", i, j)
			}
		}
	}
	
	return nil
}

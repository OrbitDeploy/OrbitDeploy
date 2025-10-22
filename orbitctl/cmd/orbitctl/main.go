package main

import (
	"flag"
	"fmt"
	"os"
)

func usage() {
	fmt.Println("orbitctl - OrbitDeploy CLI (MVP)\n")
	fmt.Println("用法:")
	fmt.Println("  orbitctl auth login    [-u 用户名 -p 密码 --token 令牌 --api-base URL]")
	fmt.Println("  orbitctl auth logout")
	fmt.Println("  orbitctl auth refresh")
	fmt.Println("  orbitctl init          [--name 应用名] [--project 项目名] [--env 环境名]")
	fmt.Println("  orbitctl spec-validate [-f 文件]")
	fmt.Println("  orbitctl deploy        [--project 项目名] [--env 环境名] [--dry-run]")
	fmt.Println("  orbitctl env list      [--project 项目名] [--env 环境名]")
	fmt.Println("  orbitctl env set       KEY=VALUE [--project 项目名] [--env 环境名]")
	fmt.Println("  orbitctl env unset     KEY [--project 项目名] [--env 环境名]")
	fmt.Println("  orbitctl scale         副本数 [--project 项目名] [--env 环境名]")

	fmt.Println("  orbitctl status        [--project 项目名] [--env 环境名]")
	fmt.Println("  orbitctl logs          [-f] [--project 项目名] [--env 环境名]")
	fmt.Println("  orbitctl inspect       [--project 项目名] [--env 环境名]")
	fmt.Println("")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	// 支持 --help 和 help 命令
	if os.Args[1] == "--help" || os.Args[1] == "help" {
		usage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "auth":
		if len(os.Args) < 3 {
			usage()
			os.Exit(1)
		}
		sub := os.Args[2]
		switch sub {
		case "login":
			loginCmd := flag.NewFlagSet("login", flag.ExitOnError)
			user := loginCmd.String("u", "", "用户名")
			pass := loginCmd.String("p", "", "密码")
			token := loginCmd.String("token", "", "直接使用访问令牌登录")
			apiBase := loginCmd.String("api-base", os.Getenv("ORBIT_API_BASE"), "API服务器地址")
			_ = loginCmd.Parse(os.Args[3:])
			if err := cmdAuthLogin(*user, *pass, *token, *apiBase); err != nil {
				fmt.Fprintf(os.Stderr, "登录失败: %v\n", err)
				os.Exit(1)
			}
		case "logout":
			if err := cmdAuthLogout(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		case "refresh":
			if err := cmdAuthRefresh(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
		default:
			usage()
			os.Exit(1)
		}
	case "init":
		initCmd := flag.NewFlagSet("init", flag.ExitOnError)
		name := initCmd.String("name", "", "应用名称")
		project := initCmd.String("project", "", "项目名称")
		env := initCmd.String("env", "dev", "环境名称")
		_ = initCmd.Parse(os.Args[2:])
		if err := cmdInit(*name, *project, *env); err != nil {
			fmt.Fprintf(os.Stderr, "初始化失败: %v\n", err)
			os.Exit(1)
		}
	case "spec-validate":
		fs := flag.NewFlagSet("spec-validate", flag.ExitOnError)
		file := fs.String("f", "orbitctl.toml", "配置文件路径")
		_ = fs.Parse(os.Args[2:])
		if err := cmdSpecValidate(*file); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	case "deploy":
		deployCmd := flag.NewFlagSet("deploy", flag.ExitOnError)
		project := deployCmd.String("project", "", "项目名称")
		env := deployCmd.String("env", "dev", "环境名称")
		dryRun := deployCmd.Bool("dry-run", false, "仅显示部署计划，不实际执行")
		_ = deployCmd.Parse(os.Args[2:])
		if err := cmdDeploy(*project, *env, *dryRun); err != nil {
			fmt.Fprintf(os.Stderr, "部署失败: %v\n", err)
			os.Exit(1)
		}
	case "env":
		if len(os.Args) < 3 {
			usage()
			os.Exit(1)
		}
		sub := os.Args[2]
		switch sub {
		case "list":
			envCmd := flag.NewFlagSet("env-list", flag.ExitOnError)
			project := envCmd.String("project", "", "项目名称")
			env := envCmd.String("env", "dev", "环境名称")
			_ = envCmd.Parse(os.Args[3:])
			if err := cmdEnvList(*project, *env); err != nil {
				fmt.Fprintf(os.Stderr, "获取环境变量失败: %v\n", err)
				os.Exit(1)
			}
		case "set":
			if len(os.Args) < 4 {
				usage()
				os.Exit(1)
			}
			envCmd := flag.NewFlagSet("env-set", flag.ExitOnError)
			project := envCmd.String("project", "", "项目名称")
			env := envCmd.String("env", "dev", "环境名称")
			_ = envCmd.Parse(os.Args[4:])
			if err := cmdEnvSet(os.Args[3], *project, *env); err != nil {
				fmt.Fprintf(os.Stderr, "设置环境变量失败: %v\n", err)
				os.Exit(1)
			}
		case "unset":
			if len(os.Args) < 4 {
				usage()
				os.Exit(1)
			}
			envCmd := flag.NewFlagSet("env-unset", flag.ExitOnError)
			project := envCmd.String("project", "", "项目名称")
			env := envCmd.String("env", "dev", "环境名称")
			_ = envCmd.Parse(os.Args[4:])
			if err := cmdEnvUnset(os.Args[3], *project, *env); err != nil {
				fmt.Fprintf(os.Stderr, "删除环境变量失败: %v\n", err)
				os.Exit(1)
			}
		default:
			usage()
			os.Exit(1)
		}
	case "scale":
		if len(os.Args) < 3 {
			usage()
			os.Exit(1)
		}
		scaleCmd := flag.NewFlagSet("scale", flag.ExitOnError)
		project := scaleCmd.String("project", "", "项目名称")
		env := scaleCmd.String("env", "dev", "环境名称")
		_ = scaleCmd.Parse(os.Args[3:])
		if err := cmdScale(os.Args[2], *project, *env); err != nil {
			fmt.Fprintf(os.Stderr, "扩缩容失败: %v\n", err)
			os.Exit(1)
		}
	case "status":
		statusCmd := flag.NewFlagSet("status", flag.ExitOnError)
		project := statusCmd.String("project", "", "项目名称")
		env := statusCmd.String("env", "dev", "环境名称")
		_ = statusCmd.Parse(os.Args[2:])
		if err := cmdStatus(*project, *env); err != nil {
			fmt.Fprintf(os.Stderr, "获取状态失败: %v\n", err)
			os.Exit(1)
		}
	case "logs":
		logsCmd := flag.NewFlagSet("logs", flag.ExitOnError)
		follow := logsCmd.Bool("f", false, "持续跟踪日志")
		project := logsCmd.String("project", "", "项目名称")
		env := logsCmd.String("env", "dev", "环境名称")
		_ = logsCmd.Parse(os.Args[2:])
		if err := cmdLogs(*follow, *project, *env); err != nil {
			fmt.Fprintf(os.Stderr, "获取日志失败: %v\n", err)
			os.Exit(1)
		}
	case "inspect":
		inspectCmd := flag.NewFlagSet("inspect", flag.ExitOnError)
		project := inspectCmd.String("project", "", "项目名称")
		env := inspectCmd.String("env", "dev", "环境名称")
		_ = inspectCmd.Parse(os.Args[2:])
		if err := cmdInspect(*project, *env); err != nil {
			fmt.Fprintf(os.Stderr, "检查配置失败: %v\n", err)
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(1)
	}
}

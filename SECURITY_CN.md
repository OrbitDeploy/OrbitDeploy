# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| Latest  | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of OrbitDeploy seriously. If you have discovered a security vulnerability, we appreciate your help in disclosing it to us in a responsible manner.

### How to Report

**Please do NOT report security vulnerabilities through public GitHub issues.**

Instead, please report them via one of the following methods:

1. **GitHub Security Advisories**: Use the [GitHub Security Advisory](https://github.com/OrbitDeploy/OrbitDeploy/security/advisories/new) feature
2. **Email**: Send details to the maintainers (contact information can be found in the repository)

### What to Include

When reporting a vulnerability, please include the following information:

- Type of vulnerability (e.g., SQL injection, cross-site scripting, authentication bypass, etc.)
- Full paths of source file(s) related to the manifestation of the vulnerability
- The location of the affected source code (tag/branch/commit or direct URL)
- Any special configuration required to reproduce the issue
- Step-by-step instructions to reproduce the issue
- Proof-of-concept or exploit code (if possible)
- Impact of the vulnerability, including how an attacker might exploit it

### Response Timeline

- We will acknowledge receipt of your vulnerability report within 48 hours
- We will provide a more detailed response within 7 days indicating the next steps in handling your report
- We will keep you informed about the progress towards a fix and full announcement
- We may ask for additional information or guidance

## Security Best Practices for Deployment

When deploying OrbitDeploy in production, please follow these security best practices:

### 1. Environment Variables

**Always** use environment variables for sensitive configuration:

```bash
# Required for production
export ORBIT_ENCRYPTION_KEY="your-secure-random-key-here"
export JWT_ACCESS_SECRET="your-jwt-access-secret-here"
export JWT_REFRESH_SECRET="your-jwt-refresh-secret-here"
```

See [Environment Variables Configuration](DOC/environment-variables.md) for complete details.

### 2. Generate Strong Secrets

Use cryptographically secure random generators for all secrets:

```bash
# Generate a secure key
openssl rand -base64 32

# Or using Python
python3 -c "import secrets; print(secrets.token_urlsafe(32))"

# Or using Node.js
node -e "console.log(require('crypto').randomBytes(32).toString('base64'))"
```

### 3. Network Security

- Run OrbitDeploy behind a reverse proxy (like Caddy, Nginx, or Traefik)
- Enable HTTPS/TLS for all connections
- Use firewall rules to restrict access to necessary ports only
- Consider using a VPN or IP allowlist for administrative access

### 4. File Permissions

- Ensure the database file has appropriate permissions (600 or 640)
- Restrict access to configuration files containing secrets
- Run the application with a dedicated user account (not root)

### 5. Regular Updates

- Keep OrbitDeploy updated to the latest version
- Monitor the repository for security advisories
- Subscribe to release notifications

### 6. Backup and Recovery

- Regularly backup your database
- Encrypt backups if they contain sensitive data
- Test your recovery procedures

### 7. Monitoring and Logging

- Enable appropriate logging levels
- Monitor logs for suspicious activity
- Set up alerts for authentication failures
- Regularly review access logs

## Known Security Considerations

### Encryption Key Rotation

If you need to rotate the `ORBIT_ENCRYPTION_KEY`:

1. Existing encrypted data will not be decryptable with the new key
2. Plan for a migration strategy if you need to preserve encrypted data
3. Consider implementing a key rotation procedure during maintenance windows

### JWT Token Security

- Access tokens expire after 15 minutes by default (configurable via `JWT_ACCESS_TTL`)
- Refresh tokens expire after 30 days by default (configurable via `JWT_REFRESH_TTL`)
- Tokens are signed but not encrypted - avoid storing sensitive data in JWT claims
- Implement token revocation if needed for your use case

### Container Security

- OrbitDeploy manages containers via Podman - ensure Podman is properly secured
- Review container security best practices for your deployment
- Keep container images updated
- Use trusted image sources

## Disclosure Policy

When we receive a security bug report, we will:

1. Confirm the problem and determine affected versions
2. Audit code to find any similar problems
3. Prepare fixes for all supported versions
4. Release new security fix versions as soon as possible

## Comments on this Policy

If you have suggestions on how this process could be improved, please submit a pull request or open an issue.

## Attribution

This security policy is adapted from the [GitHub Security Policy Template](https://docs.github.com/en/code-security/getting-started/adding-a-security-policy-to-your-repository).

---

## Chinese

### 安全政策

#### 支持的版本

我们为以下版本发布安全漏洞补丁：

| 版本 | 支持情况          |
| ---- | ---------------- |
| 最新版 | :white_check_mark: |
| < 1.0 | :x:               |

#### 报告漏洞

我们非常重视 OrbitDeploy 的安全性。如果您发现了安全漏洞，我们感谢您以负责任的方式向我们披露。

##### 如何报告

**请勿通过公共 GitHub 问题报告安全漏洞。**

请通过以下方式之一报告：

1. **GitHub 安全建议**：使用 [GitHub 安全建议](https://github.com/OrbitDeploy/OrbitDeploy/security/advisories/new) 功能
2. **电子邮件**：发送详细信息给维护者（联系信息可在仓库中找到）

##### 应包含的内容

报告漏洞时，请包括以下信息：

- 漏洞类型（例如，SQL 注入、跨站脚本、身份验证绕过等）
- 与漏洞表现相关的源文件完整路径
- 受影响源代码的位置（标签/分支/提交或直接 URL）
- 重现问题所需的任何特殊配置
- 重现问题的逐步说明
- 概念验证或利用代码（如果可能）
- 漏洞的影响，包括攻击者可能如何利用它

##### 响应时间表

- 我们将在 48 小时内确认收到您的漏洞报告
- 我们将在 7 天内提供更详细的响应，说明处理报告的下一步
- 我们将让您了解修复和完整公告的进展
- 我们可能会要求额外信息或指导

#### 部署的安全最佳实践

在生产环境中部署 OrbitDeploy 时，请遵循以下安全最佳实践：

##### 1. 环境变量

**始终** 为敏感配置使用环境变量：

```bash
# 生产环境必需
export ORBIT_ENCRYPTION_KEY="your-secure-random-key-here"
export JWT_ACCESS_SECRET="your-jwt-access-secret-here"
export JWT_REFRESH_SECRET="your-jwt-refresh-secret-here"
```

有关完整详细信息，请参阅[环境变量配置](DOC/environment-variables.md)。

##### 2. 生成强密钥

为所有密钥使用加密安全的随机生成器：

```bash
# 生成安全密钥
openssl rand -base64 32

# 或使用 Python
python3 -c "import secrets; print(secrets.token_urlsafe(32))"

# 或使用 Node.js
node -e "console.log(require('crypto').randomBytes(32).toString('base64'))"
```

##### 3. 网络安全

- 在反向代理后面运行 OrbitDeploy（如 Caddy、Nginx 或 Traefik）
- 为所有连接启用 HTTPS/TLS
- 使用防火墙规则仅限制对必要端口的访问
- 考虑为管理访问使用 VPN 或 IP 白名单

##### 4. 文件权限

- 确保数据库文件具有适当权限（600 或 640）
- 限制对包含密钥的配置文件访问
- 使用专用用户账户运行应用程序（非 root）

##### 5. 定期更新

- 保持 OrbitDeploy 更新到最新版本
- 监控仓库的安全建议
- 订阅发布通知

##### 6. 备份和恢复

- 定期备份您的数据库
- 如果备份包含敏感数据，则加密备份
- 测试您的恢复程序

##### 7. 监控和日志记录

- 启用适当的日志级别
- 监控日志以查找可疑活动
- 为身份验证失败设置警报
- 定期审查访问日志

#### 已知安全注意事项

##### 加密密钥轮换

如果您需要轮换 `ORBIT_ENCRYPTION_KEY`：

1. 现有加密数据将无法使用新密钥解密
2. 如果需要保留加密数据，请规划迁移策略
3. 考虑在维护窗口期间实施密钥轮换程序

##### JWT 令牌安全

- 访问令牌默认在 15 分钟后过期（可通过 `JWT_ACCESS_TTL` 配置）
- 刷新令牌默认在 30 天后过期（可通过 `JWT_REFRESH_TTL` 配置）
- 令牌已签名但未加密 - 避免在 JWT 声明中存储敏感数据
- 根据需要实施令牌撤销

##### 容器安全

- OrbitDeploy 通过 Podman 管理容器 - 确保 Podman 已正确安全
- 为您的部署审查容器安全最佳实践
- 保持容器镜像更新
- 使用可信镜像源

#### 披露政策

当我们收到安全漏洞报告时，我们将：

1. 确认问题并确定受影响版本
2. 审核代码以查找任何类似问题
3. 为所有支持版本准备修复
4. 尽快发布新的安全修复版本

#### 对本政策的评论

如果您对如何改进此过程有建议，请提交拉取请求或打开问题。

#### 归属

此安全政策改编自 [GitHub 安全政策模板](https://docs.github.com/en/code-security/getting-started/adding-a-security-policy-to-your-repository)

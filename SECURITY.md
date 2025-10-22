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

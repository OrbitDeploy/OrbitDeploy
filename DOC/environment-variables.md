# Environment Variables Configuration

This document describes the environment variables used by OrbitDeploy for configuration.

## Security-Critical Environment Variables

### Encryption Key

**ORBIT_ENCRYPTION_KEY**
- **Description**: Master encryption key used for encrypting sensitive data like application tokens and credentials
- **Required**: Highly recommended for production
- **Default**: A default development key (insecure - should be changed in production)
- **Format**: Any string (will be hashed with SHA-256 to derive the actual encryption key)
- **Example**: `export ORBIT_ENCRYPTION_KEY="your-secure-random-string-here"`
- **Recommendation**: Generate a strong random string (at least 32 characters)

### JWT Secrets

**JWT_ACCESS_SECRET**
- **Description**: Secret key used to sign access tokens
- **Required**: Highly recommended for production
- **Default**: `access-secret-key-change-in-production`
- **Format**: String
- **Example**: `export JWT_ACCESS_SECRET="your-access-secret-here"`
- **Recommendation**: Use a strong random string (at least 32 characters)

**JWT_REFRESH_SECRET**
- **Description**: Secret key used to sign refresh tokens
- **Required**: Highly recommended for production
- **Default**: `refresh-secret-key-change-in-production`
- **Format**: String
- **Example**: `export JWT_REFRESH_SECRET="your-refresh-secret-here"`
- **Recommendation**: Use a strong random string different from JWT_ACCESS_SECRET

### JWT Token Configuration

**JWT_ACCESS_TTL**
- **Description**: Time-to-live for access tokens
- **Required**: No
- **Default**: `15m` (15 minutes)
- **Format**: Duration string (e.g., `15m`, `1h`, `30s`)
- **Example**: `export JWT_ACCESS_TTL="30m"`

**JWT_REFRESH_TTL**
- **Description**: Time-to-live for refresh tokens
- **Required**: No
- **Default**: `720h` (30 days)
- **Format**: Duration string (e.g., `720h`, `168h` for 7 days)
- **Example**: `export JWT_REFRESH_TTL="168h"`

**JWT_ISSUER**
- **Description**: JWT issuer claim
- **Required**: No
- **Default**: `go-webui`
- **Format**: String
- **Example**: `export JWT_ISSUER="orbitdeploy"`

**JWT_AUDIENCE**
- **Description**: JWT audience claim
- **Required**: No
- **Default**: `go-webui-users`
- **Format**: String
- **Example**: `export JWT_AUDIENCE="orbitdeploy-users"`

## Application Configuration

**SERVER_ADDR**
- **Description**: Server listening address and port
- **Required**: No
- **Default**: `:8285`
- **Format**: `host:port` or `:port`
- **Example**: `export SERVER_ADDR="0.0.0.0:8080"`

**DB_PATH**
- **Description**: Path to SQLite database file
- **Required**: No
- **Default**: `orbit_app.db`
- **Format**: File path
- **Example**: `export DB_PATH="/var/lib/orbitdeploy/orbit.db"`

**LOG_LEVEL**
- **Description**: Logging level
- **Required**: No
- **Default**: `info`
- **Format**: One of: `debug`, `info`, `warn`, `error`
- **Example**: `export LOG_LEVEL="debug"`

**WEBHOOK_URL**
- **Description**: URL for webhook notifications
- **Required**: No
- **Default**: Empty (disabled)
- **Format**: URL
- **Example**: `export WEBHOOK_URL="https://hooks.example.com/notify"`

**WEBHOOK_TOKEN**
- **Description**: Authentication token for webhook requests
- **Required**: No (but recommended if WEBHOOK_URL is set)
- **Default**: Empty
- **Format**: String
- **Example**: `export WEBHOOK_TOKEN="your-webhook-token"`

## Production Deployment Example

```bash
# Create a secure configuration file (keep this file secure!)
cat > /etc/orbitdeploy/env << 'EOF'
# Security
export ORBIT_ENCRYPTION_KEY="$(openssl rand -base64 32)"
export JWT_ACCESS_SECRET="$(openssl rand -base64 32)"
export JWT_REFRESH_SECRET="$(openssl rand -base64 32)"

# Application
export SERVER_ADDR=":8285"
export DB_PATH="/var/lib/orbitdeploy/orbit.db"
export LOG_LEVEL="info"
EOF

# Source the configuration before starting the service
source /etc/orbitdeploy/env
./orbitdeploy
```

## Security Best Practices

1. **Never commit secrets to version control**: Always use environment variables for sensitive configuration
2. **Use strong random keys**: Generate keys using cryptographically secure random generators
3. **Rotate secrets regularly**: Update encryption keys and JWT secrets periodically
4. **Restrict file permissions**: Ensure configuration files containing secrets have appropriate permissions (e.g., 600)
5. **Use different secrets for different environments**: Never reuse production secrets in development or testing
6. **Store secrets securely**: Consider using secret management tools like HashiCorp Vault, AWS Secrets Manager, etc.

## Testing

For testing purposes, the default values are acceptable. However, you can override them:

```bash
export ORBIT_ENCRYPTION_KEY="test-encryption-key-for-testing"
export JWT_ACCESS_SECRET="test-jwt-access-secret"
export JWT_REFRESH_SECRET="test-jwt-refresh-secret"
go test ./...
```

## Troubleshooting

### "Failed to decrypt" errors
- Check that the `ORBIT_ENCRYPTION_KEY` environment variable is set correctly
- Ensure the same encryption key is used for both encryption and decryption
- If you changed the encryption key, previously encrypted data will not be decryptable

### "Invalid token" errors
- Verify that `JWT_ACCESS_SECRET` and `JWT_REFRESH_SECRET` are set correctly
- Ensure the secrets haven't changed since the token was generated
- Check that tokens haven't expired (see TTL settings)

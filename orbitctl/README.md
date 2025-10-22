# OrbitDeploy CLI (orbitctl) - 最小可用原型

## API Base 与 URL 构造（重要变更）
- ORBIT_API_BASE 支持两种形式：带 `/api` 或不带 `/api`，内部会自动规范化为不带 `/api` 的服务端基础地址（如 `http://localhost:8285`）。
- CLI 内部统一使用集中式 URL 构造：
  - `serverBase()`：返回不带 `/api` 的基础地址
  - `apiRoot()`：返回 `serverBase() + "/api"`
  - `apiURLf("path/%s", arg)`：拼接为 `apiRoot() + "/" + fmt.Sprintf(...)`
- 这样可以避免各处手动拼接 URL，确保行为一致、易于维护。

本目录包含一个最小可用的 Go CLI 原型（可独立为子仓库）。

功能（MVP）
- auth login/logout/refresh：对接 /login、/logout、/refresh_token（参考 doc/PRD_AUTH_TOKENS_JWT.md）
- spec-validate：解析并校验 orbitdeploy.toml（参考 doc/ORBITDEPLOY_SPEC_TOML.md）

注意
- 这是一个独立 Go 模块。后续可将整个 `orbitctl/` 目录移动至上一级目录并初始化为单独仓库。
- API Base 默认从环境变量 `ORBIT_API_BASE` 读取，未设置时默认 `http://localhost:8285`。支持传入带或不带 `/api` 的形式，内部会自动规范化为不带 `/api` 的服务端基础地址，并在实际请求时统一加上 `/api` 前缀。
- Access Token 存储在 `~/.orbitdeploy/tokens.json`，权限建议 0600；Refresh Token 可选存放于 `~/.orbitdeploy/refresh_token`。

使用示例
```bash
# 构建
cd orbitctl && go build ./cmd/orbitctl

# 登录（交互式输入用户名密码）
./orbitctl auth login -u admin -p 123456 --api-base http://localhost:8285

# 刷新（手动）
./orbitctl auth refresh

# 登出
./orbitctl auth logout

# 校验 Spec（默认 orbitdeploy.toml）
./orbitctl spec-validate -f ./orbitdeploy.toml
```

Roadmap
- v0.1：auth + spec-validate + 401 自动刷新
- v0.2：image push/deploy 预留实现（registry / ssh）
- v0.3：deploy logs（SSE 跟随）

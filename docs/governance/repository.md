# StuTrade Engine 工程边界

本文只保存 `stu-trade-engine` 特有的架构、协议、权威来源、验证命令和交付边界，不增加代码排版、命名、注释形式或其他代码风格约束。

## 验证入口

- 完整本地验证依次运行 `go vet ./...`、`go test ./...`、`go test -race ./...` 和 `go mod tidy -diff`。

## 文档所有权

- 根 README 负责当前组件目的、能力、边界、仓库结构、开发命令和文档索引。
- `docs/project-description.md` 负责当前 engine 模型、用例、协议和组件所有权。
- `docs/implementation_plan.md` 只记录尚未完成的目标和实施顺序，不证明当前行为。
- 父仓库的 `docs/project-description.md` 负责产品范围、用户可见交易语义、部署拓扑和跨组件边界。

## 权威来源

- 导出的 Go 代码和行为测试负责通用 engine 模型、入口、逻辑持久化协议、恢复规则、错误和状态机行为。
- Web 仓库的 `openapi/stu-trade.yaml` 负责 `/api/v1` 路径、传输 DTO、状态码和错误结构；父仓库负责运行时实现以及到 engine 操作的转换。
- 父仓库的 migration 负责 PostgreSQL 物理 schema、约束和迁移顺序。
- 父仓库的 query 文件负责 sqlc 查询文本和生成的数据库绑定；这些绑定留在实现 engine 协议或 host 持久化的父仓库 adapter 中。
- `go.mod` 负责 module path、Go 版本和实际依赖。

## 组件边界

- `github.com/Jason-skd/stu-trade-engine` 是 Go 1.27.1 library，公共 API 保持在根 package，生产 module 不使用非标准库依赖。
- Engine 保持同步、确定性，不负责 HTTP、物理数据库访问、进程配置、日志、后台 goroutine、host 领域类型、可执行产物、部署或产品发布。
- Engine 通过最小能力协议表达逻辑持久化和 host 能力；具体 adapter、依赖注入、生命周期与基础设施归调用方所有。
- Engine 自身的逻辑状态和恢复语义由本仓库协议负责，物理存储结构由父仓库负责。

## 测试边界

- 行为测试与实现并置，覆盖公共 API、领域不变量、状态转换、协议调用和恢复语义。
- 完整验证使用 race detector；本仓库不以 host、数据库或网络服务作为测试前置条件。

## 交付边界

- GitHub Actions 验证 pull request 和 `main` revision。
- 父仓库以 gitlink 记录精确 engine commit；engine revision 不会在 gitlink 更新并完成集成验证前进入产品集成。
- 本仓库只交付 library revision，不拥有可执行制品、环境部署或产品 release。

# StuTrade Engine

StuTrade Engine 是通用撮合 Go module。它管理挂单、价格兼容、撮合请求状态及这些通用状态的逻辑存取；业务特有数据与规则、物理数据库适配、HTTP、日志和成交后的业务流程由宿主负责。

## 当前能力

- `Amount` 表示以人民币分计价的正整数金额。
- `Listing` 管理买卖方向以及 open、reserved、matched、withdrawn 状态转换。
- `MatchRequest` 管理自动撮合和直接成交的确认与终态转换。
- `RestoreMatchRequest` 校验撮合请求的持久化表示及来源关联。
- `PriceCompatible` 判断卖方报价是否不高于买方出价。
- `PublishListing` 通过调用方定义的 `ListingCreator` 创建并保存合法的 open Listing。
- `RestoreListing` 校验宿主持久化表示，`ReadListing` 和 `ReadListings` 按请求顺序恢复通用状态。
- `WithdrawListing` 通过批量锁 Protocol 在宿主事务内幂等撤回属于调用者的 open Listing。
- `MatchValidator` 用 `(bool, error)` 区分校验通过、正常业务不匹配和依赖故障。
- `PurchaseSellListing` 和 `AcceptBuyListing` 通过宿主 `WithTx` 工作单元原子完成锁定、复查、撮合请求、订单和挂单更新。
- 领域错误可通过 `errors.Is` 和 `errors.As` 稳定识别。

候选发现、自动撮合确认、拒绝释放和请求查询仍在实施计划中。主仓库固定本仓库的精确 commit；当前产品语义以主仓库文档为准。

## 边界

依赖方向始终为 `host -> engine`。engine 保持同步、确定性，不读取环境变量、不直接连接数据库、不启动后台 goroutine，也不依赖 HTTP、ORM、日志库或宿主类型。engine 的用例定义当前调用方所需的小型存取 interface，并通过它恢复和持久化通用实体；宿主只装配物理适配器，不在业务代码中复制这些状态操作。

当前生产代码只使用 Go 标准库，根 package 按责任拆文件：

```text
engine.go          发布挂单用例
listing_usecases.go 读取与撤回挂单用例
direct_trade.go    直接购买与接单用例
listing.go         Listing 与 Amount
match_request.go   MatchRequest
matching.go        价格规则
protocol.go        当前最小能力接口
errors.go          稳定领域错误
*_test.go          并置行为测试
```

详细模型和边界见[组件说明](docs/project-description.md)，后续能力见[实施方案](docs/implementation_plan.md)。

## 开发

```bash
test -z "$(gofmt -l .)"
find .agents/skills -type f -name '*.sh' -print0 | xargs -0 bash -n
go vet ./...
go test ./...
go test -race ./...
go mod tidy -diff
git diff --check
```

## 文档索引

- [组件说明](docs/project-description.md)
- [实施方案](docs/implementation_plan.md)
- [工作流治理](docs/governance/workflow.md)
- [工程治理](docs/governance/engineering.md)
- [验证治理](docs/governance/verification.md)
- [治理元规则](docs/governance/meta.md)
- [交付治理](docs/governance/delivery.md)
- [仓库治理补充](docs/governance/repository.md)

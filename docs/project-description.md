# StuTrade Engine 组件说明

## 组件职责

StuTrade Engine 是由业务宿主嵌入的同步 Go library。它拥有跨业务场景可复用的挂单和撮合请求模型，以及这些通用状态的命令、查询、恢复和逻辑持久化契约。宿主拥有用户、资源、订单、业务特有的挂单资料、领域规则与物理数据库适配。

engine 不提供通用规则语言，也不感知 ISBN、图书、Gin、PostgreSQL 或前端。跨边界只保存宿主生成的标识和不透明资源引用，不提供任意 `Constants` 或 `Variables` 扩展容器。

宿主通过 engine 声明的细粒度 Protocol 提供 PostgreSQL 等基础设施能力，但宿主业务代码不直接构造、恢复或修改 Listing、MatchRequest。宿主以 Listing ID 或资源引用保存二手书等业务资料，并负责将 engine 返回的通用状态与这些资料组装为业务对象。物理表可以位于同一个数据库，并由外键和共享工作单元保证两部分完整，但这不会改变各自的逻辑所有权。

## Listing

Listing 包含 ID、所有者 ID、资源引用、买卖方向、价格、状态和创建/更新时间。状态转换为：

```text
open -----> reserved -----> matched
  |             |
  |             +--------> open
  |
  +----------------------> withdrawn
```

金额必须是以最小货币单位表示的正整数。时间不能早于实体先前的更新时间；失败的转换不修改实体。

## MatchRequest

MatchRequest 记录买卖双方 Listing、成交价、来源、双方确认、状态、宿主订单引用和时间。状态转换为：

```text
pending -----> completed
   |
   +---------> declined
   |
   +---------> cancelled
```

自动撮合从 pending 开始并分别记录买卖双方确认。直接接单或购买可以创建已包含相应确认的请求，并在同一业务工作单元中直接完成。终态不能再次转换。

## 当前用例与协议

`PublishListing` 校验输入、构造 open Listing，再调用：

```go
type ListingCreator interface {
	CreateListing(context.Context, *Listing) error
}
```

interface 由 engine 的消费方定义，只包含当前用例需要的能力。engine 在输入无效时不会调用依赖，并保留持久化错误的原因。

## 所有权

本仓库的导出 Go 代码和行为测试是 engine 运行契约的权威来源。产品成交语义和跨组件边界由 StuTrade 主仓库负责；后续尚未实现的 engine 用例记录在[实施方案](implementation_plan.md)，不构成当前 API。

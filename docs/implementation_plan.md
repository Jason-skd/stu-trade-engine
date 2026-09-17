# StuTrade Engine 实施方案

本文记录 engine 尚未实现的能力。现有导出 API 以代码和测试为准，完整产品语义以 StuTrade 主仓库文档为准。

## 下一阶段能力

从主仓库三条交易纵切的真实调用方提取以下最小能力：

1. 撤回 reserved Listing，并取消关联 pending MatchRequest、释放有效的对手 Listing。
2. 为指定 Listing 同步发现和排序候选。
3. 在共享工作单元中锁定 Listing、重做规则检查并创建唯一 MatchRequest。
4. 直接接受需求或购买供给，使用当前存在的目标 Listing 创建 completed MatchRequest 和宿主订单，不制造虚假反向 Listing。
5. 幂等确认或拒绝 pending MatchRequest，并正确释放或完成 Listing。
6. 查询宿主需要展示的 MatchRequest，但不承载领域搜索读模型。

## 目标协作边界

engine 是通用 Listing、MatchRequest 的唯一领域访问入口，并按用例消费细粒度基础设施与宿主能力，而不是暴露一个通用 repository：

- engine 定义通用状态所需的创建、读取、锁定和保存 Protocol，并拥有从持久化表示恢复实体的合法性规则。
- 宿主实现物理数据库适配；宿主业务代码不直接构造、恢复或修改 engine 实体。
- 宿主按 Listing ID 或资源引用保存 ISBN、成色等业务资料，并组装 engine 状态、业务资料和 HTTP DTO，不复制价格、方向、状态或确认等通用事实。
- 业务搜索可以先返回候选 Listing ID，engine 再批量取得通用状态；候选发现只缩小范围，不代替事务内最终检查。
- 宿主规则负责 ISBN、自买限制和资源有效性。
- 共享工作单元负责稳定加锁，并原子持久化 engine 状态、宿主业务扩展和订单。
- 订单创建由宿主提供，并以 MatchRequest 保证唯一。
- engine 错误保持稳定，由宿主翻译为 HTTP 或数据库语义。
- 当前用例没有任意宿主元数据读写需求，不引入 `Constants`、`Variables` 或空的对称扩展表。

接口签名在首个真实宿主调用方落地时确定，不提前公开 `CandidateFinder`、`MatchPolicy`、`UnitOfWork` 或 `TradeCreator` 的假定形态。

## 实施顺序

1. 从主动购买纵切提取规则和订单创建能力，并调整直接成交模型以只引用真实存在的 Listing。
2. 复用相同核心实现卖方接单，避免 mode 驱动的平行流程。
3. 加入候选排序和自动撮合，创建 pending MatchRequest。
4. 加入确认、拒绝和 reserved Listing 撤回协作。
5. 根据宿主组装和查询入口补充 MatchRequest 的最小读取能力。
6. 通过主仓库真实 PostgreSQL 适配和跨仓库测试验证协议。

## 验证目标

- 价格等于边界时兼容，超出边界时拒绝。
- 相同价格按创建时间和 ID 稳定排序。
- engine 存取往返保持通用实体事实，宿主扩展不复制或改写这些事实。
- 宿主规则拒绝或资源失效时不留下预留状态。
- 重复确认和重复直接成交不重复创建订单。
- 两个请求争抢同一 Listing 时最多一个成功。
- 确认、拒绝和撤回竞争时只产生一个合法终态。
- 所有失败保留原因，且失败转换不修改已有实体。

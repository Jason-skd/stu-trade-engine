---
name: resolve-review
description: 针对具体 revision 验证并在获得授权时解决用户提供的 code review finding。不用于独立全面 review、普通 bug 报告或功能设计。
---

# 处理 Review Finding

将每条 comment 视为需要验证的假设。遵循 `AGENTS.md` 和适用工程约束；评估本身不授权修复。

将 comment 锚定到具体 revision、branch、路径和符号，然后检查：

- 当前实现。
- 调用方。
- 契约。
- 测试。

将每条 finding 分类为：

- 当前缺陷。
- 已经解决。
- 误报。
- 需要产品或所有权决策。

使用具体证据进行判断，不为了满足 comment 的文字表述而修改代码。

获得修复授权后：

1. 按根因合并相关的当前缺陷。
2. 行为变化时增加行为或回归覆盖。
3. 先运行最近的检查，再扩大验证范围。
4. 保持 diff 聚焦，不进行无关清理。

报告首先列出按严重程度排序的未解决 finding，并提供文件证据；然后说明已修复项、拒绝或已过时的 comment 及原因，以及验证证据。

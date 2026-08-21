Bug：策略创建命令和 HTTP DTO 复用了调用方的环境、域名切片，后续修改原切片会污染策略约束。

触发：分别执行 collection.json 中的两条定向测试；基线会在调用方修改切片后观察到策略或命令内容被改写。

错误信息：测试报告 policy retained caller-owned slice 或 command retained request-owned slice。

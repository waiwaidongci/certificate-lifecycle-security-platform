# Bug 复现记录

- Bug：分发 webhook 的响应体生命周期跨层传递时断开，成功、HTTP 错误和读取错误路径都可能绕过统一关闭逻辑。
- 触发：让 webhook transport 返回 200、502，或返回读取错误的 `ReadCloser`，分别调用发送流程。
- 错误信息：定向检查报告 `Send() left the response body open`；读取错误场景返回 `Send() error = <nil>, want wrapped response read error`。

# Remote Agent -> APP Local Workspace Test Matrix

## Scope

This matrix tracks the current project status for the "APP workspace first, remote agent returns artifacts back to the local thread workspace" model.

- Bridge RPC surface: `session.start`, `session.message`
- APP local-first result handling: write returned `artifacts` into the current local thread workspace
- Remote execution metadata: keep `remoteWorkingDirectory` / `remoteWorkspaceRefKind` as thread metadata only

## Matrix

| Case | 输入类型 | 目标能力 | 当前项目标记 | 对应接口 | 关键断言 | 主要代码路径 | 当前测试落点 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `pptx` | 文稿生成 + 二次追问 | 单 agent + 追问续写 | 已实现，待补专项 E2E 断言 | `session.start`, `session.message` | 首轮生成 `.pptx`；同线程追问继续修改；APP 线程 `workspaceBinding` 保持 `localFs`；远程目录只记录到 metadata | Bridge: `internal/acp/server.go`, `internal/acp/execution.go`；APP: `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/lib/app/app_controller_desktop_single_agent_go_task_flow.dart`, `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/lib/runtime/go_task_service_client.dart`, `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/lib/runtime/runtime_models_runtime_payloads.dart` | APP runtime: `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/test/runtime/app_controller_ai_gateway_chat_suite_single_agent.dart`；Bridge E2E: `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/test/runtime/bridge_real_e2e_suite.dart` |
| `docx` | 周报生成 | 文档产出 | 已实现，待补内容结构专项断言 | `session.start` | 生成 `.docx`；结果写回当前线程本地目录；同名文件版本化为 `.v2`；内容结构可进一步做文档级校验 | Bridge: `internal/acp/server.go`, `internal/acp/execution.go`；APP: `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/lib/app/app_controller_desktop_single_agent_go_task_flow.dart` | APP runtime: `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/test/runtime/app_controller_ai_gateway_chat_suite_single_agent.dart` |
| `xlsx` | 公式表格 | 表格计算 | 已实现，待补公式专项 E2E | `session.start`, `session.message` | 生成 `.xlsx`；写回当前线程；后续追问仍走同一本地 workspace；公式存在且可计算 | Bridge: `internal/acp/server.go`, `internal/acp/execution.go`；APP: `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/lib/app/app_controller_desktop_single_agent_go_task_flow.dart`, `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/lib/runtime/go_task_service_client.dart` | APP runtime: `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/test/runtime/app_controller_ai_gateway_chat_suite_single_agent.dart`；Bridge E2E: `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/test/runtime/bridge_real_e2e_suite.dart` |
| `pdf` | 合并 / 转换 | 文件变换 | 已实现，待补页数/内容专项校验 | `session.start` | 产出 `.pdf`；页数或文本内容符合预期；结果写回当前线程本地目录；线程 workspace 不切到远程目录 | Bridge: `internal/acp/server.go`, `internal/acp/execution.go`；APP: `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/lib/app/app_controller_desktop_single_agent_go_task_flow.dart` | Bridge E2E: `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/test/runtime/bridge_real_e2e_suite.dart` |
| `image-resizer` | 图片处理 | 图片回写 | 已实现，待补尺寸断言自动化 | `session.start` | 输出图片写回当前线程；尺寸符合请求；同线程 artifact 面板可见；workspace 不变 | Bridge: `internal/acp/server.go`, `internal/acp/execution.go`；APP: `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/lib/app/app_controller_desktop_single_agent_go_task_flow.dart`, `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/lib/runtime/desktop_thread_artifact_service.dart` | 规划/E2E: `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/test/runtime/bridge_real_e2e_suite.dart`；artifact surface: `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/test/runtime/desktop_thread_artifact_service_test.dart` |
| `browser` | 浏览器自动化 | 浏览 / 摘要 / 截图 / 日志 | Bridge 结果面已支持，真实 browser skill 链路待专项收口 | `session.start`, `session.message` | 文本摘要、截图、日志至少两类 artifact/summary 回写本地线程；继续追问复用本地 workspace，并可附带上次远程目录 hint | Bridge: `internal/acp/server.go`, `internal/acp/execution.go`；APP: `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/lib/app/app_controller_desktop_single_agent_go_task_flow.dart`, `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/lib/runtime/go_task_service_client.dart`, `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/lib/runtime/desktop_thread_artifact_service.dart` | Bridge E2E: `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/test/runtime/bridge_real_e2e_suite.dart`；APP runtime: `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/test/runtime/app_controller_ai_gateway_chat_suite_single_agent.dart` |

## Shared Interfaces

| Interface | 用途 | 代码路径 |
| --- | --- | --- |
| `session.start` | 首轮执行；生成初始结果与 artifact payload | `internal/acp/server.go` |
| `session.message` | 同线程 follow-up；复用本地 workspace，并允许携带远程目录 hint | `internal/acp/server.go` |
| `artifacts[]` | 结构化产物回传；APP 侧据此落盘 | `internal/acp/execution.go`, `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/lib/runtime/go_task_service_client.dart` |
| `remoteWorkingDirectory` | 远程执行目录元数据，不替代 APP 本地 workspace | `internal/acp/execution.go`, `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/lib/runtime/runtime_models_runtime_payloads.dart` |
| `remoteWorkspaceRefKind` | 远程目录类型元数据，不替代 APP 本地 workspace | `internal/acp/execution.go`, `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/lib/runtime/runtime_models_runtime_payloads.dart` |
| `remoteWorkingDirectoryHint` | APP follow-up 请求时附带上一次远程目录提示 | `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/lib/runtime/go_task_service_client.dart`, `/Users/shenlan/workspaces/cloud-neutral-toolkit/xworkmate-app/lib/app/app_controller_desktop_single_agent_go_task_flow.dart` |

## Current Status Summary

| Area | 状态 | 说明 |
| --- | --- | --- |
| Bridge 结构化结果面 | 已实现 | 成功结果会补 `artifacts`, `resultSummary`, `remoteWorkingDirectory`, `remoteWorkspaceRefKind` |
| APP 本地落盘 | 已实现 | inline artifact 会写回当前线程本地目录，并自动版本化同名文件 |
| 线程 metadata 分离 | 已实现 | 远程目录写入 thread context metadata，不再覆盖 `workspaceBinding` |
| 单元 / runtime 回归 | 已覆盖 | 已有 bridge 和 APP runtime 测试覆盖核心本地优先语义 |
| 六类真实场景专项验收 | 待继续补全 | 还需要把 `pptx/docx/xlsx/pdf/image-resizer/browser` 的真实 E2E 断言逐项补细 |

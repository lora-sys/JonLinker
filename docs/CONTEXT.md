# JobLinker

JobLinker 是一个 AI 招聘助手系统，通过多阶段自主 Agent 帮助候选人完成求职全流程。

## Language

**Search Agent**:
搜索职位的 Eino ReAct Agent。接收用户 Query，调用 `query_jobs` 工具，返回排名结果。
_Avoid_: Job Agent, Main Agent

**Resume Agent**:
简历聊天的 Eino ChatModel Agent。通过多轮对话 + PDF 解析补全候选人个人资料。
_Avoid_: Chat Agent, Profile Agent

**CandidateProfile**:
候选人的完整求职资料，包括技能、工作经历、教育背景、兴趣爱好等，由 Resume Agent 聊天生成。
_Avoid_: Resume, CV, 简历

**Application**:
针对特定职位的申请包，包含定制求职信 + 针对性简历，由 Search Agent 的 `apply_job` 工具生成。
_Avoid_: Apply request, 投递

**SearchIsland**:
前端搜索职位的客户端岛屿组件。聊天式界面，支持搜索和申请一体化对话。
_Avoid_: Search page, Search form

**ResumeChatIsland**:
前端简历聊天的客户端岛屿组件。
_Avoid_: Chat page, Profile form

**SSE Streaming**:
Go 后端通过 Server-Sent Events 向前端推送 Agent 对话流的通信方式。
_Avoid_: WebSocket, polling

**Application**:
针对特定职位的申请包，包含 CoverLetter + 定制简历 + 匹配亮点。
_Avoid_: Apply request, 投递

**CheckpointStore**:
Eino `compose.CheckPointStore` 接口，用于在 Resume Agent 和 Search Agent 之间共享 CandidateProfile。
_Avoid_: Database, session store

**CoverLetter**:
由 `apply_job` 工具为特定职位生成的定制求职信（Markdown 格式）。
_Avoid_: 自我介绍, 申请信

**Islands Architecture**:
前端架构模式，交互性组件作为客户端岛屿嵌入服务端渲染页面。
_Avoid_: SPA, CSR-only

## Relationships

- **Search Agent** 拥有 `query_jobs` 工具
- **Search Agent** 拥有 `apply_job` 工具（从 CheckpointStore 读 CandidateProfile + 接收 Job，生成 Application）
- **Resume Agent** 拥有 `parse_resume_pdf` 工具（通过 Firecrawl CLI 解析 PDF）
- **Resume Agent** 通过 SSE Streaming 与 ResumeChatIsland 对话
- **Resume Agent** 将 CandidateProfile 写入 CheckpointStore
- **Search Agent** 从 CheckpointStore 读取 CandidateProfile
- 一个 **CandidateProfile** 可为多个 **Application** 提供基础数据
- 一个 **Application** 对应一个 **Job**

## Example dialogue

> **Dev:** "用户搜索到职位后，点生成申请，Search Agent 怎么拿到候选人的资料？"
> **Domain expert:** "CheckpointStore。Resume Agent 聊天过程中已经把资料写进去了，Search Agent 的 apply_job 工具直接读就行。"

## Flagged ambiguities

- "简历" 可能指 **上传的 PDF 文件** 或 **CandidateProfile** — 已区分：PDF 是原始输入，CandidateProfile 是结构化资料。
- "投递" 可能指 **生成申请内容** 或 **实际提交到招聘平台** — Phase 2 只做前者。

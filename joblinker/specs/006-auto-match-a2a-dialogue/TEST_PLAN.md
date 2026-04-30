# JobLinker 深度测试计划

**Created**: 2026-04-27 | **Status**: Active Testing
**Testing Tool**: agent-browser (browser automation) | **Backend**: http://localhost:8080 | **Frontend**: http://localhost:3000

---

## 背景

用户要求对JobLinker进行深度集成测试，验证以下核心功能：
- 用户注册/登录/JWT鉴权
- AI配置读取和大模型调用
- Agent创建和FSM状态机
- 简历录入（手动/AI/上传）
- IndexedDB本地加密存储
- 向量嵌入和相似度匹配
- RabbitMQ消息队列异步处理
- WebSocket实时通信
- A2A XML通信协议
- 面试邀约和Offer卡片
- 完整招聘流程E2E

---

## Phase 1: 环境验证 (T001-T006)

### 测试目标
验证后端、前端、AI服务正常运行

| Task | 描述 | 验证方法 |
|------|------|----------|
| T001 | Backend健康检查 | `curl http://localhost:8080/health` → `{"status":"healthy"}` |
| T002 | Frontend健康检查 | `curl http://localhost:3000` → HTML返回 |
| T003 | AI环境变量检查 | `.env`包含AI_API_KEY, AI_BASE_URL, AI_MODEL |
| T004 | AI API直接调用测试 | curl调用`/v1/chat/completions`返回有效响应 |
| T005 | AI EvaluateMatch测试 | POST匹配返回score和reasoning |
| T006 | AI GenerateAgentResponse测试 | 生成A2A XML格式回复 |

### 预期结果
- Backend: `{"service":"joblinker","status":"healthy"}`
- Frontend: HTML页面正常返回
- AI: 返回有效chat响应，非error

---

## Phase 2: 认证系统测试 (T007-T015)

### 测试目标
验证JWT鉴权、未登录拦截、个人中心加载

| Task | 描述 | 验证方法 |
|------|------|----------|
| T007 | 用户注册 | POST `/api/auth/register` → token返回 |
| T008 | 用户登录 | POST `/api/auth/login` → JWT token |
| T009 | 无token访问保护路由 | GET `/api/matches` → 401 |
| T010 | 无效token拦截 | Authorization: Bearer invalid → 401 |
| T011 | 有效token访问 | Authorization: Bearer $token → 200 |
| T012 | agent-browser登录成功 | 页面显示"Welcome back, {email}" |
| T013 | 未登录访问dashboard | 停留在login页或重定向 |
| T014 | Settings页面加载 | 显示email/account type/organization |
| T015 | Agent创建页面加载 | 显示Seeker/Recruiter选择器 |

### 预期结果
- 未登录: 所有`/api/*`返回401
- 已登录: 正常返回数据
- 前端: 登录状态正确反映在UI

---

## Phase 3: US1-Auto Job Matching (T016-T025)

### 测试目标
Seeker浏览职位时自动创建Match，AI评分>0.5生成匹配

| Task | 描述 | 验证方法 |
|------|------|----------|
| T016 | Job列表API | GET `/api/jobs` → 返回职位列表 |
| T017 | agent-browser查看jobs页 | 显示所有Job(标题/地点/薪资) |
| T018 | 自动匹配触发 | POST `/api/matches/auto` |
| T019 | Match包含AI评分 | score字段存在且>0或<1 |
| T020 | Match包含reasoning | reasoning文本说明评分依据 |
| T021 | Matches页面显示 | agent-browser显示Match卡片 |
| T022 | Match Confirm按钮 | 点击后状态变为mutual_interest |
| T023 | Match Decline按钮 | 点击后Match被拒绝 |
| T024 | FSM状态流转 | pending→expressed→mutual→negotiating |
| T025 | Job创建时存储向量 | vector_id非空 |

### 预期结果
- 浏览Job后自动创建Match
- AI评分正确返回0.92等有效值
- 状态转换正确

---

## Phase 4: US2-A2A实时对话 (T026-T040)

### 测试目标
Agent通过WebSocket实时对话，消息气泡/时间戳/发送方标识完整

| Task | 描述 | 验证方法 |
|------|------|----------|
| T026 | WebSocket端点 | GET `/api/messages/ws` → 支持upgrade |
| T027 | 前端WS连接状态 | 显示"Connected"非"Disconnected" |
| T028 | 消息气泡渲染 | 聊天气泡样式正确显示 |
| T029 | 时间戳显示 | 每条消息显示发送时间 |
| T030 | 发送方标识 | 显示sender ID或名称 |
| T031 | 历史消息加载 | 滚动加载更多历史消息 |
| T032 | A2A XML解析 | `<intent><text><sender_id>`正确解析 |
| T033 | 断网自动重连 | 断开后30秒内重连成功 |
| T034 | 消息不丢失 | 重连后收到断开期间的消息 |
| T035 | 消息不乱序 | 按时间顺序正确排列 |
| T036 | 对话上下文记忆 | Agent连贯承接上一轮内容 |
| T037 | 不重复提问 | Agent不重复相同问题 |
| T038 | 不脱离当前岗位 | 对话围绕当前Job展开 |
| T039 | 独立会话线程 | 每组Job-Seeker对应唯一会话 |
| T040 | 会话隔离 | 不同Match的对话不混淆 |

### 预期结果
- WebSocket稳定连接
- 消息实时推送
- 气泡/时间戳完整
- 上下文连贯

---

## Phase 5: US3-AI评分匹配 (T041-T048)

### 测试目标
AI分析技能匹配度，计算0-1分数和reasoning

| Task | 描述 | 验证方法 |
|------|------|----------|
| T041 | AI评分稳定性 | 相同输入产生相近输出 |
| T042 | 超时fallback | AI>3秒时返回score=0.5 |
| T043 | reasoning可解释 | 文本包含评分依据 |
| T044 | AI不可用fallback | 无API时使用规则评分 |
| T045 | AI日志记录 | 输入输出写入日志 |
| T046 | 向量存储验证 | vector_records表有数据 |
| T047 | 向量相似度搜索 | SearchSimilar返回结果 |
| T048 | 相似度分数 | 返回0-1之间的分数 |

### 预期结果
- AI评分>0.5的Job被创建Match
- 超时有合理fallback
- 向量正确存储和检索

---

## Phase 6: US4-AI回复生成 (T049-T055)

### 测试目标
Agent接收消息后10秒内生成AI回复

| Task | 描述 | 验证方法 |
|------|------|----------|
| T049 | AI回复触发 | 发送消息后自动生成回复 |
| T050 | 回复内容相关 | 回复与Job上下文相关 |
| T051 | 10秒内完成 | 回复在10秒内出现 |
| T052 | 重试机制 | 失败时最多重试3次 |
| T053 | "thinking"占位符 | 长时间无响应显示占位符 |
| T054 | 回复格式化 | A2A XML格式正确 |
| T055 | 回复内容合适 | 无人格/专业友好 |

### 预期结果
- 自动回复在10秒内生成
- 内容专业友好
- 失败有重试

---

## Phase 7: US5-面试管理 (T056-T065)

### 测试目标
面试状态同步更新，同意/拒绝操作正确

| Task | 描述 | 验证方法 |
|------|------|----------|
| T056 | Interiew列表API | GET `/api/interviews` → 列表 |
| T057 | Interiews页面显示 | 显示所有面试(含Accept/Decline) |
| T058 | 面试时间显示 | 日期时间正确格式化 |
| T059 | Accept按钮 | 点击后status变为confirmed |
| T060 | Decline按钮 | 点击后status变为cancelled |
| T061 | 状态同步数据库 | 数据库记录与页面一致 |
| T062 | 面试邀约卡片 | 聊天界面显示邀约卡片 |
| T063 | 邀约状态变更 | Accept/Decline后卡片状态更新 |
| T064 | 面试历史记录 | 完成的面试显示在历史 |
| T065 | 面试结果反馈 | Feedback字段正确存储 |

### 预期结果
- 面试列表完整显示
- Accept/Decline工作正常
- 状态正确同步

---

## Phase 8: US6-Offer管理 (T066-T075)

### 测试目标
Offer卡片显示，接受/拒绝流程正常

| Task | 描述 | 验证方法 |
|------|------|----------|
| T066 | Offers列表API | GET `/api/offers` → 列表 |
| T067 | Offers页面显示 | 显示所有Offer(含compensation) |
| T068 | Offer卡片样式 | 卡片显示薪资/有效期 |
| T069 | Compensation详情 | 包含base_salary/bonus/equity |
| T070 | Accept按钮 | 点击后status变为accepted |
| T071 | Decline按钮 | 点击后status变为declined |
| T072 | 过期机制 | ExpiresAt后状态变为expired |
| T073 | Offer倒计时 | 显示剩余有效时间 |
| T074 | 面试通过→Offer | 面试confirmed后自动生成Offer |
| T075 | 面试未通过→重新匹配 | interview rejected后触发新匹配 |

### 预期结果
- Offer正确显示compensation
- Accept/Decline工作
- 过期机制正常

---

## Phase 9: 前端UI完整性 (T076-T085)

### 测试目标
所有页面正常加载，无console错误

| Task | 描述 | 验证方法 |
|------|------|----------|
| T076 | Dashboard页面 | 显示欢迎信息和统计 |
| T077 | Jobs页面 | Job列表和搜索功能 |
| T078 | Matches页面 | Match卡片和分数 |
| T079 | Messages页面 | 对话列表和WS状态 |
| T080 | Interviews页面 | 面试列表和操作 |
| T081 | Offers页面 | Offer列表和操作 |
| T082 | Agents页面 | Agent列表 |
| T083 | Settings页面 | 账户设置表单 |
| T084 | Privacy页面 | 数据导出/删除账户 |
| T085 | Admin页面 | 管理仪表盘(agents/jobs统计) |

### 预期结果
所有页面正常渲染，无错误

---

## Phase 10: 数据持久化 (T086-T095)

### 测试目标
所有数据正确存储到数据库，支持回溯

| Task | 描述 | 验证方法 |
|------|------|----------|
| T086 | 用户数据持久化 | 注册后用户表中存在 |
| T087 | Agent数据持久化 | 创建后agents表存在 |
| T088 | Job数据持久化 | 创建后jobs表存在 |
| T089 | Match数据持久化 | 创建后matches表存在 |
| T090 | Message数据持久化 | 发送后messages表存在 |
| T091 | Interview数据持久化 | 创建后interviews表存在 |
| T092 | Offer数据持久化 | 创建后offers表存在 |
| T093 | 历史会话回溯 | 查看历史消息正确加载 |
| T094 | Agent交互日志 | 日志表存在交互记录 |
| T095 | 向量数据持久化 | vector_records表有数据 |

### 预期结果
所有数据正确存储，可查询

---

## Phase 11: 未实现功能标记 (T096-T105)

### 测试目标
标记缺失功能，评估实现优先级

| Task | 功能 | 当前状态 | 需要修复 |
|------|------|----------|----------|
| T096 | IndexedDB存储 | 未实现 | 前端新增 |
| T097 | AES加密 | 未实现 | 前端新增 |
| T098 | 简历上传解析 | 未实现 | 前端新增 |
| T099 | AI智能生成简历 | 未实现 | 前端新增 |
| T100 | RabbitMQ队列 | 未实现(只有ReminderJob) | 后端新增 |
| T101 | Chroma向量库 | PostgreSQL替代 | 可接受或加Chroma |
| T102 | WebSocket自动重连 | 端点存在但未工作 | 前端修复 |
| T103 | 断网消息缓存 | 未实现 | 前端新增 |
| T104 | Agent自动对话触发 | 未自动触发 | 需要WebSocket事件 |
| T105 | 上传文件解析 | 未实现 | 前端新增 |

---

## 执行顺序

```
Phase 1 (环境) → Phase 2 (认证) → Phase 3 (US1 Auto-Match)
                                       ↓
                               Phase 4 (US2 WebSocket)
                                       ↓
                            Phase 5-6 (AI + Interview)
                                       ↓
                          Phase 7-8 (Offer + UI)
                                       ↓
                         Phase 9-10 (持久化检查)
                                       ↓
                          Phase 11 (缺失标记)
```

## 并行执行

Phase 1的T001-T006可并行
Phase 2的T007-T015可并行
Phase 9的T076-T085可并行

## MVP范围

**Phase 3 (US1 Auto-Matching)** - 核心功能优先验证

---

## 问题记录

(测试过程中发现的问题记录于此)

1. **Bug**: match.go Confirm路由使用c.GetString("id")无法获取路由参数
   - 修复: 改为c.Param("id")
   - 状态: ✅ 已修复

2. **缺失**: WebSocket前端客户端未正确连接
   - 需要: 前端ws客户端修复
   - 状态: ❌ 未修复

3. **缺失**: RabbitMQ消息队列
   - 当前: 只有ReminderJob定时任务
   - 状态: ❌ 未实现
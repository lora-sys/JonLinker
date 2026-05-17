# E2E 测试方案 — playwright-cli + 后端日志 + 数据库验证

> 基于 10 条决策设计，覆盖所有 Phase 0 变更验证

## 1. 测试架构

```
┌──────────────────────────────────────────────────────────────┐
│                    playwright-cli (浏览器驱动)                  │
│  ┌─────────────────┐  ┌─────────────────┐                    │
│  │  -s=seeker      │  │  -s=recruiter   │                    │
│  │  (Chrome)       │  │  (Chrome)       │                    │
│  └────────┬────────┘  └────────┬────────┘                    │
│           │                    │                              │
│           ▼                    ▼                              │
│  ┌──────────────────────────────────────────────────┐        │
│  │            curl / API 调用 (数据操作)              │        │
│  │  POST /api/auth/login → 获取 token               │        │
│  │  POST /api/matches/:id/confirm → 确认 Match      │        │
│  │  GET /api/matches/:id → 查 DB 状态               │        │
│  └──────────────────────┬───────────────────────────┘        │
│                         │                                     │
│                         ▼                                     │
│  ┌──────────────────────────────────────────────────┐        │
│  │           后端日志监测 (并行 tail -f)              │        │
│  │  grep "FSM state changed" / "agent_response"     │        │
│  │  grep "human_confirm" / "broadcastStateChange"   │        │
│  └──────────────────────────────────────────────────┘        │
└──────────────────────────────────────────────────────────────┘
```

## 2. 前置条件 / Fixtures

### 2.1 启动基础设施

```bash
# 容器
docker compose -f joblinker/docker-compose.yml up -d

# 后端
cd joblinker/backend && go run cmd/seed/main.go  # 导入 JSON fixtures
cd joblinker/backend && go run cmd/server/main.go

# 前端
cd joblinker/frontend && npm run dev
```

### 2.2 Seed 数据可用资源

seed 提供了可直接用的用户/Agent/Job/Match：

| 角色 | 邮箱 | 密码 |
|------|------|------|
| Seeker | seeker.frank@email.com | password123 |
| Recruiter | hr.bob@techcorp.com | password123 |

预创建了若干 Match，其中 `pending` 状态的可以直接用。

### 2.3 Mock LLM 模式 (可选)

在 `.env` 设置 `AI_MOCK_MODE=true` 时，后端 AI client 返回**可预测的 FSM 推进回复**:

```json
// mock 回复序列 (每个轮次推进一个阶段):
Round 1: {"intent": "INTRODUCTION", "message": "介绍职位..."}
Round 2: {"intent": "INTEREST", "message": "表示兴趣..."}
Round 3: {"intent": "NEGOTIATION", "message": "开始谈判..."}
Round 4: {"intent": "OFFER", "message": "发送Offer..."}
Round 5: {"intent": "CONFIRM", "message": "确认接受..."}
```

有 Mock 时测试**确定性强、速度快**。无 Mock 时测试**真实、但慢**。

## 3. P0 核心测试 (必须全绿)

### TC1: 注册 + 登录 (来源: 旧 T1)

```bash
# 准备
rm -f seeker.json recruiter.json

# Seeker 注册
playwright-cli -s=seeker open http://localhost:3000/register
# → snapshot → fill email "seeker.frank@email.com" / password "password123"
# → click submit → await redirect to /dashboard
playwright-cli -s=seeker snapshot
# 验证: 页面包含 "Dashboard" 标题

# Recruiter 注册
playwright-cli -s=recruiter open http://localhost:3000/register
# → fill email "hr.bob@techcorp.com" / password "password123"
# → click submit → await redirect to /dashboard
playwright-cli -s=recruiter snapshot
# 验证: 页面包含 "Dashboard"

# 保存登录状态 (避免后续重复登录)
playwright-cli -s=seeker state-save seeker.json
playwright-cli -s=recruiter state-save recruiter.json

# 检查 console: 只有 favicon 404 (expected)
playwright-cli -s=seeker console error
```

### TC2: 种子数据验证 + 确认 Match (来源: 旧 T3+T4)

```bash
# 1. 用 curl 获取 seed 数据中第一个 pending match
MATCH_ID=$(curl -s http://localhost:8080/api/matches \
  -H "Authorization: Bearer $(curl -s -X POST http://localhost:8080/api/auth/login \
    -d '{"email":"seeker.frank@email.com","password":"password123"}' | jq -r '.token')" \
  | jq -r '.data[] | select(.status=="pending") | .id' | head -1)

echo "Match ID: $MATCH_ID"

# 2. 确认 Match (→ autoStartA2A 触发)
curl -s -X POST "http://localhost:8080/api/matches/$MATCH_ID/confirm" \
  -H "Authorization: Bearer $TOKEN"

# 3. 验证 Match 状态 → mutual_interest
curl -s "http://localhost:8080/api/matches/$MATCH_ID" | jq '.status'
# EXPECTED: "mutual_interest"
```

### TC3: Conversation 页面加载 (来源: 旧 T4)

```bash
# Seeker 加载 conversation 页面
playwright-cli -s=seeker goto "http://localhost:3000/conversation/$MATCH_ID"
playwright-cli -s=seeker snapshot

# 验证点:
# - FSMStatusBar 可见且显示正确阶段 (INTRODUCTION)
# - WebSocket 状态显示 "Live" / "Connected"
# - 聊天区域可见
# - 0 个 business console error
playwright-cli -s=seeker console error
```

### TC4: A2A 自动对话 — 无人类干预推进 (来源: 旧 T5+T6)

**核心验证**: autoStartA2A 自动发布了 INQUIRY → Agent 回复 → UI 显示

```bash
# 1. 等几秒让 autoStartA2A 触发
sleep 5

# 2. snapshot 看第一条 AI 回复是否出现
playwright-cli -s=seeker snapshot
# 验证: 聊天区域有 agent 回复文本 (非 XML)
# 验证: 文本中不应有 <message> <payload> 等 XML 标签

# 3. 查看后端日志确认 FSM 推进
tail -100 /tmp/backend.log | grep "FSM state changed"
# EXPECTED: "FSM state changed: intent=INTRODUCTION -> state=mutual_interest"
#           "FSM state changed: intent=INTEREST -> state=negotiating"
```

### TC5: 发送人类消息 (来源: 旧 T5)

**核心验证**: handleSubmit → 只 POST REST → 不调 sendMessage() → agent_response 经 WS 回流

```bash
# 1. 打开 DevTools 网络记录
playwright-cli -s=seeker requests

# 2. 在聊天输入框输入消息
playwright-cli -s=seeker type "Tell me more about the team structure"
playwright-cli -s=seeker press Enter

# 3. 等几秒让 AI 回复
sleep 8

# 4. 验证: 人类消息出现 + Agent 回复出现
playwright-cli -s=seeker snapshot
# 验证: 聊天区域新增了两条消息 (人类 + Agent)

# 5. 验证: 只有 POST /api/messages/:matchId, 没有 POST /api/chat-proxy
playwright-cli -s=seeker requests
# 注意: /api/chat-proxy 可能作为降级出现, 但不应该是主通道

# 6. 验证: FSM 状态更新
playwright-cli -s=seeker snapshot | grep -i "stage\|status"
```

### TC6: FSM 自动推进 → 人类确认 (来源: 旧 T6+T8)

等待多轮 Agent 对话后验证到达 Offer/Schedule 阶段：

```bash
# 1. 循环等待 FSM 推进 (最多等 60s)
for i in $(seq 1 12); do
  FSM_STATUS=$(curl -s "http://localhost:8080/api/matches/$MATCH_ID" | jq -r '.status')
  echo "Loop $i: FSM=$FSM_STATUS"
  if [ "$FSM_STATUS" = "offered" ] || [ "$FSM_STATUS" = "interviewing" ]; then
    echo "FSM reached confirmation point: $FSM_STATUS"
    break
  fi
  sleep 5
done

# 2. 验证前端显示确认模态框
playwright-cli -s=seeker snapshot
# 验证: HumanConfirmModal 可见
# 验证: 模态框包含 "review and confirm" 文字

# 3. 检查 console
playwright-cli -s=seeker console error
# 只有 favicon 404 允许
```

### TC7: 人类批准 OFFER (来源: 旧 T8)

```bash
# 1. 点击 "Approve" 按钮
playwright-cli -s=seeker click "Approve"

# 2. 验证 Match 状态 → hired
sleep 2
curl -s "http://localhost:8080/api/matches/$MATCH_ID" | jq '.status'
# EXPECTED: "hired"

# 3. 验证后端日志
tail -5 /tmp/backend.log | grep "human_confirm\|hired"
# EXPECTED: "Human APPROVED confirmation for match" 
#           "match status updated to hired"

# 4. 验证 Modal 消失 + 系统消息出现
playwright-cli -s=seeker snapshot
# 验证: Modal 关闭
# 验证: 聊天区域最后一条消息是 "Offer accepted"
```

## 4. 边界测试 (可黄, 记为 Known Issue)

### TC8: 页面刷新恢复历史

```bash
# 1. 刷新 conversation 页面
playwright-cli -s=seeker reload
sleep 2
playwright-cli -s=seeker snapshot

# 验证: 历史消息完整 (消息数量与刷新前一致)
# 验证: FSMStatusBar 显示正确阶段
# 验证: WebSocket 重新连接
```

### TC9: 10 轮 A2A 不中断

```bash
# 使用 Mock LLM 模式 (回复快速且可预测)
# 修改 .env: AI_MOCK_MODE=true

# 启动新 match → confirm → 等 10 轮对话完成
# 验证: 没有因为超时或错误中断
# 验证: DB 中消息数 ≥ 20 (10 轮 × 2 Agent)
```

### TC10: 多 Match 并发

```bash
# 为同一个 Seeker 确认 2 个不同的 Match
MATCH_A="..."
MATCH_B="..."

# 同时确认
echo "Confirming both matches..."
curl -X POST "$BASE/api/matches/$MATCH_A/confirm" &
curl -X POST "$BASE/api/matches/$MATCH_B/confirm" &
wait

# 验证两个对话独立进行
# 验证 A2A 在两个 match 中都触发
# 验证消息没有交叉
```

## 5. 调试与报告

### 失败时输出上下文

每轮测试结束后/失败时输出调试信息脚本:

```bash
dump_debug_context() {
  local match_id=$1
  echo "=== DEBUG CONTEXT ==="
  echo "Match ID: $match_id"

  # DB 状态
  curl -s "http://localhost:8080/api/matches/$match_id" | jq '{status, score, created_at}'

  # 对话轮数
  curl -s "http://localhost:8080/api/messages/$match_id" | jq 'length'
  echo "Total messages: $(curl -s "http://localhost:8080/api/messages/$match_id" | jq '.data | length')"

  # 最后 5 条消息
  curl -s "http://localhost:8080/api/messages/$match_id" | jq '.data[-5:] | .[] | {intent_type, sender, created_at}'

  # FSM 日志
  tail -20 /tmp/backend.log | grep -E "FSM|broadcast|human_confirm"

  # Agent 轮数
  grep "conversationRounds" /tmp/backend.log | tail -3

  echo "=== END DEBUG ==="
}
```

### 通过标准

| 级别 | 测试 | 要求 |
|------|------|------|
| **P0** | TC1-TC7 | **100% 绿**，失败即阻止发布 |
| **边界** | TC8-TC10 | 可黄，但需记录 Issue |
| **Console** | 所有 | 仅允许 favicon 404 + /api/chat-proxy 400 |

## 6. 执行脚本

```bash
#!/bin/bash
set -euo pipefail

# 0. 清理
rm -f seeker.json recruiter.json /tmp/debug-*.txt
playwright-cli kill-all

# 1. 基础设施检查
echo "=== Checking services ==="
pg_isready -h localhost
curl -sf http://localhost:8080/api/health
curl -sf http://localhost:3000 > /dev/null

# 2. 种子数据
echo "=== Seeding fixtures ==="
cd joblinker/backend
go run cmd/seed/main.go

# 3. 开始后端日志监控
tail -f /tmp/backend.log > /tmp/backend-tail.log &
TAIL_PID=$!

# 4. 执行 TC1-TC7 (核心流程)
echo "=== Running P0 tests ==="
# TC1: Register + Login
# TC2: Seed verify + Confirm match
# ...

# 5. 生成报告
echo "=== Results ==="
grep -c "PASS" /tmp/test-results.txt
grep -c "FAIL" /tmp/test-results.txt

# 6. 清理
kill $TAIL_PID 2>/dev/null
playwright-cli close-all
```

## 7. Mock LLM 实现方案

在后端 `pkg/ai/client.go` 加一个 mock 模式:

```go
// 环境变量 AI_MOCK_MODE=true 时启用
func (c *Client) ChatWithTools(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
    if os.Getenv("AI_MOCK_MODE") == "true" {
        return c.mockResponse(ctx, req)
    }
    return c.realChat(ctx, req)
}

// mockResponse 返回可预测的 FSM 推进回复
var mockCycle = []string{"INTRODUCTION", "INTEREST", "NEGOTIATION", "OFFER", "CONFIRM"}
```

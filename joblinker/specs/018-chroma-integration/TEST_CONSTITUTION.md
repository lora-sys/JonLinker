# 测试宪法：防止造假的 7 条铁律

## 核心原则

测试必须验证**真实行为**，不能验证**测试代码本身的正确性**。所有测试在代码有 bug 时必须 FAIL，在代码正确时必须 PASS。

---

## 铁律

### 铁律 1：禁止 t.Skip — 用 build tag 替代

```go
// ✅ 正确：环境不可用则 FAIL
//go:build integration
func TestRealDB(t *testing.T) {
    if os.Getenv("TEST_DATABASE_URL") == "" {
        t.Fatal("TEST_DATABASE_URL environment variable is required")
    }
}

// ❌ 错误：永远不要 t.Skip
if os.Getenv("TEST_DATABASE_URL") == "" {
    t.Skip("Skipping: TEST_DATABASE_URL not set")
}
```

**规则**：如果依赖不可用，测试必须 `t.Fatal("required env var missing")`，不是 Skip。

---

### 铁律 2：禁止在测试中创建匿名 handler

```go
// ❌ 错误：匿名 handler 不测试真实代码路径
r.POST("/api/auth/register", func(c *gin.Context) {
    c.JSON(201, gin.H{"status": "ok"})
})

// ✅ 正确：使用真实 handler 实例
authHandler := handler.NewAuthHandler(userRepo, jwtSecret)
r.POST("/api/auth/register", authHandler.Register)
```

**规则**：集成测试必须使用真实 handler，禁止在测试中内联 handler 逻辑。

---

### 铁律 3：禁止 nil repo — 必须连接真实数据库

```go
// ❌ 错误：nil repo 无法测试真实数据路径
exec := agent.NewToolExecutor(nil, nil, nil, nil, nil)

// ✅ 正确：使用真实 repo 实例
exec := agent.NewToolExecutor(jobRepo, agentRepo, matchRepo, interviewRepo, offerRepo)
```

**规则**：所有 service/handler 测试必须使用真实 repo 实例或测试数据库连接。

---

### 铁律 4：断言必须验证具体内容

```go
// ❌ 错误：只检查非 nil，无法发现 bug
if err != nil {
    t.Error("unexpected error")
}

// ✅ 正确：验证具体值
if result.Data["title"] != "Senior Go Developer" {
    t.Errorf("expected title 'Senior Go Developer', got %v", result.Data["title"])
}
```

**规则**：每个断言必须验证具体值，能在代码有 bug 时 FAIL。

---

### 铁律 5：E2E 测试必须发起真实 HTTP 请求

```go
// ❌ 错误：内存操作，不测试网络层
seed := CreateSeedData()
seed.TestMatch(...)

// ✅ 正确：真实 HTTP 请求
resp, err := http.Post("http://localhost:8080/api/matches/auto", ...)
if resp.StatusCode != 201 {
    t.Errorf("expected 201, got %d", resp.StatusCode)
}
```

**规则**：E2E 测试必须向运行中的服务器发起真实 HTTP/WebSocket 请求。

---

### 铁律 6：每个测试必须包含"反硬编码"断言

```go
// 验证 query_jobs 返回的不是硬编码假数据
if job.Title == "Sample Job" || job.Title == "Mock Job" {
    t.Fatal("FAKE: query_jobs returned hardcoded mock data, not real DB data")
}

// 验证 AI 响应不是 fallback
if strings.Contains(response, "Thank you for your message") {
    t.Fatal("FAKE: AI returned fallback response, not real AI response")
}

// 验证数值不是已知的硬编码值
if salary == 150000 || salary == 120000 {
    t.Fatal("FAKE: Offer salary is hardcoded 150000/120000, not from Job.salary_min/max")
}

// 验证时间戳不是硬编码
if timestamp == "2026-04-23T00:00:00Z" {
    t.Fatal("FAKE: timestamp is hardcoded, not time.Now()")
}
```

**已知硬编码值**（出现在测试中 = FAKE）：
- `"Sample Job"`, `"Mock Job"`, `"Sample Candidate"`, `"Mock Candidate"`
- `"Thank you for your message"`, `"Thank you for your introduction"`
- `"2026-06-01T10:00:00Z"`（硬编码面试时间）
- `"2026-07-01"`（硬编码 offer 开始日期）
- `150000`, `120000`（硬编码 salary）
- `0.9`（硬编码 match score）
- `"2026-04-23T00:00:00Z"`（硬编码 timestamp）

---

### 铁律 7：CI 必须检查输出

```bash
# -count=1 禁止缓存，-v 显示详细输出
go test ./... -count=1 -v 2>&1 | tee test_output.txt

# 检查是否有 SKIP
SKIP_COUNT=$(grep -c "SKIP" test_output.txt || true)
if [ "$SKIP_COUNT" -gt 0 ]; then
    echo "ERROR: $SKIP_COUNT tests were SKIPPED"
    exit 1
fi

# 检查是否有已知假数据
if grep -q "Sample Job\|Sample Candidate\|Thank you for your message" test_output.txt; then
    echo "ERROR: Tests contain hardcoded fake data"
    exit 1
fi
```

---

## 测试基础设施

### tests/testutil/setup.go

```go
package testutil

import (
    "fmt"
    "os"
    "github.com/google/uuid"
    "github.com/stretchr/testify/require"
    "gorm.io/gorm"
)

// SetupTestDB 连接测试数据库，环境变量 TEST_DATABASE_URL 必须设置
func SetupTestDB(t *testing.T) *gorm.DB {
    dbURL := os.Getenv("TEST_DATABASE_URL")
    if dbURL == "" {
        t.Fatal("TEST_DATABASE_URL environment variable is required")
    }
    db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{})
    if err != nil {
        t.Fatalf("failed to connect to test DB: %v", err)
    }
    return db
}

// RandomEmail 生成随机测试邮箱
func RandomEmail() string {
    return fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8])
}

// RandomUUID 生成随机 UUID 字符串
func RandomUUID() string {
    return uuid.New().String()
}
```

### tests/testutil/anti_cheat.go

```go
package testutil

import (
    "strings"
    "testing"
)

// KnownHardcodedValues 已知硬编码假数据，测试中出现这些值 = FAKE
var KnownHardcodedValues = []string{
    "Sample Job",
    "Mock Job",
    "Sample Candidate",
    "Mock Candidate",
    "Thank you for your message",
    "Thank you for your introduction",
    "2026-06-01T10:00:00Z",
    "2026-07-01",
    "2026-04-23T00:00:00Z",
}

var KnownHardcodedSalaries = []int{150000, 120000}
var KnownHardcodedScore = 0.9

// AssertNotHardcoded 检查字符串不包含已知硬编码值
func AssertNotHardcoded(t *testing.T, fieldName string, value string) {
    for _, fake := range KnownHardcodedValues {
        if strings.Contains(value, fake) {
            t.Fatalf("FAKE DATA: %s contains known hardcoded value %q", fieldName, fake)
        }
    }
}

// AssertSalaryNotHardcoded 检查 salary 不是已知的硬编码值
func AssertSalaryNotHardcoded(t *testing.T, salary int) {
    for _, fake := range KnownHardcodedSalaries {
        if salary == fake {
            t.Fatalf("FAKE DATA: salary is hardcoded value %d, not from Job.salary_min/max", fake)
        }
    }
}

// AssertScoreNotHardcoded 检查 score 不是已知的硬编码值
func AssertScoreNotHardcoded(t *testing.T, score float64) {
    if score == KnownHardcodedScore {
        t.Fatalf("FAKE DATA: score is hardcoded %.1f, not from AI evaluation", score)
    }
}
```

---

## Phase 1: 删除假测试

### 必须删除的文件

| 文件 | 原因 |
|------|------|
| `tests/integration/e2e_test.go` | 全部 20 个场景是恒真断言，无任何验证 |
| `tests/unit/handler/interview_test.go` | 测试 Go map 字面量语法，不是真实功能 |
| `tests/unit/handler/offer_test.go` | 测试 Go slice 字面量，不是真实功能 |
| `tests/unit/handler/job_test.go` | 测试 Go slice 长度，不是真实功能 |
| `tests/unit/model/user_test.go` | 测试 Go struct 字段赋值，不是真实功能 |
| `tests/integration/api_test.go` | 测试假匿名 handler |
| `tests/integration/job_test.go` | 测试假匿名 handler |
| `tests/integration/message_service_test.go` | 永远 t.Skip，无实际测试 |

### 保留的有效测试

| 文件 | 保留原因 |
|------|----------|
| `tests/unit/agent/agent_test.go` | 真实 FSM 状态机逻辑 |
| `tests/unit/service/agent_service_test.go` | 真实业务逻辑 |
| `tests/unit/service/agent_prompt_service_test.go` | 真实字符串验证 |
| `tests/unit/agent/function_definitions_test.go` | 真实结构验证 |
| `tests/unit/agent/tool_executor_test.go` | 真实工具执行（需修复 UUID） |
| `internal/cache/tool_cache_test.go` | 真实缓存逻辑 |
| `pkg/proto/serializer_test.go` | 真实序列化逻辑 |
| `tests/integration/gateway_proto_test.go` | 真实中间件 |
| `tests/integration/ws_test.go` | 保留但去掉 t.Skip |

---

## Phase 2: 创建真实测试

### tests/integration/auth_real_test.go

```go
//go:build integration

package integration

import (
    "bytes"
    "encoding/json"
    "net/http"
    "testing"

    "joblinker/tests/testutil"
)

func TestRegister_Real(t *testing.T) {
    baseURL := os.Getenv("TEST_SERVER_URL")
    if baseURL == "" {
        t.Fatal("TEST_SERVER_URL environment variable is required")
    }

    email := testutil.RandomEmail()
    payload := map[string]string{
        "email":    email,
        "password": "password123",
        "role":     "seeker",
    }
    body, _ := json.Marshal(payload)

    resp, err := http.Post(baseURL+"/api/auth/register", "application/json", bytes.NewReader(body))
    if err != nil {
        t.Fatalf("failed to send request: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != 201 {
        t.Fatalf("expected 201, got %d", resp.StatusCode)
    }

    var result map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&result)

    token, ok := result["token"].(string)
    if !ok || len(token) < 50 {
        t.Fatal("expected non-empty token with length > 50")
    }

    if _, exists := result["password_hash"]; exists {
        t.Fatal("FAKE: password_hash should not be in response")
    }
}
```

### tests/integration/agent_real_test.go

```go
//go:build integration

package integration

import (
    "testing"
    "net/http"
    "encoding/json"
    "bytes"

    "joblinker/tests/testutil"
)

func TestAgentCRUD_Real(t *testing.T) {
    baseURL := os.Getenv("TEST_SERVER_URL")

    // 1. Register
    email := testutil.RandomEmail()
    token := testutil.MustRegister(t, baseURL, "seeker")

    // 2. Create Agent
    agentPayload := map[string]interface{}{
        "name":       "Test Agent " + testutil.RandomUUID()[:8],
        "agent_type": "seeker",
        "skills":     testutil.RandomSkills(),
    }
    body, _ := json.Marshal(agentPayload)
    req, _ := http.NewRequest("POST", baseURL+"/api/agents", bytes.NewReader(body))
    req.Header.Set("Authorization", "Bearer "+token)
    resp, _ := http.DefaultClient.Do(req)

    if resp.StatusCode != 201 {
        t.Fatalf("expected 201, got %d", resp.StatusCode)
    }

    var agent map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&agent)
    agentID := agent["id"].(string)

    // 3. List Agents
    req, _ = http.NewRequest("GET", baseURL+"/api/agents", nil)
    req.Header.Set("Authorization", "Bearer "+token)
    resp, _ = http.DefaultClient.Do(req)

    if resp.StatusCode != 200 {
        t.Fatalf("expected 200, got %d", resp.StatusCode)
    }

    var agents []map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&agents)

    found := false
    for _, a := range agents {
        if a["id"] == agentID {
            found = true
            // 反硬编码验证
            testutil.AssertNotHardcoded(t, "agent_name", a["name"].(string))
            break
        }
    }
    if !found {
        t.Fatal("created agent not found in list")
    }

    // 4. Data isolation: another user should not see this agent
    otherToken := testutil.MustRegister(t, baseURL, "seeker")
    req, _ = http.NewRequest("GET", baseURL+"/api/agents", nil)
    req.Header.Set("Authorization", "Bearer "+otherToken)
    resp, _ = http.DefaultClient.Do(req)

    var otherAgents []map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&otherAgents)

    for _, a := range otherAgents {
        if a["id"] == agentID {
            t.Fatal("FAKE: user B can see user A's agent — data isolation broken")
        }
    }
}
```

### tests/integration/match_real_test.go

```go
//go:build integration

package integration

func TestAutoMatch_Real(t *testing.T) {
    baseURL := os.Getenv("TEST_SERVER_URL")

    // Setup: register 2 users + create agents + create job
    seekerToken := testutil.MustRegister(t, baseURL, "seeker")
    recruiterToken := testutil.MustRegister(t, baseURL, "recruiter")
    seekerAgentID := testutil.MustCreateAgent(t, baseURL, seekerToken, "seeker")
    recruiterAgentID := testutil.MustCreateAgent(t, baseURL, recruiterToken, "recruiter")
    jobID := testutil.MustCreateJob(t, baseURL, recruiterToken)

    // Create match
    payload := map[string]interface{}{
        "seeker_agent_id": seekerAgentID,
        "recruiter_agent_id": recruiterAgentID,
        "job_id": jobID,
    }
    body, _ := json.Marshal(payload)
    req, _ := http.NewRequest("POST", baseURL+"/api/matches/auto", bytes.NewReader(body))
    req.Header.Set("Authorization", "Bearer "+recruiterToken)

    resp, _ := http.DefaultClient.Do(req)
    if resp.StatusCode != 201 {
        t.Fatalf("expected 201, got %d", resp.StatusCode)
    }

    var match map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&match)

    score := match["score"].(float64)
    if score <= 0 || score > 1 {
        t.Fatalf("expected score between 0 and 1, got %f", score)
    }

    // 反硬编码：score 不能是固定的 0.9
    testutil.AssertScoreNotHardcoded(t, score)

    fsmState := match["fsm_state"].(string)
    if fsmState != "idle" {
        t.Errorf("expected initial fsm_state 'idle', got %s", fsmState)
    }
}
```

### tests/integration/ai_response_real_test.go

```go
//go:build integration

package integration

func TestAIResponse_Real(t *testing.T) {
    baseURL := os.Getenv("TEST_SERVER_URL")

    // Create match
    seekerToken := testutil.MustRegister(t, baseURL, "seeker")
    recruiterToken := testutil.MustRegister(t, baseURL, "recruiter")
    matchID := testutil.MustCreateMatch(t, baseURL, seekerToken, recruiterToken)

    // Send INTRODUCTION message
    msgPayload := map[string]interface{}{
        "match_id":  matchID,
        "intent":    "INTRODUCTION",
        "content":   "I am a senior Go developer with 5 years experience",
    }
    body, _ := json.Marshal(msgPayload)
    req, _ := http.NewRequest("POST", baseURL+"/api/messages", bytes.NewReader(body))
    req.Header.Set("Authorization", "Bearer "+seekerToken)
    req.Header.Set("Content-Type", "application/json")

    resp, _ := http.DefaultClient.Do(req)
    if resp.StatusCode != 200 {
        t.Fatalf("expected 200, got %d", resp.StatusCode)
    }

    // Poll for AI response (max 30s)
    var aiResp map[string]interface{}
    for i := 0; i < 30; i++ {
        time.Sleep(1 * time.Second)
        req, _ = http.NewRequest("GET", baseURL+"/api/messages?match_id="+matchID, nil)
        req.Header.Set("Authorization", "Bearer "+seekerToken)
        resp, _ := http.DefaultClient.Do(req)

        json.NewDecoder(resp.Body).Decode(&aiResp)
        messages := aiResp["messages"].([]map[string]interface{})

        if len(messages) >= 2 {
            aiResp = messages[len(messages)-1]
            break
        }
    }

    content := aiResp["content_xml"].(string)
    if len(content) < 50 {
        t.Fatalf("AI response too short: %d chars, expected > 50", len(content))
    }

    // 反硬编码验证
    testutil.AssertNotHardcoded(t, "ai_response", content)

    if strings.Contains(content, "Thank you for your message") {
        t.Fatal("FAKE: AI returned fallback text, not real AI response")
    }
}
```

### tests/integration/websocket_real_test.go

```go
//go:build integration

package integration

func TestWebSocket_Connect(t *testing.T) {
    serverURL := os.Getenv("TEST_WS_URL")
    if serverURL == "" {
        t.Fatal("TEST_WS_URL environment variable is required")
    }

    token := os.Getenv("TEST_WS_TOKEN")
    if token == "" {
        t.Fatal("TEST_WS_TOKEN environment variable is required")
    }

    conn, _, err := websocket.Dial(serverURL+"/ws?token="+token, nil)
    if err != nil {
        t.Fatalf("WebSocket connection failed: %v", err)
    }
    defer conn.Close()

    // Read connected message
    _, msg, err := conn.ReadMessage()
    if err != nil {
        t.Fatalf("failed to read connected message: %v", err)
    }

    var frame map[string]interface{}
    json.Unmarshal(msg, &frame)

    if frame["type"] != "connected" {
        t.Errorf("expected first message type 'connected', got %v", frame["type"])
    }
}

func TestWebSocket_SendReceive(t *testing.T) {
    // ... similar pattern, NO t.Skip
}
```

---

## CI 验证脚本

### scripts/verify_tests.sh

```bash
#!/bin/bash
set -e

cd "$(dirname "$0")/.."

echo "Running tests with -count=1 (no cache)..."
go test ./... -count=1 -v 2>&1 | tee test_output.txt

echo ""
echo "Checking for SKIP..."
SKIP_COUNT=$(grep -c "SKIP" test_output.txt || true)
if [ "$SKIP_COUNT" -gt 0 ]; then
    echo "ERROR: $SKIP_COUNT tests were SKIPPED. All tests must PASS or FAIL."
    exit 1
fi
echo "No skipped tests."

echo ""
echo "Checking for known hardcoded fake data..."
if grep -qE "Sample Job|Mock Job|Sample Candidate|Mock Candidate|Thank you for your message" test_output.txt; then
    echo "ERROR: Tests contain or output hardcoded fake data"
    grep -nE "Sample Job|Mock Job|Sample Candidate|Mock Candidate|Thank you for your message" test_output.txt || true
    exit 1
fi
echo "No hardcoded fake data detected."

echo ""
echo "All checks passed."
```

---

## 总结

| 规则 | 违反 = |
|------|--------|
| 无 t.Skip | FAIL（不是 Skip） |
| 真实 handler | 匿名 = FAIL |
| 真实 repo | nil = FAIL |
| 具体断言 | 只检查 nil = FAIL |
| 真实 HTTP | 内存操作 = FAIL |
| 反硬编码断言 | 包含假数据 = FAIL |
| CI 检查 | 有 Skip 或假数据 = FAIL |

# Quickstart: Production Hardening

## Verification Scenarios

### 安全测试
```bash
# 验证 JWT 密钥必须从环境变量加载
JWT_SECRET="" go run cmd/server/main.go  # 应 log.Fatalf

# 验证多租户隔离
curl -s -H "X-Tenant-ID: tenant-a" http://localhost:8080/api/agents \
  -H "Authorization: Bearer $TOKEN_A" | jq '.data | length'
curl -s -H "X-Tenant-ID: tenant-b" http://localhost:8080/api/agents \
  -H "Authorization: Bearer $TOKEN_A"  # 应 403

# 验证 cookie 属性
curl -s -D - http://localhost:3000/api/auth/login \
  -X POST -H "Content-Type: application/json" \
  -d '{"email":"test@t.com","password":"Test1234!"}' | grep -i set-cookie
# 应包含 HttpOnly; Secure; SameSite=Strict
```

### A2A 对话测试
```bash
# 创建 match 后轮询消息
MATCH_ID=$(curl -s -X POST http://localhost:8080/api/matches/auto \
  -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d "{\"job_id\":\"$JOB_ID\",\"seeker_agent_id\":\"$SEEKER_ID\"}" | jq -r '.id')

# 轮询等待完整对话链
for i in $(seq 1 20); do
  INTENTS=$(curl -s "http://localhost:8080/api/messages/$MATCH_ID" \
    -H "Authorization: Bearer $TOKEN" | jq -r '.[].intent_type' | tr '\n' ' → ')
  echo "[$i] $INTENTS"
  echo "$INTENTS" | grep -q "OFFER" && { echo "✅ A2A chain complete"; break; }
  sleep 5
done
```

### 可观测性测试
```bash
# 正常状态
curl -s http://localhost:8080/health | jq .
# → {"status":"healthy","checks":{"db":"ok","rabbitmq":"ok","ai_api":"ok","chroma":"ok"}}

# 停 RabbitMQ 后
docker stop joblinker-rabbitmq
curl -s http://localhost:8080/health | jq .
# → {"status":"degraded","checks":{"db":"ok","rabbitmq":"down","ai_api":"ok","chroma":"ok"}}
```

## 验收标准速查

| 验收项 | 验证方式 |
|--------|----------|
| JWT 密钥不在仓库 | `grep -r "JWT_SECRET" backend/.env` → 空 |
| Cookie 安全 | playwright-cli cookie-list → httpOnly=true |
| 多租户隔离 | A 用户访问 B 数据 → 403 |
| A2A 完整链路 | INTRODUCTION → ... → CONFIRM |
| 无硬编码 | `grep -r "extracted_from_conversation\|Sample Job\|Mock Candidate"` → 0 |
| MQ Service 行数 | `wc -l < 80` |
| Playwright CLI 全过 | `bash tests/e2e-cli/run-all.sh` → exit 0 |

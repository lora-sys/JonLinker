# Test Constitution Checklist

## Phase 1: Delete Fake Tests

- [ ] Delete `tests/integration/e2e_test.go` (20 fake scenarios, all tautological assertions)
- [ ] Delete `tests/unit/handler/interview_test.go` (tests Go map literal syntax)
- [ ] Delete `tests/unit/handler/offer_test.go` (tests Go slice literal)
- [ ] Delete `tests/unit/handler/job_test.go` (tests Go slice length)
- [ ] Delete `tests/unit/model/user_test.go` (tests Go struct assignment)
- [ ] Delete `tests/integration/api_test.go` (tests anonymous fake handlers)
- [ ] Delete `tests/integration/job_test.go` (tests anonymous fake handlers)
- [ ] Delete `tests/integration/message_service_test.go` (always t.Skip)
- [ ] Verify remaining tests still pass after deletions

## Phase 2: Create Test Infrastructure

- [ ] Create `tests/testutil/setup.go` — SetupTestDB, SetupTestServer, RandomEmail, RandomUUID, MustRegister, MustCreateAgent, MustCreateJob
- [ ] Create `tests/testutil/anti_cheat.go` — KnownHardcodedValues, KnownHardcodedSalaries, KnownHardcodedScore, AssertNotHardcoded, AssertSalaryNotHardcoded, AssertScoreNotHardcoded
- [ ] Verify `go build ./tests/...` succeeds

## Phase 3: Create Real Tests

### Auth Real Tests
- [ ] `tests/integration/auth_real_test.go` — TestRegister_Real (201 + token + no password_hash)
- [ ] `tests/integration/auth_real_test.go` — TestRegister_DuplicateEmail (409)
- [ ] `tests/integration/auth_real_test.go` — TestLogin_Real (200 + valid token)
- [ ] `tests/integration/auth_real_test.go` — TestLogin_WrongPassword (401)
- [ ] `tests/integration/auth_real_test.go` — TestProtectedEndpoint_NoToken (401)
- [ ] All auth tests use `build tag integration` + fail if env vars missing

### Agent Real Tests
- [ ] `tests/integration/agent_real_test.go` — TestAgentCRUD_Real (register → create → list → update → delete)
- [ ] `tests/integration/agent_real_test.go` — TestAgentDataIsolation (user B cannot see user A's agents)
- [ ] All agent tests use real HTTP requests, no anonymous handlers

### Match Real Tests
- [ ] `tests/integration/match_real_test.go` — TestAutoMatch_Real (201 + score 0-1 + not 0.9)
- [ ] `tests/integration/match_real_test.go` — TestAutoMatch_FSMState (initial state = idle)

### AI Response Real Tests
- [ ] `tests/integration/ai_response_real_test.go` — TestAIResponse_Real (content length > 50 + not fallback)
- [ ] `tests/integration/ai_response_real_test.go` — TestToolExecution_Real (query_jobs returns DB data, not "Sample Job")

### WebSocket Real Tests
- [ ] `tests/integration/ws_real_test.go` — TestWebSocket_Connect (no t.Skip, fail if cannot connect)
- [ ] `tests/integration/ws_real_test.go` — TestWebSocket_SendReceive (AI response, not hardcoded switch/case)
- [ ] `tests/integration/ws_real_test.go` — TestWebSocket_MatchIsolation (different matchIDs don't cross-talk)

### Memory Real Tests
- [ ] `tests/integration/memory_real_test.go` — TestMemoryPersistence (StoreMemory → GetRecentMemories returns non-empty)

### Security Real Tests
- [ ] `tests/integration/security_real_test.go` — TestAdminEndpoint_NonAdmin (403 for seeker token)
- [ ] `tests/integration/security_real_test.go` — TestPasswordHashNotLeaked (response has no password_hash key)

## Phase 4: CI Script

- [ ] Create `scripts/verify_tests.sh` (executable)
- [ ] CI checks: no SKIP, no hardcoded fake data, exit code 0
- [ ] Run `scripts/verify_tests.sh` — must pass

## Phase 5: Phase 0 Chroma Begin

- [ ] Create `pkg/chroma/client.go` (CreateCollection, GetOrCreateCollection, Add, Query, Delete)
- [ ] Verify Chroma Docker container reachable: `curl localhost:8000/api/v1/heartbeat`
- [ ] Chroma client passes unit tests

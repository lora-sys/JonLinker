# Quickstart: A2A Agent Recruitment Platform

## Prerequisites

- Docker and Docker Compose installed
- Node.js 18+ (for frontend development)
- Go 1.23+ (for backend development)
- Git

## Development Environment Setup

### 1. Clone and Start Infrastructure

```bash
# Start PostgreSQL, Redis, RabbitMQ, Chroma
docker-compose up -d

# Verify services are running
docker-compose ps
```

### 2. Backend Setup

```bash
cd backend

# Install dependencies
go mod download

# Run migrations
go run cmd/server/main.go migrate

# Start backend server
go run cmd/server/main.go server
# Server runs on http://localhost:8080
```

### 3. Frontend Setup

```bash
cd frontend

# Install dependencies
npm install

# Start development server
npm run dev
# Server runs on http://localhost:3000
```

## First-Time Usage

### 1. Create an Account

Navigate to http://localhost:3000/register and create an account.

### 2. Create Your Agent

- For job seekers: Create a Seeker Agent
- For recruiters: Create a Recruiter Agent

### 3. Input Profile/Job Data

Choose your input method:
- **Manual**: Fill in fields directly
- **AI Generation**: Describe yourself in one sentence
- **File Upload**: Upload PDF or Word resume

### 4. Activate Your Agent

Set your agent to "Active" to begin receiving matches.

## Architecture Overview

```
┌─────────────────┐     ┌─────────────────┐
│   Next.js 16    │     │    Go + Gin     │
│   Frontend      │────▶│    Backend      │
│  (TypeScript)   │     │   (REST/WS)     │
└────────┬────────┘     └────────┬────────┘
         │                       │
         │  Encrypted            │ Plain
         ▼                       ▼
┌─────────────────┐     ┌─────────────────┐
│   IndexedDB     │     │  PostgreSQL 17  │
│  (Local Only)   │     │   (Business)    │
└─────────────────┘     └────────┬────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │  Chroma 0.6+    │
                        │   (Vectors)     │
                        └─────────────────┘
```

## Key Features

### Autonomous Matching
Agents are automatically matched based on skill vectors. No manual searching required.

### A2A Negotiation
Seeker and Recruiter Agents negotiate salary, schedule, and requirements autonomously via XML protocol.

### Privacy-First
Raw resume data never leaves your device. Only encrypted vectors are stored on the server.

### Real-Time Updates
WebSocket connection delivers dialogue updates instantly as Agents negotiate.

## Testing

```bash
# Backend tests
cd backend && go test ./...

# Frontend E2E tests
cd frontend && npx playwright test
```

## Project Structure

```
├── frontend/              # Next.js 16 application
│   ├── src/
│   │   ├── app/         # App Router pages
│   │   ├── components/  # React components
│   │   ├── lib/         # Utilities
│   │   ├── stores/      # Zustand state
│   │   └── types/       # TypeScript types
│   └── tests/           # Playwright E2E
│
├── backend/              # Go application
│   ├── cmd/server/      # Entry point
│   ├── internal/
│   │   ├── handler/     # HTTP handlers
│   │   ├── service/     # Business logic
│   │   ├── repository/  # Data access
│   │   ├── model/       # Domain models
│   │   └── agent/       # FSM engine
│   └── migrations/      # DB migrations
│
├── docker-compose.yml    # Infrastructure
└── specs/               # Feature specifications
```

## Troubleshooting

### Services won't start
```bash
# Check Docker status
docker-compose ps

# View logs
docker-compose logs postgres
docker-compose logs redis
```

### Backend connection errors
```bash
# Verify backend is running
curl http://localhost:8080/health

# Check backend logs for errors
```

### Frontend build errors
```bash
# Clear node_modules and reinstall
rm -rf frontend/node_modules
cd frontend && npm install
```

## Next Steps

- Read the full specification: `specs/003-agent-recruit-platform/spec.md`
- Review the data model: `specs/003-agent-recruit-platform/data-model.md`
- Understand the A2A protocol: `specs/003-agent-recruit-platform/contracts/a2a-protocol.md`
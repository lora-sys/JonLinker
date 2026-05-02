#!/bin/bash
# Generate Protobuf code for Go and TypeScript
# Usage: ./scripts/generate-proto.sh

set -e

PROJECT_ROOT=$(cd "$(dirname "$0")/.." && pwd)
PROTO_DIR="$PROJECT_ROOT/proto"
OUT_GO="$PROJECT_ROOT/backend/pkg/proto"
OUT_TS="$PROJECT_ROOT/frontend/src/lib/proto"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}=== Protobuf Code Generation ===${NC}"

# Check protoc
if ! command -v protoc &> /dev/null; then
    echo -e "${RED}Error: protoc is not installed${NC}"
    echo "Install: brew install protobuf (macOS) or apt install protobuf-compiler (Ubuntu)"
    exit 1
fi

# Check protoc-gen-go
if ! command -v protoc-gen-go &> /dev/null; then
    echo -e "${YELLOW}Installing protoc-gen-go...${NC}"
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi

# Check protoc-gen-ts (ts-proto)
TS_PROTO_LOCAL="$PROJECT_ROOT/frontend/node_modules/.bin/protoc-gen-ts_proto"
if [ -f "$TS_PROTO_LOCAL" ]; then
    echo -e "${GREEN}Using local ts-proto from node_modules${NC}"
    PROTO_TS_BIN="$TS_PROTO_LOCAL"
else
    echo -e "${YELLOW}Warning: ts-proto not found${NC}"
    PROTO_TS_BIN="protoc-gen-ts"
fi

# Check if proto files exist
if [ ! "$(ls -A "$PROTO_DIR"/*.proto 2>/dev/null)" ]; then
    echo -e "${YELLOW}No .proto files found in $PROTO_DIR${NC}"
    exit 0
fi

# Create output directories
mkdir -p "$OUT_GO"
mkdir -p "$OUT_TS"

# Generate Go code
echo -e "${GREEN}Generating Go code...${NC}"
protoc \
    --proto_path="$PROTO_DIR" \
    --go_out="$OUT_GO" \
    --go_opt=paths=source_relative \
    "$PROTO_DIR"/*.proto

# Generate TypeScript code
echo -e "${GREEN}Generating TypeScript code...${NC}"
protoc \
    --proto_path="$PROTO_DIR" \
    --ts_out="$OUT_TS" \
    --ts_opt=outputServices=false \
    --ts_opt=useDate=false \
    --ts_opt=addGrpcMetadata=false \
    --ts_opt=nestedStdlib=true \
    --plugin="protoc-gen-ts=$PROTO_TS_BIN" \
    "$PROTO_DIR"/*.proto

echo -e "${GREEN}Done!${NC}"
echo "  Go: $OUT_GO"
echo "  TS: $OUT_TS"
echo ""
echo "Proto files processed:"
ls "$PROTO_DIR"/*.proto 2>/dev/null | xargs -I {} basename {}

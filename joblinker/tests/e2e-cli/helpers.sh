#!/usr/bin/env bash
set -euo pipefail

# Helpers for playwright-cli E2E tests

random_email() {
  echo "test-$(date +%s)-$RANDOM@test.com"
}

random_name() {
  echo "TestAgent-$RANDOM"
}

wait_for_message() {
  local match_id=$1
  local token=$2
  local max_wait=${3:-120}
  local interval=${4:-3}
  local target_intent=${5:-"OFFER"}

  echo "Waiting for $target_intent in match $match_id..."
  for i in $(seq 1 $((max_wait / interval))); do
    local intents
    intents=$(curl -s "http://localhost:8080/api/messages/$match_id" \
      -H "Authorization: Bearer $token" \
      | jq -r '.[].intent_type' 2>/dev/null | tr '\n' ' ')
    if echo "$intents" | grep -q "$target_intent"; then
      echo "✅ Found $target_intent after $((i * interval))s"
      echo "Intents: $intents"
      return 0
    fi
    sleep "$interval"
  done
  echo "❌ Timeout waiting for $target_intent"
  return 1
}

assert_no_hardcoded() {
  local value=$1
  local label=$2
  local bad_values=("Sample Job" "Mock Candidate" "Thank you for your message" "extracted_from_conversation")
  for bad in "${bad_values[@]}"; do
    if echo "$value" | grep -q "$bad"; then
      echo "❌ $label contains hardcoded value: $bad"
      return 1
    fi
  done
  echo "✅ $label: no hardcoded values"
}

assert_not_equal() {
  local v1=$1
  local v2=$2
  local label=$3
  if [ "$v1" = "$v2" ]; then
    echo "❌ $label: unexpected match $v1 == $v2"
    return 1
  fi
  echo "✅ $label: values differ"
}

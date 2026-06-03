import type { UIMessagePart, UIDataTypes, UITools } from "ai";
import { isTextUIPart, isReasoningUIPart } from "ai";

export function getTextFromParts(parts: UIMessagePart<UIDataTypes, UITools>[]): string {
  const textParts = parts.filter(
    (p): p is Extract<UIMessagePart<UIDataTypes, UITools>, { type: "text" | "reasoning" }> =>
      isTextUIPart(p) || isReasoningUIPart(p),
  );
  return textParts.map((p) => p.text).join("");
}

export function getToolCallName(part: UIMessagePart<UIDataTypes, UITools>): string {
  if ("toolName" in part) return (part as any).toolName;
  if (part.type.startsWith("tool-")) return part.type.slice("tool-".length);
  return "unknown";
}

export function getToolStateLabel(state: string): string {
  const labels: Record<string, string> = {
    "input-streaming": "执行中",
    "input-available": "待执行",
    "approval-requested": "等待确认",
    "approval-responded": "已确认",
    "output-available": "已完成",
    "output-error": "执行失败",
    "output-denied": "已拒绝",
  };
  return labels[state] || state;
}

export const TOOL_DISPLAY_MAP: Record<string, { icon: string; label: string }> = {
  query_jobs: { icon: "🔍", label: "搜索职位" },
  apply_job: { icon: "📝", label: "生成求职申请" },
  parseresume: { icon: "📄", label: "解析简历" },
};

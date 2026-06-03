"use client";

import { memo } from "react";
import { cn } from "@/lib/utils";
import { Spinner } from "@/components/ui/spinner";
import { getToolStateLabel, TOOL_DISPLAY_MAP } from "@/lib/ai-utils";

type ToolCallCardProps = {
  toolName: string;
  state: string;
  isStreaming: boolean;
  errorText?: string;
};

export const ToolCallCard = memo(function ToolCallCard({
  toolName,
  state,
  isStreaming,
  errorText,
}: ToolCallCardProps) {
  const display = TOOL_DISPLAY_MAP[toolName] ?? { icon: "🔧", label: toolName };
  const running = state === "input-streaming" || state === "input-available";
  const complete = state === "output-available";
  const error = state === "output-error" || state === "output-denied";
  const stateLabel = getToolStateLabel(state);

  return (
    <div
      className={cn(
        "flex items-center gap-2 rounded-lg border px-3 py-2 text-sm",
        running && "bg-muted/50 border-muted",
        complete && "bg-green-50 border-green-200 dark:bg-green-950/30 dark:border-green-800",
        error && "bg-red-50 border-red-200 dark:bg-red-950/30 dark:border-red-800",
      )}
    >
      {running && (isStreaming ? <Spinner className="size-4 shrink-0" /> : <span className="size-4 shrink-0">○</span>)}
      {complete && <span className="text-green-600 shrink-0">✓</span>}
      {error && <span className="text-red-600 shrink-0">✗</span>}
      <span>{display.icon} {display.label}</span>
      <span className="text-xs text-muted-foreground ml-auto">{stateLabel}</span>
      {error && errorText && (
        <span className="text-xs text-red-600 ml-2">{errorText}</span>
      )}
    </div>
  );
});

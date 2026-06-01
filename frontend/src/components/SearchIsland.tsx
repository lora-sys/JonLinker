"use client";

import { useState } from "react";
import { useChat } from "@ai-sdk/react";
import { DefaultChatTransport } from "ai";
import {
  Conversation, ConversationContent, ConversationEmptyState, ConversationScrollButton,
} from "@/components/ai-elements/conversation";
import {
  Message, MessageContent, MessageResponse,
} from "@/components/ai-elements/message";
import {
  PromptInput, PromptInputTextarea, PromptInputSubmit,
} from "@/components/ai-elements/prompt-input";
import { getTextFromParts } from "@/lib/chat";
import type { RankedJob, Application } from "@/lib/types";
import { JobCard } from "@/components/JobCard";
import { ApplicationCard } from "@/components/ApplicationCard";

export function SearchIsland({ sessionId }: { sessionId: string | null }) {
  const [applying, setApplying] = useState<string | null>(null);
  const [structured, setStructured] = useState<{ jobs?: RankedJob[]; application?: Application }>({});

  const { messages, sendMessage, setMessages, status, error } = useChat({
    transport: new DefaultChatTransport({
      api: '/api/chat',
      body: sessionId ? { session_id: sessionId } : undefined,
    }),
    onData: (dataPart: any) => {
      if (dataPart.type === 'data-jobs' && Array.isArray(dataPart.data)) {
        setStructured({ jobs: dataPart.data });
      } else if (dataPart.type === 'data-application' && dataPart.data) {
        setStructured({ application: dataPart.data });
      }
    },
  });

  async function handleApply(job: RankedJob) {
    if (!sessionId) {
      setMessages((prev) => [...prev,
        { id: `user-${Date.now()}`, role: "user", parts: [{ type: "text" as const, text: "申请职位" }] },
        { id: "apply-hint", role: "assistant", parts: [{ type: "text" as const, text: "⚠️ 请先在左侧「简历助手」上传简历并完善个人资料，然后才能生成申请。" }] },
      ]);
      return;
    }
    setApplying(job.url);
    try {
      const res = await fetch("/api/apply", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ job_url: job.url, session_id: sessionId }),
      });
      if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || "apply failed");
      const app: Application = await res.json();
      setMessages((prev) => [...prev,
        { id: `user-${Date.now()}`, role: "user", parts: [{ type: "text" as const, text: `申请职位: ${job.title}` }] },
      ]);
      setStructured({ application: app });
    } catch (e) {
      /* error handled silently */
    } finally { setApplying(null); }
  }

  const streaming = status === "streaming" || status === "submitted";

  return (
    <div className="flex flex-col h-full bg-card rounded-xl border overflow-hidden">
      <div className="px-4 py-3 border-b">
        <h2 className="text-sm font-semibold">职位搜索</h2>
        <p className="text-xs text-muted-foreground mt-0.5">聊天搜索职位，AI 智能匹配</p>
      </div>

      <Conversation>
        <ConversationContent>
          {messages.length === 0 ? (
            <ConversationEmptyState
              title="开始搜索职位"
              description="描述你想要的职位，例如：找北京的前端岗位，3-5年经验"
            />
          ) : (
            messages.map((m, i) => (
              <Message key={m.id} from={m.role}>
                <MessageContent>
                  {getTextFromParts(m.parts) && (
                    <MessageResponse
                      mode={streaming && i === messages.length - 1 ? "streaming" : "static"}
                      isAnimating={streaming && i === messages.length - 1}
                      animated
                      caret={streaming && i === messages.length - 1 ? "block" : undefined}
                    >
                      {getTextFromParts(m.parts)}
                    </MessageResponse>
                  )}
                  {error && i === messages.length - 1 && (
                    <div className="p-3 bg-destructive/10 border border-destructive/20 rounded-lg text-sm text-destructive mt-2">
                      {error.message || "请求失败"}
                    </div>
                  )}
                  {i === messages.length - 1 && structured.jobs && (
                    <div className="space-y-3 mt-2">
                      <p className="text-xs text-muted-foreground">找到 {structured.jobs.length} 个匹配职位：</p>
                      {structured.jobs.map((job, j) => (
                        <JobCard
                          key={j} job={job} sessionId={sessionId}
                          applying={applying} onApply={handleApply}
                        />
                      ))}
                    </div>
                  )}
                  {i === messages.length - 1 && structured.application && (
                    <ApplicationCard app={structured.application} />
                  )}
                </MessageContent>
              </Message>
            ))
          )}
        </ConversationContent>
        <ConversationScrollButton />
      </Conversation>

      <PromptInput
        onSubmit={async (msg) => {
          if (msg.text.trim()) {
            sendMessage({ text: msg.text });
          }
        }}
        className="border-t p-3"
      >
        <PromptInputTextarea placeholder="输入职位需求..." />
        <PromptInputSubmit status={streaming ? "streaming" : "ready"} />
      </PromptInput>
    </div>
  );
}

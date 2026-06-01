"use client";

import { useState } from "react";
import { useSSEChat } from "@/lib/chat";
import {
  Conversation, ConversationContent, ConversationEmptyState, ConversationScrollButton,
} from "@/components/ai-elements/conversation";
import {
  Message, MessageContent, MessageResponse,
} from "@/components/ai-elements/message";
import {
  PromptInput, PromptInputTextarea, PromptInputSubmit,
} from "@/components/ai-elements/prompt-input";

export function ResumeChatIsland({ onSessionReady }: { onSessionReady: (sid: string) => void }) {
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [uploading, setUploading] = useState(false);
  const [profileReady, setProfileReady] = useState(false);
  const { messages, streaming, send, reset } = useSSEChat(sessionId, () => setProfileReady(true));

  async function handleUpload(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setUploading(true);
    setProfileReady(false);
    reset();
    try {
      const fd = new FormData();
      fd.set("file", file);
      const res = await fetch("/api/resume/upload", { method: "POST", body: fd });
      const data = await res.json();
      if (data.session_id) {
        setSessionId(data.session_id);
        onSessionReady(data.session_id);
      }
    } finally {
      setUploading(false);
    }
  }

  if (!sessionId) {
    return (
      <div className="flex flex-col h-full bg-card rounded-xl border overflow-hidden">
        <div className="px-4 py-3 border-b">
          <h2 className="text-sm font-semibold">简历助手</h2>
          <p className="text-xs text-muted-foreground mt-0.5">上传简历并聊天完善资料</p>
        </div>
        <div className="flex-1 flex flex-col items-center justify-center gap-4 p-6">
          <label className="cursor-pointer inline-flex items-center gap-2 px-6 py-3 bg-primary text-primary-foreground text-sm font-medium rounded-lg hover:opacity-90 transition-opacity">
            {uploading ? (
              <><span className="animate-spin h-4 w-4 border-2 border-current border-t-transparent rounded-full" /> 解析中...</>
            ) : (
              "上传 PDF 简历"
            )}
            <input type="file" accept=".pdf" className="hidden" onChange={handleUpload} disabled={uploading} />
          </label>
          <p className="text-xs text-muted-foreground">支持 PDF 格式</p>
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full bg-card rounded-xl border overflow-hidden">
      <div className="px-4 py-3 border-b flex items-center justify-between">
        <div>
          <h2 className="text-sm font-semibold">简历助手</h2>
          <p className="text-xs text-muted-foreground mt-0.5">聊天完善个人资料</p>
        </div>
        {profileReady && (
          <span className="text-xs text-green-600 bg-green-50 px-2 py-0.5 rounded-full">资料已就绪</span>
        )}
      </div>

      <Conversation>
        <ConversationContent>
          {messages.length === 0 ? (
            <ConversationEmptyState
              title="简历已解析完成"
              description="开始聊天完善资料..."
            />
          ) : (
            messages.map((m, i) => (
              <Message key={i} from={m.role as "user" | "assistant"}>
                <MessageContent>
                  <MessageResponse
                    mode="streaming"
                    isAnimating={streaming && i === messages.length - 1}
                    animated
                    caret="block"
                  >
                    {m.content}
                  </MessageResponse>
                </MessageContent>
              </Message>
            ))
          )}
        </ConversationContent>
        <ConversationScrollButton />
      </Conversation>

      <PromptInput
        onSubmit={async (msg) => { if (msg.text.trim()) await send(msg.text); }}
        className="border-t p-3"
      >
        <PromptInputTextarea placeholder="输入信息..." />
        <PromptInputSubmit status={streaming ? "streaming" : "ready"} />
      </PromptInput>
    </div>
  );
}

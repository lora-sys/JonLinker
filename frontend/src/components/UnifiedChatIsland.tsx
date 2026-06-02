"use client";

import { useState, useMemo, useEffect, useCallback } from "react";
import { useChat } from "@ai-sdk/react";
import { DefaultChatTransport } from "ai";
import {
  Conversation, ConversationContent, ConversationEmptyState, ConversationScrollButton,
} from "@/components/ai-elements/conversation";
import {
  Message, MessageContent, MessageResponse,
} from "@/components/ai-elements/message";
import {
  PromptInput,
  PromptInputActionAddAttachments,
  PromptInputActionMenu,
  PromptInputActionMenuContent,
  PromptInputActionMenuTrigger,
  PromptInputBody,
  PromptInputHeader,
  PromptInputSubmit,
  PromptInputTextarea,
  PromptInputFooter,
  PromptInputTools,
  usePromptInputAttachments,
  type PromptInputMessage,
} from "@/components/ai-elements/prompt-input";
import { XIcon } from "lucide-react";
import { getTextFromParts } from "@/lib/ai-utils";
import { ErrorBanner } from "@/components/ui/error-banner";
import type { RankedJob, Application } from "@/lib/types";
import { JobCard } from "@/components/JobCard";
import { ApplicationCard } from "@/components/ApplicationCard";

function ResumeAttachment() {
  const attachments = usePromptInputAttachments();
  const file = attachments.files[0];
  if (!file) return null;
  return (
    <div className="flex items-center gap-2 px-3 py-1.5 bg-muted rounded-lg text-xs">
      <span className="truncate max-w-[200px]">{file.filename}</span>
      <button onClick={() => attachments.remove(file.id)} className="hover:text-destructive">
        <XIcon size={14} />
      </button>
    </div>
  );
}

export function UnifiedChatIsland() {
  const [sessionId, setSessionId] = useState<string | null>(null);
  const [uploading, setUploading] = useState(false);
  const [profileReady, setProfileReady] = useState(false);
  const [applying, setApplying] = useState<string | null>(null);
  const [application, setApplication] = useState<Application | undefined>();
  const [uploadError, setUploadError] = useState<string | null>(null);
  const [applyError, setApplyError] = useState<string | null>(null);

  const transport = useMemo(
    () => new DefaultChatTransport({
      api: "/api/chat/unified",
      body: sessionId ? { session_id: sessionId } : undefined,
    }),
    [sessionId],
  );

  const { messages, sendMessage, status, error } = useChat({
    id: sessionId ?? "no-session",
    transport,
    experimental_throttle: 50,
  });

  const streaming = status === "streaming" || status === "submitted";

  const jobs = useMemo(() => {
    for (let i = messages.length - 1; i >= 0; i--) {
      const msg = messages[i];
      if (msg.role === 'assistant') {
        for (const part of msg.parts) {
          if (part.type === 'data-jobs' && Array.isArray((part as any).data)) {
            return (part as any).data as RankedJob[];
          }
        }
        break;
      }
    }
  }, [messages]);

  const appFromMessages = useMemo(() => {
    for (let i = messages.length - 1; i >= 0; i--) {
      const msg = messages[i];
      for (const part of msg.parts) {
        if (part.type === 'data-application' && (part as any).data) {
          return (part as any).data as Application;
        }
      }
    }
    return undefined;
  }, [messages]);

  useEffect(() => {
    for (const msg of messages) {
      for (const part of msg.parts) {
        if (part.type === "data-resume") {
          if ((part as any).data?.complete) {
            setProfileReady(true);
            return;
          }
        }
      }
    }
  }, [messages]);

  async function uploadResume(file: File): Promise<string | null> {
    setUploading(true);
    setUploadError(null);
    setApplication(undefined);
    setApplyError(null);
    try {
      const fd = new FormData();
      fd.set("file", file);
      const res = await fetch("/api/resume/upload", { method: "POST", body: fd });
      if (!res.ok) {
        const errData = await res.json().catch(() => ({}));
        throw new Error(errData.error || "上传失败，请重试");
      }
      const data = await res.json();
      return data.session_id || null;
    } catch (e) {
      setUploadError(e instanceof Error ? e.message : "上传失败，请重试");
      return null;
    } finally {
      setUploading(false);
    }
  }

  const handleSubmit = useCallback(async (msg: PromptInputMessage) => {
    const hasText = Boolean(msg.text?.trim());
    const hasFiles = msg.files.length > 0;

    if (!hasText && !hasFiles) return;

    // First upload: attach PDF → upload to backend → get sessionId
    if (!sessionId && hasFiles) {
      const pdf = msg.files[0];
      const resp = await fetch(pdf.url);
      const blob = await resp.blob();
      const file = new File([blob], pdf.filename || "resume.pdf", { type: pdf.mediaType || "application/pdf" });
      const sid = await uploadResume(file);
      if (!sid) return;
      setSessionId(sid);
      setProfileReady(true);
      // Send the text message after session is ready
      if (msg.text?.trim()) {
        sendMessage({ text: msg.text.trim() });
      }
      return;
    }

    if (hasText) {
      sendMessage({ text: msg.text.trim() });
    }
  }, [sessionId, sendMessage, uploadResume]);

  async function handleApply(job: RankedJob) {
    if (!sessionId) return;
    setApplyError(null);
    setApplying(job.url);
    try {
      const res = await fetch("/api/apply", {
        method: "POST", headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ job_url: job.url, session_id: sessionId }),
      });
      if (!res.ok) throw new Error((await res.json().catch(() => ({}))).error || "apply failed");
      const app: Application = await res.json();
      setApplication(app);
    } catch (e) {
      setApplyError(e instanceof Error ? e.message : "申请失败，请稍后重试");
    } finally { setApplying(null); }
  }

  return (
    <div className="flex flex-col h-full bg-card rounded-xl border overflow-hidden">
      <div className="px-4 py-3 border-b flex items-center justify-between">
        <div>
          <h2 className="text-sm font-semibold">JobLinker AI 助手</h2>
          <p className="text-xs text-muted-foreground mt-0.5">聊天搜索职位，AI 智能匹配</p>
        </div>
        {profileReady && (
          <span className="text-xs text-green-600 bg-green-50 px-2 py-0.5 rounded-full">资料已就绪</span>
        )}
        {!sessionId && !profileReady && (
          <span className="text-xs text-amber-600 bg-amber-50 px-2 py-0.5 rounded-full">请先上传简历</span>
        )}
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
                    <ErrorBanner message={error.message || "请求失败"} />
                  )}
                  {i === messages.length - 1 && jobs && (
                    <div className="space-y-3 mt-2">
                      <p className="text-xs text-muted-foreground">找到 {jobs.length} 个匹配职位：</p>
                      {jobs.map((job) => (
                        <JobCard
                          key={job.url} job={job} sessionId={sessionId || ""}
                          profileReady={profileReady} applying={applying} onApply={handleApply}
                        />
                      ))}
                    </div>
                  )}
                  {i === messages.length - 1 && applyError && (
                    <ErrorBanner message={applyError} />
                  )}
                  {i === messages.length - 1 && (appFromMessages ?? application) && (
                    <ApplicationCard app={appFromMessages ?? application!} />
                  )}
                </MessageContent>
              </Message>
            ))
          )}
        </ConversationContent>
        <ConversationScrollButton />
      </Conversation>

      <PromptInput
        onSubmit={handleSubmit}
        className="border-t p-3"
        accept=".pdf"
        multiple={false}
        maxFiles={1}
      >
        <PromptInputHeader>
          {uploading ? (
            <div className="flex items-center gap-2 px-3 py-1.5 bg-muted rounded-lg text-xs">
              <span className="animate-spin h-3 w-3 border-2 border-current border-t-transparent rounded-full" />
              解析简历中...
            </div>
          ) : (
            !sessionId && <ResumeAttachment />
          )}
        </PromptInputHeader>
        <PromptInputBody>
          <PromptInputTextarea
            placeholder={!sessionId ? "📎 点击 + 上传简历开始" : "输入职位需求..."}
          />
        </PromptInputBody>
        <PromptInputFooter>
          <PromptInputTools>
            <PromptInputActionMenu>
              <PromptInputActionMenuTrigger />
              <PromptInputActionMenuContent>
                <PromptInputActionAddAttachments label="上传简历 (PDF)" />
              </PromptInputActionMenuContent>
            </PromptInputActionMenu>
          </PromptInputTools>
          <PromptInputSubmit disabled={uploading} status={status} />
        </PromptInputFooter>
      </PromptInput>

      {uploadError && (
        <div className="px-3 pb-3">
          <ErrorBanner message={uploadError} />
        </div>
      )}
    </div>
  );
}

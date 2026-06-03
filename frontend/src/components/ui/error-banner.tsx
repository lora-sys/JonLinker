export function ErrorBanner({
  message,
  onRetry,
}: {
  message: string;
  onRetry?: () => void;
}) {
  return (
    <div className="p-3 bg-destructive/10 border border-destructive/20 rounded-lg text-sm text-destructive mt-2 flex items-start justify-between gap-2">
      <span className="flex-1">{message}</span>
      {onRetry && (
        <button
          onClick={onRetry}
          className="shrink-0 text-xs font-medium underline underline-offset-2 hover:text-destructive/80 transition-colors"
        >
          重试
        </button>
      )}
    </div>
  );
}

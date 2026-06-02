export function ErrorBanner({ message }: { message: string }) {
  return (
    <div className="p-3 bg-destructive/10 border border-destructive/20 rounded-lg text-sm text-destructive mt-2">
      {message}
    </div>
  );
}

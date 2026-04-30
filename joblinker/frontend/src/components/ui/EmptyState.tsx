import React from 'react';
import Link from 'next/link';
import { Plus, Inbox } from 'lucide-react';
import Button from './Button';

interface EmptyStateProps {
  icon?: React.ReactNode;
  title: string;
  description?: string;
  actionLabel?: string;
  href?: string;
  actionHref?: string;
  onAction?: () => void;
  className?: string;
}

export default function EmptyState({
  icon,
  title,
  description,
  actionLabel,
  href,
  actionHref,
  onAction,
  className = '',
}: EmptyStateProps) {
  const finalHref = href || actionHref;
  return (
    <div className={`flex flex-col items-center justify-center py-16 px-4 text-center ${className}`}>
      <div className="w-20 h-20 rounded-2xl bg-gradient-to-br from-slate-100 to-slate-50 flex items-center justify-center mb-6">
        {icon || <Inbox className="w-10 h-10 text-slate-400" />}
      </div>
      <h3 className="text-xl font-semibold text-slate-900 mb-2">{title}</h3>
      {description && (
        <p className="text-slate-600 mb-8 max-w-md">{description}</p>
      )}
      {actionLabel && (finalHref || onAction) && (
        finalHref ? (
          <Link href={finalHref}>
            <Button className="flex items-center gap-2">
              <Plus className="w-4 h-4" />
              {actionLabel}
            </Button>
          </Link>
        ) : (
          <Button onClick={onAction} className="flex items-center gap-2">
            <Plus className="w-4 h-4" />
            {actionLabel}
          </Button>
        )
      )}
    </div>
  );
}

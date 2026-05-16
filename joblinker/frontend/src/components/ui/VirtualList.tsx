'use client'

import { useVirtualizer } from '@tanstack/react-virtual'
import React, { useRef } from 'react'

interface VirtualListProps<T> {
  items: T[]
  height: number | string
  itemHeight?: number
  overscan?: number
  renderItem: (item: T, index: number) => React.ReactNode
  className?: string
}

export default function VirtualList<T>({
  items,
  height,
  itemHeight = 60,
  overscan = 5,
  renderItem,
  className = '',
}: VirtualListProps<T>) {
  const parentRef = useRef<HTMLDivElement>(null)

  const virtualizer = useVirtualizer({
    count: items.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => itemHeight,
    overscan,
  })

  return (
    <div
      ref={parentRef}
      className={`overflow-auto ${className}`}
      style={{ height: typeof height === 'number' ? `${height}px` : height }}
    >
      <div
        style={{
          height: `${virtualizer.getTotalSize()}px`,
          width: '100%',
          position: 'relative',
        }}
      >
        {virtualizer.getVirtualItems().map(virtualItem => (
          <div
            key={virtualItem.key}
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              width: '100%',
              height: `${virtualItem.size}px`,
              transform: `translateY(${virtualItem.start}px)`,
            }}
          >
            {renderItem(items[virtualItem.index]!, virtualItem.index)}
          </div>
        ))}
      </div>
    </div>
  )
}

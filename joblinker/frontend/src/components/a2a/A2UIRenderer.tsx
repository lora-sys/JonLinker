'use client'

import React from 'react'

import type { A2UIComponent } from '@/hooks/useSSE'

import { useA2UI } from './A2UIProvider'

export function A2UIRenderer({ sessionId }: { sessionId: string }) {
  const { getSurface } = useA2UI()
  const surface = getSurface(sessionId)

  if (!surface) {
    return <div className="text-gray-400 text-sm p-4">Waiting for agent...</div>
  }

  const components = Array.from(surface.components.values())
  if (components.length === 0) {
    return <div className="text-gray-400 text-sm p-4">Waiting for agent...</div>
  }

  return (
    <div className="space-y-3 p-4">
      {components.map(comp => (
        <A2UIComponentRenderer
          key={comp.id}
          component={comp}
          dataModels={surface.dataModels}

        />
      ))}
    </div>
  )
}

function A2UIComponentRenderer({
  component,
  dataModels,
}: {
  component: A2UIComponent
  dataModels: Map<string, string>
}) {
  const { type, props, dataKey } = component
  const hint = props?.hint as string | undefined

  switch (type) {
    case 'text':
      return <TextComponent dataKey={dataKey} content={props?.content as string} dataModels={dataModels} hint={hint} />
    case 'card':
      return <CardComponent content={props?.content as string} hint={hint} />
    default:
      return <div className="text-gray-500 text-sm">{props?.content as string}</div>
  }
}

function TextComponent({
  dataKey,
  content,
  dataModels,
  hint,
}: {
  dataKey?: string
  content?: string
  dataModels: Map<string, string>
  hint?: string
}) {
  const displayText = dataKey ? (dataModels.get(dataKey) || '') : (content || '')

  if (hint === 'assistant') {
    return (
      <div className="flex justify-start">
        <div className="bg-blue-50 border border-blue-200 rounded-2xl rounded-tl-sm px-4 py-3 max-w-[80%]">
          <p className="text-sm text-blue-900 whitespace-pre-wrap">{displayText}</p>
        </div>
      </div>
    )
  }

  return <p className="text-sm text-gray-700">{displayText}</p>
}

function CardComponent({
  content,
  hint,
}: {
  content?: string
  hint?: string
}) {
  if (hint === 'tool_call') {
    return (
      <div className="bg-amber-50 border border-amber-200 rounded-xl px-4 py-3">
        <div className="flex items-center gap-2 mb-1">
          <span className="w-2 h-2 rounded-full bg-amber-500 animate-pulse" />
          <span className="text-xs font-medium text-amber-700">Using tool...</span>
        </div>
        <p className="text-sm text-amber-800 font-mono text-xs">{content}</p>
      </div>
    )
  }

  if (hint === 'tool_result') {
    return (
      <div className="bg-emerald-50 border border-emerald-200 rounded-xl px-4 py-3">
        <div className="flex items-center gap-2 mb-1">
          <span className="w-2 h-2 rounded-full bg-emerald-500" />
          <span className="text-xs font-medium text-emerald-700">Tool result</span>
        </div>
        <p className="text-sm text-emerald-800">{content}</p>
      </div>
    )
  }

  return <p className="text-sm text-gray-700">{content}</p>
}

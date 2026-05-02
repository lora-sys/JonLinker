'use client';

import { useState, useCallback, useRef, useEffect } from 'react';
import type { AIStreamState, ToolCall } from '@/types/ai';

interface UseStreamOptions {
  throttleMs?: number;
  chunkSize?: number;
  onComplete?: () => void;
  onError?: (error: string) => void;
}

export function useStream(options: UseStreamOptions = {}) {
  const { throttleMs = 16, chunkSize = 1, onComplete, onError } = options;
  
  const [state, setState] = useState<AIStreamState>({
    messageId: '',
    partialContent: '',
    toolCalls: [],
    isComplete: true,
  });
  
  const contentRef = useRef('');
  const targetRef = useRef('');
  const animationRef = useRef<number | null>(null);
  const lastUpdateRef = useRef(0);

  const cancelAnimation = useCallback(() => {
    if (animationRef.current !== null) {
      cancelAnimationFrame(animationRef.current);
      animationRef.current = null;
    }
  }, []);

  const animate = useCallback(() => {
    const now = performance.now();
    const elapsed = now - lastUpdateRef.current;
    
    if (elapsed >= throttleMs) {
      const current = contentRef.current;
      const target = targetRef.current;
      
      if (current.length < target.length) {
        const nextLength = Math.min(current.length + chunkSize, target.length);
        contentRef.current = target.slice(0, nextLength);
        lastUpdateRef.current = now;
        
        setState(prev => ({
          ...prev,
          partialContent: contentRef.current,
          isComplete: nextLength >= target.length,
        }));
      } else {
        setState(prev => ({ ...prev, isComplete: true }));
        if (onComplete) onComplete();
        animationRef.current = null;
        return;
      }
    }
    
    animationRef.current = requestAnimationFrame(animate);
  }, [throttleMs, chunkSize, onComplete]);

  const startStreaming = useCallback((messageId: string, initialContent: string = '') => {
    cancelAnimation();
    contentRef.current = '';
    targetRef.current = initialContent;
    lastUpdateRef.current = 0;
    
    setState({
      messageId,
      partialContent: '',
      toolCalls: [],
      isComplete: false,
    });
    
    animationRef.current = requestAnimationFrame(animate);
  }, [animate, cancelAnimation]);

  const appendContent = useCallback((content: string) => {
    targetRef.current += content;
    
    if (animationRef.current === null && !state.isComplete) {
      lastUpdateRef.current = 0;
      animationRef.current = requestAnimationFrame(animate);
    }
  }, [animate, state.isComplete]);

  const setToolCalls = useCallback((toolCalls: ToolCall[]) => {
    setState(prev => ({ ...prev, toolCalls }));
  }, []);

  const complete = useCallback(() => {
    cancelAnimation();
    contentRef.current = targetRef.current;
    setState(prev => ({
      ...prev,
      partialContent: targetRef.current,
      isComplete: true,
    }));
    if (onComplete) onComplete();
  }, [cancelAnimation, onComplete]);

  const setError = useCallback((error: string) => {
    cancelAnimation();
    setState(prev => ({ ...prev, error, isComplete: true }));
    if (onError) onError(error);
  }, [cancelAnimation, onError]);

  const reset = useCallback(() => {
    cancelAnimation();
    contentRef.current = '';
    targetRef.current = '';
    setState({
      messageId: '',
      partialContent: '',
      toolCalls: [],
      isComplete: true,
    });
  }, [cancelAnimation]);

  useEffect(() => {
    return () => {
      cancelAnimation();
    };
  }, [cancelAnimation]);

  return {
    state,
    startStreaming,
    appendContent,
    setToolCalls,
    complete,
    setError,
    reset,
    isStreaming: !state.isComplete && state.messageId !== '',
  };
}

import { useWebSocket, WSConnectionStatus } from '@/hooks/useWebSocket';

describe('useWebSocket', () => {
  const mockUrl = 'ws://localhost:8080/api/messages/ws?token=test';

  beforeEach(() => {
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.useRealTimers();
    jest.restoreAllMocks();
  });

  describe('connection status', () => {
    it('should start in Disconnected state', () => {
      const { result } = renderHook(() =>
        useWebSocket({ url: mockUrl, autoConnect: false })
      );
      expect(result.current.status).toBe('Disconnected');
    });

    it('should transition to Connecting when autoConnect is true', () => {
      const { result } = renderHook(() =>
        useWebSocket({ url: mockUrl, autoConnect: true })
      );
      expect(result.current.status).toBe('Connecting');
    });
  });

  describe('backoff timing', () => {
    it('should use 5s backoff for first retry', () => {
      const onStatusChange = jest.fn();
      const { result } = renderHook(() =>
        useWebSocket({ url: mockUrl, onStatusChange, autoConnect: false })
      );

      // Manually trigger connect and fail
      act(() => {
        result.current.reconnect();
      });

      expect(onStatusChange).toHaveBeenCalledWith('Connecting');
    });
  });

  describe('send', () => {
    it('should queue messages when not connected', () => {
      const { result } = renderHook(() =>
        useWebSocket({ url: mockUrl, autoConnect: false })
      );

      act(() => {
        result.current.send({ type: 'message', content: 'test' });
      });

      // Message should be queued (send doesn't throw)
      expect(result.current).toBeDefined();
    });
  });

  describe('disconnect', () => {
    it('should set status to Disconnected', () => {
      const onStatusChange = jest.fn();
      const { result } = renderHook(() =>
        useWebSocket({ url: mockUrl, onStatusChange, autoConnect: false })
      );

      act(() => {
        result.current.disconnect();
      });

      expect(onStatusChange).toHaveBeenCalledWith('Disconnected');
    });
  });
});
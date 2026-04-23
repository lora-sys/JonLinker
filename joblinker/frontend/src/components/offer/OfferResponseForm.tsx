'use client';

import { useState } from 'react';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';

interface OfferResponseFormProps {
  isOpen: boolean;
  onClose: () => void;
  onSubmit: (data: {
    response: 'accept' | 'decline' | 'negotiate';
    counterSalary?: number;
    message?: string;
  }) => Promise<void>;
  offerId: string;
  currentSalary: number;
}

export function OfferResponseForm({
  isOpen,
  onClose,
  onSubmit,
  offerId,
  currentSalary,
}: OfferResponseFormProps) {
  const [response, setResponse] = useState<'accept' | 'decline' | 'negotiate' | null>(null);
  const [counterSalary, setCounterSalary] = useState(currentSalary);
  const [message, setMessage] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!response) return;

    setLoading(true);
    try {
      await onSubmit({
        response,
        counterSalary: response === 'negotiate' ? counterSalary : undefined,
        message: response === 'negotiate' ? message : undefined,
      });
      onClose();
    } catch (err) {
      console.error('Failed to submit response:', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title="Respond to Offer"
      size="md"
    >
      <form onSubmit={handleSubmit} className="space-y-6">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-3">
            Your Response
          </label>
          <div className="space-y-2">
            <button
              type="button"
              onClick={() => setResponse('accept')}
              className={`w-full p-4 rounded-lg border-2 text-left transition-colors ${
                response === 'accept'
                  ? 'border-green-500 bg-green-50'
                  : 'border-gray-200 hover:border-gray-300'
              }`}
            >
              <div className="flex items-center gap-3">
                <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center ${
                  response === 'accept' ? 'border-green-500 bg-green-500' : 'border-gray-300'
                }`}>
                  {response === 'accept' && (
                    <svg className="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                    </svg>
                  )}
                </div>
                <div>
                  <p className="font-medium text-gray-900">Accept Offer</p>
                  <p className="text-sm text-gray-500">Accept the compensation package as-is</p>
                </div>
              </div>
            </button>

            <button
              type="button"
              onClick={() => setResponse('negotiate')}
              className={`w-full p-4 rounded-lg border-2 text-left transition-colors ${
                response === 'negotiate'
                  ? 'border-blue-500 bg-blue-50'
                  : 'border-gray-200 hover:border-gray-300'
              }`}
            >
              <div className="flex items-center gap-3">
                <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center ${
                  response === 'negotiate' ? 'border-blue-500 bg-blue-500' : 'border-gray-300'
                }`}>
                  {response === 'negotiate' && (
                    <svg className="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                    </svg>
                  )}
                </div>
                <div>
                  <p className="font-medium text-gray-900">Negotiate</p>
                  <p className="text-sm text-gray-500">Request a higher salary or better terms</p>
                </div>
              </div>
            </button>

            <button
              type="button"
              onClick={() => setResponse('decline')}
              className={`w-full p-4 rounded-lg border-2 text-left transition-colors ${
                response === 'decline'
                  ? 'border-red-500 bg-red-50'
                  : 'border-gray-200 hover:border-gray-300'
              }`}
            >
              <div className="flex items-center gap-3">
                <div className={`w-5 h-5 rounded-full border-2 flex items-center justify-center ${
                  response === 'decline' ? 'border-red-500 bg-red-500' : 'border-gray-300'
                }`}>
                  {response === 'decline' && (
                    <svg className="w-3 h-3 text-white" fill="currentColor" viewBox="0 0 20 20">
                      <path fillRule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clipRule="evenodd" />
                    </svg>
                  )}
                </div>
                <div>
                  <p className="font-medium text-gray-900">Decline</p>
                  <p className="text-sm text-gray-500">Decline this offer politely</p>
                </div>
              </div>
            </button>
          </div>
        </div>

        {response === 'negotiate' && (
          <>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Counter Offer Salary
              </label>
              <div className="relative">
                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500">$</span>
                <input
                  type="number"
                  value={counterSalary}
                  onChange={(e) => setCounterSalary(Number(e.target.value))}
                  className="w-full pl-8 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  min={0}
                  step={1000}
                />
              </div>
              <p className="mt-1 text-sm text-gray-500">
                Current offer: ${currentSalary.toLocaleString()}
              </p>
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Message (optional)
              </label>
              <textarea
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                rows={3}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                placeholder="Explain your reasoning for the counter offer..."
              />
            </div>
          </>
        )}

        <div className="flex justify-end gap-3 pt-4 border-t">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" isLoading={loading} disabled={!response}>
            Submit Response
          </Button>
        </div>
      </form>
    </Modal>
  );
}

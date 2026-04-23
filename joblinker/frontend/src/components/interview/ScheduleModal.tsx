'use client';

import { useState } from 'react';
import { Modal } from '@/components/ui/Modal';
import { Button } from '@/components/ui/Button';

interface ScheduleModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSchedule: (data: {
    scheduledAt: string;
    format: string;
    location: string;
  }) => Promise<void>;
  matchId: string;
  defaultValues?: {
    scheduledAt?: string;
    format?: string;
    location?: string;
  };
  mode?: 'schedule' | 'reschedule';
}

export function ScheduleModal({
  isOpen,
  onClose,
  onSchedule,
  matchId,
  defaultValues,
  mode = 'schedule',
}: ScheduleModalProps) {
  const [loading, setLoading] = useState(false);
  const [formData, setFormData] = useState({
    scheduledAt: defaultValues?.scheduledAt || '',
    format: defaultValues?.format || 'video',
    location: defaultValues?.location || '',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      await onSchedule(formData);
      onClose();
    } catch (err) {
      console.error('Failed to schedule interview:', err);
    } finally {
      setLoading(false);
    }
  };

  const timeSlots = [
    '09:00', '10:00', '11:00', '12:00', '13:00', '14:00', '15:00', '16:00', '17:00',
  ];

  const formats = [
    { value: 'video', label: 'Video Call', icon: '📹' },
    { value: 'phone', label: 'Phone Call', icon: '📞' },
    { value: 'onsite', label: 'On-site', icon: '🏢' },
  ];

  const tomorrow = new Date();
  tomorrow.setDate(tomorrow.getDate() + 1);
  const minDate = tomorrow.toISOString().split('T')[0];

  return (
    <Modal
      isOpen={isOpen}
      onClose={onClose}
      title={mode === 'reschedule' ? 'Reschedule Interview' : 'Schedule Interview'}
      size="lg"
    >
      <form onSubmit={handleSubmit} className="space-y-6">
        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Date
          </label>
          <input
            type="date"
            required
            min={minDate}
            value={formData.scheduledAt.split('T')[0]}
            onChange={(e) =>
              setFormData((prev) => ({
                ...prev,
                scheduledAt: e.target.value ? `${e.target.value}T${prev.scheduledAt.split('T')[1] || '10:00'}` : '',
              }))
            }
            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
          />
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Time
          </label>
          <div className="grid grid-cols-3 gap-2">
            {timeSlots.map((slot) => (
              <button
                key={slot}
                type="button"
                onClick={() =>
                  setFormData((prev) => ({
                    ...prev,
                    scheduledAt: prev.scheduledAt
                      ? `${prev.scheduledAt.split('T')[0]}T${slot}:00`
                      : `${new Date().toISOString().split('T')[0]}T${slot}:00`,
                  }))
                }
                className={`px-3 py-2 text-sm rounded-lg border transition-colors ${
                  formData.scheduledAt.includes(slot)
                    ? 'border-blue-500 bg-blue-50 text-blue-700'
                    : 'border-gray-300 hover:border-gray-400'
                }`}
              >
                {slot}
              </button>
            ))}
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Format
          </label>
          <div className="grid grid-cols-3 gap-3">
            {formats.map((fmt) => (
              <button
                key={fmt.value}
                type="button"
                onClick={() => setFormData((prev) => ({ ...prev, format: fmt.value }))}
                className={`p-3 rounded-lg border-2 text-center transition-colors ${
                  formData.format === fmt.value
                    ? 'border-blue-500 bg-blue-50'
                    : 'border-gray-200 hover:border-gray-300'
                }`}
              >
                <span className="text-xl">{fmt.icon}</span>
                <p className="text-sm font-medium mt-1">{fmt.label}</p>
              </button>
            ))}
          </div>
        </div>

        {formData.format === 'onsite' && (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Location
            </label>
            <input
              type="text"
              required
              value={formData.location}
              onChange={(e) =>
                setFormData((prev) => ({ ...prev, location: e.target.value }))
              }
              placeholder="Enter office address"
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
          </div>
        )}

        {formData.format === 'video' && (
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Meeting Link (optional)
            </label>
            <input
              type="url"
              value={formData.location}
              onChange={(e) =>
                setFormData((prev) => ({ ...prev, location: e.target.value }))
              }
              placeholder="https://meet.example.com/..."
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
            />
          </div>
        )}

        <div className="flex justify-end gap-3 pt-4 border-t">
          <Button type="button" variant="secondary" onClick={onClose}>
            Cancel
          </Button>
          <Button type="submit" isLoading={loading}>
            {mode === 'reschedule' ? 'Reschedule' : 'Schedule'}
          </Button>
        </div>
      </form>
    </Modal>
  );
}

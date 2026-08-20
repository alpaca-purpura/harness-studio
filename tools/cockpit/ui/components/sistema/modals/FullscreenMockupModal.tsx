'use client';

import { Modal } from '@/components/ui/Modal';

interface FullscreenMockupModalProps {
  open: boolean;
  onClose: () => void;
  /** URL del HTML/imagen a renderizar */
  src: string;
  title?: string;
}

export function FullscreenMockupModal({
  open,
  onClose,
  src,
  title,
}: FullscreenMockupModalProps) {
  return (
    <Modal
      open={open}
      onClose={onClose}
      size="full"
      title={title ?? 'Mockup preview'}
    >
      <div className="w-full" style={{ height: '85vh' }}>
        <iframe
          src={src}
          title={title ?? 'mockup'}
          className="w-full h-full border border-[var(--color-border)] rounded bg-white"
        />
      </div>
    </Modal>
  );
}

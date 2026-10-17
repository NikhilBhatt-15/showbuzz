import { useEffect } from 'react';
import { createPortal } from 'react-dom';
import styled, { css, keyframes } from 'styled-components';
import { useDispatch, useSelector } from 'react-redux';
import { dropToast } from '../store/ui.js';
import { CloseIcon } from './icons.jsx';
import { useMounted } from '../lib/hooks.js';

export const Container = styled.div`
  width: 100%;
  max-width: ${(p) => p.theme.container};
  margin: 0 auto;
  padding: 0 16px;
  @media (min-width: 1272px) {
    padding: 0;
  }
`;

export const Button = styled.button`
  border: none;
  border-radius: ${(p) => (p.$size === 'sm' ? '4px' : '8px')};
  background: ${(p) => (p.$variant === 'outline' ? '#fff' : p.$variant === 'ghost' ? 'transparent' : p.theme.brand)};
  color: ${(p) => (p.$variant === 'outline' || p.$variant === 'ghost' ? p.theme.brand : '#fff')};
  ${(p) => p.$variant === 'outline' && css`border: 1px solid ${p.theme.brand};`}
  font-weight: 500;
  font-size: ${(p) => (p.$size === 'sm' ? '12px' : p.$size === 'lg' ? '18px' : '14px')};
  padding: ${(p) => (p.$size === 'sm' ? '6px 14px' : p.$size === 'lg' ? '14px 40px' : '10px 24px')};
  width: ${(p) => (p.$block ? '100%' : 'auto')};
  transition: background 0.15s, opacity 0.15s;
  &:hover:not(:disabled) {
    background: ${(p) => (p.$variant ? p.theme.brandSoft : p.theme.brandDark)};
  }
  &:disabled {
    opacity: 0.45;
  }
`;

export const Chip = styled.button`
  border: 1px solid ${(p) => (p.$active ? p.theme.brand : p.theme.border)};
  background: ${(p) => (p.$active ? p.theme.brand : '#fff')};
  color: ${(p) => (p.$active ? '#fff' : p.theme.brandDark)};
  border-radius: ${(p) => (p.$pill ? '30px' : '4px')};
  padding: ${(p) => (p.$pill ? '8px 12px' : '6px 12px')};
  font-size: 14px;
  line-height: 18px;
  white-space: nowrap;
  &:hover {
    border-color: ${(p) => p.theme.brand};
  }
`;

const spin = keyframes`to { transform: rotate(360deg); }`;
export const Spinner = styled.div`
  width: ${(p) => p.$size || 32}px;
  height: ${(p) => p.$size || 32}px;
  border: 3px solid ${(p) => p.theme.brandSoft};
  border-top-color: ${(p) => p.theme.brand};
  border-radius: 50%;
  animation: ${spin} 0.8s linear infinite;
  margin: ${(p) => (p.$inline ? '0' : '80px auto')};
`;

const shimmer = keyframes`0% { background-position: -400px 0; } 100% { background-position: 400px 0; }`;
export const Skeleton = styled.div`
  background: linear-gradient(90deg, #eee 0px, #f7f7f7 120px, #eee 240px);
  background-size: 800px 100%;
  animation: ${shimmer} 1.4s linear infinite;
  border-radius: ${(p) => p.$radius || '8px'};
  height: ${(p) => p.$h || '20px'};
  width: ${(p) => p.$w || '100%'};
`;

const Backdrop = styled.div`
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.55);
  z-index: 1000;
  display: flex;
  align-items: ${(p) => (p.$top ? 'flex-start' : 'center')};
  justify-content: center;
  padding: ${(p) => (p.$top ? '64px 16px 16px' : '16px')};
  overflow-y: auto;
`;

const Dialog = styled.div`
  background: #fff;
  border-radius: ${(p) => p.theme.radius};
  width: 100%;
  max-width: ${(p) => p.$width || 480}px;
  position: relative;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.25);
`;

const CloseBtn = styled.button`
  position: absolute;
  top: 12px;
  right: 12px;
  background: none;
  border: none;
  color: #666;
  padding: 4px;
  z-index: 2;
`;

export function Modal({ open, onClose, width, top, children, closable = true, label }) {
  const mounted = useMounted();
  useEffect(() => {
    if (!open) return undefined;
    const onKey = (e) => e.key === 'Escape' && closable && onClose?.();
    document.addEventListener('keydown', onKey);
    const prev = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => {
      document.removeEventListener('keydown', onKey);
      document.body.style.overflow = prev;
    };
  }, [open, onClose, closable]);
  if (!open || !mounted) return null;
  return createPortal(
    <Backdrop $top={top} onMouseDown={(e) => e.target === e.currentTarget && closable && onClose?.()}>
      <Dialog $width={width} role="dialog" aria-modal="true" aria-label={label}>
        {closable && (
          <CloseBtn onClick={onClose} aria-label="Close">
            <CloseIcon />
          </CloseBtn>
        )}
        {children}
      </Dialog>
    </Backdrop>,
    document.body,
  );
}

const EmptyWrap = styled.div`
  text-align: center;
  padding: 64px 16px;
  img {
    margin: 0 auto 24px;
    width: 200px;
  }
  h3 {
    font-size: 18px;
    font-weight: 500;
    margin-bottom: 8px;
  }
  p {
    color: ${(p) => p.theme.muted};
    max-width: 320px;
    margin: 0 auto 24px;
    line-height: 1.5;
  }
`;

export function EmptyState({ title, message, cta, onCta, art = 'popcorn' }) {
  return (
    <EmptyWrap>
      <EmptyArt kind={art} />
      <h3>{title}</h3>
      {message && <p>{message}</p>}
      {cta && <Button onClick={onCta}>{cta}</Button>}
    </EmptyWrap>
  );
}

function EmptyArt({ kind }) {
  const emoji = { popcorn: '🍿', error: '🛠️', ticket: '🎟️', search: '🔍' }[kind] || '🍿';
  return (
    <div style={{ width: 120, height: 120, borderRadius: '50%', background: '#FDEEF1', margin: '0 auto 24px', display: 'grid', placeItems: 'center', fontSize: 56 }}>
      {emoji}
    </div>
  );
}

export function ErrorState({ onRetry }) {
  return (
    <EmptyState art="error" title="Sorry for bug-ging" message="Some issue from our side. Kindly refresh the page for more entertainment." cta="Refresh" onCta={onRetry || (() => window.location.reload())} />
  );
}

const ToastStack = styled.div`
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 2000;
  display: flex;
  flex-direction: column;
  gap: 8px;
  align-items: center;
`;

const ToastItem = styled.div`
  background: ${(p) => (p.$kind === 'error' ? '#D32F2F' : p.$kind === 'success' ? p.theme.green : '#333')};
  color: #fff;
  padding: 12px 20px;
  border-radius: 8px;
  font-size: 14px;
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.25);
  max-width: 480px;
`;

function Toast({ toast }) {
  const dispatch = useDispatch();
  useEffect(() => {
    const id = setTimeout(() => dispatch(dropToast(toast.id)), 4500);
    return () => clearTimeout(id);
  }, [dispatch, toast.id]);
  return <ToastItem $kind={toast.kind} role="status">{toast.message}</ToastItem>;
}

export function Toasts() {
  const toasts = useSelector((s) => s.ui.toasts);
  return (
    <ToastStack>
      {toasts.map((t) => (
        <Toast key={t.id} toast={t} />
      ))}
    </ToastStack>
  );
}

export const SectionTitle = styled.h2`
  font-size: 24px;
  font-weight: 700;
  color: ${(p) => (p.$dark ? '#fff' : p.theme.text)};
`;

export const Card = styled.div`
  background: #fff;
  border-radius: ${(p) => p.theme.radius};
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
`;

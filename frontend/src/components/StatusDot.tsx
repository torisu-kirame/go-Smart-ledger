export function StatusDot({ ok, label }: { ok: boolean; label: string }) {
  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', gap: '0.4rem' }}>
      <span
        style={{
          width: 8,
          height: 8,
          borderRadius: '50%',
          background: ok ? 'var(--success)' : 'var(--danger)',
          boxShadow: ok ? '0 0 8px var(--success)' : 'none',
        }}
      />
      {label}
    </span>
  )
}

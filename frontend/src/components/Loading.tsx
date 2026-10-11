interface LoadingProps {
  size?: 'sm' | 'md' | 'lg';
  text?: string;
}

export function Loading({ size = 'md', text }: LoadingProps) {
  return (
    <div className="loading" role="status" aria-live="polite">
      <div
        className="spinner"
        style={{
          width: size === 'sm' ? 16 : size === 'md' ? 24 : 48,
          height: size === 'sm' ? 16 : size === 'md' ? 24 : 48,
          borderWidth: size === 'sm' ? 2 : size === 'md' ? 3 : 4,
        }}
      />
      {text && <span>{text}</span>}
    </div>
  );
}
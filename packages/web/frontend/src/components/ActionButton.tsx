import type { ReactNode } from "react";

export type ActionIntent = "default" | "primary" | "accent" | "warn" | "danger";

interface Props {
  label: string;
  icon?: string;
  intent?: ActionIntent;
  disabled?: boolean;
  onClick: () => void;
}

export function ActionButton({
  label,
  icon,
  intent = "default",
  disabled = false,
  onClick,
}: Props): ReactNode {
  const display = icon ?? label.slice(0, 1);
  return (
    <button
      type="button"
      className={`action-btn ${intent}`}
      disabled={disabled}
      onClick={onClick}
    >
      <span className="action-icon" aria-hidden="true">
        {display}
      </span>
      <span>{label}</span>
    </button>
  );
}

export function ActionBar({ children }: { children: ReactNode }): ReactNode {
  return <div className="actions">{children}</div>;
}

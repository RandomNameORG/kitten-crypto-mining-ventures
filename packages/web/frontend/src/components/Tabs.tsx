import { useEffect } from "react";
import type { TabId } from "../types";

const TABS: ReadonlyArray<readonly [TabId, string, string, string]> = [
  ["store", "商店", "S", "1"],
  ["rooms", "房间", "R", "2"],
  ["gpus", "显卡", "G", "3"],
  ["defense", "防御", "D", "4"],
  ["skills", "技能", "T", "5"],
  ["mercs", "雇佣", "H", "6"],
  ["log", "日志", "L", "7"],
  ["stats", "状态", "I", "8"],
];

interface Props {
  active: TabId;
  onSelect: (id: TabId) => void;
}

export function Tabs({ active, onSelect }: Props) {
  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) return;
      if (e.metaKey || e.ctrlKey || e.altKey) return;
      const idx = Number(e.key);
      if (idx >= 1 && idx <= TABS.length) {
        onSelect(TABS[idx - 1][0]);
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [onSelect]);

  return (
    <nav className="tabs">
      {TABS.map(([id, label, icon, key]) => (
        <button
          key={id}
          type="button"
          className={id === active ? "active" : ""}
          title={`${label}  (${key})`}
          onClick={() => onSelect(id)}
        >
          <span className="tab-icon" aria-hidden="true">{icon}</span>
          <span className="tab-label">{label}</span>
          <kbd className="kbd">{key}</kbd>
        </button>
      ))}
    </nav>
  );
}

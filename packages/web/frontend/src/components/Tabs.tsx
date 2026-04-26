import type { TabId } from "../types";

const TABS: ReadonlyArray<readonly [TabId, string, string]> = [
  ["store", "商店", "S"],
  ["rooms", "房间", "R"],
  ["gpus", "显卡", "G"],
  ["defense", "防御", "D"],
  ["skills", "技能", "T"],
  ["mercs", "雇佣", "H"],
  ["log", "日志", "L"],
  ["stats", "状态", "I"],
];

interface Props {
  active: TabId;
  onSelect: (id: TabId) => void;
}

export function Tabs({ active, onSelect }: Props) {
  return (
    <nav className="tabs">
      {TABS.map(([id, label, icon]) => (
        <button
          key={id}
          type="button"
          className={id === active ? "active" : ""}
          title={label}
          onClick={() => onSelect(id)}
        >
          <span className="tab-icon" aria-hidden="true">
            {icon}
          </span>
          <span className="tab-label">{label}</span>
        </button>
      ))}
    </nav>
  );
}

import { ActionBar, ActionButton } from "../components/ActionButton";
import type { ActionRequest, Snapshot } from "../types";
import { defenseIconSrc } from "../util";

interface Props {
  snapshot: Snapshot;
  dispatch: (payload: ActionRequest) => void;
}

const DIMS: ReadonlyArray<readonly [keyof Snapshot["rooms"][number]["defense"], string]> = [
  ["lock", "门锁"],
  ["cctv", "监控"],
  ["wiring", "布线"],
  ["cooling", "散热"],
  ["armor", "装甲"],
];

export function DefensePanel({ snapshot, dispatch }: Props) {
  const room = snapshot.rooms.find((r) => r.id === snapshot.state.current_room);
  if (!room) return null;
  const d = room.defense || { lock: 0, cctv: 0, wiring: 0, cooling: 0, armor: 0 };
  return (
    <>
      <div className="list">
        {DIMS.map(([id, label]) => {
          const level = d[id] || 0;
          const maxed = level >= 8;
          return (
            <article key={id} className="row item-row">
              <img className="item-icon" src={defenseIconSrc(id)} alt="" loading="lazy" />
              <div className="item-content">
                <div className="row-head">
                  <span className="row-title">{label}</span>
                  <span className="tag">L{level}</span>
                </div>
                <div className="facts">
                  <span className="fact">cost {(level + 1) * 250}</span>
                  <span className="fact">max 8</span>
                </div>
                <ActionBar>
                  <ActionButton
                    label={maxed ? "已满级" : "升级"}
                    icon="升"
                    intent="primary"
                    disabled={maxed}
                    onClick={() => dispatch({ action: "upgrade_defense", dim: id })}
                  />
                </ActionBar>
              </div>
            </article>
          );
        })}
      </div>
    </>
  );
}

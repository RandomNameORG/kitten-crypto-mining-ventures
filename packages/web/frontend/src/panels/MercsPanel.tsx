import { ActionBar, ActionButton } from "../components/ActionButton";
import { PanelSummary } from "../components/PanelSummary";
import type { ActionRequest, Snapshot } from "../types";

interface Props {
  snapshot: Snapshot;
  currentRoomName: string;
  dispatch: (payload: ActionRequest) => void;
}

export function MercsPanel({ snapshot, currentRoomName, dispatch }: Props) {
  const roomName = (id: string) =>
    snapshot.rooms.find((r) => r.id === id)?.name ?? id;

  return (
    <>
      <h2>雇佣猫</h2>
      <PanelSummary
        items={[
          ["已雇佣", `${snapshot.mercs.length}`],
          ["当前房间", currentRoomName],
        ]}
      />
      <div className="list">
        {snapshot.mercs.length === 0 ? (
          <div className="empty">暂无雇佣</div>
        ) : (
          snapshot.mercs.map((merc) => (
            <article key={merc.instance_id} className="row">
              <div className="row-head">
                <span className="row-title">
                  #{merc.instance_id} {merc.name}
                </span>
                <span className="tag">{merc.loyalty}</span>
              </div>
              <div className="facts">
                <span className="fact">{roomName(merc.room_id)}</span>
              </div>
              <ActionBar>
                <ActionButton
                  label="打赏"
                  icon="赏"
                  intent="accent"
                  onClick={() => dispatch({ action: "bribe_merc", instance_id: merc.instance_id })}
                />
                <ActionButton
                  label="解雇"
                  icon="离"
                  intent="danger"
                  onClick={() => dispatch({ action: "fire_merc", instance_id: merc.instance_id })}
                />
              </ActionBar>
            </article>
          ))
        )}
        {snapshot.merc_defs.map((def) => (
          <article key={def.id} className="row">
            <div className="row-head">
              <span className="row-title">{def.name}</span>
              <span className="tag">{def.hire_cost_fmt}</span>
            </div>
            <div className="copy">{def.flavor}</div>
            <div className="facts">
              <span className="fact">{def.specialty}</span>
              <span className="fact">wage {def.wage_fmt}</span>
            </div>
            <ActionBar>
              <ActionButton
                label="雇佣"
                icon="雇"
                intent="primary"
                onClick={() => dispatch({ action: "hire_merc", id: def.id })}
              />
            </ActionBar>
          </article>
        ))}
      </div>
    </>
  );
}

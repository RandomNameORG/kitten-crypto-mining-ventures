import { ActionBar, ActionButton } from "../components/ActionButton";
import { PanelSummary } from "../components/PanelSummary";
import type { ActionRequest, Snapshot } from "../types";
import { gpuIconSrc } from "../util";

interface Props {
  snapshot: Snapshot;
  dispatch: (payload: ActionRequest) => void;
}

export function GPUsPanel({ snapshot, dispatch }: Props) {
  const room = snapshot.rooms.find((r) => r.id === snapshot.state.current_room) || null;
  const gpus = snapshot.gpus.filter((g) => g.room === snapshot.state.current_room);
  if (!gpus.length) {
    return (
      <>
        <h2>当前房间显卡</h2>
        <PanelSummary
          items={[
            ["槽位", room ? `0/${room.slots}` : "0"],
            ["提示", "去商店购买"],
          ]}
        />
        <div className="empty">空槽位</div>
      </>
    );
  }
  const broken = gpus.filter((g) => g.status === "broken").length;
  return (
    <>
      <h2>当前房间显卡</h2>
      <PanelSummary
        items={[
          ["槽位", room ? `${room.gpu_count}/${room.slots}` : `${gpus.length}`],
          ["损坏", `${broken}`],
        ]}
      />
      <div className="list">
        {gpus.map((gpu) => (
          <article key={gpu.instance_id} className="row item-row">
            <img
              className="item-icon gpu-icon"
              src={gpuIconSrc(gpu.def_id || "scrap")}
              alt=""
              loading="lazy"
            />
            <div className="item-content">
              <div className="row-head">
                <span className="row-title">
                  #{gpu.instance_id} {gpu.name}
                </span>
                <span className="tag">{gpu.status}</span>
              </div>
              <div className="facts">
                <span className="fact">L{gpu.upgrade}</span>
                <span className="fact">OC {gpu.oc_level}</span>
                <span className="fact">{gpu.earn_fmt}</span>
                <span className="fact">{gpu.hours_left.toFixed(1)}h</span>
              </div>
              <ActionBar>
                <ActionButton
                  label="升级"
                  icon="升"
                  intent="primary"
                  onClick={() => dispatch({ action: "upgrade_gpu", instance_id: gpu.instance_id })}
                />
                <ActionButton
                  label="超频"
                  icon="频"
                  intent="accent"
                  onClick={() => dispatch({ action: "cycle_oc", instance_id: gpu.instance_id })}
                />
                <ActionButton
                  label={gpu.repairable ? "维修" : "正常"}
                  icon="修"
                  intent="warn"
                  disabled={!gpu.repairable}
                  onClick={() => dispatch({ action: "repair_gpu", instance_id: gpu.instance_id })}
                />
                <ActionButton
                  label="拆解"
                  icon="拆"
                  intent="danger"
                  onClick={() => dispatch({ action: "scrap_gpu", instance_id: gpu.instance_id })}
                />
              </ActionBar>
            </div>
          </article>
        ))}
      </div>
    </>
  );
}

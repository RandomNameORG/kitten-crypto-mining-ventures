import { ActionBar, ActionButton } from "../components/ActionButton";
import type { ActionRequest, Snapshot } from "../types";

interface Props {
  snapshot: Snapshot;
  dispatch: (payload: ActionRequest) => void;
}

export function SkillsPanel({ snapshot, dispatch }: Props) {
  const visible = snapshot.skills.slice(0, 18);
  return (
    <>
      <div className="flex gap-2 mb-2 text-[11px] text-muted">
        <span><span className="text-gold">TP</span> {snapshot.state.tech_point}</span>
        <span><span className="text-gold">碎片</span> {snapshot.state.research_frags}</span>
      </div>
      <div className="list">
        {visible.map((skill) => {
          const prereqOk =
            !skill.prereq ||
            snapshot.skills.find((item) => item.id === skill.prereq)?.unlocked;
          const canBuy =
            !skill.unlocked && !!prereqOk && snapshot.state.tech_point >= skill.cost;
          return (
            <article key={skill.id} className="row">
              <div className="row-head">
                <span className="row-title">{skill.name}</span>
                <span className="tag">{skill.cost} TP</span>
              </div>
              <div className="copy">{skill.desc}</div>
              <div className="facts">
                <span className="fact">{skill.lane}</span>
                <span className="fact">
                  {skill.unlocked ? "已学会" : prereqOk ? "可研究" : "前置未解"}
                </span>
              </div>
              <ActionBar>
                <ActionButton
                  label={skill.unlocked ? "已研究" : "研究"}
                  icon="研"
                  intent="primary"
                  disabled={!canBuy}
                  onClick={() => dispatch({ action: "unlock_skill", id: skill.id })}
                />
              </ActionBar>
            </article>
          );
        })}
      </div>
    </>
  );
}

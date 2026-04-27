# Room-view IA — 30-minute Sketch

Branch: `feat-2d-assets` · Date: 2026-04-26 · Trigger: autoplan premise-gate Q2

## Today: 8 tabs

```
[store]  [rooms]  [gpus]  [defense]  [skills]  [mercs]  [log]  [stats]
```

The player wants to "manage one room." That requires hopping across `gpus → defense → mercs`. Heat is shown on the stage foot bar, not in any tab. Skills are global tech tree (not room-scoped, but unlocking GPU L4 affects rigs in the active room).

## Proposed: 6 tabs

```
[store]  [rooms]  [房间]  [skills]  [log]  [stats]
                    ▲
                    └── new: collapses gpus + defense + mercs-in-room
```

`[房间]` (Room) becomes a single page with the active room as the spatial unit. Inside it:

```
┌─ Room: 阴湿地下室 ────── 6/8 机位 · 4 运行 · 1 运输 · 1 损坏 ─────┐
│                                                                  │
│  ┌─ 显卡机架 (gpus) ────────────────────────────────────────┐    │
│  │   [filter: 全部 / 运行 / 损坏]                            │    │
│  │                                                          │    │
│  │  🐈──────────●▾  在途 1 · ETA 47s                        │    │
│  │                                                          │    │
│  │  [#3 RTX-4090] [L2] [OC1] [+₿0.012/s] [12.4h]    升级 ▸  │    │
│  │  [#7 GTX-1060] [L0] [OC0] [+₿0.001/s] [3.1h]     拆解 ▸  │    │
│  │  [#8 RX-580]  ⚠损坏  [可维修]                    维修 ▸  │    │
│  │  ...                                                     │    │
│  └──────────────────────────────────────────────────────────┘    │
│                                                                  │
│  ┌─ 防御 (defense) ────────────────────────┐  ┌─ 雇佣猫 ────┐    │
│  │  锁 Lv2  ▸                              │  │ 阿橘 (修)  │    │
│  │  监控 Lv1 ▸                             │  │ 大白 (防)  │    │
│  │  线路 Lv3 ▸ (热量上限 +20%)              │  │ + 雇佣 ▸    │    │
│  │  冷却 Lv2 ▸ (散热 +15%)                  │  │             │    │
│  │  装甲 Lv0 ▸                             │  │             │    │
│  │                                         │  │             │    │
│  │  当前热量 ▮▮▮▮▮▮░░░░ 62/100             │  │             │    │
│  │  收支 +₿0.013/s  电费 -₿0.002/s         │  │             │    │
│  └─────────────────────────────────────────┘  └─────────────┘    │
└──────────────────────────────────────────────────────────────────┘
```

Layout: GPU rack on top (vertical scroll), Defense + Mercs side-by-side at bottom (3:1 split on desktop, stacked on mobile). Single sticky header shows room title + slot stats. Heat moves out of stage foot and into Defense card where it belongs (heat is a room property managed by cooling).

## Obvious-win checklist (3 bullets)

- ☑ **Eliminates tab-hop?** YES. The player's "upgrade my rig's cooling" loop becomes scroll-down instead of tab-switch. The stage foot heat bar still works as a glance, but the actionable controls are in one place.
- ☐ **Implementable in similar budget as Approach B (~90 min CC)?** NO. Cost analysis:
  - New `RoomTab.tsx` panel (~150 LOC)
  - Move `DefensePanel` rendering inside it (existing component, not deletion)
  - Move `MercsPanel` filtered to current room
  - Update `Tabs.tsx` to drop `gpus`/`defense`/`mercs` and add `room`
  - Migration: `App.tsx` `useState<TabId>("store")` defaults still work; old saves' deep links to `?tab=gpus` need a redirect
  - Estimate: **~3-4 hr CC**, double Approach B
- ☑ **Avoids breaking existing flows?** YES. All actions still exist; routing is internal layout only.

## Decision: NOT an obvious win

Score: 2/3. Eliminates tab-hop (the strongest argument), but doubles the effort budget on a polish pass that was already approved at 90 min. The IA insight is correct, but it's a separate ~4-hour plan, not a swap-in for B.

## Recommendation

**Ship Approach B (sticky ship strip + sort/filter + buy-CTA slot integration) NOW.** Open a follow-up plan `web-room-view-ia.md` for the Room-view consolidation. Approach B's work is fully portable: the sort/filter/affordability logic moves into RoomTab unchanged; the SlotMeter component just gets a different parent; tier.ts is shared.

Sticky ship strip is *not wasted* — when the Room-view ships, the strip lives at the top of the Room tab in the same form.

## What changes in the autoplan plan

- Approach B proceeds as specified.
- Add to TODOS.md: "Room-view IA refactor — collapse gpus/defense/mercs into single Room tab. Estimated 3-4hr CC. Trigger: post web-ui-polish merge."
- No changes to the 17 auto-decisions or the 1 taste decision.

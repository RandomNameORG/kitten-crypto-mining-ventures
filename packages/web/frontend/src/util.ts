export const GPU_ICON_ROOT = "/assets/2d/items/gpu_cards";
export const DEFENSE_ICON_ROOT = "/assets/2d/items/defense";
export const SKILL_ICON_ROOT = "/assets/2d/items/skills";

export function gpuIconSrc(id: string | undefined): string {
  return `${GPU_ICON_ROOT}/${encodeURIComponent(id || "scrap")}.png`;
}

export function defenseIconSrc(id: string | undefined): string {
  return `${DEFENSE_ICON_ROOT}/${encodeURIComponent(id || "lock")}.png`;
}

export function skillIconSrc(id: string | undefined): string {
  const key = skillIconKey(id || "");
  return `${SKILL_ICON_ROOT}/${encodeURIComponent(key)}.png`;
}

function skillIconKey(id: string): string {
  if (id.startsWith("undervolt") || id.startsWith("neighbor_leech")) return "undervolt";
  if (id.startsWith("overclock") || id.startsWith("pump_dump")) return "overclock";
  if (id.startsWith("pcb_surgery")) return "pcb_repair";
  if (id.startsWith("auto_repair")) return "auto_repair";
  if (id === "rd_unlock") return "meowcore";
  if (id.startsWith("smart_invoicing")) return "smart_invoicing";
  if (id.startsWith("tax_opt")) return "tax_optimization";
  if (id.startsWith("hedged_wallet") || id === "venture_cap") return "hedged_wallet";
  if (id.startsWith("chain_ghost")) return "chain_ghost";
  return "meowcore";
}

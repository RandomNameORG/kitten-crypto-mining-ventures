export const GPU_ICON_ROOT = "/assets/2d/items/gpu_cards";

export function gpuIconSrc(id: string | undefined): string {
  return `${GPU_ICON_ROOT}/${encodeURIComponent(id || "scrap")}.png`;
}

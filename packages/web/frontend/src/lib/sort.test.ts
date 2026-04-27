import { describe, expect, it } from "vitest";
import { affordableOnly, sortGpuDefs } from "./sort";
import type { GPUDef } from "../types";

function def(id: string, tier: string, efficiency: number, price: number): GPUDef {
  return {
    id,
    name: id,
    flavor: "",
    tier,
    efficiency,
    power_draw: 1,
    heat_output: 1,
    price,
    price_fmt: `${price}`,
    scrap_fmt: `${price / 2}`,
  };
}

describe("sortGpuDefs", () => {
  it("orders legendary > epic > rare > common > trash", () => {
    const out = sortGpuDefs([
      def("a", "common", 0.1, 100),
      def("b", "legendary", 0.4, 5000),
      def("c", "trash", 0.05, 10),
      def("d", "epic", 0.3, 2000),
      def("e", "rare", 0.2, 500),
    ]);
    expect(out.map((d) => d.id)).toEqual(["b", "d", "e", "a", "c"]);
  });

  it("breaks ties within tier by efficiency descending", () => {
    const out = sortGpuDefs([
      def("low", "rare", 0.1, 500),
      def("mid", "rare", 0.2, 600),
      def("hi", "rare", 0.3, 700),
    ]);
    expect(out.map((d) => d.id)).toEqual(["hi", "mid", "low"]);
  });

  it("treats unknown tier as common (defensive — keeps the list visible)", () => {
    const out = sortGpuDefs([
      def("ghost", "weird-future-tier", 0.5, 1),
      def("known", "common", 0.1, 1),
    ]);
    // Both at common rank; tie-break by efficiency.
    expect(out.map((d) => d.id)).toEqual(["ghost", "known"]);
  });

  it("does not mutate the input", () => {
    const input = [def("a", "trash", 0.1, 1), def("b", "epic", 0.2, 2)];
    const before = input.map((d) => d.id);
    sortGpuDefs(input);
    expect(input.map((d) => d.id)).toEqual(before);
  });
});

describe("affordableOnly", () => {
  it("keeps defs at or below current btc", () => {
    const out = affordableOnly(
      [def("cheap", "trash", 0.1, 100), def("mid", "common", 0.2, 1000), def("dear", "epic", 0.3, 50000)],
      1000,
    );
    expect(out.map((d) => d.id)).toEqual(["cheap", "mid"]);
  });

  it("returns empty when nothing affordable (filter UI shows the empty state)", () => {
    const out = affordableOnly([def("dear", "epic", 0.3, 50000)], 100);
    expect(out).toEqual([]);
  });
});

import { describe, expect, it } from "vitest";
import { reviewGuidance } from "../../src/pages/hotspots/review-guidance";

const day = { up: 1000, highs: 20, above: 60, eligible: 300, highEligible: 4500 };
describe("收盘观察规则", () => {
  it("includes all matching conditions and prioritizes scarce new highs", () => {
    const result = reviewGuidance({...day, up: 999, highs: 19, above: 59}, 1);
    expect(result.focus.title).toBe("新高股不足 20 家");
    expect(result.others.map(rule => rule.title)).toEqual(["普跌中仍有新高股增加", "沪深 300 均线上方成份偏少"]);
  });
  it("treats the boundary values as neutral and checks an actual recovery", () => {
    expect(reviewGuidance(day, 0).focus.metric).toBeNull();
    expect(reviewGuidance({...day, above: 210}, 0).focus.metric).toBeNull();
    expect(reviewGuidance({...day, highs: 21}, 1).focus.title).toBe("新高家数重新超过 20");
    expect(reviewGuidance({...day, highs: 22}, 1).focus.metric).toBeNull();
  });
  it("keeps missing samples out of signals", () => {
    const result = reviewGuidance({...day, eligible: 0, highEligible: 0, highs: 0}, null);
    expect(result.focus.metric).toBeNull();
    expect(result.gaps).toContain("前日变化不可比");
  });
});

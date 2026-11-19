import { z } from "zod";

import { RequestGet, RequestPost } from "@/api/request-schema";
import { apiPathsV1 } from "@/api/paths/v1";

export const researchSkillSchema = z.object({
  id: z.string(), name: z.string(), instructions: z.string(), required: z.boolean(), builtIn: z.boolean(),
  defaultSelected: z.boolean().optional(),
});

const sourceSchema = z.object({
  id: z.string(), provider: z.string(), title: z.string(), url: z.string(),
  kind: z.enum(["board", "search", "original"]), date: z.string().optional(), excerpt: z.string().optional(),
});

const candidateSchema = z.object({
  code: z.string(), name: z.string(), status: z.enum(["double", "combined", "single"]),
  topics: z.array(z.string()), sourceIds: z.array(z.string()),
});

const resultSchema = z.object({
  topics: z.array(z.string()), match: z.enum(["all", "any"]),
  candidates: z.array(candidateSchema), sources: z.array(sourceSchema), notes: z.array(z.string()),
  reply: z.string().optional(), conclusion: z.string().optional(),
  skillChecks: z.array(z.object({
    method: z.enum(["orders", "industry", "risks"]),
    findings: z.array(z.object({
      subject: z.string().optional(),
      note: z.string(),
      evidence: z.array(z.object({ sourceId: z.string(), quote: z.string() })),
    })),
  })).optional(),
  observations: z.array(z.object({
    code: z.string(), name: z.string().optional(), note: z.string(), sourceIds: z.array(z.string()), quoteUrl: z.string().optional(),
  })).optional(),
  optimus: z.object({
    stages: z.array(z.string()),
    stocks: z.array(z.object({
      code: z.string(), name: z.string(), stage: z.string(), quoteUrl: z.string(),
      status: z.enum(["confirmed", "sample", "denied", "reported", "unverified"]),
      evidence: z.array(z.object({
        sourceId: z.string(), quote: z.string(), fullText: z.boolean(),
        status: z.enum(["confirmed", "sample", "denied", "reported", "unverified"]),
      })),
    })),
  }).optional(),
});

export const runSchema = z.object({
  id: z.number(), question: z.string(), model: z.string(),
  status: z.enum(["running", "done", "failed"]), created_at: z.string(), updated_at: z.string(),
  referenceId: z.number().optional(), error: z.string().optional(),
  skills: z.array(researchSkillSchema).optional(), result: resultSchema.optional(),
});

export type ResearchSkill = z.infer<typeof researchSkillSchema>;
export type ResearchRun = z.infer<typeof runSchema>;
export type ResearchReport = z.infer<typeof resultSchema>;
export type ResearchCandidate = z.infer<typeof candidateSchema>;
export type ResearchSource = z.infer<typeof sourceSchema>;

export const listResearchSkills = async () =>
  (await RequestGet(`${apiPathsV1.topicResearch}/skills`, { schema: z.array(researchSkillSchema) })).data;

export const saveResearchSkill = async (values: { id?: number; name: string; instructions: string }) =>
  (await RequestPost(`${apiPathsV1.topicResearch}/skills`, values, { schema: researchSkillSchema })).data;

export const listResearchRuns = async (before?: number) =>
  (await RequestGet(`${apiPathsV1.topicResearch}/runs`, { params: { before }, schema: z.array(runSchema) })).data;

export const readResearchRun = async (id: number) =>
  (await RequestGet(`${apiPathsV1.topicResearch}/runs/${id}`, { schema: runSchema })).data;

export const startResearchRun = async (values: { question: string; skillIds: string[]; referenceId?: number }) =>
  (await RequestPost(`${apiPathsV1.topicResearch}/runs`, values, { schema: runSchema })).data;

export function researchError(error: unknown, fallback: string) {
  return typeof error === "object" && error && "message" in error && typeof error.message === "string"
    ? error.message : fallback;
}

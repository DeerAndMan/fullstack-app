import z from "zod";

export const SubscribeItemSchema = z.object({
  id: z.string(),
  user_id: z.string(),
  description: z.string().optional().nullable(),
  enabled: z.boolean(),
  is_starred: z.boolean().default(false),
  sort_order: z.number().int().min(0).max(999999).default(0),
});

export const SubscribeListSchema = z.array(SubscribeItemSchema);

export type SubscribeItem = z.infer<typeof SubscribeItemSchema>;

export const ThemeContentItemSchema = z.object({
  id: z.number(),
  user_id: z.number(),
  screen_name: z.string().optional().nullable(),
  created_at: z.string().optional().nullable(),
  edited_at: z.string().optional().nullable(),
  text: z.string().optional().nullable(),
  meta_keywords: z.string().optional().nullable(),
});

export type ThemeContentItem = z.infer<typeof ThemeContentItemSchema>;

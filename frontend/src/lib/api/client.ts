import createClient from "openapi-fetch";

import type { components, paths } from "@/lib/api/schema";

// 本番は FE と API が同じオリジン（Echo）なので /api。
// ローカルで next dev（3000番）から API（8080番）を呼ぶときは、
// NEXT_PUBLIC_API_BASE_URL=http://localhost:8080/api を指定する。
const baseUrl = process.env.NEXT_PUBLIC_API_BASE_URL ?? "/api";

/** API のクライアント。app/ の中からだけ使う（components/ からは使わない） */
export const api = createClient<paths>({ baseUrl });

type Schemas = components["schemas"];
export type Stall = Schemas["Stall"];
export type StallSummary = Schemas["StallSummary"];
export type Ticket = Schemas["Ticket"];
export type TicketDetail = Schemas["TicketDetail"];
export type ApiError = Schemas["Error"];
export type CommentTemplate = Schemas["CommentTemplate"];

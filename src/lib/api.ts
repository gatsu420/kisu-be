export async function fetchAnswer(
  prompt: string,
  param: string,
  filter: string,
): Promise<unknown> {
  const url = new URL("/answer/v1/answer", window.location.origin);
  url.searchParams.set("prompt", prompt);
  url.searchParams.set("param", param);
  url.searchParams.set("filter", filter);

  const res = await fetch(url.toString(), {
    method: "GET",
    credentials: "include",
  });

  if (res.status === 401) {
    throw new Error("unauthorized");
  }

  if (!res.ok) {
    throw new Error("request failed");
  }

  return res.json();
}

export async function validateToolQuery(query: string): Promise<boolean> {
  const url = new URL("/answer/v1/validate-tool-query", window.location.origin);
  url.searchParams.set("query", query);

  const res = await fetch(url.toString(), {
    method: "GET",
    credentials: "include",
  });

  if (res.status === 401) {
    throw new Error("unauthorized");
  }

  if (!res.ok) {
    throw new Error("verify request failed");
  }

  const data: unknown = await res.json();
  const is_valid =
    typeof data === "object" &&
    data !== null &&
    "is_valid" in data &&
    typeof data.is_valid === "boolean";
  if (!is_valid) {
    throw new Error("verify response is invalid");
  }
  return (data as { is_valid: boolean }).is_valid;
}

export type ToolType = "table" | "query";

// Legacy rows may have an empty type. Treat them as table tools.
export function toolTypeOf(tool: Tool): ToolType {
  return tool.type === "query" ? "query" : "table";
}

export interface AddToolPayload {
  tool_description: string;
  project: string;
  dataset: string;
  table_name: string;
  columns: { name: string; type: string; description: string }[];
  type: ToolType;
  examples: { description: string; query: string }[];
  param_name: string;
  param_type: string;
  param_description: string;
}

export async function addTool(payload: AddToolPayload): Promise<string> {
  const res = await fetch("/answer/v1/tool", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    credentials: "include",
    body: JSON.stringify(payload),
  });

  if (res.status === 401) {
    throw new Error("unauthorized");
  }

  if (!res.ok) {
    throw new Error("request failed");
  }

  return res.text();
}

export interface ToolColumn {
  name: string;
  type: string;
  description: string;
}

export interface ToolQueryExamples {
  description: string;
  query: string;
}

export interface Tool {
  tool_description: string;
  project: string;
  dataset: string;
  table_name: string;
  columns: ToolColumn[];
  type: ToolType;
  examples: ToolQueryExamples[];
  param_name: string;
  param_type: string;
  param_description: string;
}

export async function fetchTool(): Promise<Tool[]> {
  const res = await fetch("/answer/v1/tool", {
    method: "GET",
    credentials: "include",
  });

  if (res.status === 401) {
    throw new Error("unauthorized");
  }

  if (!res.ok) {
    throw new Error("request failed");
  }

  return res.json();
}

export function redirectToLogin() {
  window.location.href = "/auth/v1/get-permission";
}

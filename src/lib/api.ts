export interface AnswerResult {
  answer: unknown;
  stringified_func_calls: string;
}

// Route prompts to a tool. The BE stores the result in an HttpOnly cookie.
export async function routeTool(
  prompt: string,
  paramValue: string,
  paramName: string,
): Promise<void> {
  const url = new URL("/answer/v1/route", window.location.origin);
  url.searchParams.set("prompt", prompt);
  url.searchParams.set("param_value", paramValue);
  url.searchParams.set("param_name", paramName);

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
}

// Call the routed tool. The BE reads the tool from the cookie.
export async function callTool(
  limit: number,
  offset: number,
): Promise<AnswerResult> {
  const url = new URL("/answer/v1/answer", window.location.origin);
  url.searchParams.set("limit", String(limit));
  url.searchParams.set("offset", String(offset));

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

interface ToolTableMetadataArgs {
  type: ToolType;
  project: string;
  dataset: string;
  table_name: string;
  builder_query: string;
}

interface ToolTableMetadata {
  description: string;
  columns: ToolColumn[];
}

export async function fetchToolTableMetadata(
  args: ToolTableMetadataArgs,
): Promise<ToolTableMetadata> {
  const url = new URL(
    "/answer/v1/get-tool-table-metadata",
    window.location.origin,
  );
  url.searchParams.set("type", args.type);
  url.searchParams.set("project", args.project);
  url.searchParams.set("dataset", args.dataset);
  url.searchParams.set("table_name", args.table_name);
  url.searchParams.set("builder_query", args.builder_query);

  const res = await fetch(url.toString(), {
    method: "GET",
    credentials: "include",
  });

  if (res.status === 401) {
    throw new Error("unauthorized");
  }

  if (!res.ok) {
    throw new Error("metadata request failed");
  }

  const data: unknown = await res.json();
  if (typeof data !== "object" || data === null || !("columns" in data)) {
    throw new Error("metadata response is invalid");
  }
  const { columns, description } = data as {
    columns?: unknown;
    description?: unknown;
  };
  if (!Array.isArray(columns)) {
    throw new Error("metadata response is invalid");
  }
  return {
    description: typeof description === "string" ? description : "",
    columns: columns as ToolColumn[],
  };
}

export type ToolType = "table" | "query";

// Legacy rows may have an empty type. Treat them as table tools.
export function toolTypeOf(tool: Tool): ToolType {
  return tool.type === "query" ? "query" : "table";
}

interface AddToolPayload {
  tool_description: string;
  project: string;
  dataset: string;
  table_name: string;
  columns: { name: string; type: string; description: string }[];
  type: ToolType;
  examples: { description: string; query: string }[];
  param_names: string[];
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

interface ToolColumn {
  name: string;
  type: string;
  description: string;
}

interface ToolQueryExamples {
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
  param_names: string[];
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

export async function uploadCsv(name: string, content: string): Promise<string> {
  const res = await fetch("/answer/v1/upload", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    credentials: "include",
    body: JSON.stringify({ name, content }),
  });

  if (res.status === 401) {
    throw new Error("unauthorized");
  }

  if (!res.ok) {
    throw new Error("upload failed");
  }

  const data: unknown = await res.json();
  if (
    typeof data === "object" &&
    data !== null &&
    "url" in data &&
    typeof (data as { url: unknown }).url === "string"
  ) {
    return (data as { url: string }).url;
  }
  throw new Error("upload response is invalid");
}

export function redirectToLogin() {
  window.location.href = "/auth/v1/get-permission";
}

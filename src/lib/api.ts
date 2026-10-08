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

export async function addBookmark(args: {
  id?: string;
  name: string;
  param_name: string;
  param_value: string;
  query: string;
}): Promise<void> {
  const res = await fetch("/answer/v1/bookmark", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    credentials: "include",
    body: JSON.stringify(args),
  });

  if (res.status === 401) {
    throw new Error("unauthorized");
  }

  if (!res.ok) {
    throw new Error("request failed");
  }
}

export interface Bookmark {
  id: string;
  name: string;
  updated_at: string;
}

export interface BookmarkDetail {
  name: string;
  param_name: string;
  param_value: string;
  query: string;
  hashed_tool: string;
  updated_at: string;
}

// List the current user's bookmarks. The BE returns an array.
export async function fetchBookmarks(): Promise<Bookmark[]> {
  const res = await fetch("/answer/v1/bookmarks", {
    method: "GET",
    credentials: "include",
  });

  if (res.status === 401) {
    throw new Error("unauthorized");
  }

  if (!res.ok) {
    throw new Error("request failed");
  }

  const data: unknown = await res.json();
  if (!Array.isArray(data)) {
    throw new Error("bookmarks response is invalid");
  }
  return data.map((row) => {
    const r = row as Record<string, unknown>;
    return {
      id: typeof r.id === "string" ? r.id : "",
      name: typeof r.name === "string" ? r.name : "",
      updated_at: typeof r.updated_at === "string" ? r.updated_at : "",
    };
  });
}

// Fetch one bookmark. The BE also sets the hashed_tool cookie, so a
// following callTool() runs the bookmarked tool.
export async function fetchBookmark(id: string): Promise<BookmarkDetail> {
  const url = new URL("/answer/v1/bookmark", window.location.origin);
  url.searchParams.set("id", id);

  const res = await fetch(url.toString(), {
    method: "GET",
    credentials: "include",
  });

  if (res.status === 401) {
    throw new Error("unauthorized");
  }

  if (res.status === 404) {
    throw new Error("bookmark not found");
  }

  if (!res.ok) {
    throw new Error("request failed");
  }

  const data: unknown = await res.json();
  if (typeof data !== "object" || data === null) {
    throw new Error("bookmark response is invalid");
  }
  const d = data as Record<string, unknown>;
  return {
    name: typeof d.name === "string" ? d.name : "",
    param_name: typeof d.param_name === "string" ? d.param_name : "",
    param_value: typeof d.param_value === "string" ? d.param_value : "",
    query: typeof d.query === "string" ? d.query : "",
    hashed_tool: typeof d.hashed_tool === "string" ? d.hashed_tool : "",
    updated_at: typeof d.updated_at === "string" ? d.updated_at : "",
  };
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

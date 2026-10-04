import {
  useState,
  useRef,
  useEffect,
  type FormEvent,
  type KeyboardEvent,
} from "react";
import { useNavigate } from "react-router-dom";
import { fetchAnswer, uploadCsv } from "../lib/api";
import Header from "../components/Header";
import HighlightedCode from "../components/HighlightedCode";
import toolStyles from "./Tool.module.css";
import styles from "./Query.module.css";

interface Message {
  prompt: string;
  result?: unknown;
  funcCalls?: string;
  error?: string;
}

const PAGE_SIZE = 20;

export default function Query() {
  const [paramValue, setParamValue] = useState("");
  const [paramValueDraft, setParamValueDraft] = useState("");
  const [paramName, setParamName] = useState("");
  const [paramNameDraft, setParamNameDraft] = useState("");
  const [prompt, setPrompt] = useState("");
  const [message, setMessage] = useState<Message | null>(null);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const promptRef = useRef<HTMLTextAreaElement>(null);
  const valueRefs = useRef<Array<HTMLParagraphElement | null>>([]);
  const [clipped, setClipped] = useState<boolean[]>([false, false]);
  const [resultTab, setResultTab] = useState<"answer" | "tool">("answer");
  const [saveOpen, setSaveOpen] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [uploadUrl, setUploadUrl] = useState<string | null>(null);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const saveRef = useRef<HTMLDivElement>(null);
  const [fullValue, setFullValue] = useState<{
    label: string;
    value: string;
  } | null>(null);
  const [resetTarget, setResetTarget] = useState<"value" | "name" | null>(null);
  const [resetDraft, setResetDraft] = useState("");
  const navigate = useNavigate();

  const paramLocked = paramValue.length > 0 && paramName.length > 0;
  const funcCalls = message?.funcCalls ?? "";
  const answerRows =
    message && !message.error && message.result !== undefined
      ? unwrapAnswer(message.result)
      : null;
  const canUpload = answerRows !== null && answerRows.length > 0;
  // The BE returns no total count. Assume another page exists when the
  // current page is full.
  const canGoNext = answerRows !== null && answerRows.length >= PAGE_SIZE;

  useEffect(() => {
    if (paramLocked) promptRef.current?.focus();
  }, [paramLocked]);

  // Show "See more" only when the value needs more than 5 lines.
  useEffect(() => {
    setClipped(
      valueRefs.current.map((el) =>
        el ? el.scrollHeight > el.clientHeight + 1 : false,
      ),
    );
  }, [paramValue, paramName, paramLocked]);

  useEffect(() => {
    if (!fullValue) return;
    const onKey = (e: globalThis.KeyboardEvent) => {
      if (e.key === "Escape") setFullValue(null);
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [fullValue]);

  // Close the param reset popup with Escape.
  useEffect(() => {
    if (!resetTarget) return;
    const onKey = (e: globalThis.KeyboardEvent) => {
      if (e.key === "Escape") setResetTarget(null);
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [resetTarget]);

  // Close the Save result menu when clicking outside it.
  useEffect(() => {
    if (!saveOpen) return;
    const onDown = (e: globalThis.MouseEvent) => {
      if (saveRef.current && !saveRef.current.contains(e.target as Node)) {
        setSaveOpen(false);
      }
    };
    document.addEventListener("mousedown", onDown);
    return () => document.removeEventListener("mousedown", onDown);
  }, [saveOpen]);

  const handleSetParam = (e: FormEvent) => {
    e.preventDefault();
    const trimmedParamValue = paramValueDraft.trim();
    const trimmedParamName = paramNameDraft.trim();
    if (trimmedParamValue && trimmedParamName) {
      setParamValue(trimmedParamValue);
      setParamName(trimmedParamName);
    }
  };

  // Open the popup that changes one param, without leaving this page.
  const openReset = (target: "value" | "name") => {
    setResetDraft(target === "value" ? paramValue : paramName);
    setResetTarget(target);
  };

  const handleResetSave = (e: FormEvent) => {
    e.preventDefault();
    const nextValue = resetDraft.trim();
    if (!nextValue || !resetTarget) return;
    if (resetTarget === "value") setParamValue(nextValue);
    else setParamName(nextValue);
    setResetDraft("");
    setResetTarget(null);
  };

  // Run a query and replace the result. Return true on success.
  const runQuery = async (queryPrompt: string, offset: number) => {
    setLoading(true);
    try {
      const data = await fetchAnswer(
        queryPrompt,
        paramValue,
        paramName,
        PAGE_SIZE,
        offset,
      );
      setMessage({
        prompt: queryPrompt,
        result: data,
        funcCalls: getFuncCalls(data),
      });
      return true;
    } catch (e) {
      if (e instanceof Error && e.message === "unauthorized") {
        navigate("/login");
        return false;
      }
      const errMsg = e instanceof Error ? e.message : "something went wrong";
      setMessage({ prompt: queryPrompt, error: errMsg });
      return false;
    } finally {
      setLoading(false);
    }
  };

  const handleAsk = async () => {
    if (!prompt.trim() || !paramLocked || loading) return;
    const currentPrompt = prompt.trim();

    // Set placeholder message (overwrites previous)
    setMessage({ prompt: currentPrompt });
    setUploadUrl(null);
    setUploadError(null);
    setResultTab("answer");
    setSaveOpen(false);
    setPage(1);

    await runQuery(currentPrompt, 0);
  };

  // Load another page of the current result. The BE re-runs the query.
  const goToPage = async (nextPage: number) => {
    if (!message || message.error || loading || nextPage < 1) return;
    setResultTab("answer");
    setSaveOpen(false);
    const ok = await runQuery(message.prompt, (nextPage - 1) * PAGE_SIZE);
    if (ok) setPage(nextPage);
  };

  const handleUpload = async () => {
    if (!answerRows || answerRows.length === 0 || uploading) return;
    setUploading(true);
    setUploadError(null);
    setUploadUrl(null);

    try {
      const url = await uploadCsv(csvFileName(), rowsToCsv(answerRows));
      setUploadUrl(url);
    } catch (e) {
      if (e instanceof Error && e.message === "unauthorized") {
        navigate("/login");
        return;
      }
      setUploadError(e instanceof Error ? e.message : "upload failed");
    } finally {
      setUploading(false);
      setSaveOpen(false);
    }
  };

  const handleDownload = () => {
    if (!answerRows || answerRows.length === 0) return;
    const blob = new Blob([rowsToCsv(answerRows)], {
      type: "text/csv;charset=utf-8",
    });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = csvFileName();
    document.body.appendChild(link);
    link.click();
    link.remove();
    URL.revokeObjectURL(url);
    setSaveOpen(false);
  };

  const handlePromptKey = (e: KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleAsk();
    }
  };

  return (
    <div className={styles.page}>
      {uploadUrl && (
        <div className={toolStyles.snackbar}>
          <span className={toolStyles.snackbarText}>
            {"CSV has been uploaded successfully! "}
            <a
              className={styles.snackLink}
              href={uploadUrl}
              target="_blank"
              rel="noreferrer"
            >
              Open in Google Drive
            </a>
          </span>
          <button
            type="button"
            className={toolStyles.snackbarClose}
            onClick={() => setUploadUrl(null)}
          >
            x
          </button>
        </div>
      )}
      <Header />

      <main className={styles.main}>
        <div className={styles.card}>
          <h2 className={styles.cardTitle}>Query</h2>
          {/* Param setup */}
          {!paramLocked ? (
            <form onSubmit={handleSetParam} className={styles.setupCard}>
              <div className={styles.setupRow}>
                <div className={styles.setupField}>
                  <p className={styles.setupLabel}>Find these records ...</p>
                  <textarea
                    className={styles.setupInputTextarea}
                    placeholder="e.g. alice@example.com"
                    value={paramValueDraft}
                    onChange={(e) => setParamValueDraft(e.target.value)}
                    rows={3}
                    autoFocus
                  />
                </div>
                <div className={styles.setupField}>
                  <p className={styles.setupLabel}>... on this column</p>
                  <input
                    className={styles.setupInput}
                    placeholder="e.g. email"
                    value={paramNameDraft}
                    onChange={(e) => setParamNameDraft(e.target.value)}
                  />
                </div>
              </div>
              <div className={styles.setupSubmit}>
                <button
                  type="submit"
                  className={styles.setupBtn}
                  style={{
                    opacity:
                      paramValueDraft.trim() && paramNameDraft.trim() ? 1 : 0.5,
                  }}
                  disabled={!paramValueDraft.trim() || !paramNameDraft.trim()}
                >
                  Set
                </button>
              </div>
            </form>
          ) : (
            <>
              <div className={styles.rows}>
                {/* Info cards */}
                <div className={styles.infoCards}>
                  <div className={styles.infoCard}>
                    <p className={styles.infoCardLabel}>
                      Find these records <span className={styles.paren}>(</span>
                      <button
                        type="button"
                        className={styles.resetLink}
                        onClick={() => openReset("value")}
                      >
                        reset
                      </button>
                      <span className={styles.paren}>)</span>
                    </p>
                    <p
                      className={styles.infoCardValue}
                      ref={(el) => {
                        valueRefs.current[0] = el;
                      }}
                    >
                      {paramValue}
                    </p>
                    {clipped[0] && (
                      <button
                        type="button"
                        className={styles.seeMore}
                        onClick={() =>
                          setFullValue({
                            label: "Find these records ...",
                            value: paramValue,
                          })
                        }
                      >
                        See more
                      </button>
                    )}
                  </div>
                  <div className={styles.infoCard}>
                    <p className={styles.infoCardLabel}>
                      {"on this column "}
                      <span className={styles.paren}>(</span>
                      <button
                        type="button"
                        className={styles.resetLink}
                        onClick={() => openReset("name")}
                      >
                        reset
                      </button>
                      <span className={styles.paren}>)</span>
                    </p>
                    <p
                      className={styles.infoCardValue}
                      ref={(el) => {
                        valueRefs.current[1] = el;
                      }}
                    >
                      {paramName}
                    </p>
                    {clipped[1] && (
                      <button
                        type="button"
                        className={styles.seeMore}
                        onClick={() =>
                          setFullValue({
                            label: "On this column",
                            value: paramName,
                          })
                        }
                      >
                        See more
                      </button>
                    )}
                  </div>
                </div>

                {/* Input bar */}
                <div className={styles.inputBar}>
                  <textarea
                    ref={promptRef}
                    className={styles.promptInput}
                    placeholder="What do you want to know?"
                    value={prompt}
                    onChange={(e) => setPrompt(e.target.value)}
                    onKeyDown={handlePromptKey}
                    rows={1}
                  />
                  <button
                    className={styles.askBtn}
                    style={{ opacity: loading || !prompt.trim() ? 0.5 : 1 }}
                    onClick={handleAsk}
                    disabled={loading || !prompt.trim()}
                  >
                    {loading ? "..." : "Get Answer"}
                  </button>
                </div>

                {/* Result tabs, own row above the result box */}
                {message !== null &&
                  !message.error &&
                  message.result !== undefined && (
                    <>
                      <div className={toolStyles.tabs}>
                        <button
                          type="button"
                          className={`${toolStyles.tab} ${
                            resultTab === "answer" ? toolStyles.tabActive : ""
                          }`}
                          onClick={() => setResultTab("answer")}
                        >
                          Answer
                        </button>
                        <div className={styles.saveWrap} ref={saveRef}>
                          <button
                            type="button"
                            className={`${toolStyles.tab} ${
                              canUpload ? "" : styles.tabDisabled
                            }`}
                            aria-expanded={saveOpen}
                            onClick={() => setSaveOpen((v) => !v)}
                            disabled={!canUpload}
                          >
                            Save result{" "}
                            <span className={styles.saveCaret}>
                              {saveOpen ? "\u25b2" : "\u25bc"}
                            </span>
                          </button>
                          {saveOpen && (
                            <div className={styles.saveMenu}>
                              <button
                                type="button"
                                className={styles.saveMenuItem}
                                onClick={handleUpload}
                                disabled={uploading}
                              >
                                {uploading
                                  ? "Uploading..."
                                  : "Upload to Google Drive as CSV"}
                              </button>
                              <button
                                type="button"
                                className={styles.saveMenuItem}
                                onClick={handleDownload}
                              >
                                Download as CSV
                              </button>
                            </div>
                          )}
                        </div>
                        <button
                          type="button"
                          className={`${toolStyles.tab} ${
                            resultTab === "tool" ? toolStyles.tabActive : ""
                          } ${funcCalls ? "" : styles.tabDisabled}`}
                          onClick={() => setResultTab("tool")}
                          disabled={!funcCalls}
                        >
                          Tool call
                        </button>
                      </div>
                      {uploadError && (
                        <p className={styles.uploadError}>{uploadError}</p>
                      )}
                    </>
                  )}

                {/* Result box: answer tab shows table, tool call tab shows code */}
                {resultTab === "answer" ? (
                  <div className={styles.tableContainer}>
                    {message === null ? (
                      <p className={styles.emptyTable}>
                        Ask a question above to see results.
                      </p>
                    ) : message.error ? (
                      <p className={styles.error}>{message.error}</p>
                    ) : message.result !== undefined ? (
                      <>
                        <ResultTable data={message.result} />
                        {answerRows !== null && (
                          <div className={styles.pager}>
                            <button
                              type="button"
                              className={styles.pagerBtn}
                              onClick={() => goToPage(page - 1)}
                              disabled={page <= 1 || loading}
                            >
                              Prev
                            </button>
                            <span className={styles.pagerPage}>
                              Page {page}
                            </span>
                            <button
                              type="button"
                              className={styles.pagerBtn}
                              onClick={() => goToPage(page + 1)}
                              disabled={!canGoNext || loading}
                            >
                              Next
                            </button>
                          </div>
                        )}
                      </>
                    ) : (
                      <p className={styles.thinking}>Thinking...</p>
                    )}
                  </div>
                ) : (
                  <div className={styles.funcCallContainer}>
                    <HighlightedCode
                      code={funcCalls}
                      lang="json"
                      className={styles.funcCallBody}
                    />
                  </div>
                )}
              </div>
            </>
          )}
          {fullValue && (
            <div
              className={styles.modalOverlay}
              onClick={() => setFullValue(null)}
            >
              <div
                className={styles.modal}
                role="dialog"
                aria-modal="true"
                onClick={(e) => e.stopPropagation()}
              >
                <div className={styles.modalHeader}>
                  <p className={styles.infoCardLabel}>{fullValue.label}</p>
                  <button
                    type="button"
                    className={styles.modalClose}
                    aria-label="Close"
                    onClick={() => setFullValue(null)}
                  >
                    ×
                  </button>
                </div>
                <p className={styles.modalValue}>{fullValue.value}</p>
              </div>
            </div>
          )}
          {/* Param reset popup: change one value, keep prompt and result. */}
          {resetTarget && (
            <div
              className={styles.modalOverlay}
              onClick={() => setResetTarget(null)}
            >
              <form
                className={styles.modal}
                role="dialog"
                aria-modal="true"
                onClick={(e) => e.stopPropagation()}
                onSubmit={handleResetSave}
              >
                <div className={styles.modalHeader}>
                  <p className={styles.infoCardLabel}>
                    {resetTarget === "value"
                      ? "Find these records ..."
                      : "On this column"}
                  </p>
                  <button
                    type="button"
                    className={styles.modalClose}
                    aria-label="Close"
                    onClick={() => setResetTarget(null)}
                  >
                    ×
                  </button>
                </div>
                {resetTarget === "value" ? (
                  <textarea
                    className={styles.setupInputTextarea}
                    placeholder="e.g. alice@example.com"
                    value={resetDraft}
                    onChange={(e) => setResetDraft(e.target.value)}
                    rows={3}
                    autoFocus
                  />
                ) : (
                  <input
                    className={styles.setupInput}
                    placeholder="e.g. email"
                    value={resetDraft}
                    onChange={(e) => setResetDraft(e.target.value)}
                    autoFocus
                  />
                )}
                <div className={styles.setupSubmit}>
                  <button
                    type="submit"
                    className={styles.setupBtn}
                    style={{ opacity: resetDraft.trim() ? 1 : 0.5 }}
                    disabled={!resetDraft.trim()}
                  >
                    Set
                  </button>
                </div>
              </form>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

interface ResultTableProps {
  data: unknown;
}

function ResultTable({ data }: ResultTableProps) {
  const rows = unwrapAnswer(data);

  if (rows && rows.length === 0) {
    return <p className={styles.emptyResult}>No results found.</p>;
  }

  if (
    rows &&
    rows.length > 0 &&
    typeof rows[0] === "object" &&
    rows[0] !== null
  ) {
    const columns = Object.keys(rows[0]);
    return (
      <div className={styles.tableWrap}>
        <table className={styles.table}>
          <thead>
            <tr>
              {columns.map((col) => (
                <th key={col} className={styles.th}>
                  {col}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((row, i) => (
              <tr key={i} className={i % 2 === 0 ? styles.trEven : undefined}>
                {columns.map((col) => (
                  <td key={col} className={styles.td}>
                    {renderCell(row[col])}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  }

  const textAnswer = getTextAnswer(data);
  if (textAnswer !== null) {
    return <p className={styles.emptyResult}>{textAnswer}</p>;
  }

  return <pre className={styles.pre}>{JSON.stringify(data, null, 2)}</pre>;
}

function unwrapAnswer(data: unknown): Record<string, unknown>[] | null {
  if (Array.isArray(data)) return data as Record<string, unknown>[];
  if (data && typeof data === "object" && !Array.isArray(data)) {
    const obj = data as Record<string, unknown>;
    if ("answer" in obj && Array.isArray(obj.answer)) {
      return obj.answer as Record<string, unknown>[];
    }
  }
  return null;
}

function renderCell(value: unknown): string {
  if (value === null || value === undefined) return "";
  if (typeof value === "object") return JSON.stringify(value);
  return String(value);
}

// The endpoint returns { answer, stringified_func_calls }.
// Show the func call only when it is a non-empty string.
function getFuncCalls(data: unknown): string | undefined {
  if (data && typeof data === "object" && !Array.isArray(data)) {
    const value = (data as Record<string, unknown>).stringified_func_calls;
    if (typeof value === "string" && value.trim() !== "") return value;
  }
  return undefined;
}

function getTextAnswer(data: unknown): string | null {
  if (data && typeof data === "object" && !Array.isArray(data)) {
    const obj = data as Record<string, unknown>;
    if ("answer" in obj && typeof obj.answer === "string") {
      return obj.answer as string;
    }
  }
  return null;
}

function rowsToCsv(rows: Record<string, unknown>[]): string {
  const columns = Object.keys(rows[0]);
  const lines = [columns.map(csvCell).join(",")];
  for (const row of rows) {
    lines.push(columns.map((col) => csvCell(renderCell(row[col]))).join(","));
  }
  return lines.join("\n");
}

function csvCell(value: string): string {
  if (/[",\n\r]/.test(value)) {
    return `"${value.replace(/"/g, '""')}"`;
  }
  return value;
}

function csvFileName(): string {
  const d = new Date();
  const pad = (n: number) => String(n).padStart(2, "0");
  const date = `${d.getFullYear()}${pad(d.getMonth() + 1)}${pad(d.getDate())}`;
  const time = `${pad(d.getHours())}${pad(d.getMinutes())}${pad(d.getSeconds())}`;
  return `answer_${date}_${time}.csv`;
}

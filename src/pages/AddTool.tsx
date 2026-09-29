import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { addTool, validateToolQuery, type ToolType } from "../lib/api";
import Header from "../components/Header";
import styles from "./Tool.module.css";

interface Column {
  name: string;
  type: string;
  description: string;
}

type QueryStatus = "unverified" | "checking" | "valid" | "invalid";

interface Query {
  description: string;
  query: string;
  status: QueryStatus;
  message?: string;
}

interface FormData {
  tool_description: string;
  project: string;
  dataset: string;
  table_name: string;
  columns: Column[];
  query: Query[];
}

export default function AddTool() {
  const navigate = useNavigate();
  const [mode, setMode] = useState<ToolType>("table");
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);
  // Index of the column checked as Param. Its name becomes param_name.
  const [paramIndex, setParamIndex] = useState<number | null>(null);

  const changeMode = (next: ToolType) => {
    setSuccess(false);
    setError(null);
    if (next === "query") {
      // Query mode allows only one query.
      setFormData((prev) => ({ ...prev, query: prev.query.slice(0, 1) }));
    }
    setMode(next);
  };

  const [formData, setFormData] = useState<FormData>({
    tool_description: "",
    project: "",
    dataset: "",
    table_name: "",
    columns: [{ name: "", type: "", description: "" }],
    query: [{ description: "", query: "", status: "unverified" }],
  });

  const handleInputChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
    field: keyof FormData,
  ) => {
    setFormData((prev) => ({ ...prev, [field]: e.target.value }));
  };

  const handleColumnChange = (
    index: number,
    field: keyof Column,
    value: string,
  ) => {
    const newColumns = [...formData.columns];
    newColumns[index] = { ...newColumns[index], [field]: value };
    setFormData((prev) => ({ ...prev, columns: newColumns }));
  };

  const addColumn = () => {
    setFormData((prev) => ({
      ...prev,
      columns: [...prev.columns, { name: "", type: "", description: "" }],
    }));
  };

  const removeColumn = (index: number) => {
    if (formData.columns.length <= 1) return;
    const newColumns = formData.columns.filter((_, i) => i !== index);
    setFormData((prev) => ({ ...prev, columns: newColumns }));
    if (paramIndex === index) setParamIndex(null);
    else if (paramIndex !== null && index < paramIndex) {
      setParamIndex(paramIndex - 1);
    }
  };

  const toggleParam = (index: number, checked: boolean) => {
    setParamIndex(checked ? index : null);
  };

  const handleQueryChange = (
    index: number,
    field: keyof Query,
    value: string,
  ) => {
    const newQuery = [...formData.query];
    newQuery[index] = { ...newQuery[index], [field]: value };
    // Query text changed, so the old verification no longer counts.
    if (field === "query") {
      newQuery[index].status = "unverified";
      newQuery[index].message = undefined;
    }
    setFormData((prev) => ({ ...prev, query: newQuery }));
  };

  const addQuery = () => {
    setFormData((prev) => ({
      ...prev,
      query: [
        ...prev.query,
        { description: "", query: "", status: "unverified" as const },
      ],
    }));
  };

  const removeQuery = (index: number) => {
    if (formData.query.length <= 1) return;
    const newQuery = formData.query.filter((_, i) => i !== index);
    setFormData((prev) => ({ ...prev, query: newQuery }));
  };

  // Tab inserts a tab character. Shift+Tab removes one indent level.
  const handleTabKey = (
    e: React.KeyboardEvent<HTMLTextAreaElement>,
    value: string,
    apply: (next: string) => void,
  ) => {
    if (e.key !== "Tab") return;
    e.preventDefault();
    const el = e.currentTarget;
    const start = el.selectionStart;
    const end = el.selectionEnd;

    if (!e.shiftKey) {
      const next = value.slice(0, start) + "\t" + value.slice(end);
      apply(next);
      requestAnimationFrame(() => el.setSelectionRange(start + 1, start + 1));
      return;
    }

    // Shift+Tab: remove one indent level from every line in the block.
    const blockStart = value.lastIndexOf("\n", start - 1) + 1;
    const nextBreak = value.indexOf("\n", end);
    const blockEnd = nextBreak === -1 ? value.length : nextBreak;
    const block = value.slice(blockStart, blockEnd);

    const removeIndent = (line: string) => {
      if (line.startsWith("\t")) return line.slice(1);
      if (line.startsWith("    ")) return line.slice(4);
      return line;
    };

    const lines = block.split("\n");
    const dedented = lines.map(removeIndent).join("\n");
    apply(value.slice(0, blockStart) + dedented + value.slice(blockEnd));

    const hadSelection = end > start;
    const caret = hadSelection
      ? blockStart + dedented.length
      : Math.max(
          blockStart,
          start - (lines[0].length - removeIndent(lines[0]).length),
        );
    requestAnimationFrame(() => el.setSelectionRange(caret, caret));
  };

  const setQueryStatus = (
    index: number,
    text: string,
    status: QueryStatus,
    message?: string,
  ) => {
    setFormData((prev) => {
      // Text changed or row removed while the request ran.
      // Keep the row unverified and drop the stale result.
      if (prev.query[index]?.query !== text) return prev;
      return {
        ...prev,
        query: prev.query.map((q, i) =>
          i === index ? { ...q, status, message } : q,
        ),
      };
    });
  };

  const verifyQuery = async (index: number) => {
    const text = formData.query[index].query;
    if (!text.trim()) {
      setQueryStatus(index, text, "unverified", "Query is empty");
      return;
    }

    setQueryStatus(index, text, "checking");
    try {
      const valid = await validateToolQuery(text);
      setQueryStatus(index, text, valid ? "valid" : "invalid");
    } catch (e) {
      if (e instanceof Error && e.message === "unauthorized") {
        navigate("/login");
        return;
      }
      setQueryStatus(
        index,
        text,
        "unverified",
        "Failed to verify",
      );
    }
  };

  // Every query must pass verification before the tool can be added.
  const allQueriesValid =
    formData.query.length > 0 &&
    formData.query.every((q) => q.status === "valid");

  const verifyRow = (index: number) => {
    const { status, message } = formData.query[index];
    // No text until the user verifies or an error happens.
    const text =
      message ??
      (status === "checking"
        ? "Checking..."
        : status === "valid"
          ? "Verified"
          : status === "invalid"
            ? "Invalid query"
            : null);
    const tone =
      status === "valid"
        ? styles.verifyOk
        : status === "invalid" || message
          ? styles.verifyBad
          : styles.verifyIdle;

    return (
      <div className={styles.verifyRow}>
        <button
          type="button"
          className={styles.verifyBtn}
          onClick={() => verifyQuery(index)}
          disabled={status === "checking"}
        >
          {status === "checking" ? "Verifying..." : "Verify"}
        </button>
        {text !== null && (
          <span className={`${styles.verifyStatus} ${tone}`}>{text}</span>
        )}
      </div>
    );
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(false);

    // Validate
    const missingTableName = mode === "table" && !formData.table_name.trim();
    const paramColumn =
      paramIndex !== null ? formData.columns[paramIndex] : undefined;
    if (
      !formData.tool_description.trim() ||
      !formData.project.trim() ||
      !formData.dataset.trim() ||
      missingTableName ||
      !paramColumn ||
      !paramColumn.name.trim()
    ) {
      setError("Please fill in all fields");
      return;
    }

    if (!allQueriesValid) {
      setError("Verify all queries before adding the tool");
      return;
    }

    setLoading(true);
    try {
      await addTool({
        tool_description: formData.tool_description,
        project: formData.project,
        dataset: formData.dataset,
        table_name: mode === "table" ? formData.table_name : "",
        columns: formData.columns,
        type: mode,
        examples: formData.query.map(({ description, query }) => ({
          description,
          query,
        })),
        param_name: paramColumn.name.trim(),
        param_type: "",
        param_description: "",
      });
      setSuccess(true);
      window.scrollTo(0, 0);
      // Reset form
      setFormData({
        tool_description: "",
        project: "",
        dataset: "",
        table_name: "",
        columns: [{ name: "", type: "", description: "" }],
        query: [{ description: "", query: "", status: "unverified" }],
      });
      setParamIndex(null);
    } catch (e) {
      if (e instanceof Error && e.message === "unauthorized") {
        navigate("/login");
        return;
      }
      setError(e instanceof Error ? e.message : "Failed to add tool");
    } finally {
      setLoading(false);
    }
  };

  const queryForm = (
    <div className={styles.section}>
      <p className={styles.sectionHeader}>
        {mode === "table" ? "Example" : "Query"}
      </p>
      {mode === "table" ? (
        <div className={styles.greyCardList}>
          {formData.query.map((entry, index) => (
            <div
              key={index}
              className={`${styles.greyCard} ${styles.exampleCard}`}
            >
              {formData.query.length > 1 && (
                <button
                  type="button"
                  className={styles.removeBtn}
                  onClick={() => removeQuery(index)}
                >
                  x
                </button>
              )}
              <div className={styles.formGroup}>
                <div className={styles.fieldHead}>
                  <label className={styles.label}>Description</label>
                </div>
                <input
                  className={styles.input}
                  type="text"
                  placeholder="e.g. find user by email"
                  value={entry.description}
                  onChange={(e) =>
                    handleQueryChange(index, "description", e.target.value)
                  }
                />
              </div>
              <div className={styles.formGroup}>
                <label className={styles.label}>Query</label>
                <textarea
                  className={styles.queryArea}
                  placeholder="e.g. SELECT * FROM users WHERE email = ?"
                  value={entry.query}
                  rows={15}
                  onKeyDown={(e) =>
                    handleTabKey(e, entry.query, (next) =>
                      handleQueryChange(index, "query", next),
                    )
                  }
                  onChange={(e) =>
                    handleQueryChange(index, "query", e.target.value)
                  }
                />
                {verifyRow(index)}
              </div>
            </div>
          ))}
        </div>
      ) : (
        <>
          <textarea
            className={styles.queryArea}
            placeholder="e.g. SELECT * FROM users WHERE email = ?"
            value={formData.query[0].query}
            rows={15}
            onKeyDown={(e) =>
              handleTabKey(e, formData.query[0].query, (next) =>
                handleQueryChange(0, "query", next),
              )
            }
            onChange={(e) => handleQueryChange(0, "query", e.target.value)}
          />
          {verifyRow(0)}
        </>
      )}
      {mode === "table" && (
        <button type="button" className={styles.addBtn} onClick={addQuery}>
          + Add Example
        </button>
      )}
    </div>
  );

  const columnsForm = (
    <div className={styles.section}>
      <p className={styles.sectionHeader}>Columns</p>
      <div className={styles.greyCardList}>
        {formData.columns.map((col, index) => (
          <div key={index} className={styles.greyCard}>
            <div className={styles.formGroup}>
              <div className={styles.arrayRow}>
                <span className={styles.arrayLabel}>Name</span>
                <span className={styles.arrayLabel}>Type</span>
                <span className={styles.arrayLabelFixed}>Set as param</span>
                {formData.columns.length > 1 && (
                  <span className={styles.removeSpacer} />
                )}
              </div>
              <div className={styles.arrayRow}>
                <input
                  className={styles.arrayInput}
                  type="text"
                  placeholder="e.g. email"
                  value={col.name}
                  onChange={(e) =>
                    handleColumnChange(index, "name", e.target.value)
                  }
                />
                <input
                  className={styles.arrayInput}
                  type="text"
                  placeholder="e.g. string"
                  value={col.type}
                  onChange={(e) =>
                    handleColumnChange(index, "type", e.target.value)
                  }
                />
                <div className={styles.checkCol}>
                  <input
                    className={styles.check}
                    type="checkbox"
                    checked={paramIndex === index}
                    onChange={(e) => toggleParam(index, e.target.checked)}
                  />
                </div>
                {formData.columns.length > 1 && (
                  <button
                    type="button"
                    className={styles.removeBtn}
                    onClick={() => removeColumn(index)}
                  >
                    x
                  </button>
                )}
              </div>
            </div>
            <div className={styles.formGroup}>
              <label className={styles.label}>Description</label>
              <input
                className={styles.arrayInput}
                type="text"
                placeholder="e.g. user email"
                value={col.description}
                onChange={(e) =>
                  handleColumnChange(index, "description", e.target.value)
                }
              />
            </div>
          </div>
        ))}
      </div>
      <button type="button" className={styles.addBtn} onClick={addColumn}>
        + Add Column
      </button>
    </div>
  );

  return (
    <div className={styles.page}>
      {success && (
        <div className={styles.snackbar}>
          <span className={styles.snackbarText}>
            Tool has been added successfully! You can now use it in the query.
          </span>
          <button
            type="button"
            className={styles.snackbarClose}
            onClick={() => setSuccess(false)}
          >
            x
          </button>
        </div>
      )}
      <Header />

      <main className={styles.main}>
        <div className={styles.card}>
          <h2 className={styles.cardTitle}>Add New Tool</h2>

          <div className={styles.tabs}>
            <button
              type="button"
              className={`${styles.tab} ${
                mode === "table" ? styles.tabActive : ""
              }`}
              onClick={() => changeMode("table")}
            >
              By Table
            </button>
            <button
              type="button"
              className={`${styles.tab} ${
                mode === "query" ? styles.tabActive : ""
              }`}
              onClick={() => changeMode("query")}
            >
              By Query
            </button>
          </div>

          <form onSubmit={handleSubmit} className={styles.form}>
            {/* Description */}
            <div className={styles.section}>
              <p className={styles.sectionHeader}>Description</p>
              <div className={styles.formGroup}>
                <textarea
                  className={styles.textarea}
                  placeholder="Describe what this tool does..."
                  value={formData.tool_description}
                  onChange={(e) => handleInputChange(e, "tool_description")}
                  onKeyDown={(e) =>
                    handleTabKey(e, formData.tool_description, (next) =>
                      setFormData((prev) => ({
                        ...prev,
                        tool_description: next,
                      })),
                    )
                  }
                  rows={3}
                />
              </div>
            </div>

            {/* Location */}
            <div className={styles.section}>
              <p className={styles.sectionHeader}>Location</p>
              <div className={styles.formRow}>
                <div className={styles.formGroup}>
                  <label className={styles.label}>Project</label>
                  <input
                    className={styles.input}
                    type="text"
                    placeholder="e.g. my-project"
                    value={formData.project}
                    onChange={(e) => handleInputChange(e, "project")}
                  />
                </div>
                <div className={styles.formGroup}>
                  <label className={styles.label}>Dataset</label>
                  <input
                    className={styles.input}
                    type="text"
                    placeholder="e.g. my_dataset"
                    value={formData.dataset}
                    onChange={(e) => handleInputChange(e, "dataset")}
                  />
                </div>
                {mode === "table" && (
                  <div className={styles.formGroup}>
                    <label className={styles.label}>Table Name</label>
                    <input
                      className={styles.input}
                      type="text"
                      placeholder="e.g. users"
                      value={formData.table_name}
                      onChange={(e) => handleInputChange(e, "table_name")}
                    />
                  </div>
                )}
              </div>
            </div>

            {/* Table: Columns then Example. Query: Example then Columns. */}
            {mode === "table" ? columnsForm : queryForm}
            {mode === "table" ? queryForm : columnsForm}

            {/* Submit */}
            <div className={styles.submitRow}>
              <button
                type="submit"
                className={styles.submitBtn}
                disabled={loading || !allQueriesValid}
              >
                {loading ? "Adding..." : "Add Tool"}
              </button>
              {error && <div className={styles.error}>{error}</div>}
            </div>
          </form>
        </div>
      </main>
    </div>
  );
}

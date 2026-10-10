import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import {
  addTool,
  fetchToolTableMetadata,
  validateToolQuery,
  type ToolType,
} from "../lib/api";
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

const emptyForm = (): FormData => ({
  tool_description: "",
  project: "",
  dataset: "",
  table_name: "",
  columns: [{ name: "", type: "", description: "" }],
  query: [{ description: "", query: "", status: "unverified" }],
});

export default function AddTool() {
  const navigate = useNavigate();
  const [mode, setMode] = useState<ToolType>("table");
  const [loading, setLoading] = useState(false);
  const [success, setSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [filling, setFilling] = useState(false);

  // Each tab keeps its own form, so a half filled form is kept on tab switch.
  const [tableData, setTableData] = useState<FormData>(emptyForm);
  const [queryData, setQueryData] = useState<FormData>(emptyForm);
  // Indexes of the columns checked as Param. Their names become param_names.
  const [tableParamIndexes, setTableParamIndexes] = useState<number[]>([]);
  const [queryParamIndexes, setQueryParamIndexes] = useState<number[]>([]);
  const [tableFillError, setTableFillError] = useState<string | null>(null);
  const [queryFillError, setQueryFillError] = useState<string | null>(null);

  // Read and write only the state of the active tab.
  const formData = mode === "table" ? tableData : queryData;
  const paramIndexes = mode === "table" ? tableParamIndexes : queryParamIndexes;
  const fillError = mode === "table" ? tableFillError : queryFillError;
  const setFormData: React.Dispatch<React.SetStateAction<FormData>> = (
    value,
  ) => (mode === "table" ? setTableData(value) : setQueryData(value));
  const setParamIndexes: React.Dispatch<React.SetStateAction<number[]>> = (
    value,
  ) =>
    mode === "table" ? setTableParamIndexes(value) : setQueryParamIndexes(value);
  const setFillError: React.Dispatch<React.SetStateAction<string | null>> = (
    value,
  ) =>
    mode === "table" ? setTableFillError(value) : setQueryFillError(value);

  const changeMode = (next: ToolType) => {
    setSuccess(false);
    setError(null);
    setTableFillError(null);
    setQueryFillError(null);
    setMode(next);
  };

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
    // Keep the other rows selected and shift the indexes after the removal.
    setParamIndexes((prev) =>
      prev
        .filter((i) => i !== index)
        .map((i) => (i > index ? i - 1 : i)),
    );
  };

  const toggleParam = (index: number, checked: boolean) => {
    setParamIndexes((prev) =>
      checked
        ? prev.includes(index)
          ? prev
          : [...prev, index]
        : prev.filter((i) => i !== index),
    );
  };

  // Ask the backend for table metadata and auto fill the column rows.
  const autoFillColumns = async () => {
    if (filling) return;
    setFillError(null);
    // Fail early when the needed form fields are empty.
    const missing =
      mode === "table"
        ? !formData.project.trim() ||
          !formData.dataset.trim() ||
          !formData.table_name.trim()
        : !(formData.query[0]?.query ?? "").trim();
    if (missing) {
      setFillError(
        mode === "table"
          ? "Please complete the project, dataset, and table name forms"
          : "Please complete the query form",
      );
      return;
    }
    setFilling(true);
    try {
      const { description, columns } = await fetchToolTableMetadata({
        type: mode,
        project: formData.project.trim(),
        dataset: formData.dataset.trim(),
        table_name: formData.table_name.trim(),
        builder_query: formData.query[0]?.query ?? "",
      });
      if (columns.length === 0) {
        setFillError("No columns found");
        return;
      }
      setFormData((prev) => ({
        ...prev,
        columns,
        // Table tool takes its description from the backend.
        tool_description:
          mode === "table" && description
            ? description
            : prev.tool_description,
      }));
      // Old param indexes no longer match the new rows.
      setParamIndexes([]);
    } catch (e) {
      if (e instanceof Error && e.message === "unauthorized") {
        navigate("/login");
        return;
      }
      setFillError("Failed to autofill columns");
    } finally {
      setFilling(false);
    }
  };

  const handleQueryChange = (
    index: number,
    field: keyof Query,
    value: string,
  ) => {
    const newQuery = [...formData.query];
    newQuery[index] = { ...newQuery[index], [field]: value };
    // Description or query changed, so the old verification no longer counts.
    if (field === "query" || field === "description") {
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
    // Table example needs both description and query.
    const missing =
      mode === "table"
        ? !formData.query[index].description.trim() || !text.trim()
        : !text.trim();
    if (missing) {
      setQueryStatus(
        index,
        text,
        "unverified",
        mode === "table"
          ? "Please complete the description and query forms in this example"
          : "Please complete the query form",
      );
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
      setQueryStatus(index, text, "unverified", "Failed to verify");
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
          ? "Query is verified"
          : status === "invalid"
            ? "Invalid query"
            : null);
    // Errors share the same text style as the autofill errors.
    const isError = status === "invalid" || !!message;
    const textClass = isError
      ? styles.error
      : `${styles.verifyStatus} ${
          status === "valid" ? styles.verifyOk : styles.verifyIdle
        }`;

    const verifyBtn = (
      <button
        type="button"
        className={`${styles.actionBtn} ${styles.verifyBtn}`}
        onClick={() => verifyQuery(index)}
        disabled={status === "checking"}
      >
        {status === "checking" ? "Verifying..." : "Verify"}
      </button>
    );
    const verifyMsg =
      text !== null ? <span className={textClass}>{text}</span> : null;

    return (
      <div className={styles.verifyRow}>
        {verifyBtn}
        {verifyMsg}
      </div>
    );
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(false);

    // Validate
    const missingTableName = mode === "table" && !formData.table_name.trim();
    const paramNames = paramIndexes
      .map((i) => formData.columns[i]?.name.trim() ?? "")
      .filter((name) => name.length > 0);
    const hasEmptyParam = paramIndexes.some(
      (i) => !(formData.columns[i]?.name.trim() ?? ""),
    );
    if (paramNames.length === 0 || hasEmptyParam) {
      setError("Select at least one column as param and give it a name");
      return;
    }
    if (
      !formData.tool_description.trim() ||
      (mode === "table" &&
        (!formData.project.trim() || !formData.dataset.trim())) ||
      missingTableName
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
        // Query tools have no location. Backend fills it later.
        project: mode === "table" ? formData.project : "",
        dataset: mode === "table" ? formData.dataset : "",
        table_name: mode === "table" ? formData.table_name : "",
        columns: formData.columns,
        type: mode,
        examples: formData.query.map(({ description, query }) => ({
          description,
          query,
        })),
        param_names: paramNames,
      });
      setSuccess(true);
      window.scrollTo(0, 0);
      // Reset both tabs.
      setTableData(emptyForm());
      setQueryData(emptyForm());
      setTableParamIndexes([]);
      setQueryParamIndexes([]);
      setTableFillError(null);
      setQueryFillError(null);
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
        {mode === "table" ? "Examples" : "Query"}
      </p>
      {mode === "table" ? (
        <div className={styles.greyCardList}>
          {formData.query.map((entry, index) => (
            <div
              key={index}
              className={`${styles.greyCard} ${styles.exampleCards}`}
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
        <button type="button" className={styles.actionBtn} onClick={addQuery}>
          + Add Example
        </button>
      )}
    </div>
  );

  const columnsForm = (
    <div className={styles.section}>
      <p className={styles.sectionHeader}>Columns</p>
      {mode === "query" && (
        <div className={styles.fillRow}>
          <button
            type="button"
            className={`${styles.actionBtn} ${styles.fillBtn}`}
            onClick={autoFillColumns}
            disabled={filling}
          >
            {filling ? "Autofilling..." : "Autofill"}
          </button>
          {fillError && <span className={styles.error}>{fillError}</span>}
        </div>
      )}
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
                    checked={paramIndexes.includes(index)}
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
      <button type="button" className={styles.actionBtn} onClick={addColumn}>
        + Add Column
      </button>
    </div>
  );

  const descriptionForm = (
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
  );

  return (
    <div className={styles.page}>
      {success && (
        <div className={styles.snackbar}>
          <span className={styles.snackbarText}>
            Tool has been added successfully! You can now use it in the query.
          </span>
          <span
            className={styles.snackbarClose}
            role="button"
            tabIndex={0}
            aria-label="Close"
            onClick={() => setSuccess(false)}
            onKeyDown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                setSuccess(false);
              }
            }}
          >
            ×
          </span>
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
            {/* Location: table tools only */}
            {mode === "table" && (
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
                </div>
                <div className={styles.fillRow}>
                  <button
                    type="button"
                    className={`${styles.actionBtn} ${styles.fillBtnWide}`}
                    onClick={autoFillColumns}
                    disabled={filling}
                  >
                    {filling
                      ? "Autofilling..."
                      : "Autofill description and columns"}
                  </button>
                  {fillError && (
                    <span className={styles.error}>{fillError}</span>
                  )}
                </div>
              </div>
            )}

            {/* Table: Description, Columns, Examples. Query: Query, Description, Columns. */}
            {mode === "table" ? (
              <>
                {descriptionForm}
                {columnsForm}
                {queryForm}
              </>
            ) : (
              <>
                {queryForm}
                {descriptionForm}
                {columnsForm}
              </>
            )}

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

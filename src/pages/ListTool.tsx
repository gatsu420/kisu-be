import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { fetchTool, toolTypeOf, type Tool, type ToolType } from "../lib/api";
import { highlightSql } from "../lib/highlight";
import Header from "../components/Header";
import styles from "./Tool.module.css";

export default function ListTool() {
  const navigate = useNavigate();
  const [mode, setMode] = useState<ToolType>("table");
  const [tool, setTool] = useState<Tool[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const changeMode = (next: ToolType) => {
    setMode(next);
  };

  const visibleTool = tool.filter((item) => toolTypeOf(item) === mode);

  useEffect(() => {
    let active = true;

    const load = async () => {
      setLoading(true);
      setError(null);
      try {
        const data = await fetchTool();
        if (active) setTool(data);
      } catch (e) {
        if (e instanceof Error && e.message === "unauthorized") {
          navigate("/login");
          return;
        }
        if (active) {
          setError(e instanceof Error ? e.message : "Failed to load tool");
        }
      } finally {
        if (active) setLoading(false);
      }
    };

    load();
    return () => {
      active = false;
    };
  }, [navigate]);

  return (
    <div className={styles.page}>
      <Header />

      <main className={styles.main}>
        <div className={styles.card}>
          <h2 className={styles.cardTitle}>List Tools</h2>

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

          {loading && <p className={styles.emptyState}>Loading tool...</p>}

          {error && <div className={styles.error}>{error}</div>}

          {!loading && !error && visibleTool.length === 0 && (
            <p className={styles.emptyState}>
              No tool yet. Add one from the Add Tool page.
            </p>
          )}

          {!loading && !error && visibleTool.length > 0 && (
            <div className={styles.toolList}>
              {visibleTool.map((item, index) => (
                <ToolCard key={index} tool={item} />
              ))}
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

interface ToolCardProps {
  tool: Tool;
}

function ToolCard({ tool }: ToolCardProps) {
  const type = toolTypeOf(tool);

  const locationSection =
    type === "table" ? (
      <div>
        <p className={styles.sectionTitle}>Location</p>
        <div className={styles.locationList}>
          <p className={styles.locationRow}>
            <span className={styles.locationLabel}>Project</span>
            <span className={styles.locationValue}>{tool.project}</span>
          </p>
          <p className={styles.locationRow}>
            <span className={styles.locationLabel}>Dataset</span>
            <span className={styles.locationValue}>{tool.dataset}</span>
          </p>
          <p className={styles.locationRow}>
            <span className={styles.locationLabel}>Table Name</span>
            <span className={styles.locationValue}>{tool.table_name}</span>
          </p>
        </div>
      </div>
    ) : null;

  const columnsSection =
    tool.columns.length > 0 ? (
      <div>
        <p className={styles.sectionTitle}>Columns</p>
        <div className={styles.tableWrap}>
          <table className={styles.table}>
            <thead>
              <tr>
                <th className={styles.th}>Name</th>
                <th className={styles.th}>Type</th>
                <th className={styles.th}>Description</th>
              </tr>
            </thead>
            <tbody>
              {tool.columns.map((col, i) => (
                <tr key={i} className={i % 2 === 0 ? styles.trEven : undefined}>
                  <td className={styles.td}>
                    {tool.param_name && col.name === tool.param_name
                      ? `${col.name} (param)`
                      : col.name}
                  </td>
                  <td className={styles.td}>{col.type}</td>
                  <td className={styles.td}>{col.description}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    ) : null;

  const examplesSection =
    tool.examples.length > 0 ? (
      <div>
        <p className={styles.sectionTitle}>
          {type === "table" ? "Examples" : "Query"}
        </p>
        <div className={styles.exampleLists}>
          {tool.examples.map((examples, i) =>
            type === "table" ? (
              <details key={i} className={styles.collapsible}>
                <summary className={styles.collapsibleTitle}>
                  {examples.description || `Example ${i + 1}`}
                </summary>
                <HighlightedCode code={examples.query} />
              </details>
            ) : (
              <div key={i} className={styles.examples}>
                {examples.description && (
                  <p className={styles.exampleDescriptions}>
                    {examples.description}
                  </p>
                )}
                <HighlightedCode code={examples.query} />
              </div>
            ),
          )}
        </div>
      </div>
    ) : null;

  const title = tool.tool_description || `${tool.project}.${tool.dataset}`;

  return (
    <details className={styles.toolItem}>
      <summary className={styles.collapsibleTitle}>{title}</summary>
      <div className={styles.toolCardBody}>
        {locationSection}

        {type === "table" ? (
          <>
            {columnsSection}
            {examplesSection}
          </>
        ) : (
          <>
            {examplesSection}
            {columnsSection}
          </>
        )}
      </div>
    </details>
  );
}

function HighlightedCode({ code }: { code: string }) {
  const [html, setHtml] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    highlightSql(code).then((result) => {
      if (active) setHtml(result);
    });
    return () => {
      active = false;
    };
  }, [code]);

  if (html) {
    return (
      <div
        className={styles.codeBlock}
        dangerouslySetInnerHTML={{ __html: html }}
      />
    );
  }

  return <div className={styles.codeBlock}>{code}</div>;
}

type WebBundle = typeof import("shiki/bundle/web");
type ShikiHighlighter = Awaited<ReturnType<WebBundle["createHighlighter"]>>;

const THEME = "github-light";
export type HighlightLang = "sql" | "json";
const LANGS: HighlightLang[] = ["sql", "json"];

let highlighterPromise: Promise<ShikiHighlighter> | null = null;

function loadHighlighter(): Promise<ShikiHighlighter> {
  highlighterPromise ??= import("shiki/bundle/web").then((mod) =>
    mod.createHighlighter({ themes: [THEME], langs: LANGS }),
  );
  return highlighterPromise;
}

/** Returns highlighted HTML, or null when highlighting is not ready. */
export async function highlightCode(
  code: string,
  lang: HighlightLang,
): Promise<string | null> {
  if (!code) return null;
  try {
    const highlighter = await loadHighlighter();
    return highlighter.codeToHtml(code, { lang, theme: THEME });
  } catch {
    return null;
  }
}

import { useEffect, useState } from "react";
import { highlightCode, type HighlightLang } from "../lib/highlight";

interface HighlightedCodeProps {
  code: string;
  lang: HighlightLang;
  className?: string;
}

// Shows shiki HTML when ready, plain text until then.
export default function HighlightedCode({
  code,
  lang,
  className,
}: HighlightedCodeProps) {
  const [html, setHtml] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    highlightCode(code, lang).then((result) => {
      if (active) setHtml(result);
    });
    return () => {
      active = false;
    };
  }, [code, lang]);

  if (html) {
    return (
      <div
        className={className}
        dangerouslySetInnerHTML={{ __html: html }}
      />
    );
  }

  return <div className={className}>{code}</div>;
}

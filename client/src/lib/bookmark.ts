// The BE sends updated_at as an RFC 3339 UTC string. The user's system
// timezone is applied here, at render time.
export function formatDayMonth(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  const day = d.getDate();
  const month = d.toLocaleString(undefined, { month: "short" });
  return `${day} ${month}`;
}

export function formatTime(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return "";
  return d.toLocaleTimeString(undefined, {
    hour: "2-digit",
    minute: "2-digit",
    hourCycle: "h23",
  });
}

export function updatedAtTime(iso: string): number {
  const t = Date.parse(iso);
  return Number.isNaN(t) ? 0 : t;
}

// Drawer item: "HH:MM · name".
export function bookmarkItemLabel(name: string, updatedAt: string): string {
  const time = formatTime(updatedAt);
  return time ? `${time} · ${name}` : name;
}

// Page title: "9 Oct HH:MM · name".
export function bookmarkTitle(name: string, updatedAt: string): string {
  const stamp = [formatDayMonth(updatedAt), formatTime(updatedAt)]
    .filter(Boolean)
    .join(" ");
  return stamp ? `${stamp} · ${name}` : name;
}

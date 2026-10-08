import { useEffect, useMemo, useState } from "react";
import { useNavigate, useLocation } from "react-router-dom";
import { type Bookmark } from "../lib/api";
import {
  bookmarkItemLabel,
  formatDayMonth,
  updatedAtTime,
} from "../lib/bookmark";
import { useBookmarks } from "./BookmarkContext";
import Logo from "./Logo";
import styles from "./Header.module.css";

export default function Header() {
  const navigate = useNavigate();
  const { pathname } = useLocation();
  const {
    bookmarks,
    bookmarksError,
    bookmarksLoading,
    bookmarksLoaded,
    refreshBookmarks,
  } = useBookmarks();
  const [bookmarksOpen, setBookmarksOpen] = useState(false);

  const isQuery = pathname === "/";
  const isTool = pathname.startsWith("/tool");

  // Group by day-month (newest day first), then by time (newest first).
  const bookmarkGroups = useMemo(() => {
    const groups = new Map<
      string,
      { label: string; sortKey: number; items: Bookmark[] }
    >();
    for (const b of bookmarks) {
      const label = formatDayMonth(b.updated_at);
      const sortKey = updatedAtTime(b.updated_at);
      const group = groups.get(label);
      if (group) {
        group.items.push(b);
        group.sortKey = Math.max(group.sortKey, sortKey);
      } else {
        groups.set(label, { label, sortKey, items: [b] });
      }
    }
    return Array.from(groups.values())
      .sort((a, b) => b.sortKey - a.sortKey)
      .map((group) => ({
        ...group,
        items: [...group.items].sort(
          (a, b) => updatedAtTime(b.updated_at) - updatedAtTime(a.updated_at),
        ),
      }));
  }, [bookmarks]);

  const toggleBookmarks = () => {
    const next = !bookmarksOpen;
    setBookmarksOpen(next);
    if (next && !bookmarksLoaded) void refreshBookmarks();
  };

  const selectBookmark = (bookmark: Bookmark) => {
    setBookmarksOpen(false);
    navigate("/", { state: { bookmark }, replace: pathname === "/" });
  };

  // Close the drawer with Escape.
  useEffect(() => {
    if (!bookmarksOpen) return;
    const onKey = (e: globalThis.KeyboardEvent) => {
      if (e.key === "Escape") setBookmarksOpen(false);
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [bookmarksOpen]);

  return (
    <>
      <header className={styles.header}>
        <button
          type="button"
          className={styles.menuBtn}
          aria-label="Bookmarks"
          aria-expanded={bookmarksOpen}
          onClick={toggleBookmarks}
        >
          ≡
        </button>
        <Logo />
        <nav className={styles.nav}>
          <button
            className={`${styles.navBtn} ${isQuery ? styles.navBtnActive : ""}`}
            onClick={() =>
              navigate("/", {
                state: { newQuery: true },
                replace: pathname === "/",
              })
            }
          >
            Query
          </button>
          <div className={styles.dropdown}>
            <button
              className={`${styles.navBtn} ${
                isTool ? styles.navBtnActive : ""
              }`}
              onClick={() => navigate("/tool/add")}
            >
              Tools
            </button>
            <div className={styles.dropdownMenu}>
              <button
                className={`${styles.dropdownItem} ${
                  pathname === "/tool/add" ? styles.dropdownItemActive : ""
                }`}
                onClick={() => navigate("/tool/add")}
              >
                Add New Tool
              </button>
              <button
                className={`${styles.dropdownItem} ${
                  pathname === "/tool/list" ? styles.dropdownItemActive : ""
                }`}
                onClick={() => navigate("/tool/list")}
              >
                List Tools
              </button>
            </div>
          </div>
        </nav>
        <button
          className={styles.signOutBtn}
          onClick={() => navigate("/login?signout")}
        >
          Sign out
        </button>
      </header>

      {bookmarksOpen && (
        <>
          <div
            className={styles.drawerBackdrop}
            onClick={() => setBookmarksOpen(false)}
          />
          <aside className={styles.drawer} aria-label="Bookmarks">
            <div className={styles.drawerHeader}>
              <button
                type="button"
                className={styles.menuBtn}
                aria-label="Close bookmarks"
                onClick={() => setBookmarksOpen(false)}
              >
                ×
              </button>
              <Logo />
            </div>
            <div className={styles.drawerBody}>
              <p className={styles.drawerTitle}>Bookmark</p>
              {bookmarksError ? (
                <p className={styles.drawerError}>{bookmarksError}</p>
              ) : bookmarksLoading ? (
                <p className={styles.drawerEmpty}>Loading...</p>
              ) : bookmarks.length === 0 ? (
                <p className={styles.drawerEmpty}>No bookmarks yet.</p>
              ) : (
                <ul className={styles.drawerList}>
                  {bookmarkGroups.map((group) => (
                    <li key={group.label || "unknown"} className={styles.drawerGroup}>
                      {group.label && (
                        <p className={styles.drawerGroupDate}>{group.label}</p>
                      )}
                      <ul className={styles.drawerGroupList}>
                        {group.items.map((b) => (
                          <li key={b.id}>
                            <button
                              type="button"
                              className={styles.drawerItem}
                              onClick={() => selectBookmark(b)}
                            >
                              {bookmarkItemLabel(
                                b.name || b.id,
                                b.updated_at,
                              )}
                            </button>
                          </li>
                        ))}
                      </ul>
                    </li>
                  ))}
                </ul>
              )}
            </div>
          </aside>
        </>
      )}
    </>
  );
}

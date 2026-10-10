import { useCallback, useState, type ReactNode } from "react";
import { useNavigate } from "react-router-dom";
import { fetchBookmarks, type Bookmark } from "../lib/api";
import { BookmarkContext } from "./BookmarkContext";

export function BookmarkProvider({ children }: { children: ReactNode }) {
  const navigate = useNavigate();
  const [bookmarks, setBookmarks] = useState<Bookmark[]>([]);
  const [bookmarksError, setBookmarksError] = useState<string | null>(null);
  const [bookmarksLoading, setBookmarksLoading] = useState(false);
  const [bookmarksLoaded, setBookmarksLoaded] = useState(false);

  const refreshBookmarks = useCallback(async () => {
    setBookmarksLoading(true);
    try {
      setBookmarks(await fetchBookmarks());
      setBookmarksError(null);
      setBookmarksLoaded(true);
    } catch (e) {
      if (e instanceof Error && e.message === "unauthorized") {
        navigate("/login");
        return;
      }
      setBookmarksError(
        e instanceof Error ? e.message : "failed to load bookmarks",
      );
    } finally {
      setBookmarksLoading(false);
    }
  }, [navigate]);

  // No fetch on mount. The drawer calls refreshBookmarks when it opens, so a
  // logged-out page load does not hit /answer and log a 401.
  return (
    <BookmarkContext.Provider
      value={{
        bookmarks,
        bookmarksError,
        bookmarksLoading,
        bookmarksLoaded,
        refreshBookmarks,
      }}
    >
      {children}
    </BookmarkContext.Provider>
  );
}

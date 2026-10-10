import { createContext, useContext } from "react";
import type { Bookmark } from "../lib/api";

export interface BookmarkContextValue {
  bookmarks: Bookmark[];
  bookmarksError: string | null;
  bookmarksLoading: boolean;
  bookmarksLoaded: boolean;
  refreshBookmarks: () => Promise<void>;
}

export const BookmarkContext = createContext<BookmarkContextValue | null>(null);

export function useBookmarks() {
  const ctx = useContext(BookmarkContext);
  if (!ctx) {
    throw new Error("useBookmarks must be used within a BookmarkProvider");
  }
  return ctx;
}

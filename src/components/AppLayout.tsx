import { Outlet } from "react-router-dom";
import { BookmarkProvider } from "./BookmarkProvider";

// Wraps the authenticated pages. The login page stays outside, so it does
// not fetch bookmarks before the user has a session cookie.
export default function AppLayout() {
  return (
    <BookmarkProvider>
      <Outlet />
    </BookmarkProvider>
  );
}

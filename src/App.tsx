import { BrowserRouter, Routes, Route, Navigate } from "react-router-dom";
import AppLayout from "./components/AppLayout";
import Login from "./pages/Login";
import Query from "./pages/Query";
import AddTool from "./pages/AddTool";
import ListTool from "./pages/ListTool";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<Login />} />
        <Route element={<AppLayout />}>
          <Route path="/" element={<Query />} />
          <Route path="/tool/add" element={<AddTool />} />
          <Route path="/tool/list" element={<ListTool />} />
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}

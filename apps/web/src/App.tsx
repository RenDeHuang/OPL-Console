import { BrowserRouter, Link, Route, Routes } from "react-router";
import { AdminOverviewPage } from "./pages/AdminOverviewPage";
import { LoginPage } from "./pages/LoginPage";
import { OwnerOverviewPage } from "./pages/OwnerOverviewPage";

export function App() {
  return (
    <BrowserRouter>
      <nav className="topbar">
        <Link to="/">OPL Console</Link>
        <Link to="/login">Login</Link>
        <Link to="/admin">Admin</Link>
      </nav>
      <Routes>
        <Route path="/" element={<OwnerOverviewPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/admin" element={<AdminOverviewPage />} />
      </Routes>
    </BrowserRouter>
  );
}

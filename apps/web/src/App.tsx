import { BrowserRouter, NavLink, Route, Routes } from "react-router";
import { LayoutDashboard, LockKeyhole, Shield } from "lucide-react";
import { AdminOverviewPage } from "./pages/AdminOverviewPage";
import { LoginPage } from "./pages/LoginPage";
import { OwnerOverviewPage } from "./pages/OwnerOverviewPage";

export function App() {
  return (
    <BrowserRouter>
      <nav className="topbar">
        <NavLink to="/" end>
          <LayoutDashboard size={16} />
          OPL Console
        </NavLink>
        <NavLink to="/login"><LockKeyhole size={16} /> Login</NavLink>
        <NavLink to="/admin"><Shield size={16} /> Admin</NavLink>
      </nav>
      <Routes>
        <Route path="/" element={<OwnerOverviewPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/admin" element={<AdminOverviewPage />} />
      </Routes>
    </BrowserRouter>
  );
}

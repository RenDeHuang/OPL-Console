import { BrowserRouter, NavLink, Route, Routes } from "react-router";
import { AdminOverviewPage } from "./pages/AdminOverviewPage";
import { LoginPage } from "./pages/LoginPage";
import { OwnerOverviewPage } from "./pages/OwnerOverviewPage";

export function App() {
  return (
    <BrowserRouter>
      <nav className="topbar">
        <NavLink to="/" end>
          OPL Console
        </NavLink>
        <NavLink to="/login">Login</NavLink>
        <NavLink to="/admin">Admin</NavLink>
      </nav>
      <Routes>
        <Route path="/" element={<OwnerOverviewPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/admin" element={<AdminOverviewPage />} />
      </Routes>
    </BrowserRouter>
  );
}

import { BrowserRouter, NavLink, Route, Routes } from "react-router";
import { CreditCard, HelpCircle, LayoutDashboard, LockKeyhole, Shield, Server, Stamp } from "lucide-react";
import { AdminOverviewPage } from "./pages/AdminOverviewPage";
import { ApprovalsPage } from "./pages/ApprovalsPage";
import { BillingPage } from "./pages/BillingPage";
import { LoginPage } from "./pages/LoginPage";
import { OwnerOverviewPage } from "./pages/OwnerOverviewPage";
import { PoliciesPage } from "./pages/PoliciesPage";
import { SupportPage } from "./pages/SupportPage";
import { WorkspacesPage } from "./pages/WorkspacesPage";

export function App() {
  return (
    <BrowserRouter>
      <nav className="topbar">
        <NavLink to="/" end>
          <LayoutDashboard size={16} />
          控制台
        </NavLink>
        <NavLink to="/workspaces"><Server size={16} /> 工作空间</NavLink>
        <NavLink to="/billing"><CreditCard size={16} /> 账单</NavLink>
        <NavLink to="/support"><HelpCircle size={16} /> 工单</NavLink>
        <NavLink to="/login"><LockKeyhole size={16} /> 登录</NavLink>
        <NavLink to="/admin"><Shield size={16} /> 管理</NavLink>
        <NavLink to="/admin/policies"><Stamp size={16} /> 策略</NavLink>
        <NavLink to="/admin/approvals"><Shield size={16} /> 审批</NavLink>
      </nav>
      <Routes>
        <Route path="/" element={<OwnerOverviewPage />} />
        <Route path="/workspaces" element={<WorkspacesPage />} />
        <Route path="/billing" element={<BillingPage />} />
        <Route path="/support" element={<SupportPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/admin" element={<AdminOverviewPage />} />
        <Route path="/admin/policies" element={<PoliciesPage />} />
        <Route path="/admin/approvals" element={<ApprovalsPage />} />
      </Routes>
    </BrowserRouter>
  );
}

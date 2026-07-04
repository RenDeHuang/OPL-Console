import { BrowserRouter, Navigate, NavLink, Route, Routes } from "react-router";
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
        <NavLink to="/console/overview">
          <LayoutDashboard size={16} />
          控制台
        </NavLink>
        <NavLink to="/console/workspaces"><Server size={16} /> 工作空间</NavLink>
        <NavLink to="/console/billing"><CreditCard size={16} /> 账单</NavLink>
        <NavLink to="/console/support"><HelpCircle size={16} /> 工单</NavLink>
        <NavLink to="/login"><LockKeyhole size={16} /> 登录</NavLink>
        <NavLink to="/admin/overview"><Shield size={16} /> 管理</NavLink>
        <NavLink to="/admin/ledger"><Stamp size={16} /> 策略</NavLink>
        <NavLink to="/admin/runtime"><Shield size={16} /> 审批</NavLink>
      </nav>
      <Routes>
        <Route path="/" element={<Navigate to="/console/overview" replace />} />
        <Route path="/console" element={<Navigate to="/console/overview" replace />} />
        <Route path="/console/overview" element={<OwnerOverviewPage />} />
        <Route path="/console/workspaces" element={<WorkspacesPage />} />
        <Route path="/console/billing" element={<BillingPage />} />
        <Route path="/console/billing/wallet" element={<BillingPage />} />
        <Route path="/console/support" element={<SupportPage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/admin" element={<Navigate to="/admin/overview" replace />} />
        <Route path="/admin/overview" element={<AdminOverviewPage />} />
        <Route path="/admin/users" element={<AdminOverviewPage />} />
        <Route path="/admin/billing" element={<BillingPage />} />
        <Route path="/admin/ledger" element={<PoliciesPage />} />
        <Route path="/admin/runtime" element={<ApprovalsPage />} />
        <Route path="/admin/support" element={<SupportPage />} />
      </Routes>
    </BrowserRouter>
  );
}

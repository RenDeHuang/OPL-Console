import type { ReactNode } from "react";
import { ProLayout } from "@ant-design/pro-components";
import { Button, Tag } from "antd";
import {
  Activity,
  Bell,
  Boxes,
  CreditCard,
  Database,
  FileText,
  Gauge,
  Headphones,
  KeyRound,
  Layers,
  LogOut,
  Server,
  ShieldCheck,
  UserRound,
  WalletCards,
} from "lucide-react";
import { Navigate, useLocation, useNavigate } from "react-router";
import { useQuery } from "@tanstack/react-query";
import { loadConsoleState, loadManagementState, loadOperatorSummary, loadTickets } from "../api/consoleApi";
import { adminMenuRoutes, findRoute, ownerMenuRoutes, routeTo, type OplRoute } from "../routes/oplRoutes";
import { AdminRoutePage, ConsoleRoutePage } from "./RoutePages";
import { ErrorPage, HomePage, LoginPage } from "./PublicPages";

function menuIcon(path: string) {
  const map: Record<string, ReactNode> = {
    "/console/overview": <Gauge size={17} />,
    "/console/compute": <Boxes size={17} />,
    "/console/storage": <Database size={17} />,
    "/console/attachments": <Layers size={17} />,
    "/console/workspaces": <Server size={17} />,
    "/console/gateway": <KeyRound size={17} />,
    "/console/billing": <WalletCards size={17} />,
    "/console/account": <UserRound size={17} />,
    "/console/support": <Headphones size={17} />,
    "/console/alerts": <Bell size={17} />,
    "/admin/overview": <Gauge size={17} />,
    "/admin/users": <UserRound size={17} />,
    "/admin/billing": <CreditCard size={17} />,
    "/admin/ledger": <Database size={17} />,
    "/admin/runtime": <Activity size={17} />,
    "/admin/support": <Headphones size={17} />,
  };
  return map[path] || <FileText size={17} />;
}

function buildMenu(isAdmin: boolean) {
  const owner = ownerMenuRoutes.map((route) => ({
    path: route.path,
    name: route.label,
    icon: menuIcon(route.path),
  }));
  const admin = isAdmin ? [{
    path: "/admin",
    name: "Admin",
    icon: <ShieldCheck size={17} />,
    children: adminMenuRoutes.map((route) => ({
      path: route.path,
      name: route.label,
      icon: menuIcon(route.path),
    })),
  }] : [];
  return [...owner, ...admin];
}

export function AppRouter() {
  const location = useLocation();
  const route = findRoute(location.pathname);

  if (route.redirectRouteId) return <Navigate to={routeTo(route.redirectRouteId)} replace />;
  if (route.area === "public") return <HomePage />;
  if (route.id === "auth.login" || route.id === "auth.logout") return <LoginPage route={route} />;
  if (route.id === "error.forbidden") return <ErrorPage code="403" title="无权限" />;
  if (route.id === "error.server") return <ErrorPage code="500" title="服务异常" />;
  if (route.id === "error.notFound") return <ErrorPage code="404" title="页面不存在" />;

  return <ConsoleShell route={route} />;
}

function ConsoleShell({ route }: { route: OplRoute }) {
  const navigate = useNavigate();
  const location = useLocation();
  const isAdmin = route.area === "admin";
  const state = useQuery({ queryKey: ["console-state"], queryFn: loadConsoleState });
  const management = useQuery({ queryKey: ["management-state"], queryFn: loadManagementState });
  const operator = useQuery({ queryKey: ["operator-summary"], queryFn: loadOperatorSummary });
  const tickets = useQuery({ queryKey: ["support-tickets"], queryFn: loadTickets });

  const ctx = {
    route,
    path: location.pathname,
    state: state.data,
    wallet: state.data?.wallet,
    management: management.data,
    operator: operator.data,
    tickets: tickets.data,
  };

  if (!state.data) return <div className="loading">Loading OPL Console...</div>;

  return (
    <ProLayout
      title="OPL Console"
      logo={<div className="proLogo">OPL</div>}
      location={{ pathname: location.pathname }}
      layout="mix"
      navTheme="light"
      ErrorBoundary={false}
      menuDataRender={() => buildMenu(isAdmin)}
      menuItemRender={(item, dom) => (
        <a
          href={item.path}
          onClick={(event) => {
            event.preventDefault();
            navigate(item.path || routeTo("console.overview"));
          }}
        >
          {dom}
        </a>
      )}
      actionsRender={() => [
        <Tag color={isAdmin ? "purple" : "blue"} key="role">{isAdmin ? "Admin" : "Lab Owner"}</Tag>,
        <Button key="logout" icon={<LogOut size={15} />} onClick={() => navigate("/")}>退出</Button>,
      ]}
      avatarProps={{
        title: <span className="shellEmail">{isAdmin ? "admin@opl.local" : "owner@opl.local"}</span>,
        size: "small",
        icon: <UserRound size={16} />,
      }}
    >
      {isAdmin ? (
        <AdminRoutePage {...ctx} state={state.data} />
      ) : (
        <ConsoleRoutePage {...ctx} state={state.data} />
      )}
    </ProLayout>
  );
}

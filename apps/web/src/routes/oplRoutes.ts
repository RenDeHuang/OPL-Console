export type ActorRole = "public" | "lab_owner" | "admin";
export type RouteArea = "public" | "auth" | "console" | "admin" | "error";

export type OplRoute = {
  id: string;
  path: string;
  label: string;
  area: RouteArea;
  role: ActorRole;
  routeKind: string;
  objectKind?: string;
  menu?: boolean;
  adminMenu?: boolean;
  hiddenInMenu?: boolean;
  redirectRouteId?: string;
  status?: "implemented" | "folded_into_parent" | "external";
  contractLifecycle?: "current" | "folded_parent";
  serviceBoundary: string;
  apiRoutes?: string[];
  capabilities?: string[];
};

function currentRoute(route: Omit<OplRoute, "status" | "contractLifecycle"> & Partial<Pick<OplRoute, "status" | "contractLifecycle">>): OplRoute {
  return {
    status: "implemented",
    contractLifecycle: "current",
    capabilities: ["read"],
    ...route,
  };
}

export const oplRoutes = Object.freeze([
  currentRoute({ id: "public.home", path: "/", label: "首页", area: "public", role: "public", routeKind: "static_content", serviceBoundary: "StaticPublicContent" }),
  currentRoute({ id: "public.pricing", path: "/pricing", label: "价格", area: "public", role: "public", status: "folded_into_parent", contractLifecycle: "folded_parent", routeKind: "static_content", serviceBoundary: "StaticPublicContent" }),
  currentRoute({ id: "public.docs", path: "/docs", label: "文档", area: "public", role: "public", status: "folded_into_parent", contractLifecycle: "folded_parent", routeKind: "static_content", serviceBoundary: "StaticPublicContent" }),
  currentRoute({ id: "public.status", path: "/status", label: "服务状态", area: "public", role: "public", status: "folded_into_parent", contractLifecycle: "folded_parent", routeKind: "static_content", serviceBoundary: "StaticPublicContent" }),
  currentRoute({ id: "auth.login", path: "/login", label: "登录", area: "auth", role: "public", routeKind: "auth_flow", serviceBoundary: "AuthController", apiRoutes: ["POST /api/auth/login"], capabilities: ["read", "authenticate", "session"] }),
  currentRoute({ id: "auth.logout", path: "/logout", label: "退出登录", area: "auth", role: "lab_owner", hiddenInMenu: true, routeKind: "auth_flow", serviceBoundary: "AuthController", apiRoutes: ["POST /api/auth/logout"], capabilities: ["read", "session"] }),
  currentRoute({ id: "console.root", path: "/console", label: "控制台", area: "console", role: "lab_owner", redirectRouteId: "console.overview", hiddenInMenu: true, status: "folded_into_parent", contractLifecycle: "folded_parent", routeKind: "read_model", serviceBoundary: "ConsoleRouter" }),
  currentRoute({ id: "console.overview", path: "/console/overview", label: "概览", area: "console", role: "lab_owner", menu: true, routeKind: "read_model", serviceBoundary: "ConsoleReadModelService", apiRoutes: ["GET /api/state", "GET /api/support/tickets"], capabilities: ["read", "summary"] }),
  currentRoute({ id: "compute-pools.list", path: "/console/compute/pools", label: "计算资源池", area: "console", role: "admin", hiddenInMenu: true, routeKind: "read_model", objectKind: "ComputePool", serviceBoundary: "ComputePoolCatalogService", apiRoutes: ["GET /api/compute-pools"], capabilities: ["list", "read", "evidence"] }),
  currentRoute({ id: "compute-allocations.list", path: "/console/compute", label: "计算资源", area: "console", role: "lab_owner", menu: true, routeKind: "business_object", objectKind: "ComputeAllocation", serviceBoundary: "ComputeAllocationService", apiRoutes: ["GET /api/state", "GET /api/compute-allocations"], capabilities: ["list", "read", "evidence"] }),
  currentRoute({ id: "compute-allocations.create", path: "/console/compute/new", label: "开通计算资源", area: "console", role: "lab_owner", hiddenInMenu: true, routeKind: "business_object", objectKind: "ComputeAllocation", serviceBoundary: "ComputeAllocationService", apiRoutes: ["GET /api/compute-pools", "POST /api/compute-allocations"], capabilities: ["read", "write", "action", "evidence"] }),
  currentRoute({ id: "compute-allocations.detail", path: "/console/compute/:id", label: "计算资源详情", area: "console", role: "lab_owner", hiddenInMenu: true, routeKind: "business_object", objectKind: "ComputeAllocation", serviceBoundary: "ComputeAllocationService", apiRoutes: ["GET /api/compute-allocations/:id", "POST /api/compute-allocations/:id/destroy"], capabilities: ["detail", "read", "action", "evidence"] }),
  currentRoute({ id: "storage.list", path: "/console/storage", label: "存储资源", area: "console", role: "lab_owner", menu: true, routeKind: "business_object", objectKind: "StorageVolume", serviceBoundary: "ResourceProvisioningService", apiRoutes: ["GET /api/state"], capabilities: ["list", "read", "action", "evidence"] }),
  currentRoute({ id: "storage.create", path: "/console/storage/new", label: "开通存储资源", area: "console", role: "lab_owner", hiddenInMenu: true, routeKind: "business_object", objectKind: "StorageVolume", serviceBoundary: "ResourceProvisioningService", apiRoutes: ["GET /api/state", "POST /api/storage-volumes"], capabilities: ["read", "write", "action", "evidence"] }),
  currentRoute({ id: "storage.detail", path: "/console/storage/:id", label: "存储资源详情", area: "console", role: "lab_owner", hiddenInMenu: true, routeKind: "business_object", objectKind: "StorageVolume", serviceBoundary: "ResourceProvisioningService", apiRoutes: ["GET /api/state", "POST /api/storage-volumes/destroy"], capabilities: ["detail", "read", "action", "evidence"] }),
  currentRoute({ id: "attachment.list", path: "/console/attachments", label: "挂载关系", area: "console", role: "lab_owner", menu: true, routeKind: "business_object", objectKind: "StorageAttachment", serviceBoundary: "ResourceProvisioningService", apiRoutes: ["GET /api/state"], capabilities: ["list", "read", "action", "evidence"] }),
  currentRoute({ id: "attachment.create", path: "/console/attachments/new", label: "挂载存储", area: "console", role: "lab_owner", hiddenInMenu: true, routeKind: "business_object", objectKind: "StorageAttachment", serviceBoundary: "ResourceProvisioningService", apiRoutes: ["GET /api/state", "POST /api/storage-attachments"], capabilities: ["read", "write", "action", "evidence"] }),
  currentRoute({ id: "attachment.detail", path: "/console/attachments/:id", label: "挂载详情", area: "console", role: "lab_owner", hiddenInMenu: true, routeKind: "business_object", objectKind: "StorageAttachment", serviceBoundary: "ResourceProvisioningService", apiRoutes: ["GET /api/state", "POST /api/storage-attachments/detach"], capabilities: ["detail", "read", "action", "evidence"] }),
  currentRoute({ id: "resources.relationships", path: "/console/resources/relationships", label: "资源关系", area: "console", role: "lab_owner", menu: true, routeKind: "read_model", objectKind: "ResourceRelationship", serviceBoundary: "ConsoleReadModelService", apiRoutes: ["GET /api/state"], capabilities: ["read", "summary", "evidence"] }),
  currentRoute({ id: "workspace.list", path: "/console/workspaces", label: "工作区入口", area: "console", role: "lab_owner", menu: true, routeKind: "business_object", objectKind: "Workspace", serviceBoundary: "WorkspaceEntryService", apiRoutes: ["GET /api/state", "POST /api/workspaces/reset-token", "POST /api/workspaces/delete-token"], capabilities: ["list", "read", "action"] }),
  currentRoute({ id: "workspace.create", path: "/console/workspaces/new", label: "创建工作区入口", area: "console", role: "lab_owner", hiddenInMenu: true, routeKind: "business_object", objectKind: "Workspace", serviceBoundary: "WorkspaceEntryService", apiRoutes: ["GET /api/state", "POST /api/workspaces"], capabilities: ["read", "write"] }),
  currentRoute({ id: "workspace.detail", path: "/console/workspaces/:id", label: "工作区入口详情", area: "console", role: "lab_owner", hiddenInMenu: true, routeKind: "business_object", objectKind: "Workspace", serviceBoundary: "WorkspaceEntryService", apiRoutes: ["GET /api/state", "POST /api/workspaces/reset-token", "POST /api/workspaces/delete-token", "POST /api/workspaces/runtime-status"], capabilities: ["detail", "read", "action", "evidence"] }),
  currentRoute({ id: "gateway.external", path: "/console/gateway", label: "网关", area: "console", role: "lab_owner", menu: true, status: "external", routeKind: "external_integration", objectKind: "GatewayIntegration", serviceBoundary: "OPLGatewayExternalIntegration", capabilities: ["read", "external_link"] }),
  currentRoute({ id: "billing.overview", path: "/console/billing", label: "账单", area: "console", role: "lab_owner", menu: true, routeKind: "read_model", objectKind: "Wallet", serviceBoundary: "WalletService", apiRoutes: ["GET /api/state"], capabilities: ["read", "list", "detail"] }),
  currentRoute({ id: "billing.wallet", path: "/console/billing/wallet", label: "钱包与冻结", area: "console", role: "lab_owner", hiddenInMenu: true, status: "folded_into_parent", contractLifecycle: "folded_parent", routeKind: "read_model", objectKind: "Wallet", serviceBoundary: "WalletService", apiRoutes: ["GET /api/state"], capabilities: ["read", "detail"] }),
  currentRoute({ id: "account.overview", path: "/console/account", label: "账号与实验室", area: "console", role: "lab_owner", menu: true, routeKind: "read_model", objectKind: "User", serviceBoundary: "ManagementModel", apiRoutes: ["GET /api/state", "GET /api/auth/me"], capabilities: ["read", "detail"] }),
  currentRoute({ id: "support.list", path: "/console/support", label: "工单", area: "console", role: "lab_owner", menu: true, routeKind: "business_object", objectKind: "SupportTicket", serviceBoundary: "SupportTicketService", apiRoutes: ["GET /api/support/tickets"], capabilities: ["list", "read", "audit"] }),
  currentRoute({ id: "support.create", path: "/console/support/new", label: "新建工单", area: "console", role: "lab_owner", hiddenInMenu: true, routeKind: "business_object", objectKind: "SupportTicket", serviceBoundary: "SupportTicketService", apiRoutes: ["GET /api/support/tickets", "POST /api/support/tickets"], capabilities: ["read", "write", "audit"] }),
  currentRoute({ id: "support.detail", path: "/console/support/:id", label: "工单详情", area: "console", role: "lab_owner", hiddenInMenu: true, routeKind: "business_object", objectKind: "SupportTicket", serviceBoundary: "SupportTicketService", apiRoutes: ["GET /api/support/tickets"], capabilities: ["detail", "read", "audit"] }),
  currentRoute({ id: "alerts.list", path: "/console/alerts", label: "提醒", area: "console", role: "lab_owner", menu: true, routeKind: "read_model", objectKind: "RuntimeReadiness", serviceBoundary: "ConsoleReadModelService", apiRoutes: ["GET /api/state", "GET /api/support/tickets"], capabilities: ["read", "list"] }),
  currentRoute({ id: "admin.root", path: "/admin", label: "管理", area: "admin", role: "admin", redirectRouteId: "admin.overview", hiddenInMenu: true, status: "folded_into_parent", contractLifecycle: "folded_parent", routeKind: "read_model", serviceBoundary: "ConsoleRouter" }),
  currentRoute({ id: "admin.overview", path: "/admin/overview", label: "管理概览", area: "admin", role: "admin", adminMenu: true, routeKind: "read_model", serviceBoundary: "ConsoleReadModelService", apiRoutes: ["GET /api/state", "GET /api/operator/summary"], capabilities: ["read", "summary"] }),
  currentRoute({ id: "admin.users", path: "/admin/users", label: "用户", area: "admin", role: "admin", adminMenu: true, routeKind: "read_model", objectKind: "User", serviceBoundary: "ManagementModel", apiRoutes: ["GET /api/management/state", "POST /api/users", "POST /api/users/disable", "POST /api/users/delete", "POST /api/billing/topups"], capabilities: ["list", "read", "action", "audit"] }),
  currentRoute({ id: "admin.billing", path: "/admin/billing", label: "账单运营", area: "admin", role: "admin", adminMenu: true, routeKind: "read_model", objectKind: "Wallet", serviceBoundary: "WalletService", apiRoutes: ["GET /api/state", "POST /api/billing/topups"], capabilities: ["read", "list", "action", "audit"] }),
  currentRoute({ id: "admin.ledger", path: "/admin/ledger", label: "账本", area: "admin", role: "admin", adminMenu: true, routeKind: "read_model", objectKind: "Usage", serviceBoundary: "LedgerEvidenceService", apiRoutes: ["GET /api/state", "GET /api/ledger/task-receipts"], capabilities: ["read", "list", "evidence", "audit"] }),
  currentRoute({ id: "admin.runtime", path: "/admin/runtime", label: "运行状态", area: "admin", role: "admin", adminMenu: true, routeKind: "read_model", objectKind: "RuntimeReadiness", serviceBoundary: "RuntimeOperationService", apiRoutes: ["GET /api/runtime/readiness", "GET /api/production/readiness", "GET /api/operator/summary"], capabilities: ["read", "detail", "audit"] }),
  currentRoute({ id: "admin.diagnostics", path: "/admin/diagnostics", label: "线上诊断", area: "admin", role: "admin", adminMenu: true, routeKind: "read_model", objectKind: "RuntimeReadiness", serviceBoundary: "RuntimeOperationService", apiRoutes: ["GET /api/runtime/readiness", "GET /api/production/readiness", "GET /api/operator/summary"], capabilities: ["read", "detail", "audit"] }),
  currentRoute({ id: "admin.e2e", path: "/admin/e2e", label: "E2E记录", area: "admin", role: "admin", adminMenu: true, routeKind: "read_model", objectKind: "ProductionVerification", serviceBoundary: "RuntimeOperationService", apiRoutes: ["GET /api/operator/summary"], capabilities: ["read", "summary", "audit"] }),
  currentRoute({ id: "admin.cleanup", path: "/admin/cleanup", label: "入口清理", area: "admin", role: "admin", adminMenu: true, routeKind: "read_model", objectKind: "RuntimeReadiness", serviceBoundary: "WorkspaceEntryService", apiRoutes: ["GET /api/management/state", "GET /api/operator/summary", "POST /api/operator/cleanup-workspace-access"], capabilities: ["read", "action", "audit"] }),
  currentRoute({ id: "admin.support", path: "/admin/support", label: "工单运营", area: "admin", role: "admin", adminMenu: true, routeKind: "business_object", objectKind: "SupportTicket", serviceBoundary: "SupportTicketService", apiRoutes: ["GET /api/support/tickets"], capabilities: ["list", "read", "audit"] }),
  currentRoute({ id: "error.forbidden", path: "/403", label: "无权限", area: "error", role: "public", hiddenInMenu: true, routeKind: "static_content", serviceBoundary: "ConsoleRouter" }),
  currentRoute({ id: "error.notFound", path: "/404", label: "未找到", area: "error", role: "public", hiddenInMenu: true, routeKind: "static_content", serviceBoundary: "ConsoleRouter" }),
  currentRoute({ id: "error.server", path: "/500", label: "错误", area: "error", role: "public", hiddenInMenu: true, routeKind: "static_content", serviceBoundary: "ConsoleRouter" }),
] satisfies OplRoute[]);

export const routesById = new Map(oplRoutes.map((route) => [route.id, route]));

function escapeRegex(value: string) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
}

function routePattern(path: string) {
  return new RegExp(`^${path.split("/").map((part) => (part.startsWith(":") ? "[^/]+" : escapeRegex(part))).join("/")}$`);
}

export function normalizePath(pathname: string) {
  if (!pathname || pathname === "") return "/";
  return pathname.length > 1 ? pathname.replace(/\/+$/, "") : pathname;
}

export function routeTo(id: string, params: Record<string, string> = {}) {
  const route = routesById.get(id);
  if (!route) throw new Error(`unknown route id: ${id}`);
  return route.path.replace(/:([^/]+)/g, (_, key: string) => encodeURIComponent(params[key] ?? ""));
}

export function findRoute(pathname: string): OplRoute {
  const normalized = normalizePath(pathname);
  const route = oplRoutes.find((entry) => entry.path === normalized)
    ?? oplRoutes.find((entry) => entry.path.includes(":") && routePattern(entry.path).test(normalized));
  return route ?? routesById.get("error.notFound")!;
}

export function menuRoutesFor(role: "lab_owner" | "admin") {
  return oplRoutes.filter((route) => (role === "admin" ? route.adminMenu : route.menu));
}

export const ownerMenuRoutes = menuRoutesFor("lab_owner");
export const adminMenuRoutes = menuRoutesFor("admin");

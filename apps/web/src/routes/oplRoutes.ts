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
  currentRoute({ id: "public.home", path: "/", label: "Public Home", area: "public", role: "public", routeKind: "static_content", serviceBoundary: "StaticPublicContent" }),
  currentRoute({ id: "public.pricing", path: "/pricing", label: "Pricing", area: "public", role: "public", status: "folded_into_parent", contractLifecycle: "folded_parent", routeKind: "static_content", serviceBoundary: "StaticPublicContent" }),
  currentRoute({ id: "public.docs", path: "/docs", label: "Docs", area: "public", role: "public", status: "folded_into_parent", contractLifecycle: "folded_parent", routeKind: "static_content", serviceBoundary: "StaticPublicContent" }),
  currentRoute({ id: "public.status", path: "/status", label: "Service Status", area: "public", role: "public", status: "folded_into_parent", contractLifecycle: "folded_parent", routeKind: "static_content", serviceBoundary: "StaticPublicContent" }),
  currentRoute({ id: "auth.login", path: "/login", label: "Login", area: "auth", role: "public", routeKind: "auth_flow", serviceBoundary: "AuthController", apiRoutes: ["POST /api/auth/login"], capabilities: ["read", "authenticate", "session"] }),
  currentRoute({ id: "auth.logout", path: "/logout", label: "Logout", area: "auth", role: "lab_owner", hiddenInMenu: true, routeKind: "auth_flow", serviceBoundary: "AuthController", apiRoutes: ["POST /api/auth/logout"], capabilities: ["read", "session"] }),
  currentRoute({ id: "console.root", path: "/console", label: "Console", area: "console", role: "lab_owner", redirectRouteId: "console.overview", hiddenInMenu: true, status: "folded_into_parent", contractLifecycle: "folded_parent", routeKind: "read_model", serviceBoundary: "ConsoleRouter" }),
  currentRoute({ id: "console.overview", path: "/console/overview", label: "Overview", area: "console", role: "lab_owner", menu: true, routeKind: "read_model", serviceBoundary: "ConsoleReadModelService", apiRoutes: ["GET /api/state", "GET /api/support/tickets"], capabilities: ["read", "summary"] }),
  currentRoute({ id: "compute-pools.list", path: "/console/compute/pools", label: "Compute Pools", area: "console", role: "admin", hiddenInMenu: true, routeKind: "read_model", objectKind: "ComputePool", serviceBoundary: "ComputePoolCatalogService", apiRoutes: ["GET /api/compute-pools"], capabilities: ["list", "read", "evidence"] }),
  currentRoute({ id: "compute-allocations.list", path: "/console/compute", label: "Compute", area: "console", role: "lab_owner", menu: true, routeKind: "business_object", objectKind: "ComputeAllocation", serviceBoundary: "ComputeAllocationService", apiRoutes: ["GET /api/state", "GET /api/compute-allocations"], capabilities: ["list", "read", "evidence"] }),
  currentRoute({ id: "compute-allocations.create", path: "/console/compute/new", label: "Create Compute Allocation", area: "console", role: "lab_owner", hiddenInMenu: true, routeKind: "business_object", objectKind: "ComputeAllocation", serviceBoundary: "ComputeAllocationService", apiRoutes: ["GET /api/compute-pools", "POST /api/compute-allocations"], capabilities: ["read", "write", "action", "evidence"] }),
  currentRoute({ id: "compute-allocations.detail", path: "/console/compute/:id", label: "Compute Allocation Detail", area: "console", role: "lab_owner", hiddenInMenu: true, routeKind: "business_object", objectKind: "ComputeAllocation", serviceBoundary: "ComputeAllocationService", apiRoutes: ["GET /api/compute-allocations/:id", "POST /api/compute-allocations/:id/destroy"], capabilities: ["detail", "read", "action", "evidence"] }),
  currentRoute({ id: "storage.list", path: "/console/storage", label: "Storage", area: "console", role: "lab_owner", menu: true, routeKind: "business_object", objectKind: "StorageVolume", serviceBoundary: "ResourceProvisioningService", apiRoutes: ["GET /api/state"], capabilities: ["list", "read", "action", "evidence"] }),
  currentRoute({ id: "storage.create", path: "/console/storage/new", label: "Create Storage", area: "console", role: "lab_owner", hiddenInMenu: true, routeKind: "business_object", objectKind: "StorageVolume", serviceBoundary: "ResourceProvisioningService", apiRoutes: ["GET /api/state", "POST /api/storage-volumes"], capabilities: ["read", "write", "action", "evidence"] }),
  currentRoute({ id: "storage.detail", path: "/console/storage/:id", label: "Storage Detail", area: "console", role: "lab_owner", hiddenInMenu: true, routeKind: "business_object", objectKind: "StorageVolume", serviceBoundary: "ResourceProvisioningService", apiRoutes: ["GET /api/state", "POST /api/storage-volumes/destroy"], capabilities: ["detail", "read", "action", "evidence"] }),
  currentRoute({ id: "attachment.list", path: "/console/attachments", label: "Attachments", area: "console", role: "lab_owner", menu: true, routeKind: "business_object", objectKind: "StorageAttachment", serviceBoundary: "ResourceProvisioningService", apiRoutes: ["GET /api/state"], capabilities: ["list", "read", "action", "evidence"] }),
  currentRoute({ id: "attachment.create", path: "/console/attachments/new", label: "Attach Storage", area: "console", role: "lab_owner", hiddenInMenu: true, routeKind: "business_object", objectKind: "StorageAttachment", serviceBoundary: "ResourceProvisioningService", apiRoutes: ["GET /api/state", "POST /api/storage-attachments"], capabilities: ["read", "write", "action", "evidence"] }),
  currentRoute({ id: "attachment.detail", path: "/console/attachments/:id", label: "Attachment Detail", area: "console", role: "lab_owner", hiddenInMenu: true, routeKind: "business_object", objectKind: "StorageAttachment", serviceBoundary: "ResourceProvisioningService", apiRoutes: ["GET /api/state", "POST /api/storage-attachments/detach"], capabilities: ["detail", "read", "action", "evidence"] }),
  currentRoute({ id: "workspace.list", path: "/console/workspaces", label: "Workspaces", area: "console", role: "lab_owner", menu: true, routeKind: "business_object", objectKind: "Workspace", serviceBoundary: "WorkspaceEntryService", apiRoutes: ["GET /api/state", "POST /api/workspaces/reset-token", "POST /api/workspaces/delete-token"], capabilities: ["list", "read", "action"] }),
  currentRoute({ id: "workspace.create", path: "/console/workspaces/new", label: "Create Workspace", area: "console", role: "lab_owner", hiddenInMenu: true, routeKind: "business_object", objectKind: "Workspace", serviceBoundary: "WorkspaceEntryService", apiRoutes: ["GET /api/state", "POST /api/workspaces"], capabilities: ["read", "write"] }),
  currentRoute({ id: "workspace.detail", path: "/console/workspaces/:id", label: "Workspace Detail", area: "console", role: "lab_owner", hiddenInMenu: true, routeKind: "business_object", objectKind: "Workspace", serviceBoundary: "WorkspaceEntryService", apiRoutes: ["GET /api/state", "POST /api/workspaces/reset-token", "POST /api/workspaces/delete-token", "POST /api/workspaces/runtime-status"], capabilities: ["detail", "read", "action", "evidence"] }),
  currentRoute({ id: "gateway.external", path: "/console/gateway", label: "Gateway", area: "console", role: "lab_owner", menu: true, status: "external", routeKind: "external_integration", objectKind: "GatewayIntegration", serviceBoundary: "OPLGatewayExternalIntegration", capabilities: ["read", "external_link"] }),
  currentRoute({ id: "billing.overview", path: "/console/billing", label: "Billing", area: "console", role: "lab_owner", menu: true, routeKind: "read_model", objectKind: "Wallet", serviceBoundary: "WalletService", apiRoutes: ["GET /api/state"], capabilities: ["read", "list", "detail"] }),
  currentRoute({ id: "billing.wallet", path: "/console/billing/wallet", label: "Wallet and Holds", area: "console", role: "lab_owner", hiddenInMenu: true, status: "folded_into_parent", contractLifecycle: "folded_parent", routeKind: "read_model", objectKind: "Wallet", serviceBoundary: "WalletService", apiRoutes: ["GET /api/state"], capabilities: ["read", "detail"] }),
  currentRoute({ id: "account.overview", path: "/console/account", label: "Account & Lab", area: "console", role: "lab_owner", menu: true, routeKind: "read_model", objectKind: "User", serviceBoundary: "ManagementModel", apiRoutes: ["GET /api/state", "GET /api/auth/me"], capabilities: ["read", "detail"] }),
  currentRoute({ id: "support.list", path: "/console/support", label: "Support", area: "console", role: "lab_owner", menu: true, routeKind: "business_object", objectKind: "SupportTicket", serviceBoundary: "SupportTicketService", apiRoutes: ["GET /api/support/tickets"], capabilities: ["list", "read", "audit"] }),
  currentRoute({ id: "support.create", path: "/console/support/new", label: "New Ticket", area: "console", role: "lab_owner", hiddenInMenu: true, routeKind: "business_object", objectKind: "SupportTicket", serviceBoundary: "SupportTicketService", apiRoutes: ["GET /api/support/tickets", "POST /api/support/tickets"], capabilities: ["read", "write", "audit"] }),
  currentRoute({ id: "support.detail", path: "/console/support/:id", label: "Ticket Detail", area: "console", role: "lab_owner", hiddenInMenu: true, routeKind: "business_object", objectKind: "SupportTicket", serviceBoundary: "SupportTicketService", apiRoutes: ["GET /api/support/tickets"], capabilities: ["detail", "read", "audit"] }),
  currentRoute({ id: "alerts.list", path: "/console/alerts", label: "Alerts", area: "console", role: "lab_owner", menu: true, routeKind: "read_model", objectKind: "RuntimeReadiness", serviceBoundary: "ConsoleReadModelService", apiRoutes: ["GET /api/state", "GET /api/support/tickets"], capabilities: ["read", "list"] }),
  currentRoute({ id: "admin.root", path: "/admin", label: "Admin", area: "admin", role: "admin", redirectRouteId: "admin.overview", hiddenInMenu: true, status: "folded_into_parent", contractLifecycle: "folded_parent", routeKind: "read_model", serviceBoundary: "ConsoleRouter" }),
  currentRoute({ id: "admin.overview", path: "/admin/overview", label: "Admin Overview", area: "admin", role: "admin", adminMenu: true, routeKind: "read_model", serviceBoundary: "ConsoleReadModelService", apiRoutes: ["GET /api/state", "GET /api/operator/summary"], capabilities: ["read", "summary"] }),
  currentRoute({ id: "admin.users", path: "/admin/users", label: "Users", area: "admin", role: "admin", adminMenu: true, routeKind: "read_model", objectKind: "User", serviceBoundary: "ManagementModel", apiRoutes: ["GET /api/management/state", "POST /api/users", "POST /api/billing/topups"], capabilities: ["list", "read", "action", "audit"] }),
  currentRoute({ id: "admin.billing", path: "/admin/billing", label: "Billing Ops", area: "admin", role: "admin", adminMenu: true, routeKind: "read_model", objectKind: "Wallet", serviceBoundary: "WalletService", apiRoutes: ["GET /api/state", "POST /api/billing/topups"], capabilities: ["read", "list", "action", "audit"] }),
  currentRoute({ id: "admin.ledger", path: "/admin/ledger", label: "Ledger", area: "admin", role: "admin", adminMenu: true, routeKind: "read_model", objectKind: "Usage", serviceBoundary: "LedgerEvidenceService", apiRoutes: ["GET /api/state", "GET /api/ledger/task-receipts"], capabilities: ["read", "list", "evidence", "audit"] }),
  currentRoute({ id: "admin.runtime", path: "/admin/runtime", label: "Runtime", area: "admin", role: "admin", adminMenu: true, routeKind: "read_model", objectKind: "RuntimeReadiness", serviceBoundary: "RuntimeOperationService", apiRoutes: ["GET /api/runtime/readiness", "GET /api/production/readiness", "GET /api/operator/summary"], capabilities: ["read", "detail", "audit"] }),
  currentRoute({ id: "admin.support", path: "/admin/support", label: "Support Ops", area: "admin", role: "admin", adminMenu: true, routeKind: "business_object", objectKind: "SupportTicket", serviceBoundary: "SupportTicketService", apiRoutes: ["GET /api/support/tickets"], capabilities: ["list", "read", "audit"] }),
  currentRoute({ id: "error.forbidden", path: "/403", label: "Forbidden", area: "error", role: "public", hiddenInMenu: true, routeKind: "static_content", serviceBoundary: "ConsoleRouter" }),
  currentRoute({ id: "error.notFound", path: "/404", label: "Not Found", area: "error", role: "public", hiddenInMenu: true, routeKind: "static_content", serviceBoundary: "ConsoleRouter" }),
  currentRoute({ id: "error.server", path: "/500", label: "Error", area: "error", role: "public", hiddenInMenu: true, routeKind: "static_content", serviceBoundary: "ConsoleRouter" }),
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

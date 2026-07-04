export type Healthz = { ok: boolean };
export type Readiness = { ready: boolean; checks: Record<string, boolean> };
export type UserView = { id: string; email: string; role: string; status: string };
export type OrganizationView = { id: string; name: string; status: string };
export type TeamView = { id: string; organizationId: string; name: string; status: string };
export type RoleView = { id: string; organizationId?: string; name: string; scope: string; permissions?: unknown };
export type Me = { user: UserView; organization: OrganizationView };
export type Session = { user: UserView; csrfToken: string; expiresAt: string };
export type WorkspacePackage = {
  id: string;
  name: string;
  cpu: number;
  memoryGb: number;
  storageGb: number;
  computeHourlyFen: number;
  storageGbMonthFen: number;
  available: boolean;
};
export type ManagedWorkspace = {
  id: string;
  name: string;
  state: string;
  policy: string;
  url?: string;
  provider?: string;
};
export type ManagedResourceView = {
  id: string;
  organizationId: string;
  resourceType: string;
  resourceId: string;
  displayName: string;
  provider: string;
  status: string;
  policyState: string;
  workspaceId?: string;
  metadata?: unknown;
};
export type CreateWorkspacePayload = {
  workspaceId: string;
  name: string;
  billingAccountId: string;
  packageId: string;
  token: string;
};
export type CreateWorkspaceResult = { workspaceId: string; url: string };
export type WalletView = {
  billingAccountId: string;
  balanceFen: number;
  frozenFen: number;
  availableFen: number;
};
export type BillingLedgerEntry = {
  id: string;
  workspaceId?: string;
  resourceType: string;
  resourceId?: string;
  amountFen: number;
  kind: string;
  description: string;
  createdAt: string;
};
export type SupportTicket = {
  id: string;
  workspaceId?: string;
  subject: string;
  body?: string;
  status: string;
  createdAt?: string;
};
export type CreateSupportTicketPayload = {
  workspaceId?: string;
  subject: string;
  body: string;
};
export type PolicyView = {
  id: string;
  organizationId: string;
  name: string;
  policyType: string;
  status: string;
  rules?: unknown;
};
export type CreatePolicyPayload = {
  organizationId: string;
  name: string;
  policyType: string;
  rules: unknown;
};
export type ApprovalView = {
  id: string;
  organizationId?: string;
  policyId?: string;
  requesterUserId?: string;
  reviewerUserId?: string;
  action?: string;
  objectType?: string;
  objectId?: string;
  status: string;
  reason?: string;
  decisionNote?: string;
};
export type ConsoleState = {
  me: Me;
  packages: WorkspacePackage[];
  workspaces: ManagedWorkspace[];
  wallet: WalletView;
  ledger: BillingLedgerEntry[];
  tickets: SupportTicket[];
};
export type ManagementState = {
  users: UserView[];
  organizations: OrganizationView[];
  teams: TeamView[];
  roles: RoleView[];
  resources: ManagedResourceView[];
  policies: PolicyView[];
  approvals: ApprovalView[];
};

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const csrfToken = typeof window === "undefined" ? "" : window.localStorage.getItem("opl_csrf_token") || "";
  const headers = new Headers(init?.headers);
  if (init?.body) {
    headers.set("content-type", "application/json");
  }
  if (csrfToken && init?.method && init.method !== "GET") {
    headers.set("x-opl-csrf-token", csrfToken);
  }
  const response = await fetch(path, {
    credentials: "include",
    headers,
    ...init
  });
  if (!response.ok) {
    throw new Error(`request_failed:${path}:${response.status}`);
  }
  const payload = await response.json() as T;
  if (path === "/api/auth/login" || path === "/api/auth/operator-login" || path === "/api/auth/me") {
    const maybeSession = payload as Session;
    if (maybeSession.csrfToken && typeof window !== "undefined") {
      window.localStorage.setItem("opl_csrf_token", maybeSession.csrfToken);
    }
  }
  if (path === "/api/auth/logout" && typeof window !== "undefined") {
    window.localStorage.removeItem("opl_csrf_token");
  }
  return payload;
}

export const api = {
  healthz: () => request<Healthz>("/api/healthz"),
  runtimeReadiness: () => request<Readiness>("/api/runtime/readiness"),
  productionReadiness: () => request<Readiness>("/api/production/readiness"),
  state: () => request<ConsoleState>("/api/state"),
  managementState: () => request<ManagementState>("/api/management/state"),
  login: (email: string, password: string) =>
    request<Session>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password })
    }),
  operatorLogin: (operatorToken: string) =>
    request<Session>("/api/auth/operator-login", {
      method: "POST",
      body: JSON.stringify({ operatorToken })
    }),
  session: () => request<Session>("/api/auth/me"),
  logout: () => request<{ ok: boolean }>("/api/auth/logout", { method: "POST" }),
  me: async () => (await api.state()).me,
  packages: async () => (await api.state()).packages,
  workspaces: async () => (await api.state()).workspaces,
  createWorkspace: (payload: CreateWorkspacePayload) =>
    request<CreateWorkspaceResult>("/api/workspaces", {
      method: "POST",
      body: JSON.stringify(payload)
    }),
  resetWorkspaceToken: (id: string, token: string) =>
    request<CreateWorkspaceResult>("/api/workspaces/reset-token", {
      method: "POST",
      body: JSON.stringify({ workspaceId: id, token })
    }),
  deleteWorkspaceToken: (id: string) =>
    request<CreateWorkspaceResult>("/api/workspaces/delete-token", {
      method: "POST",
      body: JSON.stringify({ workspaceId: id })
    }),
  wallet: async () => (await api.state()).wallet,
  billingLedger: async () => (await api.state()).ledger,
  supportTickets: async () => (await request<{ tickets: SupportTicket[] }>("/api/support/tickets")).tickets,
  createSupportTicket: (payload: CreateSupportTicketPayload) =>
    request<SupportTicket>("/api/support/tickets", {
      method: "POST",
      body: JSON.stringify(payload)
    }),
  adminUsers: async () => (await api.managementState()).users,
  adminOrganizations: async () => (await api.managementState()).organizations,
  adminTeams: async () => (await api.managementState()).teams,
  adminRoles: async () => (await api.managementState()).roles,
  adminResources: async () => (await api.managementState()).resources,
  adminPolicies: async () => (await api.managementState()).policies,
  adminApprovals: async () => (await api.managementState()).approvals
};

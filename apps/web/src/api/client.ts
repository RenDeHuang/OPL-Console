export type Healthz = { ok: boolean };
export type Readiness = { ready: boolean; checks: Record<string, boolean> };
export type UserView = { id: string; email: string; role: string; status: string };
export type OrganizationView = { id: string; name: string; status: string };
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

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    credentials: "include",
    headers: init?.body ? { "content-type": "application/json", ...init.headers } : init?.headers,
    ...init
  });
  if (!response.ok) {
    throw new Error(`request_failed:${path}:${response.status}`);
  }
  return response.json() as Promise<T>;
}

export const api = {
  healthz: () => request<Healthz>("/api/healthz"),
  runtimeReadiness: () => request<Readiness>("/api/runtime/readiness"),
  productionReadiness: () => request<Readiness>("/api/production/readiness"),
  login: (email: string, password: string) =>
    request<Session>("/api/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password })
    }),
  session: () => request<Session>("/api/auth/session"),
  logout: () => request<{ ok: boolean }>("/api/auth/logout", { method: "POST" }),
  me: () => request<Me>("/api/me"),
  packages: () => request<WorkspacePackage[]>("/api/packages"),
  workspaces: () => request<ManagedWorkspace[]>("/api/workspaces"),
  createWorkspace: (payload: CreateWorkspacePayload) =>
    request<CreateWorkspaceResult>("/api/workspaces", {
      method: "POST",
      body: JSON.stringify(payload)
    }),
  wallet: () => request<WalletView>("/api/billing/wallet"),
  billingLedger: () => request<BillingLedgerEntry[]>("/api/billing/ledger"),
  supportTickets: () => request<SupportTicket[]>("/api/support/tickets"),
  createSupportTicket: (payload: CreateSupportTicketPayload) =>
    request<SupportTicket>("/api/support/tickets", {
      method: "POST",
      body: JSON.stringify(payload)
    }),
  adminUsers: () => request<UserView[]>("/api/admin/users"),
  adminPolicies: () => request<PolicyView[]>("/api/admin/policies"),
  createPolicy: (payload: CreatePolicyPayload) =>
    request<PolicyView>("/api/admin/policies", {
      method: "POST",
      body: JSON.stringify(payload)
    }),
  adminApprovals: () => request<ApprovalView[]>("/api/admin/approvals"),
  approveApproval: (id: string, decisionNote: string) =>
    request<ApprovalView>(`/api/admin/approvals/${id}/approve`, {
      method: "POST",
      body: JSON.stringify({ decisionNote })
    }),
  rejectApproval: (id: string, decisionNote: string) =>
    request<ApprovalView>(`/api/admin/approvals/${id}/reject`, {
      method: "POST",
      body: JSON.stringify({ decisionNote })
    })
};

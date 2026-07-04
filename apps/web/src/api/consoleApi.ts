export type ConsoleState = {
  account: { id: string; name: string; status: string };
  wallet: { balance: number; frozen: number; totalRecharged: number };
  packages: Record<string, unknown>[];
  computePools: Record<string, unknown>[];
  computeAllocations: Record<string, unknown>[];
  storageVolumes: Record<string, unknown>[];
  storageAttachments: Record<string, unknown>[];
  workspaces: Record<string, unknown>[];
  manualTopups: Record<string, unknown>[];
  walletTransactions: Record<string, unknown>[];
  billingLedger: Record<string, unknown>[];
  resourceUsageLogs: Record<string, unknown>[];
  requestUsageLogs: Record<string, unknown>[];
  notifications: Record<string, unknown>[];
};

export type ManagementState = {
  organizations: Record<string, unknown>[];
  accounts: Record<string, unknown>[];
  users: Record<string, unknown>[];
};

export type SupportTickets = {
  tickets: Record<string, unknown>[];
};

export type AuthSession = {
  user: {
    id: string;
    email: string;
    name: string;
    role: "lab_owner" | "admin";
    accountId: string;
  };
  csrfToken: string;
};

const fallbackState: ConsoleState = {
  account: { id: "acct-demo", name: "演示实验室", status: "active" },
  wallet: { balance: 2500, frozen: 420, totalRecharged: 5000 },
  packages: [
    {
      id: "basic",
      name: "基础工作区",
      server: "标准 CPU",
      cpu: 2,
      memoryGb: 4,
      diskGb: 10,
      available: true,
      price: { computeHourly: 0.39, storageGbMonth: 0.36 },
    },
    {
      id: "pro",
      name: "专业工作区",
      server: "标准 CPU",
      cpu: 8,
      memoryGb: 16,
      diskGb: 100,
      available: true,
      price: { computeHourly: 3.09, storageGbMonth: 0.36 },
    },
  ],
  computePools: [{ id: "pool-standard-cpu", name: "标准 CPU", cpu: 8, memoryGb: 32, status: "available" }],
  computeAllocations: [{ id: "cmp-alpha", name: "Alpha 计算", status: "running", cpu: 4, memoryGb: 16 }],
  storageVolumes: [{ id: "vol-alpha", name: "Alpha 数据卷", sizeGb: 100, status: "available" }],
  storageAttachments: [{ id: "att-alpha", computeId: "cmp-alpha", storageId: "vol-alpha", mountPath: "/data", status: "attached" }],
  workspaces: [{
    id: "ws-alpha",
    name: "Alpha 实验室",
    state: "running",
    status: "running",
    runtimeStatus: "ready",
    url: "https://workspace.example/ws-alpha",
    packageId: "basic",
    computeAllocationId: "cmp-alpha",
    storageId: "vol-alpha",
    attachmentId: "att-alpha",
    access: { tokenStatus: "active" },
  }],
  manualTopups: [{ id: "topup-demo-1", targetAccountId: "acct-demo", amount: 5000, reason: "初始额度" }],
  walletTransactions: [{ id: "txn-demo-1", type: "topup", accountId: "acct-demo", amount: 5000 }],
  billingLedger: [],
  resourceUsageLogs: [],
  requestUsageLogs: [],
  notifications: [{ id: "note-demo-1", type: "runtime", severity: "info", workspaceId: "ws-alpha" }],
};

const fallbackManagement: ManagementState = {
  organizations: [{ id: "org-demo", name: "演示组织", billingAccountId: "acct-demo", status: "active" }],
  accounts: [{ id: "acct-demo", balance: 2500, frozen: 420, totalRecharged: 5000, status: "active" }],
  users: [
    { id: "user-demo-owner", email: "owner@opl.local", name: "OPL 所有者", role: "lab_owner", accountId: "acct-demo", status: "active" },
    { id: "user-demo-admin", email: "admin@opl.local", name: "OPL 管理员", role: "admin", accountId: "acct-operator", status: "active" },
  ],
};

async function getJSON<T>(path: string, fallback: T): Promise<T> {
  try {
    const response = await fetch(path, { headers: { accept: "application/json" } });
    if (!response.ok) return fallback;
    return (await response.json()) as T;
  } catch {
    return fallback;
  }
}

async function postJSON<T>(path: string, body: Record<string, unknown>): Promise<T> {
  const response = await fetch(path, {
    method: "POST",
    headers: {
      accept: "application/json",
      "content-type": "application/json",
    },
    body: JSON.stringify(body),
  });
  const payload = await response.json().catch(() => ({}));
  if (!response.ok) {
    throw new Error(String((payload as { error?: string }).error || "request_failed"));
  }
  return payload as T;
}

export function loginOwner(credentials: { email: string; password: string }) {
  return postJSON<AuthSession>("/api/auth/login", credentials).catch((error) => demoLogin(credentials, error));
}

export function loginOperator(credentials: { email: string; password: string; operatorToken: string }) {
  return postJSON<AuthSession>("/api/auth/operator-login", credentials).catch((error) => demoOperatorLogin(credentials, error));
}

export function logoutSession(csrfToken: string) {
  return postJSON<{ ok: boolean }>("/api/auth/logout", { csrfToken });
}

function demoLogin(credentials: { email: string; password: string }, error: unknown): AuthSession {
  if (!isNetworkError(error)) throw error;
  if (credentials.password !== "password") {
    throw new Error("invalid_credentials");
  }
  if (credentials.email === "admin@opl.local") {
    return {
      user: {
        id: "user-demo-admin",
        email: "admin@opl.local",
        name: "OPL 管理员",
        role: "admin",
        accountId: "acct-operator",
      },
      csrfToken: "csrf-demo-token",
    };
  }
  if (credentials.email !== "owner@opl.local") {
    throw new Error("invalid_credentials");
  }
  return {
    user: {
      id: "user-demo-owner",
      email: "owner@opl.local",
      name: "OPL 所有者",
      role: "lab_owner",
      accountId: "acct-demo",
    },
    csrfToken: "csrf-demo-token",
  };
}

function demoOperatorLogin(credentials: { email: string; password: string; operatorToken: string }, error: unknown): AuthSession {
  if (!isNetworkError(error)) throw error;
  if (credentials.email !== "admin@opl.local" || credentials.password !== "password" || credentials.operatorToken !== "operator-dev-token") {
    throw new Error("invalid_operator_credentials");
  }
  return {
    user: {
      id: "user-demo-admin",
      email: "admin@opl.local",
      name: "OPL 管理员",
      role: "admin",
      accountId: "acct-operator",
    },
    csrfToken: "csrf-demo-token",
  };
}

function isNetworkError(error: unknown) {
  return error instanceof TypeError
    || (error instanceof Error && ["Failed to fetch", "request_failed"].includes(error.message));
}

export function loadConsoleState() {
  return getJSON<ConsoleState>("/api/state", fallbackState);
}

export function loadManagementState() {
  return getJSON<ManagementState>("/api/management/state", fallbackManagement);
}

export function loadOperatorSummary() {
  return getJSON<Record<string, unknown>>("/api/operator/summary", {
    accounts: { total: 1, frozen: 420 },
    workspaces: { total: 1, running: 1 },
    runtimeOperations: { failed: 0 },
    notifications: { total: 1, error: 0, recent: [] },
  });
}

export function loadTickets() {
  return getJSON<SupportTickets>("/api/support/tickets", {
    tickets: [{ id: "ticket-demo-1", title: "工作区开通问题", category: "workspace", status: "open" }],
  });
}

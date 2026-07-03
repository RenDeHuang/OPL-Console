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
  adminUsers: () => request<UserView[]>("/api/admin/users")
};

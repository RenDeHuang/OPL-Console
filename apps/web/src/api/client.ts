export type Healthz = { ok: boolean };
export type Readiness = { ready: boolean; checks: Record<string, boolean> };

async function request<T>(path: string): Promise<T> {
  const response = await fetch(path, { credentials: "include" });
  if (!response.ok) {
    throw new Error(`request_failed:${path}:${response.status}`);
  }
  return response.json() as Promise<T>;
}

export const api = {
  healthz: () => request<Healthz>("/api/healthz"),
  runtimeReadiness: () => request<Readiness>("/api/runtime/readiness")
};

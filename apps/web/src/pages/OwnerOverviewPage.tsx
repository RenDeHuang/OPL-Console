import { useQuery } from "@tanstack/react-query";
import { Box, Database, Server } from "lucide-react";
import { api } from "../api/client";

export function OwnerOverviewPage() {
  const me = useQuery({ queryKey: ["me"], queryFn: api.me, retry: false });
  const packages = useQuery({ queryKey: ["packages"], queryFn: api.packages, retry: false });
  const workspaces = useQuery({ queryKey: ["workspaces"], queryFn: api.workspaces, retry: false });
  const readiness = useQuery({ queryKey: ["runtime-readiness"], queryFn: api.runtimeReadiness });
  let readinessStatus = "Not ready";

  if (readiness.isLoading) {
    readinessStatus = "Checking";
  } else if (readiness.isError) {
    readinessStatus = "Unavailable";
  } else if (readiness.data?.ready) {
    readinessStatus = "Ready";
  }

  return (
    <main className="shell">
      <div className="page-header">
        <div>
          <h1>Managed Workspaces</h1>
          <p>{me.data?.organization.name ?? "Organization governance surface"}</p>
        </div>
        <span className={`status ${readiness.data?.ready ? "ok" : "warn"}`}>{readinessStatus}</span>
      </div>
      <div className="grid three">
        <section className="panel metric">
          <Server size={18} />
          <span>Runtime</span>
          <strong>{readinessStatus}</strong>
        </section>
        <section className="panel metric">
          <Box size={18} />
          <span>Managed Workspaces</span>
          <strong>{workspaces.data?.length ?? 0}</strong>
        </section>
        <section className="panel metric">
          <Database size={18} />
          <span>Packages</span>
          <strong>{packages.data?.length ?? 0}</strong>
        </section>
      </div>
      <section className="panel">
        <h2>Workspace Packages</h2>
        <div className="table">
          {(packages.data ?? []).map((item) => (
            <div className="row" key={item.id}>
              <span>{item.name}</span>
              <span>{item.cpu} CPU / {item.memoryGb} GB</span>
              <span>{item.storageGb} GB storage</span>
              <span>{item.available ? "Available" : "Unavailable"}</span>
            </div>
          ))}
        </div>
      </section>
      <section className="panel">
        <h2>Managed Workspace Views</h2>
        <div className="table">
          {(workspaces.data ?? []).map((workspace) => (
            <div className="row" key={workspace.id}>
              <span>{workspace.name}</span>
              <span>{workspace.state}</span>
              <span>{workspace.policy}</span>
              <span>{workspace.provider || "console-facade"}</span>
            </div>
          ))}
          {workspaces.data?.length === 0 ? <p className="muted">No managed workspaces yet.</p> : null}
        </div>
      </section>
    </main>
  );
}

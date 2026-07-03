import { useQuery } from "@tanstack/react-query";
import { ShieldCheck, Users } from "lucide-react";
import { api } from "../api/client";

export function AdminOverviewPage() {
  const users = useQuery({ queryKey: ["admin-users"], queryFn: api.adminUsers, retry: false });
  const runtime = useQuery({ queryKey: ["runtime-readiness"], queryFn: api.runtimeReadiness });
  const production = useQuery({ queryKey: ["production-readiness"], queryFn: api.productionReadiness });

  return (
    <main className="shell">
      <div className="page-header">
        <div>
          <h1>Admin</h1>
          <p>Governance, readiness, and managed resource oversight</p>
        </div>
        <span className={`status ${production.data?.ready ? "ok" : "warn"}`}>
          {production.data?.ready ? "Production ready" : "Production gated"}
        </span>
      </div>
      <div className="grid three">
        <section className="panel metric">
          <Users size={18} />
          <span>Users</span>
          <strong>{users.data?.length ?? 0}</strong>
        </section>
        <section className="panel metric">
          <ShieldCheck size={18} />
          <span>Runtime</span>
          <strong>{runtime.data?.ready ? "Ready" : "Blocked"}</strong>
        </section>
        <section className="panel metric">
          <ShieldCheck size={18} />
          <span>Production</span>
          <strong>{production.data?.ready ? "Ready" : "Blocked"}</strong>
        </section>
      </div>
      <section className="panel">
        <h2>Users</h2>
        <div className="table">
          {(users.data ?? []).map((user) => (
            <div className="row" key={user.id}>
              <span>{user.email}</span>
              <span>{user.role}</span>
              <span>{user.status}</span>
            </div>
          ))}
        </div>
      </section>
      <section className="panel">
        <h2>Readiness</h2>
        <div className="table">
          <div className="row">
            <span>Runtime</span>
            <span>{runtime.data?.ready ? "Ready" : "Not ready"}</span>
          </div>
          <div className="row">
            <span>Production</span>
            <span>{production.data?.ready ? "Ready" : "Not ready"}</span>
          </div>
        </div>
      </section>
    </main>
  );
}

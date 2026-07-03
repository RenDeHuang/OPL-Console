import { useQuery } from "@tanstack/react-query";
import { ShieldCheck, Users } from "lucide-react";
import { api } from "../api/client";

export function AdminOverviewPage() {
  const users = useQuery({ queryKey: ["admin-users"], queryFn: api.adminUsers, retry: false });
  const organizations = useQuery({ queryKey: ["admin-organizations"], queryFn: api.adminOrganizations, retry: false });
  const teams = useQuery({ queryKey: ["admin-teams"], queryFn: api.adminTeams, retry: false });
  const roles = useQuery({ queryKey: ["admin-roles"], queryFn: api.adminRoles, retry: false });
  const resources = useQuery({ queryKey: ["admin-resources"], queryFn: api.adminResources, retry: false });
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
          <span>Organizations</span>
          <strong>{organizations.data?.length ?? 0}</strong>
        </section>
        <section className="panel metric">
          <ShieldCheck size={18} />
          <span>Resources</span>
          <strong>{resources.data?.length ?? 0}</strong>
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
        <h2>Organizations</h2>
        <div className="table">
          {(organizations.data ?? []).map((organization) => (
            <div className="row" key={organization.id}>
              <span>{organization.name}</span>
              <span>{organization.id}</span>
              <span>{organization.status}</span>
              <span />
            </div>
          ))}
        </div>
      </section>
      <section className="panel">
        <h2>Teams And Roles</h2>
        <div className="table">
          {(teams.data ?? []).map((team) => (
            <div className="row" key={team.id}>
              <span>{team.name}</span>
              <span>{team.organizationId}</span>
              <span>{team.status}</span>
              <span>team</span>
            </div>
          ))}
          {(roles.data ?? []).map((role) => (
            <div className="row" key={role.id}>
              <span>{role.name}</span>
              <span>{role.organizationId || "system"}</span>
              <span>{role.scope}</span>
              <span>role</span>
            </div>
          ))}
        </div>
      </section>
      <section className="panel">
        <h2>Managed Resources</h2>
        <div className="table">
          {(resources.data ?? []).map((resource) => (
            <div className="row" key={resource.id}>
              <span>{resource.displayName}</span>
              <span>{resource.resourceType}</span>
              <span>{resource.status}</span>
              <span>{resource.provider}</span>
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

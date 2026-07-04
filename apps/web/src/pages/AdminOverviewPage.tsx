import { useQuery } from "@tanstack/react-query";
import { ShieldCheck, Users } from "lucide-react";
import { api } from "../api/client";
import { fen, providerText, readinessText, resourceText, scopeText, statusText } from "../format";

export function AdminOverviewPage() {
  const users = useQuery({ queryKey: ["admin-users"], queryFn: api.adminUsers, retry: false });
  const organizations = useQuery({ queryKey: ["admin-organizations"], queryFn: api.adminOrganizations, retry: false });
  const teams = useQuery({ queryKey: ["admin-teams"], queryFn: api.adminTeams, retry: false });
  const roles = useQuery({ queryKey: ["admin-roles"], queryFn: api.adminRoles, retry: false });
  const resources = useQuery({ queryKey: ["admin-resources"], queryFn: api.adminResources, retry: false });
  const ledger = useQuery({ queryKey: ["admin-ledger"], queryFn: api.adminLedger, retry: false });
  const tickets = useQuery({ queryKey: ["admin-support-tickets"], queryFn: api.adminSupportTickets, retry: false });
  const runtime = useQuery({ queryKey: ["runtime-readiness"], queryFn: api.runtimeReadiness });
  const production = useQuery({ queryKey: ["production-readiness"], queryFn: api.productionReadiness });

  return (
    <main className="shell">
      <div className="page-header">
        <div>
          <h1>管理</h1>
          <p>组织、成员、资源和上线状态。</p>
        </div>
        <span className={`status ${production.data?.ready ? "ok" : "warn"}`}>
          {production.data?.ready ? "可上线" : "未达上线条件"}
        </span>
      </div>
      <div className="grid three">
        <section className="panel metric">
          <Users size={18} />
          <span>用户</span>
          <strong>{users.data?.length ?? 0}</strong>
        </section>
        <section className="panel metric">
          <ShieldCheck size={18} />
          <span>组织</span>
          <strong>{organizations.data?.length ?? 0}</strong>
        </section>
        <section className="panel metric">
          <ShieldCheck size={18} />
          <span>资源</span>
          <strong>{resources.data?.length ?? 0}</strong>
        </section>
      </div>
      <section className="panel">
        <h2>用户</h2>
        <div className="table">
          {(users.data ?? []).map((user) => (
            <div className="row" key={user.id}>
              <span>{user.email}</span>
              <span>{user.role === "admin" ? "管理员" : "负责人"}</span>
              <span>{statusText(user.status)}</span>
            </div>
          ))}
        </div>
      </section>
      <section className="panel">
        <h2>组织</h2>
        <div className="table">
          {(organizations.data ?? []).map((organization) => (
            <div className="row" key={organization.id}>
              <span>{organization.name}</span>
              <span>{organization.id}</span>
              <span>{statusText(organization.status)}</span>
              <span />
            </div>
          ))}
        </div>
      </section>
      <section className="panel">
        <h2>团队与角色</h2>
        <div className="table">
          {(teams.data ?? []).map((team) => (
            <div className="row" key={team.id}>
              <span>{team.name}</span>
              <span>{team.organizationId}</span>
              <span>{statusText(team.status)}</span>
              <span>团队</span>
            </div>
          ))}
          {(roles.data ?? []).map((role) => (
            <div className="row" key={role.id}>
              <span>{role.name}</span>
              <span>{role.organizationId || "系统"}</span>
              <span>{scopeText(role.scope)}</span>
              <span>角色</span>
            </div>
          ))}
        </div>
      </section>
      <section className="panel">
        <h2>托管资源</h2>
        <div className="table">
          {(resources.data ?? []).map((resource) => (
            <div className="row" key={resource.id}>
              <span>{resource.displayName}</span>
              <span>{resourceText(resource.resourceType)}</span>
              <span>{statusText(resource.status)}</span>
              <span>{providerText(resource.provider)}</span>
            </div>
          ))}
        </div>
      </section>
      <section className="panel">
        <h2>上线检查</h2>
        <div className="table">
          <div className="row">
            <span>运行环境</span>
            <span>{runtime.data?.ready ? "就绪" : "未就绪"}</span>
          </div>
          <div className="row">
            <span>生产环境</span>
            <span>{production.data?.ready ? "就绪" : "未就绪"}</span>
          </div>
          {Object.entries(production.data?.checks ?? {}).map(([key, value]) => (
            <div className="row" key={key}>
              <span>{readinessText(key)}</span>
              <span>{value ? "通过" : "未通过"}</span>
              <span>{key}</span>
              <span />
            </div>
          ))}
        </div>
      </section>
      <section className="panel">
        <h2>Ledger 事件</h2>
        <div className="table">
          {(ledger.data ?? []).map((entry) => (
            <div className="row" key={entry.id}>
              <span>{entry.kind}</span>
              <span>{fen(entry.amountFen)}</span>
              <span>{entry.workspaceId || entry.resourceId || "账户"}</span>
              <span>{entry.description}</span>
            </div>
          ))}
          {ledger.data?.length === 0 ? <p className="muted">暂无 Ledger 事件。</p> : null}
        </div>
      </section>
      <section className="panel">
        <h2>支持队列</h2>
        <div className="table">
          {(tickets.data ?? []).map((ticket) => (
            <div className="row" key={ticket.id}>
              <span>{ticket.subject}</span>
              <span>{ticket.workspaceId || "通用"}</span>
              <span>{ticket.failedLifecycleStep || statusText(ticket.status)}</span>
              <span>{ticket.fabricErrorCode || ticket.createdAt}</span>
            </div>
          ))}
          {tickets.data?.length === 0 ? <p className="muted">暂无工单。</p> : null}
        </div>
      </section>
    </main>
  );
}

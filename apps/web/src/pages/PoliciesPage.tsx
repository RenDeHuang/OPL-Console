import { useQuery } from "@tanstack/react-query";
import { api } from "../api/client";
import { policyTypeText, statusText } from "../format";

export function PoliciesPage() {
  const policies = useQuery({ queryKey: ["admin-policies"], queryFn: api.adminPolicies, retry: false });

  return (
    <main className="shell">
      <div className="page-header">
        <div>
          <h1>策略</h1>
          <p>控制工作空间开通、配额和审批。</p>
        </div>
        <span className="status ok">{policies.data?.length ?? 0} 条策略</span>
      </div>
      <section className="panel">
        <h2>策略列表</h2>
        <div className="table">
          {(policies.data ?? []).map((policy) => (
            <div className="row" key={policy.id}>
              <span>{policy.name}</span>
              <span>{policyTypeText(policy.policyType)}</span>
              <span>{statusText(policy.status)}</span>
              <span>{policy.organizationId}</span>
            </div>
          ))}
          {policies.data?.length === 0 ? <p className="muted">暂无策略。</p> : null}
        </div>
      </section>
    </main>
  );
}

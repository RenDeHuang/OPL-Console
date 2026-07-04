import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus } from "lucide-react";
import { api } from "../api/client";
import { policyTypeText, statusText } from "../format";

export function PoliciesPage() {
  const queryClient = useQueryClient();
  const me = useQuery({ queryKey: ["me"], queryFn: api.me, retry: false });
  const policies = useQuery({ queryKey: ["admin-policies"], queryFn: api.adminPolicies, retry: false });
  const [organizationId, setOrganizationID] = useState("");
  const [name, setName] = useState("工作空间审批");
  const [policyType, setPolicyType] = useState("workspace_lifecycle");
  const [rules, setRules] = useState('{"requiresApproval":true}');
  const targetOrganizationID = organizationId || me.data?.organization.id || "";
  const createPolicy = useMutation({
    mutationFn: () => api.createPolicy({ organizationId: targetOrganizationID, name, policyType, rules: JSON.parse(rules) as unknown }),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["admin-policies"] });
    }
  });

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
        <h2>创建策略</h2>
        <form
          className="form-grid"
          onSubmit={(event) => {
            event.preventDefault();
            createPolicy.mutate();
          }}
        >
          <label>
            组织 ID
            <input value={targetOrganizationID} onChange={(event) => setOrganizationID(event.target.value)} />
          </label>
          <label>
            名称
            <input value={name} onChange={(event) => setName(event.target.value)} />
          </label>
          <label>
            类型
            <input value={policyType} onChange={(event) => setPolicyType(event.target.value)} />
          </label>
          <label className="wide">
            规则
            <textarea value={rules} onChange={(event) => setRules(event.target.value)} />
          </label>
          <button type="submit" disabled={!targetOrganizationID || !name || !policyType || createPolicy.isPending}>
            <Plus size={16} />
            {createPolicy.isPending ? "创建中" : "创建"}
          </button>
        </form>
        {createPolicy.isError ? <p className="error">创建失败，请检查规则、权限和登录状态。</p> : null}
      </section>
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

import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Plus } from "lucide-react";
import { api } from "../api/client";

export function PoliciesPage() {
  const queryClient = useQueryClient();
  const me = useQuery({ queryKey: ["me"], queryFn: api.me, retry: false });
  const policies = useQuery({ queryKey: ["admin-policies"], queryFn: api.adminPolicies, retry: false });
  const [organizationId, setOrganizationID] = useState("");
  const [name, setName] = useState("Managed Workspace Approval");
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
          <h1>Policies</h1>
          <p>Admin policy definitions for managed resources and lifecycle gates.</p>
        </div>
        <span className="status ok">{policies.data?.length ?? 0} policies</span>
      </div>
      <section className="panel">
        <h2>Create Policy</h2>
        <form
          className="form-grid"
          onSubmit={(event) => {
            event.preventDefault();
            createPolicy.mutate();
          }}
        >
          <label>
            Organization ID
            <input value={targetOrganizationID} onChange={(event) => setOrganizationID(event.target.value)} />
          </label>
          <label>
            Name
            <input value={name} onChange={(event) => setName(event.target.value)} />
          </label>
          <label>
            Type
            <input value={policyType} onChange={(event) => setPolicyType(event.target.value)} />
          </label>
          <label className="wide">
            Rules JSON
            <textarea value={rules} onChange={(event) => setRules(event.target.value)} />
          </label>
          <button type="submit" disabled={!targetOrganizationID || !name || !policyType || createPolicy.isPending}>
            <Plus size={16} />
            {createPolicy.isPending ? "Creating" : "Create"}
          </button>
        </form>
        {createPolicy.isError ? <p className="error">Policy create failed. Check JSON and admin session.</p> : null}
      </section>
      <section className="panel">
        <h2>Policy List</h2>
        <div className="table">
          {(policies.data ?? []).map((policy) => (
            <div className="row" key={policy.id}>
              <span>{policy.name}</span>
              <span>{policy.policyType}</span>
              <span>{policy.status}</span>
              <span>{policy.organizationId}</span>
            </div>
          ))}
          {policies.data?.length === 0 ? <p className="muted">No policies configured.</p> : null}
        </div>
      </section>
    </main>
  );
}

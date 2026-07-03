import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { KeyRound, Pause, Plus, Server, Settings, Trash2 } from "lucide-react";
import { api } from "../api/client";

function randomToken() {
  return `opl-${Math.random().toString(36).slice(2)}${Date.now().toString(36)}`;
}

export function WorkspacesPage() {
  const queryClient = useQueryClient();
  const packages = useQuery({ queryKey: ["packages"], queryFn: api.packages });
  const workspaces = useQuery({ queryKey: ["workspaces"], queryFn: api.workspaces, retry: false });
  const wallet = useQuery({ queryKey: ["wallet"], queryFn: api.wallet, retry: false });
  const [name, setName] = useState("Managed Lab Workspace");
  const [workspaceId, setWorkspaceId] = useState(`ws-${Date.now().toString(36)}`);
  const [packageId, setPackageId] = useState("");
  const selectedPackageId = packageId || packages.data?.[0]?.id || "";
  const canCreate = Boolean(workspaceId && name && selectedPackageId && wallet.data?.billingAccountId);
  const createWorkspace = useMutation({
    mutationFn: () =>
      api.createWorkspace({
        workspaceId,
        name,
        packageId: selectedPackageId,
        billingAccountId: wallet.data?.billingAccountId ?? "",
        token: randomToken()
      }),
    onSuccess: () => {
      setWorkspaceId(`ws-${Date.now().toString(36)}`);
      void queryClient.invalidateQueries({ queryKey: ["workspaces"] });
      void queryClient.invalidateQueries({ queryKey: ["billing-ledger"] });
    }
  });
  const lifecycle = useMutation({
    mutationFn: ({ id, action }: { id: string; action: "configure" | "suspend" | "delete" | "reset-token" | "delete-token" }) => {
      if (action === "configure") return api.configureWorkspace(id);
      if (action === "suspend") return api.suspendWorkspace(id);
      if (action === "delete") return api.deleteWorkspace(id);
      if (action === "reset-token") return api.resetWorkspaceToken(id, randomToken());
      return api.deleteWorkspaceToken(id);
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["workspaces"] });
      void queryClient.invalidateQueries({ queryKey: ["billing-ledger"] });
    }
  });
  const workspaceCount = workspaces.data?.length ?? 0;
  const readyPackages = useMemo(() => (packages.data ?? []).filter((item) => item.available), [packages.data]);

  return (
    <main className="shell">
      <div className="page-header">
        <div>
          <h1>Managed Workspaces</h1>
          <p>Create, view, and govern Console-managed Workspace lifecycle.</p>
        </div>
        <span className="status ok">{workspaceCount} managed</span>
      </div>
      <section className="panel">
        <h2>Create Workspace</h2>
        <form
          className="form-grid"
          onSubmit={(event) => {
            event.preventDefault();
            createWorkspace.mutate();
          }}
        >
          <label>
            Workspace ID
            <input value={workspaceId} onChange={(event) => setWorkspaceId(event.target.value)} />
          </label>
          <label>
            Name
            <input value={name} onChange={(event) => setName(event.target.value)} />
          </label>
          <label>
            Package
            <select value={selectedPackageId} onChange={(event) => setPackageId(event.target.value)}>
              {readyPackages.map((item) => (
                <option key={item.id} value={item.id}>
                  {item.name} - {item.cpu} CPU / {item.memoryGb} GB / {item.storageGb} GB
                </option>
              ))}
            </select>
          </label>
          <label>
            Billing account
            <input value={wallet.data?.billingAccountId ?? "loading"} disabled />
          </label>
          <button type="submit" disabled={!canCreate || createWorkspace.isPending}>
            <Plus size={16} />
            {createWorkspace.isPending ? "Creating" : "Create"}
          </button>
        </form>
        {createWorkspace.isError ? <p className="error">Workspace create failed</p> : null}
        {createWorkspace.data ? <p className="muted">Created route: {createWorkspace.data.url}</p> : null}
      </section>
      <section className="panel">
        <h2>Managed Workspace Views</h2>
        <div className="table">
          {(workspaces.data ?? []).map((workspace) => (
            <div className="workspace-row" key={workspace.id}>
              <span>{workspace.name}</span>
              <span>{workspace.state}</span>
              <span>{workspace.policy}</span>
              <span>{workspace.url ? <a href={workspace.url}>{workspace.provider || "open"}</a> : workspace.provider || "console-facade"}</span>
              <div className="button-row">
                <button type="button" title="Configure" disabled={lifecycle.isPending} onClick={() => lifecycle.mutate({ id: workspace.id, action: "configure" })}>
                  <Settings size={16} />
                </button>
                <button type="button" title="Suspend" disabled={lifecycle.isPending} onClick={() => lifecycle.mutate({ id: workspace.id, action: "suspend" })}>
                  <Pause size={16} />
                </button>
                <button type="button" title="Reset token" disabled={lifecycle.isPending} onClick={() => lifecycle.mutate({ id: workspace.id, action: "reset-token" })}>
                  <KeyRound size={16} />
                </button>
                <button type="button" title="Delete token" disabled={lifecycle.isPending} onClick={() => lifecycle.mutate({ id: workspace.id, action: "delete-token" })}>
                  <KeyRound size={16} />
                </button>
                <button className="danger" type="button" title="Delete" disabled={lifecycle.isPending} onClick={() => lifecycle.mutate({ id: workspace.id, action: "delete" })}>
                  <Trash2 size={16} />
                </button>
              </div>
            </div>
          ))}
          {workspaces.data?.length === 0 ? <p className="muted">No managed workspaces yet.</p> : null}
        </div>
      </section>
      <section className="panel">
        <h2>Packages</h2>
        <div className="table">
          {(packages.data ?? []).map((item) => (
            <div className="row" key={item.id}>
              <span><Server size={14} /> {item.name}</span>
              <span>{item.cpu} CPU / {item.memoryGb} GB</span>
              <span>{item.storageGb} GB storage</span>
              <span>{item.available ? "Available" : "Unavailable"}</span>
            </div>
          ))}
        </div>
      </section>
    </main>
  );
}

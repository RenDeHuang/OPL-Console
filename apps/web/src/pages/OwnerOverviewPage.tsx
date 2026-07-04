import { useQuery } from "@tanstack/react-query";
import { Box, Database, Server } from "lucide-react";
import { api } from "../api/client";
import { packageText, statusText } from "../format";

export function OwnerOverviewPage() {
  const me = useQuery({ queryKey: ["me"], queryFn: api.me, retry: false });
  const packages = useQuery({ queryKey: ["packages"], queryFn: api.packages, retry: false });
  const workspaces = useQuery({ queryKey: ["workspaces"], queryFn: api.workspaces, retry: false });
  const readiness = useQuery({ queryKey: ["runtime-readiness"], queryFn: api.runtimeReadiness });
  let readinessStatus = "未就绪";

  if (readiness.isLoading) {
    readinessStatus = "检查中";
  } else if (readiness.isError) {
    readinessStatus = "不可用";
  } else if (readiness.data?.ready) {
    readinessStatus = "就绪";
  }

  return (
    <main className="shell">
      <div className="page-header">
        <div>
          <h1>控制台概览</h1>
          <p>{me.data?.organization.name ?? "组织治理"}</p>
        </div>
        <span className={`status ${readiness.data?.ready ? "ok" : "warn"}`}>{readinessStatus}</span>
      </div>
      <div className="grid three">
        <section className="panel metric">
          <Server size={18} />
          <span>运行环境</span>
          <strong>{readinessStatus}</strong>
        </section>
        <section className="panel metric">
          <Box size={18} />
          <span>工作空间</span>
          <strong>{workspaces.data?.length ?? 0}</strong>
        </section>
        <section className="panel metric">
          <Database size={18} />
          <span>套餐</span>
          <strong>{packages.data?.length ?? 0}</strong>
        </section>
      </div>
      <section className="panel">
        <h2>可用套餐</h2>
        <div className="table">
          {(packages.data ?? []).map((item) => (
            <div className="row" key={item.id}>
              <span>{packageText(item.name)}</span>
              <span>{item.cpu} 核 / {item.memoryGb}GB</span>
              <span>{item.storageGb}GB 存储</span>
              <span>{item.available ? "可开通" : "不可用"}</span>
            </div>
          ))}
        </div>
      </section>
      <section className="panel">
        <h2>托管工作空间</h2>
        <div className="table">
          {(workspaces.data ?? []).map((workspace) => (
            <div className="row" key={workspace.id}>
              <span>{workspace.name}</span>
              <span>{statusText(workspace.state)}</span>
              <span>{statusText(workspace.policy)}</span>
              <span>{workspace.provider || "控制台托管"}</span>
            </div>
          ))}
          {workspaces.data?.length === 0 ? <p className="muted">暂无托管工作空间。</p> : null}
        </div>
      </section>
    </main>
  );
}

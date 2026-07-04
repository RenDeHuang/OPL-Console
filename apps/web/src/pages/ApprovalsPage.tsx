import { useQuery } from "@tanstack/react-query";
import { api } from "../api/client";
import { actionText, resourceText, statusText } from "../format";

export function ApprovalsPage() {
  const approvals = useQuery({ queryKey: ["admin-approvals"], queryFn: api.adminApprovals, retry: false });

  return (
    <main className="shell">
      <div className="page-header">
        <div>
          <h1>审批</h1>
          <p>处理工作空间开通和策略拦截。</p>
        </div>
        <span className="status ok">{approvals.data?.filter((item) => item.status === "pending").length ?? 0} 待处理</span>
      </div>
      <section className="panel">
        <h2>审批队列</h2>
        <div className="table">
          {(approvals.data ?? []).map((approval) => (
            <div className="approval-row" key={approval.id}>
              <div>
                <strong>{actionText(approval.action)}</strong>
                <p className="muted">{resourceText(approval.objectType) || "对象"} {approval.objectId || approval.id}</p>
              </div>
              <span>{statusText(approval.status)}</span>
              <span>{approval.reason || "无备注"}</span>
              <span>{approval.decisionNote || "未处理"}</span>
            </div>
          ))}
          {approvals.data?.length === 0 ? <p className="muted">暂无审批。</p> : null}
        </div>
      </section>
    </main>
  );
}

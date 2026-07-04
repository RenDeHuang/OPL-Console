import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Check, X } from "lucide-react";
import { api } from "../api/client";
import { actionText, fen, resourceText, statusText } from "../format";

function approvalContext(value: unknown) {
  if (!value || typeof value !== "object") return {};
  return value as Record<string, unknown>;
}

export function ApprovalsPage() {
  const queryClient = useQueryClient();
  const approvals = useQuery({ queryKey: ["admin-approvals"], queryFn: api.adminApprovals, retry: false });
  const [noteByID, setNoteByID] = useState<Record<string, string>>({});
  const approve = useMutation({
    mutationFn: (id: string) => api.approveApproval(id, noteByID[id] ?? ""),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["admin-approvals"] })
  });
  const reject = useMutation({
    mutationFn: (id: string) => api.rejectApproval(id, noteByID[id] ?? ""),
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: ["admin-approvals"] })
  });

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
                <p className="muted">
                  {String(approvalContext(approval.context).policyRuleTriggered ?? "策略拦截")} ·
                  冻结 {fen(Number(approvalContext(approval.context).estimatedHoldFen ?? 0))}
                </p>
                <p className="muted">
                  后续动作：{Array.isArray(approvalContext(approval.context).postApprovalActions)
                    ? (approvalContext(approval.context).postApprovalActions as unknown[]).join(" / ")
                    : "冻结余额 / 创建资源 / 生成访问地址"}
                </p>
              </div>
              <span>{statusText(approval.status)}</span>
              <input
                value={noteByID[approval.id] ?? ""}
                onChange={(event) => setNoteByID((current) => ({ ...current, [approval.id]: event.target.value }))}
                placeholder="审批备注"
              />
              <div className="button-row">
                <button type="button" disabled={approval.status !== "pending" || approve.isPending} onClick={() => approve.mutate(approval.id)}>
                  <Check size={16} />
                  通过
                </button>
                <button className="danger" type="button" disabled={approval.status !== "pending" || reject.isPending} onClick={() => reject.mutate(approval.id)}>
                  <X size={16} />
                  拒绝
                </button>
              </div>
            </div>
          ))}
          {approvals.data?.length === 0 ? <p className="muted">暂无审批。</p> : null}
        </div>
      </section>
    </main>
  );
}

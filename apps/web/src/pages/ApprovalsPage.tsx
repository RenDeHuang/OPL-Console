import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Check, X } from "lucide-react";
import { api } from "../api/client";

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
          <h1>Approvals</h1>
          <p>Review pending policy gates for managed Workspace operations.</p>
        </div>
        <span className="status ok">{approvals.data?.filter((item) => item.status === "pending").length ?? 0} pending</span>
      </div>
      <section className="panel">
        <h2>Approval Queue</h2>
        <div className="table">
          {(approvals.data ?? []).map((approval) => (
            <div className="approval-row" key={approval.id}>
              <div>
                <strong>{approval.action || "request"}</strong>
                <p className="muted">{approval.objectType || "object"} {approval.objectId || approval.id}</p>
              </div>
              <span>{approval.status}</span>
              <input
                value={noteByID[approval.id] ?? ""}
                onChange={(event) => setNoteByID((current) => ({ ...current, [approval.id]: event.target.value }))}
                placeholder="decision note"
              />
              <div className="button-row">
                <button type="button" disabled={approval.status !== "pending" || approve.isPending} onClick={() => approve.mutate(approval.id)}>
                  <Check size={16} />
                  Approve
                </button>
                <button className="danger" type="button" disabled={approval.status !== "pending" || reject.isPending} onClick={() => reject.mutate(approval.id)}>
                  <X size={16} />
                  Reject
                </button>
              </div>
            </div>
          ))}
          {approvals.data?.length === 0 ? <p className="muted">No approvals in queue.</p> : null}
        </div>
      </section>
    </main>
  );
}

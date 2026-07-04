import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Eye, Send } from "lucide-react";
import { api } from "../api/client";
import { statusText } from "../format";

export function SupportPage() {
  const queryClient = useQueryClient();
  const tickets = useQuery({ queryKey: ["support-tickets"], queryFn: api.supportTickets, retry: false });
  const [workspaceId, setWorkspaceId] = useState("");
  const [subject, setSubject] = useState("");
  const [body, setBody] = useState("");
  const [openTicketId, setOpenTicketId] = useState("");
  const createTicket = useMutation({
    mutationFn: () => api.createSupportTicket({ workspaceId, subject, body }),
    onSuccess: () => {
      setSubject("");
      setBody("");
      void queryClient.invalidateQueries({ queryKey: ["support-tickets"] });
    }
  });

  return (
    <main className="shell">
      <div className="page-header">
        <div>
          <h1>工单</h1>
          <p>查看和提交工作空间运维工单。</p>
        </div>
        <span className="status ok">{tickets.data?.length ?? 0} 个工单</span>
      </div>
      <section className="panel">
        <h2>提交工单</h2>
        <form
          className="form-grid"
          onSubmit={(event) => {
            event.preventDefault();
            createTicket.mutate();
          }}
        >
          <label>
            工作空间 ID
            <input value={workspaceId} onChange={(event) => setWorkspaceId(event.target.value)} placeholder="可选" />
          </label>
          <label>
            标题
            <input value={subject} onChange={(event) => setSubject(event.target.value)} />
          </label>
          <label className="wide">
            内容
            <textarea value={body} onChange={(event) => setBody(event.target.value)} />
          </label>
          <button type="submit" disabled={!subject || !body || createTicket.isPending}>
            <Send size={16} />
            {createTicket.isPending ? "提交中" : "提交"}
          </button>
        </form>
      </section>
      <section className="panel">
        <h2>工单列表</h2>
        <div className="table">
          {(tickets.data ?? []).map((ticket) => (
            <div className="ticket-row" key={ticket.id}>
              <span>{ticket.subject}</span>
              <span>{ticket.workspaceId || "通用"}</span>
              <span>{statusText(ticket.status)}</span>
              <button type="button" onClick={() => setOpenTicketId(openTicketId === ticket.id ? "" : ticket.id)}>
                <Eye size={16} />
                查看
              </button>
              {openTicketId === ticket.id ? (
                <div className="ticket-detail">
                  <strong>{ticket.createdAt || "未记录时间"}</strong>
                  <p>{ticket.body || "无详细内容"}</p>
                </div>
              ) : null}
            </div>
          ))}
          {tickets.data?.length === 0 ? <p className="muted">暂无工单。</p> : null}
        </div>
      </section>
    </main>
  );
}

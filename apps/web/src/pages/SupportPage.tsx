import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Send } from "lucide-react";
import { api } from "../api/client";

export function SupportPage() {
  const queryClient = useQueryClient();
  const tickets = useQuery({ queryKey: ["support-tickets"], queryFn: api.supportTickets, retry: false });
  const [workspaceId, setWorkspaceId] = useState("");
  const [subject, setSubject] = useState("");
  const [body, setBody] = useState("");
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
          <h1>Support</h1>
          <p>Owner support tickets for managed Workspace operations.</p>
        </div>
        <span className="status ok">{tickets.data?.length ?? 0} tickets</span>
      </div>
      <section className="panel">
        <h2>Create Ticket</h2>
        <form
          className="form-grid"
          onSubmit={(event) => {
            event.preventDefault();
            createTicket.mutate();
          }}
        >
          <label>
            Workspace ID
            <input value={workspaceId} onChange={(event) => setWorkspaceId(event.target.value)} placeholder="optional" />
          </label>
          <label>
            Subject
            <input value={subject} onChange={(event) => setSubject(event.target.value)} />
          </label>
          <label className="wide">
            Body
            <textarea value={body} onChange={(event) => setBody(event.target.value)} />
          </label>
          <button type="submit" disabled={!subject || !body || createTicket.isPending}>
            <Send size={16} />
            {createTicket.isPending ? "Sending" : "Send"}
          </button>
        </form>
      </section>
      <section className="panel">
        <h2>Tickets</h2>
        <div className="table">
          {(tickets.data ?? []).map((ticket) => (
            <div className="row" key={ticket.id}>
              <span>{ticket.subject}</span>
              <span>{ticket.workspaceId || "general"}</span>
              <span>{ticket.status}</span>
              <span>{ticket.createdAt || ""}</span>
            </div>
          ))}
          {tickets.data?.length === 0 ? <p className="muted">No support tickets yet.</p> : null}
        </div>
      </section>
    </main>
  );
}

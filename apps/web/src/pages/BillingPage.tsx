import { useQuery } from "@tanstack/react-query";
import { CreditCard } from "lucide-react";
import { api } from "../api/client";

function fen(value?: number) {
  return `¥${((value ?? 0) / 100).toFixed(2)}`;
}

export function BillingPage() {
  const wallet = useQuery({ queryKey: ["wallet"], queryFn: api.wallet, retry: false });
  const ledger = useQuery({ queryKey: ["billing-ledger"], queryFn: api.billingLedger, retry: false });

  return (
    <main className="shell">
      <div className="page-header">
        <div>
          <h1>Managed Billing</h1>
          <p>Console shows billing evidence without taking over Ledger platform ownership.</p>
        </div>
        <span className="status ok">{wallet.data?.billingAccountId ?? "account"}</span>
      </div>
      <div className="grid three">
        <section className="panel metric">
          <CreditCard size={18} />
          <span>Balance</span>
          <strong>{fen(wallet.data?.balanceFen)}</strong>
        </section>
        <section className="panel metric">
          <CreditCard size={18} />
          <span>Frozen</span>
          <strong>{fen(wallet.data?.frozenFen)}</strong>
        </section>
        <section className="panel metric">
          <CreditCard size={18} />
          <span>Available</span>
          <strong>{fen(wallet.data?.availableFen)}</strong>
        </section>
      </div>
      <section className="panel">
        <h2>Ledger Evidence</h2>
        <div className="table">
          {(ledger.data ?? []).map((entry) => (
            <div className="row" key={entry.id}>
              <span>{entry.kind}</span>
              <span>{fen(entry.amountFen)}</span>
              <span>{entry.resourceType} {entry.workspaceId || entry.resourceId}</span>
              <span>{entry.description}</span>
            </div>
          ))}
          {ledger.data?.length === 0 ? <p className="muted">No billing evidence yet.</p> : null}
        </div>
      </section>
    </main>
  );
}

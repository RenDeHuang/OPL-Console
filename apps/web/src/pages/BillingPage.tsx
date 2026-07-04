import { useQuery } from "@tanstack/react-query";
import { CreditCard } from "lucide-react";
import { api } from "../api/client";
import { fen, ledgerKindText, resourceText } from "../format";

export function BillingPage() {
  const wallet = useQuery({ queryKey: ["wallet"], queryFn: api.wallet, retry: false });
  const ledger = useQuery({ queryKey: ["billing-ledger"], queryFn: api.billingLedger, retry: false });

  return (
    <main className="shell">
      <div className="page-header">
        <div>
          <h1>账单</h1>
          <p>查看余额、冻结金额和托管资源费用证据。</p>
        </div>
        <span className="status ok">{wallet.data?.billingAccountId ?? "账户"}</span>
      </div>
      <div className="grid three">
        <section className="panel metric">
          <CreditCard size={18} />
          <span>余额</span>
          <strong>{fen(wallet.data?.balanceFen)}</strong>
        </section>
        <section className="panel metric">
          <CreditCard size={18} />
          <span>冻结</span>
          <strong>{fen(wallet.data?.frozenFen)}</strong>
        </section>
        <section className="panel metric">
          <CreditCard size={18} />
          <span>可用</span>
          <strong>{fen(wallet.data?.availableFen)}</strong>
        </section>
      </div>
      <section className="panel">
        <h2>费用明细</h2>
        <div className="table">
          {(ledger.data ?? []).map((entry) => (
            <div className="row" key={entry.id}>
              <span>{ledgerKindText(entry.kind)}</span>
              <span>{fen(entry.amountFen)}</span>
              <span>{resourceText(entry.resourceType)} {entry.workspaceId || entry.resourceId}</span>
              <span>{entry.description}</span>
            </div>
          ))}
          {ledger.data?.length === 0 ? <p className="muted">暂无费用记录。</p> : null}
        </div>
      </section>
    </main>
  );
}

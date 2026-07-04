import { useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Archive, HardDrive, KeyRound, Pause, Play, RotateCcw, Server, Trash2 } from "lucide-react";
import { Link, useParams } from "react-router";
import { api } from "../api/client";
import { actionText, fen, lifecycleStepText, packageText, statusText } from "../format";

function randomToken() {
  return `opl-${Math.random().toString(36).slice(2)}${Date.now().toString(36)}`;
}

export function WorkspaceDetailPage() {
  const { id = "" } = useParams();
  const queryClient = useQueryClient();
  const workspace = useQuery({ queryKey: ["workspace", id], queryFn: () => api.workspace(id), enabled: Boolean(id), retry: false });
  const [storageConfirm, setStorageConfirm] = useState("");
  const lifecycle = useMutation({
    mutationFn: (action: "stop" | "restart" | "destroy-compute" | "destroy-storage" | "backup" | "restore" | "reset-token" | "delete-token") => {
      if (action === "stop") return api.stopCompute(id);
      if (action === "restart") return api.restartCompute(id);
      if (action === "destroy-compute") return api.destroyCompute(id);
      if (action === "destroy-storage") return api.destroyStorage(id, storageConfirm === workspace.data?.name);
      if (action === "backup") return api.createBackup(id);
      if (action === "restore") return api.restoreBackup(id);
      if (action === "reset-token") return api.resetWorkspaceToken(id, randomToken());
      return api.deleteWorkspaceToken(id);
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["workspace", id] });
      void queryClient.invalidateQueries({ queryKey: ["workspaces"] });
      void queryClient.invalidateQueries({ queryKey: ["billing-ledger"] });
    }
  });

  const detail = workspace.data;

  return (
    <main className="shell">
      <div className="page-header">
        <div>
          <Link to="/workspaces" className="muted">返回工作空间</Link>
          <h1>{detail?.name ?? "工作空间详情"}</h1>
          <p>{detail?.id ?? id}</p>
        </div>
        <span className={`status ${detail?.state === "running" ? "ok" : "warn"}`}>{statusText(detail?.state)}</span>
      </div>

      <div className="grid three">
        <section className="panel metric">
          <Server size={18} />
          <span>计算</span>
          <strong>{statusText(detail?.computeStatus)}</strong>
        </section>
        <section className="panel metric">
          <HardDrive size={18} />
          <span>存储</span>
          <strong>{statusText(detail?.storageStatus)}</strong>
        </section>
        <section className="panel metric">
          <KeyRound size={18} />
          <span>访问地址</span>
          <strong>{detail?.url ? "已生成" : statusText(detail?.tokenStatus) || "待生成"}</strong>
        </section>
      </div>

      <section className="panel">
        <h2>访问与生命周期</h2>
        <div className="button-row">
          {detail?.url ? <a className="button-link" href={detail.url}>打开</a> : null}
          <button type="button" disabled={lifecycle.isPending} onClick={() => lifecycle.mutate("stop")}><Pause size={16} />停止计算</button>
          <button type="button" disabled={lifecycle.isPending} onClick={() => lifecycle.mutate("restart")}><Play size={16} />重启计算</button>
          <button type="button" disabled={lifecycle.isPending} onClick={() => lifecycle.mutate("destroy-compute")}><Server size={16} />销毁计算</button>
          <button type="button" disabled={lifecycle.isPending} onClick={() => lifecycle.mutate("backup")}><Archive size={16} />创建备份</button>
          <button type="button" disabled={lifecycle.isPending} onClick={() => lifecycle.mutate("restore")}><RotateCcw size={16} />恢复备份</button>
          <button type="button" disabled={lifecycle.isPending} onClick={() => lifecycle.mutate("reset-token")}><KeyRound size={16} />重置地址</button>
          <button type="button" disabled={lifecycle.isPending} onClick={() => lifecycle.mutate("delete-token")}><KeyRound size={16} />停用地址</button>
        </div>
        <div className="danger-zone">
          <label>
            销毁存储确认
            <input value={storageConfirm} onChange={(event) => setStorageConfirm(event.target.value)} placeholder={`输入 ${detail?.name ?? "工作空间名称"}`} />
          </label>
          <button className="danger" type="button" disabled={lifecycle.isPending || storageConfirm !== detail?.name} onClick={() => lifecycle.mutate("destroy-storage")}>
            <Trash2 size={16} />
            销毁存储
          </button>
          <p className="danger-note">销毁存储是唯一会停止存储计费的动作，必须输入工作空间名称确认。</p>
        </div>
        {lifecycle.isError ? <p className="error">操作失败，请检查资源状态、余额和权限。</p> : null}
      </section>

      <section className="panel">
        <h2>资源与费用</h2>
        <div className="row">
          <span>{packageText(detail?.package?.name)}</span>
          <span>{detail?.package?.cpu ?? 0} 核 / {detail?.package?.memoryGb ?? 0}GB</span>
          <span>{detail?.package?.storageGb ?? 0}GB 存储</span>
          <span>7 天冻结 {fen(detail?.estimatedHoldFen)}</span>
        </div>
        <div className="row">
          <span>计算 {detail?.computeId || "未登记"}</span>
          <span>存储 {detail?.storageId || "未登记"}</span>
          <span>挂载 {statusText(detail?.attachmentStatus)}</span>
          <span>账单账户 {detail?.billingAccountId}</span>
        </div>
      </section>

      <section className="panel">
        <h2>生命周期</h2>
        <div className="table">
          {(detail?.lifecycleSteps ?? []).map((step) => (
            <div className="row" key={step.stepName}>
              <span>{lifecycleStepText(step.stepName)}</span>
              <span>{statusText(step.actualState)}</span>
              <span>{step.providerResourceId || "未登记"}</span>
              <span>{step.errorCode || step.lastCheckedAt}</span>
            </div>
          ))}
          {detail?.lifecycleSteps?.length === 0 ? <p className="muted">暂无生命周期步骤。</p> : null}
        </div>
      </section>

      <section className="panel">
        <h2>费用与凭证</h2>
        <div className="table">
          {(detail?.ledgerEntries ?? []).map((entry) => (
            <div className="row" key={entry.id}>
              <span>{entry.kind}</span>
              <span>{fen(entry.amountFen)}</span>
              <span>{entry.resourceType}</span>
              <span>{entry.description}</span>
            </div>
          ))}
          {(detail?.receipts ?? []).map((receipt) => (
            <div className="row" key={receipt.id}>
              <span>{receipt.receiptType}</span>
              <span>{receipt.subjectType}</span>
              <span>{receipt.subjectId}</span>
              <span>{receipt.operationId || "凭证"}</span>
            </div>
          ))}
        </div>
      </section>

      <section className="panel">
        <h2>工单与审计</h2>
        <div className="table">
          {(detail?.supportTickets ?? []).map((ticket) => (
            <div className="row" key={ticket.id}>
              <span>{ticket.subject}</span>
              <span>{statusText(ticket.status)}</span>
              <span>{ticket.failedLifecycleStep || "未关联故障"}</span>
              <span>{ticket.fabricErrorCode || ticket.createdAt}</span>
            </div>
          ))}
          {(detail?.auditEvents ?? []).map((event) => (
            <div className="row" key={event.id}>
              <span>{actionText(event.action)}</span>
              <span>{statusText(event.result)}</span>
              <span>{event.actorUserId}</span>
              <span>{event.createdAt}</span>
            </div>
          ))}
        </div>
      </section>
    </main>
  );
}

import { useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { HardDrive, Plus, Route, Server, Settings } from "lucide-react";
import { Link } from "react-router";
import { api } from "../api/client";
import { fen, packageText, providerText, statusText } from "../format";

function randomToken() {
  return `opl-${Math.random().toString(36).slice(2)}${Date.now().toString(36)}`;
}

export function WorkspacesPage() {
  const queryClient = useQueryClient();
  const packages = useQuery({ queryKey: ["packages"], queryFn: api.packages });
  const workspaces = useQuery({ queryKey: ["workspaces"], queryFn: api.workspaces, retry: false });
  const wallet = useQuery({ queryKey: ["wallet"], queryFn: api.wallet, retry: false });
  const [name, setName] = useState("实验工作空间");
  const [workspaceId, setWorkspaceId] = useState(`ws-${Date.now().toString(36)}`);
  const [packageId, setPackageId] = useState("");
  const selectedPackageId = packageId || packages.data?.[0]?.id || "";
  const quote = useQuery({
    queryKey: ["workspace-quote", wallet.data?.billingAccountId, selectedPackageId],
    queryFn: () => api.workspaceQuote(wallet.data?.billingAccountId ?? "", selectedPackageId),
    enabled: Boolean(wallet.data?.billingAccountId && selectedPackageId),
    retry: false
  });
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
  const workspaceCount = workspaces.data?.length ?? 0;
  const readyPackages = useMemo(() => (packages.data ?? []).filter((item) => item.available), [packages.data]);

  return (
    <main className="shell">
      <div className="page-header">
        <div>
          <h1>托管工作空间</h1>
          <p>开通、配置、暂停和删除组织托管的工作空间。</p>
        </div>
        <span className="status ok">{workspaceCount} 个</span>
      </div>
      <section className="panel">
        <h2>开通工作空间</h2>
        <form
          className="form-grid"
          onSubmit={(event) => {
            event.preventDefault();
            createWorkspace.mutate();
          }}
        >
          <label>
            工作空间编号
            <input value={workspaceId} onChange={(event) => setWorkspaceId(event.target.value)} />
          </label>
          <label>
            名称
            <input value={name} onChange={(event) => setName(event.target.value)} />
          </label>
          <label>
            套餐
            <select value={selectedPackageId} onChange={(event) => setPackageId(event.target.value)}>
              {readyPackages.map((item) => (
                <option key={item.id} value={item.id}>
                  {packageText(item.name)} - {item.cpu} 核 / {item.memoryGb}GB 内存 / {item.storageGb}GB 存储
                </option>
              ))}
            </select>
          </label>
          <label>
            账单账户
            <input value={wallet.data?.billingAccountId ?? "加载中"} disabled />
          </label>
          <button type="submit" disabled={!canCreate || createWorkspace.isPending}>
            <Plus size={16} />
            {createWorkspace.isPending ? "开通中" : "开通"}
          </button>
        </form>
        <div className="process-grid">
          <div><HardDrive size={16} /><span>1. 开通存储</span></div>
          <div><Server size={16} /><span>2. 开通计算</span></div>
          <div><Settings size={16} /><span>3. 挂载存储</span></div>
          <div><Route size={16} /><span>4. 生成访问地址</span></div>
        </div>
        <div className="cost-strip">
          <span>计算冻结：{fen(quote.data?.computeHoldFen)}</span>
          <span>存储冻结：{fen(quote.data?.storageHoldFen)}</span>
          <strong>{quote.data?.sufficientBalance === false ? "余额不足：" : "预计冻结："}{fen(quote.data?.totalHoldFen)}</strong>
        </div>
        {createWorkspace.isError ? <p className="error">开通失败，请检查余额、策略审批和运行环境。</p> : null}
        {createWorkspace.data ? <p className="muted">访问地址：{createWorkspace.data.url || "等待审批或运行时生成"}</p> : null}
      </section>
      <section className="panel">
        <h2>工作空间列表</h2>
        <div className="table">
          {(workspaces.data ?? []).map((workspace) => (
            <div className="workspace-row" key={workspace.id}>
              <span>{workspace.name}</span>
              <span>{statusText(workspace.state)}</span>
              <span>{statusText(workspace.computeStatus)} / {statusText(workspace.storageStatus)}</span>
              <span>{workspace.url ? <a href={workspace.url}>打开</a> : workspace.provider ? providerText(workspace.provider) : "待生成"}</span>
              <div className="button-row">
                <Link className="button-link" to={`/workspaces/${workspace.id}`}>
                  <Settings size={16} />
                  详情
                </Link>
              </div>
            </div>
          ))}
          {workspaces.data?.length === 0 ? <p className="muted">还没有托管工作空间。</p> : null}
        </div>
      </section>
      <section className="panel">
        <h2>套餐价格</h2>
        <div className="table">
          {(packages.data ?? []).map((item) => (
            <div className="row" key={item.id}>
              <span><Server size={14} /> {packageText(item.name)}</span>
              <span>{item.cpu} 核 / {item.memoryGb}GB</span>
              <span>{item.storageGb}GB 存储</span>
              <span>{fen(item.computeHourlyFen)} / 小时，{fen(item.storageGbMonthFen)} / GB·月</span>
            </div>
          ))}
        </div>
      </section>
    </main>
  );
}

import { Button, Empty, Form, Input, InputNumber, Select, Typography } from "antd";
import { AlertTriangle, Database, HardDrive, Headphones, Link as LinkIcon, Plus, Server, Settings2, WalletCards } from "lucide-react";
import type { ConsoleState, ManagementState, SupportTickets } from "../api/consoleApi";
import type { OplRoute } from "../routes/oplRoutes";
import { routeTo } from "../routes/oplRoutes";
import {
  ActionGroup,
  ConsoleSurface,
  InsightPanel,
  MetricStrip,
  ObjectTable,
  ResourceSplit,
  StatusPill,
  TimelineList,
} from "./shared/commercial-console";
import { available, fieldLabel, money, packageText, statusColor, statusLabel, valueLabel } from "./shared/formatters";

type ConsoleProps = {
  route: OplRoute;
  path: string;
  state: ConsoleState;
  wallet?: ConsoleState["wallet"];
  tickets?: SupportTickets;
};

type AdminProps = {
  route: OplRoute;
  state: ConsoleState;
  management?: ManagementState;
  operator?: Record<string, unknown>;
  tickets?: SupportTickets;
};

export function ConsoleRoutePage({ route, path, state, wallet = state.wallet, tickets = { tickets: [] } }: ConsoleProps) {
  if (route.id === "console.overview") return <OverviewPage route={route} state={state} wallet={wallet} tickets={tickets} />;
  if (path.startsWith("/console/compute/new")) return <CreateResourcePage route={route} title="开通计算" action="开通计算" fields={["name", "packageId"]} />;
  if (path.startsWith("/console/compute/")) return <DetailPage route={route} data={state.computeAllocations} id={path.split("/").at(-1) || ""} title="计算" />;
  if (path.startsWith("/console/compute")) return <ComputePage route={route} state={state} />;
  if (path.startsWith("/console/storage/new")) return <CreateResourcePage route={route} title="开通存储" action="开通存储" fields={["name", "packageId", "sizeGb"]} />;
  if (path.startsWith("/console/storage/")) return <DetailPage route={route} data={state.storageVolumes} id={path.split("/").at(-1) || ""} title="存储" />;
  if (path.startsWith("/console/storage")) return <ResourceListPage route={route} title="存储" eyebrow="TKE 资源" data={state.storageVolumes} createPath="/console/storage/new" createLabel="开通存储" />;
  if (path.startsWith("/console/attachments/new")) return <CreateResourcePage route={route} title="挂载存储" action="挂载存储" fields={["computeId", "storageId", "mountPath"]} />;
  if (path.startsWith("/console/attachments")) return <ResourceListPage route={route} title="挂载" eyebrow="存储挂载" data={state.storageAttachments} createPath="/console/attachments/new" createLabel="挂载存储" />;
  if (path.startsWith("/console/resources/relationships")) return <ResourceRelationshipPage route={route} state={state} />;
  if (path.startsWith("/console/workspaces/new")) return <CreateResourcePage route={route} title="创建工作区" action="创建工作区" fields={["name", "packageId", "token"]} />;
  if (path.startsWith("/console/workspaces/")) return <DetailPage route={route} data={state.workspaces} id={path.split("/").at(-1) || ""} title="工作区" />;
  if (path.startsWith("/console/workspaces")) return <WorkspacesPage route={route} state={state} wallet={wallet} />;
  if (path.startsWith("/console/billing")) return <BillingPage route={route} state={state} wallet={wallet} />;
  if (path.startsWith("/console/account")) return <AccountPage route={route} state={state} />;
  if (path.startsWith("/console/support/new")) return <CreateResourcePage route={route} title="新建工单" action="创建工单" fields={["title", "category", "body"]} />;
  if (path.startsWith("/console/support")) return <ResourceListPage route={route} title="支持" eyebrow="工单" data={tickets.tickets} createPath="/console/support/new" createLabel="新建工单" />;
  if (path.startsWith("/console/alerts")) return <AlertsPage route={route} state={state} />;
  if (path.startsWith("/console/gateway")) return <GatewayPage route={route} />;
  return <OverviewPage route={route} state={state} wallet={wallet} tickets={tickets} />;
}

export function AdminRoutePage({ route, state, management, operator, tickets = { tickets: [] } }: AdminProps) {
  if (route.id === "admin.users") return <AdminUsersPage route={route} management={management} />;
  if (route.id === "admin.billing") return <ResourceListPage route={route} title="计费运营" eyebrow="管理" data={state.manualTopups} />;
  if (route.id === "admin.ledger") return <ResourceListPage route={route} title="账本" eyebrow="管理" data={state.billingLedger} />;
  if (route.id === "admin.runtime") return <AdminRuntimePage route={route} operator={operator} />;
  if (route.id === "admin.diagnostics") return <AdminDiagnosticsPage route={route} operator={operator} />;
  if (route.id === "admin.e2e") return <AdminE2EPage route={route} operator={operator} />;
  if (route.id === "admin.cleanup") return <AdminCleanupPage route={route} management={management} operator={operator} />;
  if (route.id === "admin.support") return <ResourceListPage route={route} title="支持运营" eyebrow="管理" data={tickets.tickets} />;
  return <AdminOverviewPage route={route} state={state} operator={operator} />;
}

function OverviewPage({ route, state, wallet, tickets }: { route: OplRoute; state: ConsoleState; wallet: ConsoleState["wallet"]; tickets: SupportTickets }) {
  const computeRunning = state.computeAllocations.filter((compute) => compute.status === "running").length;
  const storageAvailable = state.storageVolumes.filter((storage) => storage.status !== "destroyed").length;
  const activeTickets = tickets?.tickets.filter((ticket) => ticket.status !== "closed").length || 0;
  const usable = available(wallet);
  const freezeRatio = Number(wallet?.balance || 0) > 0
    ? Math.min(100, Math.max(0, (Number(wallet?.frozen || 0) / Number(wallet?.balance || 1)) * 100))
    : 0;
  const recentSignals = [
    ...state.notifications.slice(-5).reverse().map((event) => ({
      title: String(event.type || "alert"),
      description: String(event.workspaceId || event.accountId || ""),
      meta: String(event.severity || "signal"),
      tone: event.severity === "error" ? "danger" as const : "warn" as const,
    })),
    ...(tickets?.tickets || []).slice(-3).reverse().map((ticket) => ({
      title: String(ticket.title || ticket.id),
      description: String(ticket.category || ""),
      meta: String(ticket.status || "open"),
      tone: "info" as const,
    })),
  ].slice(0, 6);

  return (
    <ConsoleSurface
      title="总览"
      eyebrow="OPL 控制台"
      subtitle="钱包、工作区交付、网关用量与支持工单"
      extra={<Button type="primary" icon={<Plus size={15} />} href={routeTo("workspace.create")}>创建工作区</Button>}
    >
      <MetricStrip
        items={[
          { label: "可用余额", value: money(usable), caption: `${money(wallet?.frozen)} 已冻结`, icon: <WalletCards size={16} />, tone: usable > 0 ? "good" : "warn" },
          { label: "工作区", value: state.workspaces.length, caption: `${computeRunning} 个计算运行中`, icon: <Server size={16} />, tone: computeRunning ? "good" : "neutral" },
          { label: "存储资源", value: storageAvailable, caption: "存储卷", icon: <HardDrive size={16} />, tone: storageAvailable ? "info" : "neutral" },
          { label: "网关请求", value: state.requestUsageLogs.length, caption: "gflabtoken.cn", icon: <LinkIcon size={16} />, tone: "info" },
          { label: "工单", value: activeTickets, caption: `共 ${tickets?.tickets.length || 0} 个`, icon: <Headphones size={16} />, tone: activeTickets ? "warn" : "neutral" },
          { label: "告警", value: state.notifications.length, caption: "所有者可见", icon: <AlertTriangle size={16} />, tone: state.notifications.length ? "danger" : "good" },
        ]}
      />

      <div className="consoleGrid">
        <InsightPanel
          title="业务链"
          eyebrow="开通链路"
          actions={<ActionGroup actions={[
            { label: "工作区", type: "primary", icon: <Plus size={15} />, onClick: () => { window.location.href = routeTo("workspace.create"); } },
            { label: "钱包", icon: <WalletCards size={15} />, onClick: () => { window.location.href = routeTo("billing.wallet"); } },
            { label: "工单", icon: <Headphones size={15} />, onClick: () => { window.location.href = routeTo("support.create"); } },
          ]} />}
        >
          <ResourceSplit
            items={[
              { label: "充值与冻结", value: `${money(wallet?.balance)} / ${money(wallet?.frozen)}`, meta: "余额 / 冻结金额", status: `${Math.round(freezeRatio)}% 已冻结`, tone: freezeRatio > 70 ? "warn" : "info" },
              { label: "计算交付", value: `${computeRunning}/${state.computeAllocations.length}`, meta: "运行中的计算分配", status: computeRunning ? "有效" : "空闲", tone: computeRunning ? "good" : "neutral" },
              { label: "存储资源", value: storageAvailable, meta: "持久化存储卷", status: storageAvailable ? "可用" : "空", tone: storageAvailable ? "good" : "neutral" },
              { label: "访问链接分发", value: state.workspaces.filter((workspace) => (workspace.access as { tokenStatus?: string } | undefined)?.tokenStatus === "active").length, meta: "有效的工作区访问链接", status: "已限定", tone: "info" },
              { label: "支持闭环", value: activeTickets, meta: "待处理工单", status: activeTickets ? "待处理" : "已清理", tone: activeTickets ? "warn" : "good" },
            ]}
          />
        </InsightPanel>

        <InsightPanel title="最近信号" eyebrow="关注">
          <TimelineList items={recentSignals} emptyText="当前没有告警或待处理工单" />
        </InsightPanel>
      </div>

      <InsightPanel title="最近工作区" eyebrow="交付">
        <WorkspaceTable workspaces={state.workspaces.slice(-5).reverse()} packages={state.packages} />
      </InsightPanel>
    </ConsoleSurface>
  );
}

function WorkspacesPage({ route, state, wallet }: { route: OplRoute; state: ConsoleState; wallet: ConsoleState["wallet"] }) {
  const running = state.workspaces.filter((workspace) => workspace.state === "running").length;
  const activeUrls = state.workspaces.filter((workspace) => (workspace.access as { tokenStatus?: string } | undefined)?.tokenStatus === "active").length;
  const attachedEntries = state.workspaces.filter((workspace) => workspace.attachmentId).length;
  return (
    <ConsoleSurface
      title="工作区"
      eyebrow="交付"
      subtitle="计算分配、存储卷、访问令牌与计费冻结"
      extra={<Button type="primary" icon={<Plus size={15} />} href={routeTo("workspace.create")}>创建工作区</Button>}
    >
      <MetricStrip
        items={[
          { label: "总数", value: state.workspaces.length, caption: "当前实验室名下", tone: state.workspaces.length ? "info" : "neutral" },
          { label: "运行中", value: running, caption: "工作区访问入口", tone: running ? "good" : "neutral" },
          { label: "链接可用", value: activeUrls, caption: "可分享访问链接", tone: activeUrls ? "good" : "warn" },
          { label: "已挂载", value: attachedEntries, caption: "已绑定存储挂载", tone: attachedEntries ? "info" : "neutral" },
          { label: "可用余额", value: money(available(wallet)), caption: "扣除冻结后", tone: available(wallet) > 0 ? "good" : "warn" },
        ]}
      />
      <InsightPanel title="工作区对象" eyebrow="当前">
        <WorkspaceTable workspaces={state.workspaces} packages={state.packages} />
      </InsightPanel>
      <ContractPanel route={route} />
    </ConsoleSurface>
  );
}

function WorkspaceTable({ workspaces, packages }: { workspaces: Record<string, unknown>[]; packages: Record<string, unknown>[] }) {
  const planById = Object.fromEntries(packages.map((plan) => [String(plan.id), plan]));
  return (
    <ObjectTable
      rowKey="id"
      data={workspaces}
      emptyText="暂无工作区"
      columns={[
        { title: "名称", dataIndex: "name", render: (_: unknown, row: Record<string, unknown>) => <Button type="link" href={routeTo("workspace.detail", { id: String(row.id) })}>{String(row.name || row.id)}</Button> },
        { title: "状态", dataIndex: "state", render: (_: unknown, row: Record<string, unknown>) => <StatusPill label={statusLabel(row)} tone={toneForStatus(row.state)} /> },
        { title: "套餐", dataIndex: "packageId", render: (value: unknown) => packageText(planById[String(value)]) },
        { title: "计算分配", render: (_: unknown, row: Record<string, unknown>) => <Typography.Text ellipsis>{String(row.computeAllocationId || row.computeId || "-")}</Typography.Text> },
        { title: "存储卷", render: (_: unknown, row: Record<string, unknown>) => <Typography.Text ellipsis>{String(row.storageId || "-")}</Typography.Text> },
        { title: "访问链接", dataIndex: "url", ellipsis: true, render: (value: unknown) => <Typography.Text className="inlineCode">{String(value || "-")}</Typography.Text> },
      ]}
    />
  );
}

function ComputePage({ route, state }: { route: OplRoute; state: ConsoleState }) {
  const running = state.computeAllocations.filter((item) => item.status === "running").length;
  return (
    <ConsoleSurface
      title="计算"
      eyebrow="TKE 分配"
      subtitle="账号名下专属 CVM 计算分配"
      extra={<Button type="primary" icon={<Plus size={15} />} href="/console/compute/new">开通计算</Button>}
    >
      <MetricStrip
        items={[
          { label: "计算分配", value: state.computeAllocations.length, caption: "当前账号名下", tone: state.computeAllocations.length ? "info" : "neutral" },
          { label: "运行中", value: running, caption: "计费中的 CVM 运行态", tone: "good" },
        ]}
      />
      <ResourceListPage route={route} title="计算分配" eyebrow="计算分配" data={state.computeAllocations} embedded />
    </ConsoleSurface>
  );
}

function ResourceRelationshipPage({ route, state }: { route: OplRoute; state: ConsoleState }) {
  const rows = state.workspaces.map((workspace) => ({
    id: workspace.id,
    workspace: workspace.name || workspace.id,
    computeAllocationId: workspace.computeAllocationId || workspace.computeId || "-",
    storageId: workspace.storageId || "-",
    attachmentId: workspace.attachmentId || "-",
    tokenStatus: (workspace.access as { tokenStatus?: string } | undefined)?.tokenStatus || "-",
  }));

  return (
    <ConsoleSurface title="资源关系" eyebrow="读模型" subtitle="工作区入口、计算资源、存储资源与挂载关系">
      <MetricStrip
        items={[
          { label: "工作区入口", value: state.workspaces.length, caption: "当前账号名下", tone: state.workspaces.length ? "info" : "neutral" },
          { label: "计算资源", value: state.computeAllocations.length, caption: "计算分配", tone: state.computeAllocations.length ? "good" : "neutral" },
          { label: "存储资源", value: state.storageVolumes.length, caption: "存储卷", tone: state.storageVolumes.length ? "info" : "neutral" },
          { label: "挂载关系", value: state.storageAttachments.length, caption: "存储挂载", tone: state.storageAttachments.length ? "good" : "neutral" },
        ]}
      />
      <InsightPanel title="关系表" eyebrow="资源">
        <ObjectTable data={rows} emptyText="暂无资源关系" columns={columnsFor(rows)} />
      </InsightPanel>
      <ContractPanel route={route} />
    </ConsoleSurface>
  );
}

function BillingPage({ route, state, wallet }: { route: OplRoute; state: ConsoleState; wallet: ConsoleState["wallet"] }) {
  const usable = available(wallet);
  const frozenPercent = wallet.balance > 0 ? Math.min(100, Math.round((wallet.frozen / wallet.balance) * 100)) : 0;
  return (
    <ConsoleSurface title="计费" eyebrow="钱包" subtitle="预付余额、冻结金额与资源用量">
      <MetricStrip
        items={[
          { label: "可用", value: money(usable), caption: "可开通计算或存储", tone: usable > 0 ? "good" : "warn" },
          { label: "冻结", value: money(wallet.frozen), caption: `占余额 ${frozenPercent}%`, tone: frozenPercent > 70 ? "warn" : "info" },
          { label: "余额", value: money(wallet.balance), caption: "可用加冻结", tone: "neutral" },
          { label: "累计充值", value: money(wallet.totalRecharged), caption: "人工充值账本", tone: "good" },
          { label: "扣费记录", value: state.walletTransactions.length, caption: "最近资源事件", tone: state.walletTransactions.length ? "info" : "neutral" },
        ]}
      />
      <div className="consoleGrid">
        <InsightPanel title="钱包拆分" eyebrow="余额">
          <div className="stackList">
            <div className="walletBar"><span style={{ width: `${frozenPercent}%` }} /></div>
            <div className="stackRow"><span>可用余额</span><strong>{money(usable)}</strong></div>
            <div className="stackRow"><span>冻结金额</span><strong>{money(wallet.frozen)}</strong></div>
            <div className="stackRow"><span>总余额</span><strong>{money(wallet.balance)}</strong></div>
          </div>
        </InsightPanel>
        <InsightPanel title="钱包流水" eyebrow="交易">
          <TimelineList items={state.walletTransactions.slice(-6).reverse().map((event) => ({
            title: String(event.type || event.id),
            description: String(event.workspaceId || event.accountId || ""),
            meta: money(event.amount),
            tone: Number(event.amount || 0) < 0 ? "warn" : "good",
          }))} emptyText="暂无钱包流水" />
        </InsightPanel>
      </div>
      <ContractPanel route={route} />
    </ConsoleSurface>
  );
}

function AccountPage({ route, state }: { route: OplRoute; state: ConsoleState }) {
  return (
    <ConsoleSurface title="账号与实验室" eyebrow="管理" subtitle="账号、组织与计费范围">
      <ResourceSplit items={[
        { label: "账号", value: state.account.name, meta: state.account.id, status: valueLabel(state.account.status), tone: "good" },
        { label: "钱包", value: money(state.wallet.balance), meta: `${money(state.wallet.frozen)} 已冻结`, status: "有效", tone: "info" },
      ]} />
      <ContractPanel route={route} />
    </ConsoleSurface>
  );
}

function AlertsPage({ route, state }: { route: OplRoute; state: ConsoleState }) {
  return (
    <ConsoleSurface title="告警" eyebrow="信号" subtitle="所有者可见的运行态与支持信号">
      <InsightPanel title="告警" eyebrow="运行态">
        <TimelineList items={state.notifications.map((event) => ({
          title: String(event.type || event.id),
          description: String(event.workspaceId || event.accountId || ""),
          meta: String(event.severity || "signal"),
          tone: event.severity === "error" ? "danger" : "warn",
        }))} emptyText="暂无告警" />
      </InsightPanel>
      <ContractPanel route={route} />
    </ConsoleSurface>
  );
}

function GatewayPage({ route }: { route: OplRoute }) {
  return (
    <ConsoleSurface title="网关" eyebrow="外部集成" subtitle="OPL 网关外部集成">
      <ResourceSplit items={[
        { label: "网关", value: "外部链接", meta: "控制台只负责用量与策略可视化", status: "外部", tone: "info" },
        { label: "边界", value: "此处不实现", meta: "提供方路由归属网关服务", status: "已限定", tone: "good" },
      ]} />
      <ContractPanel route={route} />
    </ConsoleSurface>
  );
}

function ResourceListPage({ route, title, eyebrow, data, createPath, createLabel, embedded = false }: {
  route: OplRoute;
  title: string;
  eyebrow: string;
  data: Record<string, unknown>[];
  createPath?: string;
  createLabel?: string;
  embedded?: boolean;
}) {
  const panel = (
    <InsightPanel title={title} eyebrow={eyebrow}>
      <ObjectTable data={data} emptyText="暂无数据" columns={columnsFor(data)} />
    </InsightPanel>
  );
  if (embedded) return panel;
  return (
    <ConsoleSurface
      title={title}
      eyebrow={eyebrow}
      extra={createPath && <Button type="primary" icon={<Plus size={15} />} href={createPath}>{createLabel}</Button>}
    >
      {panel}
      <ContractPanel route={route} />
    </ConsoleSurface>
  );
}

function DetailPage({ route, data, id, title }: { route: OplRoute; data: Record<string, unknown>[]; id: string; title: string }) {
  const resource = data.find((item) => item.id === id);
  if (!resource) return <ConsoleSurface title={title} eyebrow="详情"><Empty description="未找到资源" /></ConsoleSurface>;
  return (
    <ConsoleSurface title={String(resource.name || resource.id)} eyebrow={`${title}详情`}>
      <ResourceSplit items={Object.entries(resource).slice(0, 6).map(([key, value]) => ({
        label: fieldLabel(key),
        value: renderCell(value),
        status: key.toLowerCase().includes("status") || key === "state" ? valueLabel(value) : undefined,
        tone: key.toLowerCase().includes("status") || key === "state" ? toneForStatus(value) : "neutral",
      }))} />
      <ContractPanel route={route} />
    </ConsoleSurface>
  );
}

function CreateResourcePage({ route, title, action, fields }: { route: OplRoute; title: string; action: string; fields: string[] }) {
  return (
    <ConsoleSurface title={title} eyebrow="开通" subtitle={route.serviceBoundary} compact>
      <InsightPanel title={action} eyebrow={objectLabel(route.objectKind || route.routeKind)}>
        <Form layout="vertical" initialValues={{ packageId: "basic", sizeGb: 10, mountPath: "/data" }}>
          {fields.map((field) => (
            <Form.Item key={field} name={field} label={labelize(field)} rules={[{ required: field !== "body" }]}>
              {field === "packageId" ? (
                <Select options={[{ label: "基础工作区", value: "basic" }, { label: "专业工作区", value: "pro" }]} />
              ) : field.toLowerCase().includes("gb") ? (
                <InputNumber min={1} className="fullWidth" />
              ) : (
                <Input.TextArea rows={field === "body" ? 4 : 1} />
              )}
            </Form.Item>
          ))}
          <Button className="formSubmit" type="primary" htmlType="submit">{action}</Button>
        </Form>
      </InsightPanel>
      <ContractPanel route={route} />
    </ConsoleSurface>
  );
}

function AdminOverviewPage({ route, state, operator }: { route: OplRoute; state: ConsoleState; operator?: Record<string, unknown> }) {
  return (
    <ConsoleSurface title="管理总览" eyebrow="运营" subtitle="账号、工作区与运行证据">
      <MetricStrip
        items={[
          { label: "账号", value: valueFrom(operator, "accounts.total", 1), caption: "已管理计费账号", tone: "info" },
          { label: "工作区", value: valueFrom(operator, "workspaces.total", state.workspaces.length), caption: "运行中的工作区入口", tone: "good" },
          { label: "失败操作", value: valueFrom(operator, "runtimeOperations.failed", 0), caption: "运行操作失败数", tone: "good" },
          { label: "冻结总额", value: money(valueFrom(operator, "accounts.frozen", state.wallet.frozen)), caption: "全部账号", tone: "warn" },
          { label: "告警", value: valueFrom(operator, "notifications.total", state.notifications.length), caption: "运营可见", tone: "neutral" },
        ]}
      />
      <div className="consoleGrid equal">
        <InsightPanel title="运行态" eyebrow="运行">
          <ResourceSplit items={[
            { label: "运行态", value: "就绪", meta: "运行就绪检查", status: "通过", tone: "good" },
            { label: "上线", value: "阻塞", meta: "生产上线门禁", status: "待检查", tone: "warn" },
            { label: "计算分配", value: state.computeAllocations.length, meta: "CVM 分配证据", status: "已跟踪", tone: "info" },
          ]} />
        </InsightPanel>
        <InsightPanel title="最近告警" eyebrow="信号">
          <TimelineList items={[]} emptyText="暂无运营告警" />
        </InsightPanel>
      </div>
      <ContractPanel route={route} />
    </ConsoleSurface>
  );
}

function AdminUsersPage({ route, management }: { route: OplRoute; management?: ManagementState }) {
  return (
    <ConsoleSurface title="用户" eyebrow="管理" subtitle="登录用户、计费账号与钱包操作" extra={<Button type="primary" icon={<Plus size={15} />}>新建用户</Button>}>
      <InsightPanel title="用户钱包" eyebrow="管理">
        <ObjectTable data={management?.users || []} emptyText="暂无用户" columns={columnsFor(management?.users || [])} />
      </InsightPanel>
      <ContractPanel route={route} />
    </ConsoleSurface>
  );
}

function AdminRuntimePage({ route }: { route: OplRoute; operator?: Record<string, unknown> }) {
  return (
    <ConsoleSurface title="运行态" eyebrow="管理" subtitle="就绪门禁与上线阻塞项">
      <div className="consoleGrid equal">
        <InsightPanel title="就绪检查" eyebrow="运行态">
          <ResourceSplit items={[
            { label: "Fabric", value: "就绪", meta: "运行提供方", status: "通过", tone: "good" },
            { label: "上线", value: "阻塞", meta: "生产门禁", status: "待检查", tone: "warn" },
            { label: "环境", value: 0, meta: "缺失环境输入", status: "环境", tone: "good" },
            { label: "工具", value: 0, meta: "宿主工具检查", status: "工具", tone: "good" },
          ]} />
        </InsightPanel>
        <InsightPanel title="阻塞项" eyebrow="检查">
          <TimelineList items={[]} emptyText="暂无阻塞项" />
        </InsightPanel>
      </div>
      <ContractPanel route={route} />
    </ConsoleSurface>
  );
}

function AdminDiagnosticsPage({ route, operator }: { route: OplRoute; operator?: Record<string, unknown> }) {
  return (
    <ConsoleSurface title="线上诊断" eyebrow="运行状态" subtitle="运行就绪、生产就绪与运营摘要">
      <div className="consoleGrid equal">
        <InsightPanel title="诊断摘要" eyebrow="运行">
          <ResourceSplit items={[
            { label: "工作区总数", value: valueFrom(operator, "workspaces.total", 0), meta: "运营摘要", status: "已读取", tone: "info" },
            { label: "运行中", value: valueFrom(operator, "workspaces.running", 0), meta: "工作区入口", status: "运行", tone: "good" },
            { label: "失败操作", value: valueFrom(operator, "runtimeOperations.failed", 0), meta: "运行操作", status: "跟踪", tone: Number(valueFrom(operator, "runtimeOperations.failed", 0)) > 0 ? "warn" : "good" },
            { label: "告警", value: valueFrom(operator, "notifications.total", 0), meta: "运营可见", status: "信号", tone: "neutral" },
          ]} />
        </InsightPanel>
        <InsightPanel title="最近信号" eyebrow="诊断">
          <TimelineList items={(valueFrom(operator, "notifications.recent", []) as unknown[]).map((entry) => ({
            title: renderCell(entry),
            description: "运营信号",
            meta: "最近",
            tone: "info" as const,
          }))} emptyText="暂无诊断信号" />
        </InsightPanel>
      </div>
      <ContractPanel route={route} />
    </ConsoleSurface>
  );
}

function AdminE2EPage({ route, operator }: { route: OplRoute; operator?: Record<string, unknown> }) {
  return (
    <ConsoleSurface title="E2E记录" eyebrow="生产验证" subtitle="端到端验证与上线证据">
      <MetricStrip
        items={[
          { label: "工作区总数", value: valueFrom(operator, "workspaces.total", 0), caption: "验证对象", tone: "info" },
          { label: "运行中", value: valueFrom(operator, "workspaces.running", 0), caption: "可访问入口", tone: "good" },
          { label: "失败操作", value: valueFrom(operator, "runtimeOperations.failed", 0), caption: "运行操作失败", tone: Number(valueFrom(operator, "runtimeOperations.failed", 0)) > 0 ? "warn" : "good" },
        ]}
      />
      <InsightPanel title="验证记录" eyebrow="E2E">
        <TimelineList items={[]} emptyText="暂无端到端验证记录" />
      </InsightPanel>
      <ContractPanel route={route} />
    </ConsoleSurface>
  );
}

function AdminCleanupPage({ route, management, operator }: { route: OplRoute; management?: ManagementState; operator?: Record<string, unknown> }) {
  return (
    <ConsoleSurface title="入口清理" eyebrow="工作区入口" subtitle="清理失效访问入口与账号关联">
      <MetricStrip
        items={[
          { label: "用户", value: management?.users.length || 0, caption: "管理模型", tone: "info" },
          { label: "账号", value: management?.accounts.length || 0, caption: "计费账号", tone: "neutral" },
          { label: "工作区", value: valueFrom(operator, "workspaces.total", 0), caption: "运营摘要", tone: "good" },
        ]}
      />
      <InsightPanel title="清理候选" eyebrow="入口">
        <ObjectTable data={management?.users || []} emptyText="暂无清理候选" columns={columnsFor(management?.users || [])} />
      </InsightPanel>
      <ContractPanel route={route} />
    </ConsoleSurface>
  );
}

function ContractPanel({ route }: { route: OplRoute }) {
  return (
    <InsightPanel title="API 契约" eyebrow="边界">
      <ResourceSplit items={(route.apiRoutes || ["无直接 API 路由"]).map((apiRoute) => ({
        label: String(apiRoute).split(" ")[0] || "API",
        value: apiRoute,
        meta: route.serviceBoundary,
        status: "已固定",
        tone: "info",
      }))} />
    </InsightPanel>
  );
}

function columnsFor(data: Record<string, unknown>[]) {
  const keys = Array.from(new Set(data.flatMap((row) => Object.keys(row)))).slice(0, 5);
  return keys.map((key) => ({
    title: labelize(key),
    dataIndex: key,
    ellipsis: true,
    render: (value: unknown) => renderCell(value),
  }));
}

function renderCell(value: unknown) {
  if (typeof value === "boolean") return value ? "是" : "否";
  if (value === null || value === undefined || value === "") return "-";
  if (typeof value === "object") return renderObject(value as Record<string, unknown>);
  return valueLabel(value);
}

function renderObject(value: Record<string, unknown>) {
  return Object.entries(value)
    .map(([key, entry]) => `${fieldLabel(key)}: ${typeof entry === "object" ? JSON.stringify(entry) : valueLabel(entry)}`)
    .join("，");
}

function toneForStatus(value: unknown) {
  const color = statusColor(value);
  if (color === "green") return "good" as const;
  if (color === "red") return "danger" as const;
  if (color === "orange") return "warn" as const;
  return "info" as const;
}

function valueFrom<T>(source: Record<string, unknown> | undefined, path: string, fallback: T): T {
  if (!source) return fallback;
  return path.split(".").reduce<unknown>((current, key) => {
    if (current && typeof current === "object" && key in current) return (current as Record<string, unknown>)[key];
    return fallback;
  }, source) as T;
}

function labelize(value: string) {
  return fieldLabel(value);
}

function objectLabel(value: string | undefined) {
  const labels: Record<string, string> = {
    business_object: "业务对象",
    ComputeAllocation: "计算分配",
    ComputePool: "计算资源池",
    external_integration: "外部集成",
    GatewayIntegration: "网关集成",
    read_model: "读取模型",
    RuntimeReadiness: "运行就绪",
    static_content: "静态内容",
    StorageAttachment: "存储挂载",
    StorageVolume: "存储卷",
    SupportTicket: "支持工单",
    User: "用户",
    Wallet: "钱包",
    Workspace: "工作区",
  };
  return labels[value || ""] || value || "-";
}

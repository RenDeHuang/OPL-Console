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
import { available, money, packageText, statusColor, statusLabel } from "./shared/formatters";

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
  if (path.startsWith("/console/compute/new")) return <CreateResourcePage route={route} title="Create Compute" action="开通计算" fields={["name", "packageId"]} />;
  if (path.startsWith("/console/compute/")) return <DetailPage route={route} data={state.computeAllocations} id={path.split("/").at(-1) || ""} title="Compute" />;
  if (path.startsWith("/console/compute")) return <ComputePage route={route} state={state} />;
  if (path.startsWith("/console/storage/new")) return <CreateResourcePage route={route} title="Create Storage" action="开通存储" fields={["name", "packageId", "sizeGb"]} />;
  if (path.startsWith("/console/storage/")) return <DetailPage route={route} data={state.storageVolumes} id={path.split("/").at(-1) || ""} title="Storage" />;
  if (path.startsWith("/console/storage")) return <ResourceListPage route={route} title="Storage" eyebrow="TKE resources" data={state.storageVolumes} createPath="/console/storage/new" createLabel="开通存储" />;
  if (path.startsWith("/console/attachments/new")) return <CreateResourcePage route={route} title="Attach Storage" action="挂载存储" fields={["computeId", "storageId", "mountPath"]} />;
  if (path.startsWith("/console/attachments")) return <ResourceListPage route={route} title="Attachments" eyebrow="StorageAttachment" data={state.storageAttachments} createPath="/console/attachments/new" createLabel="挂载存储" />;
  if (path.startsWith("/console/workspaces/new")) return <CreateResourcePage route={route} title="Create Workspace" action="创建 Workspace" fields={["name", "packageId", "token"]} />;
  if (path.startsWith("/console/workspaces/")) return <DetailPage route={route} data={state.workspaces} id={path.split("/").at(-1) || ""} title="Workspace" />;
  if (path.startsWith("/console/workspaces")) return <WorkspacesPage route={route} state={state} wallet={wallet} />;
  if (path.startsWith("/console/billing")) return <BillingPage route={route} state={state} wallet={wallet} />;
  if (path.startsWith("/console/account")) return <AccountPage route={route} state={state} />;
  if (path.startsWith("/console/support/new")) return <CreateResourcePage route={route} title="New Ticket" action="创建工单" fields={["title", "category", "body"]} />;
  if (path.startsWith("/console/support")) return <ResourceListPage route={route} title="Support" eyebrow="SupportTicket" data={tickets.tickets} createPath="/console/support/new" createLabel="新建工单" />;
  if (path.startsWith("/console/alerts")) return <AlertsPage route={route} state={state} />;
  if (path.startsWith("/console/gateway")) return <GatewayPage route={route} />;
  return <OverviewPage route={route} state={state} wallet={wallet} tickets={tickets} />;
}

export function AdminRoutePage({ route, state, management, operator, tickets = { tickets: [] } }: AdminProps) {
  if (route.id === "admin.users") return <AdminUsersPage route={route} management={management} />;
  if (route.id === "admin.billing") return <ResourceListPage route={route} title="Billing Ops" eyebrow="Admin" data={state.manualTopups} />;
  if (route.id === "admin.ledger") return <ResourceListPage route={route} title="Ledger" eyebrow="Admin" data={state.billingLedger} />;
  if (route.id === "admin.runtime") return <AdminRuntimePage route={route} operator={operator} />;
  if (route.id === "admin.support") return <ResourceListPage route={route} title="Support Ops" eyebrow="Admin" data={tickets.tickets} />;
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
      title="Overview"
      eyebrow="OPL Console"
      subtitle="Wallet, Workspace delivery, Gateway usage, Support"
      extra={<Button type="primary" icon={<Plus size={15} />} href={routeTo("workspace.create")}>创建 Workspace</Button>}
    >
      <MetricStrip
        items={[
          { label: "可用余额", value: money(usable), caption: `${money(wallet?.frozen)} frozen`, icon: <WalletCards size={16} />, tone: usable > 0 ? "good" : "warn" },
          { label: "Workspace", value: state.workspaces.length, caption: `${computeRunning} compute running`, icon: <Server size={16} />, tone: computeRunning ? "good" : "neutral" },
          { label: "存储资源", value: storageAvailable, caption: "StorageVolume", icon: <HardDrive size={16} />, tone: storageAvailable ? "info" : "neutral" },
          { label: "Gateway 请求", value: state.requestUsageLogs.length, caption: "gflabtoken.cn", icon: <LinkIcon size={16} />, tone: "info" },
          { label: "工单", value: activeTickets, caption: `${tickets?.tickets.length || 0} total`, icon: <Headphones size={16} />, tone: activeTickets ? "warn" : "neutral" },
          { label: "告警", value: state.notifications.length, caption: "owner visible", icon: <AlertTriangle size={16} />, tone: state.notifications.length ? "danger" : "good" },
        ]}
      />

      <div className="consoleGrid">
        <InsightPanel
          title="业务链"
          eyebrow="Launch loop"
          actions={<ActionGroup actions={[
            { label: "Workspace", type: "primary", icon: <Plus size={15} />, onClick: () => { window.location.href = routeTo("workspace.create"); } },
            { label: "钱包", icon: <WalletCards size={15} />, onClick: () => { window.location.href = routeTo("billing.wallet"); } },
            { label: "工单", icon: <Headphones size={15} />, onClick: () => { window.location.href = routeTo("support.create"); } },
          ]} />}
        >
          <ResourceSplit
            items={[
              { label: "充值与冻结", value: `${money(wallet?.balance)} / ${money(wallet?.frozen)}`, meta: "Balance / frozen hold", status: `${Math.round(freezeRatio)}% frozen`, tone: freezeRatio > 70 ? "warn" : "info" },
              { label: "计算交付", value: `${computeRunning}/${state.computeAllocations.length}`, meta: "Running compute allocations", status: computeRunning ? "active" : "idle", tone: computeRunning ? "good" : "neutral" },
              { label: "存储资源", value: storageAvailable, meta: "StorageVolume persistent data", status: storageAvailable ? "available" : "empty", tone: storageAvailable ? "good" : "neutral" },
              { label: "URL 分发", value: state.workspaces.filter((workspace) => (workspace.access as { tokenStatus?: string } | undefined)?.tokenStatus === "active").length, meta: "Active Workspace URLs", status: "scoped", tone: "info" },
              { label: "支持闭环", value: activeTickets, meta: "Open support tickets", status: activeTickets ? "open" : "clear", tone: activeTickets ? "warn" : "good" },
            ]}
          />
        </InsightPanel>

        <InsightPanel title="最近信号" eyebrow="Attention">
          <TimelineList items={recentSignals} emptyText="当前没有告警或待处理工单" />
        </InsightPanel>
      </div>

      <InsightPanel title="最近 Workspace" eyebrow="Delivery">
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
      title="Workspaces"
      eyebrow="Delivery"
      subtitle="Compute allocation, storage volume, URL token, billing hold"
      extra={<Button type="primary" icon={<Plus size={15} />} href={routeTo("workspace.create")}>创建 Workspace</Button>}
    >
      <MetricStrip
        items={[
          { label: "总数", value: state.workspaces.length, caption: "owned by this lab", tone: state.workspaces.length ? "info" : "neutral" },
          { label: "运行中", value: running, caption: "Workspace URL entries", tone: running ? "good" : "neutral" },
          { label: "URL 可用", value: activeUrls, caption: "shareable links", tone: activeUrls ? "good" : "warn" },
          { label: "已挂载", value: attachedEntries, caption: "attachment-backed entries", tone: attachedEntries ? "info" : "neutral" },
          { label: "可用余额", value: money(available(wallet)), caption: "after frozen hold", tone: available(wallet) > 0 ? "good" : "warn" },
        ]}
      />
      <InsightPanel title="Workspace 对象" eyebrow="Current">
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
      emptyText="暂无 Workspace"
      columns={[
        { title: "名称", dataIndex: "name", render: (_: unknown, row: Record<string, unknown>) => <Button type="link" href={routeTo("workspace.detail", { id: String(row.id) })}>{String(row.name || row.id)}</Button> },
        { title: "状态", dataIndex: "state", render: (_: unknown, row: Record<string, unknown>) => <StatusPill label={statusLabel(row)} tone={toneForStatus(row.state)} /> },
        { title: "套餐", dataIndex: "packageId", render: (value: unknown) => packageText(planById[String(value)]) },
        { title: "计算分配", render: (_: unknown, row: Record<string, unknown>) => <Typography.Text ellipsis>{String(row.computeAllocationId || row.computeId || "-")}</Typography.Text> },
        { title: "存储卷", render: (_: unknown, row: Record<string, unknown>) => <Typography.Text ellipsis>{String(row.storageId || "-")}</Typography.Text> },
        { title: "Workspace URL", dataIndex: "url", ellipsis: true, render: (value: unknown) => <Typography.Text className="inlineCode">{String(value || "-")}</Typography.Text> },
      ]}
    />
  );
}

function ComputePage({ route, state }: { route: OplRoute; state: ConsoleState }) {
  const running = state.computeAllocations.filter((item) => item.status === "running").length;
  return (
    <ConsoleSurface
      title="Compute"
      eyebrow="TKE allocations"
      subtitle="Account-owned dedicated CVM allocations"
      extra={<Button type="primary" icon={<Plus size={15} />} href="/console/compute/new">开通计算</Button>}
    >
      <MetricStrip
        items={[
          { label: "计算分配", value: state.computeAllocations.length, caption: "owned by this account", tone: state.computeAllocations.length ? "info" : "neutral" },
          { label: "运行中", value: running, caption: "billable CVM runtime", tone: "good" },
        ]}
      />
      <ResourceListPage route={route} title="计算分配" eyebrow="ComputeAllocation" data={state.computeAllocations} embedded />
    </ConsoleSurface>
  );
}

function BillingPage({ route, state, wallet }: { route: OplRoute; state: ConsoleState; wallet: ConsoleState["wallet"] }) {
  const usable = available(wallet);
  const frozenPercent = wallet.balance > 0 ? Math.min(100, Math.round((wallet.frozen / wallet.balance) * 100)) : 0;
  return (
    <ConsoleSurface title="Billing" eyebrow="Wallet" subtitle="Prepaid balance, holds, resource usage">
      <MetricStrip
        items={[
          { label: "可用", value: money(usable), caption: "can open compute or storage", tone: usable > 0 ? "good" : "warn" },
          { label: "冻结", value: money(wallet.frozen), caption: `${frozenPercent}% of balance`, tone: frozenPercent > 70 ? "warn" : "info" },
          { label: "余额", value: money(wallet.balance), caption: "available plus frozen", tone: "neutral" },
          { label: "累计充值", value: money(wallet.totalRecharged), caption: "manual top-up ledger", tone: "good" },
          { label: "扣费记录", value: state.walletTransactions.length, caption: "recent resource events", tone: state.walletTransactions.length ? "info" : "neutral" },
        ]}
      />
      <div className="consoleGrid">
        <InsightPanel title="钱包拆分" eyebrow="Balance">
          <div className="stackList">
            <div className="walletBar"><span style={{ width: `${frozenPercent}%` }} /></div>
            <div className="stackRow"><span>可用余额</span><strong>{money(usable)}</strong></div>
            <div className="stackRow"><span>冻结金额</span><strong>{money(wallet.frozen)}</strong></div>
            <div className="stackRow"><span>总余额</span><strong>{money(wallet.balance)}</strong></div>
          </div>
        </InsightPanel>
        <InsightPanel title="钱包流水" eyebrow="Transactions">
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
    <ConsoleSurface title="Account & Lab" eyebrow="Management" subtitle="Account, organization, and billing scope">
      <ResourceSplit items={[
        { label: "Account", value: state.account.name, meta: state.account.id, status: state.account.status, tone: "good" },
        { label: "Wallet", value: money(state.wallet.balance), meta: `${money(state.wallet.frozen)} frozen`, status: "active", tone: "info" },
      ]} />
      <ContractPanel route={route} />
    </ConsoleSurface>
  );
}

function AlertsPage({ route, state }: { route: OplRoute; state: ConsoleState }) {
  return (
    <ConsoleSurface title="Alerts" eyebrow="Signals" subtitle="Owner-visible runtime and support signals">
      <InsightPanel title="告警" eyebrow="Runtime">
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
    <ConsoleSurface title="Gateway" eyebrow="External" subtitle="OPL Gateway external integration">
      <ResourceSplit items={[
        { label: "Gateway", value: "External link", meta: "Console owns usage and policy visibility", status: "external", tone: "info" },
        { label: "Boundary", value: "Not implemented here", meta: "Provider routing belongs to Gateway", status: "scoped", tone: "good" },
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
  if (!resource) return <ConsoleSurface title={title} eyebrow="Detail"><Empty description="未找到资源" /></ConsoleSurface>;
  return (
    <ConsoleSurface title={String(resource.name || resource.id)} eyebrow={`${title} detail`}>
      <ResourceSplit items={Object.entries(resource).slice(0, 6).map(([key, value]) => ({
        label: key,
        value: String(value && typeof value === "object" ? JSON.stringify(value) : value || "-"),
        status: key.toLowerCase().includes("status") || key === "state" ? String(value) : undefined,
        tone: key.toLowerCase().includes("status") || key === "state" ? toneForStatus(value) : "neutral",
      }))} />
      <ContractPanel route={route} />
    </ConsoleSurface>
  );
}

function CreateResourcePage({ route, title, action, fields }: { route: OplRoute; title: string; action: string; fields: string[] }) {
  return (
    <ConsoleSurface title={title} eyebrow="Provision" subtitle={route.serviceBoundary} compact>
      <InsightPanel title={action} eyebrow={route.objectKind || route.routeKind}>
        <Form layout="vertical" initialValues={{ packageId: "basic", sizeGb: 10, mountPath: "/data" }}>
          {fields.map((field) => (
            <Form.Item key={field} name={field} label={labelize(field)} rules={[{ required: field !== "body" }]}>
              {field === "packageId" ? (
                <Select options={[{ label: "Basic Workspace", value: "basic" }, { label: "Pro Workspace", value: "pro" }]} />
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
    <ConsoleSurface title="Admin Overview" eyebrow="Operations" subtitle="Accounts, Workspaces, runtime evidence">
      <MetricStrip
        items={[
          { label: "账号", value: valueFrom(operator, "accounts.total", 1), caption: "managed billing accounts", tone: "info" },
          { label: "Workspace", value: valueFrom(operator, "workspaces.total", state.workspaces.length), caption: "running workspace entries", tone: "good" },
          { label: "失败操作", value: valueFrom(operator, "runtimeOperations.failed", 0), caption: "runtime operation failures", tone: "good" },
          { label: "冻结总额", value: money(valueFrom(operator, "accounts.frozen", state.wallet.frozen)), caption: "all accounts", tone: "warn" },
          { label: "告警", value: valueFrom(operator, "notifications.total", state.notifications.length), caption: "operator visible", tone: "neutral" },
        ]}
      />
      <div className="consoleGrid equal">
        <InsightPanel title="运行态" eyebrow="Runtime">
          <ResourceSplit items={[
            { label: "Runtime", value: "Ready", meta: "runtime readiness", status: "pass", tone: "good" },
            { label: "Launch", value: "Blocked", meta: "production launch gates", status: "check", tone: "warn" },
            { label: "计算分配", value: state.computeAllocations.length, meta: "CVM allocation evidence", status: "tracked", tone: "info" },
          ]} />
        </InsightPanel>
        <InsightPanel title="最近告警" eyebrow="Signals">
          <TimelineList items={[]} emptyText="暂无运营告警" />
        </InsightPanel>
      </div>
      <ContractPanel route={route} />
    </ConsoleSurface>
  );
}

function AdminUsersPage({ route, management }: { route: OplRoute; management?: ManagementState }) {
  return (
    <ConsoleSurface title="Users" eyebrow="Admin" subtitle="Login users, billing accounts, and wallet operations" extra={<Button type="primary" icon={<Plus size={15} />}>新建用户</Button>}>
      <InsightPanel title="用户钱包" eyebrow="Management">
        <ObjectTable data={management?.users || []} emptyText="暂无用户" columns={columnsFor(management?.users || [])} />
      </InsightPanel>
      <ContractPanel route={route} />
    </ConsoleSurface>
  );
}

function AdminRuntimePage({ route }: { route: OplRoute; operator?: Record<string, unknown> }) {
  return (
    <ConsoleSurface title="Runtime" eyebrow="Admin" subtitle="Readiness gates and launch blockers">
      <div className="consoleGrid equal">
        <InsightPanel title="Readiness" eyebrow="Runtime">
          <ResourceSplit items={[
            { label: "Fabric", value: "Ready", meta: "runtime provider", status: "pass", tone: "good" },
            { label: "Launch", value: "Blocked", meta: "production gates", status: "check", tone: "warn" },
            { label: "Env", value: 0, meta: "missing environment inputs", status: "env", tone: "good" },
            { label: "Tools", value: 0, meta: "host tool checks", status: "tools", tone: "good" },
          ]} />
        </InsightPanel>
        <InsightPanel title="Blockers" eyebrow="Checks">
          <TimelineList items={[]} emptyText="No blockers" />
        </InsightPanel>
      </div>
      <ContractPanel route={route} />
    </ConsoleSurface>
  );
}

function ContractPanel({ route }: { route: OplRoute }) {
  return (
    <InsightPanel title="API contract" eyebrow="Boundary">
      <ResourceSplit items={(route.apiRoutes || ["No direct API route"]).map((apiRoute) => ({
        label: String(apiRoute).split(" ")[0] || "API",
        value: apiRoute,
        meta: route.serviceBoundary,
        status: "fixed",
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
  if (typeof value === "boolean") return value ? "Yes" : "No";
  if (value === null || value === undefined || value === "") return "-";
  if (typeof value === "object") return JSON.stringify(value);
  return String(value);
}

function toneForStatus(value: unknown) {
  const color = statusColor(value);
  if (color === "green") return "good" as const;
  if (color === "red") return "danger" as const;
  if (color === "orange") return "warn" as const;
  return "info" as const;
}

function valueFrom(source: Record<string, unknown> | undefined, path: string, fallback: string | number) {
  if (!source) return fallback;
  return path.split(".").reduce<unknown>((current, key) => {
    if (current && typeof current === "object" && key in current) return (current as Record<string, unknown>)[key];
    return fallback;
  }, source) as string | number;
}

function labelize(value: string) {
  return value.replace(/([a-z])([A-Z])/g, "$1 $2").replace(/^./, (char) => char.toUpperCase());
}

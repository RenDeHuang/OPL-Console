export function money(value: unknown) {
  const numeric = Number(value || 0);
  return `¥${numeric.toLocaleString("zh-CN", { maximumFractionDigits: 2 })}`;
}

export function available(wallet: { balance?: number; frozen?: number } | undefined) {
  return Number(wallet?.balance || 0) - Number(wallet?.frozen || 0);
}

export function statusColor(value: unknown) {
  const normalized = String(value || "pending").toLowerCase();
  if (["running", "ready", "active", "available", "attached", "bound"].includes(normalized)) return "green";
  if (["failed", "error", "destroyed", "deleted", "detached"].includes(normalized)) return "red";
  if (["creating", "pending", "starting", "stopping"].includes(normalized)) return "orange";
  return "blue";
}

export function statusLabel(value: unknown) {
  if (value && typeof value === "object" && "state" in value) return valueLabel((value as { state?: unknown }).state || "pending");
  return valueLabel(value || "pending");
}

export function packageText(plan: Record<string, unknown> | undefined) {
  if (!plan) return "-";
  return `${plan.name || plan.id} · ${plan.cpu || "-"} CPU / ${plan.memoryGb || "-"}GB`;
}

export function valueLabel(value: unknown) {
  const normalized = String(value || "").toLowerCase();
  const labels: Record<string, string> = {
    active: "有效",
    admin: "管理员",
    attached: "已挂载",
    available: "可用",
    blocked: "阻塞",
    bound: "已绑定",
    check: "待检查",
    clear: "已清理",
    creating: "创建中",
    deleted: "已删除",
    destroyed: "已销毁",
    detached: "已卸载",
    empty: "空",
    env: "环境",
    error: "错误",
    external: "外部",
    failed: "失败",
    fixed: "已固定",
    good: "正常",
    hold: "冻结",
    idle: "空闲",
    info: "信息",
    lab_owner: "实验室所有者",
    open: "待处理",
    pass: "通过",
    pending: "待处理",
    ready: "就绪",
    running: "运行中",
    scoped: "已限定",
    signal: "信号",
    starting: "启动中",
    stopping: "停止中",
    stopped: "已停止",
    tools: "工具",
    topup: "充值",
    warn: "警告",
  };
  return labels[normalized] || String(value || "-");
}

export function fieldLabel(value: string) {
  const labels: Record<string, string> = {
    access: "访问",
    accountId: "账号 ID",
    attachmentId: "挂载 ID",
    available: "可用",
    balance: "余额",
    billingAccount: "计费账号",
    billingAccountId: "计费账号 ID",
    body: "内容",
    category: "分类",
    computeAllocationId: "计算分配 ID",
    computeId: "计算 ID",
    cpu: "CPU",
    createdAt: "创建时间",
    diskGb: "磁盘 GB",
    email: "邮箱",
    frozen: "冻结金额",
    id: "ID",
    memoryGb: "内存 GB",
    mountPath: "挂载路径",
    name: "名称",
    packageId: "套餐",
    poolId: "资源池",
    price: "价格",
    reason: "原因",
    role: "角色",
    runtimeStatus: "运行状态",
    server: "服务器",
    severity: "级别",
    sizeGb: "容量 GB",
    state: "状态",
    status: "状态",
    storageId: "存储卷 ID",
    subject: "主题",
    targetAccountId: "目标账号",
    title: "标题",
    token: "访问令牌",
    tokenStatus: "令牌状态",
    totalRecharged: "累计充值",
    type: "类型",
    url: "访问链接",
    workspaceId: "工作区 ID",
  };
  return labels[value] || value.replace(/([a-z])([A-Z])/g, "$1 $2");
}

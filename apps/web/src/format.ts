export function statusText(value?: string) {
  const labels: Record<string, string> = {
    active: "启用",
    approved: "已通过",
    approval_required: "待审批",
    attached: "已挂载",
    available: "可用",
    cancelled: "已取消",
    configured: "已配置",
    deleted: "已删除",
    disabled: "停用",
    managed: "托管",
    open: "处理中",
    pending: "待处理",
    ready: "就绪",
    rejected: "已拒绝",
    running: "运行中",
    suspended: "已暂停",
    token_deleted: "地址已停用"
  };
  return labels[value ?? ""] ?? value ?? "";
}

export function resourceText(value?: string) {
  const labels: Record<string, string> = {
    attachment: "挂载",
    compute: "计算",
    route: "访问地址",
    storage: "存储",
    workspace: "工作空间"
  };
  return labels[value ?? ""] ?? value ?? "";
}

export function packageText(value?: string) {
  const labels: Record<string, string> = {
    "Basic Workspace": "基础套餐",
    "Pro Workspace": "专业套餐",
    basic: "基础套餐",
    pro: "专业套餐"
  };
  return labels[value ?? ""] ?? value ?? "";
}

export function ledgerKindText(value?: string) {
  const labels: Record<string, string> = {
    compute_hold: "计算冻结",
    governance_receipt: "治理凭证",
    manual_topup: "人工充值",
    request_debit: "请求扣费",
    storage_hold: "存储冻结",
    workspace_hold: "工作空间冻结"
  };
  return labels[value ?? ""] ?? value ?? "";
}

export function policyTypeText(value?: string) {
  const labels: Record<string, string> = {
    quota: "配额",
    workspace_lifecycle: "工作空间生命周期"
  };
  return labels[value ?? ""] ?? value ?? "";
}

export function scopeText(value?: string) {
  const labels: Record<string, string> = {
    organization: "组织",
    system: "系统",
    team: "团队"
  };
  return labels[value ?? ""] ?? value ?? "";
}

export function providerText(value?: string) {
  const labels: Record<string, string> = {
    local: "本地运行",
    mock: "模拟运行",
    tke: "腾讯云 TKE",
    "tencent-tke": "腾讯云 TKE"
  };
  return labels[value ?? ""] ?? value ?? "控制台托管";
}

export function actionText(value?: string) {
  const labels: Record<string, string> = {
    "workspace.create": "开通工作空间",
    "workspace.configure": "配置工作空间",
    "workspace.delete": "删除工作空间",
    "workspace.suspend": "暂停工作空间",
    "workspace.token.delete": "停用访问地址",
    "workspace.token.reset": "重置访问地址"
  };
  return labels[value ?? ""] ?? value ?? "请求";
}

export function fen(value?: number) {
  return `¥${((value ?? 0) / 100).toFixed(2)}`;
}

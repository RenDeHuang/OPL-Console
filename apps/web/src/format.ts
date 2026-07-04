export function statusText(value?: string) {
  const labels: Record<string, string> = {
    active: "启用",
    approved: "已通过",
    approval_required: "待审批",
    attached: "已挂载",
    available: "可用",
    cancelled: "已取消",
    creating_storage_backup: "创建备份中",
    destroyed: "已销毁",
    destroying_disk: "销毁存储中",
    destroying_server: "销毁计算中",
    deleted: "已删除",
    detached_retained: "已卸载并保留",
    disabled: "停用",
    failed: "失败",
    managed: "托管",
    open: "处理中",
    pending: "待处理",
    ready: "就绪",
    rejected: "已拒绝",
    restoring_storage_backup: "恢复备份中",
    running: "运行中",
    server_destroyed_disk_retained: "计算已销毁，存储保留",
    stopped: "已停止",
    stopped_server_disk_retained: "计算已停止，存储保留",
    stopping_server: "停止计算中",
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
    "workspace.compute.destroy": "销毁计算",
    "workspace.compute.restart": "重启计算",
    "workspace.compute.stop": "停止计算",
    "workspace.storage.backup": "创建存储备份",
    "workspace.storage.destroy": "销毁存储",
    "workspace.storage.restore": "恢复存储备份",
    "workspace.token.delete": "停用访问地址",
    "workspace.token.reset": "重置访问地址"
  };
  return labels[value ?? ""] ?? value ?? "请求";
}

export function fen(value?: number) {
  return `¥${((value ?? 0) / 100).toFixed(2)}`;
}

export function lifecycleStepText(value?: string) {
  const labels: Record<string, string> = {
    attach_storage: "挂载存储",
    create_compute: "开通计算",
    create_route: "生成访问地址",
    create_storage: "开通存储"
  };
  return labels[value ?? ""] ?? value ?? "";
}

export function readinessText(value: string) {
  const labels: Record<string, string> = {
    "auth.seed_without_defaults": "默认账号禁用",
    "console.public_https_url": "公网 HTTPS",
    "database.postgres_url": "PostgreSQL",
    "fabric.external_contract": "Fabric API 合同",
    "fabric.provider": "Fabric 提供方",
    "kubernetes.config": "K8s 配置",
    "kubernetes.ingress_class": "Ingress",
    "kubernetes.namespace": "命名空间",
    "kubernetes.storage_class": "存储类",
    "ledger.external_contract": "Ledger API 合同",
    "registry.workspace_image": "工作空间镜像",
    "secrets.fabric_token": "Fabric 密钥",
    "secrets.ledger_token": "Ledger 密钥",
    "workspace.domain": "工作空间域名"
  };
  return labels[value] ?? value;
}

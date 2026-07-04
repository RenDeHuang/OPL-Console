export function money(value: unknown) {
  const numeric = Number(value || 0);
  return `$${numeric.toLocaleString(undefined, { maximumFractionDigits: 2 })}`;
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
  if (value && typeof value === "object" && "state" in value) return String((value as { state?: unknown }).state || "pending");
  return String(value || "pending");
}

export function packageText(plan: Record<string, unknown> | undefined) {
  if (!plan) return "-";
  return `${plan.name || plan.id} · ${plan.cpu || "-"} CPU / ${plan.memoryGb || "-"}GB`;
}

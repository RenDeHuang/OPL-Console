import type { ReactNode } from "react";
import { PageContainer, ProTable } from "@ant-design/pro-components";
import { Button, Empty, List, Space, Tag, Typography } from "antd";

type Tone = "good" | "warn" | "danger" | "info" | "neutral";

function toneClass(tone: Tone = "neutral") {
  return ["good", "warn", "danger", "info"].includes(tone) ? tone : "neutral";
}

function tagColor(tone: Tone = "neutral") {
  return {
    good: "green",
    warn: "orange",
    danger: "red",
    info: "blue",
    neutral: "default",
  }[toneClass(tone)];
}

export function ConsoleSurface({ title, eyebrow, subtitle, extra, children, compact = false }: {
  title: ReactNode;
  eyebrow?: ReactNode;
  subtitle?: ReactNode;
  extra?: ReactNode;
  children: ReactNode;
  compact?: boolean;
}) {
  return (
    <PageContainer
      title={(
        <div className="surfaceTitle">
          {eyebrow && <span>{eyebrow}</span>}
          <strong>{title}</strong>
        </div>
      )}
      subTitle={subtitle}
      extra={extra}
    >
      <div className={compact ? "consoleSurface compact" : "consoleSurface"}>
        {children}
      </div>
    </PageContainer>
  );
}

export function MetricStrip({ items = [] }: { items: Array<{ label: ReactNode; value: ReactNode; caption?: ReactNode; icon?: ReactNode; tone?: Tone }> }) {
  return (
    <section className="metricStrip" aria-label="Console metrics">
      {items.map((item) => (
        <article className={`metricTile ${toneClass(item.tone)}`} key={String(item.label)}>
          <div className="metricTopline">
            <span>{item.label}</span>
            {item.icon && <span className="metricIcon">{item.icon}</span>}
          </div>
          <strong>{item.value}</strong>
          {item.caption && <small>{item.caption}</small>}
        </article>
      ))}
    </section>
  );
}

export function InsightPanel({ title, eyebrow, actions, children, tone = "neutral", className = "" }: {
  title: ReactNode;
  eyebrow?: ReactNode;
  actions?: ReactNode;
  children: ReactNode;
  tone?: Tone;
  className?: string;
}) {
  return (
    <section className={`insightPanel ${toneClass(tone)} ${className}`.trim()}>
      <header>
        <div>
          {eyebrow && <span>{eyebrow}</span>}
          <h2>{title}</h2>
        </div>
        {actions && <div className="panelActions">{actions}</div>}
      </header>
      <div className="panelBody">{children}</div>
    </section>
  );
}

export function StatusPill({ label, tone = "neutral" }: { label: ReactNode; tone?: Tone }) {
  return <Tag color={tagColor(tone)} className="statusPill">{label}</Tag>;
}

export function ResourceSplit({ items = [] }: { items: Array<{ label: ReactNode; value: ReactNode; meta?: ReactNode; status?: ReactNode; tone?: Tone }> }) {
  return (
    <div className="resourceSplit">
      {items.map((item) => (
        <article key={String(item.label)}>
          <div className="resourceLabel">
            <span>{item.label}</span>
            {item.status && <StatusPill label={item.status} tone={item.tone} />}
          </div>
          <strong>{item.value}</strong>
          {item.meta && <small>{item.meta}</small>}
        </article>
      ))}
    </div>
  );
}

export function ActionGroup({ actions = [] }: { actions: Array<ReactNode | { key?: string; label: ReactNode; type?: "primary" | "default"; danger?: boolean; icon?: ReactNode; disabled?: boolean; onClick?: () => void }> }) {
  return (
    <Space wrap size={8} className="actionGroup">
      {actions.map((action) => {
        if (!action || typeof action !== "object" || !("label" in action)) return action;
        return (
          <Button
            key={action.key || String(action.label)}
            type={action.type}
            danger={action.danger}
            icon={action.icon}
            disabled={action.disabled}
            onClick={action.onClick}
          >
            {action.label}
          </Button>
        );
      })}
    </Space>
  );
}

export function TimelineList({ items = [], emptyText = "暂无记录" }: { items: Array<{ title: ReactNode; description?: ReactNode; meta?: ReactNode; tone?: Tone }>; emptyText?: string }) {
  if (!items.length) {
    return <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={emptyText} />;
  }

  return (
    <List
      className="timelineList"
      dataSource={items}
      renderItem={(item) => (
        <List.Item>
          <div className={`timelineDot ${toneClass(item.tone)}`} />
          <div className="timelineContent">
            <strong>{item.title}</strong>
            {item.description && <Typography.Text type="secondary">{item.description}</Typography.Text>}
          </div>
          {item.meta && <Typography.Text type="secondary" className="timelineMeta">{item.meta}</Typography.Text>}
        </List.Item>
      )}
    />
  );
}

export function ObjectTable({ rowKey = "id", data = [], columns = [], emptyText = "暂无数据", ...rest }: {
  rowKey?: string | ((row: Record<string, unknown>) => string);
  data?: Record<string, unknown>[];
  columns?: Array<Record<string, unknown>>;
  emptyText?: string;
  [key: string]: unknown;
}) {
  return (
    <ProTable
      className="objectTable"
      rowKey={rowKey}
      search={false}
      options={false}
      pagination={false}
      size="small"
      scroll={{ x: "max-content" }}
      dataSource={data}
      columns={columns}
      locale={{ emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={emptyText} /> }}
      {...rest}
    />
  );
}

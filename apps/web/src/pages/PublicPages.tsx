import type { FormEvent, ReactNode } from "react";
import { useState } from "react";
import { ArrowLeft, ArrowRight, Database, Headphones, KeyRound, LogIn, Server, ShieldCheck, WalletCards } from "lucide-react";
import type { OplRoute } from "../routes/oplRoutes";
import { loginOwner } from "../api/consoleApi";

export function HomePage() {
  const target = "/login";
  return (
    <div className="publicShell">
      <header className="publicNav">
        <a className="wordmark" href="/">
          <span>OPL</span>
          <strong>OPL Console</strong>
        </a>
        <nav className="publicLinks">
          <a href="/console/workspaces">Workspace</a>
          <a href="/console/billing">Billing</a>
          <a href="/console/support">Support</a>
          <a className="navButton" href={target}>登录</a>
        </nav>
      </header>

      <main>
        <section className="publicConsole">
          <div className="publicConsoleCopy">
            <p className="eyebrow">OPL Console</p>
            <h1>OPL Console</h1>
            <p>开通 Workspace，分发访问 URL，按计算、存储和 Gateway 请求扣费。</p>
            <a className="primaryLink" href={target}>进入控制台 <ArrowRight size={16} /></a>
          </div>

          <div className="publicConsolePanel" aria-label="OPL Console product surface">
            <div className="publicPanelTop">
              <strong>Business chain</strong>
              <span>Live Console</span>
            </div>
            <div className="publicMetrics">
              <PublicMetric icon={<WalletCards />} label="Wallet" value="Balance + holds" />
              <PublicMetric icon={<Server />} label="Workspace" value="Compute + storage" />
              <PublicMetric icon={<KeyRound />} label="URL" value="Scoped access" />
              <PublicMetric icon={<Database />} label="Ledger" value="Usage evidence" />
            </div>
            <div className="publicFlow">
              <span>Top up</span>
              <span>Create</span>
              <span>Share URL</span>
              <span>Meter</span>
            </div>
          </div>
        </section>

        <section className="homeBand">
          <article>
            <ShieldCheck />
            <h2>Lab Owner</h2>
            <p>余额、Workspace、URL、工单。</p>
          </article>
          <article>
            <WalletCards />
            <h2>Billing</h2>
            <p>充值、冻结、小时扣费。</p>
          </article>
          <article>
            <Headphones />
            <h2>Admin</h2>
            <p>用户、充值、运行证据。</p>
          </article>
        </section>
      </main>
    </div>
  );
}

export function LoginPage({ route }: { route?: OplRoute }) {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const routeMode = route?.path || "/login";

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitting(true);
    setError("");
    try {
      const session = await loginOwner({ email, password });
      window.location.href = session.user.role === "admin" ? "/admin/overview" : "/console/overview";
    } catch (err) {
      setError(err instanceof Error ? err.message : "login_failed");
    } finally {
      setSubmitting(false);
    }
  }

  if (routeMode !== "/login" && routeMode !== "/logout") {
    return (
      <div className="loginShell">
        <a className="backLink" href="/"><ArrowLeft size={16} /> OPL Console</a>
        <main className="loginPanel compactAuth">
          <div className="loginBrand">
            <div className="brandIcon">OPL</div>
            <div>
              <p className="eyebrow">OPL Console</p>
              <h1>无法访问</h1>
            </div>
          </div>
          <div className="emptyState">
            <strong>当前入口不可用</strong>
            <span>请使用已开通的 Console 账号登录。</span>
          </div>
          <a className="primaryLink" href="/login">返回登录</a>
        </main>
      </div>
    );
  }

  return (
    <div className="loginShell">
      <a className="backLink" href="/"><ArrowLeft size={16} /> OPL Console</a>
      <main className="loginPanel">
        <div className="loginBrand">
          <div className="brandIcon">OPL</div>
          <div>
            <p className="eyebrow">OPL Console</p>
            <h1>登录</h1>
          </div>
        </div>
        <form onSubmit={submit}>
          <label>
            邮箱
            <input value={email} onChange={(event) => setEmail(event.target.value)} type="email" autoComplete="email" required />
          </label>
          <label>
            密码
            <input value={password} onChange={(event) => setPassword(event.target.value)} type="password" autoComplete="current-password" required />
          </label>
          {error && <div className="error">{errorLabel(error)}</div>}
          <button className="primary wide" disabled={submitting}>
            <LogIn size={16} /> {submitting ? "登录中..." : "登录"}
          </button>
        </form>
        <div className="securityNote">
          <ShieldCheck size={16} />
          <span>Secure cookie + CSRF</span>
        </div>
      </main>
    </div>
  );
}

function errorLabel(value: string) {
  const labels: Record<string, string> = {
    login_failed: "登录失败",
    invalid_credentials: "邮箱或密码不正确",
    request_failed: "请求失败",
  };
  return labels[value] || value;
}

export function ErrorPage({ code, title }: { code: string; title: string }) {
  return (
    <div className="loginShell">
      <a className="backLink" href="/"><ArrowLeft size={16} /> OPL Console</a>
      <main className="loginPanel compactAuth">
        <p className="eyebrow">{code}</p>
        <div className="emptyState">
          <strong>{title}</strong>
          <span>请返回已开通的 Console 页面。</span>
        </div>
        <a className="primaryLink" href="/console/overview">返回 Console</a>
      </main>
    </div>
  );
}

function PublicMetric({ icon, label, value }: { icon: ReactNode; label: string; value: string }) {
  return (
    <article className="publicMetric">
      {icon}
      <div>
        <strong>{label}</strong>
        <span>{value}</span>
      </div>
    </article>
  );
}

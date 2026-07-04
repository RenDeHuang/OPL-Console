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
          <strong>OPL 控制台</strong>
        </a>
        <nav className="publicLinks">
          <a href="/console/workspaces">工作区</a>
          <a href="/console/billing">计费</a>
          <a href="/console/support">支持</a>
          <a className="navButton" href={target}>登录</a>
        </nav>
      </header>

      <main>
        <section className="publicConsole">
          <div className="publicConsoleCopy">
            <p className="eyebrow">OPL 控制台</p>
            <h1>OPL 控制台</h1>
            <p>开通工作区，分发访问链接，按计算、存储和网关请求扣费。</p>
            <a className="primaryLink" href={target}>进入控制台 <ArrowRight size={16} /></a>
          </div>

          <div className="publicConsolePanel" aria-label="OPL 控制台产品界面">
            <div className="publicPanelTop">
              <strong>业务链路</strong>
              <span>控制台</span>
            </div>
            <div className="publicMetrics">
              <PublicMetric icon={<WalletCards />} label="钱包" value="余额与冻结" />
              <PublicMetric icon={<Server />} label="工作区" value="计算与存储" />
              <PublicMetric icon={<KeyRound />} label="访问链接" value="范围化访问" />
              <PublicMetric icon={<Database />} label="账本" value="用量凭证" />
            </div>
            <div className="publicFlow">
              <span>充值</span>
              <span>创建</span>
              <span>分发链接</span>
              <span>计量</span>
            </div>
          </div>
        </section>

        <section className="homeBand">
          <article>
            <ShieldCheck />
            <h2>实验室所有者</h2>
            <p>余额、工作区、访问链接、工单。</p>
          </article>
          <article>
            <WalletCards />
            <h2>计费</h2>
            <p>充值、冻结、小时扣费。</p>
          </article>
          <article>
            <Headphones />
            <h2>管理</h2>
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
        <a className="backLink" href="/"><ArrowLeft size={16} /> OPL 控制台</a>
        <main className="loginPanel compactAuth">
          <div className="loginBrand">
            <div className="brandIcon">OPL</div>
            <div>
              <p className="eyebrow">OPL 控制台</p>
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
      <a className="backLink" href="/"><ArrowLeft size={16} /> OPL 控制台</a>
      <main className="loginPanel">
        <div className="loginBrand">
          <div className="brandIcon">OPL</div>
          <div>
            <p className="eyebrow">OPL 控制台</p>
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
          <span>安全 Cookie 与 CSRF 防护</span>
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
      <a className="backLink" href="/"><ArrowLeft size={16} /> OPL 控制台</a>
      <main className="loginPanel compactAuth">
        <p className="eyebrow">{code}</p>
        <div className="emptyState">
          <strong>{title}</strong>
          <span>请返回已开通的 Console 页面。</span>
        </div>
        <a className="primaryLink" href="/console/overview">返回控制台</a>
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

import { useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { LogIn } from "lucide-react";
import { api } from "../api/client";

export function LoginPage() {
  const [email, setEmail] = useState("admin@opl.local");
  const [password, setPassword] = useState("");
  const login = useMutation({
    mutationFn: () => api.login(email, password),
    onSuccess: () => {
      window.location.href = "/console/overview";
    }
  });

  return (
    <main className="shell narrow">
      <h1>OPL 控制台</h1>
      <form
        className="panel"
        method="post"
        onSubmit={(event) => {
          event.preventDefault();
          login.mutate();
        }}
        >
        <label>
          邮箱
          <input type="email" name="email" autoComplete="email" value={email} onChange={(event) => setEmail(event.target.value)} />
        </label>
        <label>
          密码
          <input type="password" name="password" autoComplete="current-password" value={password} onChange={(event) => setPassword(event.target.value)} />
        </label>
        <button type="submit" disabled={login.isPending}>
          <LogIn size={16} />
          {login.isPending ? "登录中" : "登录"}
        </button>
        {login.isError ? <p className="error">登录失败，请检查账号或密码。</p> : null}
      </form>
    </main>
  );
}

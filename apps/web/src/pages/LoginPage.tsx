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
      window.location.href = "/";
    }
  });

  return (
    <main className="shell narrow">
      <h1>OPL Console</h1>
      <form
        className="panel"
        method="post"
        onSubmit={(event) => {
          event.preventDefault();
          login.mutate();
        }}
      >
        <label>
          Email
          <input type="email" name="email" autoComplete="email" value={email} onChange={(event) => setEmail(event.target.value)} />
        </label>
        <label>
          Password
          <input type="password" name="password" autoComplete="current-password" value={password} onChange={(event) => setPassword(event.target.value)} />
        </label>
        <button type="submit" disabled={login.isPending}>
          <LogIn size={16} />
          {login.isPending ? "Signing in" : "Sign in"}
        </button>
        {login.isError ? <p className="error">Login failed</p> : null}
      </form>
    </main>
  );
}

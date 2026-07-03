export function LoginPage() {
  return (
    <main className="shell narrow">
      <h1>OPL Console</h1>
      <form className="panel">
        <label>
          Email
          <input type="email" name="email" autoComplete="email" />
        </label>
        <label>
          Password
          <input type="password" name="password" autoComplete="current-password" />
        </label>
        <button type="submit">Sign in</button>
      </form>
    </main>
  );
}

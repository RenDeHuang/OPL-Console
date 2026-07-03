import { useQuery } from "@tanstack/react-query";
import { api } from "../api/client";

export function OwnerOverviewPage() {
  const readiness = useQuery({ queryKey: ["runtime-readiness"], queryFn: api.runtimeReadiness });

  return (
    <main className="shell">
      <h1>Workspaces</h1>
      <section className="panel">
        <h2>Runtime</h2>
        <p>{readiness.data?.ready ? "Ready" : "Not ready"}</p>
      </section>
    </main>
  );
}

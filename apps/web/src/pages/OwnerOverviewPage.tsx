import { useQuery } from "@tanstack/react-query";
import { api } from "../api/client";

export function OwnerOverviewPage() {
  const readiness = useQuery({ queryKey: ["runtime-readiness"], queryFn: api.runtimeReadiness });
  let readinessStatus = "Not ready";

  if (readiness.isLoading) {
    readinessStatus = "Checking";
  } else if (readiness.isError) {
    readinessStatus = "Unavailable";
  } else if (readiness.data?.ready) {
    readinessStatus = "Ready";
  }

  return (
    <main className="shell">
      <h1>Workspaces</h1>
      <section className="panel">
        <h2>Runtime</h2>
        <p>{readinessStatus}</p>
      </section>
    </main>
  );
}

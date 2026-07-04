import { BrowserRouter } from "react-router";
import { AppRouter } from "./pages/ConsoleShell";

export function App() {
  return (
    <BrowserRouter>
      <AppRouter />
    </BrowserRouter>
  );
}

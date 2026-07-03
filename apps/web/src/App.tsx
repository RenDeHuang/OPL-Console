import { BrowserRouter, Route, Routes } from "react-router";

export function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<main className="shell"><h1>OPL Console</h1></main>} />
      </Routes>
    </BrowserRouter>
  );
}

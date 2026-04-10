import ReactDOM from "react-dom/client";
import { ThemeProvider } from "@/theme";
import { Toastify } from "@/components";
import App from "./App";
import "./styles/index.css";

const rootEl = document.getElementById("root");
if (rootEl) {
  const root = ReactDOM.createRoot(rootEl);
  root.render(
    <ThemeProvider>
      <Toastify />
      <App />
    </ThemeProvider>,
  );
}

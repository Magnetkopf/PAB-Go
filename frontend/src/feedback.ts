import { snackbar } from "mdui/functions/snackbar.js";

const htmlEntities: Record<string, string> = {
  "&": "&amp;",
  "<": "&lt;",
  ">": "&gt;",
  '"': "&quot;",
  "'": "&#039;",
};

// MDUI's helper assigns `message` with innerHTML, so escape API-provided errors first.
const escapeHtml = (message: string) => message.replace(/[&<>"']/g, (character) => htmlEntities[character]);

export function showSnackbar(message: string) {
  snackbar({
    message: escapeHtml(message),
    closeable: true,
    messageLine: 2,
    queue: "pab-feedback",
  });
}

export function showRequestError(error: unknown, fallback: string) {
  showSnackbar(error instanceof Error && error.message ? error.message : fallback);
}

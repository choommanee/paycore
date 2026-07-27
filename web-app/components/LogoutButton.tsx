"use client";

import { useState } from "react";

// LogoutButton clears the session, then sends the merchant back to the public
// landing. Posting a plain <form> to /api/auth/logout would navigate the browser
// to the endpoint's JSON response instead, so do it via fetch + redirect.
export default function LogoutButton() {
  const [busy, setBusy] = useState(false);

  async function logout() {
    setBusy(true);
    try {
      await fetch("/api/auth/logout", { method: "POST" });
    } finally {
      window.location.href = "/";
    }
  }

  return (
    <button
      onClick={logout}
      disabled={busy}
      className="text-sm text-paycore-muted hover:text-paycore-text disabled:opacity-60"
    >
      {busy ? "กำลังออก…" : "ออกจากระบบ"}
    </button>
  );
}

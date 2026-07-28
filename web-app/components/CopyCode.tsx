"use client";

import { useState } from "react";

// A code block with a copy-to-clipboard button. Client component because it needs
// navigator.clipboard + local "copied" state. `label` is an optional caption shown
// above the block (e.g. "cURL", "Node.js"); `code` is the raw text to render/copy.
export default function CopyCode({ code, label }: { code: string; label?: string }) {
  const [copied, setCopied] = useState(false);

  async function copy() {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      // Clipboard blocked (insecure context / permissions): fail quietly, the
      // code is still selectable by hand.
    }
  }

  return (
    <div className="overflow-hidden rounded-xl2 border border-paycore-line bg-paycore-surface2">
      <div className="flex items-center justify-between border-b border-paycore-line px-4 py-2">
        <span className="text-[11px] font-semibold uppercase tracking-wider text-paycore-muted">
          {label ?? "โค้ด"}
        </span>
        <button
          type="button"
          onClick={copy}
          className="rounded-md px-2 py-1 text-[11px] font-semibold text-paycore-accent transition-colors hover:bg-paycore-line2"
        >
          {copied ? "คัดลอกแล้ว ✓" : "คัดลอก"}
        </button>
      </div>
      <pre className="overflow-x-auto px-4 py-3 text-[12.5px] leading-relaxed">
        <code className="font-mono text-paycore-text">{code}</code>
      </pre>
    </div>
  );
}

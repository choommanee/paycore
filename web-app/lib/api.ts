import { cookies } from "next/headers";

// serverGet calls the Go backend directly from the Next server, forwarding the
// incoming request cookies so the pc_session cookie reaches the API. It does NOT
// derive the target from the inbound Host header (which is attacker-controllable
// and would be an SSRF vector) — it uses the trusted BACKEND_URL env var.
const BACKEND_URL = process.env.BACKEND_URL || "http://localhost:8080";

export async function serverGet(path: string): Promise<Response> {
  const cookieHeader = cookies().toString();
  return fetch(`${BACKEND_URL}/v1${path}`, {
    headers: { cookie: cookieHeader },
    cache: "no-store",
  });
}

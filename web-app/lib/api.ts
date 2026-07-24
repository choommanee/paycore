import { cookies, headers } from "next/headers";

// serverGet calls the backend through the same-origin /api proxy, forwarding the
// incoming request cookies so the pc_session cookie reaches the Go API.
export async function serverGet(path: string): Promise<Response> {
  const h = headers();
  const host = h.get("host");
  const proto = h.get("x-forwarded-proto") ?? "http";
  const cookieHeader = cookies().toString();
  return fetch(`${proto}://${host}/api${path}`, {
    headers: { cookie: cookieHeader },
    cache: "no-store",
  });
}

import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

// First-line UX guard: bounce to /login when the session cookie is absent.
// This is NOT the authorization boundary — the Go backend's merchantAuth is.
// Cookie presence is not proof of a valid session, so pages still handle a 401
// from serverGet by redirecting to /login.
export function middleware(req: NextRequest) {
  const hasSession = req.cookies.has("pc_session");
  if (!hasSession) {
    const url = req.nextUrl.clone();
    url.pathname = "/login";
    return NextResponse.redirect(url);
  }
  return NextResponse.next();
}

// Guard only the dashboard pages. /login, /pay/*, /api/*, and Next internals are
// excluded so public checkout and the login page stay reachable without a cookie.
export const config = {
  matcher: ["/", "/transactions/:path*", "/settings/:path*", "/links/:path*"],
};

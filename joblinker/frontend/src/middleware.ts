import { NextResponse } from 'next/server';
import type { NextRequest } from 'next/server';

const PUBLIC_PATHS = [
  '/login',
  '/register',
  '/privacy',
  '/api/auth/login',
  '/api/auth/register',
  '/api/chat',
  '/',
];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Allow public paths
  if (PUBLIC_PATHS.some(path => pathname === path || pathname.startsWith('/api/auth'))) {
    return NextResponse.next();
  }

  // Check for auth session cookie OR Authorization header
  const authCookie = request.cookies.get('joblinker-auth');
  const authHeader = request.headers.get('Authorization');
  const hasBearerToken = authHeader?.startsWith('Bearer ');

  if (!authCookie?.value && !hasBearerToken) {
    // Redirect to login for protected pages
    const loginUrl = new URL('/login', request.url);
    loginUrl.searchParams.set('redirect', pathname);
    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
}

export const config = {
  matcher: [
    /*
     * Match all request paths except:
     * - _next/static (static files)
     * - _next/image (image optimization)
     * - favicon.ico (favicon)
     * - public files
     */
    '/((?!_next/static|_next/image|favicon.ico|.*\\.(?:svg|png|jpg|jpeg|gif|webp)$).*)',
  ],
};
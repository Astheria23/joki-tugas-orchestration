import { useState } from 'react';
import { Link, Navigate, useLocation, useNavigate } from 'react-router-dom';
import { AlertTriangle, ArrowRight } from 'lucide-react';
import { BrandLogo } from '../components/BrandLogo';
import { useAuth } from '../hooks/useAuth';

type Mode = 'login' | 'register';

export function AuthPage({ mode }: { mode: Mode }) {
  const { login, register, isAuthenticated } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const from = (location.state as { from?: string } | null)?.from || '/app';

  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  if (isAuthenticated) {
    return <Navigate to={from} replace />;
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!username.trim() || !password.trim()) {
      setError('Username dan password wajib diisi.');
      return;
    }
    if (password.length < 6) {
      setError('Password minimal 6 karakter.');
      return;
    }

    setSubmitting(true);
    try {
      if (mode === 'login') {
        await login(username.trim(), password);
      } else {
        await register(username.trim(), password);
      }
      navigate(from, { replace: true });
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Gagal masuk. Coba lagi ya.');
    } finally {
      setSubmitting(false);
    }
  };

  const isLogin = mode === 'login';

  return (
    <div className="relative min-h-svh bg-canvas text-ink flex flex-col overflow-hidden">
      <div className="absolute inset-0 hero-plane" aria-hidden />
      <div className="absolute inset-0 hero-grain" aria-hidden />

      <header className="relative z-20 px-5 sm:px-6 py-5">
        <div className="max-w-6xl mx-auto">
          <BrandLogo size="sm" variant="wordmark" />
        </div>
      </header>

      <main className="relative z-10 flex-1 flex items-center justify-center px-5 sm:px-6 py-10">
        <div className="w-full max-w-md animate-rise">
          <div className="rounded-3xl bg-white border border-ink/8 shadow-[0_24px_60px_rgba(15,20,25,0.08)] p-7 sm:p-9">
            <h1 className="font-display text-2xl sm:text-3xl font-bold tracking-tight text-ink mb-2 text-center">
              {isLogin ? 'Masuk dulu' : 'Buat akun gratis'}
            </h1>
            <p className="text-ink/50 mb-7 text-sm leading-relaxed text-center">
              {isLogin
                ? 'Lanjut beresin tugas yang kemarin sempat tertunda.'
                : 'Daftar sekali — setelah itu langsung bisa minta bantuan ngerjain tugas.'}
            </p>

            {error && (
              <div className="mb-5 flex items-start gap-2 rounded-2xl border border-rose-300 bg-rose-50 text-rose-800 px-4 py-3 text-sm">
                <AlertTriangle className="h-4 w-4 shrink-0 mt-0.5" />
                <span>{error}</span>
              </div>
            )}

            <form onSubmit={(e) => void handleSubmit(e)} className="space-y-4">
              <div>
                <label
                  htmlFor="username"
                  className="block text-xs font-semibold uppercase tracking-wider text-ink/45 mb-1.5"
                >
                  Username
                </label>
                <input
                  id="username"
                  type="text"
                  autoComplete="username"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  className="w-full rounded-2xl border border-ink/10 bg-canvas px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-banana focus:bg-white transition"
                  placeholder="nama_kamu"
                  required
                />
              </div>
              <div>
                <label
                  htmlFor="password"
                  className="block text-xs font-semibold uppercase tracking-wider text-ink/45 mb-1.5"
                >
                  Password
                </label>
                <input
                  id="password"
                  type="password"
                  autoComplete={isLogin ? 'current-password' : 'new-password'}
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  className="w-full rounded-2xl border border-ink/10 bg-canvas px-4 py-3 text-sm focus:outline-none focus:ring-2 focus:ring-banana focus:bg-white transition"
                  placeholder="Minimal 6 karakter"
                  required
                  minLength={6}
                />
              </div>

              <button
                type="submit"
                disabled={submitting}
                className="w-full inline-flex items-center justify-center gap-2 rounded-full bg-banana text-ink font-bold py-3.5 text-sm hover:bg-banana-deep disabled:opacity-50 transition-colors shadow-[0_8px_24px_rgba(245,197,24,0.35)] mt-2"
              >
                {submitting ? 'Memproses…' : isLogin ? 'Masuk' : 'Daftar'}
                {!submitting && <ArrowRight className="h-4 w-4" />}
              </button>
            </form>

            <p className="mt-6 text-sm text-ink/50 text-center">
              {isLogin ? (
                <>
                  Belum punya akun?{' '}
                  <Link
                    to="/register"
                    className="font-semibold text-navy underline underline-offset-2 hover:text-ink"
                  >
                    Daftar
                  </Link>
                </>
              ) : (
                <>
                  Sudah punya akun?{' '}
                  <Link
                    to="/login"
                    className="font-semibold text-navy underline underline-offset-2 hover:text-ink"
                  >
                    Masuk
                  </Link>
                </>
              )}
            </p>
          </div>
        </div>
      </main>
    </div>
  );
}

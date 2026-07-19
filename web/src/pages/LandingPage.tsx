import { Link } from 'react-router-dom';
import { ArrowRight, Check } from 'lucide-react';
import { BrandLogo } from '../components/BrandLogo';
import { useAuth } from '../hooks/useAuth';

const NAV = [
  { href: '#apa-itu', label: 'Apa itu' },
  { href: '#cara-kerja', label: 'Cara kerja' },
  { href: '#harga', label: 'Harga' },
];

const PLANS = [
  {
    id: 'coba',
    name: 'Coba dulu',
    price: 'Gratis',
    period: '',
    blurb: 'Pas buat ngetes apakah cocok buat ritme kuliahmu.',
    features: ['5 chat per akun', 'Hasil teks & file', 'Riwayat chat tersimpan'],
    cta: 'Mulai gratis',
    highlighted: false,
  },
  {
    id: 'mahasiswa',
    name: 'Mahasiswa',
    price: 'Rp49rb',
    period: '/bulan',
    blurb: 'Untuk yang sering numpuk deadline di satu minggu.',
    features: [
      'Kuota chat lebih besar (menyusul)',
      'Prioritas pengerjaan',
      'Unduh file tanpa batas',
      'Riwayat lengkap',
    ],
    cta: 'Pilih paket ini',
    highlighted: true,
  },
  {
    id: 'semester',
    name: 'Semester',
    price: 'Rp199rb',
    period: '/4 bulan',
    blurb: 'Hemat buat satu semester penuh biar tidak bayar tiap bulan.',
    features: [
      'Kuota chat paling longgar (menyusul)',
      'Prioritas tertinggi',
      'Dukungan chat prioritas',
      'Cocok buat skripsi & proyek besar',
    ],
    cta: 'Ambil paket semester',
    highlighted: false,
  },
];

export function LandingPage() {
  const { isAuthenticated } = useAuth();
  const startTo = isAuthenticated ? '/app' : '/register';

  return (
    <div className="min-h-screen flex flex-col bg-canvas text-ink overflow-x-hidden">
      <header className="relative z-30 px-5 sm:px-6 py-4">
        <div className="max-w-6xl mx-auto flex items-center justify-between gap-4">
          <BrandLogo size="sm" variant="wordmark" />

          <nav className="hidden md:flex items-center gap-8" aria-label="Navigasi utama">
            {NAV.map((item) => (
              <a
                key={item.href}
                href={item.href}
                className="text-sm font-medium text-ink/55 hover:text-ink transition-colors"
              >
                {item.label}
              </a>
            ))}
          </nav>

          <div className="flex items-center gap-2 sm:gap-3">
            {isAuthenticated ? (
              <Link
                to="/app"
                className="rounded-full bg-navy text-white px-4 py-2 text-sm font-semibold hover:bg-ink transition-colors"
              >
                Lanjut ngerjain
              </Link>
            ) : (
              <>
                <Link
                  to="/login"
                  className="hidden sm:inline text-sm font-medium text-ink/60 hover:text-ink transition-colors px-2"
                >
                  Masuk
                </Link>
                <Link
                  to="/register"
                  className="rounded-full bg-navy text-white px-4 py-2 text-sm font-semibold hover:bg-ink transition-colors"
                >
                  Daftar
                </Link>
              </>
            )}
          </div>
        </div>
      </header>

      {/* Hero — brand first, one composition, no clutter */}
      <section className="relative flex flex-col justify-center min-h-[calc(100svh-4.5rem)]">
        <div className="absolute inset-0 hero-plane" aria-hidden />
        <div className="absolute inset-0 hero-grain" aria-hidden />

        <div className="relative z-10 max-w-6xl mx-auto w-full px-5 sm:px-6 py-16 md:py-20">
          <div className="max-w-2xl">
            <div className="mb-3 md:mb-4 leading-none">
              <BrandLogo size="hero" to={null} />
            </div>

            <h1 className="font-display text-4xl sm:text-5xl md:text-[3.5rem] font-bold tracking-tight leading-[1.08] text-ink mb-5 animate-rise">
              Deadline numpuk?
              <span className="block mt-1">Serahin aja ke kami.</span>
            </h1>

            <p className="text-lg md:text-xl text-ink/60 max-w-md leading-relaxed mb-9 animate-rise-delay">
              Ceritain tugas kuliahmu. Bananacademic yang bantu kerjakan biar kamu bisa napas lagi.
            </p>

            <div className="flex flex-wrap items-center gap-3 animate-rise-delay">
              <Link
                to={startTo}
                className="inline-flex items-center gap-2 rounded-full bg-banana text-ink px-6 py-3.5 text-sm font-bold hover:bg-banana-deep transition-all shadow-[0_12px_32px_rgba(245,197,24,0.4)] hover:-translate-y-0.5"
              >
                {isAuthenticated ? 'Buka aplikasi' : 'Coba gratis'}
                <ArrowRight className="h-4 w-4" />
              </Link>
              <a
                href="#cara-kerja"
                className="inline-flex items-center rounded-full border border-ink/12 bg-white/55 backdrop-blur px-6 py-3.5 text-sm font-semibold text-ink/80 hover:bg-white transition-colors"
              >
                Lihat cara kerjanya
              </a>
            </div>
          </div>
        </div>
      </section>

      {/* What is Bananacademic */}
      <section id="apa-itu" className="relative z-10 bg-white border-t border-ink/6 scroll-mt-20">
        <div className="max-w-6xl mx-auto px-5 sm:px-6 py-20 md:py-28">
          <p className="text-xs font-semibold uppercase tracking-[0.2em] text-navy/50 mb-4">
            Apa itu Bananacademic
          </p>
          <h2 className="font-display text-3xl md:text-5xl font-bold tracking-tight text-ink max-w-3xl leading-[1.12] mb-6">
            Teman ngerjain tugas kuliah tanpa drama, tanpa begadang sendirian.
          </h2>
          <p className="text-base md:text-lg text-ink/55 max-w-2xl leading-relaxed mb-12">
            Kamu tulis aja mau apa: ringkas materi, bikin slide, rapikan tulisan, terjemahin, atau
            susun laporan. Kami olah permintaannya sampai hasilnya siap kamu pakai buat dikumpulin.
          </p>

          <ul className="grid sm:grid-cols-3 gap-8 md:gap-12 max-w-4xl">
            {[
              {
                title: 'Bahasa manusia',
                body: 'Nggak perlu format aneh. Tulis kayak chat ke temen: “tolong ringkas bab 3 terus jadiin PPT”.',
              },
              {
                title: 'Hasil siap pakai',
                body: 'Dapat teks yang bisa langsung disalin, atau file yang bisa langsung diunduh.',
              },
              {
                title: 'Riwayat tersimpan',
                body: 'Tugas yang sudah selesai tetap ada di akunmu tinggal buka lagi kalau perlu.',
              },
            ].map((item) => (
              <li key={item.title}>
                <h3 className="font-display text-xl font-bold text-ink mb-2">{item.title}</h3>
                <p className="text-sm text-ink/55 leading-relaxed">{item.body}</p>
              </li>
            ))}
          </ul>
        </div>
      </section>

      {/* How it works */}
      <section id="cara-kerja" className="relative z-10 bg-canvas border-t border-ink/6 scroll-mt-20">
        <div className="max-w-6xl mx-auto px-5 sm:px-6 py-20 md:py-28">
          <p className="text-xs font-semibold uppercase tracking-[0.2em] text-navy/50 mb-4">
            Cara kerja
          </p>
          <h2 className="font-display text-3xl md:text-5xl font-bold tracking-tight text-ink max-w-2xl leading-[1.12] mb-4">
            Tiga langkah. Selesai.
          </h2>
          <p className="text-ink/55 max-w-xl mb-14 text-base md:text-lg leading-relaxed">
            Simpel banget dari cerita tugas sampai file di tangan.
          </p>

          <ol className="grid md:grid-cols-3 gap-10 md:gap-8">
            {[
              {
                n: '1',
                title: 'Ceritain tugasmu',
                body: 'Masuk, tulis apa yang dibutuhkan. Bisa sekalian kasih link atau detail singkat.',
              },
              {
                n: '2',
                title: 'Pantau sampai beres',
                body: 'Kamu bisa lihat progresnya jalan. Nggak perlu refresh berkali-kali statusnya update sendiri.',
              },
              {
                n: '3',
                title: 'Ambil hasilnya',
                body: 'Salin teksnya, atau unduh filenya. Simpan di riwayat buat jaga-jaga.',
              },
            ].map((step) => (
              <li key={step.n} className="relative">
                <span className="font-display text-6xl md:text-7xl font-bold text-banana leading-none select-none">
                  {step.n}
                </span>
                <h3 className="mt-4 text-xl font-bold text-ink">{step.title}</h3>
                <p className="mt-2 text-sm text-ink/55 leading-relaxed max-w-xs">{step.body}</p>
              </li>
            ))}
          </ol>
        </div>
      </section>

      {/* Pricing */}
      <section id="harga" className="relative z-10 bg-white border-t border-ink/6 scroll-mt-20">
        <div className="max-w-6xl mx-auto px-5 sm:px-6 py-20 md:py-28">
          <p className="text-xs font-semibold uppercase tracking-[0.2em] text-navy/50 mb-4">
            Harga
          </p>
          <h2 className="font-display text-3xl md:text-5xl font-bold tracking-tight text-ink max-w-2xl leading-[1.12] mb-4">
            Pilih yang pas sama kantong mahasiswa.
          </h2>
          <p className="text-ink/55 max-w-xl mb-14 text-base md:text-lg leading-relaxed">
            Mulai gratis. Naik paket kalau ternyata sering kepake.
          </p>

          <div className="grid md:grid-cols-3 gap-5 md:gap-6 items-stretch">
            {PLANS.map((plan) => (
              <div
                key={plan.id}
                className={`relative flex flex-col rounded-3xl p-6 md:p-7 transition-transform hover:-translate-y-1 ${
                  plan.highlighted
                    ? 'bg-navy text-white shadow-[0_20px_50px_rgba(27,42,74,0.28)] ring-2 ring-banana'
                    : 'bg-canvas border border-ink/8 text-ink'
                }`}
              >
                {plan.highlighted && (
                  <span className="absolute -top-3 left-6 rounded-full bg-banana text-ink text-[11px] font-bold uppercase tracking-wider px-3 py-1">
                    Paling dipilih
                  </span>
                )}

                <h3
                  className={`font-display text-2xl font-bold ${
                    plan.highlighted ? 'text-white' : 'text-ink'
                  }`}
                >
                  {plan.name}
                </h3>
                <p
                  className={`mt-2 text-sm leading-relaxed ${
                    plan.highlighted ? 'text-white/65' : 'text-ink/50'
                  }`}
                >
                  {plan.blurb}
                </p>

                <div className="mt-6 mb-6 flex items-baseline gap-1">
                  <span className="font-display text-4xl font-bold tracking-tight">{plan.price}</span>
                  {plan.period && (
                    <span
                      className={`text-sm ${plan.highlighted ? 'text-white/50' : 'text-ink/40'}`}
                    >
                      {plan.period}
                    </span>
                  )}
                </div>

                <ul className="space-y-3 mb-8 flex-1">
                  {plan.features.map((f) => (
                    <li key={f} className="flex items-start gap-2.5 text-sm">
                      <Check
                        className={`h-4 w-4 shrink-0 mt-0.5 ${
                          plan.highlighted ? 'text-banana' : 'text-leaf'
                        }`}
                      />
                      <span className={plan.highlighted ? 'text-white/85' : 'text-ink/70'}>{f}</span>
                    </li>
                  ))}
                </ul>

                <Link
                  to={startTo}
                  className={`inline-flex items-center justify-center gap-2 rounded-full py-3 text-sm font-bold transition-colors ${
                    plan.highlighted
                      ? 'bg-banana text-ink hover:bg-banana-deep'
                      : 'bg-navy text-white hover:bg-ink'
                  }`}
                >
                  {plan.cta}
                  <ArrowRight className="h-4 w-4" />
                </Link>
              </div>
            ))}
          </div>

          <p className="mt-8 text-center text-xs text-ink/40">
            Harga bisa berubah. Pembayaran paket berbayar menyusul — untuk sekarang kamu bisa mulai dari
            paket gratis.
          </p>
        </div>
      </section>

      {/* Closing CTA */}
      <section className="relative z-10 border-t border-ink/6 overflow-hidden">
        <div className="absolute inset-0 hero-plane opacity-80" aria-hidden />
        <div className="relative max-w-6xl mx-auto px-5 sm:px-6 py-20 md:py-24 text-center">
          <h2 className="font-display text-3xl md:text-5xl font-bold tracking-tight text-ink mb-4">
            Udah siap beresin tugas?
          </h2>
          <p className="text-ink/55 max-w-md mx-auto mb-8 leading-relaxed">
            Daftar gratis, tulis tugas pertama, dan ambil hasilnya hari ini juga.
          </p>
          <Link
            to={startTo}
            className="inline-flex items-center gap-2 rounded-full bg-banana text-ink px-8 py-3.5 text-sm font-bold hover:bg-banana-deep transition-all shadow-[0_12px_32px_rgba(245,197,24,0.4)] hover:-translate-y-0.5"
          >
            {isAuthenticated ? 'Masuk ke aplikasi' : 'Mulai sekarang'}
            <ArrowRight className="h-4 w-4" />
          </Link>
        </div>
      </section>

      <footer className="relative z-10 border-t border-ink/8 bg-white px-5 sm:px-6 py-10">
        <div className="max-w-6xl mx-auto flex flex-col sm:flex-row items-center justify-between gap-6">
          <BrandLogo size="sm" variant="wordmark" />
          <nav className="flex flex-wrap justify-center gap-6 text-sm text-ink/45">
            {NAV.map((item) => (
              <a key={item.href} href={item.href} className="hover:text-ink transition-colors">
                {item.label}
              </a>
            ))}
            <Link to="/login" className="hover:text-ink transition-colors">
              Masuk
            </Link>
          </nav>
          <p className="text-xs text-ink/35">
            © {new Date().getFullYear()} Bananacademic
          </p>
        </div>
      </footer>
    </div>
  );
}

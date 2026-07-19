import { Link } from 'react-router-dom';

interface BrandLogoProps {
  size?: 'sm' | 'md' | 'lg' | 'hero';
  /** Text wordmark (navbar) instead of image mark */
  variant?: 'mark' | 'wordmark';
  to?: string | null;
  className?: string;
}

const markSize = {
  sm: 'h-10 w-auto',
  md: 'h-12 w-auto',
  lg: 'h-20 w-auto',
  hero: 'h-36 sm:h-40 md:h-44 w-auto max-w-full',
} as const;

const wordSize = {
  sm: 'text-xl',
  md: 'text-2xl',
  lg: 'text-3xl',
  hero: 'text-4xl md:text-5xl',
} as const;

function Wordmark({ size }: { size: keyof typeof wordSize }) {
  return (
    <span
      className={`font-display font-extrabold tracking-tight leading-none ${wordSize[size]}`}
    >
      <span className="text-banana">Banana</span>
      <span className="text-navy">academic</span>
    </span>
  );
}

export function BrandLogo({
  size = 'md',
  variant = 'mark',
  to = '/',
  className = '',
}: BrandLogoProps) {
  const content =
    variant === 'wordmark' ? (
      <span className={`inline-flex items-center ${className}`}>
        <Wordmark size={size} />
      </span>
    ) : (
      <span className={`inline-flex items-center ${className}`}>
        <img
          src="/bananacademic.png"
          alt="Bananacademic"
          className={`${markSize[size]} shrink-0 object-contain`}
        />
      </span>
    );

  if (to === null) return content;
  return (
    <Link
      to={to}
      className="inline-flex focus:outline-none focus-visible:ring-2 focus-visible:ring-banana rounded-sm"
      aria-label="Bananacademic beranda"
    >
      {content}
    </Link>
  );
}

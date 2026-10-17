import { useParams } from 'react-router-dom';
import { useSelector } from 'react-redux';
import { useGetRegionsQuery } from '../store/api.js';

export const REGION_COOKIE = 'sb_region';
export const DEFAULT_REGION = 'mumbai';

export function writeRegionCookie(slug) {
  if (typeof document !== 'undefined') {
    document.cookie = `${REGION_COOKIE}=${slug}; path=/; max-age=${60 * 60 * 24 * 365}; samesite=lax`;
  }
}

export function parseCookies(header = '') {
  return Object.fromEntries(
    header
      .split(';')
      .map((c) => c.trim().split('='))
      .filter(([k]) => k)
      .map(([k, ...v]) => [k, decodeURIComponent(v.join('='))]),
  );
}

// Region slug for the current page: the URL's :city segment wins, then the
// cookie-backed store value, then the default city.
export function useRegionSlug() {
  const params = useParams();
  const stored = useSelector((s) => s.ui.regionSlug);
  const fromListing = params.listing?.match(/^(?:movies|upcoming-movies|events|plays|sports|activities)-(.+)$/)?.[1];
  return params.city || fromListing || stored || DEFAULT_REGION;
}

export function useRegion() {
  const slug = useRegionSlug();
  const { data } = useGetRegionsQuery();
  const all = data ? [...data.popular, ...data.others] : [];
  return all.find((r) => r.slug === slug) || all[0] || { slug, name: slug, code: slug.toUpperCase() };
}

// Replace the city segment of a path when the user switches cities.
export function pathForCity(pathname, slug) {
  const swaps = [
    [/^\/explore\/home\/[^/]+/, `/explore/home/${slug}`],
    [/^\/explore\/(movies|upcoming-movies|events|plays|sports|activities)-[^/?]+/, `/explore/$1-${slug}`],
    [/^\/movies\/[^/]+\/([^/]+\/[A-Z0-9]+)$/, `/movies/${slug}/$1`],
    [/^\/[^/]+\/cinemas$/, `/${slug}/cinemas`],
  ];
  for (const [re, to] of swaps) if (re.test(pathname)) return pathname.replace(re, to);
  return `/explore/home/${slug}`;
}

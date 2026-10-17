const I = ({ children, size = 20, ...p }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" {...p}>
    {children}
  </svg>
);

export const SearchIcon = (p) => (<I {...p}><circle cx="11" cy="11" r="7" /><path d="m20 20-3.5-3.5" /></I>);
export const ChevronDown = (p) => (<I {...p}><path d="m6 9 6 6 6-6" /></I>);
export const ChevronUp = (p) => (<I {...p}><path d="m18 15-6-6-6 6" /></I>);
export const ChevronLeft = (p) => (<I {...p}><path d="m15 18-6-6 6-6" /></I>);
export const ChevronRight = (p) => (<I {...p}><path d="m9 18 6-6-6-6" /></I>);
export const MenuIcon = (p) => (<I {...p}><path d="M4 7h16M4 12h16M4 17h16" /></I>);
export const CloseIcon = (p) => (<I {...p}><path d="M6 6l12 12M18 6 6 18" /></I>);
export const PinIcon = (p) => (<I {...p}><path d="M12 21s-7-6.2-7-12a7 7 0 1 1 14 0c0 5.8-7 12-7 12z" /><circle cx="12" cy="9" r="2.5" /></I>);
export const HeartIcon = (p) => (<I {...p}><path d="M20.8 5.6a5 5 0 0 0-7.1 0L12 7.3l-1.7-1.7a5 5 0 0 0-7.1 7.1L12 21.5l8.8-8.8a5 5 0 0 0 0-7.1z" /></I>);
export const InfoIcon = (p) => (<I {...p}><circle cx="12" cy="12" r="9" /><path d="M12 11v5M12 8h.01" /></I>);
export const ShareIcon = (p) => (<I {...p}><circle cx="18" cy="5" r="3" /><circle cx="6" cy="12" r="3" /><circle cx="18" cy="19" r="3" /><path d="m8.6 13.5 6.8 4M15.4 6.5l-6.8 4" /></I>);
export const PlayIcon = (p) => (<I {...p} fill="currentColor" stroke="none"><path d="M8 5v14l11-7z" /></I>);
export const EditIcon = (p) => (<I {...p}><path d="M12 20h9M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4z" /></I>);
export const CheckIcon = (p) => (<I {...p}><path d="M20 6 9 17l-5-5" /></I>);
export const BellIcon = (p) => (<I {...p}><path d="M18 8a6 6 0 1 0-12 0c0 7-3 9-3 9h18s-3-2-3-9M13.7 21a2 2 0 0 1-3.4 0" /></I>);
export const BagIcon = (p) => (<I {...p}><path d="M6 2 3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4zM3 6h18M16 10a4 4 0 0 1-8 0" /></I>);
export const TvIcon = (p) => (<I {...p}><rect x="2" y="7" width="20" height="15" rx="2" /><path d="m17 2-5 5-5-5" /></I>);
export const CardIcon = (p) => (<I {...p}><rect x="2" y="5" width="20" height="14" rx="2" /><path d="M2 10h20" /></I>);
export const HelpIcon = (p) => (<I {...p}><circle cx="12" cy="12" r="9" /><path d="M9.1 9a3 3 0 0 1 5.8 1c0 2-3 3-3 3M12 17h.01" /></I>);
export const GearIcon = (p) => (<I {...p}><circle cx="12" cy="12" r="3" /><path d="M19.4 15a1.7 1.7 0 0 0 .3 1.8l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-1.8-.3 1.7 1.7 0 0 0-1 1.5V21a2 2 0 1 1-4 0v-.1a1.7 1.7 0 0 0-1.1-1.5 1.7 1.7 0 0 0-1.8.3l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.7 1.7 0 0 0 .3-1.8 1.7 1.7 0 0 0-1.5-1H3a2 2 0 1 1 0-4h.1a1.7 1.7 0 0 0 1.5-1.1 1.7 1.7 0 0 0-.3-1.8l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.7 1.7 0 0 0 1.8.3H9a1.7 1.7 0 0 0 1-1.5V3a2 2 0 1 1 4 0v.1a1.7 1.7 0 0 0 1 1.5 1.7 1.7 0 0 0 1.8-.3l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0-.3 1.8V9a1.7 1.7 0 0 0 1.5 1H21a2 2 0 1 1 0 4h-.1a1.7 1.7 0 0 0-1.5 1z" /></I>);
export const GiftIcon = (p) => (<I {...p}><path d="M20 12v10H4V12M2 7h20v5H2zM12 22V7M12 7H7.5a2.5 2.5 0 0 1 0-5C11 2 12 7 12 7zM12 7h4.5a2.5 2.5 0 0 0 0-5C13 2 12 7 12 7z" /></I>);
export const SmileIcon = (p) => (<I {...p}><circle cx="12" cy="12" r="9" /><path d="M8 14s1.5 2 4 2 4-2 4-2M9 9h.01M15 9h.01" /></I>);
export const UserIcon = (p) => (<I {...p}><circle cx="12" cy="8" r="4" /><path d="M4 21a8 8 0 0 1 16 0" /></I>);
export const LogoutIcon = (p) => (<I {...p}><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9" /></I>);
export const MoonIcon = (p) => (<I {...p}><path d="M21 12.8A9 9 0 1 1 11.2 3a7 7 0 0 0 9.8 9.8z" /></I>);
export const SunIcon = (p) => (<I {...p}><circle cx="12" cy="12" r="4" /><path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4" /></I>);
export const TicketIcon = (p) => (<I {...p}><path d="M3 8a2 2 0 0 0 0 4v4a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-4a2 2 0 0 1 0-4V6a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2z" /><path d="M13 4v2M13 10v2M13 16v2" /></I>);
export const FoodIcon = (p) => (<I {...p}><path d="M3 11h18l-2 10H5zM7 11V7a5 5 0 0 1 10 0v4" /></I>);
export const ClockIcon = (p) => (<I {...p}><circle cx="12" cy="12" r="9" /><path d="M12 7v5l3 2" /></I>);
export const ShieldIcon = (p) => (<I {...p}><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" /></I>);

export const StarFilled = ({ size = 18, color = '#F84464' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" aria-hidden="true">
    <path fill={color} d="M12 2l2.9 6.9 7.1.6-5.4 4.7 1.7 7-6.3-3.9-6.3 3.9 1.7-7L2 9.5l7.1-.6z" />
  </svg>
);

export const ThumbUp = ({ size = 18, color = '#1EA83C' }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" aria-hidden="true">
    <path fill={color} d="M2 21h4V9H2v12zm20-11c0-1.1-.9-2-2-2h-6.3l1-4.6v-.3c0-.4-.2-.8-.4-1.1L13.2 1 6.6 7.6C6.2 7.9 6 8.4 6 9v10c0 1.1.9 2 2 2h9c.8 0 1.5-.5 1.8-1.2l3-7.1c.1-.2.2-.5.2-.7v-2z" />
  </svg>
);

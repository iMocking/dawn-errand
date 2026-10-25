/**
 * 内联 SVG 线框图标注册表（stroke-width 2，viewBox 0 0 24 24）。
 * 一律 stroke="currentColor"，由外部 color 控制；禁止 emoji。
 * 后续新增图标在此登记即可。
 */
export const ICONS: Record<string, string> = {
  // 品牌 / 通用
  clock: '<circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>',
  search: '<circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/>',
  'chevron-right': '<polyline points="9 18 15 12 9 6"/>',
  'chevrons-left': '<polyline points="11 17 6 12 11 7"/><polyline points="18 17 13 12 18 7"/>',
  'chevrons-right': '<polyline points="13 17 18 12 13 7"/><polyline points="6 17 11 12 6 7"/>',
  logout:
    '<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/>',
  bell: '<path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.73 21a2 2 0 0 1-3.46 0"/>',
  download:
    '<path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/>',
  plus: '<line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>',
  'check-circle':
    '<path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/>',
  close: '<line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/>',

  // 菜单
  chart: '<path d="M3 3v18h18"/><path d="M7 14l4-4 3 3 5-6"/>',
  dashboard:
    '<rect x="3" y="3" width="7" height="9" rx="1"/><rect x="14" y="3" width="7" height="5" rx="1"/><rect x="14" y="12" width="7" height="9" rx="1"/><rect x="3" y="16" width="7" height="5" rx="1"/>',
  order:
    '<path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/>',
  clipboard:
    '<rect x="5" y="4" width="14" height="17" rx="2"/><path d="M9 4.5V3h6v1.5"/><path d="M8.5 9h7M8.5 13h7M8.5 17h4"/>',
  complaint:
    '<path d="M3 11h18M5 11V8a2 2 0 0 1 2-2h10a2 2 0 0 1 2 2v3M4 11l1 9h14l1-9"/>',
  headset:
    '<path d="M4 13a8 8 0 0 1 16 0"/><path d="M4 13v3a2 2 0 0 0 2 2h2v-6H6a2 2 0 0 0-2 2z"/><path d="M20 13v3a2 2 0 0 1-2 2h-2v-6h2a2 2 0 0 1 2 2z"/><path d="M12 18v2a2 2 0 0 0 2 2h1"/>',
  merchant:
    '<path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><path d="M9 22V12h6v10"/>',
  storefront:
    '<path d="M3 9l2-5h14l2 5"/><path d="M3 9h18"/><path d="M4 9v10h16V9"/><path d="M8 19v-5h5v5"/><path d="M15 12h3v3h-3z"/>',
  'storefront-check':
    '<path d="M3 9l2-5h14l2 5"/><path d="M3 9h18"/><path d="M4 9v10h16V9"/><path d="M8 19v-5h5v5"/><path d="M15 12h3v3h-3z"/><circle cx="19" cy="5" r="3"/><path d="M17.8 5l.8.8 1.6-1.6"/>',
  shield:
    '<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/><path d="M9 12l2 2 4-4"/>',
  grid: '<rect x="4" y="4" width="16" height="16" rx="2"/><path d="M9 4v16M15 4v16M4 9h16M4 15h16"/>',
  'rider-audit':
    '<path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2"/><circle cx="9" cy="7" r="4"/><polyline points="16 11 18 13 22 9"/>',
  'rider-scooter':
    '<circle cx="5.5" cy="19" r="2"/><circle cx="18.5" cy="19" r="2"/><path d="M7.5 19h8.5"/><path d="M8 16h5l2.5 3"/><path d="M15.5 10.5l1.5 5.5"/><path d="M15.5 10.5h2.5"/><circle cx="11" cy="4.5" r="1.6"/><path d="M11 6.1v4.4"/><path d="M11 7.5l4 1"/><path d="M11 10.5l3 5"/>',
  scooter:
    '<circle cx="5.5" cy="19" r="2"/><circle cx="18.5" cy="19" r="2"/><path d="M7.5 19h8.5"/><path d="M8 16h5l2.5 3"/><path d="M15.5 10.5l1.5 5.5"/><path d="M15.5 10.5h2.5"/><path d="M8 16l-1-3h4"/>',
  'scooter-coin':
    '<circle cx="6" cy="6" r="3"/><path d="M5 4.5l1 1.5 1-1.5"/><path d="M5 6.8h2"/><path d="M6 6v2.5"/><circle cx="8.5" cy="19" r="2"/><circle cx="19" cy="19" r="2"/><path d="M10.5 19h6.5"/><path d="M11 16h4l2.5 3"/><path d="M16.5 10.5l1.5 5.5"/><path d="M16.5 10.5h2.5"/>',
  user: '<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>',
  'id-card':
    '<rect x="3" y="5" width="18" height="14" rx="2"/><circle cx="8" cy="11" r="2"/><path d="M5.5 16a2.5 2.5 0 0 1 5 0"/><path d="M14 10h4M14 13h4M14 16h3"/>',
  money: '<line x1="12" y1="1" x2="12" y2="23"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/>',
  wallet:
    '<rect x="3" y="6" width="18" height="14" rx="2"/><path d="M3 10h18"/><path d="M15 14h4v3h-4a1.5 1.5 0 0 1 0-3z"/>',
  list: '<path d="M8 6h13M8 12h13M8 18h13"/><path d="M3 6h.01M3 12h.01M3 18h.01"/>',
  receipt:
    '<path d="M6 3h12v18l-3-2-3 2-3-2-3 2z"/><path d="M9 8h6M9 12h6M9 16h3"/>',
  withdraw:
    '<circle cx="12" cy="12" r="10"/><polyline points="16 12 12 8 8 12"/><line x1="12" y1="8" x2="12" y2="16"/>',
  settings:
    '<circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 1 1-4 0v-.09a1.65 1.65 0 0 0-1-1.51 1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 1 1 0-4h.09a1.65 1.65 0 0 0 1.51-1 1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33h.09a1.65 1.65 0 0 0 1-1.51V3a2 2 0 1 1 4 0v.09a1.65 1.65 0 0 0 1 1.51h.09a1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82v.09a1.65 1.65 0 0 0 1.51 1H21a2 2 0 1 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>',
  tag: '<path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z"/><line x1="7" y1="7" x2="7.01" y2="7"/>',
  megaphone:
    '<path d="M3 10v4a1 1 0 0 0 1 1h3l7 4V5L7 9H4a1 1 0 0 0-1 1z"/><path d="M18 8a4 4 0 0 1 0 8"/><path d="M7 15v4"/>',
  mail: '<path d="M4 4h16c1.1 0 2 .9 2 2v12c0 1.1-.9 2-2 2H4c-1.1 0-2-.9-2-2V6c0-1.1.9-2 2-2z"/><polyline points="22 6 12 13 2 6"/>',
  message:
    '<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>',

  // 登录页
  lock: '<rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>',
  eye: '<path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>',
  captcha: '<rect x="3" y="3" width="18" height="18" rx="2"/><path d="M8 12l2.5 2.5L16 8"/>',
  checkbox: '<polyline points="20 6 9 17 4 12"/>',
}

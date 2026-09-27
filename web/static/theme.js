/* Shared Sir John Shell theme format and runtime helpers. */
const SJS_THEME_FORMAT = 'sir-john-shell-theme';
const SJS_THEME_VERSION = 1;
const SJS_FONT_PRESETS = [
  { value: 'ui-monospace', label: 'SYSTEM MONOSPACE' },
  { value: '"JetBrains Mono"', label: 'JETBRAINS MONO' },
  { value: '"Fira Code"', label: 'FIRA CODE' },
  { value: '"IBM Plex Mono"', label: 'IBM PLEX MONO' },
  { value: '"Courier New"', label: 'COURIER NEW' }
];
const SJS_SCHEMES = {
  khaki: {
    name: 'Khaki',
    palette: { bg: '#101010', panel: '#101010', window: '#101010', border: '#242424', green: '#9ac16a', greenDim: '#4a5a2f', cyan: '#8be9fd', yellow: '#ffbd2e', red: '#ff5f56', gray: '#888888', dim: '#555555', userText: '#9ac16a', agentText: '#c0c0c0', userName: '#ffbd2e', agentName: '#8be9fd' }
  },
  monokai: {
    name: 'Monokai',
    palette: { bg: '#272822', panel: '#1e1f1c', window: '#2d2e27', border: '#49483e', green: '#a6e22e', greenDim: '#557018', cyan: '#66d9ef', yellow: '#e6db74', red: '#f92672', gray: '#c5c8b6', dim: '#75715e', userText: '#a6e22e', agentText: '#f8f8f2', userName: '#e6db74', agentName: '#66d9ef' }
  },
  dracula: {
    name: 'Dracula',
    palette: { bg: '#282a36', panel: '#21222c', window: '#343746', border: '#6272a4', green: '#50fa7b', greenDim: '#236b3b', cyan: '#8be9fd', yellow: '#f1fa8c', red: '#ff5555', gray: '#f8f8f2', dim: '#6272a4', userText: '#50fa7b', agentText: '#f8f8f2', userName: '#f1fa8c', agentName: '#8be9fd' }
  },
  nord: {
    name: 'Nord',
    palette: { bg: '#2e3440', panel: '#3b4252', window: '#434c5e', border: '#4c566a', green: '#a3be8c', greenDim: '#4f6846', cyan: '#88c0d0', yellow: '#ebcb8b', red: '#bf616a', gray: '#d8dee9', dim: '#616e88', userText: '#a3be8c', agentText: '#d8dee9', userName: '#ebcb8b', agentName: '#88c0d0' }
  },
  solarized: {
    name: 'Solarized',
    palette: { bg: '#002b36', panel: '#073642', window: '#0b3d48', border: '#586e75', green: '#859900', greenDim: '#4d5900', cyan: '#2aa198', yellow: '#b58900', red: '#dc322f', gray: '#93a1a1', dim: '#657b83', userText: '#859900', agentText: '#eee8d5', userName: '#b58900', agentName: '#2aa198' }
  }
};

function sjsDebugLog(event, data) {
  if (!window.fetch) return;
  fetch('/debug/theme-log', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ event, data: data || {} }),
    keepalive: true
  }).catch(() => {});
}

function sjsApplyFont(font) {
  const allowed = SJS_FONT_PRESETS.map(item => item.value);
  const selected = allowed.includes(font) ? font : 'ui-monospace';
  document.documentElement.style.setProperty('--ui-font', selected);
  return selected;
}

const SJS_THEME_DEFAULTS = {
  name: 'Unnamed theme',
  palette: {
    bg: '#141414', panel: '#141414', window: '#141414', border: '#242424',
    green: '#9ac16a', greenDim: '#4a5a2f', cyan: '#8be9fd',
    yellow: '#ffbd2e', red: '#ff5f56', gray: '#888888', dim: '#555555',
    userText: '#9ac16a', agentText: '#c0c0c0', userName: '#ffbd2e', agentName: '#8be9fd'
  },
  brightness: { user: 100, agent: 100, terminal: 100, thoughts: 100, workspace: 100, dim: 100 },
  backgrounds: {
    global: { type: 'scanlines', c1: '#141414', c2: '#242424', opacity: 0.03 },
    chat: { type: 'scanlines', c1: '#141414', c2: '#242424', opacity: 0.5 },
    workspace: { type: 'scanlines', c1: '#141414', c2: '#242424', opacity: 0.5 },
    thoughts: { type: 'scanlines', c1: '#141414', c2: '#242424', opacity: 0.5 },
    terminal: { type: 'scanlines', c1: '#141414', c2: '#242424', opacity: 0.5 }
  },
  chat: {
    user: { background: 'transparent', border: '#4a5a2f' },
    agent: { background: 'transparent', border: '#242424' }
  },
  scanlines: false
};

function sjsClone(value) {
  return JSON.parse(JSON.stringify(value));
}

function sjsIsColor(value) {
  return typeof value === 'string' && /^#[0-9a-f]{6}$/i.test(value);
}

function sjsColor(value, fallback) {
  return sjsIsColor(value) ? value : fallback;
}

function sjsClamp(value, min, max) {
  return Math.min(max, Math.max(min, Number(value) || min));
}

function sjsNormalizeTheme(input) {
  const source = input && typeof input === 'object' ? input : {};
  const uiSource = source.ui && typeof source.ui === 'object' ? source.ui : source;
  const hasUI = Boolean(source.ui || source.backgrounds || source.chat || Object.prototype.hasOwnProperty.call(source, 'scanlines'));
  const theme = sjsClone(SJS_THEME_DEFAULTS);
  theme.scheme = typeof source.scheme === 'string' ? source.scheme : '';
  theme.name = typeof source.name === 'string' && source.name.trim() ? source.name.trim() : theme.name;

  const palette = source.palette || {};
  for (const key of Object.keys(theme.palette)) {
    theme.palette[key] = sjsColor(palette[key], theme.palette[key]);
  }
  for (const key of Object.keys(theme.brightness)) {
    theme.brightness[key] = Math.round(sjsClamp(source.brightness?.[key], 50, 200));
  }
  for (const key of Object.keys(theme.backgrounds)) {
    const value = uiSource.backgrounds?.[key] || {};
    const target = theme.backgrounds[key];
    target.type = ['none', 'mono', 'tg-gradient', 'tg-dots', 'scanlines', 'noise'].includes(value.type) ? value.type : target.type;
    target.c1 = sjsColor(value.c1, target.c1);
    target.c2 = sjsColor(value.c2, target.c2);
    target.opacity = sjsClamp(value.opacity, 0, 1);
  }
  for (const key of Object.keys(theme.chat)) {
    const value = uiSource.chat?.[key] || {};
    theme.chat[key].background = value.background === 'transparent' ? 'transparent' : sjsColor(value.background, theme.chat[key].background);
    theme.chat[key].border = sjsColor(value.border, theme.chat[key].border);
  }
  theme.scanlines = Boolean(uiSource.scanlines);
  theme.ui = hasUI ? { backgrounds: theme.backgrounds, chat: theme.chat, scanlines: theme.scanlines } : null;
  return theme;
}

function sjsThemePreset(scheme) {
  const preset = SJS_SCHEMES[scheme] || SJS_SCHEMES.khaki;
  const theme = sjsClone(SJS_THEME_DEFAULTS);
  theme.name = preset.name;
  theme.scheme = scheme;
  Object.assign(theme.palette, preset.palette);
  return { name: theme.name, scheme: theme.scheme, palette: theme.palette, brightness: theme.brightness };
}

function sjsMakeNoiseData(color, opacity) {
  return `url("data:image/svg+xml,%3Csvg viewBox='0 0 200 200' xmlns='http://www.w3.org/2000/svg'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.8' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' fill='${encodeURIComponent(color)}'/%3E%3Crect width='100%25' height='100%25' filter='url(%23n)' opacity='${opacity}'/%3E%3C/svg%3E")`;
}

function sjsBackgroundCss(config) {
  const c1 = config.c1;
  const c2 = config.c2;
  switch (config.type) {
    case 'none': return 'transparent';
    case 'tg-gradient': return `linear-gradient(135deg, ${c1}, ${c2})`;
    case 'tg-dots': return `radial-gradient(circle at 2px 2px, ${c2} 1px, transparent 1px), ${c1}`;
    case 'scanlines': return `repeating-linear-gradient(to bottom, ${c2} 0px, ${c2} 1px, transparent 1px, transparent 3px), ${c1}`;
    case 'noise': return `${sjsMakeNoiseData(c2, config.opacity)}, ${c1}`;
    case 'mono':
    default: return c1;
  }
}

function sjsBrightness(hex, percent) {
  const match = hex.match(/^#(..)(..)(..)$/);
  if (!match) return hex;
  const rgb = match.slice(1).map(v => parseInt(v, 16) / 255);
  const max = Math.max(...rgb), min = Math.min(...rgb);
  let h = 0, s = 0, l = (max + min) / 2;
  if (max !== min) {
    const d = max - min;
    s = l > 0.5 ? d / (2 - max - min) : d / (max + min);
    if (max === rgb[0]) h = (rgb[1] - rgb[2]) / d + (rgb[1] < rgb[2] ? 6 : 0);
    else if (max === rgb[1]) h = (rgb[2] - rgb[0]) / d + 2;
    else h = (rgb[0] - rgb[1]) / d + 4;
    h /= 6;
  }
  l = Math.min(1, Math.max(0.05, l * percent / 100));
  const hue = (p, q, t) => {
    if (t < 0) t += 1;
    if (t > 1) t -= 1;
    if (t < 1 / 6) return p + (q - p) * 6 * t;
    if (t < 1 / 2) return q;
    if (t < 2 / 3) return p + (q - p) * (2 / 3 - t) * 6;
    return p;
  };
  if (s === 0) rgb[0] = rgb[1] = rgb[2] = l;
  else {
    const q = l < 0.5 ? l * (1 + s) : l + s - l * s;
    const p = 2 * l - q;
    rgb[0] = hue(p, q, h + 1 / 3); rgb[1] = hue(p, q, h); rgb[2] = hue(p, q, h - 1 / 3);
  }
  return '#' + rgb.map(v => Math.round(v * 255).toString(16).padStart(2, '0')).join('');
}

function sjsApplyTheme(input) {
  const theme = sjsNormalizeTheme(input);
  const root = document.documentElement;
  const p = theme.palette;
  const vars = {
    userText: sjsBrightness(p.userText, theme.brightness.user),
    agentText: sjsBrightness(p.agentText, theme.brightness.agent), userName: p.userName, agentName: p.agentName,
    terminalText: sjsBrightness(p.gray, theme.brightness.terminal),
    thoughtsText: sjsBrightness(p.gray, theme.brightness.thoughts),
    workspaceText: sjsBrightness(p.gray, theme.brightness.workspace), dim: sjsBrightness(p.dim, theme.brightness.dim)
  };
  // Themes are text-only. Surfaces, borders, layout and effects remain product UI.
  for (const [key, value] of Object.entries(vars)) root.style.setProperty(`--${key.replace(/[A-Z]/g, m => '-' + m.toLowerCase())}`, value);
  if (theme.ui) {
    const ui = theme.ui;
    for (const [key, target] of Object.entries({
      global: '--global-background', chat: '--chat-background', workspace: '--workspace-background',
      thoughts: '--thoughts-background', terminal: '--terminal-background'
    })) root.style.setProperty(target, sjsBackgroundCss(ui.backgrounds[key]));
    root.style.setProperty('--user-message-background', ui.chat.user.background);
    root.style.setProperty('--user-message-border', ui.chat.user.border);
    root.style.setProperty('--agent-message-background', ui.chat.agent.background);
    root.style.setProperty('--agent-message-border', ui.chat.agent.border);
    document.body.classList.toggle('theme-scanlines', ui.scanlines);
  }
  return theme;
}

function sjsExportTheme(input, name) {
  const theme = sjsNormalizeTheme(input);
  if (name) theme.name = name;
  const result = {
    format: SJS_THEME_FORMAT,
    version: SJS_THEME_VERSION,
    name: theme.name,
    scheme: theme.scheme,
    palette: theme.palette,
    brightness: theme.brightness
  };
  if (theme.ui) result.ui = theme.ui;
  return result;
}

function sjsDownloadTheme(input) {
  const theme = sjsExportTheme(input);
  const blob = new Blob([JSON.stringify(theme, null, 2) + '\n'], { type: 'application/json' });
  const link = document.createElement('a');
  link.href = URL.createObjectURL(blob);
  link.download = (theme.name || 'sir-john-shell-theme').toLowerCase().replace(/[^a-z0-9-_]+/gi, '-') + '.json';
  link.click();
  URL.revokeObjectURL(link.href);
}

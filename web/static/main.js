/* ============================================================
   DEVIN.SYS // WEB UI — live WebSocket client
   ============================================================ */

const $ = (id) => document.getElementById(id);

function esc(s) {
  return s.replace(/[&<>]/g, c => ({ '&':'&amp;','<':'&lt;','>':'&gt;' }[c]));
}

function autoScroll(el) {
  el.scrollTop = el.scrollHeight;
}

/* ============================================================
   FOCUS SWAP
   ============================================================ */
function initSwap() {
  const mainSlot = $('slot-main');
  const sideIds = ['slot-side1', 'slot-side2'];
  document.querySelectorAll('.titlebar').forEach(bar => {
    bar.addEventListener('click', (e) => {
      if (e.target.tagName === 'BUTTON') return; // don't swap on switcher etc
      const slot = e.currentTarget.closest('.slot');
      if (!slot) return;
      const a = mainSlot.firstElementChild;
      if (!a) return;

      if (slot === mainSlot) {
        // main panel: bring .main-view back to main, push current main to its side slot
        const target = sideIds.map(id => $(id)).find(s => s.firstElementChild && s.firstElementChild.id === 'main-view');
        if (!target) return;
        const b = target.firstElementChild;
        mainSlot.appendChild(b);
        target.appendChild(a);
      } else {
        const b = slot.firstElementChild;
        if (b) {
          mainSlot.appendChild(b);
          slot.appendChild(a);
        }
      }
    });
  });
}

/* ============================================================
   SYNTAX HIGHLIGHTING
   ============================================================ */

const GO_KEYWORDS = ['package','import','func','var','const','type','struct','interface','return','if','else','for','range','go','defer','chan','select','switch','case','default','break','continue','map','nil','true','false','make','new','len','cap','append'];
const GO_TYPES = ['int','int64','int32','uint8','float64','string','bool','byte','rune','error','any'];
const JS_KEYWORDS = ['const','let','var','function','return','if','else','for','while','switch','case','break','continue','new','this','true','false','null','undefined','async','await','import','export','from','class'];

function highlightByRegex(code, kw, ty, fnRegex) {
  const kwRe = kw.join('|');
  const tyRe = ty.join('|');
  const parts = [
    '(//[^\\n]*)',
    '(/\\*[^]*?\\*/)',
    '("(?:[^"\\\\]|\\\\.)*")',
    "('(?:[^'\\\\]|\\\\.)*')",
    '(`[^`]*`)',
    '\\b(\\d+(?:\\.\\d+)?)\\b',
    '\\b(' + kwRe + ')\\b',
  ];
  if (tyRe) parts.push('\\b(' + tyRe + ')\\b');
  if (fnRegex) parts.push(fnRegex);
  parts.push('([{}()\\[\\];,.*+/=<>!:-])');
  const re = new RegExp(parts.join('|'), 'g');

  let out = '', last = 0, m;
  while ((m = re.exec(code)) !== null) {
    out += esc(code.slice(last, m.index));
    const cls = m[1] ? 'com' : m[2] ? 'com' : m[3] ? 'str' : m[4] ? 'str' : m[5] ? 'str' :
                m[6] ? 'num' : m[7] ? 'kw' : m[8] ? 'typ' : m[9] ? 'fn' : m[10] ? 'pun' : '';
    out += cls ? `<span class="${cls}">${esc(m[0])}</span>` : esc(m[0]);
    last = re.lastIndex;
  }
  out += esc(code.slice(last));
  return out;
}

function highlightGo(code) { return highlightByRegex(code, GO_KEYWORDS, GO_TYPES, '\\b([A-Za-z_]\\w*)(?=\\s*\\()'); }
function highlightJS(code) { return highlightByRegex(code, JS_KEYWORDS, [], '\\b([A-Za-z_$][\\w$]*)(?=\\s*\\()'); }

function highlightCSS(code) {
  const re = /(\/\*[^]*?\*\/)|("(?:[^"\\]|\\.)*")|('(?:[^'\\]|\\.)*')|(\b\d+(?:\.\d+)?(?:px|em|rem|%|s|ms|vh|vw|ch|ex|fr|deg)?\b)|([\w-]+)(?=\s*:)|([\w-]+)(?=\s*\()|([{}();,:\-])|(\.[A-Za-z_-][\w-]*)|(#\w+)/g;
  let out = '', last = 0, m;
  while ((m = re.exec(code)) !== null) {
    out += esc(code.slice(last, m.index));
    const cls = m[1] ? 'com' : m[2] ? 'str' : m[3] ? 'str' : m[4] ? 'num' : m[5] ? 'typ' : m[6] ? 'fn' : m[7] ? 'pun' : m[8] ? 'fn' : m[9] ? 'num' : '';
    out += cls ? `<span class="${cls}">${esc(m[0])}</span>` : esc(m[0]);
    last = re.lastIndex;
  }
  out += esc(code.slice(last));
  return out;
}

function highlightGeneric(code) {
  const re = /(\/\/[^\n]*)|(\/\*[^]*?\*\/)|("(?:[^"\\]|\\.)*")|('(?:[^'\\]|\\.)*')|(`[^`]*`)|(\b\d+(?:\.\d+)?\b)|([{}()\[\];,=+\-*\/])|(\b\w+\b)/g;
  let out = '', last = 0, m;
  while ((m = re.exec(code)) !== null) {
    out += esc(code.slice(last, m.index));
    const cls = m[1] ? 'com' : m[2] ? 'com' : m[3] ? 'str' : m[4] ? 'str' : m[5] ? 'str' :
                m[6] ? 'num' : m[7] ? 'pun' : '';
    out += cls ? `<span class="${cls}">${esc(m[0])}</span>` : esc(m[0]);
    last = re.lastIndex;
  }
  out += esc(code.slice(last));
  return out;
}

function detectLang(name, content) {
  if (/\.css$/i.test(name)) return 'css';
  if (/\.js$/i.test(name)) return 'js';
  if (/\.go$/i.test(name)) return 'go';
  if (/\.json$/i.test(name)) return 'json';
  if (/\bpackage\s+\w+\b/.test(content) && /\bfunc\b/.test(content)) return 'go';
  if (/\bbody\s*\{/.test(content) && /:\s*#/.test(content)) return 'css';
  return 'generic';
}

function highlightCode(name, content) {
  const lang = detectLang(name, content);
  if (lang === 'css') return highlightCSS(content);
  if (lang === 'js') return highlightJS(content);
  if (lang === 'go') return highlightGo(content);
  return highlightGeneric(content);
}

/* ============================================================
   TEXT ENRICHMENT
   ============================================================ */
function enrichText(text) {
  let s = esc(text);
  s = s.replace(/`([^`]+)`/g, '<code class="str">$1</code>');
  s = s.replace(/"([^"]*)"/g, '<span class="str">"$1"</span>');
  s = s.replace(/'([^']*)'/g, '<span class="str">\'$1\'</span>');
  s = s.replace(/(https?:\/\/\S+)/g, '<a class="typ" href="$1" target="_blank">$1</a>');
  s = s.replace(/(\/(?:[A-Za-z0-9_.\-]+(?:\/|$))+)/g, '<span class="typ">$1</span>');
  s = s.replace(/\b(\d+(?:\.\d+)?)\b/g, '<span class="num">$1</span>');
  s = s.replace(/\b([A-Za-z0-9_-]+\.(go|css|js|json|html|md|txt|mod|sum))\b/g, '<span class="typ">$1</span>');
  return s;
}

/* ============================================================
   VIEW SWITCHER
   ============================================================ */
document.querySelectorAll('.sw').forEach(btn => {
  btn.addEventListener('click', () => {
    const view = btn.dataset.view;
    document.querySelectorAll('.sw').forEach(b => b.classList.remove('active'));
    document.querySelectorAll('.view').forEach(v => v.classList.remove('active'));
    btn.classList.add('active');
    $(`view-${view}`).classList.add('active');
  });
});

/* ============================================================
   CHAT
   ============================================================ */
const currentMsg = { agent: null, thought: null };

function addUserMessage(text) {
  const el = document.createElement('div');
  el.className = 'msg user';
  el.innerHTML = `<div class="meta">Sir John Smith ${timeNow()}</div>` + enrichText(text);
  $('chat-body').appendChild(el);
  autoScroll($('chat-body'));
}

function ensureAgentMessage() {
  if (!currentMsg.agent) {
    const el = document.createElement('div');
    el.className = 'msg agent';
    el.innerHTML = `<div class="meta">AGENT ${timeNow()}</div><span class="text"></span>`;
    $('chat-body').appendChild(el);
    currentMsg.agent = el.querySelector('.text');
  }
  return currentMsg.agent;
}

function appendAgentText(text) {
  const span = ensureAgentMessage();
  span.textContent += text;
  autoScroll($('chat-body'));
}

function finalizeAgentMessage() {
  if (!currentMsg.agent) return;
  const span = currentMsg.agent;
  const text = span.textContent;
  span.innerHTML = enrichText(text);
  currentMsg.agent = null;
}

function ensureThought() {
  if (!currentMsg.thought) {
    const el = document.createElement('div');
    el.className = 'thought';
    el.innerHTML = '<div class="t-title">thinking</div><div class="t-text"></div><div class="t-status">running</div>';
    $('thoughts-body').appendChild(el);
    currentMsg.thought = el.querySelector('.t-text');
  }
  return currentMsg.thought;
}

function appendThought(text) {
  const t = ensureThought();
  t.textContent += text;
  autoScroll($('thoughts-body'));
}

function finalizeThought() {
  if (!currentMsg.thought) return;
  const el = currentMsg.thought.parentElement;
  const status = el.querySelector('.t-status');
  status.textContent = 'done';
  currentMsg.thought = null;
}

/* ============================================================
   TERMINAL / TOOLS / SKILLS
   ============================================================ */
function logTerminal(level, text) {
  const panel = $('panel-terminal');
  const line = document.createElement('div');
  line.className = `t-line t-${level}`;
  line.textContent = text;
  panel.appendChild(line);
  autoScroll(panel);
}

function renderConfig(configOptions) {
  const panel = $('panel-skills');
  let html = '';
  configOptions.forEach(opt => {
    html += `<div style="color:#666;margin:10px 0 4px;">${esc(opt.name.toUpperCase())}</div>`;
    html += `<div class="list-row"><span class="code">${esc(opt.id)}</span><span class="name">${esc(opt.currentValue)}</span></div>`;
    if (opt.options) {
      opt.options.forEach(o => {
        html += `<div class="list-row"><span class="code">${esc(o.value)}</span><span class="name">${esc(o.name)}</span></div>`;
      });
    }
  });
  panel.innerHTML = html;
}

function renderSkills(commands) {
  const panel = $('panel-skills');
  panel.innerHTML = '';
  commands.forEach(cmd => {
    const row = document.createElement('div');
    row.className = 'list-row';
    row.innerHTML = `<span class="code">/${esc(cmd.name)}</span><span class="name">${esc(cmd.description || '')}</span>`;
    panel.appendChild(row);
  });
}

/* ============================================================
   WORKSPACE
   ============================================================ */
function renderTree(files) {
  const tree = $('ws-tree');
  tree.innerHTML = '';
  files.forEach(f => {
    const row = document.createElement('div');
    row.className = f.dir ? 'f dir' : 'f';
    row.textContent = (f.dir ? '[D] ' : '[F] ') + f.name;
    row.addEventListener('click', () => showFile(f.name, f.content || '// empty'));
    tree.appendChild(row);
  });
}

function showFile(name, content) {
  const file = $('ws-file');
  const highlighted = highlightCode(name, content);
  file.innerHTML = `<div style="color:#666;margin-bottom:6px;">${esc(name)}</div>` +
                   `<pre class="code-block" style="white-space:pre-wrap;">${highlighted}</pre>`;
}

/* ============================================================
   PROMPT
   ============================================================ */
const promptEl = $('prompt');
promptEl.addEventListener('keydown', (e) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    sendPrompt();
  }
});

$('send-btn').addEventListener('click', sendPrompt);

function sendPrompt() {
  const text = promptEl.value.trim();
  if (!text || !socket) return;
  socket.send(JSON.stringify({ type: 'prompt', text }));
  addUserMessage(text);
  promptEl.value = '';
  promptEl.rows = 1;
  currentMsg.agent = null;
}

/* ============================================================
   WEBSOCKET
   ============================================================ */
let socket;

function timeNow() {
  const d = new Date();
  return `${String(d.getHours()).padStart(2,'0')}:${String(d.getMinutes()).padStart(2,'0')}`;
}

function connect() {
  const proto = location.protocol === 'https:' ? 'wss' : 'ws';
  const url = `${proto}://${location.host}/ws`;
  socket = new WebSocket(url);

  socket.onopen = () => {
    $('conn-pill').innerHTML = '<span class="led ok"></span> ONLINE';
    logTerminal('info', '[ws] connected');
  };

  socket.onclose = () => {
    $('conn-pill').innerHTML = '<span class="led" style="background:#666"></span> OFFLINE';
    logTerminal('warn', '[ws] disconnected');
    setTimeout(connect, 3000);
  };

  socket.onerror = (e) => {
    logTerminal('err', '[ws] error');
  };

  socket.onmessage = (e) => {
    try {
      const msg = JSON.parse(e.data);
      handleMessage(msg);
    } catch (err) {
      logTerminal('err', '[ws] invalid json: ' + e.data.slice(0, 80));
    }
  };
}

function handleMessage(msg) {
  // Server welcome
  if (msg.type === 'welcome') {
    $('session-pill').textContent = 'SESSION: ' + msg.sessionId;
    $('ws-path').textContent = msg.cwd;
    logTerminal('info', '[session] ' + msg.sessionId);
    return;
  }

  const method = msg.method || '';
  const params = msg.params || {};

  if (method === 'session/update') {
    const up = params.update || {};
    const type = up.sessionUpdate;

    switch (type) {
      case 'agent_message_chunk':
        appendAgentText(chunkText(up.content));
        break;

      case 'agent_thought_chunk':
        appendThought(chunkText(up.content));
        break;

      case 'config_option_update':
        updateHeaderFromConfig(up.configOptions);
        renderConfig(up.configOptions);
        break;

      case 'available_commands_update':
        renderSkills(up.availableCommands);
        break;

      case 'current_mode_update':
        $('mode-pill').textContent = 'MODE: ' + up.currentModeId;
        break;

      case 'session_info_update':
        if (up.title) logTerminal('info', '[title] ' + up.title);
        break;

      case 'usage_update':
        const meta = up._meta || {};
        const out = meta['cognition.ai/outputTokens'] || 0;
        const inTok = meta['cognition.ai/inputTokens'] || 0;
        logTerminal('info', `[usage] in:${inTok} out:${out}`);
        if (typeof up.used === 'number' && typeof up.size === 'number') {
          updateContextBar(up.used, up.size);
        }
        break;

      case 'command_revision':
      case 'command_complete':
      case 'command_start':
        logTerminal('info', `[cmd] ${type}`);
        break;

      case 'permission_request':
        logTerminal('warn', `[perm] ${up.toolName || '?'}`);
        break;

      case 'tool_call_update':
        logTerminal('info', `[tool] ${up.toolName || '?'}`);
        break;

      default:
        logTerminal('info', '[session/update] ' + type);
    }
  } else if (method === '_cognition.ai/output') {
    const level = params.level || 'info';
    const channel = params.channel || 'devin';
    const text = `[${channel}] ${params.message}`;
    logTerminal(level, text);
  } else if (method === '_cognition.ai/agent_stopped') {
    finalizeAgentMessage();
    finalizeThought();
    const stats = params.stats || {};
    logTerminal('info', `[done] outputTokens=${stats.outputTokens} toolCalls=${stats.toolCalls}`);
  } else if (method === '_cognition.ai/thinking_complete') {
    finalizeThought();
  } else if (method === '_cognition.ai/mcp/serversChanged') {
    logTerminal('info', '[mcp] servers changed');
  } else if (method === '_cognition.ai/turn_stats') {
    logTerminal('info', '[stats] turn complete');
  } else {
    logTerminal('info', `[${method}] ${JSON.stringify(params).slice(0,80)}`);
  }
}

function chunkText(content) {
  if (!content) return '';
  if (typeof content === 'string') return content;
  if (content.type === 'text' && content.text) return content.text;
  return JSON.stringify(content);
}

function populateSelect(id, label, opt) {
  const sel = $(id);
  if (!sel || !opt || !opt.options) return;
  const current = opt.currentValue || '';
  const wasOpen = document.activeElement === sel;
  sel.innerHTML = '';
  opt.options.forEach(o => {
    const option = document.createElement('option');
    option.value = o.value;
    option.textContent = o.name || o.value;
    if (o.value === current) option.selected = true;
    sel.appendChild(option);
  });
  const first = document.createElement('option');
  first.value = '';
  first.textContent = label;
  first.disabled = true;
  sel.prepend(first);
  if (wasOpen) sel.blur();
}

function updateHeaderFromConfig(opts) {
  if (!opts) return;
  opts.forEach(opt => {
    if (opt.id === 'mode' && opt.currentValue) {
      populateSelect('mode-select', 'MODE', opt);
      const ms = $('mode-select');
      if (ms) ms.dataset.current = opt.currentValue;
    }
    if (opt.id === 'model' && opt.currentValue) {
      populateSelect('model-select', 'MODEL', opt);
      const ms = $('model-select');
      if (ms) ms.dataset.current = opt.currentValue;
      $('cf-model').textContent = opt.currentValue;
    }
  });
}

function formatTokens(n) {
  if (!n) return '0';
  if (n >= 1000000) return (n / 1000000).toFixed(1).replace(/\.0$/, '') + 'M';
  if (n >= 1000) return (n / 1000).toFixed(0) + 'K';
  return String(n);
}

function contextColor(pct, light) {
  const h = Math.max(0, Math.round(90 - pct * 0.9));
  return `hsl(${h}, 35%, ${light}%)`;
}

function updateContextBar(used, size) {
  used = Number(used) || 0;
  size = Number(size) || 1;
  const pct = Math.min(100, Math.max(0, Math.round(used / size * 100)));
  const fill = $('ctx-fill');
  fill.style.width = pct + '%';
  fill.style.background = `linear-gradient(90deg, ${contextColor(pct, 48)} 0%, ${contextColor(pct, 30)} 100%)`;
  fill.style.boxShadow = `0 0 5px ${contextColor(pct, 42)}`;
  $('cf-ctx').textContent = `${formatTokens(used)} / ${formatTokens(size)} (${pct}%)`;
}

function demoContextBar() {
  const size = 262000;
  let start = null;
  const duration = 4000;
  function frame(ts) {
    if (!start) start = ts;
    const p = Math.min(1, (ts - start) / duration);
    const used = Math.round(p * size);
    updateContextBar(used, size);
    if (p < 1) {
      requestAnimationFrame(frame);
    } else {
      setTimeout(() => updateContextBar(0, size), 600);
    }
  }
  requestAnimationFrame(frame);
}

/* ============================================================
   VOICE
   ============================================================ */
let recording = false;
const voiceBtn = $('voice-btn');
let mediaRecorder;
let recordedChunks = [];

voiceBtn.addEventListener('click', async () => {
  if (!navigator.mediaDevices) {
    logTerminal('err', '[voice] not supported');
    return;
  }
  if (!recording) {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true });
      mediaRecorder = new MediaRecorder(stream);
      recordedChunks = [];
      mediaRecorder.ondataavailable = (e) => recordedChunks.push(e.data);
      mediaRecorder.onstop = async () => {
        // TODO: send audio to Go endpoint for transcription
        logTerminal('info', '[voice] recording stopped');
      };
      mediaRecorder.start();
      recording = true;
      voiceBtn.classList.add('recording');
      logTerminal('info', '[voice] recording started');
    } catch (err) {
      logTerminal('err', '[voice] ' + err.message);
    }
  } else {
    mediaRecorder.stop();
    recording = false;
    voiceBtn.classList.remove('recording');
  }
});

function initHeaderControls() {
  const modeCmd = { 'accept-edits': '/code', 'ask': '/ask', 'plan': '/plan', 'smart': '/smart', 'bypass': '/bypass' };

  $('mode-select').addEventListener('change', (e) => {
    const val = e.target.value;
    const cmd = modeCmd[val];
    if (cmd && socket && val !== e.target.dataset.current) {
      socket.send(JSON.stringify({ type: 'prompt', text: cmd }));
      addUserMessage(cmd);
    }
  });

  $('model-select').addEventListener('change', (e) => {
    const val = e.target.value;
    if (val && socket && val !== e.target.dataset.current) {
      socket.send(JSON.stringify({ type: 'prompt', text: `/model ${val}` }));
      addUserMessage(`/model ${val}`);
    }
  });
}

/* ============================================================
   DEMO INIT
   ============================================================ */
function initDemo() {
  initSwap();
  initHeaderControls();
  demoContextBar();
  renderTree([
    { name: 'cmd', dir: true },
    { name: 'internal', dir: true },
    { name: 'web', dir: true },
    { name: 'go.mod', dir: false, content: 'module sir-john-shell\n\ngo 1.26.5' },
    { name: 'main.go', dir: false, content: 'package main\n\nimport (\n\t"fmt"\n)\n\nfunc main() {\n\tfmt.Println("boot")\n}' },
    { name: 'style.css', dir: false, content: 'body {\n  background: #0a0a0a;\n  color: #9ac16a;\n}\n\n.title {\n  font-size: 13px;\n}' },
    { name: 'README.md', dir: false, content: '# Sir John Shell\n\nBrowser-native UI for Devin. Single Go binary.' }
  ]);
  showFile('main.go', 'package main\n\nimport (\n\t"fmt"\n)\n\nfunc main() {\n\tfmt.Println("boot")\n}');
  connect();
}

initDemo();

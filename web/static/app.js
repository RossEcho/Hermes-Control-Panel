/* ============================================================
   Hermes Control Panel — app.js
   Plain vanilla JS. No build step, no npm required.
   ============================================================ */

// ----------------------------------------------------------------
// Chat SSE streaming
// ----------------------------------------------------------------

let activeSSE = null;  // active EventSource
let activeSessionID = null;

/**
 * Send the composer message: POST /api/chat/send, then open SSE stream.
 */
async function sendMessage() {
  const input = document.getElementById('chat-input');
  const sessionIDEl = document.getElementById('session-id');
  if (!input || !sessionIDEl) return;

  const message = input.value.trim();
  const sessionID = sessionIDEl.value.trim();

  if (!message || !sessionID) {
    if (!sessionID) alert('No session selected. Create or open a session first.');
    return;
  }

  // Append user message immediately
  appendUserMessage(message);
  input.value = '';
  input.style.height = '';

  // Show streaming placeholder
  showStreamingMessage();
  setSendState(true);

  // POST the message
  let resp;
  try {
    const form = new FormData();
    form.append('session_id', sessionID);
    form.append('message', message);
    resp = await fetch('/api/chat/send', { method: 'POST', body: form });
    if (!resp.ok) throw new Error('send failed: ' + resp.status);
  } catch (err) {
    appendErrorMessage('Failed to send message: ' + err.message);
    setSendState(false);
    hideStreamingMessage();
    return;
  }

  const data = await resp.json();
  openSSEStream(data.stream_url, sessionID);
}

/**
 * Open an SSE connection to the stream URL.
 */
function openSSEStream(url, sessionID) {
  if (activeSSE) {
    activeSSE.close();
    activeSSE = null;
  }
  activeSessionID = sessionID;

  const es = new EventSource(url);
  activeSSE = es;

  let currentToolCallID = null;
  let currentToolCallEl = null;

  es.addEventListener('message_start', () => {
    // streaming placeholder already visible
  });

  es.addEventListener('tool_call_start', (e) => {
    const payload = safeJSON(e.data);
    if (!payload) return;
    currentToolCallID = payload.meta && payload.meta.tool_id;
    currentToolCallEl = createToolCallBlock(currentToolCallID, payload.meta && payload.meta.tool_name, 'running');
    const toolContainer = document.getElementById('streaming-tool-calls');
    if (toolContainer) toolContainer.appendChild(currentToolCallEl);
  });

  es.addEventListener('tool_call_update', (e) => {
    const payload = safeJSON(e.data);
    if (!payload || !currentToolCallEl) return;
    updateToolCallInput(currentToolCallEl, payload.content || '');
  });

  es.addEventListener('tool_call_end', (e) => {
    const payload = safeJSON(e.data);
    if (!payload || !currentToolCallEl) return;
    updateToolCallOutput(currentToolCallEl, payload.content || '');
    setToolCallStatus(currentToolCallEl, (payload.meta && payload.meta.status) || 'done');
    currentToolCallID = null;
    currentToolCallEl = null;
  });

  es.addEventListener('message_delta', (e) => {
    const payload = safeJSON(e.data);
    if (!payload) return;
    appendStreamingChunk(payload.content || '');
  });

  es.addEventListener('message_end', () => {
    finaliseStreamingMessage();
    setSendState(false);
    es.close();
    activeSSE = null;
  });

  es.addEventListener('error', (e) => {
    if (es.readyState === EventSource.CLOSED) {
      finaliseStreamingMessage();
      setSendState(false);
      activeSSE = null;
    } else {
      appendErrorMessage('Stream error. Connection may have dropped.');
      setSendState(false);
      hideStreamingMessage();
      es.close();
      activeSSE = null;
    }
  });
}

/**
 * Cancel the active SSE stream.
 */
async function stopStream() {
  if (activeSSE) {
    activeSSE.close();
    activeSSE = null;
  }
  const sessionID = document.getElementById('session-id');
  if (sessionID && sessionID.value) {
    const form = new FormData();
    form.append('session_id', sessionID.value);
    await fetch('/api/chat/stop', { method: 'POST', body: form }).catch(() => {});
  }
  finaliseStreamingMessage();
  setSendState(false);
}

// ----------------------------------------------------------------
// DOM helpers — chat
// ----------------------------------------------------------------

function appendUserMessage(text) {
  const list = document.getElementById('message-list');
  if (!list) return;
  const div = document.createElement('div');
  div.className = 'message message-user';
  div.innerHTML = `
    <div class="message-header">
      <span class="message-role">user</span>
      <span class="message-ts">${new Date().toISOString()}</span>
    </div>
    <div class="message-content">${escapeHTML(text)}</div>
  `;
  // Remove empty state if present
  const empty = list.querySelector('.message-empty');
  if (empty) empty.remove();
  // Insert before the streaming placeholder
  const streaming = document.getElementById('streaming-message');
  list.insertBefore(div, streaming);
  scrollToBottom();
}

function showStreamingMessage() {
  const el = document.getElementById('streaming-message');
  const ts = document.getElementById('streaming-ts');
  const content = document.getElementById('streaming-content');
  const indicator = document.getElementById('streaming-indicator');
  const toolCalls = document.getElementById('streaming-tool-calls');
  if (el) el.classList.remove('hidden');
  if (ts) ts.textContent = new Date().toISOString();
  if (content) content.textContent = '';
  if (indicator) indicator.classList.remove('hidden');
  if (toolCalls) toolCalls.innerHTML = '';
  scrollToBottom();
}

function hideStreamingMessage() {
  const el = document.getElementById('streaming-message');
  if (el) el.classList.add('hidden');
}

function appendStreamingChunk(chunk) {
  const content = document.getElementById('streaming-content');
  if (content) content.textContent += chunk;
  scrollToBottom();
}

function finaliseStreamingMessage() {
  const el = document.getElementById('streaming-message');
  const indicator = document.getElementById('streaming-indicator');
  if (indicator) indicator.classList.add('hidden');
  if (!el) return;

  const content = document.getElementById('streaming-content');
  if (content && content.textContent.trim()) {
    // Clone the streaming message into a permanent bubble
    const list = document.getElementById('message-list');
    const clone = el.cloneNode(true);
    clone.id = '';
    // Remove the streaming indicator from the clone
    const ind = clone.querySelector('.streaming-indicator');
    if (ind) ind.remove();
    list.insertBefore(clone, el);
  }
  // Reset streaming placeholder
  hideStreamingMessage();
  if (content) content.textContent = '';
  const toolCalls = document.getElementById('streaming-tool-calls');
  if (toolCalls) toolCalls.innerHTML = '';
  scrollToBottom();
}

function appendErrorMessage(text) {
  const list = document.getElementById('message-list');
  if (!list) return;
  const div = document.createElement('div');
  div.className = 'message';
  div.style.background = 'rgba(224,49,49,0.08)';
  div.style.border = '1px solid rgba(224,49,49,0.3)';
  div.innerHTML = `<div class="message-content" style="color:var(--error)">⚠ ${escapeHTML(text)}</div>`;
  const streaming = document.getElementById('streaming-message');
  list.insertBefore(div, streaming);
  scrollToBottom();
}

function scrollToBottom() {
  const list = document.getElementById('message-list');
  if (list) list.scrollTop = list.scrollHeight;
}

function setSendState(streaming) {
  const sendBtn = document.getElementById('send-btn');
  const stopBtn = document.getElementById('stop-btn');
  if (sendBtn) sendBtn.disabled = streaming;
  if (stopBtn) {
    if (streaming) stopBtn.classList.remove('hidden');
    else stopBtn.classList.add('hidden');
  }
}

// ----------------------------------------------------------------
// Tool call DOM helpers
// ----------------------------------------------------------------

function createToolCallBlock(id, name, status) {
  const el = document.createElement('details');
  el.className = 'tool-call-block';
  el.dataset.toolId = id || '';
  el.open = true;
  el.innerHTML = `
    <summary class="tool-call-summary">
      <span class="tool-icon">⚙</span>
      <span class="tool-name">${escapeHTML(name || 'tool')}</span>
      <span class="tool-status tool-status-${escapeHTML(status)}">${escapeHTML(status)}</span>
    </summary>
    <div class="tool-call-body">
      <div class="tool-section-label">Input</div>
      <pre class="tool-pre tool-input"></pre>
    </div>
  `;
  return el;
}

function updateToolCallInput(el, text) {
  const pre = el.querySelector('.tool-input');
  if (pre) pre.textContent += text;
}

function updateToolCallOutput(el, text) {
  const body = el.querySelector('.tool-call-body');
  if (!body) return;
  const label = document.createElement('div');
  label.className = 'tool-section-label';
  label.textContent = 'Output';
  const pre = document.createElement('pre');
  pre.className = 'tool-pre';
  pre.textContent = text;
  body.appendChild(label);
  body.appendChild(pre);
}

function setToolCallStatus(el, status) {
  const badge = el.querySelector('.tool-status');
  if (badge) {
    badge.className = `tool-status tool-status-${status}`;
    badge.textContent = status;
  }
}

// ----------------------------------------------------------------
// Skills filter
// ----------------------------------------------------------------

function filterSkills(query) {
  const items = document.querySelectorAll('#skill-list .skill-item');
  const q = query.toLowerCase().trim();
  items.forEach(item => {
    const name = (item.dataset.name || '').toLowerCase();
    const category = (item.dataset.category || '').toLowerCase();
    const match = !q || name.includes(q) || category.includes(q);
    item.style.display = match ? '' : 'none';
  });
}

function selectSkill(id) {
  // Deselect all
  document.querySelectorAll('#skill-list .skill-item').forEach(el => {
    el.classList.remove('skill-selected');
  });
  document.querySelectorAll('.skill-detail-panel').forEach(el => {
    el.classList.add('hidden');
  });
  document.querySelector('.skill-placeholder') && document.querySelector('.skill-placeholder').classList.add('hidden');

  // Select target
  const item = document.querySelector(`[data-id="${CSS.escape(id)}"]`);
  if (item) item.classList.add('skill-selected');

  const panel = document.getElementById(`skill-panel-${id}`);
  if (panel) panel.classList.remove('hidden');
}

// ----------------------------------------------------------------
// Keyboard shortcuts
// ----------------------------------------------------------------

document.addEventListener('keydown', function(e) {
  const input = document.getElementById('chat-input');
  if (!input) return;

  // Enter (without Shift) = send; Shift+Enter = newline
  if (e.target === input && e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    sendMessage();
  }
});

// Auto-resize textarea
document.addEventListener('input', function(e) {
  if (e.target && e.target.id === 'chat-input') {
    e.target.style.height = 'auto';
    e.target.style.height = Math.min(e.target.scrollHeight, 200) + 'px';
  }
});

// ----------------------------------------------------------------
// Copy-to-clipboard on code blocks
// ----------------------------------------------------------------

document.addEventListener('click', function(e) {
  if (e.target.classList.contains('copy-btn')) {
    const code = e.target.closest('.code-wrapper');
    if (code) {
      const pre = code.querySelector('pre');
      if (pre) navigator.clipboard.writeText(pre.textContent).then(() => {
        e.target.textContent = 'Copied!';
        setTimeout(() => { e.target.textContent = 'Copy'; }, 1500);
      });
    }
  }
});

// ----------------------------------------------------------------
// Utilities
// ----------------------------------------------------------------

function escapeHTML(str) {
  return String(str)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

function safeJSON(text) {
  try { return JSON.parse(text); } catch { return null; }
}

// ----------------------------------------------------------------
// Floating Chat Windows  (Windows messenger style)
// ----------------------------------------------------------------

const chatWindows = new Map(); // sessionID → { el, sse, minimized }
const CW_WIDTH  = 320;
const CW_GAP    = 8;
const CW_RIGHT  = 16;

/**
 * Open (or restore) a floating chat window for the given session.
 */
function openChatWindow(sessionID, title, status) {
  if (chatWindows.has(sessionID)) {
    const win = chatWindows.get(sessionID);
    if (win.minimized) cwToggleMinimize(sessionID);
    return;
  }

  const rightPos = CW_RIGHT + chatWindows.size * (CW_WIDTH + CW_GAP);
  const safeStatus = (status === 'running') ? 'running' : 'idle';

  const el = document.createElement('div');
  el.className = `chat-window cw-status-${safeStatus}`;
  el.id = `chatwin-${sessionID}`;
  el.style.right = rightPos + 'px';

  el.innerHTML = `
    <div class="cw-titlebar" onclick="cwToggleMinimize('${sessionID}')">
      <span class="cw-dot"></span>
      <span class="cw-title">${escapeHTML(title)}</span>
      <div class="cw-controls" onclick="event.stopPropagation()">
        <button class="cw-btn cw-enlarge"  title="Expand"   onclick="cwEnlarge('${sessionID}','${escapeHTML(title)}','${safeStatus}')">⛶</button>
        <button class="cw-btn cw-minimize" title="Minimize" onclick="cwToggleMinimize('${sessionID}')">—</button>
        <button class="cw-btn cw-close"    title="Close"    onclick="cwClose('${sessionID}')">×</button>
      </div>
    </div>
    <div class="cw-body" id="cwbody-${sessionID}">
      <div class="cw-messages" id="cwmsgs-${sessionID}">
        <div class="cw-loading">Loading…</div>
      </div>
      <div class="cw-composer">
        <textarea class="cw-input" id="cwinput-${sessionID}"
          placeholder="Message… (Enter to send)"
          rows="2"
          onkeydown="cwKeyDown(event,'${sessionID}')"></textarea>
        <div class="cw-actions">
          <button class="cw-stop hidden" id="cwstop-${sessionID}" title="Stop" onclick="cwStop('${sessionID}')">■</button>
          <button class="cw-send" title="Send" onclick="cwSend('${sessionID}')">▶</button>
        </div>
      </div>
    </div>
  `;

  document.body.appendChild(el);
  chatWindows.set(sessionID, { el, sse: null, minimized: false });
  cwLoadMessages(sessionID);
}

function cwClose(sessionID) {
  const win = chatWindows.get(sessionID);
  if (!win) return;
  if (win.sse) win.sse.close();
  win.el.remove();
  chatWindows.delete(sessionID);
  cwReposition();
}

function cwToggleMinimize(sessionID) {
  const win = chatWindows.get(sessionID);
  if (!win) return;
  win.minimized = !win.minimized;
  const body = document.getElementById(`cwbody-${sessionID}`);
  if (body) body.style.display = win.minimized ? 'none' : 'flex';
  win.el.classList.toggle('cw-minimized', win.minimized);
}

function cwReposition() {
  let i = 0;
  chatWindows.forEach(win => {
    win.el.style.right = (CW_RIGHT + i * (CW_WIDTH + CW_GAP)) + 'px';
    i++;
  });
}

async function cwLoadMessages(sessionID) {
  const msgsEl = document.getElementById(`cwmsgs-${sessionID}`);
  if (!msgsEl) return;
  try {
    const resp = await fetch(`/api/sessions/${sessionID}/messages`);
    if (!resp.ok) throw new Error('fetch failed');
    const data = await resp.json();
    const messages = data.messages || [];
    msgsEl.innerHTML = '';
    if (messages.length === 0) {
      msgsEl.innerHTML = '<div class="cw-empty">Send a message to start.</div>';
    } else {
      messages.forEach(msg => cwAppendMessage(sessionID, msg.Role || msg.role, msg.Content || msg.content));
    }
    cwScrollBottom(sessionID);
  } catch {
    msgsEl.innerHTML = '<div class="cw-empty">Could not load messages.</div>';
  }
}

function cwAppendMessage(sessionID, role, content) {
  const msgsEl = document.getElementById(`cwmsgs-${sessionID}`);
  if (!msgsEl) return;
  const empty = msgsEl.querySelector('.cw-empty, .cw-loading');
  if (empty) empty.remove();
  const div = document.createElement('div');
  div.className = `cw-message cw-msg-${role}`;
  div.innerHTML = `<span class="cw-msg-role">${escapeHTML(role)}</span><div class="cw-msg-content">${escapeHTML(content)}</div>`;
  msgsEl.appendChild(div);
}

async function cwSend(sessionID) {
  const input = document.getElementById(`cwinput-${sessionID}`);
  if (!input) return;
  const message = input.value.trim();
  if (!message) return;

  const win = chatWindows.get(sessionID);
  if (!win) return;

  input.value = '';
  cwAppendMessage(sessionID, 'user', message);
  cwScrollBottom(sessionID);

  // Streaming placeholder
  const msgsEl = document.getElementById(`cwmsgs-${sessionID}`);
  const streamEl = document.createElement('div');
  streamEl.className = 'cw-message cw-msg-assistant';
  streamEl.id = `cwstream-${sessionID}`;
  streamEl.innerHTML = `
    <span class="cw-msg-role">assistant</span>
    <div class="cw-msg-content" id="cwstreamtext-${sessionID}"></div>
    <div class="cw-blink"><span></span><span></span><span></span></div>
  `;
  if (msgsEl) msgsEl.appendChild(streamEl);
  cwScrollBottom(sessionID);

  // Disable send, show stop
  const stopBtn = document.getElementById(`cwstop-${sessionID}`);
  if (stopBtn) stopBtn.classList.remove('hidden');

  try {
    const form = new FormData();
    form.append('session_id', sessionID);
    form.append('message', message);
    const resp = await fetch('/api/chat/send', { method: 'POST', body: form });
    if (!resp.ok) throw new Error('send failed');
    const data = await resp.json();
    cwOpenSSE(sessionID, data.stream_url);
  } catch (err) {
    const textEl = document.getElementById(`cwstreamtext-${sessionID}`);
    if (textEl) { textEl.style.color = 'var(--error)'; textEl.textContent = '⚠ ' + err.message; }
    cwFinalise(sessionID);
  }
}

function cwOpenSSE(sessionID, url) {
  const win = chatWindows.get(sessionID);
  if (!win) return;
  if (win.sse) win.sse.close();

  const es = new EventSource(url);
  win.sse = es;

  es.addEventListener('message_delta', e => {
    const payload = safeJSON(e.data);
    if (!payload) return;
    const textEl = document.getElementById(`cwstreamtext-${sessionID}`);
    if (textEl) textEl.textContent += (payload.content || '');
    cwScrollBottom(sessionID);
  });

  es.addEventListener('message_end', () => { cwFinalise(sessionID); es.close(); win.sse = null; });
  es.addEventListener('error', ()      => { cwFinalise(sessionID); es.close(); win.sse = null; });
}

function cwFinalise(sessionID) {
  const streamEl = document.getElementById(`cwstream-${sessionID}`);
  if (streamEl) {
    streamEl.id = '';
    const blink = streamEl.querySelector('.cw-blink');
    if (blink) blink.remove();
    const textEl = document.getElementById(`cwstreamtext-${sessionID}`);
    if (textEl) textEl.id = '';
  }
  const stopBtn = document.getElementById(`cwstop-${sessionID}`);
  if (stopBtn) stopBtn.classList.add('hidden');
}

function cwStop(sessionID) {
  const win = chatWindows.get(sessionID);
  if (win && win.sse) { win.sse.close(); win.sse = null; }
  cwFinalise(sessionID);
  const form = new FormData();
  form.append('session_id', sessionID);
  fetch('/api/chat/stop', { method: 'POST', body: form }).catch(() => {});
}

function cwKeyDown(e, sessionID) {
  if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); cwSend(sessionID); }
}

function cwScrollBottom(sessionID) {
  const msgsEl = document.getElementById(`cwmsgs-${sessionID}`);
  if (msgsEl) msgsEl.scrollTop = msgsEl.scrollHeight;
}

// ----------------------------------------------------------------
// Enlarged modal chat window
// ----------------------------------------------------------------

// modalSSE tracks active streams for modal windows: sessionID → EventSource
const modalSSE = new Map();

function cwEnlarge(sessionID, title, status) {
  // Minimize the corner window so it doesn't overlap
  const win = chatWindows.get(sessionID);
  if (win && !win.minimized) cwToggleMinimize(sessionID);

  // If modal already exists, just show it
  const existing = document.getElementById(`cwmodal-${sessionID}`);
  if (existing) { existing.style.display = 'flex'; return; }

  const safeStatus = (status === 'running') ? 'running' : 'idle';

  const overlay = document.createElement('div');
  overlay.className = 'cw-modal-overlay';
  overlay.id = `cwmodal-${sessionID}`;
  // Click outside modal → collapse back to corner
  overlay.addEventListener('click', e => { if (e.target === overlay) cwCollapseModal(sessionID); });

  overlay.innerHTML = `
    <div class="cw-modal cw-status-${safeStatus}">
      <div class="cw-modal-titlebar">
        <span class="cw-dot"></span>
        <span class="cw-modal-title">${escapeHTML(title)}</span>
        <div class="cw-controls">
          <button class="cw-btn" title="Collapse to corner" onclick="cwCollapseModal('${sessionID}')">⊟</button>
          <button class="cw-btn cw-close" title="Close entirely" onclick="cwDestroyModal('${sessionID}')">×</button>
        </div>
      </div>
      <div class="cw-modal-messages" id="cwmodalmsgs-${sessionID}">
        <div class="cw-loading">Loading…</div>
      </div>
      <div class="cw-modal-composer">
        <textarea class="cw-input cw-modal-input" id="cwmodalinput-${sessionID}"
          placeholder="Message… (Enter to send, Shift+Enter for newline)"
          rows="3"
          onkeydown="cwModalKeyDown(event,'${sessionID}')"></textarea>
        <div class="cw-modal-actions">
          <button class="cw-stop hidden" id="cwmodalstop-${sessionID}" title="Stop" onclick="cwModalStop('${sessionID}')">■ Stop</button>
          <button class="cw-send cw-modal-send" title="Send" onclick="cwModalSend('${sessionID}')">Send ▶</button>
        </div>
      </div>
    </div>
  `;

  document.body.appendChild(overlay);
  cwModalLoadMessages(sessionID);
}

function cwCollapseModal(sessionID) {
  const overlay = document.getElementById(`cwmodal-${sessionID}`);
  if (overlay) overlay.style.display = 'none';
  // Restore the corner window
  const win = chatWindows.get(sessionID);
  if (win && win.minimized) cwToggleMinimize(sessionID);
}

function cwDestroyModal(sessionID) {
  const sse = modalSSE.get(sessionID);
  if (sse) { sse.close(); modalSSE.delete(sessionID); }
  const overlay = document.getElementById(`cwmodal-${sessionID}`);
  if (overlay) overlay.remove();
  cwClose(sessionID); // also close the corner window
}

async function cwModalLoadMessages(sessionID) {
  const msgsEl = document.getElementById(`cwmodalmsgs-${sessionID}`);
  if (!msgsEl) return;
  try {
    const resp = await fetch(`/api/sessions/${sessionID}/messages`);
    if (!resp.ok) throw new Error('fetch failed');
    const data = await resp.json();
    const messages = data.messages || [];
    msgsEl.innerHTML = '';
    if (messages.length === 0) {
      msgsEl.innerHTML = '<div class="cw-empty">Send a message to start.</div>';
    } else {
      messages.forEach(msg => cwModalAppendMessage(sessionID, msg.Role || msg.role, msg.Content || msg.content));
    }
    cwModalScrollBottom(sessionID);
  } catch {
    msgsEl.innerHTML = '<div class="cw-empty">Could not load messages.</div>';
  }
}

function cwModalAppendMessage(sessionID, role, content) {
  const msgsEl = document.getElementById(`cwmodalmsgs-${sessionID}`);
  if (!msgsEl) return;
  const empty = msgsEl.querySelector('.cw-empty, .cw-loading');
  if (empty) empty.remove();
  const div = document.createElement('div');
  div.className = `cw-message cw-msg-${role}`;
  div.innerHTML = `<span class="cw-msg-role">${escapeHTML(role)}</span><div class="cw-msg-content">${escapeHTML(content)}</div>`;
  msgsEl.appendChild(div);
}

async function cwModalSend(sessionID) {
  const input = document.getElementById(`cwmodalinput-${sessionID}`);
  if (!input) return;
  const message = input.value.trim();
  if (!message) return;

  input.value = '';
  cwModalAppendMessage(sessionID, 'user', message);
  cwModalScrollBottom(sessionID);

  // Streaming placeholder
  const msgsEl = document.getElementById(`cwmodalmsgs-${sessionID}`);
  const streamEl = document.createElement('div');
  streamEl.className = 'cw-message cw-msg-assistant';
  streamEl.id = `cwmodalstream-${sessionID}`;
  streamEl.innerHTML = `
    <span class="cw-msg-role">assistant</span>
    <div class="cw-msg-content" id="cwmodalstreamtext-${sessionID}"></div>
    <div class="cw-blink"><span></span><span></span><span></span></div>
  `;
  if (msgsEl) msgsEl.appendChild(streamEl);
  cwModalScrollBottom(sessionID);

  const stopBtn = document.getElementById(`cwmodalstop-${sessionID}`);
  if (stopBtn) stopBtn.classList.remove('hidden');

  try {
    const form = new FormData();
    form.append('session_id', sessionID);
    form.append('message', message);
    const resp = await fetch('/api/chat/send', { method: 'POST', body: form });
    if (!resp.ok) throw new Error('send failed');
    const data = await resp.json();
    cwModalOpenSSE(sessionID, data.stream_url);
  } catch (err) {
    const textEl = document.getElementById(`cwmodalstreamtext-${sessionID}`);
    if (textEl) { textEl.style.color = 'var(--error)'; textEl.textContent = '⚠ ' + err.message; }
    cwModalFinalise(sessionID);
  }
}

function cwModalOpenSSE(sessionID, url) {
  const existing = modalSSE.get(sessionID);
  if (existing) existing.close();

  const es = new EventSource(url);
  modalSSE.set(sessionID, es);

  es.addEventListener('message_delta', e => {
    const payload = safeJSON(e.data);
    if (!payload) return;
    const textEl = document.getElementById(`cwmodalstreamtext-${sessionID}`);
    if (textEl) textEl.textContent += (payload.content || '');
    cwModalScrollBottom(sessionID);
  });

  es.addEventListener('message_end', () => { cwModalFinalise(sessionID); es.close(); modalSSE.delete(sessionID); });
  es.addEventListener('error',       () => { cwModalFinalise(sessionID); es.close(); modalSSE.delete(sessionID); });
}

function cwModalFinalise(sessionID) {
  const streamEl = document.getElementById(`cwmodalstream-${sessionID}`);
  if (streamEl) {
    streamEl.id = '';
    const blink = streamEl.querySelector('.cw-blink');
    if (blink) blink.remove();
    const textEl = document.getElementById(`cwmodalstreamtext-${sessionID}`);
    if (textEl) textEl.id = '';
  }
  const stopBtn = document.getElementById(`cwmodalstop-${sessionID}`);
  if (stopBtn) stopBtn.classList.add('hidden');
}

function cwModalStop(sessionID) {
  const sse = modalSSE.get(sessionID);
  if (sse) { sse.close(); modalSSE.delete(sessionID); }
  cwModalFinalise(sessionID);
  const form = new FormData();
  form.append('session_id', sessionID);
  fetch('/api/chat/stop', { method: 'POST', body: form }).catch(() => {});
}

function cwModalKeyDown(e, sessionID) {
  if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); cwModalSend(sessionID); }
}

function cwModalScrollBottom(sessionID) {
  const msgsEl = document.getElementById(`cwmodalmsgs-${sessionID}`);
  if (msgsEl) msgsEl.scrollTop = msgsEl.scrollHeight;
}

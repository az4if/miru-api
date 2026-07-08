package docs

import (
	"github.com/gofiber/fiber/v2"
)

const swaggerHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>Miru API — Docs</title>
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
<style>body{margin:0;background:#0b0d12} .topbar{display:none}</style>
</head>
<body>
<div id="ui"></div>
<script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
<script>
window.onload = () => {
  SwaggerUIBundle({
    url: "/openapi.json",
    dom_id: "#ui",
    deepLinking: true,
    persistAuthorization: true,
    tryItOutEnabled: true,
    docExpansion: "list",
    defaultModelsExpandDepth: -1,
  });
};
</script>
</body></html>`

const demoHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1,viewport-fit=cover"/>
<title>Miru — Player</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
<style>
:root{
  color-scheme:dark;
  --bg:#07090d; --panel:#0e131c; --panel-2:#141b27; --line:#1f2632;
  --text:#e7eaf3; --mute:#9aa4b2; --dim:#5a6473;
  --brand:#7c5cff; --brand-2:#5b8cff; --accent:#22d3ee;
  --good:#22c55e; --bad:#ef4444; --warn:#f59e0b;
  --r:14px;
}
*{box-sizing:border-box}
html,body{height:100%}
body{
  margin:0;background:radial-gradient(1200px 600px at 80% -10%,rgba(124,92,255,.18),transparent 60%),
       radial-gradient(900px 500px at -10% 110%,rgba(34,211,238,.10),transparent 60%),var(--bg);
  color:var(--text);font:14px/1.5 'Inter',ui-sans-serif,system-ui,-apple-system,'Segoe UI',Roboto,sans-serif;
  -webkit-font-smoothing:antialiased;
}
.shell{max-width:1180px;margin:0 auto;padding:24px 20px 80px}
.head{display:flex;align-items:center;gap:14px;margin-bottom:20px}
.logo{
  width:38px;height:38px;border-radius:11px;
  background:conic-gradient(from 220deg,var(--brand),var(--brand-2),var(--accent),var(--brand));
  box-shadow:0 8px 28px rgba(124,92,255,.35),inset 0 0 0 1px rgba(255,255,255,.08);
  position:relative;
}
.logo::after{content:"";position:absolute;inset:9px;border-radius:6px;background:#07090d}
.brand{font-weight:700;letter-spacing:.2px}
.brand small{display:block;color:var(--mute);font-weight:500;font-size:11px;letter-spacing:.06em;text-transform:uppercase;margin-top:2px}
.spacer{flex:1}
.tag{display:inline-flex;align-items:center;gap:6px;padding:6px 10px;border-radius:999px;background:#0f1623;border:1px solid var(--line);font-size:12px;color:var(--mute)}
.dot{width:7px;height:7px;border-radius:50%;background:var(--good);box-shadow:0 0 12px var(--good)}

/* Search bar */
.bar{display:grid;grid-template-columns:1fr 110px auto;gap:10px;background:var(--panel);border:1px solid var(--line);border-radius:var(--r);padding:10px;margin-bottom:14px}
.bar input,.bar button{height:42px;border-radius:10px;border:1px solid var(--line);background:#0b1220;color:var(--text);padding:0 12px;font:inherit;outline:none}
.bar input:focus{border-color:var(--brand);box-shadow:0 0 0 3px rgba(124,92,255,.18)}
.bar button{background:linear-gradient(180deg,var(--brand),#5a3df0);border:0;color:#fff;font-weight:600;cursor:pointer;padding:0 18px;display:inline-flex;align-items:center;gap:8px;transition:transform .08s ease, filter .15s}
.bar button:hover{filter:brightness(1.08)}
.bar button:active{transform:translateY(1px)}

/* Toolbar (chips) */
.tools{display:flex;flex-wrap:wrap;gap:14px;margin-bottom:14px}
.group{display:flex;align-items:center;gap:8px;background:var(--panel);border:1px solid var(--line);border-radius:999px;padding:6px}
.group .label{padding:0 10px;color:var(--mute);font-size:11px;text-transform:uppercase;letter-spacing:.08em}
.chip{padding:7px 13px;border-radius:999px;border:0;background:transparent;color:var(--mute);font-weight:500;cursor:pointer;font:inherit;transition:background .15s, color .15s}
.chip:hover{color:var(--text)}
.chip.active{background:linear-gradient(180deg,var(--brand),#5a3df0);color:#fff;box-shadow:0 6px 18px rgba(124,92,255,.35)}
.chip .lat{font-size:10px;color:rgba(255,255,255,.65);margin-left:6px}
.chip.bad{opacity:.45}

/* Player */
.player{
  position:relative;background:#000;border-radius:18px;overflow:hidden;
  border:1px solid var(--line);box-shadow:0 30px 80px -30px rgba(0,0,0,.7), 0 0 0 1px rgba(255,255,255,.02) inset;
  aspect-ratio:16/9;
}
.player video{width:100%;height:100%;display:block;background:#000}
.player.loading::after{
  content:"";position:absolute;inset:0;
  background:radial-gradient(120px 120px at 50% 50%,rgba(124,92,255,.25),transparent 70%);
  animation:pulse 1.6s ease-in-out infinite;pointer-events:none;
}
@keyframes pulse{0%,100%{opacity:.6}50%{opacity:1}}
.center-msg{
  position:absolute;inset:0;display:none;align-items:center;justify-content:center;
  color:var(--mute);font-weight:500;text-align:center;padding:20px;pointer-events:none;
}
.center-msg.show{display:flex}
/* Download progress pill — sits at the TOP of the player so it never overlaps
   the big play/pause button in the center. */
.dl-toast{
  position:absolute;left:50%;top:14px;transform:translateX(-50%);
  display:none;align-items:center;gap:10px;
  padding:8px 14px;border-radius:999px;
  background:rgba(10,14,22,.82);backdrop-filter:blur(10px);
  border:1px solid rgba(255,255,255,.14);
  color:#fff;font-size:13px;font-weight:500;
  box-shadow:0 10px 30px rgba(0,0,0,.45);
  max-width:calc(100% - 28px);white-space:nowrap;overflow:hidden;text-overflow:ellipsis;
  z-index:5;pointer-events:none;
}
.dl-toast.show{display:inline-flex}
.dl-toast .mini-spin{
  width:14px;height:14px;border-radius:50%;
  border:2px solid rgba(255,255,255,.18);border-top-color:var(--brand);
  animation:spin 1s linear infinite;flex:none;
}
.spinner{
  width:42px;height:42px;border-radius:50%;
  border:3px solid rgba(255,255,255,.1);border-top-color:var(--brand);
  animation:spin 1s linear infinite;margin:0 auto 10px;
}
@keyframes spin{to{transform:rotate(360deg)}}

/* Controls overlay */
.controls{
  position:absolute;left:0;right:0;bottom:0;padding:14px 16px 12px;
  background:linear-gradient(180deg,transparent,rgba(0,0,0,.85));
  opacity:0;transition:opacity .25s ease;
  display:flex;flex-direction:column;gap:10px;
  pointer-events:none;
}
.player.show-controls .controls,
.player:hover .controls,
.player.paused .controls{opacity:1;pointer-events:auto}

.scrub{position:relative;height:18px;display:flex;align-items:center;cursor:pointer}
.scrub .track{position:relative;height:4px;width:100%;background:rgba(255,255,255,.18);border-radius:999px;overflow:hidden;transition:height .12s}
.scrub:hover .track{height:6px}
.scrub .buffered{position:absolute;left:0;top:0;bottom:0;background:rgba(255,255,255,.32);border-radius:999px}
.scrub .filled{position:absolute;left:0;top:0;bottom:0;background:linear-gradient(90deg,var(--brand),var(--brand-2));border-radius:999px}
.scrub .knob{position:absolute;top:50%;width:14px;height:14px;border-radius:50%;background:#fff;transform:translate(-50%,-50%);box-shadow:0 0 0 4px rgba(124,92,255,.3);opacity:0;transition:opacity .15s}
.scrub:hover .knob{opacity:1}

.row-c{display:flex;align-items:center;gap:6px;color:#fff}
.row-c .grow{flex:1}
.icbtn{
  width:38px;height:38px;border-radius:10px;border:0;background:transparent;color:#fff;cursor:pointer;
  display:inline-flex;align-items:center;justify-content:center;transition:background .15s, transform .08s;
}
.icbtn:hover{background:rgba(255,255,255,.12)}
.icbtn:active{transform:scale(.94)}
.icbtn svg{width:20px;height:20px;display:block}
.time{font-variant-numeric:tabular-nums;color:rgba(255,255,255,.85);font-size:12px;padding:0 6px}
.vol{display:flex;align-items:center;gap:6px}
.vol input[type=range]{width:0;opacity:0;transition:width .2s, opacity .2s;height:4px;-webkit-appearance:none;background:rgba(255,255,255,.3);border-radius:999px}
.vol:hover input[type=range],.vol:focus-within input[type=range]{width:90px;opacity:1}
.vol input[type=range]::-webkit-slider-thumb{-webkit-appearance:none;width:12px;height:12px;background:#fff;border-radius:50%;cursor:pointer}

/* Big play */
.bigplay{
  position:absolute;left:50%;top:50%;transform:translate(-50%,-50%);
  width:84px;height:84px;border-radius:50%;border:0;cursor:pointer;color:#fff;
  background:rgba(0,0,0,.55);backdrop-filter:blur(8px);
  display:flex;align-items:center;justify-content:center;
  transition:transform .15s, opacity .2s;opacity:0;pointer-events:none;
}
.bigplay svg{width:38px;height:38px;margin-left:4px}
.player.paused:not(.loading) .bigplay{opacity:1;pointer-events:auto}
.bigplay:hover{transform:translate(-50%,-50%) scale(1.06)}

/* Skip buttons */
.skipbtn{
  position:absolute;right:18px;bottom:80px;
  display:none;align-items:center;gap:8px;padding:10px 16px;border-radius:999px;
  background:rgba(20,27,39,.92);backdrop-filter:blur(8px);
  border:1px solid rgba(255,255,255,.14);color:#fff;cursor:pointer;font:inherit;font-weight:600;
  box-shadow:0 12px 36px rgba(0,0,0,.5);
}
.skipbtn.show{display:inline-flex}
.skipbtn:hover{background:rgba(28,38,56,.95)}
.skipbtn svg{width:16px;height:16px}

/* Settings sheet — anchored to the bottom-right of the player and constrained
   to the player's height so it never spills outside the video area. */
.sheet{
  position:absolute;right:14px;bottom:64px;
  width:280px;max-width:calc(100% - 28px);
  max-height:calc(100% - 80px);
  background:rgba(14,19,28,.95);backdrop-filter:blur(14px);
  border:1px solid rgba(255,255,255,.08);border-radius:14px;
  box-shadow:0 24px 60px rgba(0,0,0,.55);
  overflow-y:auto;overflow-x:hidden;display:none;
  overscroll-behavior:contain;
  scrollbar-width:thin;scrollbar-color:rgba(255,255,255,.18) transparent;
}
.sheet::-webkit-scrollbar{width:8px}
.sheet::-webkit-scrollbar-thumb{background:rgba(255,255,255,.18);border-radius:8px}
.sheet::-webkit-scrollbar-track{background:transparent}
.sheet.open{display:block}
.sheet h4{margin:0;padding:12px 14px 6px;color:var(--mute);font-size:11px;letter-spacing:.08em;text-transform:uppercase;font-weight:600;position:sticky;top:0;background:rgba(14,19,28,.95);backdrop-filter:blur(14px);z-index:1}
.sheet h4:first-child{padding-top:14px}
.sheet .item{
  display:flex;align-items:center;gap:10px;padding:10px 14px;color:#fff;cursor:pointer;
  border:0;background:transparent;width:100%;text-align:left;font:inherit;transition:background .12s;
}
.sheet .item:hover{background:rgba(255,255,255,.06)}
.sheet .item .meta{margin-left:auto;color:var(--mute);font-size:12px}
.sheet .item.active{color:#fff}
.sheet .item.active .meta{color:var(--brand)}
.sheet .check{width:14px;display:inline-block;color:var(--brand)}
.sheet hr{border:0;border-top:1px solid rgba(255,255,255,.06);margin:6px 0}

/* Download modal */
.dl-modal{position:fixed;inset:0;display:none;align-items:center;justify-content:center;background:rgba(4,6,10,.7);backdrop-filter:blur(6px);z-index:50;padding:20px}
.dl-modal.open{display:flex}
.dl-card{width:100%;max-width:760px;max-height:86vh;display:flex;flex-direction:column;background:var(--panel);border:1px solid var(--line);border-radius:18px;box-shadow:0 30px 80px rgba(0,0,0,.6);overflow:hidden}
.dl-head{display:flex;align-items:center;gap:12px;padding:18px 20px 12px;border-bottom:1px solid var(--line);flex-wrap:wrap}
.dl-title{font-weight:700;font-size:17px;color:var(--text)}
.dl-sub{color:var(--mute);font-size:12px;margin-top:2px}
/* Source picker — Pahe (default, per-episode MP4) vs Tosho (p2p/DDL).
   Visually a segmented control so it's obvious you can flip between them
   in one click without losing the modal context. */
.dl-source-tabs{display:inline-flex;background:#0b1220;border:1px solid var(--line);border-radius:999px;padding:3px;gap:2px}
.dl-source-tabs .src-tab{display:inline-flex;align-items:center;gap:6px;padding:6px 14px;border-radius:999px;border:0;background:transparent;color:var(--mute);font-size:12px;font-weight:600;cursor:pointer;transition:all .15s;font-family:inherit}
.dl-source-tabs .src-tab:hover{color:var(--text)}
.dl-source-tabs .src-tab.active{background:linear-gradient(180deg,var(--brand),#5a3df0);color:#fff;box-shadow:0 4px 14px rgba(124,92,255,.35)}
.dl-source-tabs .src-tab .src-dot{width:6px;height:6px;border-radius:50%;background:currentColor;opacity:.7}
.dl-source-tabs .src-tab.active .src-dot{opacity:1}
.dl-source-tabs .src-tab.loading{opacity:.6;pointer-events:none}
.dl-filters{display:flex;flex-wrap:wrap;gap:8px;padding:12px 20px;border-bottom:1px solid var(--line);background:rgba(255,255,255,.02)}
.dl-filters .fchip{padding:6px 12px;border-radius:999px;background:#0b1220;border:1px solid var(--line);color:var(--mute);font-size:12px;font-weight:500;cursor:pointer;transition:all .12s}
.dl-filters .fchip:hover{color:var(--text);border-color:var(--brand)}
.dl-filters .fchip.active{background:linear-gradient(180deg,var(--brand),#5a3df0);color:#fff;border-color:transparent}
.dl-body{flex:1;overflow-y:auto;padding:8px 20px 20px}
.dl-empty{padding:40px 20px;text-align:center;color:var(--mute)}
.dl-loading{padding:40px 20px;text-align:center;color:var(--mute)}
.dl-loading .spinner{margin:0 auto 10px}
.dl-group{margin-top:14px}
.dl-group-h{display:flex;align-items:center;gap:10px;padding:8px 0;color:var(--text);font-weight:600;font-size:13px}
.dl-group-h .badge{padding:2px 8px;border-radius:999px;background:rgba(124,92,255,.15);color:var(--brand);font-size:10px;font-weight:600;text-transform:uppercase;letter-spacing:.06em}
.dl-row{display:grid;grid-template-columns:1fr auto;gap:12px;align-items:center;padding:12px;background:#0b1220;border:1px solid var(--line);border-radius:10px;margin-bottom:8px;transition:border-color .12s}
.dl-row:hover{border-color:rgba(124,92,255,.4)}
.dl-row .meta-line{color:var(--text);font-size:13px;font-weight:500;line-height:1.35;word-break:break-word}
.dl-row .sub-line{color:var(--mute);font-size:11px;margin-top:4px;display:flex;flex-wrap:wrap;gap:8px}
.dl-row .sub-line span{display:inline-flex;align-items:center;gap:4px}
.dl-row .qbadge{display:inline-block;padding:1px 6px;border-radius:4px;background:rgba(34,211,238,.12);color:var(--accent);font-weight:600}
.dl-row .gbadge{display:inline-block;padding:1px 6px;border-radius:4px;background:rgba(124,92,255,.12);color:var(--brand);font-weight:600}
.dl-actions{display:flex;gap:6px;flex-wrap:wrap;justify-content:flex-end}
.dl-btn{padding:7px 12px;border-radius:8px;border:1px solid var(--line);background:#0e131c;color:var(--text);font-size:12px;font-weight:600;cursor:pointer;text-decoration:none;display:inline-flex;align-items:center;gap:5px;transition:all .12s}
.dl-btn:hover{border-color:var(--brand);color:#fff}
.dl-btn.primary{background:linear-gradient(180deg,var(--brand),#5a3df0);border-color:transparent;color:#fff}
.dl-btn.primary:hover{filter:brightness(1.08)}
.dl-btn svg{width:13px;height:13px}

/* Status / log */
.statbar{margin-top:14px;display:flex;flex-wrap:wrap;align-items:center;gap:10px;color:var(--mute);font-size:13px}
.statbar .pill{padding:4px 10px;border-radius:999px;background:#0f1623;border:1px solid var(--line);color:var(--text);font-size:12px}
.statbar .pill.ok{border-color:rgba(34,197,94,.45);color:#9af0b8}
.statbar .pill.err{border-color:rgba(239,68,68,.5);color:#ffb6b6}
details{margin-top:14px;background:var(--panel);border:1px solid var(--line);border-radius:12px}
details summary{cursor:pointer;padding:10px 14px;color:var(--mute);font-size:12px;text-transform:uppercase;letter-spacing:.06em;list-style:none}
details summary::-webkit-details-marker{display:none}
details pre{margin:0;padding:0 14px 14px;color:#cdd5e0;font-size:12px;max-height:300px;overflow:auto}

/* Subtitles look — Netflix-style */
::cue{
  background:rgba(0,0,0,.55);color:#fff;font-family:Inter,system-ui,sans-serif;
  font-size:.95em;text-shadow:0 1px 2px rgba(0,0,0,.7);line-height:1.3;padding:.1em .4em;
}

/* Mobile */
@media (max-width:680px){
  .shell{padding:14px 12px 60px}
  .bar{grid-template-columns:1fr 80px auto}
  .tools{gap:8px}
  .group{padding:4px}
  .chip{padding:6px 10px;font-size:13px}
  .icbtn{width:34px;height:34px}
  .bigplay{width:64px;height:64px}
  .sheet{right:8px;bottom:58px;width:240px}
}
</style>
</head>
<body>
<div class="shell">
  <div class="head">
    <div class="logo"></div>
    <div class="brand">Miru Player <small>HLS · Soft subs · Auto server</small></div>
    <div class="spacer"></div>
    <a class="tag" href="/docs"><span class="dot"></span>API · v1.0.0</a>
  </div>

  <div class="bar">
    <input id="id" placeholder="Anime ID" value="6989bcdf29cf95f4eb03e9f5" autocomplete="off"/>
    <input id="ep" placeholder="Episode" value="1" autocomplete="off"/>
    <button id="go">
      <svg viewBox="0 0 24 24" fill="currentColor" width="18" height="18"><path d="M8 5v14l11-7z"/></svg>
      Play
    </button>
  </div>

  <div class="tools">
    <div class="group" id="typeGrp">
      <span class="label">Audio</span>
      <button class="chip active" data-type="sub">Sub</button>
      <button class="chip" data-type="dub">Dub</button>
    </div>
    <div class="group" id="serverGrp">
      <span class="label">Server</span>
      <button class="chip active" data-server="auto">Auto</button>
    </div>
    <div class="group" style="margin-left:auto">
      <button class="chip" id="probe" title="Probe servers">Probe servers</button>
      <button class="chip" id="dl" title="Show real downloadable releases for this episode">Download</button>
    </div>
  </div>

  <div class="dl-modal" id="dlModal" aria-hidden="true">
    <div class="dl-card">
      <div class="dl-head">
        <div style="flex:1;min-width:0">
          <div class="dl-title" id="dlTitle">Downloads</div>
          <div class="dl-sub" id="dlSub">Pick a release to download.</div>
        </div>
        <div class="dl-source-tabs" id="dlSourceTabs" role="tablist" aria-label="Download source">
          <button class="src-tab active" data-src="pahe" role="tab" aria-selected="true" title="Per-episode MP4s from animepahe (small files, 1-click download)">
            <span class="src-dot"></span>Pahe
          </button>
          <button class="src-tab" data-src="tosho" role="tab" aria-selected="false" title="P2P / DDL releases from Anime Tosho (full-quality, every group)">
            <span class="src-dot"></span>Tosho
          </button>
        </div>
        <button class="icbtn" id="dlClose" aria-label="Close">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18M6 6l12 12"/></svg>
        </button>
      </div>
      <div class="dl-filters" id="dlFilters"></div>
      <div class="dl-body" id="dlBody"></div>
    </div>
  </div>

  <div class="player" id="player">
    <video id="v" playsinline crossorigin="anonymous" preload="metadata"></video>
    <div class="center-msg" id="centerMsg"></div>
    <div class="dl-toast" id="dlToast"></div>
    <button class="bigplay" id="bigplay" aria-label="Play">
      <svg viewBox="0 0 24 24" fill="currentColor"><path d="M8 5v14l11-7z"/></svg>
    </button>
    <button class="skipbtn" id="skipIntro">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><polygon points="5 4 15 12 5 20 5 4"/><line x1="19" y1="5" x2="19" y2="19"/></svg>
      Skip intro
    </button>
    <button class="skipbtn" id="skipOutro" style="bottom:130px">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><polygon points="5 4 15 12 5 20 5 4"/><line x1="19" y1="5" x2="19" y2="19"/></svg>
      Next episode
    </button>

    <div class="sheet" id="settings"></div>

    <div class="controls">
      <div class="scrub" id="scrub">
        <div class="track">
          <div class="buffered" id="buf"></div>
          <div class="filled" id="fill"></div>
        </div>
        <div class="knob" id="knob"></div>
      </div>
      <div class="row-c">
        <button class="icbtn" id="playBtn" aria-label="Play/Pause">
          <svg id="pIcon" viewBox="0 0 24 24" fill="currentColor"><path d="M8 5v14l11-7z"/></svg>
        </button>
        <button class="icbtn" id="back10" aria-label="Back 10s">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 12a9 9 0 1 0 3-6.7L3 8"/><path d="M3 3v5h5"/><text x="12" y="15.5" text-anchor="middle" font-size="7" font-weight="700" fill="currentColor" stroke="none">10</text></svg>
        </button>
        <button class="icbtn" id="fwd10" aria-label="Forward 10s">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12a9 9 0 1 1-3-6.7L21 8"/><path d="M21 3v5h-5"/><text x="12" y="15.5" text-anchor="middle" font-size="7" font-weight="700" fill="currentColor" stroke="none">10</text></svg>
        </button>
        <div class="vol">
          <button class="icbtn" id="muteBtn" aria-label="Mute">
            <svg id="mIcon" viewBox="0 0 24 24" fill="currentColor"><path d="M3 10v4h4l5 5V5L7 10H3zm13.5 2c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02zM14 3.23v2.06c2.89.86 5 3.54 5 6.71s-2.11 5.85-5 6.71v2.06c4.01-.91 7-4.49 7-8.77s-2.99-7.86-7-8.77z"/></svg>
          </button>
          <input type="range" id="volRange" min="0" max="1" step="0.01" value="1"/>
        </div>
        <span class="time" id="time">0:00 / 0:00</span>
        <div class="grow"></div>
        <button class="icbtn" id="ccBtn" aria-label="Subtitles" title="Subtitles">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="5" width="18" height="14" rx="3"/><path d="M7 13c.7 1 1.7 1.5 2.8 1.5"/><path d="M14 13c.7 1 1.7 1.5 2.8 1.5"/></svg>
        </button>
        <button class="icbtn" id="setBtn" aria-label="Settings" title="Settings">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.7 1.7 0 0 0 .3 1.8l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-1.8-.3 1.7 1.7 0 0 0-1 1.5V21a2 2 0 1 1-4 0v-.1a1.7 1.7 0 0 0-1.1-1.5 1.7 1.7 0 0 0-1.8.3l-.1.1A2 2 0 1 1 4.3 17l.1-.1a1.7 1.7 0 0 0 .3-1.8 1.7 1.7 0 0 0-1.5-1H3a2 2 0 1 1 0-4h.1a1.7 1.7 0 0 0 1.5-1.1 1.7 1.7 0 0 0-.3-1.8l-.1-.1A2 2 0 1 1 7 4.3l.1.1a1.7 1.7 0 0 0 1.8.3H9a1.7 1.7 0 0 0 1-1.5V3a2 2 0 1 1 4 0v.1a1.7 1.7 0 0 0 1 1.5 1.7 1.7 0 0 0 1.8-.3l.1-.1A2 2 0 1 1 19.7 7l-.1.1a1.7 1.7 0 0 0-.3 1.8V9a1.7 1.7 0 0 0 1.5 1H21a2 2 0 1 1 0 4h-.1a1.7 1.7 0 0 0-1.5 1z"/></svg>
        </button>
        <button class="icbtn" id="pipBtn" aria-label="PiP" title="Picture-in-picture">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="5" width="18" height="14" rx="2"/><rect x="12" y="11" width="7" height="6" rx="1" fill="currentColor"/></svg>
        </button>
        <button class="icbtn" id="fsBtn" aria-label="Fullscreen" title="Fullscreen">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 9V5h4"/><path d="M20 9V5h-4"/><path d="M4 15v4h4"/><path d="M20 15v4h-4"/></svg>
        </button>
      </div>
    </div>
  </div>

  <div class="statbar">
    <span id="statText">Ready.</span>
    <span class="pill" id="srvPill" style="display:none"></span>
    <span class="pill" id="qPill" style="display:none"></span>
    <span class="pill" id="subPill" style="display:none"></span>
  </div>

  <details><summary>Debug log</summary><pre id="log"></pre></details>
</div>

<script src="https://cdn.jsdelivr.net/npm/hls.js@1.5.17/dist/hls.min.js"></script>
<script>
const $ = id => document.getElementById(id);
const v = $("v"), player = $("player"), settings = $("settings");
const SVG_PLAY = '<svg viewBox="0 0 24 24" fill="currentColor"><path d="M8 5v14l11-7z"/></svg>';
const SVG_PAUSE = '<svg viewBox="0 0 24 24" fill="currentColor"><path d="M6 5h4v14H6zM14 5h4v14h-4z"/></svg>';
const SVG_VOL = '<svg viewBox="0 0 24 24" fill="currentColor"><path d="M3 10v4h4l5 5V5L7 10H3zm13.5 2c0-1.77-1.02-3.29-2.5-4.03v8.05c1.48-.73 2.5-2.25 2.5-4.02zM14 3.23v2.06c2.89.86 5 3.54 5 6.71s-2.11 5.85-5 6.71v2.06c4.01-.91 7-4.49 7-8.77s-2.99-7.86-7-8.77z"/></svg>';
const SVG_MUTE = '<svg viewBox="0 0 24 24" fill="currentColor"><path d="M16.5 12c0-1.77-1.02-3.29-2.5-4.03v2.21l2.45 2.45c.03-.2.05-.41.05-.63zM19 12c0 .94-.2 1.82-.54 2.64l1.51 1.51A8.796 8.796 0 0 0 21 12c0-4.28-2.99-7.86-7-8.77v2.06c2.89.86 5 3.54 5 6.71zM4.27 3L3 4.27 7.73 9H3v6h4l5 5v-6.73l4.25 4.25c-.67.52-1.42.93-2.25 1.18v2.06c1.38-.31 2.63-.95 3.69-1.81L19.73 21 21 19.73l-9-9L4.27 3zM12 4L9.91 6.09 12 8.18V4z"/></svg>';

let hls = null, recoverCount = 0, lastRecover = 0;
let watchData = null;     // last /api/watch response
let allSources = [];      // [{quality,url,proxy_url,type}]
let currentQuality = 'auto';
let subTracks = [];       // [{label,url,lang,default}]
let activeSubIdx = -1;
let availableServers = [{id:'auto',tip:'Auto (best)'}];
let currentServer = 'auto';
let currentType = 'sub';
let skipIntroShown = false, skipOutroShown = false;

const log = (...x) => {
  const s = x.map(z => typeof z==='string'?z:JSON.stringify(z,null,2)).join(' ');
  $("log").textContent = s + '\n\n' + $("log").textContent.slice(0, 12000);
};
const setStat = t => $("statText").textContent = t;
const showCenter = (html) => { const m = $("centerMsg"); if (!html) { m.classList.remove('show'); m.innerHTML=''; } else { m.innerHTML = html; m.classList.add('show'); } };
const fmt = s => { if (!isFinite(s) || s<0) return '0:00'; s=Math.floor(s); const h=Math.floor(s/3600), m=Math.floor((s%3600)/60), x=s%60; return (h?h+':':'')+(h?String(m).padStart(2,'0'):m)+':'+String(x).padStart(2,'0'); };

// ---- Pill groups -------------------------------------------------------
function bindGroup(groupId, attr, onChange) {
  const g = $(groupId);
  g.addEventListener('click', e => {
    const b = e.target.closest('button.chip'); if (!b || !b.hasAttribute('data-'+attr)) return;
    [...g.querySelectorAll('button.chip[data-'+attr+']')].forEach(x => x.classList.remove('active'));
    b.classList.add('active');
    onChange(b.getAttribute('data-'+attr) || '');
  });
}
bindGroup('typeGrp','type', v => { currentType = v; renderServerChips(); });
bindGroup('serverGrp','server', v => { currentServer = v; });

function renderServerChips() {
  const g = $("serverGrp");
  // remove existing dynamic chips
  [...g.querySelectorAll('button.chip[data-server]')].forEach(b => b.remove());
  const make = (id, label, latency) => {
    const b = document.createElement('button');
    b.className = 'chip' + (id === currentServer ? ' active' : '');
    b.setAttribute('data-server', id);
    b.textContent = label;
    if (typeof latency === 'number') {
      const l = document.createElement('span'); l.className='lat'; l.textContent = latency+'ms';
      b.appendChild(l);
    }
    g.appendChild(b);
  };
  make('auto', 'Auto');
  for (const s of availableServers) {
    if (s.id === 'auto') continue;
    make(s.id, s.id);
  }
}
renderServerChips();

// ---- HLS player --------------------------------------------------------
function destroyHls(){ if (hls) { try{hls.destroy();}catch(e){} hls=null; } }

function attachStream(url, type) {
  destroyHls();
  // Remove old <track> elements (we'll add via TextTrack API after manifest).
  [...v.querySelectorAll('track')].forEach(t => t.remove());

  const isHls = (type && /mpegurl/i.test(type)) || /\.m3u8(\?|$)/i.test(url);
  player.classList.add('loading');
  showCenter('<div><div class="spinner"></div>Loading stream…</div>');

  if (!isHls) {
    v.src = url;
    bindMediaReady();
    v.play().catch(()=>{});
    return;
  }
  if (window.Hls && Hls.isSupported()) {
    recoverCount = 0;
    hls = new Hls({
      enableWorker:false, lowLatencyMode:false, backBufferLength:60,
      maxBufferLength:30, maxMaxBufferLength:60,
      fragLoadingMaxRetry:6, manifestLoadingMaxRetry:4, levelLoadingMaxRetry:4,
    });
    hls.loadSource(url); hls.attachMedia(v);
    hls.on(Hls.Events.MANIFEST_PARSED, () => {
      bindMediaReady();
      attachSubtitles();
      v.play().catch(()=>{});
      buildSettingsSheet();
    });
    hls.on(Hls.Events.LEVEL_SWITCHED, (_, d) => buildSettingsSheet());
    hls.on(Hls.Events.ERROR, (_, d) => {
      log('hls', d.type, d.details);
      if (!d.fatal) return;
      const now = performance.now();
      if (now - lastRecover < 1500) recoverCount++; else recoverCount = 1;
      lastRecover = now;
      if (recoverCount > 4) { setStat('Playback failed. Try another server.'); destroyHls(); player.classList.remove('loading'); showCenter('Playback failed. Pick a different server above.'); return; }
      if (d.type === Hls.ErrorTypes.NETWORK_ERROR) { try{hls.startLoad();}catch(e){} }
      else if (d.type === Hls.ErrorTypes.MEDIA_ERROR) { try{hls.recoverMediaError();}catch(e){} }
      else { destroyHls(); }
    });
  } else if (v.canPlayType('application/vnd.apple.mpegurl')) {
    v.src = url; bindMediaReady(); v.play().catch(()=>{});
    // Native — add subs as <track>
    setTimeout(attachSubtitlesNative, 80);
  } else {
    showCenter('Browser cannot play HLS.');
  }
}

function bindMediaReady() {
  v.addEventListener('loadedmetadata', () => {
    player.classList.remove('loading'); showCenter('');
  }, { once:true });
}

// ---- Subtitles (TextTrack API; works around hls.js + <track> CORS quirks)
async function attachSubtitles() {
  // Clear existing app text tracks (browsers expose them as readonly,
  // but disabling all then adding new addTextTrack instances is supported).
  for (let i = 0; i < v.textTracks.length; i++) v.textTracks[i].mode = 'disabled';
  if (!subTracks.length) { $("subPill").style.display='none'; activeSubIdx=-1; buildSettingsSheet(); return; }

  // Fetch each VTT through our subtitle proxy and inject as cues, so we
  // bypass any browser cross-origin quirks with <track src>.
  for (const t of subTracks) {
    if (t._tt) continue;
    try {
      const r = await fetch(t.url, { credentials:'omit' });
      const txt = await r.text();
      const tt = v.addTextTrack('subtitles', t.label, t.lang || 'en');
      parseVTT(txt).forEach(cue => { try { tt.addCue(cue); } catch(e){} });
      t._tt = tt;
    } catch(e) { log('sub fail', t.label, e.message); }
  }
  // Auto-enable default (or first) track
  let pickIdx = subTracks.findIndex(t => t.default);
  if (pickIdx < 0) pickIdx = 0;
  setActiveSub(pickIdx);
  buildSettingsSheet();
}

function attachSubtitlesNative() {
  // For iOS Safari etc., let <track> handle it.
  for (const t of subTracks) {
    const el = document.createElement('track');
    el.kind = 'subtitles'; el.label = t.label; el.srclang = t.lang || 'en'; el.src = t.url;
    if (t.default) el.default = true;
    v.appendChild(el);
  }
}

function parseVTT(text) {
  // Minimal WEBVTT parser — handles HH:MM:SS.mmm and MM:SS.mmm timestamps
  // and multi-line cues. Skips NOTE blocks. Robust enough for upstream subs.
  const cues = [];
  const lines = text.replace(/\r/g,'').split('\n');
  let i = 0;
  // Skip header
  if (lines[i] && lines[i].startsWith('WEBVTT')) i++;
  while (i < lines.length) {
    while (i < lines.length && lines[i].trim() === '') i++;
    if (i >= lines.length) break;
    if (lines[i].startsWith('NOTE')) { while (i < lines.length && lines[i].trim() !== '') i++; continue; }
    let header = lines[i];
    if (!header.includes('-->')) { i++; if (i >= lines.length) break; header = lines[i]; }
    if (!header.includes('-->')) { i++; continue; }
    const m = header.match(/(\d{1,2}:)?(\d{1,2}):(\d{1,2})[.,](\d{1,3})\s*-->\s*(\d{1,2}:)?(\d{1,2}):(\d{1,2})[.,](\d{1,3})/);
    if (!m) { i++; continue; }
    const start = (parseInt(m[1]||'0',10)*3600) + parseInt(m[2],10)*60 + parseInt(m[3],10) + parseInt(m[4],10)/1000;
    const end   = (parseInt(m[5]||'0',10)*3600) + parseInt(m[6],10)*60 + parseInt(m[7],10) + parseInt(m[8],10)/1000;
    i++;
    const buf = [];
    while (i < lines.length && lines[i].trim() !== '') { buf.push(lines[i]); i++; }
    const text = buf.join('\n').replace(/<[^>]+>/g,'');
    if (text && end > start) cues.push(new VTTCue(start, end, text));
  }
  return cues;
}

function setActiveSub(idx) {
  activeSubIdx = idx;
  for (let i = 0; i < subTracks.length; i++) {
    if (subTracks[i]._tt) subTracks[i]._tt.mode = (i === idx) ? 'showing' : 'disabled';
  }
  // Native <track> path
  for (let i = 0; i < v.textTracks.length; i++) {
    const tt = v.textTracks[i];
    if (subTracks.some(t => t._tt === tt)) continue;
    tt.mode = (i === idx) ? 'showing' : 'disabled';
  }
  $("subPill").style.display = idx >= 0 ? '' : 'none';
  $("subPill").textContent = idx >= 0 ? ('CC: ' + (subTracks[idx]?.label||'')) : '';
}

// ---- Quality selection -------------------------------------------------
// Two paths:
//   1) Upstream returned multiple per-quality sources (the common case for
//      the kite/fsoft/pahe servers): switch by re-attaching a different
//      source URL while preserving the playhead, mute, volume and rate.
//   2) The chosen source is a master playlist with multiple levels:
//      use hls.currentLevel.
function uniqueQualityList() {
  // Returns [{key, label}] for the *source* list (variant URLs from upstream).
  // Filters out the synthetic "master" entry and de-duplicates by label.
  const seen = new Set();
  const out = [];
  for (const s of allSources) {
    if (!s || !s.quality) continue;
    if (/master/i.test(s.quality)) continue;
    const key = String(s.quality).toLowerCase();
    if (seen.has(key)) continue;
    seen.add(key);
    out.push({ key, label: s.quality });
  }
  // Sort high → low by extracted height
  const h = q => { const m = /(\d{3,4})/.exec(q); return m ? parseInt(m[1],10) : 0; };
  out.sort((a,b) => h(b.label) - h(a.label));
  return out;
}

function findSourceByQualityKey(key) {
  if (!key) return null;
  return allSources.find(s => String(s.quality).toLowerCase() === key) || null;
}

async function setQuality(q) {
  currentQuality = q;
  // hls.js master-playlist path
  if (hls && hls.levels && hls.levels.length > 1) {
    if (q === 'auto') { hls.currentLevel = -1; buildSettingsSheet(); return; }
    const idx = hls.levels.findIndex(L => String(L.height||L.bitrate) === String(q) || (L.height && (q+'p') === (L.height+'p')));
    if (idx >= 0) { hls.currentLevel = idx; buildSettingsSheet(); return; }
  }
  // Multi-source path: switch the underlying URL.
  if (q === 'auto') {
    const best = pickBest(allSources);
    if (best) reattachPreservingState(best);
    buildSettingsSheet();
    return;
  }
  const target = findSourceByQualityKey(q);
  if (!target) { setStat('Quality '+q+' is not available for this server.'); return; }
  reattachPreservingState(target);
  $("qPill").style.display=''; $("qPill").textContent = target.quality;
  buildSettingsSheet();
}

function reattachPreservingState(source) {
  const t = v.currentTime, paused = v.paused, vol = v.volume, muted = v.muted, rate = v.playbackRate;
  const onReady = () => {
    try { v.currentTime = t; } catch(e) {}
    v.volume = vol; v.muted = muted; v.playbackRate = rate;
    if (!paused) v.play().catch(()=>{});
    v.removeEventListener('loadedmetadata', onReady);
  };
  v.addEventListener('loadedmetadata', onReady);
  attachStream(source.proxy_url, source.type);
}

// ---- Settings sheet ---------------------------------------------------
function buildSettingsSheet() {
  const parts = [];
  const checkOn = '<span class="check">&#10003;</span>';
  const checkOff = '<span class="check"></span>';
  // Quality
  parts.push('<h4>Quality</h4>');
  const isAuto = currentQuality === 'auto';
  // Determine current display label
  let curLabel = '';
  if (hls && hls.levels && hls.levels.length > 1 && hls.currentLevel >= 0) {
    curLabel = (hls.levels[hls.currentLevel].height||'') + 'p';
  } else if (!isAuto) {
    const cur = findSourceByQualityKey(currentQuality);
    if (cur) curLabel = cur.quality;
  }
  parts.push('<button class="item '+(isAuto?'active':'')+'" data-act="q" data-v="auto">'+(isAuto?checkOn:checkOff)+'Auto<span class="meta">'+curLabel+'</span></button>');

  // Prefer multi-source list when available (the common case in this API).
  const sourceQs = uniqueQualityList();
  if (sourceQs.length > 1) {
    sourceQs.forEach(q => {
      const act = !isAuto && String(currentQuality) === q.key;
      parts.push('<button class="item '+(act?'active':'')+'" data-act="q" data-v="'+q.key+'">'+(act?checkOn:checkOff)+q.label+'</button>');
    });
  } else if (hls && hls.levels && hls.levels.length) {
    [...hls.levels].sort((a,b)=>(b.height||0)-(a.height||0)).forEach(L => {
      const label = L.height ? (L.height + 'p') : Math.round((L.bitrate||0)/1000)+'kbps';
      const v2 = L.height || L.bitrate;
      const act = !isAuto && String(currentQuality) === String(v2);
      parts.push('<button class="item '+(act?'active':'')+'" data-act="q" data-v="'+v2+'">'+(act?checkOn:checkOff)+label+'</button>');
    });
  }
  // Speed
  parts.push('<hr><h4>Playback speed</h4>');
  [0.5,0.75,1,1.25,1.5,2].forEach(sp => {
    const act = Math.abs(v.playbackRate - sp) < 0.01;
    const lbl = sp===1?'Normal':(sp+'x');
    parts.push('<button class="item '+(act?'active':'')+'" data-act="speed" data-v="'+sp+'">'+(act?checkOn:checkOff)+lbl+'</button>');
  });
  // Subtitles
  if (subTracks.length) {
    parts.push('<hr><h4>Subtitles</h4>');
    parts.push('<button class="item '+(activeSubIdx<0?'active':'')+'" data-act="sub" data-v="-1">'+(activeSubIdx<0?checkOn:checkOff)+'Off</button>');
    subTracks.forEach((t,i) => {
      const act = i === activeSubIdx;
      parts.push('<button class="item '+(act?'active':'')+'" data-act="sub" data-v="'+i+'">'+(act?checkOn:checkOff)+t.label+'</button>');
    });
  }
  settings.innerHTML = parts.join('');
}
settings.addEventListener('click', e => {
  const b = e.target.closest('button.item'); if (!b) return;
  const act = b.getAttribute('data-act'), val = b.getAttribute('data-v');
  if (act === 'q') setQuality(val);
  else if (act === 'speed') v.playbackRate = parseFloat(val);
  else if (act === 'sub') setActiveSub(parseInt(val,10));
  buildSettingsSheet();
});

// ---- Player controls --------------------------------------------------
$("playBtn").onclick = () => v.paused ? v.play() : v.pause();
$("bigplay").onclick = () => v.play();
$("back10").onclick = () => v.currentTime = Math.max(0, v.currentTime - 10);
$("fwd10").onclick = () => v.currentTime = Math.min(v.duration||0, v.currentTime + 10);
$("muteBtn").onclick = () => { v.muted = !v.muted; updateVolIcon(); };
$("volRange").addEventListener('input', e => { v.volume = parseFloat(e.target.value); v.muted = v.volume === 0; updateVolIcon(); });
$("ccBtn").onclick = () => {
  if (!subTracks.length) { setStat('No subtitles available for this episode.'); return; }
  setActiveSub(activeSubIdx >= 0 ? -1 : (subTracks.findIndex(t=>t.default) >= 0 ? subTracks.findIndex(t=>t.default) : 0));
  buildSettingsSheet();
};
$("setBtn").onclick = e => { e.stopPropagation(); settings.classList.toggle('open'); buildSettingsSheet(); };
document.addEventListener('click', e => { if (!settings.contains(e.target) && e.target.id !== 'setBtn') settings.classList.remove('open'); });

$("pipBtn").onclick = async () => { try { if (document.pictureInPictureElement) await document.exitPictureInPicture(); else await v.requestPictureInPicture(); } catch(e){ log('pip', e.message); } };
$("fsBtn").onclick = () => { if (document.fullscreenElement) document.exitFullscreen(); else player.requestFullscreen(); };
v.addEventListener('play',  () => { player.classList.remove('paused'); $("pIcon").outerHTML = SVG_PAUSE.replace('<svg', '<svg id="pIcon"'); });
v.addEventListener('pause', () => { player.classList.add('paused');    $("pIcon").outerHTML = SVG_PLAY.replace('<svg', '<svg id="pIcon"'); });
v.addEventListener('volumechange', updateVolIcon);
function updateVolIcon() { $("muteBtn").innerHTML = (v.muted || v.volume === 0) ? SVG_MUTE : SVG_VOL; $("volRange").value = v.muted ? 0 : v.volume; }
v.addEventListener('waiting', () => player.classList.add('loading'));
v.addEventListener('playing', () => { player.classList.remove('loading'); showCenter(''); });

// scrubber
const scrub = $("scrub"), fill = $("fill"), buf = $("buf"), knob = $("knob");
function pctOf(e) { const r = scrub.getBoundingClientRect(); return Math.min(1, Math.max(0, (e.clientX - r.left) / r.width)); }
let dragging = false;
scrub.addEventListener('pointerdown', e => { dragging = true; scrub.setPointerCapture(e.pointerId); seekTo(pctOf(e)); });
scrub.addEventListener('pointermove', e => { if (dragging) seekTo(pctOf(e)); else hoverKnob(pctOf(e)); });
scrub.addEventListener('pointerup',   e => { dragging = false; });
scrub.addEventListener('pointerleave',() => { if (!dragging) knob.style.left = (parseFloat(fill.style.width)||0)+'%'; });
function seekTo(p) { if (!v.duration) return; v.currentTime = p * v.duration; }
function hoverKnob(p) { knob.style.left = (p*100)+'%'; }
v.addEventListener('timeupdate', () => {
  const d = v.duration || 0, c = v.currentTime || 0;
  const p = d ? (c/d)*100 : 0;
  fill.style.width = p+'%'; if (!dragging) knob.style.left = p+'%';
  $("time").textContent = fmt(c)+' / '+fmt(d);
  // skip intro/outro
  const skips = watchData?.skips;
  if (skips?.intro && skips.intro.end > skips.intro.start && c >= skips.intro.start && c < skips.intro.end) {
    if (!skipIntroShown) { $("skipIntro").classList.add('show'); skipIntroShown = true; }
  } else if (skipIntroShown) { $("skipIntro").classList.remove('show'); skipIntroShown = false; }
});
v.addEventListener('progress', () => {
  if (!v.buffered.length || !v.duration) return;
  const e = v.buffered.end(v.buffered.length - 1);
  buf.style.width = ((e / v.duration) * 100) + '%';
});
$("skipIntro").onclick = () => { const s = watchData?.skips?.intro; if (s) v.currentTime = s.end + 0.1; };
$("skipOutro").onclick = () => { const ep = parseInt($("ep").value,10)||1; $("ep").value = ep+1; $("go").click(); };

// keyboard
document.addEventListener('keydown', e => {
  if (e.target.tagName === 'INPUT') return;
  if (e.code === 'Space') { e.preventDefault(); $("playBtn").click(); }
  else if (e.code === 'ArrowLeft')  v.currentTime = Math.max(0, v.currentTime - 5);
  else if (e.code === 'ArrowRight') v.currentTime = Math.min(v.duration||0, v.currentTime + 5);
  else if (e.code === 'ArrowUp')    { v.volume = Math.min(1, v.volume + 0.05); v.muted = false; updateVolIcon(); }
  else if (e.code === 'ArrowDown')  { v.volume = Math.max(0, v.volume - 0.05); updateVolIcon(); }
  else if (e.key === 'f') $("fsBtn").click();
  else if (e.key === 'm') $("muteBtn").click();
  else if (e.key === 'c') $("ccBtn").click();
});

// ---- Source picker -----------------------------------------------------
function pickBest(sources) {
  if (!sources.length) return null;
  return sources.find(s => /master/i.test(s.quality))
      || sources.find(s => /1080/.test(s.quality))
      || sources.find(s => /720/.test(s.quality))
      || sources[0];
}

// ---- Actions -----------------------------------------------------------
async function loadAndPlay() {
  const id = $("id").value.trim(), ep = $("ep").value.trim();
  if (!id) { setStat('Enter an anime ID first.'); return; }
  // Kick off server probe in parallel so chips populate with the real list ASAP.
  probeServers(true);
  setStat('Resolving sources…'); showCenter('<div><div class="spinner"></div>Finding the best server…</div>'); player.classList.add('loading');
  $("subPill").style.display='none'; $("srvPill").style.display='none'; $("qPill").style.display='none';
  const t0 = performance.now();
  try {
    const r = await fetch('/api/watch?id='+encodeURIComponent(id)+'&ep='+encodeURIComponent(ep)+'&server='+currentServer+'&source_type='+currentType).then(r=>r.json());
    const dt = Math.round(performance.now() - t0);
    log('watch '+dt+'ms', r);
    if (!r.success) { setStat('Error: ' + (r.error?.message||'unknown')); showCenter('Error: ' + (r.error?.message||'unknown')); player.classList.remove('loading'); return; }
    watchData = r.data;
    allSources = watchData.sources || [];
    subTracks = (watchData.subtitles || []).map(s => ({ label:s.label||s.lang||'Subtitle', url:s.url, lang:s.lang||'en', default:!!s.default }));
    const best = pickBest(allSources);
    if (!best) { setStat('No sources returned.'); showCenter('No sources returned.'); player.classList.remove('loading'); return; }
    setStat('Playing · resolved in '+dt+'ms');
    $("srvPill").style.display=''; $("srvPill").textContent = 'Server: '+watchData.server;
    $("srvPill").classList.add('ok');
    $("qPill").style.display=''; $("qPill").textContent = best.quality;
    attachStream(best.proxy_url, best.type);
    skipIntroShown = false; skipOutroShown = false;
  } catch (e) {
    setStat('Failed: ' + e.message); showCenter('Failed: ' + e.message); player.classList.remove('loading');
  }
}
$("go").onclick = loadAndPlay;
$("id").addEventListener('keydown', e => { if (e.key==='Enter') loadAndPlay(); });
$("ep").addEventListener('keydown', e => { if (e.key==='Enter') loadAndPlay(); });

async function probeServers(silent) {
  const id = $("id").value.trim(), ep = $("ep").value.trim();
  if (!id) return;
  if (!silent) setStat('Probing servers…');
  const t0 = performance.now();
  try {
    const r = await fetch('/api/anime/'+id+'/servers/'+ep+'?source_type='+currentType).then(r=>r.json());
    const dt = Math.round(performance.now() - t0);
    log('servers '+dt+'ms', r);
    if (!r.success) { if (!silent) setStat('Probe failed.'); return; }
    availableServers = (r.data||[]).map(s => ({ id:s.id, working:s.working, latency:s.latency_ms, default:s.default }));
    const g = $("serverGrp");
    [...g.querySelectorAll('button.chip[data-server]')].forEach(b => b.remove());
    const make = (id, label, latency, working) => {
      const b = document.createElement('button');
      b.className = 'chip' + (id === currentServer ? ' active' : '') + (working === false ? ' bad' : '');
      b.setAttribute('data-server', id);
      b.textContent = label;
      if (typeof latency === 'number') {
        const l = document.createElement('span'); l.className='lat'; l.textContent = latency+'ms';
        b.appendChild(l);
      }
      g.appendChild(b);
    };
    make('auto', 'Auto');
    for (const s of availableServers) make(s.id, s.id, s.latency, s.working);
    if (!silent) setStat('Probed in '+dt+'ms · '+availableServers.filter(s=>s.working).length+'/'+availableServers.length+' working');
  } catch (e) { if (!silent) setStat('Probe failed: '+e.message); }
}
$("probe").onclick = () => probeServers(false);

// ---- Real downloads ---------------------------------------------------
// Two sources, switchable in one click via the Pahe / Tosho tabs:
//   - Pahe  (default) → animepahe per-episode MP4s. Small (~30-100 MB),
//                       1-click download via pahe.win → kwik.cx.
//                       Matches the dropdown miru.live's website shows.
//   - Tosho           → Anime Tosho p2p / magnet / DDL mirrors. Full
//                       quality, every fansub group, every episode.
const dlModal = $("dlModal"), dlBody = $("dlBody"), dlFilters = $("dlFilters"), dlTitle = $("dlTitle"), dlSub = $("dlSub");
const dlSourceTabs = $("dlSourceTabs");
let dlState = {
  groups: [], flat: [],
  filter: { type: 'all', quality: 'all', group: 'all' },
  title: '', episode: '',
  source: 'pahe',          // currently displayed source
  currentSource: 'pahe',   // user's selected tab (default = pahe)
};

function openDownloads() {
  const id = $("id").value.trim(), ep = $("ep").value.trim();
  if (!id) { setStat('Enter an anime ID first.'); return; }
  dlModal.classList.add('open');
  dlModal.setAttribute('aria-hidden','false');
  dlTitle.textContent = 'Downloads';
  setSourceTab(dlState.currentSource);
  loadDownloadsForSource(id, ep, dlState.currentSource);
}
function closeDownloads() { dlModal.classList.remove('open'); dlModal.setAttribute('aria-hidden','true'); }
$("dl").onclick = openDownloads;
$("dlClose").onclick = closeDownloads;
dlModal.addEventListener('click', e => { if (e.target === dlModal) closeDownloads(); });
document.addEventListener('keydown', e => { if (e.key === 'Escape' && dlModal.classList.contains('open')) closeDownloads(); });

// One-click source switcher. Re-fetches with the new source and keeps the
// modal open so the flip feels instant.
dlSourceTabs.addEventListener('click', e => {
  const b = e.target.closest('.src-tab'); if (!b) return;
  const src = b.getAttribute('data-src');
  if (!src || src === dlState.currentSource) return;
  dlState.currentSource = src;
  setSourceTab(src);
  const id = $("id").value.trim(), ep = $("ep").value.trim();
  loadDownloadsForSource(id, ep, src);
});

function setSourceTab(src) {
  [...dlSourceTabs.querySelectorAll('.src-tab')].forEach(b => {
    const active = b.getAttribute('data-src') === src;
    b.classList.toggle('active', active);
    b.setAttribute('aria-selected', active ? 'true' : 'false');
  });
}

function loadDownloadsForSource(id, ep, src) {
  const label = src === 'pahe' ? 'animepahe' : 'Anime Tosho';
  dlSub.textContent = 'Episode ' + ep + ' · fetching releases from ' + label + '…';
  dlFilters.innerHTML = '';
  dlBody.innerHTML = '<div class="dl-loading"><div class="spinner"></div>Searching ' + label + ' for releases…</div>';
  fetchDownloads(id, ep, src);
}

async function fetchDownloads(id, ep, src) {
  try {
    const url = '/api/anime/'+encodeURIComponent(id)+'/downloads/'+encodeURIComponent(ep)+
                '?limit=20&source='+encodeURIComponent(src||'pahe');
    const r = await fetch(url).then(r=>r.json());
    log('downloads ('+src+')', r);
    if (!r.success) {
      dlBody.innerHTML = '<div class="dl-empty">Error: '+(r.error?.message||'unknown')+
        '<br><span style="font-size:11px;margin-top:8px;display:inline-block">'+
        'Try the other source above.</span></div>';
      return;
    }
    dlState.groups = r.data?.groups || [];
    dlState.flat = r.data?.flat || [];
    dlState.title = r.data?.title || '';
    dlState.episode = r.data?.episode || ep;
    dlState.source = r.data?.source || src || 'pahe';
    dlState.filter = { type: 'all', quality: 'all', group: 'all' };
    dlTitle.textContent = dlState.title || 'Downloads';
    // If pahe silently fell back to tosho server-side, keep the user's
    // selected tab in sync so the UI reflects what's actually shown.
    if (r.data?.fallback_from === 'pahe' && dlState.source === 'tosho') {
      dlState.currentSource = 'tosho';
      setSourceTab('tosho');
    }
    if (!dlState.flat.length) {
      const otherSrc = dlState.source === 'pahe' ? 'tosho' : 'pahe';
      dlSub.textContent = 'Episode '+dlState.episode+' · no releases found via '+dlState.source+'.';
      dlFilters.innerHTML = '';
      dlBody.innerHTML = '<div class="dl-empty">No downloadable releases found for this episode on <b>'+dlState.source+'</b>.<br>'+
        '<button class="dl-btn primary" style="margin-top:14px" onclick="document.querySelector(\'.src-tab[data-src='+otherSrc+']\').click()">Try '+otherSrc+' instead →</button></div>';
      return;
    }
    let subText = 'Episode '+dlState.episode+' · '+dlState.flat.length+' release'+(dlState.flat.length===1?'':'s')+' from '+dlState.source;
    if (r.data?.fallback_from) subText += ' (fell back from '+r.data.fallback_from+')';
    dlSub.textContent = subText;
    renderDlFilters();
    renderDlBody();
  } catch (e) {
    log('downloads error', e.message);
    dlBody.innerHTML = '<div class="dl-empty">Failed to load downloads: '+e.message+'</div>';
  }
}

function renderDlFilters() {
  const types = ['all', ...new Set(dlState.flat.map(r => r.type).filter(Boolean))];
  const qualities = ['all', ...[...new Set(dlState.flat.map(r => r.quality).filter(q => q && q !== 'unknown'))].sort((a,b)=>parseInt(b)-parseInt(a))];
  const groups = ['all', ...new Set(dlState.groups.map(g => g.group))];
  const make = (key, val, label) => '<button class="fchip '+(dlState.filter[key]===val?'active':'')+'" data-k="'+key+'" data-v="'+val+'">'+label+'</button>';
  let html = '<span style="color:var(--mute);font-size:11px;align-self:center;margin-right:4px">Type:</span>';
  types.forEach(t => html += make('type', t, t === 'all' ? 'All' : t.charAt(0).toUpperCase()+t.slice(1)));
  html += '<span style="color:var(--mute);font-size:11px;align-self:center;margin:0 4px 0 8px">Quality:</span>';
  qualities.forEach(q => html += make('quality', q, q === 'all' ? 'All' : q));
  if (groups.length > 2) {
    html += '<span style="color:var(--mute);font-size:11px;align-self:center;margin:0 4px 0 8px">Group:</span>';
    groups.forEach(g => html += make('group', g, g === 'all' ? 'All' : g));
  }
  dlFilters.innerHTML = html;
}
dlFilters.addEventListener('click', e => {
  const b = e.target.closest('.fchip'); if (!b) return;
  dlState.filter[b.getAttribute('data-k')] = b.getAttribute('data-v');
  renderDlFilters();
  renderDlBody();
});

function renderDlBody() {
  const f = dlState.filter;
  const matches = r =>
    (f.type === 'all' || r.type === f.type) &&
    (f.quality === 'all' || r.quality === f.quality) &&
    (f.group === 'all' || r.group === f.group);

  const groupsHtml = [];
  for (const g of dlState.groups) {
    if (f.group !== 'all' && g.group !== f.group) continue;
    const rows = (g.releases || []).filter(matches);
    if (!rows.length) continue;
    groupsHtml.push(
      '<div class="dl-group">'+
        '<div class="dl-group-h">'+escapeHtml(g.group)+
          ' <span class="badge">'+escapeHtml(g.type||'sub')+'</span>'+
          ' <span style="color:var(--mute);font-weight:500;font-size:11px">'+rows.length+' release'+(rows.length===1?'':'s')+'</span>'+
        '</div>'+
        rows.map(renderDlRow).join('')+
      '</div>'
    );
  }
  if (!groupsHtml.length) {
    dlBody.innerHTML = '<div class="dl-empty">No releases match the current filters.</div>';
    return;
  }
  dlBody.innerHTML = groupsHtml.join('');
}

function renderDlRow(r) {
  const seedStr = (typeof r.seeders === 'number') ? (r.seeders + ' seeders') : '';
  const sizeStr = r.size_human || '';
  const langStr = r.language || (r.type === 'dub' ? 'English Dub' : 'English Sub');
  const meta = [
    r.quality && r.quality !== 'unknown' ? '<span class="qbadge">'+escapeHtml(r.quality)+'</span>' : '',
    r.container ? '<span>'+escapeHtml(r.container.toUpperCase())+'</span>' : '',
    sizeStr ? '<span>'+escapeHtml(sizeStr)+'</span>' : '',
    seedStr ? '<span>'+escapeHtml(seedStr)+'</span>' : '',
    langStr ? '<span>'+escapeHtml(langStr)+'</span>' : '',
  ].filter(Boolean).join('');

  // Pahe rows have only view_page (the pahe.win continue page → kwik
  // player → real download button). Tosho rows have view_page (Anime
  // Tosho mirror list) PLUS .p2p / magnet / .nzb. Re-label the
  // primary action so the wording matches each source's UX.
  const actions = [];
  const isPahe = dlState.source === 'pahe' && r.view_page && !r.p2p_url && !r.magnet_uri;
  if (r.view_page) {
    actions.push('<a class="dl-btn primary" href="'+escapeAttr(r.view_page)+'" target="_blank" rel="noopener">'+
      iconDownload()+(isPahe ? 'Download' : 'View mirrors')+'</a>');
  }
  if (r.p2p_url) actions.push('<a class="dl-btn" href="'+escapeAttr(r.p2p_url)+'" target="_blank" rel="noopener">'+iconDownload()+'.p2p</a>');
  if (r.magnet_uri) actions.push('<a class="dl-btn" href="'+escapeAttr(r.magnet_uri)+'">'+iconMagnet()+'Magnet</a>');
  if (r.nzb_url) actions.push('<a class="dl-btn" href="'+escapeAttr(r.nzb_url)+'" target="_blank" rel="noopener">.nzb</a>');

  return (
    '<div class="dl-row">'+
      '<div>'+
        '<div class="meta-line">'+escapeHtml(r.title||'(untitled)')+'</div>'+
        '<div class="sub-line">'+meta+'</div>'+
      '</div>'+
      '<div class="dl-actions">'+actions.join('')+'</div>'+
    '</div>'
  );
}

function iconDownload() { return '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>'; }
function iconExternal() { return '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg>'; }
function iconMagnet() { return '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 3v8a7 7 0 0 0 14 0V3"/><path d="M5 3h4v8"/><path d="M15 3h4v8"/></svg>'; }
function escapeHtml(s) { return String(s==null?'':s).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c])); }
function escapeAttr(s) { return escapeHtml(s); }


// initial volume restore
v.volume = 1; updateVolIcon();
</script>
</body></html>`

// Mount registers /docs, /openapi.json, and /demo on the given app/router.
func Mount(app *fiber.App) {
	app.Get("/docs", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(swaggerHTML)
	})
	app.Get("/demo", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html; charset=utf-8")
		return c.SendString(demoHTML)
	})
	app.Get("/openapi.json", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json; charset=utf-8")
		return c.SendString(openapiJSON)
	})
}

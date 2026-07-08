package docs

const openapiJSON = `{
  "openapi": "3.0.3",
  "info": {
    "title": "Miru API",
    "version": "1.0.0",
    "description": "Thin, cached, normalized REST wrapper combining Animetsu, Animex.one, and Miruro.to streams. Built-in HLS proxy, automatic server fallback, and a /demo player.\n\n- Use proxy_url returned by /api/watch with hls.js, ExoPlayer, AVPlayer, or VLC.\n- server=auto picks a working server automatically (recommended)."
  },
  "servers": [{ "url": "/" }],
  "tags": [
    { "name": "discovery" },
    { "name": "details" },
    { "name": "stream" },
    { "name": "proxy" },
    { "name": "meta" }
  ],
  "paths": {
    "/healthz": { "get": { "tags":["meta"], "summary": "Liveness probe", "responses": { "200": { "description": "ok" } } } },
    "/api/home": { "get": { "tags":["discovery"], "summary": "All home rails (trending, seasonal, popular, top, upcoming)", "responses": { "200": { "description": "ok" } } } },
    "/api/trending": { "get": { "tags":["discovery"], "summary": "Trending rail", "responses": { "200": { "description": "ok" } } } },
    "/api/season": { "get": { "tags":["discovery"], "summary": "This-season rail", "responses": { "200": { "description": "ok" } } } },
    "/api/popular": { "get": { "tags":["discovery"], "summary": "All-time popular rail", "responses": { "200": { "description": "ok" } } } },
    "/api/top-rated": { "get": { "tags":["discovery"], "summary": "Top-rated rail", "responses": { "200": { "description": "ok" } } } },
    "/api/upcoming": { "get": { "tags":["discovery"], "summary": "Top upcoming rail", "responses": { "200": { "description": "ok" } } } },
    "/api/recent": {
      "get": {
        "tags": ["discovery"], "summary": "Recently updated (paginated)",
        "parameters": [
          { "name": "page", "in": "query", "schema": { "type": "integer", "default": 1 } },
          { "name": "per_page", "in": "query", "schema": { "type": "integer", "default": 12 } }
        ],
        "responses": { "200": { "description": "ok" } }
      }
    },
    "/api/schedule": { "get": { "tags":["discovery"], "summary": "Today's airing schedule", "responses": { "200": { "description": "ok" } } } },
    "/api/random": { "get": { "tags":["discovery"], "summary": "Random anime ID", "responses": { "200": { "description": "ok" } } } },
    "/api/search": {
      "get": {
        "tags": ["discovery"], "summary": "Search + filter (AniList vocabulary)",
        "parameters": [
          { "name":"q","in":"query","schema":{"type":"string"} },
          { "name":"genres","in":"query","schema":{"type":"string"},"description":"comma-separated, e.g. Action,Drama" },
          { "name":"format","in":"query","schema":{"type":"string","enum":["TV","TV_SHORT","MOVIE","SPECIAL","OVA","ONA","MUSIC"]} },
          { "name":"status","in":"query","schema":{"type":"string","enum":["RELEASING","FINISHED","NOT_YET_RELEASED","CANCELLED","HIATUS"]} },
          { "name":"season","in":"query","schema":{"type":"string","enum":["WINTER","SPRING","SUMMER","FALL"]} },
          { "name":"year","in":"query","schema":{"type":"integer"} },
          { "name":"sort","in":"query","schema":{"type":"string","enum":["POPULARITY_DESC","SCORE_DESC","TRENDING_DESC","UPDATED_AT_DESC","START_DATE_DESC","FAVOURITES_DESC","TITLE_ROMAJI"]} },
          { "name":"page","in":"query","schema":{"type":"integer","default":1} }
        ],
        "responses": { "200": { "description": "ok" } }
      }
    },
    "/api/anime/{id}": {
      "get": {
        "tags":["details"], "summary":"Full metadata (incl. trailer + next_airing_ep)",
        "parameters":[{ "name":"id","in":"path","required":true,"schema":{"type":"string"} }],
        "responses": { "200": { "description":"ok" } }
      }
    },
    "/api/anime/{id}/episodes": {
      "get": {
        "tags":["details"], "summary":"Episode list with thumbnails",
        "parameters":[{ "name":"id","in":"path","required":true,"schema":{"type":"string"} }],
        "responses": { "200": { "description":"ok" } }
      }
    },
    "/api/anime/{id}/views/{ep}": {
      "get": {
        "tags":["details"], "summary":"Episode view counter",
        "parameters":[
          { "name":"id","in":"path","required":true,"schema":{"type":"string"} },
          { "name":"ep","in":"path","required":true,"schema":{"type":"string"} }
        ],
        "responses": { "200": { "description":"ok" } }
      }
    },
    "/api/anime/{id}/servers/{ep}": {
      "get": {
        "tags":["stream"],
        "summary":"List stream servers AND probe each in parallel — returns working flag + latency",
        "parameters":[
          { "name":"id","in":"path","required":true,"schema":{"type":"string"} },
          { "name":"ep","in":"path","required":true,"schema":{"type":"string"} },
          { "name":"source_type","in":"query","schema":{"type":"string","enum":["sub","dub"],"default":"sub"} }
        ],
        "responses": { "200": { "description":"ok" } }
      }
    },
    "/api/anime/{id}/watch/{ep}": {
      "get": {
        "tags":["stream"],
        "summary":"Resolve playable sources (auto-fallback over working servers). Returns sources[] with proxy_url ready for hls.js / ExoPlayer / VLC, plus an optional subtitles[] array (label/url/lang/default) — wire those into <track kind=subtitles> for soft-subs that the user can toggle on/off in the player.",
        "parameters":[
          { "name":"id","in":"path","required":true,"schema":{"type":"string"} },
          { "name":"ep","in":"path","required":true,"schema":{"type":"string"} },
          { "name":"server","in":"query","schema":{"type":"string","default":"auto","enum":["auto","pahe","kite","fsoft","animex:Auto","miruro:ally"]} },
          { "name":"source_type","in":"query","schema":{"type":"string","enum":["sub","dub"],"default":"sub"} },
          { "name":"fallback","in":"query","schema":{"type":"boolean","default":true} }
        ],
        "responses": { "200": { "description":"ok" } }
      }
    },
    "/api/anime/{id}/download/{ep}": {
      "get": {
        "tags":["stream"],
        "summary":"Best-quality proxied HLS URL + suggested .mp4 filename + soft-subtitle list. Use the 'hint' field for the exact ffmpeg command that muxes the stream into a single MP4 (with optional subtitles burned-in or kept as soft tracks).",
        "parameters":[
          { "name":"id","in":"path","required":true,"schema":{"type":"string"} },
          { "name":"ep","in":"path","required":true,"schema":{"type":"string"} },
          { "name":"quality","in":"query","schema":{"type":"string","enum":["1080p","720p","480p","360p","master"]} },
          { "name":"server","in":"query","schema":{"type":"string","default":"auto","enum":["auto","pahe","kite","fsoft","animex:Auto","miruro:ally"]} },
          { "name":"source_type","in":"query","schema":{"type":"string","enum":["sub","dub"],"default":"sub"} }
        ],
        "responses": { "200": { "description":"ok" } }
      }
    },
    "/api/watch": {
      "get": {
        "tags":["stream"], "summary":"Same as /api/anime/{id}/watch/{ep} but with query params. Returns soft-subtitle URLs — the demo player and any hls.js/ExoPlayer/AVPlayer client can plug them into <track> for user-toggleable subs.",
        "parameters":[
          { "name":"id","in":"query","required":true,"schema":{"type":"string"} },
          { "name":"ep","in":"query","required":true,"schema":{"type":"string"} },
          { "name":"server","in":"query","schema":{"type":"string","default":"auto","enum":["auto","pahe","kite","fsoft","animex:Auto","miruro:ally"]} },
          { "name":"source_type","in":"query","schema":{"type":"string","default":"sub","enum":["sub","dub"]} },
          { "name":"fallback","in":"query","schema":{"type":"boolean","default":true} }
        ],
        "responses": { "200": { "description":"ok" } }
      }
    },
    "/api/proxy/hls": {
      "get": {
        "tags":["proxy"], "summary":"HLS playlist+segment+AES key proxy with URI rewriting. Streams segments without buffering so seeking is instant.",
        "parameters":[{ "name":"url","in":"query","required":true,"schema":{"type":"string"} }],
        "responses": { "200": { "description":"ok" } }
      }
    }
  }
}`

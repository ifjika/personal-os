const DEFAULT_TTS_URL = "https://ifjika-personal-os.hf.space";

async function probeGemini(apiKey) {
  if (!apiKey) return { ok: false, label: "🔴 Missing GEMINI_API_KEY" };
  const t0 = Date.now();
  try {
    const resp = await fetch(
      `https://generativelanguage.googleapis.com/v1beta/models/gemini-3-flash-preview?key=${apiKey}`,
      { signal: AbortSignal.timeout(6000) }
    );
    const ms = Date.now() - t0;
    if (resp.ok) return { ok: true, label: `🟢 Online (${ms}ms)` };
    const errBody = await resp.text().catch(() => "");
    return { ok: false, label: `🔴 Error HTTP ${resp.status} (${ms}ms)` + (errBody ? ` — ${errBody.slice(0, 80)}` : "") };
  } catch (err) {
    return { ok: false, label: `🔴 Unreachable (${err.message})` };
  }
}

async function probeHuggingFace(ttsUrl) {
  const t0 = Date.now();
  try {
    const resp = await fetch(`${ttsUrl}/gradio_api/info`, {
      signal: AbortSignal.timeout(8000)
    });
    const ms = Date.now() - t0;
    if (resp.ok) {
      const info = await resp.json().catch(() => null);
      const hasEndpoint = Boolean(info?.named_endpoints?.["/synthesize_voice"]);
      return hasEndpoint
        ? { ok: true, label: `🟢 Awake & Ready (${ms}ms)` }
        : { ok: false, label: `🟡 Online (${ms}ms), but /synthesize_voice endpoint missing` };
    }
    return { ok: false, label: `🟡 HTTP ${resp.status} (${ms}ms) — Space may be building or sleeping` };
  } catch (err) {
    if (err.name === "TimeoutError" || err.name === "AbortError") {
      return { ok: false, label: "🟡 Sleeping / Cold Start (>8s) — Wake-up triggered, retry in 30s" };
    }
    return { ok: false, label: `🔴 Unreachable (${err.message})` };
  }
}

async function probeTelegram(botToken) {
  if (!botToken) return { ok: false, label: "🔴 Missing TELEGRAM_BOT_TOKEN" };
  const t0 = Date.now();
  try {
    const resp = await fetch(`https://api.telegram.org/bot${botToken}/getMe`, {
      signal: AbortSignal.timeout(5000)
    });
    const ms = Date.now() - t0;
    if (resp.ok) {
      const data = await resp.json();
      const botName = data.result?.username ? `@${data.result.username}` : "Connected";
      return { ok: true, label: `🟢 ${botName} (${ms}ms)` };
    }
    return { ok: false, label: `🔴 Error HTTP ${resp.status} (${ms}ms)` };
  } catch (err) {
    return { ok: false, label: `🔴 Unreachable (${err.message})` };
  }
}

/**
 * Runs live active health probes in parallel across Worker secrets, Gemini, Hugging Face, and Telegram.
 */
export async function runHealthDiagnostics(env) {
  const ttsUrl = (env.TTS_URL || DEFAULT_TTS_URL).replace(/\/+$/, "");
  const missing = [
    !env.GEMINI_API_KEY && "GEMINI_API_KEY",
    !env.TELEGRAM_BOT_TOKEN && "TELEGRAM_BOT_TOKEN",
    !env.TELEGRAM_CHAT_ID && "TELEGRAM_CHAT_ID"
  ].filter(Boolean);

  const secretsOk = missing.length === 0;
  const secretsStatus = secretsOk ? "🟢 All Configured" : `🔴 Missing (${missing.join(", ")})`;

  const [geminiResult, hfResult, tgResult] = await Promise.all([
    probeGemini(env.GEMINI_API_KEY),
    probeHuggingFace(ttsUrl),
    probeTelegram(env.TELEGRAM_BOT_TOKEN)
  ]);

  const allHealthy = secretsOk && geminiResult.ok && hfResult.ok && tgResult.ok;
  const wibTime = new Date(Date.now() + 7 * 3600 * 1000).toISOString().replace("T", " ").slice(0, 19);

  return {
    ok: allHealthy,
    cloudflare: "🟢 Online (Edge Worker)",
    secrets: secretsStatus,
    gemini: geminiResult.label,
    huggingface: hfResult.label,
    telegram: tgResult.label,
    summary: allHealthy
      ? "All systems operational. Ready to generate audio!"
      : "One or more services need attention (see status above).",
    checkedAtWib: wibTime
  };
}

/**
 * Wakes up a sleeping or paused Hugging Face Space using Official Hub API (if HF_TOKEN is set)
 * or falls back to an HTML root warm-up request.
 */
export async function wakeHuggingFaceSpace(env) {
  const ttsUrl = (env.TTS_URL || DEFAULT_TTS_URL).replace(/\/+$/, "");
  const spaceId = env.HF_SPACE_ID || "ifjika/personal-os";
  const hfToken = env.HF_TOKEN;

  if (hfToken) {
    try {
      const resp = await fetch(`https://huggingface.co/api/spaces/${spaceId}/restart`, {
        method: "POST",
        headers: {
          Authorization: `Bearer ${hfToken}`,
          "Content-Type": "application/json"
        },
        signal: AbortSignal.timeout(8000)
      });

      if (resp.ok) {
        return {
          ok: true,
          method: "Official Hub Restart API",
          message: "Wake-up/restart triggered! Wait ~30–60s for F5-TTS to load, then check /ping."
        };
      }
      const errText = await resp.text().catch(() => "");
      return {
        ok: false,
        method: "Official Hub Restart API",
        message: `HTTP ${resp.status}: ${errText.slice(0, 120)}`
      };
    } catch (err) {
      return {
        ok: false,
        method: "Official Hub Restart API",
        message: err.message
      };
    }
  }

  try {
    const resp = await fetch(ttsUrl, {
      headers: { Accept: "text/html" },
      signal: AbortSignal.timeout(10000)
    });
    if (resp.ok) {
      return {
        ok: true,
        method: "HTML Warm-Up Ping",
        message: "Space responded 200 OK and is already awake!"
      };
    }
    return {
      ok: true,
      method: "HTML Warm-Up Ping",
      message: `Wake signal sent (HTTP ${resp.status}). Wait ~30–60s then check /ping. Tip: set HF_TOKEN secret for guaranteed hard-restart.`
    };
  } catch (err) {
    return {
      ok: err.name === "TimeoutError" || err.name === "AbortError",
      method: "HTML Warm-Up Ping",
      message: `Space is booting or unreachable (${err.message}). Check /ping in ~45s.`
    };
  }
}

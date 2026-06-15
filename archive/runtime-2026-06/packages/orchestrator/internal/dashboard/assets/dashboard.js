(function () {
  "use strict";

  const app = document.getElementById("dashboard-app");
  if (!app) {
    return;
  }

  const apiPrefix = app.dataset.apiPrefix || "/api/dashboard";
  const TWEAK_KEY = "lazyai.dashboard.tweaks.v1";
  const TWEAK_DEFAULTS = { theme: "light", density: "balanced", nav: "top" };
  const TWEAK_FIELDS = ["theme", "density", "nav"];

  const ROUTES = {
    OVERVIEW: "#/overview",
    RUNS: "#/runs",
    RUN_DETAIL: "#/run-detail",
    CATALOG: "#/catalog",
    ERRORS: "#/errors",
  };

  const SECTION_BY_ROUTE = {
    overview: "overview",
    runs: "runs",
    "run-detail": "run-detail",
    catalog: "catalog",
    errors: "errors",
  };

  const state = {
    currentRun: null,
    eventSource: null,
    lastEventID: 0,
    catalogItems: [],
    selectedRun: null,
    activeRoute: { name: "overview" },
    tweaks: Object.assign({}, TWEAK_DEFAULTS),
    runFilters: { attention: "", hasErrors: false },
    runs: { items: [], nextCursor: "", totalLoaded: 0 },
    activity: {
      source: null,
      paused: false,
      lastEventID: 0,
      reconnectTimer: null,
      reconnectAttempts: 0,
      overviewRefreshTimer: null,
      runsRefreshTimer: null,
      seenIDs: new Set(),
    },
  };

  const ACTIVITY_FEED_CAP = 50;
  const ACTIVITY_LEVEL_BY_TYPE = {
    run_started: "info",
    run_completed: "ok",
    run_failed: "err",
    step_started: "info",
    step_completed: "ok",
    step_failed: "err",
    task_started: "info",
    task_completed: "ok",
    task_failed: "err",
    phase_started: "info",
    phase_completed: "ok",
    phase_failed: "err",
    handoff_created: "warn",
    budget_updated: "info",
    error_recorded: "err",
  };
  const ACTIVITY_GLYPH_BY_LEVEL = { info: "·", ok: "✓", warn: "↯", err: "✕" };
  const ACTIVITY_TOAST_TYPES = new Set([
    "run_failed",
    "step_failed",
    "task_failed",
    "phase_failed",
    "error_recorded",
    "handoff_created",
  ]);
  const RUN_STATE_CHANGE_TYPES = new Set([
    "run_started",
    "run_completed",
    "run_failed",
    "phase_started",
    "phase_completed",
  ]);

  function byID(id) {
    return document.getElementById(id);
  }

  function setStatus(message, level) {
    const target = byID("status-message");
    if (!target) {
      return;
    }
    target.textContent = message;
    target.classList.remove("ok", "error");
    if (level) {
      target.classList.add(level);
    }
  }

  function buildURL(path, params) {
    const url = new URL(apiPrefix + path, window.location.origin);
    Object.entries(params || {}).forEach(([key, value]) => {
      if (value !== undefined && value !== null && String(value) !== "") {
        url.searchParams.set(key, value);
      }
    });
    return url;
  }

  async function fetchJSON(path, params) {
    const response = await fetch(buildURL(path, params), {
      headers: { Accept: "application/json" },
    });
    let payload = null;
    try {
      payload = await response.json();
    } catch (_err) {
      payload = null;
    }
    if (!response.ok) {
      const message = payload && payload.error ? payload.error.message : `Request failed with status ${response.status}`;
      throw new Error(message);
    }
    return payload;
  }

  function clear(target) {
    if (target) {
      target.replaceChildren();
    }
  }

  function text(value, fallback) {
    if (value === undefined || value === null || value === "") {
      return fallback || "—";
    }
    return String(value);
  }

  function formatDate(value) {
    if (!value) {
      return "—";
    }
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) {
      return value;
    }
    return date.toLocaleString();
  }

  function element(tag, attrs, children) {
    const node = document.createElement(tag);
    Object.entries(attrs || {}).forEach(([key, value]) => {
      if (key === "class") {
        node.className = value;
      } else if (key === "text") {
        node.textContent = value;
      } else if (key.startsWith("data")) {
        node.dataset[key.slice(4).replace(/^./, (c) => c.toLowerCase())] = value;
      } else {
        node.setAttribute(key, value);
      }
    });
    (children || []).forEach((child) => node.append(child));
    return node;
  }

  function renderEmptyState(target, message) {
    clear(target);
    if (target) {
      target.append(element("li", { class: "empty-state", text: message || target.dataset.empty || "Nothing to show." }));
    }
  }

  function renderErrorState(target, message) {
    clear(target);
    if (target) {
      target.append(element("li", { class: "empty-state", text: message || "Unable to load data." }));
    }
    setStatus(message || "Unable to load dashboard data.", "error");
  }

  function renderKV(target, entries, emptyMessage) {
    clear(target);
    const filtered = Object.entries(entries || {}).filter(([, value]) => value !== undefined && value !== null && value !== "");
    if (filtered.length === 0) {
      target.append(element("dt", { text: "Status" }), element("dd", { text: emptyMessage || target.dataset.empty || "No data." }));
      return;
    }
    filtered.forEach(([key, value]) => {
      target.append(element("dt", { text: labelize(key) }), element("dd", { text: text(value) }));
    });
  }

  function labelize(value) {
    return String(value).replace(/([A-Z])/g, " $1").replace(/[_-]/g, " ").replace(/^./, (c) => c.toUpperCase());
  }

  function activeRunEntries(activeRuns) {
    const active = activeRuns || {};
    return {
      Chains: active.chains || active.Chains || 0,
      Teams: active.teams || active.Teams || 0,
      Workflows: active.workflows || active.Workflows || 0,
    };
  }

  function renderOverview(overview) {
    renderKV(byID("health-status"), {
      Status: overview.health && overview.health.status,
      Name: overview.health && overview.health.name,
      Port: overview.health && overview.health.port,
      PID: overview.health && overview.health.pid,
      ProjectRoot: overview.health && overview.health.projectRoot,
      StartedAt: overview.health && overview.health.startedAt,
      GeneratedAt: overview.generatedAt,
    });
    renderKV(byID("active-runs"), activeRunEntries(overview.activeRuns || (overview.health && overview.health.activeRuns)));
    renderKV(byID("run-counts"), overview.runCountsByState || {}, "No runs recorded yet.");
    renderKV(byID("catalog-counts"), Object.assign({ Total: overview.catalogCounts && overview.catalogCounts.total }, (overview.catalogCounts && overview.catalogCounts.byKind) || {}), "No catalog definitions found.");
    renderRunCollection(byID("recent-runs"), overview.recentRuns || [], true);
    renderErrorCollection(byID("recent-errors"), overview.recentErrors || []);
    updateNavCounts({
      catalog: overview.catalogCounts && overview.catalogCounts.total,
      errors: (overview.recentErrors || []).length,
    });
    setStatus("Dashboard data loaded.", "ok");
  }

  async function fetchOverview() {
    try {
      const overview = await fetchJSON("/overview");
      renderOverview(overview || {});
    } catch (err) {
      renderErrorState(byID("recent-errors"), err.message);
    }
  }

  function runFilters(extra) {
    const limitValue = Number.parseInt(byID("run-limit") && byID("run-limit").value, 10);
    const boundedLimit = Number.isFinite(limitValue) ? Math.min(Math.max(limitValue, 1), 200) : 50;
    const params = {
      kind: byID("run-kind-filter") && byID("run-kind-filter").value,
      state: byID("run-state-filter") && byID("run-state-filter").value.trim(),
      search: ((byID("run-search") && byID("run-search").value) || "").trim(),
      limit: boundedLimit,
    };
    // attention=budget is handled client-side; everything else goes to the server.
    if (state.runFilters.attention && state.runFilters.attention !== "budget") {
      params.attention = state.runFilters.attention;
    }
    if (state.runFilters.hasErrors) {
      params.has_errors = "true";
    }
    return Object.assign(params, extra || {});
  }

  function applyClientAttention(items) {
    if (state.runFilters.attention !== "budget") {
      return items;
    }
    return items.filter((item) => {
      const health = String(item.budgetHealth || "").toLowerCase();
      return health === "warning" || health === "critical" || health === "over";
    });
  }

  async function fetchRuns({ append = false } = {}) {
    try {
      const params = append ? runFilters({ cursor: state.runs.nextCursor }) : runFilters();
      const payload = await fetchJSON("/runs", params);
      const incoming = (payload && payload.items) || [];
      const merged = append ? state.runs.items.concat(incoming) : incoming;
      state.runs.items = merged;
      state.runs.nextCursor = (payload && payload.nextCursor) || "";
      state.runs.totalLoaded = merged.length;
      const visible = applyClientAttention(merged);
      renderRunCollection(byID("run-list"), visible, false);
      renderRunPagination(visible.length, merged.length, !!state.runs.nextCursor);
      updateNavCounts({ runs: visible.length });
      setStatus("Run list loaded.", "ok");
    } catch (err) {
      renderErrorState(byID("run-list"), err.message);
    }
  }

  function renderRunPagination(visibleCount, loadedCount, hasMore) {
    const summary = byID("run-count-summary");
    if (summary) {
      const filterNote = state.runFilters.attention === "budget"
        ? ` · ${visibleCount} match budget chip (client-side)`
        : "";
      summary.textContent = `${loadedCount} loaded${filterNote}${hasMore ? " · more available" : ""}`;
    }
    const more = byID("run-load-more");
    if (more) {
      more.hidden = !hasMore;
    }
  }

  function syncAttentionChips() {
    document.querySelectorAll("[data-attention]").forEach((node) => {
      const key = node.dataset.attention;
      node.setAttribute("aria-pressed", state.runFilters.attention === key ? "true" : "false");
    });
    document.querySelectorAll("[data-has-errors]").forEach((node) => {
      node.setAttribute("aria-pressed", state.runFilters.hasErrors ? "true" : "false");
    });
  }

  function bindAttentionChips() {
    document.querySelectorAll("[data-attention]").forEach((node) => {
      node.addEventListener("click", () => {
        state.runFilters.attention = node.dataset.attention || "";
        syncAttentionChips();
        fetchRuns();
      });
    });
    document.querySelectorAll("[data-has-errors]").forEach((node) => {
      node.addEventListener("click", () => {
        state.runFilters.hasErrors = !state.runFilters.hasErrors;
        syncAttentionChips();
        fetchRuns();
      });
    });
  }

  function renderRunCollection(target, runs, compact) {
    clear(target);
    const empty = compact ? null : byID("run-list-empty");
    if (empty) {
      empty.hidden = runs.length > 0;
    }
    if (!runs.length) {
      if (compact) {
        renderEmptyState(target, target.dataset.empty || "No recent runs.");
      }
      return;
    }
    runs.forEach((run) => target.append(renderRunItem(run, compact)));
  }

  function renderRunItem(run, compact) {
    const item = element("li", { class: "run-item" });
    const title = `${text(run.kind)} / ${text(run.id)}`;
    const header = element("div", { class: "run-item-header" }, [
      element("strong", { text: title }),
      element("span", { class: badgeClass(run.state), text: text(run.state) }),
    ]);
    item.append(header);
    item.append(element("p", { class: "muted", text: `${text(run.definitionName)} ${text(run.definitionVersion, "")}`.trim() || "No definition metadata" }));
    item.append(element("p", { class: "muted", text: `Updated ${formatDate(run.updatedAt)} · errors ${run.errorCount || 0}` }));
    if (!compact) {
      const button = element("button", { type: "button", text: "Open details" });
      button.addEventListener("click", () => openRunFromList(run.kind, run.id));
      item.append(button);
    } else {
      // Compact rows act as buttons for both pointer and keyboard users.
      item.setAttribute("role", "button");
      item.tabIndex = 0;
      item.setAttribute("aria-label", `Open run ${text(run.kind)} ${text(run.id)}`);
      item.style.cursor = "pointer";
      item.addEventListener("click", () => openRunFromList(run.kind, run.id));
      item.addEventListener("keydown", (event) => {
        if (event.key === "Enter" || event.key === " ") {
          event.preventDefault();
          openRunFromList(run.kind, run.id);
        }
      });
    }
    return item;
  }

  function badgeClass(stateValue) {
    const normalized = String(stateValue || "").toLowerCase();
    if (normalized.includes("fail") || normalized.includes("error")) {
      return "badge error";
    }
    if (normalized.includes("running") || normalized.includes("pending")) {
      return "badge warning";
    }
    return "badge";
  }

  // Open from a run row: update hash → router will activate detail and load data.
  function openRunFromList(kind, id) {
    navigateTo(`#/runs/${encodeURIComponent(kind)}/${encodeURIComponent(id)}`);
  }

  async function openRunDetail(kind, id) {
    closeRunEvents();
    state.currentRun = { kind, id };
    state.lastEventID = 0;
    state.selectedRun = { kind, id };
    rememberSelectedRun(kind, id);
    refreshRunDetailNav();
    const empty = byID("run-detail-empty");
    if (empty) {
      empty.hidden = true;
    }
    const status = byID("run-detail-status");
    if (status) {
      status.textContent = `Loading ${kind}/${id}…`;
    }
    try {
      const [detail, budget] = await Promise.all([
        fetchJSON(`/runs/${encodeURIComponent(kind)}/${encodeURIComponent(id)}`),
        fetchJSON(`/runs/${encodeURIComponent(kind)}/${encodeURIComponent(id)}/budget`).catch(() => null),
      ]);
      if (budget) {
        detail.budget = budget;
      }
      renderRunDetail(detail || {});
      connectRunEvents(kind, id);
      setStatus(`Loaded ${kind}/${id}.`, "ok");
    } catch (err) {
      renderErrorState(byID("event-timeline"), err.message);
      if (status) {
        status.textContent = err.message;
      }
    }
  }

  function renderRunDetail(detail) {
    const summary = detail.summary || {};
    renderRunHero(summary);
    renderKV(byID("run-summary"), {
      Kind: summary.kind,
      ID: summary.id,
      Definition: summary.definitionName,
      Version: summary.definitionVersion,
      State: summary.state,
      Current: summary.current,
      ProjectRoot: summary.projectRoot,
      BudgetHealth: summary.budgetHealth,
      Errors: summary.errorCount,
      CreatedAt: formatDate(summary.createdAt),
      UpdatedAt: formatDate(summary.updatedAt),
      DecodeError: detail.stateDecodeError,
    });
    const stateTarget = byID("run-state-json");
    if (stateTarget) {
      stateTarget.textContent = JSON.stringify(detail.state || {}, null, 2);
    }
    renderBudget(detail.budget || null);
    renderBudgetCards(detail.budget || null);
    renderTimeline(detail);
    renderEventCollection(byID("event-timeline"), detail.events || []);
    renderErrorCollection(byID("run-errors"), detail.errors || []);
    renderExecutionPlan(detail.executionPlan || null);
    renderHandoffs(detail.handoffs || []);
  }

  // ===== Run detail hero ======================================================
  const RUN_STATE_GLYPHS = {
    running: { glyph: "●", cls: "chip-warn" },
    completed: { glyph: "✓", cls: "chip-ok" },
    failed: { glyph: "✕", cls: "chip-err" },
    gated: { glyph: "◇", cls: "chip-gate" },
    paused: { glyph: "❚❚", cls: "chip-warn" },
    awaiting_recovery: { glyph: "↻", cls: "chip-warn" },
    waiting_on_child: { glyph: "↯", cls: "chip-gate" },
  };

  function renderRunStateChip(stateValue) {
    const target = byID("run-detail-state-chip");
    if (!target) return;
    const key = String(stateValue || "").toLowerCase();
    const meta = RUN_STATE_GLYPHS[key] || { glyph: "·", cls: "" };
    target.className = `chip ${meta.cls}`;
    target.textContent = `${meta.glyph} ${stateValue || "—"}`;
  }

  function renderRunHero(summary) {
    const status = byID("run-detail-status");
    if (status) {
      const bestEffort = summary.kind === "team" || summary.kind === "workflow" ? " · best-effort runtime data" : "";
      const definition = summary.definitionName ? `${summary.definitionName} v${summary.definitionVersion || "—"}` : "—";
      status.textContent = `${definition} · started ${formatDate(summary.createdAt)} · updated ${formatDate(summary.updatedAt)}${bestEffort}`;
    }
    const idTarget = byID("run-detail-id");
    if (idTarget) {
      idTarget.textContent = summary.kind && summary.id ? `${summary.kind}/${summary.id}` : "—";
    }
    const eyebrow = byID("run-detail-eyebrow");
    if (eyebrow) {
      eyebrow.textContent = `${text(summary.kind)} run`;
    }
    renderRunStateChip(summary.state);
  }

  // ===== Vertical timeline ====================================================
  function renderTimeline(detail) {
    const target = byID("run-timeline");
    if (!target) return;
    clear(target);
    const summary = detail.summary || {};
    const events = detail.events || [];
    const eventsByStep = {};
    events.forEach((event) => {
      const stepID = (event.data && (event.data.stepId || event.data.taskId || event.data.phaseId)) || "_unscoped";
      eventsByStep[stepID] = eventsByStep[stepID] || [];
      eventsByStep[stepID].push(event);
    });

    let nodes = [];
    if (summary.kind === "chain" && Array.isArray(detail.steps)) {
      nodes = detail.steps.map((step) => ({
        id: step.stepId,
        title: `Step ${step.order != null ? step.order + 1 : "?"} · ${text(step.stepId)}`,
        meta: `${text(step.agent)} · ${text(step.stepType)}`,
        state: step.state,
        startedAt: step.startedAt,
        completedAt: step.completedAt,
        attempts: step.attempts,
        usage: step.usage,
        error: step.error,
      }));
    } else if (summary.kind === "team" && Array.isArray(detail.tasks)) {
      nodes = detail.tasks.map((task) => ({
        id: task.taskId,
        title: `Task · ${text(task.taskId)}`,
        meta: `${text(task.role || task.kind)} · ${text(task.agent)}`,
        state: task.state,
        startedAt: task.assignedAt || task.claimedAt,
        completedAt: task.completedAt,
        attempts: 0,
        usage: task.usage,
        error: task.error,
      }));
    } else if (summary.kind === "workflow" && Array.isArray(detail.phases)) {
      nodes = detail.phases.map((phase) => ({
        id: phase.phaseId,
        title: `Phase · ${text(phase.phaseId)}`,
        meta: `${text(phase.kind)}${phase.ref ? ` · ${phase.ref}` : ""}${phase.gate ? ` · gate=${phase.gate}` : ""}`,
        state: phase.state,
        startedAt: phase.startedAt,
        completedAt: phase.completedAt,
        attempts: 0,
        error: null,
      }));
    }

    if (nodes.length === 0) {
      target.append(element("li", { class: "empty-state", text: target.dataset.empty || "No timeline data for this run yet." }));
      return;
    }

    nodes.forEach((node) => target.append(renderTimelineNode(node, eventsByStep[node.id] || [])));
  }

  function renderTimelineNode(node, events) {
    const stateKey = String(node.state || "").toLowerCase();
    const marker = element("span", { class: "timeline-marker", "data-state": stateKey });
    if (stateKey === "completed") marker.textContent = "✓";
    else if (stateKey === "failed") marker.textContent = "✕";
    else if (stateKey === "running") marker.textContent = "●";
    else if (stateKey === "gated" || stateKey === "paused") marker.textContent = "◇";

    const titleParts = [
      element("strong", { text: node.title }),
      element("span", { class: "chip chip-toggle", "aria-pressed": "true", text: node.state || "—" }),
    ];
    if (node.attempts && node.attempts > 1) {
      titleParts.push(element("span", { class: "timeline-title-meta", text: `attempt ${node.attempts}` }));
    }
    const titleRow = element("div", { class: "timeline-title" }, titleParts);

    const metaBits = [];
    if (node.meta) metaBits.push(node.meta);
    if (node.startedAt) metaBits.push(`started ${formatDate(node.startedAt)}`);
    if (node.completedAt) metaBits.push(`done ${formatDate(node.completedAt)}`);
    if (node.usage && (node.usage.totalTokens || node.usage.costUsd || node.usage.wallClockMs)) {
      const usageBits = [];
      if (node.usage.totalTokens) usageBits.push(`${formatNumberCompact(node.usage.totalTokens)} tok`);
      if (node.usage.costUsd) usageBits.push(`$${node.usage.costUsd.toFixed(3)}`);
      if (node.usage.wallClockMs) usageBits.push(`${(node.usage.wallClockMs / 1000).toFixed(1)}s`);
      metaBits.push(usageBits.join(" · "));
    }
    const metaRow = element("p", { class: "timeline-title-meta", text: metaBits.join(" · ") || "—" });

    const eventsList = element("ul", { class: "timeline-events" });
    events.forEach((event) => {
      eventsList.append(element("li", { class: "timeline-event" }, [
        element("span", { class: "timeline-event-time", text: formatDate(event.createdAt) }),
        element("span", { text: event.eventType || "event" }),
      ]));
    });

    const body = element("div", { class: "timeline-body" }, [titleRow, metaRow, eventsList]);
    if (node.error && (node.error.message || node.error.code)) {
      body.append(element("div", { class: "timeline-error", text: `${text(node.error.code, "error")} · ${text(node.error.message)}` }));
    }
    return element("li", { class: "timeline-node" }, [marker, body]);
  }

  // ===== Budget cards =========================================================
  function renderBudgetCards(budget) {
    const target = byID("run-budget-cards");
    if (!target) return;
    clear(target);
    if (!budget || !budget.state) {
      target.append(element("p", { class: "empty-state", text: target.dataset.empty || "No budget data." }));
      return;
    }
    const dimensions = (budget.evaluation && budget.evaluation.dimensions) || {};
    const cards = [
      buildBudgetCard("tokens", budget.state.tokens, dimensions.tokens, formatNumberCompact),
      buildBudgetCard("cost · usd", budget.state.costUsd, dimensions.costUsd, (n) => `$${(n / 100).toFixed(3)}`),
      buildBudgetCard("wall clock", budget.state.wallClockMs, dimensions.wallClockMs, formatDuration),
      buildBudgetCard("retries", budget.state.retries, dimensions.retries, (n) => String(n)),
    ];
    cards.forEach((card) => card && target.append(card));
  }

  function buildBudgetCard(label, dim, health, formatter) {
    if (!dim) return null;
    const consumed = dim.consumed || 0;
    const limit = dim.limit || 0;
    const ratio = limit > 0 ? Math.min(1, consumed / limit) : 0;
    const healthValue = String(health || "").toLowerCase();
    const card = element("article", { class: "budget-card", "data-health": healthValue });
    card.append(element("span", { class: "budget-card-label", text: label }));
    card.append(element("strong", { class: "budget-card-value", text: formatter(consumed) }));
    const fill = element("span", { class: "budget-bar-fill", "data-health": healthValue });
    fill.style.width = `${Math.round(ratio * 100)}%`;
    const bar = element("div", { class: "budget-bar" }, [fill]);
    card.append(bar);
    const limitText = limit > 0 ? `of ${formatter(limit)}` : "no limit";
    const healthText = healthValue ? `· ${healthValue}` : "";
    card.append(element("span", { class: "budget-card-foot", text: `${limitText} ${healthText}`.trim() }));
    return card;
  }

  function renderExecutionPlan(plan) {
    const target = byID("run-execution-plan");
    if (!target) return;
    if (!plan || Object.keys(plan).length === 0) {
      target.textContent = target.dataset.empty || "No execution plan attached to this run.";
      return;
    }
    target.textContent = JSON.stringify(plan, null, 2);
  }

  function renderHandoffs(handoffs) {
    const target = byID("run-handoffs");
    if (!target) return;
    clear(target);
    if (!handoffs.length) {
      target.append(element("li", { class: "empty-state", text: target.dataset.empty || "No handoffs recorded." }));
      return;
    }
    handoffs.forEach((doc) => {
      target.append(element("li", { class: "error-item" }, [
        element("strong", { text: text(doc.title || doc.handoffType || "handoff") }),
        element("pre", { class: "json-block", text: JSON.stringify(doc, null, 2) }),
      ]));
    });
  }

  function formatNumberCompact(n) {
    if (n == null) return "—";
    if (n >= 1000) return `${(n / 1000).toFixed(n >= 10000 ? 0 : 1)}k`;
    return String(n);
  }

  function formatDuration(ms) {
    if (!ms) return "0s";
    if (ms < 1000) return `${ms}ms`;
    const seconds = ms / 1000;
    if (seconds < 60) return `${seconds.toFixed(1)}s`;
    const minutes = Math.floor(seconds / 60);
    const restSec = Math.round(seconds % 60);
    return restSec ? `${minutes}m ${restSec}s` : `${minutes}m`;
  }

  function bindCopyHandlers() {
    const copyBtn = byID("run-detail-copy-id");
    if (copyBtn) {
      copyBtn.addEventListener("click", () => {
        const idTarget = byID("run-detail-id");
        const idText = idTarget ? idTarget.textContent || "" : "";
        if (!idText) return;
        if (navigator.clipboard && navigator.clipboard.writeText) {
          navigator.clipboard.writeText(idText).then(
            () => setStatus("Run id copied to clipboard.", "ok"),
            () => setStatus("Could not copy to clipboard.", "error"),
          );
        }
      });
    }
  }

  function renderBudget(budget) {
    if (!budget) {
      renderKV(byID("budget-state"), {}, "No budget state available.");
      renderKV(byID("budget-evaluation"), {}, "No budget evaluation available.");
      return;
    }
    renderKV(byID("budget-state"), Object.assign({}, budget.state || {}, { LastUpdatedAt: budget.lastUpdatedAt, DecodeError: budget.decodeError }), "No budget state available.");
    renderKV(byID("budget-evaluation"), budget.evaluation || {}, "No budget evaluation available.");
  }

  function renderEventCollection(target, events) {
    clear(target);
    if (!events.length) {
      target.append(element("li", { class: "empty-state", text: target.dataset.empty || "No events for this run yet." }));
      return;
    }
    events.forEach((event) => appendEvent(event));
  }

  function appendEvent(event) {
    const target = byID("event-timeline");
    if (!target) {
      return;
    }
    if (event.id && event.id > state.lastEventID) {
      state.lastEventID = event.id;
    }
    const item = element("li", {}, [
      element("strong", { text: event.eventType || "event" }),
      element("p", { class: "muted", text: `${formatDate(event.createdAt)} · id ${text(event.id)}` }),
      element("pre", { class: "json-block", text: JSON.stringify(event.data || {}, null, 2) }),
    ]);
    target.append(item);
  }

  function renderErrorCollection(target, errors) {
    clear(target);
    if (!errors.length) {
      target.append(element("li", { class: "empty-state", text: target.dataset.empty || "No errors." }));
      return;
    }
    errors.forEach((entry) => {
      target.append(element("li", { class: "error-item" }, [
        element("strong", { text: `${text(entry.code)} · ${text(entry.category)}` }),
        element("p", { text: text(entry.message) }),
        element("p", { class: "muted", text: `${text(entry.runKind)} ${text(entry.runId)} ${formatDate(entry.createdAt)}` }),
      ]));
    });
  }

  function connectRunEvents(kind, id) {
    if (!("EventSource" in window)) {
      setStatus("Live events unavailable in this browser; showing replayed events only.", "error");
      return;
    }
    closeRunEvents();
    const url = buildURL(`/runs/${encodeURIComponent(kind)}/${encodeURIComponent(id)}/events`, { since_id: state.lastEventID });
    const source = new EventSource(url);
    state.eventSource = source;
    const handle = (event) => {
      try {
        appendEvent(JSON.parse(event.data));
      } catch (_err) {
        setStatus("Received an unreadable run event.", "error");
      }
    };
    source.onmessage = handle;
    ["run_started", "run_completed", "run_failed", "step_started", "step_completed", "step_failed", "task_started", "task_completed", "phase_started", "phase_completed", "budget_updated", "error_recorded", "handoff_created"].forEach((eventName) => {
      source.addEventListener(eventName, handle);
    });
    source.onerror = () => setStatus("Live event stream reconnecting…", "error");
  }

  function closeRunEvents() {
    if (state.eventSource) {
      state.eventSource.close();
      state.eventSource = null;
    }
  }

  async function fetchCatalog() {
    try {
      const payload = await fetchJSON("/catalog");
      state.catalogItems = (payload && payload.items) || [];
      renderCatalogList(applyCatalogFilters(state.catalogItems));
      updateNavCounts({ catalog: state.catalogItems.length });
      setStatus("Catalog loaded.", "ok");
    } catch (err) {
      renderErrorState(byID("catalog-list"), err.message);
    }
  }

  function catalogFilters() {
    return {
      kind: byID("catalog-kind-filter") && byID("catalog-kind-filter").value,
      search: ((byID("catalog-search") && byID("catalog-search").value) || "").trim().toLowerCase(),
      sort: (byID("catalog-sort") && byID("catalog-sort").value) || "name",
    };
  }

  function applyCatalogFilters(items) {
    const filters = catalogFilters();
    return (items || [])
      .filter((item) => !filters.kind || item.kind === filters.kind)
      .filter((item) => {
        if (!filters.search) {
          return true;
        }
        return `${item.kind || ""} ${item.name || ""}`.toLowerCase().includes(filters.search);
      })
      .slice()
      .sort((left, right) => compareCatalogItems(left, right, filters.sort));
  }

  function compareCatalogItems(left, right, sortKey) {
    if (sortKey === "updated") {
      const updated = String(right.updatedAt || "").localeCompare(String(left.updatedAt || ""));
      if (updated !== 0) {
        return updated;
      }
    } else if (sortKey === "kind") {
      const kind = String(left.kind || "").localeCompare(String(right.kind || ""));
      if (kind !== 0) {
        return kind;
      }
    }
    const name = String(left.name || "").localeCompare(String(right.name || ""));
    if (name !== 0) {
      return name;
    }
    return String(left.kind || "").localeCompare(String(right.kind || ""));
  }

  function groupCatalogItems(items) {
    return (items || []).reduce((groups, item) => {
      const kind = item.kind || "unknown";
      if (!groups[kind]) {
        groups[kind] = [];
      }
      groups[kind].push(item);
      return groups;
    }, {});
  }

  function refreshCatalogView() {
    renderCatalogList(applyCatalogFilters(state.catalogItems));
  }

  // First-class Errors screen: bounded recent errors backed by /errors.
  async function fetchErrors() {
    const target = byID("errors-list");
    try {
      const payload = await fetchJSON("/errors", { limit: 50 });
      const items = (payload && payload.items) || [];
      renderErrorCollection(target, items);
      updateNavCounts({ errors: items.length });
      setStatus("Errors loaded.", "ok");
    } catch (err) {
      renderErrorState(target, err.message);
    }
  }

  // Also seed Overview's recent-errors panel on first load (without forcing a /errors call).
  async function fetchOverviewErrors() {
    try {
      const payload = await fetchJSON("/errors", { limit: 25 });
      renderErrorCollection(byID("recent-errors"), (payload && payload.items) || []);
    } catch (err) {
      renderErrorState(byID("recent-errors"), err.message);
    }
  }

  function renderCatalogList(items) {
    const target = byID("catalog-list");
    clear(target);
    if (!items.length) {
      renderEmptyState(target, target.dataset.empty || "No catalog definitions found.");
      return;
    }
    const groups = groupCatalogItems(items);
    Object.keys(groups).sort().forEach((kind) => {
      const groupList = element("ul", { class: "catalog-kind-items" });
      groups[kind].forEach((item) => groupList.append(renderCatalogListItem(item)));
      target.append(element("li", { class: "catalog-kind-group" }, [
        element("h4", { text: labelize(kind) }),
        groupList,
      ]));
    });
  }

  function renderCatalogListItem(item) {
    const button = element("button", { type: "button", text: "Open definition" });
    const noActive = item.activeVersion === undefined || item.activeVersion === null;
    if (noActive) {
      button.disabled = true;
      button.title = "Definition has no active version";
      button.textContent = "No active version";
    } else {
      button.addEventListener("click", () => openCatalogDetail(item.kind, item.name));
    }
    const children = [
      element("div", { class: "catalog-item-header" }, [
        element("strong", { text: `${item.kind}/${item.name}` }),
        element("span", { class: "badge", text: `v${text(item.activeVersion, "—")}` }),
      ]),
      element("p", { class: "muted", text: `${item.totalVersions || 0} versions · updated ${formatDate(item.updatedAt)}` }),
    ];
    if (noActive) {
      children.push(element("p", { class: "muted warning", text: "Definition has no active version; select an explicit version outside this active-detail view." }));
    }
    children.push(button);
    return element("li", { class: noActive ? "catalog-item no-active" : "catalog-item" }, children);
  }

  async function openCatalogDetail(kind, name, version) {
    try {
      const params = { kind, name };
      if (version) {
        params.version = version;
      }
      const detail = await fetchJSON("/catalog/detail", params);
      renderCatalogDetail(detail || {});
      setStatus(`Loaded catalog ${kind}/${name}.`, "ok");
    } catch (err) {
      const target = byID("catalog-detail");
      if (target) {
        target.textContent = err.message;
        target.classList.add("empty-state");
      }
      setStatus(err.message, "error");
    }
  }

  function renderCatalogDetail(detail) {
    const target = byID("catalog-detail");
    if (!target) {
      return;
    }
    target.classList.remove("empty-state");
    clear(target);
    const versionList = element("div", { class: "version-list" });
    (detail.versions || []).forEach((version) => {
      const button = element("button", { type: "button", class: "secondary", text: `v${version.version}` });
      button.addEventListener("click", () => openCatalogDetail(detail.kind, detail.name, version.version));
      versionList.append(button);
    });
    target.append(
      element("h4", { text: `${text(detail.kind)}/${text(detail.name)} v${text(detail.version)}` }),
      element("p", { class: "muted", text: `Active v${text(detail.activeVersion)} · checksum ${text(detail.checksum)} · created ${formatDate(detail.createdAt)}` }),
      versionList,
      element("h5", { text: "Frontmatter" }),
      element("pre", { class: "json-block", text: JSON.stringify(detail.frontmatter || {}, null, 2) }),
      element("h5", { text: "Body" }),
      element("pre", { class: "catalog-body", text: detail.body || "" })
    );
  }

  // ==========================================================================
  // Hash routing
  // ==========================================================================
  function parseHash() {
    const raw = (window.location.hash || ROUTES.OVERVIEW).replace(/^#/, "");
    const parts = raw.split("/").filter(Boolean);
    if (parts[0] === "runs" && parts[1] && parts[2]) {
      return { name: "run-detail", kind: decodeURIComponent(parts[1]), id: decodeURIComponent(parts[2]) };
    }
    if (parts[0] === "runs") {
      return { name: "runs" };
    }
    if (parts[0] === "run-detail") {
      // Explicit run-detail route without a selection — show empty state.
      return { name: "run-detail" };
    }
    if (parts[0] === "catalog") {
      return { name: "catalog" };
    }
    if (parts[0] === "errors") {
      return { name: "errors" };
    }
    return { name: "overview" };
  }

  function navigateTo(hash) {
    if (window.location.hash === hash) {
      // Force a re-route even when hash unchanged (e.g., re-clicking active tab).
      handleRouteChange();
      return;
    }
    window.location.hash = hash;
  }

  function activateSection(name) {
    const target = SECTION_BY_ROUTE[name] || "overview";
    document.querySelectorAll("[data-dashboard-section]").forEach((node) => {
      node.hidden = node.dataset.dashboardSection !== target;
    });
  }

  function updateNavState(name) {
    const target = SECTION_BY_ROUTE[name] || "overview";
    document.querySelectorAll("[data-nav-link], [data-tab-link]").forEach((node) => {
      const linkName = node.dataset.navLink || node.dataset.tabLink;
      if (linkName === target) {
        node.setAttribute("aria-current", "page");
      } else {
        node.removeAttribute("aria-current");
      }
    });
  }

  function refreshRunDetailNav() {
    const enabled = !!state.selectedRun;
    document.querySelectorAll("[data-nav-run-detail]").forEach((node) => {
      if (enabled) {
        node.removeAttribute("disabled");
        node.removeAttribute("aria-disabled");
        node.dataset.route = `#/runs/${encodeURIComponent(state.selectedRun.kind)}/${encodeURIComponent(state.selectedRun.id)}`;
        node.title = `${state.selectedRun.kind} / ${state.selectedRun.id}`;
      } else {
        node.setAttribute("disabled", "true");
        node.setAttribute("aria-disabled", "true");
        node.dataset.route = ROUTES.RUN_DETAIL;
        node.title = "Select a run first";
      }
    });
    const hint = document.querySelector("[data-run-detail-hint]");
    if (hint) {
      hint.textContent = enabled ? state.selectedRun.id : "Select a run first";
    }
  }

  function rememberSelectedRun(kind, id) {
    try {
      window.sessionStorage.setItem("lazyai.dashboard.selectedRun", JSON.stringify({ kind, id }));
    } catch (_err) { /* sessionStorage unavailable */ }
  }

  function restoreSelectedRun() {
    try {
      const raw = window.sessionStorage.getItem("lazyai.dashboard.selectedRun");
      if (!raw) return;
      const parsed = JSON.parse(raw);
      if (parsed && parsed.kind && parsed.id) {
        state.selectedRun = parsed;
      }
    } catch (_err) { /* ignore */ }
  }

  function updateNavCounts(counts) {
    if (counts && counts.runs !== undefined) {
      const node = byID("nav-runs-count");
      if (node) node.textContent = String(counts.runs);
    }
    if (counts && counts.catalog !== undefined && counts.catalog !== null) {
      const node = byID("nav-catalog-count");
      if (node) node.textContent = String(counts.catalog);
    }
    if (counts && counts.errors !== undefined) {
      const node = byID("nav-errors-count");
      if (node) node.textContent = String(counts.errors);
    }
  }

  function handleRouteChange() {
    const route = parseHash();
    state.activeRoute = route;
    activateSection(route.name);
    updateNavState(route.name);

    if (route.name === "run-detail" && route.kind && route.id) {
      // Deep-link into a run.
      openRunDetail(route.kind, route.id);
    } else if (route.name === "run-detail" && state.selectedRun) {
      // Re-enter run-detail with stored selection.
      openRunDetail(state.selectedRun.kind, state.selectedRun.id);
    } else if (route.name === "errors") {
      fetchErrors();
    } else if (route.name === "catalog" && state.catalogItems.length === 0) {
      fetchCatalog();
    } else if (route.name === "runs") {
      // Refresh runs list each time the screen is opened.
      fetchRuns();
    }
  }

  // ==========================================================================
  // Tweaks (theme / density / nav)
  // ==========================================================================
  function loadTweaks() {
    try {
      const raw = window.localStorage.getItem(TWEAK_KEY);
      if (!raw) return Object.assign({}, TWEAK_DEFAULTS);
      const parsed = JSON.parse(raw);
      const merged = Object.assign({}, TWEAK_DEFAULTS, parsed || {});
      return merged;
    } catch (_err) {
      return Object.assign({}, TWEAK_DEFAULTS);
    }
  }

  function saveTweaks() {
    try {
      window.localStorage.setItem(TWEAK_KEY, JSON.stringify(state.tweaks));
    } catch (_err) { /* localStorage unavailable */ }
  }

  function applyTweaks() {
    document.body.dataset.theme = state.tweaks.theme;
    document.body.dataset.density = state.tweaks.density;
    document.body.dataset.nav = state.tweaks.nav;
    TWEAK_FIELDS.forEach((field) => {
      document.querySelectorAll(`#tweaks-panel input[name="${field}"]`).forEach((input) => {
        input.checked = input.value === state.tweaks[field];
      });
    });
  }

  function bindTweaks() {
    state.tweaks = loadTweaks();
    applyTweaks();
    const toggle = byID("tweaks-toggle");
    const panel = byID("tweaks-panel");
    const closeBtn = byID("tweaks-close");
    if (toggle && panel) {
      toggle.addEventListener("click", () => {
        const next = !!panel.hidden;
        panel.hidden = !next;
        toggle.setAttribute("aria-expanded", next ? "true" : "false");
      });
    }
    if (closeBtn && panel) {
      closeBtn.addEventListener("click", () => {
        panel.hidden = true;
        if (toggle) toggle.setAttribute("aria-expanded", "false");
      });
    }
    TWEAK_FIELDS.forEach((field) => {
      document.querySelectorAll(`#tweaks-panel input[name="${field}"]`).forEach((input) => {
        input.addEventListener("change", () => {
          if (!input.checked) return;
          state.tweaks[field] = input.value;
          applyTweaks();
          saveTweaks();
        });
      });
    });
  }

  // ==========================================================================
  // Wiring
  // ==========================================================================
  function bindNavClicks() {
    document.querySelectorAll("[data-route]").forEach((node) => {
      node.addEventListener("click", (event) => {
        const route = node.dataset.route;
        if (!route) return;
        if (node.disabled || node.getAttribute("aria-disabled") === "true") {
          event.preventDefault();
          return;
        }
        event.preventDefault();
        navigateTo(route);
      });
    });
  }

  function bindEvents() {
    const overviewRefresh = byID("refresh-overview");
    if (overviewRefresh) {
      overviewRefresh.addEventListener("click", fetchOverview);
    }
    const runsRefresh = byID("refresh-runs");
    if (runsRefresh) {
      runsRefresh.addEventListener("click", fetchRuns);
    }
    const errorsRefresh = byID("refresh-errors");
    if (errorsRefresh) {
      errorsRefresh.addEventListener("click", fetchErrors);
    }
    const runFiltersForm = byID("run-filters");
    if (runFiltersForm) {
      runFiltersForm.addEventListener("submit", (event) => {
        event.preventDefault();
        fetchRuns();
      });
    }
    const runSearch = byID("run-search");
    if (runSearch) {
      let debounce;
      runSearch.addEventListener("input", () => {
        window.clearTimeout(debounce);
        debounce = window.setTimeout(() => fetchRuns(), 200);
      });
    }
    const loadMore = byID("run-load-more");
    if (loadMore) {
      loadMore.addEventListener("click", () => fetchRuns({ append: true }));
    }
    bindAttentionChips();
    const catalogRefresh = byID("refresh-catalog");
    if (catalogRefresh) {
      catalogRefresh.addEventListener("click", fetchCatalog);
    }
    const catalogFiltersForm = byID("catalog-filters");
    if (catalogFiltersForm) {
      catalogFiltersForm.addEventListener("submit", (event) => {
        event.preventDefault();
        refreshCatalogView();
      });
    }
    const catalogKind = byID("catalog-kind-filter");
    if (catalogKind) {
      catalogKind.addEventListener("change", refreshCatalogView);
    }
    const catalogSearch = byID("catalog-search");
    if (catalogSearch) {
      catalogSearch.addEventListener("input", refreshCatalogView);
    }
    const catalogSort = byID("catalog-sort");
    if (catalogSort) {
      catalogSort.addEventListener("change", refreshCatalogView);
    }
    window.addEventListener("hashchange", handleRouteChange);
    window.addEventListener("beforeunload", closeRunEvents);
  }

  // ==========================================================================
  // Live activity feed (global SSE)
  // ==========================================================================
  function setActivityStatus(message) {
    const target = byID("activity-status");
    if (target) {
      target.textContent = message;
    }
  }

  function renderActivityEmpty() {
    const target = byID("activity-feed");
    if (!target) return;
    if (target.children.length === 0) {
      const empty = element("li", { class: "activity-empty", text: target.dataset.activityEmpty || "No activity yet." });
      target.append(empty);
    }
  }

  function dropActivityEmpty() {
    const target = byID("activity-feed");
    if (!target) return;
    target.querySelectorAll(".activity-empty").forEach((node) => node.remove());
  }

  function appendActivityEvent(event) {
    const target = byID("activity-feed");
    if (!target) return;
    if (event.id && state.activity.seenIDs.has(event.id)) return;
    if (event.id) state.activity.seenIDs.add(event.id);
    if (event.id && event.id > state.activity.lastEventID) {
      state.activity.lastEventID = event.id;
    }

    dropActivityEmpty();

    const level = ACTIVITY_LEVEL_BY_TYPE[event.eventType] || "info";
    const glyph = ACTIVITY_GLYPH_BY_LEVEL[level] || "·";
    const item = element("li", { class: "activity-item entering" }, [
      element("span", { class: "activity-item-time", text: formatActivityTime(event.createdAt) }),
      element("span", { class: "activity-item-glyph", "data-level": level, text: glyph }),
      activityMessage(event),
      element("span", { class: "activity-item-run", text: shortRunID(event.runId) }),
    ]);
    target.insertBefore(item, target.firstChild);
    window.setTimeout(() => item.classList.remove("entering"), 240);

    while (target.children.length > ACTIVITY_FEED_CAP) {
      target.lastElementChild.remove();
    }
  }

  function activityMessage(event) {
    const span = element("span", { class: "activity-item-msg" });
    const ev = element("span", { class: "activity-event", text: event.eventType || "event" });
    span.append(ev);
    const data = event.data || {};
    const detail = data.stepId || data.taskId || data.phaseId || data.code || data.message;
    if (detail) {
      span.append(element("span", { text: ` · ${detail}` }));
    }
    return span;
  }

  function shortRunID(runID) {
    if (!runID) return "";
    if (runID.length <= 8) return runID;
    return `…${runID.slice(-8)}`;
  }

  function formatActivityTime(value) {
    if (!value) return "";
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return value;
    return date.toTimeString().slice(0, 8);
  }

  function showToastForEvent(event) {
    if (!ACTIVITY_TOAST_TYPES.has(event.eventType)) return;
    if (event.eventType === "handoff_created") {
      pushToast({ kind: "warn", title: "Handoff created", msg: shortRunID(event.runId), glyph: "↯" });
      return;
    }
    const data = event.data || {};
    const title = event.eventType.replace(/_/g, " ");
    const msg = data.message || data.code || `${shortRunID(event.runId)}${data.stepId ? " · " + data.stepId : ""}`;
    pushToast({ kind: "err", title, msg, glyph: "✕" });
  }

  function pushToast({ kind = "info", title, msg, glyph = "·", duration = 6000 }) {
    let host = document.querySelector(".toasts");
    if (!host) {
      host = element("div", { class: "toasts" });
      document.body.append(host);
    }
    const toast = element("div", { class: `toast t-${kind}` });
    const closeBtn = element("button", { class: "toast-close", "aria-label": "dismiss", text: "×" });
    closeBtn.addEventListener("click", () => removeToast(toast));
    toast.append(
      element("span", { class: "toast-glyph", text: glyph }),
      element("div", { class: "toast-body" }, [
        element("span", { class: "toast-title", text: title || "" }),
        msg ? element("span", { class: "toast-msg", text: msg }) : document.createComment("no-msg"),
      ]),
      closeBtn,
    );
    host.append(toast);
    window.setTimeout(() => removeToast(toast), duration);
  }

  function removeToast(toast) {
    if (!toast || !toast.parentNode) return;
    toast.classList.add("leaving");
    window.setTimeout(() => {
      if (toast.parentNode) toast.parentNode.removeChild(toast);
    }, 220);
  }

  function scheduleOverviewRefresh() {
    if (state.activity.overviewRefreshTimer) return;
    state.activity.overviewRefreshTimer = window.setTimeout(() => {
      state.activity.overviewRefreshTimer = null;
      fetchOverview();
    }, 600);
  }

  function scheduleRunsRefresh() {
    if (state.activity.runsRefreshTimer) return;
    state.activity.runsRefreshTimer = window.setTimeout(() => {
      state.activity.runsRefreshTimer = null;
      fetchRuns();
    }, 600);
  }

  function handleGlobalEvent(event) {
    appendActivityEvent(event);
    showToastForEvent(event);
    if (RUN_STATE_CHANGE_TYPES.has(event.eventType)) {
      scheduleOverviewRefresh();
      scheduleRunsRefresh();
    }
  }

  function connectGlobalEvents() {
    if (!("EventSource" in window)) {
      setActivityStatus("EventSource unsupported · activity feed unavailable");
      return;
    }
    if (state.activity.paused) {
      setActivityStatus("paused");
      return;
    }
    disconnectGlobalEvents({ silent: true });

    const url = buildURL("/events", { since_id: state.activity.lastEventID, limit: 100 });
    const source = new EventSource(url);
    state.activity.source = source;

    source.onopen = () => {
      state.activity.reconnectAttempts = 0;
      setActivityStatus("live · listening for events across all runs");
    };
    const onMessage = (sseEvent) => {
      try {
        const parsed = JSON.parse(sseEvent.data);
        handleGlobalEvent(parsed);
      } catch (_err) {
        // Ignore unparseable frames; the next event will recover.
      }
    };
    source.onmessage = onMessage;
    Object.keys(ACTIVITY_LEVEL_BY_TYPE).forEach((type) => {
      source.addEventListener(type, onMessage);
    });
    source.onerror = () => {
      setActivityStatus("reconnecting…");
      disconnectGlobalEvents({ silent: true });
      const attempt = ++state.activity.reconnectAttempts;
      const backoff = Math.min(15000, 500 * Math.pow(2, attempt));
      state.activity.reconnectTimer = window.setTimeout(connectGlobalEvents, backoff);
    };

    renderActivityEmpty();
  }

  function disconnectGlobalEvents({ silent = false } = {}) {
    if (state.activity.source) {
      try { state.activity.source.close(); } catch (_err) { /* ignore */ }
      state.activity.source = null;
    }
    if (state.activity.reconnectTimer) {
      window.clearTimeout(state.activity.reconnectTimer);
      state.activity.reconnectTimer = null;
    }
    if (!silent) setActivityStatus("paused");
  }

  function setActivityPaused(paused) {
    state.activity.paused = paused;
    const toggle = byID("activity-pause-toggle");
    if (toggle) {
      toggle.setAttribute("aria-pressed", paused ? "true" : "false");
      const label = toggle.querySelector(".activity-pause-label");
      if (label) label.textContent = paused ? "resume" : "pause";
      toggle.title = paused ? "Resume live updates" : "Pause live updates";
    }
    const feed = byID("activity-feed");
    if (feed) feed.dataset.paused = paused ? "true" : "false";
    if (paused) {
      disconnectGlobalEvents();
    } else {
      connectGlobalEvents();
    }
    try {
      window.localStorage.setItem("lazyai.dashboard.activityPaused", paused ? "1" : "0");
    } catch (_err) { /* ignore */ }
  }

  function bindActivityFeed() {
    try {
      state.activity.paused = window.localStorage.getItem("lazyai.dashboard.activityPaused") === "1";
    } catch (_err) { /* ignore */ }
    const toggle = byID("activity-pause-toggle");
    if (toggle) {
      toggle.addEventListener("click", () => setActivityPaused(!state.activity.paused));
    }
    setActivityPaused(state.activity.paused);
    if (!state.activity.paused) connectGlobalEvents();
    renderActivityEmpty();
    window.addEventListener("beforeunload", () => disconnectGlobalEvents({ silent: true }));
  }

  function init() {
    bindTweaks();
    bindNavClicks();
    bindEvents();
    bindCopyHandlers();
    bindActivityFeed();
    restoreSelectedRun();
    refreshRunDetailNav();
    // Initial data fetches — don't block routing.
    fetchOverview();
    fetchRuns();
    fetchCatalog();
    fetchOverviewErrors();
    // Activate section based on initial hash (or default to overview).
    handleRouteChange();
  }

  init();
})();

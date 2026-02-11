const statusEl = document.getElementById("status");
const grid = document.getElementById("progressGrid");
const refreshBtn = document.getElementById("refreshBtn");
const statCourses = document.getElementById("statCourses");
const statDone = document.getElementById("statDone");
const statCompletion = document.getElementById("statCompletion");
const statScore = document.getElementById("statScore");

function setStatus(message, tone) {
    if (!statusEl) return;
    statusEl.innerHTML = message || "";
    statusEl.className = "status" + (tone ? " status--" + tone : "") + (message ? "" : " is-hidden");
}

function escapeHtml(s) {
    return String(s)
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;")
        .replaceAll("'", "&#039;");
}

function formatPercent(value) {
    if (!Number.isFinite(value)) return "0%";
    return `${Math.round(value)}%`;
}

function formatScore(value) {
    if (!Number.isFinite(value)) return "0";
    return value.toFixed(1).replace(".0", "");
}

function setStats({ courses, done, completion, score }) {
    if (statCourses) statCourses.textContent = courses ?? "—";
    if (statDone) statDone.textContent = done ?? "—";
    if (statCompletion) statCompletion.textContent = completion ?? "—";
    if (statScore) statScore.textContent = score ?? "—";
}

function renderCards(items) {
    grid.innerHTML = "";
    items.forEach((item, index) => {
        const card = document.createElement("a");
        card.className = "card";
        card.href = `/courses/${item.courseId || ""}`;
        card.style.setProperty("--i", index);

        const title = escapeHtml(item.courseTitle || "Курс");
        const completionRate = Number(item.completionRate || 0);
        const doneCount = Number(item.doneCount || 0);
        const itemsCount = Number(item.itemsCount || 0);
        const avgScore = Number(item.avgScore || 0);

        card.innerHTML = `
            <span class="badge">Прогресс</span>
            <div class="card__title">${title}</div>
            <div class="progress__meta">
                <span class="muted">Выполнено: ${doneCount} из ${itemsCount}</span>
                <span class="progress__value">${formatPercent(completionRate)}</span>
            </div>
            <div class="progress">
                <div class="progress__bar" style="width: ${Math.min(100, Math.max(0, completionRate))}%"></div>
            </div>
            <div class="card__meta">Средний балл: ${formatScore(avgScore)}</div>
            <div class="card__actions">
                <span class="btn btn--ghost btn--sm">Открыть курс</span>
            </div>
        `;

        grid.appendChild(card);
    });
}

async function loadProgress() {
    if (!grid) return;
    grid.innerHTML = "";
    setStatus("Загрузка...", "info");
    setStats({ courses: "—", done: "—", completion: "—", score: "—" });

    try {
        const res = await fetch("/me/progress?format=json", {
            headers: { "Accept": "application/json" },
            cache: "no-store"
        });

        if (res.status === 401) {
            setStatus("Нужна авторизация. <a href=\"/auth\">Войти</a>", "error");
            return;
        }

        if (!res.ok) {
            setStatus(`Ошибка загрузки: ${res.status}`, "error");
            return;
        }

        const items = await res.json();
        if (!Array.isArray(items) || items.length === 0) {
            setStatus("Пока нет прогресса по курсам.", "info");
            setStats({ courses: 0, done: 0, completion: "0%", score: "0" });
            return;
        }

        const totals = items.reduce(
            (acc, item) => {
                const itemsCount = Number(item.itemsCount || 0);
                const doneCount = Number(item.doneCount || 0);
                const avgScore = Number(item.avgScore || 0);
                acc.courses += 1;
                acc.items += itemsCount;
                acc.done += doneCount;
                acc.scoreSum += avgScore * itemsCount;
                return acc;
            },
            { courses: 0, items: 0, done: 0, scoreSum: 0 }
        );

        const completion = totals.items > 0 ? (totals.done / totals.items) * 100 : 0;
        const avgScore = totals.items > 0 ? totals.scoreSum / totals.items : 0;

        setStats({
            courses: totals.courses,
            done: totals.done,
            completion: formatPercent(completion),
            score: formatScore(avgScore)
        });

        setStatus("", "");
        renderCards(items);
    } catch (e) {
        setStatus("Ошибка подключения к API.", "error");
    }
}

if (refreshBtn) {
    refreshBtn.addEventListener("click", loadProgress);
}

loadProgress();

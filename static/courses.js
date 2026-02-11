const API_BASE = "";

const coursesGrid = document.getElementById("coursesGrid");
const statusEl = document.getElementById("status");
const filterSummary = document.getElementById("filterSummary");
const searchInput = document.getElementById("searchInput");
const categoryInput = document.getElementById("categoryInput");
const teacherInput = document.getElementById("teacherInput");
const sortSelect = document.getElementById("sortSelect");
const prevBtn = document.getElementById("prevBtn");
const nextBtn = document.getElementById("nextBtn");
const pageLabel = document.getElementById("pageLabel");

let currentPage = 1;
let lastPage = 1;

function escapeHtml(s) {
    return String(s)
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;")
        .replaceAll("'", "&#039;");
}

function truncate(text, maxLen) {
    const value = String(text || "");
    if (value.length <= maxLen) return value;
    return value.slice(0, maxLen).trimEnd() + "...";
}

function setStatus(message, tone) {
    statusEl.textContent = message || "";
    statusEl.className = "status" + (tone ? " status--" + tone : "") + (message ? "" : " is-hidden");
}

function updateSummary(total, filters) {
    if (!filterSummary) return;
    const chips = [];

    if (filters.search) {
        chips.push(`<span class="chip">Поиск: ${escapeHtml(filters.search)}</span>`);
    }
    if (filters.category) {
        chips.push(`<span class="chip">Категория: ${escapeHtml(filters.category)}</span>`);
    }
    if (filters.teacherId) {
        chips.push(`<span class="chip">Преподаватель: ${escapeHtml(filters.teacherId)}</span>`);
    }

    if (!chips.length) {
        chips.push("<span class=\"chip\">Фильтры не заданы</span>");
    }

    chips.push(`<span class="chip chip--accent">Найдено: ${total}</span>`);
    filterSummary.innerHTML = chips.join("");
}

function readParams() {
    const params = new URLSearchParams(window.location.search);
    searchInput.value = params.get("search") || "";
    categoryInput.value = params.get("category") || "";
    teacherInput.value = params.get("teacherId") || "";
    sortSelect.value = params.get("sort") || "createdAt_desc";
    currentPage = Math.max(1, parseInt(params.get("page") || "1", 10));
}

function updateUrl(params) {
    const query = params.toString();
    const nextUrl = query ? `${window.location.pathname}?${query}` : window.location.pathname;
    window.history.replaceState(null, "", nextUrl);
}

function buildParams(page) {
    const search = searchInput.value.trim();
    const category = categoryInput.value.trim();
    const teacherId = teacherInput.value.trim();
    const sort = sortSelect.value;

    const params = new URLSearchParams();
    params.set("page", page);
    params.set("limit", 9);
    params.set("format", "json");
    if (search) params.set("search", search);
    if (category) params.set("category", category);
    if (teacherId) params.set("teacherId", teacherId);
    if (sort) params.set("sort", sort);

    return { params, filters: { search, category, teacherId } };
}

async function loadCourses(page = 1) {
    const { params, filters } = buildParams(page);
    updateUrl(params);

    setStatus("Загрузка...", "info");
    coursesGrid.innerHTML = "";

    try {
        const res = await fetch(`${API_BASE}/courses?${params.toString()}`, {
            headers: { "Accept": "application/json" }
        });

        if (!res.ok) {
            const errText = await res.text();
            setStatus(errText || `Ошибка: ${res.status}`, "error");
            return;
        }

        const data = await res.json();
        const items = data.items || [];
        const total = data.total || 0;

        if (items.length === 0) {
            setStatus("Курсы не найдены.", "info");
        } else {
            setStatus("", "");
        }

        updateSummary(total, filters);

        items.forEach((course, index) => {
            const card = document.createElement("a");
            card.className = "card";
            card.href = `/courses/${course.id}`;
            card.style.setProperty("--i", index);

            const title = escapeHtml(course.title || "Без названия");
            const category = escapeHtml(course.category || "Без категории");
            const description = escapeHtml(truncate(course.description || "Описание пока не заполнено.", 140));
            const id = escapeHtml(course.id || "");

            card.innerHTML = `
                <span class="badge">${category}</span>
                <div class="card__title">${title}</div>
                <div class="card__desc">${description}</div>
                <div class="card__meta">ID: ${id}</div>
                <div class="card__actions">
                    <span class="btn btn--ghost btn--sm">Открыть курс</span>
                </div>
            `;
            coursesGrid.appendChild(card);
        });

        currentPage = data.page || page;
        lastPage = Math.max(1, Math.ceil(total / (data.limit || 9)));
        pageLabel.textContent = `Страница ${currentPage} из ${lastPage}`;
        prevBtn.disabled = currentPage <= 1;
        nextBtn.disabled = currentPage >= lastPage;
    } catch (e) {
        setStatus("Ошибка подключения к API.", "error");
    }
}

function resetAndLoad() {
    currentPage = 1;
    loadCourses(1);
}

prevBtn.addEventListener("click", () => {
    if (currentPage > 1) loadCourses(currentPage - 1);
});

nextBtn.addEventListener("click", () => {
    if (currentPage < lastPage) loadCourses(currentPage + 1);
});

[searchInput, categoryInput, teacherInput].forEach((el) => {
    el.addEventListener("input", () => {
        clearTimeout(el._t);
        el._t = setTimeout(resetAndLoad, 400);
    });
});

sortSelect.addEventListener("change", resetAndLoad);

document.getElementById("searchBtn").addEventListener("click", resetAndLoad);

document.getElementById("resetBtn").addEventListener("click", () => {
    searchInput.value = "";
    categoryInput.value = "";
    teacherInput.value = "";
    sortSelect.value = "createdAt_desc";
    resetAndLoad();
});

readParams();
loadCourses(currentPage);

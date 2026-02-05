const API_BASE = "";

const coursesGrid = document.getElementById("coursesGrid");
const statusEl = document.getElementById("status");
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

async function loadCourses(page = 1) {
    const search = searchInput.value.trim();
    const category = categoryInput.value.trim();
    const teacherId = teacherInput.value.trim();
    const sort = sortSelect.value;

    const params = new URLSearchParams();
    params.set("page", page);
    params.set("limit", 9);
    if (search) params.set("search", search);
    if (category) params.set("category", category);
    if (teacherId) params.set("teacherId", teacherId);
    if (sort) params.set("sort", sort);

    statusEl.textContent = "Загрузка...";
    coursesGrid.innerHTML = "";

    try {
        const res = await fetch(`${API_BASE}/courses?${params.toString()}`, {
            headers: { "Accept": "application/json" }
        });

        if (!res.ok) {
            const errText = await res.text();
            statusEl.textContent = `Ошибка: ${errText}`;
            return;
        }

        const data = await res.json();
        const items = data.items || [];
        const total = data.total || 0;

        if (items.length === 0) {
            statusEl.textContent = "Курсы не найдены.";
        } else {
            statusEl.textContent = "";
        }

        items.forEach(course => {
            const card = document.createElement("a");
            card.className = "card";
            card.href = `/courses/${course.id}`;
            card.innerHTML = `
                <div class="card__title">${escapeHtml(course.title)}</div>
                <div class="card__meta">Категория: ${escapeHtml(course.category || "—")}</div>
                <div class="card__meta">ID: ${escapeHtml(course.id)}</div>
            `;
            coursesGrid.appendChild(card);
        });

        currentPage = data.page || page;
        lastPage = Math.max(1, Math.ceil(total / (data.limit || 9)));
        pageLabel.textContent = `Страница ${currentPage} из ${lastPage}`;
        prevBtn.disabled = currentPage <= 1;
        nextBtn.disabled = currentPage >= lastPage;
    } catch (e) {
        statusEl.textContent = "Ошибка подключения к API.";
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

[searchInput, categoryInput, teacherInput].forEach(el => {
    el.addEventListener("input", () => {
        clearTimeout(el._t);
        el._t = setTimeout(resetAndLoad, 300);
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

loadCourses();

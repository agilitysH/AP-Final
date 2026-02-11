const API_BASE = "";

const statusEl = document.getElementById("status");
const grid = document.getElementById("coursesGrid");
const reloadBtn = document.getElementById("reloadBtn");
const statCourses = document.getElementById("statCourses");

function setStatus(message, tone) {
    if (!statusEl) return;
    statusEl.textContent = message || "";
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

function truncate(text, maxLen) {
    const value = String(text || "");
    if (value.length <= maxLen) return value;
    return value.slice(0, maxLen).trimEnd() + "...";
}

async function loadCourses() {
    if (!grid || !statusEl) return;
    grid.innerHTML = "";
    setStatus("Загрузка...", "info");

    try {
        const res = await fetch(`${API_BASE}/courses?limit=6&format=json`, {
            headers: { "Accept": "application/json" }
        });

        if (res.status === 401) {
            statusEl.innerHTML = "Нужна авторизация. <a href=\"/auth\">Войти</a>";
            statusEl.className = "status status--error";
            return;
        }

        if (!res.ok) {
            setStatus(`Ошибка загрузки: ${res.status}`, "error");
            return;
        }

        const data = await res.json();
        const courses = data.items || [];
        if (statCourses) statCourses.textContent = data.total ?? courses.length;

        if (!Array.isArray(courses) || courses.length === 0) {
            setStatus("Курсов пока нет.", "info");
            return;
        }

        setStatus("", "");

        courses.forEach((course, index) => {
            const card = document.createElement("article");
            card.className = "card";
            card.style.setProperty("--i", index);
            const title = escapeHtml(course.title || "Без названия");
            const category = escapeHtml(course.category || "Без категории");
            const description = escapeHtml(truncate(course.description || "Описание пока не заполнено.", 120));
            const id = escapeHtml(course.id || "");

            card.innerHTML = `
                <span class="badge">${category}</span>
                <div class="card__title">${title}</div>
                <div class="card__desc">${description}</div>
                <div class="card__meta">ID: ${id}</div>
                <div class="card__actions">
                    <a class="btn btn--primary btn--sm" href="/courses/${id}">Открыть</a>
                    <button class="btn btn--ghost btn--sm" type="button" data-action="edit">Изменить</button>
                    <button class="btn btn--danger btn--sm" type="button" data-action="delete">Удалить</button>
                </div>
            `;

            card.querySelector("[data-action='edit']").addEventListener("click", () => {
                editCourse(course.id, course.title || "");
            });

            card.querySelector("[data-action='delete']").addEventListener("click", () => {
                deleteCourse(course.id);
            });

            grid.appendChild(card);
        });
    } catch (e) {
        setStatus("Ошибка подключения к API.", "error");
    }
}

async function deleteCourse(id) {
    if (!confirm("Вы уверены, что хотите удалить этот курс?")) return;

    try {
        const res = await fetch(`${API_BASE}/courses/${id}`, { method: "DELETE" });
        if (res.ok) {
            await loadCourses();
        } else {
            const errText = await res.text();
            setStatus(errText || "Ошибка удаления.", "error");
        }
    } catch (e) {
        setStatus("Ошибка сети.", "error");
    }
}

async function editCourse(id, currentTitle) {
    const newTitle = prompt("Введите новое название курса:", currentTitle);
    if (!newTitle || newTitle === currentTitle) return;

    try {
        const res = await fetch(`${API_BASE}/courses/${id}`, {
            method: "PATCH",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ title: newTitle })
        });

        if (res.ok) {
            await loadCourses();
        } else {
            const errText = await res.text();
            setStatus(errText || "Ошибка при обновлении.", "error");
        }
    } catch (e) {
        setStatus("Ошибка сети.", "error");
    }
}

if (reloadBtn) {
    reloadBtn.addEventListener("click", loadCourses);
}

loadCourses();

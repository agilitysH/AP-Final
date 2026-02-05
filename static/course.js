const statusEl = document.getElementById("status");
const titleEl = document.getElementById("courseTitle");
const metaEl = document.getElementById("courseMeta");
const modulesEl = document.getElementById("modules");

function escapeHtml(s) {
    return String(s)
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;")
        .replaceAll("'", "&#039;");
}

function getCourseId() {
    const parts = window.location.pathname.split("/").filter(Boolean);
    return parts[1] || "";
}

async function loadCourse() {
    const courseId = getCourseId();
    if (!courseId) {
        statusEl.textContent = "Неверный курс.";
        return;
    }

    statusEl.textContent = "Загрузка...";
    modulesEl.innerHTML = "";

    try {
        const res = await fetch(`/courses/${courseId}`, { headers: { "Accept": "application/json" } });
        if (!res.ok) {
            statusEl.textContent = `Ошибка загрузки: ${res.status}`;
            return;
        }

        const course = await res.json();
        titleEl.textContent = course.title || "Курс";
        metaEl.textContent = `Категория: ${course.category || "—"} | ID: ${course.id}`;

        if (!course.modules || course.modules.length === 0) {
            statusEl.textContent = "Модулей пока нет.";
            return;
        }

        statusEl.textContent = "";

        course.modules.forEach(module => {
            const section = document.createElement("section");
            section.className = "module";
            section.innerHTML = `
                <h3>${escapeHtml(module.title || "Модуль")}</h3>
                <div class="module__meta">Порядок: ${module.order ?? 0}</div>
            `;

            const list = document.createElement("div");
            list.className = "items";

            (module.items || []).forEach(item => {
                const row = document.createElement("div");
                row.className = "item";
                row.innerHTML = `
                    <div>
                        <div class="item__title">${escapeHtml(item.title || "Элемент")}</div>
                        <div class="item__meta">Тип: ${escapeHtml(item.type || "—")} | MaxScore: ${item.maxScore ?? 0}</div>
                    </div>
                    <button class="btn" data-item-id="${item.id}">Update progress</button>
                `;
                row.querySelector("button").addEventListener("click", () => updateProgress(course.id, item.id));
                list.appendChild(row);
            });

            section.appendChild(list);
            modulesEl.appendChild(section);
        });
    } catch (e) {
        statusEl.textContent = "Ошибка подключения к API.";
    }
}

async function updateProgress(courseId, itemId) {
    const status = prompt("Статус (not_started | in_progress | done):", "done");
    if (!status) return;
    const scoreRaw = prompt("Score (число):", "0");
    const score = Number(scoreRaw || 0);

    try {
        const res = await fetch(`/courses/${courseId}/items/${itemId}/progress`, {
            method: "PUT",
            headers: { "Content-Type": "application/json", "Accept": "application/json" },
            body: JSON.stringify({ status, score })
        });

        if (res.ok) {
            alert("Progress updated");
        } else {
            const err = await res.text();
            alert(`Ошибка: ${err}`);
        }
    } catch (e) {
        alert("Ошибка сети");
    }
}

loadCourse();

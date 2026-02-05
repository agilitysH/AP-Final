const API_BASE = "";

async function loadCourses() {
    const status = document.getElementById("status");
    const grid = document.getElementById("coursesGrid");
    grid.innerHTML = "";
    status.textContent = "Загрузка...";

    try {
        const res = await fetch(`${API_BASE}/courses?limit=50`, {
            headers: { "Accept": "application/json" }
        });

        if (res.status === 401) {
            status.textContent = "Ошибка: Войдите в систему (Unauthorized).";
            return;
        }

        if (!res.ok) {
            status.textContent = `Ошибка: ${res.status}`;
            return;
        }

        const data = await res.json();
        const courses = data.items || [];
        if (!Array.isArray(courses) || courses.length === 0) {
            status.textContent = "Курсов пока нет.";
            return;
        }

        status.textContent = "";

        courses.forEach(c => {
            const card = document.createElement("div");
            card.className = "card";
            card.innerHTML = `
                <div class="card__title">${escapeHtml(c.title ?? "Без названия")}</div>
                <div class="card__meta">ID: ${c.id}</div>
                <div style="margin-top: 12px; display: flex; gap: 8px;">
                    <button class="btn" onclick="editCourse('${c.id}', '${escapeHtml(c.title)}')">Изменить</button>
                    <button class="btn" style="color: #ff5b5b; border-color: rgba(255,91,91,0.3)" onclick="deleteCourse('${c.id}')">Удалить</button>
                </div>
            `;
            grid.appendChild(card);
        });
    } catch (e) {
        status.textContent = "Ошибка подключения к API.";
    }
}

async function deleteCourse(id) {
    if (!confirm("Вы уверены, что хотите удалить этот курс?")) return;

    try {
        const res = await fetch(`${API_BASE}/courses/${id}`, { method: "DELETE" });
        if (res.ok) {
            loadCourses();
        } else {
            const errText = await res.text();
            alert("Ошибка удаления: " + errText);
        }
    } catch (e) {
        alert("Ошибка сети");
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
            loadCourses();
        } else {
            alert("Ошибка при обновлении");
        }
    } catch (e) {
        alert("Ошибка сети");
    }
}

function escapeHtml(s) {
    return s.replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;").replaceAll('"', "&quot;").replaceAll("'", "&#039;");
}

document.getElementById("reloadBtn").addEventListener("click", loadCourses);
loadCourses();

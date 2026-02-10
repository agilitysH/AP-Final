const statusEl = document.getElementById("status");
const titleEl = document.getElementById("courseTitle");
const descEl = document.getElementById("courseDescription");
const metaEl = document.getElementById("courseMeta");
const modulesEl = document.getElementById("modules");
const courseStatusEl = document.getElementById("courseStatus");
const enrollBtn = document.getElementById("enrollBtn");

let currentCourseId = "";

function escapeHtml(s) {
    return String(s)
        .replaceAll("&", "&amp;")
        .replaceAll("<", "&lt;")
        .replaceAll(">", "&gt;")
        .replaceAll('"', "&quot;")
        .replaceAll("'", "&#039;");
}

function setStatus(el, message, tone) {
    if (!el) return;
    el.textContent = message || "";
    el.className = "status" + (tone ? " status--" + tone : "") + (message ? "" : " is-hidden");
}

function setRowStatus(row, message, tone) {
    const el = row.querySelector(".item__status");
    if (!el) return;
    el.textContent = message || "";
    el.className = "item__status" + (tone ? " item__status--" + tone : "");
}

function getCourseId() {
    const parts = window.location.pathname.split("/").filter(Boolean);
    return parts[1] || "";
}

async function loadCourse() {
    const courseId = getCourseId();
    currentCourseId = courseId;

    if (enrollBtn) {
        enrollBtn.disabled = true;
        enrollBtn.textContent = "Записаться";
    }

    if (!courseId) {
        setStatus(statusEl, "Неверный курс.", "error");
        return;
    }

    setStatus(statusEl, "Загрузка...", "info");
    modulesEl.innerHTML = "";

    try {
        const res = await fetch(`/courses/${courseId}?format=json`, {
            headers: { "Accept": "application/json" },
            cache: "no-store"
        });
        if (!res.ok) {
            setStatus(statusEl, `Ошибка загрузки: ${res.status}`, "error");
            return;
        }

        const course = await res.json();
        titleEl.textContent = course.title || "Курс";
        descEl.textContent = course.description || "Описание курса пока не заполнено.";
        const teacherMeta = course.teacherId ? ` • Преподаватель: ${course.teacherId}` : "";
        metaEl.textContent = `Категория: ${course.category || "—"} • ID: ${course.id}${teacherMeta}`;

        if (enrollBtn) {
            enrollBtn.disabled = false;
        }

        if (!course.modules || course.modules.length === 0) {
            setStatus(statusEl, "Модулей пока нет.", "info");
            return;
        }

        setStatus(statusEl, "", "");

        course.modules.forEach((module, moduleIndex) => {
            const section = document.createElement("section");
            section.className = "module";
            section.style.setProperty("--i", moduleIndex);
            section.innerHTML = `
                <h3>${escapeHtml(module.title || "Модуль")}</h3>
                <div class="module__meta">Порядок: ${module.order ?? 0}</div>
            `;

            const list = document.createElement("div");
            list.className = "items";
            const items = module.items || [];

            if (items.length === 0) {
                const empty = document.createElement("div");
                empty.className = "item__meta";
                empty.textContent = "В модуле пока нет элементов.";
                list.appendChild(empty);
            }

            items.forEach((item) => {
                const row = document.createElement("div");
                row.className = "item";
                row.dataset.maxScore = item.maxScore ?? 0;
                const maxScore = Number(item.maxScore ?? 0);

                row.innerHTML = `
                    <div>
                        <div class="item__title">${escapeHtml(item.title || "Элемент")}</div>
                        <div class="item__meta">Тип: ${escapeHtml(item.type || "—")} · MaxScore: ${maxScore}</div>
                    </div>
                    <div class="item__controls">
                        <label class="field">
                            <span class="field__label">Статус</span>
                            <select class="input input--sm" data-field="status">
                                <option value="not_started">not_started</option>
                                <option value="in_progress">in_progress</option>
                                <option value="done">done</option>
                            </select>
                        </label>
                        <label class="field">
                            <span class="field__label">Score</span>
                            <input class="input input--sm" type="number" min="0" step="0.1" max="${maxScore}" value="0" data-field="score">
                        </label>
                        <button class="btn btn--primary btn--sm" type="button" data-action="save">Сохранить</button>
                        <div class="item__status muted"></div>
                    </div>
                `;

                row.querySelector("[data-action='save']").addEventListener("click", () => {
                    updateProgress(course.id, item.id, row);
                });

                list.appendChild(row);
            });

            section.appendChild(list);
            modulesEl.appendChild(section);
        });
    } catch (e) {
        setStatus(statusEl, "Ошибка подключения к API.", "error");
    }
}

async function enrollInCourse(courseId) {
    if (!courseId) return;

    if (enrollBtn) enrollBtn.disabled = true;
    setStatus(courseStatusEl, "Записываем на курс...", "info");

    try {
        const res = await fetch("/enrollments", {
            method: "POST",
            headers: { "Content-Type": "application/json", "Accept": "application/json" },
            body: JSON.stringify({ courseId })
        });

        if (res.ok) {
            setStatus(courseStatusEl, "Вы записаны на курс.", "success");
            if (enrollBtn) enrollBtn.textContent = "Вы записаны";
            return;
        }

        if (res.status === 401) {
            setStatus(courseStatusEl, "Нужна авторизация для записи.", "error");
        } else if (res.status === 409) {
            setStatus(courseStatusEl, "Вы уже записаны на этот курс.", "info");
            if (enrollBtn) enrollBtn.textContent = "Вы записаны";
        } else {
            const errText = await res.text();
            setStatus(courseStatusEl, errText || "Ошибка записи на курс.", "error");
        }
    } catch (e) {
        setStatus(courseStatusEl, "Ошибка сети.", "error");
    } finally {
        if (enrollBtn && enrollBtn.textContent !== "Вы записаны") {
            enrollBtn.disabled = false;
        }
    }
}

async function updateProgress(courseId, itemId, row) {
    const statusSelect = row.querySelector("[data-field='status']");
    const scoreInput = row.querySelector("[data-field='score']");
    const saveBtn = row.querySelector("[data-action='save']");
    const maxScore = Number(row.dataset.maxScore || 0);
    const score = Number(scoreInput.value || 0);

    if (Number.isNaN(score) || score < 0) {
        setRowStatus(row, "Введите корректный score.", "error");
        return;
    }

    if (score > maxScore) {
        setRowStatus(row, `Score не может быть больше ${maxScore}.`, "error");
        return;
    }

    saveBtn.disabled = true;
    setRowStatus(row, "Сохраняем...", "info");

    try {
        const res = await fetch(`/courses/${courseId}/items/${itemId}/progress`, {
            method: "PUT",
            headers: { "Content-Type": "application/json", "Accept": "application/json" },
            body: JSON.stringify({ status: statusSelect.value, score })
        });

        if (res.ok) {
            setRowStatus(row, "Сохранено.", "success");
            return;
        }

        if (res.status === 401) {
            setRowStatus(row, "Нужна авторизация.", "error");
            setStatus(courseStatusEl, "Авторизуйтесь, чтобы обновлять прогресс.", "error");
            return;
        }

        const errText = await res.text();
        setRowStatus(row, errText || "Ошибка сохранения.", "error");
    } catch (e) {
        setRowStatus(row, "Ошибка сети.", "error");
    } finally {
        saveBtn.disabled = false;
    }
}

if (enrollBtn) {
    enrollBtn.addEventListener("click", () => enrollInCourse(currentCourseId));
}

loadCourse();

const navUser = document.getElementById("navUser");

if (navUser) {
    fetch("/me", { headers: { "Accept": "application/json" }, cache: "no-store" })
        .then((res) => (res.ok ? res.json() : null))
        .then((data) => {
            if (!data || !data.username) return;
            navUser.textContent = data.username;
            navUser.title = data.username;
        })
        .catch(() => {});
}

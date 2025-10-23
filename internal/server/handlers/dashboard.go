package handlers

import (
	"fmt"
	"net/http"

	"github.com/mdhender/ottomat/ent/user"
	"github.com/mdhender/ottomat/internal/server/middleware"
)

func Dashboard(w http.ResponseWriter, r *http.Request) {
	u, ok := middleware.GetUser(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if u.Role == user.RoleAdmin {
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
		return
	}

	clanID := "N/A"
	if u.ClanID != nil {
		clanID = fmt.Sprintf("%d", *u.ClanID)
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Dashboard - OttoMat</title>
    <script src="https://unpkg.com/htmx.org@2.0.4"></script>
    <script src="https://unpkg.com/alpinejs@3.14.8/dist/cdn.min.js" defer></script>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-900 text-white min-h-screen">
    <div class="container mx-auto p-8">
        <div class="bg-gray-800 p-8 rounded-lg shadow-lg">
            <h1 class="text-3xl font-bold mb-6">Chief Dashboard</h1>
            <div class="mb-6">
                <p class="text-lg">Welcome, <span class="font-semibold">%s</span></p>
                <p class="text-lg">Clan Number: <span class="font-semibold">%s</span></p>
            </div>
            <form hx-post="/logout" hx-swap="none">
                <button type="submit" 
                    class="bg-red-600 hover:bg-red-700 text-white font-medium py-2 px-4 rounded transition">
                    Logout
                </button>
            </form>
        </div>
    </div>
</body>
</html>`, u.Username, clanID)

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

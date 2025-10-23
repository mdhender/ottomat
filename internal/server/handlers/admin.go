package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/mdhender/ottomat/ent"
	"github.com/mdhender/ottomat/ent/user"
	"github.com/mdhender/ottomat/internal/server/middleware"
	"golang.org/x/crypto/bcrypt"
)

func AdminDashboard(client *ent.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := middleware.GetUser(r.Context())
		if !ok || u.Role != user.RoleAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		ctx := r.Context()
		users, err := client.User.Query().All(ctx)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		userRows := ""
		for _, usr := range users {
			clanID := "N/A"
			if usr.ClanID != nil {
				clanID = fmt.Sprintf("%d", *usr.ClanID)
			}
			userRows += fmt.Sprintf(`
            <tr class="border-b border-gray-700">
                <td class="py-3 px-4">%d</td>
                <td class="py-3 px-4">%s</td>
                <td class="py-3 px-4">%s</td>
                <td class="py-3 px-4">%s</td>
                <td class="py-3 px-4">
                    <button hx-delete="/admin/users/%d" hx-confirm="Are you sure?" hx-target="closest tr" hx-swap="outerHTML"
                        class="bg-red-600 hover:bg-red-700 text-white font-medium py-1 px-3 rounded transition text-sm">
                        Delete
                    </button>
                </td>
            </tr>`, usr.ID, usr.Username, usr.Role, clanID, usr.ID)
		}

		html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Admin Dashboard - OttoMat</title>
    <script src="https://unpkg.com/htmx.org@2.0.4"></script>
    <script src="https://unpkg.com/alpinejs@3.14.8/dist/cdn.min.js" defer></script>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-900 text-white min-h-screen">
    <div class="container mx-auto p-8">
        <div class="bg-gray-800 p-8 rounded-lg shadow-lg mb-6">
            <div class="flex justify-between items-center mb-6">
                <h1 class="text-3xl font-bold">Admin Dashboard</h1>
                <form hx-post="/logout" hx-swap="none">
                    <button type="submit" 
                        class="bg-red-600 hover:bg-red-700 text-white font-medium py-2 px-4 rounded transition">
                        Logout
                    </button>
                </form>
            </div>
            
            <div class="mb-8">
                <h2 class="text-xl font-semibold mb-4">Add New User</h2>
                <form hx-post="/admin/users" hx-target="#users-table tbody" hx-swap="beforeend" class="grid grid-cols-5 gap-4">
                    <input type="text" name="username" placeholder="Username" required
                        class="px-3 py-2 bg-gray-700 border border-gray-600 rounded focus:outline-none focus:border-blue-500">
                    <input type="password" name="password" placeholder="Password" required
                        class="px-3 py-2 bg-gray-700 border border-gray-600 rounded focus:outline-none focus:border-blue-500">
                    <select name="role" required
                        class="px-3 py-2 bg-gray-700 border border-gray-600 rounded focus:outline-none focus:border-blue-500">
                        <option value="guest">Guest</option>
                        <option value="chief">Chief</option>
                        <option value="admin">Admin</option>
                    </select>
                    <input type="number" name="clan_id" placeholder="Clan ID (optional)"
                        class="px-3 py-2 bg-gray-700 border border-gray-600 rounded focus:outline-none focus:border-blue-500">
                    <button type="submit" 
                        class="bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 rounded transition">
                        Add User
                    </button>
                </form>
            </div>

            <div>
                <h2 class="text-xl font-semibold mb-4">Users</h2>
                <table id="users-table" class="w-full">
                    <thead>
                        <tr class="border-b border-gray-600">
                            <th class="py-3 px-4 text-left">ID</th>
                            <th class="py-3 px-4 text-left">Username</th>
                            <th class="py-3 px-4 text-left">Role</th>
                            <th class="py-3 px-4 text-left">Clan ID</th>
                            <th class="py-3 px-4 text-left">Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        %s
                    </tbody>
                </table>
            </div>
        </div>
    </div>
</body>
</html>`, userRows)

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, html)
	}
}

func CreateUser(client *ent.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := middleware.GetUser(r.Context())
		if !ok || u.Role != user.RoleAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")
		roleStr := r.FormValue("role")
		clanIDStr := r.FormValue("clan_id")

		ctx := r.Context()

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		create := client.User.
			Create().
			SetUsername(username).
			SetPasswordHash(string(passwordHash)).
			SetRole(user.Role(roleStr))

		if clanIDStr != "" {
			clanID, err := strconv.Atoi(clanIDStr)
			if err == nil {
				create.SetClanID(clanID)
			}
		}

		newUser, err := create.Save(ctx)
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusBadRequest)
			return
		}

		clanID := "N/A"
		if newUser.ClanID != nil {
			clanID = fmt.Sprintf("%d", *newUser.ClanID)
		}

		html := fmt.Sprintf(`
        <tr class="border-b border-gray-700">
            <td class="py-3 px-4">%d</td>
            <td class="py-3 px-4">%s</td>
            <td class="py-3 px-4">%s</td>
            <td class="py-3 px-4">%s</td>
            <td class="py-3 px-4">
                <button hx-delete="/admin/users/%d" hx-confirm="Are you sure?" hx-target="closest tr" hx-swap="outerHTML"
                    class="bg-red-600 hover:bg-red-700 text-white font-medium py-1 px-3 rounded transition text-sm">
                    Delete
                </button>
            </td>
        </tr>`, newUser.ID, newUser.Username, newUser.Role, clanID, newUser.ID)

		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, html)
	}
}

func DeleteUser(client *ent.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := middleware.GetUser(r.Context())
		if !ok || u.Role != user.RoleAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		err = client.User.DeleteOneID(id).Exec(ctx)
		if err != nil {
			http.Error(w, "Failed to delete user", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

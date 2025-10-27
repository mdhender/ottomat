package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/mdhender/ottomat"
	"github.com/mdhender/ottomat/ent"
	"github.com/mdhender/ottomat/ent/user"
	"github.com/mdhender/ottomat/internal/server/middleware"
	"github.com/mdhender/ottomat/internal/views"
	"golang.org/x/crypto/bcrypt"
)

func AdminDashboard(client *ent.Client, view views.Loader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s: entered\n", r.Method, r.URL.Path)
		u, ok := middleware.GetUser(r.Context())
		if !ok || u.Role != user.RoleAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		ctx := r.Context()
		users, err := client.User.Query().All(ctx)
		if err != nil {
			log.Printf("%s %s: query %v\n", r.Method, r.URL.Path, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		type userRow struct {
			ID       string
			Username string
			Role     string
			ClanID   string
			UserID   string
		}
		payload := struct {
			UserRows []userRow
			Version  string
		}{
			Version: ottomat.Version().String(),
		}
		for _, usr := range users {
			row := userRow{
				ID:       fmt.Sprintf("%d", usr.ID),
				Username: usr.Username,
				Role:     usr.Role.String(),
				ClanID:   "N/A",
				UserID:   fmt.Sprintf("%d", usr.ID),
			}
			if usr.ClanID != nil {
				row.ClanID = fmt.Sprintf("%04d", *usr.ClanID)
			}
			payload.UserRows = append(payload.UserRows, row)
		}
		name := "pages/admin/dashboard"
		buf, err := view.Execute(name, payload)
		if err != nil {
			log.Printf("%s %s: %s: data %+v\n", r.Method, r.URL.Path, name, payload)
			log.Printf("%s %s: %s: render %v\n", r.Method, r.URL.Path, name, err)
			http.Error(w, fmt.Sprintf("%s %s: %s: view error: %v", r.Method, r.URL.Path, name, err), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(buf.Bytes())
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

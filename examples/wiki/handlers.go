package main

import (
	"context"
	"html/template"
	"net/http"
	"time"
)

type WikiHandlers struct {
	auth         *AuthService
	wiki         *WikiService
	templates    *template.Template
}

func NewWikiHandlers(auth *AuthService, wiki *WikiService) *WikiHandlers {
	tmpl := template.Must(template.ParseGlob("templates/*.html"))
	return &WikiHandlers{
		auth:      auth,
		wiki:      wiki,
		templates: tmpl,
	}
}

func (h *WikiHandlers) HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	data := map[string]interface{}{
		"Title": "Wiki Home",
	}

	// Check for session cookie to determine if user is logged in
	if cookie, err := r.Cookie("session"); err == nil {
		if user, err := h.auth.ValidateSession(cookie.Value); err == nil && user != nil {
			data["User"] = user
			data["LoggedIn"] = true
		}
	}

	h.templates.ExecuteTemplate(w, "home.html", data)
}

func (h *WikiHandlers) LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.templates.ExecuteTemplate(w, "login.html", map[string]interface{}{
			"Title": "Login",
		})
		return
	}

	if r.Method == "POST" {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		_, sessionID, err := h.auth.Login(username, password)
		if err != nil {
			h.templates.ExecuteTemplate(w, "login.html", map[string]interface{}{
				"Title": "Login",
				"Error": "Invalid username or password",
			})
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    sessionID,
			Path:     "/",
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (h *WikiHandlers) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		h.templates.ExecuteTemplate(w, "register.html", map[string]interface{}{
			"Title": "Register",
		})
		return
	}

	if r.Method == "POST" {
		r.ParseForm()
		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirm_password")

		if password != confirmPassword {
			h.templates.ExecuteTemplate(w, "register.html", map[string]interface{}{
				"Title": "Register",
				"Error": "Passwords do not match",
			})
			return
		}

		_, err := h.auth.Register(username, email, password)
		if err != nil {
			h.templates.ExecuteTemplate(w, "register.html", map[string]interface{}{
				"Title": "Register",
				"Error": "Registration failed: " + err.Error(),
			})
			return
		}

		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func (h *WikiHandlers) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		h.auth.Logout(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *WikiHandlers) ViewPageHandler(w http.ResponseWriter, r *http.Request) {
	pageTitle := r.URL.Query().Get("title")
	if pageTitle == "" {
		http.Error(w, "Page title required", http.StatusBadRequest)
		return
	}

	// Check for session cookie to determine if user is logged in
	var currentUser *User
	if cookie, err := r.Cookie("session"); err == nil {
		if user, err := h.auth.ValidateSession(cookie.Value); err == nil && user != nil {
			currentUser = user
		}
	}

	page, err := h.wiki.GetPageByTitle(pageTitle)
	if err != nil {
		data := map[string]interface{}{
			"Title":     "Page Not Found",
			"PageTitle": pageTitle,
		}
		if currentUser != nil {
			data["User"] = currentUser
			data["LoggedIn"] = true
		}
		h.templates.ExecuteTemplate(w, "page_not_found.html", data)
		return
	}

	data := map[string]interface{}{
		"Title": page.Title,
		"Page":  page,
	}
	if currentUser != nil {
		data["User"] = currentUser
		data["LoggedIn"] = true
	}

	h.templates.ExecuteTemplate(w, "view_page.html", data)
}

func (h *WikiHandlers) EditPageHandler(w http.ResponseWriter, r *http.Request) {
	user := getUserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "GET" {
		pageTitle := r.URL.Query().Get("title")
		if pageTitle == "" {
			http.Error(w, "Page title required", http.StatusBadRequest)
			return
		}

		page, err := h.wiki.GetPageByTitle(pageTitle)
		data := map[string]interface{}{
			"Title":     "Edit Page",
			"PageTitle": pageTitle,
			"User":      user,
			"LoggedIn":  true,
		}

		if err == nil {
			data["Page"] = page
			data["IsEdit"] = true
		} else {
			data["IsEdit"] = false
			data["Page"] = &WikiPage{Title: pageTitle, Content: ""}
		}

		h.templates.ExecuteTemplate(w, "edit_page.html", data)
		return
	}

	if r.Method == "POST" {
		r.ParseForm()
		pageTitle := r.FormValue("title")
		content := r.FormValue("content")
		pageID := r.FormValue("page_id")

		if pageID != "" {
			_, err := h.wiki.UpdatePage(pageID, pageTitle, content)
			if err != nil {
				http.Error(w, "Failed to update page", http.StatusInternalServerError)
				return
			}
		} else {
			_, err := h.wiki.CreatePage(pageTitle, content, user.ID)
			if err != nil {
				http.Error(w, "Failed to create page", http.StatusInternalServerError)
				return
			}
		}

		http.Redirect(w, r, "/page?title="+pageTitle, http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func getUserFromContext(ctx context.Context) *User {
	if ctx == nil {
		return nil
	}

	if user, ok := ctx.Value("user").(*User); ok {
		return user
	}
	return nil
}
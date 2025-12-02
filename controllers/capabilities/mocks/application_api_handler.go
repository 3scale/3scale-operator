package mocks

import (
	"encoding/json"
	"net/http"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/3scale/3scale-porta-go-client/client"
)

type ApplicationAPIHandler struct {
	mux      *http.ServeMux
	services map[int64]*client.ApplicationPlanJSONList
	accounts map[int64]*client.ApplicationList
}

type ApplicationAPIHandlerOpt func(h *ApplicationAPIHandler)

func WithService(service int64, plans *client.ApplicationPlanJSONList) ApplicationAPIHandlerOpt {
	return func(m *ApplicationAPIHandler) {
		m.services[service] = plans
	}
}

func WithAccount(account int64, applications *client.ApplicationList) ApplicationAPIHandlerOpt {
	return func(m *ApplicationAPIHandler) {
		m.accounts[account] = applications
	}
}

func NewApplicationAPIHandler(opts ...ApplicationAPIHandlerOpt) *ApplicationAPIHandler {
	handler := &ApplicationAPIHandler{
		accounts: make(map[int64]*client.ApplicationList),
		services: make(map[int64]*client.ApplicationPlanJSONList),
		mux:      http.NewServeMux(),
	}

	for _, opt := range opts {
		opt(handler)
	}

	handler.registerRoutes()
	return handler
}

func (m *ApplicationAPIHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mux.ServeHTTP(w, r)
}

func (m *ApplicationAPIHandler) registerRoutes() {
	m.mux.HandleFunc("GET /admin/api/services/{serviceID}/application_plans.json", m.applicationPlanHandler)
	m.mux.HandleFunc("GET /admin/api/accounts/{accountID}/applications.json", m.applicationHandler)
	m.mux.HandleFunc("POST /admin/api/accounts/{accountID}/applications.json", m.applicationHandler)
	m.mux.HandleFunc("DELETE /admin/api/accounts/{accountID}/applications/{applicationID}", m.applicationHandler)
	m.mux.HandleFunc("PUT /admin/api/accounts/{accountID}/applications/{applicationID}", m.applicationHandler)
	m.mux.HandleFunc("PUT /admin/api/accounts/{accountID}/applications/{applicationID}/resume.json", m.applicationHandler)
	m.mux.HandleFunc("PUT /admin/api/accounts/{accountID}/applications/{applicationID}/suspend.json", m.applicationHandler)
	m.mux.HandleFunc("PUT /admin/api/accounts/{accountID}/applications/{applicationID}/change_plan.json", m.applicationHandler)
}

func (m *ApplicationAPIHandler) applicationPlanHandler(w http.ResponseWriter, r *http.Request) {
	serviceIDParam := r.PathValue("serviceID")

	serviceID, err := strconv.ParseInt(serviceIDParam, 10, 64)
	if err != nil {
		http.Error(w, serviceIDParam, http.StatusInternalServerError)
		return
	}

	applicationPlans, ok := m.services[serviceID]
	if !ok {
		errorResponse(w, "error", []string{"Access Denied"}, http.StatusForbidden)
		return
	}

	json, err := json.Marshal(applicationPlans)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(json)
}

func (m *ApplicationAPIHandler) applicationHandler(w http.ResponseWriter, r *http.Request) {
	accountIDParam := r.PathValue("accountID")

	accountID, err := strconv.ParseInt(accountIDParam, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	applications, ok := m.accounts[accountID]
	if !ok {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		json, err := json.Marshal(applications)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		_, _ = w.Write(json)
	case http.MethodPost:
		_ = r.ParseForm()
		planIDField := r.FormValue("plan_id")

		planID, err := strconv.ParseInt(planIDField, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		found := false
		var serviceID int64
		for id, service := range m.services {
			for _, plan := range service.Plans {
				if plan.Element.ID == planID {
					serviceID = id
					found = true
					break
				}
			}
		}

		if !found {
			http.NotFound(w, r)
			return
		}

		application := client.Application{
			ID:            3,
			State:         "live",
			UserAccountID: accountID,
			ServiceID:     serviceID,
			PlanID:        planID,
			AppName:       r.FormValue("name"),
			Description:   r.FormValue("description"),
		}

		elem := client.ApplicationElem{Application: application}

		applications.Applications = append(applications.Applications, elem)

		json, err := json.Marshal(elem)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(json)

	case http.MethodDelete:
		appIDParam := strings.TrimSuffix(r.PathValue("applicationID"), ".json")
		appID, err := strconv.ParseInt(appIDParam, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for idx, app := range applications.Applications {
			if app.Application.ID == appID {
				applications.Applications = slices.Delete(applications.Applications, idx, idx+1)
				w.WriteHeader(http.StatusOK)
				return
			}
		}

		http.NotFound(w, r)
		return

	case http.MethodPut:
		appIDParam := strings.TrimSuffix(r.PathValue("applicationID"), ".json")
		appID, err := strconv.ParseInt(appIDParam, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		switch path.Base(r.URL.Path) {
		case "resume.json":
			for _, app := range applications.Applications {
				if app.Application.ID == appID {
					app.Application.State = "live"

					json, err := json.Marshal(app)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					w.Header().Add("Content-Type", "application/json")
					_, _ = w.Write(json)
					return
				}
			}
		case "suspend.json":
			for _, app := range applications.Applications {
				if app.Application.ID == appID {
					app.Application.State = "suspended"

					json, err := json.Marshal(app.Application)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					w.Header().Add("Content-Type", "application/json")
					_, _ = w.Write(json)
					return
				}
			}
		case "change_plan.json":
			planIDField := r.FormValue("plan_id")

			planID, err := strconv.ParseInt(planIDField, 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			found := false

			for _, service := range m.services {
				for _, plan := range service.Plans {
					if plan.Element.ID == planID {
						found = true
						break
					}
				}
			}

			if !found {
				http.NotFound(w, r)
				return
			}

			for _, app := range applications.Applications {
				if app.Application.ID == appID {
					app.Application.PlanID = planID

					json, err := json.Marshal(app)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					w.Header().Add("Content-Type", "application/json")
					_, _ = w.Write(json)
					return
				}
			}
		default:
			for _, app := range applications.Applications {
				if app.Application.ID == appID {
					appName := r.FormValue("name")

					description := r.FormValue("description")

					if appName != "" {
						app.Application.AppName = appName
					}

					if description != "" {
						app.Application.Description = description
					}

					json, err := json.Marshal(app)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
						return
					}

					w.Header().Add("Content-Type", "application/json")
					_, _ = w.Write(json)
					return
				}
			}

			http.NotFound(w, r)
			return
		}
	}
}

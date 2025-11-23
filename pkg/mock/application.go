package mock

import (
	"encoding/json"
	"net/http"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/3scale/3scale-porta-go-client/client"
)

type mockApplicationServer struct {
	serviceID           int64
	accountID           int64
	applicationPlanList *client.ApplicationPlanJSONList
	applicationsList    *client.ApplicationList
}

func NewApplicationMockServer(accountID, serviceID int64, applicationPlanList *client.ApplicationPlanJSONList, application client.Application) http.Handler {
	appList := client.ApplicationList{
		Applications: []client.ApplicationElem{
			{Application: application},
		},
	}

	srv := &mockApplicationServer{
		serviceID:           serviceID,
		accountID:           accountID,
		applicationPlanList: applicationPlanList,
		applicationsList:    &appList,
	}

	handler := srv.GetMux()

	return handler
}

func (m *mockApplicationServer) GetMux() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /admin/api/services/{serviceID}/application_plans.json", m.applicationPlanHandler)
	mux.HandleFunc("GET /admin/api/accounts/{accountID}/applications.json", m.applicationHandler)
	mux.HandleFunc("POST /admin/api/accounts/{accountID}/applications.json", m.applicationHandler)
	mux.HandleFunc("DELETE /admin/api/accounts/{accountID}/applications/{applicationID}", m.applicationHandler)
	mux.HandleFunc("PUT /admin/api/accounts/{accountID}/applications/{applicationID}", m.applicationHandler)
	mux.HandleFunc("PUT /admin/api/accounts/{accountID}/applications/{applicationID}/resume.json", m.applicationHandler)
	mux.HandleFunc("PUT /admin/api/accounts/{accountID}/applications/{applicationID}/suspend.json", m.applicationHandler)
	mux.HandleFunc("PUT /admin/api/accounts/{accountID}/applications/{applicationID}/change_plan.json", m.applicationHandler)

	return mux
}

func (m *mockApplicationServer) applicationHandler(w http.ResponseWriter, r *http.Request) {
	accountIDParam := r.PathValue("accountID")

	accountID, err := strconv.ParseInt(accountIDParam, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if accountID != m.accountID {
		http.NotFound(w, r)
		return
	}

	switch r.Method {
	case http.MethodGet:
		json, err := json.Marshal(m.applicationsList)
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

		for _, plan := range m.applicationPlanList.Plans {
			if plan.Element.ID == planID {
				found = true
				break
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
			ServiceID:     0,
			PlanID:        planID,
			AppName:       r.FormValue("name"),
			Description:   r.FormValue("description"),
		}

		elem := client.ApplicationElem{Application: application}

		m.applicationsList.Applications = append(m.applicationsList.Applications, elem)

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

		for idx, app := range m.applicationsList.Applications {
			if app.Application.ID == appID {
				m.applicationsList.Applications = slices.Delete(m.applicationsList.Applications, idx, idx+1)
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
			for _, app := range m.applicationsList.Applications {
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
			for _, app := range m.applicationsList.Applications {
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

			for _, plan := range m.applicationPlanList.Plans {
				if plan.Element.ID == planID {
					found = true
					break
				}
			}

			if !found {
				http.NotFound(w, r)
				return
			}

			for _, app := range m.applicationsList.Applications {
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
			for _, app := range m.applicationsList.Applications {
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
		}

		http.NotFound(w, r)
		return
	}
}

func (m *mockApplicationServer) applicationPlanHandler(w http.ResponseWriter, r *http.Request) {
	serviceIDParam := r.PathValue("serviceID")

	serviceID, err := strconv.ParseInt(serviceIDParam, 10, 64)
	if err != nil {
		http.Error(w, serviceIDParam, http.StatusInternalServerError)
		return
	}

	if serviceID != m.serviceID {
		errorResponse(w, "error", []string{"Access Denied"}, http.StatusForbidden)
		return
	}

	json, err := json.Marshal(m.applicationPlanList)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	_, _ = w.Write(json)
}

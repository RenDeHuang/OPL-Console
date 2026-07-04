package api

import (
	"encoding/json"
	"net/http"

	"github.com/RenDeHuang/opl-console/internal/ledger"
)

type topUpRequest struct {
	AccountID        string `json:"accountId"`
	BillingAccountID string `json:"billingAccountId"`
	Amount           int64  `json:"amount"`
	AmountFen        int64  `json:"amountFen"`
	Reason           string `json:"reason"`
	Note             string `json:"note"`
}

func mountBillingRoutes(router Router, deps Dependencies) {
	router.Post("/api/billing/topups", func(w http.ResponseWriter, r *http.Request) {
		session, ok := requireAdmin(w, r, deps)
		if !ok {
			return
		}
		if deps.Ledger == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "ledger_not_configured"})
			return
		}
		var payload topUpRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid_json"})
			return
		}
		billingAccountID := payload.BillingAccountID
		if billingAccountID == "" {
			billingAccountID = payload.AccountID
		}
		amountFen := payload.AmountFen
		if amountFen == 0 {
			amountFen = payload.Amount
		}
		note := payload.Note
		if note == "" {
			note = payload.Reason
		}
		topUpID, err := apiID("topup")
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "topup_id_failed"})
			return
		}
		if err := deps.Ledger.RecordManualTopUp(r.Context(), ledger.TopUpRequest{
			TopUpID:          topUpID,
			BillingAccountID: billingAccountID,
			AmountFen:        amountFen,
			ActorUserID:      session.User.ID,
			Note:             note,
		}); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "billing_topup_failed"})
			return
		}
		wallet, err := deps.Ledger.GetWallet(r.Context(), billingAccountID)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]any{"id": topUpID, "billingAccountId": billingAccountID})
			return
		}
		writeJSON(w, http.StatusOK, wallet)
	})

	router.Post("/api/billing/request-usage", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireOwner(w, r, deps); !ok {
			return
		}
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "request_usage_not_implemented"})
	})

	router.Post("/api/billing/reconciliation", func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireAdmin(w, r, deps); !ok {
			return
		}
		writeJSON(w, http.StatusNotImplemented, map[string]string{"error": "billing_reconciliation_not_implemented"})
	})
}

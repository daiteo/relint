package producthandler

import "net/http"

type ProductHandler struct{}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {}

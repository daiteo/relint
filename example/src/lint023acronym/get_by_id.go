package producthandler

import "net/http"

type ProductHandler struct{}

type GetByIDInput struct{}

type GetByIDOutput struct{}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {}

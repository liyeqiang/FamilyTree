package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"familytree/interfaces"
	"familytree/models"

	"github.com/gorilla/mux"
)

// FamilyHandler 家庭关系处理器
type FamilyHandler struct {
	service interfaces.FamilyService
}

// NewFamilyHandler 创建家庭关系处理器
func NewFamilyHandler(service interfaces.FamilyService) *FamilyHandler {
	return &FamilyHandler{service: service}
}

// CreateFamily 创建家庭关系（配偶关系）
func (h *FamilyHandler) CreateFamily(w http.ResponseWriter, r *http.Request) {
	var req models.CreateFamilyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的请求数据",
		})
		return
	}

	family, err := h.service.CreateFamily(r.Context(), &req)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    family,
		Message: "配偶关系创建成功",
	})
}

// AddSpouse 添加配偶关系
func (h *FamilyHandler) AddSpouse(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	individualID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的个人ID",
		})
		return
	}

	var req struct {
		SpouseID *int `json:"spouse_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的请求数据",
		})
		return
	}

	if req.SpouseID == nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "配偶ID不能为空",
		})
		return
	}

	family, err := h.service.AddSpouse(r.Context(), individualID, *req.SpouseID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    family,
		Message: "配偶关系添加成功",
	})
} 
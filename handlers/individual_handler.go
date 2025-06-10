package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"familytree/interfaces"
	"familytree/models"

	"github.com/gorilla/mux"
)

// IndividualHandler 个人信息处理器
type IndividualHandler struct {
	service interfaces.IndividualService
}

// NewIndividualHandler 创建个人信息处理器
func NewIndividualHandler(service interfaces.IndividualService) *IndividualHandler {
	return &IndividualHandler{service: service}
}

// CreateIndividual 创建个人信息
func (h *IndividualHandler) CreateIndividual(w http.ResponseWriter, r *http.Request) {
	var req models.CreateIndividualRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	individual, err := h.service.Create(r.Context(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(individual)
}

// GetIndividual 获取个人信息
func (h *IndividualHandler) GetIndividual(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	individual, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(individual)
}

// UpdateIndividual 更新个人信息
func (h *IndividualHandler) UpdateIndividual(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	var req models.UpdateIndividualRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	individual, err := h.service.Update(r.Context(), id, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(individual)
}

// DeleteIndividual 删除个人信息
func (h *IndividualHandler) DeleteIndividual(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// SearchIndividuals 搜索个人信息
func (h *IndividualHandler) SearchIndividuals(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20 // 默认值
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0 // 默认值
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	individuals, total, err := h.service.Search(r.Context(), query, limit, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"data":   individuals,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetChildren 获取个人的子女
func (h *IndividualHandler) GetChildren(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	children, err := h.service.GetChildren(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(children)
}

// GetParents 获取个人的父母
func (h *IndividualHandler) GetParents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	father, mother, err := h.service.GetParents(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"father": father,
		"mother": mother,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetSiblings 获取个人的兄弟姐妹
func (h *IndividualHandler) GetSiblings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	siblings, err := h.service.GetSiblings(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(siblings)
}

// GetSpouses 获取个人的配偶
func (h *IndividualHandler) GetSpouses(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	spouses, err := h.service.GetSpouses(r.Context(), id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(spouses)
}

// GetAncestors 获取个人的祖先
func (h *IndividualHandler) GetAncestors(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	generationsStr := r.URL.Query().Get("generations")
	generations := 3 // 默认3代
	if generationsStr != "" {
		if g, err := strconv.Atoi(generationsStr); err == nil && g > 0 {
			generations = g
		}
	}

	ancestors, err := h.service.GetAncestors(r.Context(), id, generations)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ancestors)
}

// GetDescendants 获取个人的后代
func (h *IndividualHandler) GetDescendants(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	generationsStr := r.URL.Query().Get("generations")
	generations := 3 // 默认3代
	if generationsStr != "" {
		if g, err := strconv.Atoi(generationsStr); err == nil && g > 0 {
			generations = g
		}
	}

	descendants, err := h.service.GetDescendants(r.Context(), id, generations)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(descendants)
}

// GetFamilyTree 获取家族树
func (h *IndividualHandler) GetFamilyTree(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	generationsStr := r.URL.Query().Get("generations")
	generations := 3 // 默认3代
	if generationsStr != "" {
		if g, err := strconv.Atoi(generationsStr); err == nil && g > 0 {
			generations = g
		}
	}

	tree, err := h.service.GetFamilyTree(r.Context(), id, generations)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tree)
} 
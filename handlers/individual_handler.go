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

// APIResponse 统一API响应格式
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Total   *int        `json:"total,omitempty"`
	Limit   *int        `json:"limit,omitempty"`
	Offset  *int        `json:"offset,omitempty"`
}

// respondJSON 统一JSON响应
func respondJSON(w http.ResponseWriter, status int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

// CreateIndividual 创建个人信息
func (h *IndividualHandler) CreateIndividual(w http.ResponseWriter, r *http.Request) {
	var req models.CreateIndividualRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的请求数据",
		})
		return
	}

	individual, err := h.service.Create(r.Context(), &req)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Data:    individual,
		Message: "创建成功",
	})
}

// GetIndividual 获取个人信息
func (h *IndividualHandler) GetIndividual(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的ID",
		})
		return
	}

	individual, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    individual,
	})
}

// UpdateIndividual 更新个人信息
func (h *IndividualHandler) UpdateIndividual(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的ID",
		})
		return
	}

	var req models.UpdateIndividualRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的请求数据",
		})
		return
	}

	individual, err := h.service.Update(r.Context(), id, &req)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    individual,
		Message: "更新成功",
	})
}

// DeleteIndividual 删除个人信息
func (h *IndividualHandler) DeleteIndividual(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的ID",
		})
		return
	}

	if err := h.service.Delete(r.Context(), id); err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "删除成功",
	})
}

// SearchIndividuals 搜索个人信息
func (h *IndividualHandler) SearchIndividuals(w http.ResponseWriter, r *http.Request) {
	// 同时支持q和query参数
	query := r.URL.Query().Get("query")
	if query == "" {
		query = r.URL.Query().Get("q")
	}
	
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
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    individuals,
		Total:   &total,
		Limit:   &limit,
		Offset:  &offset,
	})
}

// GetChildren 获取个人的子女
func (h *IndividualHandler) GetChildren(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的ID",
		})
		return
	}

	children, err := h.service.GetChildren(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    children,
	})
}

// GetParents 获取个人的父母
func (h *IndividualHandler) GetParents(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的ID",
		})
		return
	}

	father, mother, err := h.service.GetParents(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// 组装父母数据为数组
	var parents []models.Individual
	if father != nil {
		parents = append(parents, *father)
	}
	if mother != nil {
		parents = append(parents, *mother)
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    parents,
	})
}

// GetSiblings 获取个人的兄弟姐妹
func (h *IndividualHandler) GetSiblings(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的ID",
		})
		return
	}

	siblings, err := h.service.GetSiblings(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    siblings,
	})
}

// GetSpouses 获取个人的配偶
func (h *IndividualHandler) GetSpouses(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的ID",
		})
		return
	}

	spouses, err := h.service.GetSpouses(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    spouses,
	})
}

// GetAncestors 获取个人的祖先
func (h *IndividualHandler) GetAncestors(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的ID",
		})
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
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    ancestors,
	})
}

// GetDescendants 获取个人的后代
func (h *IndividualHandler) GetDescendants(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的ID",
		})
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
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    descendants,
	})
}

// GetFamilyTree 获取家族树
func (h *IndividualHandler) GetFamilyTree(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的ID",
		})
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
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    tree,
	})
} 
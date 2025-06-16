package handlers

import (
	"encoding/json"
	"familytree/interfaces"
	"familytree/models"
	"net/http"
	"strconv"

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

// GetFamiliesByHusband 根据丈夫ID获取所有家庭关系
func (h *FamilyHandler) GetFamiliesByHusband(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	husbandID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的丈夫ID",
		})
		return
	}

	families, err := h.service.GetByIndividualID(r.Context(), husbandID)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// 只返回该男性作为丈夫的家庭
	husbandFamilies := make([]models.Family, 0)
	for _, family := range families {
		if family.HusbandID != nil && *family.HusbandID == husbandID {
			husbandFamilies = append(husbandFamilies, family)
		}
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    husbandFamilies,
		Message: "获取家庭关系成功",
	})
}

// GetFamily 获取家庭关系
func (h *FamilyHandler) GetFamily(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的家庭ID",
		})
		return
	}

	family, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		respondJSON(w, http.StatusNotFound, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    family,
	})
}

// UpdateFamily 更新家庭关系
func (h *FamilyHandler) UpdateFamily(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的家庭ID",
		})
		return
	}

	var req models.CreateFamilyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的请求数据",
		})
		return
	}

	family, err := h.service.Update(r.Context(), id, &req)
	if err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Data:    family,
		Message: "更新成功",
	})
}

// DeleteFamily 删除家庭关系
func (h *FamilyHandler) DeleteFamily(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的家庭ID",
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

// AddChild 为家庭添加子女
func (h *FamilyHandler) AddChild(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	familyID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的家庭ID",
		})
		return
	}

	var req struct {
		ChildID      int    `json:"child_id"`
		Relationship string `json:"relationship"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的请求数据",
		})
		return
	}

	if err := h.service.AddChild(r.Context(), familyID, req.ChildID, req.Relationship); err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusCreated, APIResponse{
		Success: true,
		Message: "子女添加成功",
	})
}

// RemoveChild 从家庭移除子女
func (h *FamilyHandler) RemoveChild(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	familyID, err := strconv.Atoi(vars["id"])
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的家庭ID",
		})
		return
	}

	childID, err := strconv.Atoi(vars["childId"])
	if err != nil {
		respondJSON(w, http.StatusBadRequest, APIResponse{
			Success: false,
			Message: "无效的子女ID",
		})
		return
	}

	if err := h.service.RemoveChild(r.Context(), familyID, childID); err != nil {
		respondJSON(w, http.StatusInternalServerError, APIResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	respondJSON(w, http.StatusOK, APIResponse{
		Success: true,
		Message: "子女移除成功",
	})
}

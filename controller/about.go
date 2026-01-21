package controller

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"myblog-gogogo/auth"
	"myblog-gogogo/db/models"
	"myblog-gogogo/db/repositories"
)

var (
	mainCardRepo repositories.AboutMainCardRepository
	subCardRepo  repositories.AboutSubCardRepository
)

// InitAboutRepositories 初始化关于页面仓库
func InitAboutRepositories(database *sql.DB) {
	mainCardRepo = repositories.NewSQLiteAboutMainCardRepository(database)
	subCardRepo = repositories.NewSQLiteAboutSubCardRepository(database)
}

// AboutMainCardsHandler 获取所有主卡片
func AboutMainCardsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	cards, err := mainCardRepo.GetAllEnabled()
	if err != nil {
		http.Error(w, "获取主卡片失败", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(cards)
}

// AboutMainCardsAdminHandler 管理员获取所有主卡片（包括禁用的）
func AboutMainCardsAdminHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 验证管理员权限
	if !auth.IsAdmin(r) {
		http.Error(w, "无权限", http.StatusUnauthorized)
		return
	}

	cards, err := mainCardRepo.GetAll()
	if err != nil {
		http.Error(w, "获取主卡片失败", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(cards)
}

// AboutMainCardCreateHandler 创建主卡片
func AboutMainCardCreateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 验证管理员权限
	if !auth.IsAdmin(r) {
		http.Error(w, "无权限", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var card models.AboutMainCard
	if err := json.NewDecoder(r.Body).Decode(&card); err != nil {
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	if err := mainCardRepo.Create(&card); err != nil {
		http.Error(w, "创建主卡片失败", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(card)
}

// AboutMainCardUpdateHandler 更新主卡片
func AboutMainCardUpdateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 验证管理员权限
	if !auth.IsAdmin(r) {
		http.Error(w, "无权限", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	var card models.AboutMainCard
	if err := json.NewDecoder(r.Body).Decode(&card); err != nil {
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	card.ID = id
	if err := mainCardRepo.Update(&card); err != nil {
		http.Error(w, "更新主卡片失败", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(card)
}

// AboutMainCardDeleteHandler 删除主卡片
func AboutMainCardDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 验证管理员权限
	if !auth.IsAdmin(r) {
		http.Error(w, "无权限", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodDelete {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	if err := mainCardRepo.Delete(id); err != nil {
		http.Error(w, "删除主卡片失败", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "删除成功"})
}

// AboutMainCardUpdateSortHandler 更新主卡片排序
func AboutMainCardUpdateSortHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 验证管理员权限
	if !auth.IsAdmin(r) {
		http.Error(w, "无权限", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	var data struct {
		SortOrder int `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	if err := mainCardRepo.UpdateSortOrder(id, data.SortOrder); err != nil {
		http.Error(w, "更新排序失败", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "更新成功"})
}

// AboutMainCardUpdateEnabledHandler 更新主卡片启用状态
func AboutMainCardUpdateEnabledHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 验证管理员权限
	if !auth.IsAdmin(r) {
		http.Error(w, "无权限", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	var data struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	if err := mainCardRepo.UpdateEnabled(id, data.Enabled); err != nil {
		http.Error(w, "更新状态失败", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "更新成功"})
}

// AboutSubCardsHandler 获取指定主卡片的所有次卡片
func AboutSubCardsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idStr := r.URL.Query().Get("main_card_id")
	mainCardID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "无效的主卡片ID", http.StatusBadRequest)
		return
	}

	cards, err := subCardRepo.GetByMainCardIDEnabled(mainCardID)
	if err != nil {
		http.Error(w, "获取次卡片失败", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(cards)
}

// AboutSubCardsAdminHandler 管理员获取所有次卡片
func AboutSubCardsAdminHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 验证管理员权限
	if !auth.IsAdmin(r) {
		http.Error(w, "无权限", http.StatusUnauthorized)
		return
	}

	cards, err := subCardRepo.GetAll()
	if err != nil {
		http.Error(w, "获取次卡片失败", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(cards)
}

// AboutSubCardCreateHandler 创建次卡片
func AboutSubCardCreateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 验证管理员权限
	if !auth.IsAdmin(r) {
		http.Error(w, "无权限", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	var card models.AboutSubCard
	if err := json.NewDecoder(r.Body).Decode(&card); err != nil {
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	if err := subCardRepo.Create(&card); err != nil {
		http.Error(w, "创建次卡片失败", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(card)
}

// AboutSubCardUpdateHandler 更新次卡片
func AboutSubCardUpdateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 验证管理员权限
	if !auth.IsAdmin(r) {
		http.Error(w, "无权限", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	var card models.AboutSubCard
	if err := json.NewDecoder(r.Body).Decode(&card); err != nil {
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	card.ID = id
	if err := subCardRepo.Update(&card); err != nil {
		http.Error(w, "更新次卡片失败", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(card)
}

// AboutSubCardDeleteHandler 删除次卡片
func AboutSubCardDeleteHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 验证管理员权限
	if !auth.IsAdmin(r) {
		http.Error(w, "无权限", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodDelete {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	if err := subCardRepo.Delete(id); err != nil {
		http.Error(w, "删除次卡片失败", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "删除成功"})
}

// AboutSubCardUpdateSortHandler 更新次卡片排序
func AboutSubCardUpdateSortHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 验证管理员权限
	if !auth.IsAdmin(r) {
		http.Error(w, "无权限", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	var data struct {
		SortOrder int `json:"sort_order"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	if err := subCardRepo.UpdateSortOrder(id, data.SortOrder); err != nil {
		http.Error(w, "更新排序失败", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "更新成功"})
}

// AboutSubCardUpdateEnabledHandler 更新次卡片启用状态
func AboutSubCardUpdateEnabledHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// 验证管理员权限
	if !auth.IsAdmin(r) {
		http.Error(w, "无权限", http.StatusUnauthorized)
		return
	}

	if r.Method != http.MethodPut {
		http.Error(w, "方法不允许", http.StatusMethodNotAllowed)
		return
	}

	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "无效的ID", http.StatusBadRequest)
		return
	}

	var data struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	if err := subCardRepo.UpdateEnabled(id, data.Enabled); err != nil {
		http.Error(w, "更新状态失败", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "更新成功"})
}
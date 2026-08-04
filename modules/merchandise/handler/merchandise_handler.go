package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/webdevtedxuniversitasairlangga/modules/merchandise/dto"
	"github.com/webdevtedxuniversitasairlangga/modules/merchandise/service"
)

type MerchandiseHandler struct {
	service service.MerchandiseService
}

func NewMerchandiseHandler(i *do.Injector) (*MerchandiseHandler, error) {
	svc := do.MustInvoke[service.MerchandiseService](i)
	return &MerchandiseHandler{
		service: svc,
	}, nil
}

func (h *MerchandiseHandler) GetAll(c *gin.Context) {
	var filter dto.MerchandiseFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data, total, err := h.service.GetAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Get all merchandise success",
		"data":    data,
		"meta": gin.H{
			"total": total,
			"page":  filter.Page,
			"limit": filter.Limit,
		},
	})
}

func (h *MerchandiseHandler) GetByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID"})
		return
	}

	data, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Merchandise not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Success", "data": data})
}

func (h *MerchandiseHandler) Create(c *gin.Context) {
	var req dto.MerchandiseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Create(c.Request.Context(), req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Merchandise created successfully"})
}

func (h *MerchandiseHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID"})
		return
	}

	var req dto.MerchandiseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.Update(c.Request.Context(), id, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Merchandise updated successfully"})
}

func (h *MerchandiseHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid UUID"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Merchandise deleted successfully"})
}


func (h *MerchandiseHandler) AddImage(c *gin.Context) {
	merchId, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Merchandise UUID"})
		return
	}

	var req dto.MerchImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.AddImage(c.Request.Context(), merchId, req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Image added successfully"})
}

func (h *MerchandiseHandler) DeleteImage(c *gin.Context) {
	imageId, err := uuid.Parse(c.Param("imageId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Image UUID"})
		return
	}

	if err := h.service.DeleteImage(c.Request.Context(), imageId); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Image deleted successfully"})
}
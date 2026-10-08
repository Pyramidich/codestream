package handler

import (
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ilya/codestream/internal/service"
)

// FileHandler handles file HTTP requests.
type FileHandler struct {
	fileService *service.FileService
}

// NewFileHandler creates a new FileHandler.
func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{fileService: fileService}
}

// CreateFileRequest represents a create file request.
type CreateFileRequest struct {
	Name     string `json:"name" binding:"required"`
	Path     string `json:"path" binding:"required"`
	Language string `json:"language"`
}

// UpdateFileRequest represents an update file request.
type UpdateFileRequest struct {
	Name        *string `json:"name,omitempty"`
	Path        *string `json:"path,omitempty"`
	Language    *string `json:"language,omitempty"`
	Content     *[]byte `json:"content,omitempty"`
	ContentB64  *string `json:"content_base64,omitempty"`
	ContentText *string `json:"content_text,omitempty"`
}

// Create handles file creation.
func (h *FileHandler) Create(c *gin.Context) {
	userID, err := getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	var req CreateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	file, err := h.fileService.Create(c.Request.Context(), projectID, userID, req.Name, req.Path, req.Language)
	if err != nil {
		status := http.StatusUnprocessableEntity
		if err.Error() == "insufficient permissions" {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, file)
}

// List handles listing files in a project.
func (h *FileHandler) List(c *gin.Context) {
	userID, err := getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	files, err := h.fileService.ListByProject(c.Request.Context(), projectID, userID)
	if err != nil {
		status := http.StatusUnprocessableEntity
		if err.Error() == "access denied" {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}

// Get handles getting a single file.
func (h *FileHandler) Get(c *gin.Context) {
	userID, err := getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	fileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}

	file, err := h.fileService.GetByID(c.Request.Context(), fileID, userID)
	if err != nil {
		status := http.StatusNotFound
		if err.Error() == "access denied" {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, file)
}

// Update handles updating a file.
func (h *FileHandler) Update(c *gin.Context) {
	userID, err := getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	fileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}

	var req UpdateFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	updates := service.FileUpdates{
		Name:        req.Name,
		Path:        req.Path,
		Language:    req.Language,
		Content:     req.Content,
		ContentText: req.ContentText,
	}

	if req.ContentB64 != nil {
		decoded, err := base64.StdEncoding.DecodeString(*req.ContentB64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid content_base64"})
			return
		}
		updates.Content = &decoded
	}

	file, err := h.fileService.Update(c.Request.Context(), fileID, userID, updates)
	if err != nil {
		status := http.StatusUnprocessableEntity
		if err.Error() == "insufficient permissions" {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, file)
}

// GetContent returns the plain-text content of a file.
func (h *FileHandler) GetContent(c *gin.Context) {
	userID, err := getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	fileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}

	file, err := h.fileService.GetByID(c.Request.Context(), fileID, userID)
	if err != nil {
		status := http.StatusNotFound
		if err.Error() == "access denied" {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	content := ""
	if file.ContentText != nil {
		content = *file.ContentText
	}

	c.JSON(http.StatusOK, gin.H{"content": content})
}

// Delete handles deleting a file.
func (h *FileHandler) Delete(c *gin.Context) {
	userID, err := getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	fileID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}

	if err := h.fileService.Delete(c.Request.Context(), fileID, userID); err != nil {
		status := http.StatusUnprocessableEntity
		if err.Error() == "insufficient permissions" {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

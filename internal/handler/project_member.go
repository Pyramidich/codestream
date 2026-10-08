package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ilya/codestream/internal/service"
)

// ProjectMemberHandler handles project membership HTTP requests.
type ProjectMemberHandler struct {
	memberService *service.ProjectMemberService
}

// NewProjectMemberHandler creates a new ProjectMemberHandler.
func NewProjectMemberHandler(memberService *service.ProjectMemberService) *ProjectMemberHandler {
	return &ProjectMemberHandler{memberService: memberService}
}

// AddMemberRequest represents an request to add a member.
type AddMemberRequest struct {
	Email string `json:"email" binding:"required"`
	Role  string `json:"role" binding:"required"`
}

// Add handles adding a member to a project.
func (h *ProjectMemberHandler) Add(c *gin.Context) {
	currentUserID, err := getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	member, err := h.memberService.AddMember(c.Request.Context(), projectID, currentUserID, strings.ToLower(req.Email), req.Role)
	if err != nil {
		status := http.StatusUnprocessableEntity
		switch err.Error() {
		case "user not found":
			status = http.StatusNotFound
		case "user is already a member":
			status = http.StatusConflict
		case "only owner or editor can add members":
			status = http.StatusForbidden
		case "invalid role", "invalid email", "cannot add yourself":
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, member)
}

// List handles listing project members.
func (h *ProjectMemberHandler) List(c *gin.Context) {
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

	members, err := h.memberService.ListMembers(c.Request.Context(), projectID, userID)
	if err != nil {
		status := http.StatusUnprocessableEntity
		if err.Error() == "access denied" {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"members": members})
}

// Remove handles removing a member from a project.
func (h *ProjectMemberHandler) Remove(c *gin.Context) {
	ownerID, err := getCurrentUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	projectID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project id"})
		return
	}

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if err := h.memberService.RemoveMember(c.Request.Context(), projectID, ownerID, userID); err != nil {
		status := http.StatusUnprocessableEntity
		switch err.Error() {
		case "only owner can remove members":
			status = http.StatusForbidden
		case "member not found":
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

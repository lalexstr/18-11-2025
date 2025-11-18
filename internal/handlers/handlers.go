package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"test/internal/models"
	"test/internal/pdf"
	"test/internal/worker"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db     *gorm.DB
	worker *worker.Worker
}

func NewHandler(db *gorm.DB, w *worker.Worker) *Handler {
	return &Handler{db: db, worker: w}
}

func (h *Handler) Register(r *gin.Engine) {
	r.POST("/links", h.POSTLinks)
	r.GET("/report", h.GetReport)
	r.GET("/ping", func(c *gin.Context) { c.JSON(200, gin.H{"pong": true}) })
}

func (h *Handler) POSTLinks(c *gin.Context) {
	var body struct {
		Links []string `json:"links"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json"})
		return
	}
	if len(body.Links) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no links provided"})
		return
	}

	set := models.LinkSet{}
	if err := h.db.Create(&set).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot create set"})
		return
	}

	results := make([]gin.H, 0, len(body.Links))
	for _, u := range body.Links {
		ln := models.Link{
			LinkSetID: set.ID,
			URL:       strings.TrimSpace(u),
			Status:    "pending",
			Processed: false,
		}
		if err := h.db.Create(&ln).Error; err != nil {
			results = append(results, gin.H{"url": u, "status": "error"})
			continue
		}

		select {
		case h.worker.Queue <- ln.ID:
		default:

		}
		results = append(results, gin.H{"url": u, "status": "pending"})
	}
	c.JSON(http.StatusOK, gin.H{"links_num": set.ID, "results": results})
}

func (h *Handler) GetReport(c *gin.Context) {
	q := c.Query("links_num")
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "links_num required"})
		return
	}
	parts := strings.Split(q, ",")
	var ids []uint
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.ParseUint(p, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, uint(n))
	}
	if len(ids) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid ids"})
		return
	}

	// достаём ссылки из БД
	var rows []models.Link
	if err := h.db.Where("link_set_id IN ?", ids).Find(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	// генерим pdf
	buf, err := pdf.GeneratePDF(ids, rows)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "pdf generation failed"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=report.pdf")
	c.Data(200, "application/pdf", buf)
}

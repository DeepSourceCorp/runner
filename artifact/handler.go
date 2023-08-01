package artifact

import (
	"fmt"
	"io"

	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slog"
)

type Handler struct {
	storage StorageClient
	bucket  string
}

func NewHandler(storage StorageClient, bucket string) *Handler {
	return &Handler{
		storage: storage,
		bucket:  bucket,
	}
}

type AnalysisArtifactsRequest struct {
	RunID      string   `json:"run_id"`
	Shortcode  string   `json:"shortcode"`
	SnippetIDs []string `json:"snippet_ids"`
}

// HandleAnalysisArtifacts handles reading and parsing the body of the POST request
// to `/artifacts/analysis` endpoint.
func (h *Handler) HandleAnalysis(c echo.Context) error {
	ctx := c.Request().Context()

	var req AnalysisArtifactsRequest
	analysisArtifactsResponse := make(map[string]string)

	if err := c.Bind(&req); err != nil {
		return err
	}

	for _, snippetID := range req.SnippetIDs {
		slog.Info(fmt.Sprintf("fetching analysis artifacts from %s", snippetID))
		r, err := h.storage.NewReader(ctx, h.bucket, snippetID)
		if err != nil {
			return err
		}
		defer r.Close()

		raw, err := io.ReadAll(r)
		if err != nil {
			return err
		}
		analysisArtifactsResponse[snippetID] = string(raw)
	}

	return c.JSON(200, analysisArtifactsResponse)
}

type AutofixArtifactsRequest struct {
	RunID      string              `json:"run_id"`
	Shortcode  string              `json:"shortcode"`
	SnippetIDs map[string][]string `json:"snippet_ids"`
}

type AutofixResultObject struct {
	BeforeHTML string `json:"before_html"`
	AfterHTML  string `json:"after_html"`
}

// HandleAutofixArtifacts handles reading and parsing the body of the POST request
// to /artifacts/autofix endpoint.
func (h *Handler) HandleAutofix(c echo.Context) error {
	ctx := c.Request().Context()

	var req AutofixArtifactsRequest
	autofixArtifactsResponse := make(map[string]AutofixResultObject)

	if err := c.Bind(&req); err != nil {
		return err
	}

	for filename, snippetIDs := range req.SnippetIDs {
		for _, snippetID := range snippetIDs {
			obj := AutofixResultObject{}
			path := fmt.Sprintf("%s/%s/%s/before", req.RunID, filename, snippetID)
			slog.Info(fmt.Sprintf("fetching the autofix patch artifact %s from bucket %s", h.bucket, path))
			r, err := h.storage.NewReader(ctx, h.bucket, path)
			if err != nil {
				return err
			}
			defer r.Close()

			raw, err := io.ReadAll(r)
			if err != nil {
				return err
			}

			before := string(raw)
			obj.BeforeHTML = before

			path = fmt.Sprintf("%s/%s/%s/after", req.RunID, filename, snippetID)
			slog.Info(fmt.Sprintf("fetching the autofix patch artifact %s from bucket %s", h.bucket, path))
			r, err = h.storage.NewReader(ctx, h.bucket, path)
			if err != nil {
				return err
			}
			defer r.Close()

			raw, err = io.ReadAll(r)
			if err != nil {
				return err
			}

			after := string(raw)
			obj.AfterHTML = after

			autofixArtifactsResponse[snippetID] = obj
		}
	}
	return c.JSON(200, autofixArtifactsResponse)
}

package orchestrator

import (
	"log"
	"net/http"

	artifact "github.com/deepsourcelabs/artifacts/types"
	"github.com/labstack/echo/v4"
	"golang.org/x/exp/slog"
)

type Handler struct {
	analysisTask    *AnalysisTask
	autofixTask     *AutofixTask
	transformerTask *TransformerTask
	cancelCheckTask *CancelCheckTask
	patcherTask     *PatcherTask
}

func NewHandler(
	opts *TaskOpts,
	driver Driver,
	provider Provider,
	signer Signer,
) *Handler {
	slog.Info("initializing orchestrator handler", slog.Any("opts", opts))
	return &Handler{
		analysisTask:    NewAnalysisTask(opts, driver, provider, signer),
		autofixTask:     NewAutofixTask(opts, driver, provider, signer),
		transformerTask: NewTransformerTask(opts, driver, provider, signer),
		cancelCheckTask: NewCancelCheckTask(opts, driver, signer, nil),
		patcherTask:     NewPatcherTask(opts, driver, provider, signer),
	}
}

func (h *Handler) HandleAnalysis(c echo.Context) error {
	log.Println("received analysis task")
	ctx := c.Request().Context()
	run := new(artifact.AnalysisRun)
	if err := c.Bind(&run); err != nil {
		log.Println(err)
		return c.HTML(http.StatusBadRequest, err.Error())
	}
	log.Println("Running analysis task")
	if err := h.analysisTask.Run(ctx, &AnalysisRunRequest{
		Run:            run,
		AppID:          c.Param("app_id"),
		InstallationID: c.Request().Header.Get("X-Installation-ID"),
	}); err != nil {
		log.Println(err)
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	return nil
}

func (h *Handler) HandleAutofix(c echo.Context) error {
	ctx := c.Request().Context()
	run := new(artifact.AutofixRun)
	if err := c.Bind(&run); err != nil {
		return c.HTML(http.StatusBadRequest, err.Error())
	}
	if err := h.autofixTask.Run(ctx, &AutofixRunRequest{
		Run:            run,
		AppID:          c.Param("app_id"),
		InstallationID: c.Request().Header.Get("X-Installation-ID"),
	}); err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	return nil
}

func (h *Handler) HandleTransformer(c echo.Context) error {
	ctx := c.Request().Context()
	run := new(artifact.TransformerRun)
	if err := c.Bind(&run); err != nil {
		return c.HTML(http.StatusBadRequest, err.Error())
	}
	if err := h.transformerTask.Run(ctx, &TransformerRunRequest{
		Run:            run,
		AppID:          c.Param("app_id"),
		InstallationID: c.Request().Header.Get("X-Installation-ID"),
	}); err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	return nil
}

// HandleCancelCheck handles the cancel check workflow.
func (h *Handler) HandleCancelCheck(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(artifact.CancelCheckRun)
	if err := c.Bind(&req); err != nil {
		return c.HTML(http.StatusBadRequest, err.Error())
	}
	if err := h.cancelCheckTask.Run(ctx, req); err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	return nil
}

// HandlePatcher handles the patching job workflow.
func (h *Handler) HandlePatcher(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(artifact.PatcherRun)
	if err := c.Bind(&req); err != nil {
		return c.HTML(http.StatusBadRequest, err.Error())
	}

	if err := h.patcherTask.Run(ctx, &PatcherRunRequest{
		Run:            req,
		AppID:          c.Param("app_id"),
		InstallationID: c.Request().Header.Get("X-Installation-ID"),
	}); err != nil {
		return c.HTML(http.StatusInternalServerError, err.Error())
	}
	return nil
}

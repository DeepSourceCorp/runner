package orchestrator

import (
	"log"

	artifact "github.com/DeepSourceCorp/artifacts/types"
	"github.com/deepsourcecorp/runner/httperror"
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
	runner *Runner,
) *Handler {
	return &Handler{
		analysisTask:    NewAnalysisTask(runner, opts, driver, provider, signer),
		autofixTask:     NewAutofixTask(runner, opts, driver, provider, signer),
		transformerTask: NewTransformerTask(runner, opts, driver, provider, signer),
		cancelCheckTask: NewCancelCheckTask(runner, opts, driver, signer, nil),
		patcherTask:     NewPatcherTask(runner, opts, driver, provider, signer),
	}
}

func (h *Handler) HandleAnalysis(c echo.Context) error {
	log.Println("received analysis task")
	ctx := c.Request().Context()
	run := new(artifact.AnalysisRun)
	if err := c.Bind(&run); err != nil {
		slog.Error("analysis task bind error", slog.Any("err", err))
		return httperror.ErrMissingParams(err)
	}
	log.Println("Running analysis task")
	if err := h.analysisTask.Run(ctx, &AnalysisRunRequest{
		Run:            run,
		AppID:          c.Param("app_id"),
		InstallationID: c.Request().Header.Get("X-Installation-ID"),
	}); err != nil {
		slog.Error("analysis task run error", slog.Any("err", err))
		return httperror.ErrUnknown(err)
	}
	return nil
}

func (h *Handler) HandleAutofix(c echo.Context) error {
	ctx := c.Request().Context()
	run := new(artifact.AutofixRun)
	if err := c.Bind(&run); err != nil {
		slog.Error("autofix task bind error", slog.Any("err", err))
		return httperror.ErrMissingParams(err)
	}
	if err := h.autofixTask.Run(ctx, &AutofixRunRequest{
		Run:            run,
		AppID:          c.Param("app_id"),
		InstallationID: c.Request().Header.Get("X-Installation-ID"),
	}); err != nil {
		slog.Error("autofix task run error", slog.Any("err", err))
		return httperror.ErrUnknown(err)
	}
	return nil
}

func (h *Handler) HandleTransformer(c echo.Context) error {
	ctx := c.Request().Context()
	run := new(artifact.TransformerRun)
	if err := c.Bind(&run); err != nil {
		slog.Error("transformer task bind error", slog.Any("err", err))
		return httperror.ErrMissingParams(err)
	}
	if err := h.transformerTask.Run(ctx, &TransformerRunRequest{
		Run:            run,
		AppID:          c.Param("app_id"),
		InstallationID: c.Request().Header.Get("X-Installation-ID"),
	}); err != nil {
		slog.Error("transformer task run error", slog.Any("err", err))
		return httperror.ErrUnknown(err)
	}
	return nil
}

// HandleCancelCheck handles the cancel check workflow.
func (h *Handler) HandleCancelCheck(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(artifact.CancelCheckRun)
	if err := c.Bind(&req); err != nil {
		slog.Error("cancel check task bind error", slog.Any("err", err))
		return httperror.ErrMissingParams(err)
	}
	if err := h.cancelCheckTask.Run(ctx, req); err != nil {
		slog.Error("cancel check task run error", slog.Any("err", err))
		return httperror.ErrUnknown(err)
	}
	return nil
}

// HandlePatcher handles the patching job workflow.
func (h *Handler) HandlePatcher(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(artifact.PatcherRun)
	if err := c.Bind(&req); err != nil {
		slog.Error("patcher task bind error", slog.Any("err", err))
		return httperror.ErrMissingParams(err)
	}

	if err := h.patcherTask.Run(ctx, &PatcherRunRequest{
		Run:            req,
		AppID:          c.Param("app_id"),
		InstallationID: c.Request().Header.Get("X-Installation-ID"),
	}); err != nil {
		slog.Error("patcher task run error", slog.Any("err", err))
		return httperror.ErrUnknown(err)
	}
	return nil
}

package app

import (
	"avito_intership/internal/config"
	handler_bid_mux_impl "avito_intership/internal/handlers/bid/mux_impl"
	handler_tender_mux_impl "avito_intership/internal/handlers/tender/mux_impl"
	"avito_intership/pkg/logger"
	"context"
	"github.com/gorilla/mux"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"net/http"
)

type App struct {
	sp *serviceProvider

	cfg *config.Config

	router *mux.Router

	logger *slog.Logger
}

func (a *App) initLogger(_ context.Context) error {
	a.logger = logger.New()
	return nil
}

func (a *App) initConfig(_ context.Context) error {
	cfg, err := config.New()
	if err != nil {
		return err
	}

	a.cfg = cfg

	return nil
}

func (a *App) initMuxHandler(_ context.Context) error {
	a.router = mux.NewRouter()
	return nil
}

func (a *App) initBidsHandler(ctx context.Context) error {
	bidsService, err := a.sp.BidService(ctx)
	if err != nil {
		return err
	}

	if err = handler_bid_mux_impl.Register(a.router, bidsService, a.logger); err != nil {
		return err
	}

	return nil
}

func (a *App) initTenderHandler(ctx context.Context) error {
	tenderService, err := a.sp.TenderService(ctx)
	if err != nil {
		return err
	}

	if err = handler_tender_mux_impl.Register(a.router, tenderService, a.logger); err != nil {
		return err
	}

	return nil
}

func (a *App) initServiceProvider(_ context.Context) error {
	a.sp = newServiceProvider(a.cfg.DB.PostgresConnStr, a.logger)
	return nil
}

func (a *App) initDeps(ctx context.Context) error {
	deps := [...]func(ctx context.Context) error{
		a.initLogger,
		a.initConfig,
		a.initServiceProvider,
		a.initMuxHandler,
		a.initBidsHandler,
		a.initTenderHandler,
	}

	for _, f := range deps {
		if err := f(ctx); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) runHttpServer() error {
	server := &http.Server{
		Addr:      a.cfg.Address,
		Handler:   a.router,
		TLSConfig: nil,
	}

	return server.ListenAndServe()
}

func (a *App) Run(ctx context.Context) error {
	g, _ := errgroup.WithContext(ctx)

	g.Go(a.runHttpServer)

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
}

func (a *App) Stop() {
	if a.sp.feedbackService != nil {
		a.sp.feedbackRepository.CloseConn()
	}
	if a.sp.decisionRepository != nil {
		a.sp.decisionRepository.CloseConn()
	}
	if a.sp.employeeRepository != nil {
		a.sp.employeeRepository.CloseConn()
	}
	if a.sp.organizationResponsibleRepository != nil {
		a.sp.organizationResponsibleRepository.CloseConn()
	}
	if a.sp.tendersRepository != nil {
		a.sp.tendersRepository.CloseConn()
	}
	if a.sp.bidRepository != nil {
		a.sp.bidRepository.CloseConn()
	}
}

func New(ctx context.Context) (*App, error) {
	a := &App{}

	if err := a.initDeps(ctx); err != nil {
		return nil, err
	}

	return a, nil
}

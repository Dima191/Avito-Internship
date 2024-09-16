package app

import (
	repository_bid "avito_intership/internal/repository/bid"
	repository_bid_postgres "avito_intership/internal/repository/bid/postgres"
	repository_decision "avito_intership/internal/repository/decision"
	repository_decision_postgres "avito_intership/internal/repository/decision/postgres"
	repository_employee "avito_intership/internal/repository/employee"
	repository_employee_postgres "avito_intership/internal/repository/employee/postgres"
	repository_feedback "avito_intership/internal/repository/feedback"
	repository_feedback_postgres "avito_intership/internal/repository/feedback/postgres"
	repository_organization_resp "avito_intership/internal/repository/organization_responsible"
	repository_organization_resp_postgres "avito_intership/internal/repository/organization_responsible/postgres"
	repository_tenders "avito_intership/internal/repository/tender"
	repository_tenders_postgres "avito_intership/internal/repository/tender/postgres"
	service_bids "avito_intership/internal/service/bid"
	service_bids_impl "avito_intership/internal/service/bid/implementation"
	service_decision "avito_intership/internal/service/decision"
	service_decision_impl "avito_intership/internal/service/decision/implementation"
	service_employee "avito_intership/internal/service/employee"
	service_employee_impl "avito_intership/internal/service/employee/implementation"
	service_feedback "avito_intership/internal/service/feedback"
	service_feedback_impl "avito_intership/internal/service/feedback/implementation"
	service_organization_resp "avito_intership/internal/service/organization_responsible"
	service_organization_resp_impl "avito_intership/internal/service/organization_responsible/implementation"
	service_tenders "avito_intership/internal/service/tender"
	service_tenders_impl "avito_intership/internal/service/tender/implementation"
	"context"
	"log/slog"
)

type serviceProvider struct {
	feedbackRepository repository_feedback.Repository
	feedbackService    service_feedback.Service

	decisionRepository repository_decision.Repository
	decisionService    service_decision.Service

	employeeRepository repository_employee.Repository
	employeeService    service_employee.Service

	organizationResponsibleRepository repository_organization_resp.Repository
	organizationResponsibleService    service_organization_resp.Service

	tendersRepository repository_tenders.Repository
	tendersService    service_tenders.Service

	bidRepository repository_bid.Repository
	bidService    service_bids.Service

	DBConnectionStr string
	logger          *slog.Logger
}

func (sp *serviceProvider) FeedbackRepository(ctx context.Context) (repository_feedback.Repository, error) {
	if sp.feedbackRepository == nil {
		repository, err := repository_feedback_postgres.New(ctx, sp.DBConnectionStr, sp.logger)
		if err != nil {
			return nil, err
		}

		sp.feedbackRepository = repository
	}
	return sp.feedbackRepository, nil
}

func (sp *serviceProvider) FeedbackService(ctx context.Context) (service_feedback.Service, error) {
	if sp.feedbackService == nil {
		repository, err := sp.FeedbackRepository(ctx)
		if err != nil {
			return nil, err
		}

		sp.feedbackService = service_feedback_impl.New(repository, sp.logger)
	}

	return sp.feedbackService, nil
}

func (sp *serviceProvider) DecisionRepository(ctx context.Context) (repository_decision.Repository, error) {
	if sp.decisionRepository == nil {
		repository, err := repository_decision_postgres.New(ctx, sp.DBConnectionStr, sp.logger)
		if err != nil {
			return nil, err
		}

		sp.decisionRepository = repository
	}
	return sp.decisionRepository, nil
}

func (sp *serviceProvider) DecisionService(ctx context.Context) (service_decision.Service, error) {
	if sp.decisionService == nil {
		repository, err := sp.DecisionRepository(ctx)
		if err != nil {
			return nil, err
		}

		sp.decisionService = service_decision_impl.New(repository, sp.logger)
	}
	return sp.decisionService, nil
}

func (sp *serviceProvider) EmployeeRepository(ctx context.Context) (repository_employee.Repository, error) {
	if sp.employeeRepository == nil {
		repository, err := repository_employee_postgres.New(ctx, sp.DBConnectionStr, sp.logger)
		if err != nil {
			return nil, err
		}

		sp.employeeRepository = repository
	}
	return sp.employeeRepository, nil
}

func (sp *serviceProvider) EmployeeService(ctx context.Context) (service_employee.Service, error) {
	if sp.employeeService == nil {
		repository, err := sp.EmployeeRepository(ctx)
		if err != nil {
			return nil, err
		}

		sp.employeeService = service_employee_impl.New(repository, sp.logger)
	}

	return sp.employeeService, nil
}

func (sp *serviceProvider) OrganizationResponsibleRepository(ctx context.Context) (repository_organization_resp.Repository, error) {
	if sp.organizationResponsibleRepository == nil {
		repository, err := repository_organization_resp_postgres.New(ctx, sp.DBConnectionStr, sp.logger)
		if err != nil {
			return nil, err
		}

		sp.organizationResponsibleRepository = repository
	}
	return sp.organizationResponsibleRepository, nil
}

func (sp *serviceProvider) OrganizationResponsibleService(ctx context.Context) (service_organization_resp.Service, error) {
	if sp.organizationResponsibleService == nil {
		repository, err := sp.OrganizationResponsibleRepository(ctx)
		if err != nil {
			return nil, err
		}

		sp.organizationResponsibleService = service_organization_resp_impl.New(repository, sp.logger)
	}

	return sp.organizationResponsibleService, nil
}

func (sp *serviceProvider) TenderRepository(ctx context.Context) (repository_tenders.Repository, error) {
	if sp.tendersRepository == nil {
		repository, err := repository_tenders_postgres.New(ctx, sp.DBConnectionStr, sp.logger)
		if err != nil {
			return nil, err
		}

		sp.tendersRepository = repository
	}

	return sp.tendersRepository, nil
}

func (sp *serviceProvider) TenderService(ctx context.Context) (service_tenders.Service, error) {
	if sp.tendersService == nil {
		repository, err := sp.TenderRepository(ctx)
		if err != nil {
			return nil, err
		}

		employeeService, err := sp.EmployeeService(ctx)
		if err != nil {
			return nil, err
		}

		organizationRespService, err := sp.OrganizationResponsibleService(ctx)
		if err != nil {
			return nil, err
		}

		sp.tendersService = service_tenders_impl.New(repository, employeeService, organizationRespService, sp.logger)
	}

	return sp.tendersService, nil
}

func (sp *serviceProvider) BidRepository(ctx context.Context) (repository_bid.Repository, error) {
	if sp.bidRepository == nil {
		repository, err := repository_bid_postgres.New(ctx, sp.DBConnectionStr, sp.logger)
		if err != nil {
			return nil, err
		}

		sp.bidRepository = repository
	}
	return sp.bidRepository, nil
}

func (sp *serviceProvider) BidService(ctx context.Context) (service_bids.Service, error) {
	if sp.bidService == nil {
		repository, err := sp.BidRepository(ctx)
		if err != nil {
			return nil, err
		}

		feedbackService, err := sp.FeedbackService(ctx)
		if err != nil {
			return nil, err
		}

		employeeService, err := sp.EmployeeService(ctx)
		if err != nil {
			return nil, err
		}

		organizationResponsibleService, err := sp.OrganizationResponsibleService(ctx)
		if err != nil {
			return nil, err
		}

		tenderService, err := sp.TenderService(ctx)
		if err != nil {
			return nil, err
		}

		decisionService, err := sp.DecisionService(ctx)
		if err != nil {
			return nil, err
		}

		sp.bidService = service_bids_impl.New(repository, employeeService, organizationResponsibleService, tenderService, decisionService, feedbackService, sp.logger)
	}
	return sp.bidService, nil
}

func newServiceProvider(DBConnectionStr string, logger *slog.Logger) *serviceProvider {
	sp := &serviceProvider{
		DBConnectionStr: DBConnectionStr,
		logger:          logger,
	}
	return sp
}

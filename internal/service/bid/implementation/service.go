package service_bids_impl

import (
	"avito_intership/internal/model"
	repository_bid "avito_intership/internal/repository/bid"
	service_bids "avito_intership/internal/service/bid"
	service_decision "avito_intership/internal/service/decision"
	service_employee "avito_intership/internal/service/employee"
	service_feedback "avito_intership/internal/service/feedback"
	service_organization_resp "avito_intership/internal/service/organization_responsible"
	service_tenders "avito_intership/internal/service/tender"
	"context"
	"errors"
	"log/slog"
)

type service struct {
	bidsRepository repository_bid.Repository

	employeeService         service_employee.Service
	organizationRespService service_organization_resp.Service
	tenderService           service_tenders.Service
	decisionService         service_decision.Service
	feedbackService         service_feedback.Service

	logger *slog.Logger
}

var (
	tenderClosedStatus = "Closed"
)

func (s *service) organizationIDAndUserIDByUsername(ctx context.Context, username string) (userID string, organizationID string, err error) {
	userID, err = s.employeeService.IDByUsername(ctx, username)
	if err != nil {
		return "", "", err
	}

	organizationID, err = s.organizationRespService.GetOrganizationIDByRepresentative(ctx, userID)
	if err != nil {
		return "", "", err
	}

	return userID, organizationID, nil
}

func (s *service) Create(ctx context.Context, bid model.Bid) (model.Bid, error) {
	bid, err := s.bidsRepository.Create(ctx, bid)
	if err != nil {
		switch {
		case errors.Is(err, repository_bid.ErrInvalidReq):
			return model.Bid{}, service_bids.ErrInvalidReq
		case errors.Is(err, repository_bid.ErrInvalidAuthorID):
			return model.Bid{}, service_bids.ErrInvalidAuthorID
		case errors.Is(err, repository_bid.ErrInvalidTenderID):
			return model.Bid{}, service_bids.ErrInvalidTenderID
		default:
			return model.Bid{}, service_bids.ErrInternal
		}
	}

	return bid, nil
}

func (s *service) BidsByUser(ctx context.Context, username string, limit int, offset int) ([]model.Bid, error) {
	userID, organizationID, err := s.organizationIDAndUserIDByUsername(ctx, username)
	if err != nil {
		if !errors.Is(err, service_organization_resp.ErrUserHasNoOrganization) {
			return nil, err
		}
	}

	bids, err := s.bidsRepository.BidsByAuthorID(ctx, userID, organizationID, limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, repository_bid.ErrNoBids):
			return nil, service_bids.ErrNoBids
		default:
			return nil, service_bids.ErrInternal
		}
	}

	return bids, nil
}

func (s *service) BidsByTenderID(ctx context.Context, tenderID string, username string, limit int, offset int) ([]model.Bid, error) {
	//CHECK USER. ONLY TENDER CREATOR CAN USE THIS FUNCTIONAL
	_, organizationID, err := s.organizationIDAndUserIDByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, service_organization_resp.ErrUserHasNoOrganization) {
			return nil, service_bids.ErrForbidden
		}
	}

	//CHECK THAT USER ORGANIZATION IS A TENDER CREATOR
	exists, err := s.tenderService.ConfirmTenderCreator(ctx, tenderID, organizationID)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, service_bids.ErrForbidden
	}

	bids, err := s.bidsRepository.BidsByTenderID(ctx, tenderID, limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, repository_bid.ErrNoBids):
			return nil, service_bids.ErrNoBids
		default:
			return nil, service_bids.ErrInternal
		}
	}

	return bids, nil
}

func (s *service) GetStatus(ctx context.Context, bidID string, username string) (status string, err error) {
	userID, organizationID, err := s.organizationIDAndUserIDByUsername(ctx, username)
	if err != nil {
		if !errors.Is(err, service_organization_resp.ErrUserHasNoOrganization) {
			return "", err
		}
	}

	status, tenderID, authorID, err := s.bidsRepository.GetStatus(ctx, bidID)
	if err != nil {
		switch {
		case errors.Is(err, repository_bid.ErrNoBids):
			return "", service_bids.ErrNoBids
		default:
			return "", service_bids.ErrInternal
		}
	}

	tenderOrganizationID, err := s.tenderService.TenderOrganizationID(ctx, tenderID)
	if err != nil {
		return "", err
	}

	if userID != authorID && organizationID != authorID && organizationID != tenderOrganizationID {
		return "", service_bids.ErrForbidden
	}

	return status, nil
}

func (s *service) ChangeStatus(ctx context.Context, bidID string, username string, status string) (bid model.Bid, err error) {
	//CHECK ACCESS
	userID, organizationID, err := s.organizationIDAndUserIDByUsername(ctx, username)
	if err != nil {
		return model.Bid{}, err
	}

	authorID, err := s.bidsRepository.BidAuthorID(ctx, bidID)
	if err != nil {
		return model.Bid{}, service_bids.ErrInternal
	}

	if authorID != userID && authorID != organizationID {
		return model.Bid{}, service_bids.ErrForbidden
	}

	//CHANGE STATUS
	bid, err = s.bidsRepository.ChangeStatus(ctx, bidID, status)
	if err != nil {
		switch {
		case errors.Is(err, repository_bid.ErrInvalidBidStatus):
			return model.Bid{}, service_bids.ErrInvalidBidStatus
		case errors.Is(err, repository_bid.ErrNoBids):
			return model.Bid{}, service_bids.ErrNoBids
		default:
			return model.Bid{}, service_bids.ErrInternal
		}
	}

	return bid, nil
}

func (s *service) Edit(ctx context.Context, bidID string, username string, bid model.Bid) (model.Bid, error) {
	//CHECK ACCESS
	userID, organizationID, err := s.organizationIDAndUserIDByUsername(ctx, username)
	if err != nil {
		return model.Bid{}, err
	}

	authorID, err := s.bidsRepository.BidAuthorID(ctx, bidID)
	if err != nil {
		return model.Bid{}, service_bids.ErrInternal
	}

	if authorID != userID && authorID != organizationID {
		return model.Bid{}, service_bids.ErrForbidden
	}

	//EDIT
	updatedBid, err := s.bidsRepository.Edit(ctx, bidID, bid)
	if err != nil {
		switch {
		case errors.Is(err, repository_bid.ErrNoSuggestionToUpdate):
			return model.Bid{}, service_bids.ErrNoSuggestionToUpdate
		case errors.Is(err, repository_bid.ErrNoBids):
			return model.Bid{}, service_bids.ErrNoBids
		default:
			return model.Bid{}, service_bids.ErrInternal
		}
	}

	return updatedBid, nil
}

func (s *service) SubmitDecision(ctx context.Context, bidID string, decision string, username string) (bid model.Bid, isWinner bool, err error) {
	//CHECK USER ACCESS
	userID, userOrganizationID, err := s.organizationIDAndUserIDByUsername(ctx, username)
	if err != nil {
		return model.Bid{}, false, err
	}

	//GET ORGANIZATION THAT LAUNCHED TENDER
	tenderID, err := s.bidsRepository.BidTenderID(ctx, bidID)
	if err != nil {
		return model.Bid{}, false, service_bids.ErrInternal
	}

	tenderOrganizationID, err := s.tenderService.TenderOrganizationID(ctx, tenderID)
	if err != nil {
		return model.Bid{}, false, err
	}

	//CHECK IF USER ORGANIZATION LAUNCHED TENDER
	if tenderOrganizationID != userOrganizationID {
		return model.Bid{}, false, service_bids.ErrForbidden
	}

	//CHECK TENDER STATUS
	_, status, err := s.tenderService.TenderStatus(ctx, tenderID)
	if err != nil {
		return model.Bid{}, false, err
	}

	if status == tenderClosedStatus {
		return model.Bid{}, false, service_bids.ErrTenderClosed
	}

	//GET DECISION STATISTIC
	applied, rejected, err := s.decisionService.DecisionStats(ctx, bidID)
	if err != nil {
		return model.Bid{}, false, err
	}
	if rejected != 0 {
		return model.Bid{}, false, service_bids.ErrBidBeenRejected
	}

	//SUBMIT DECISION
	if err = s.decisionService.SubmitDecision(ctx, userID, tenderID, bidID, decision); err != nil {
		return model.Bid{}, false, err
	}

	//ORGANIZATION REPRESENTATIVES AMOUNT
	representatives, err := s.organizationRespService.OrganizationRepresentativesAmount(ctx, tenderOrganizationID)
	if err != nil {
		return model.Bid{}, false, err
	}

	//QUORUM CHECK
	if float64(representatives)/2 <= float64(applied) {
		if err = s.tenderService.ChangeTenderStatusForce(ctx, tenderID, tenderClosedStatus); err != nil {
			return model.Bid{}, false, err
		}
		isWinner = true
	}

	bid, err = s.bidsRepository.BidByID(ctx, bidID)
	if err != nil {
		switch {
		case errors.Is(err, repository_bid.ErrNoBids):
			return model.Bid{}, false, service_bids.ErrNoBids
		default:
			return model.Bid{}, false, service_bids.ErrInternal
		}
	}
	return bid, isWinner, nil
}

func (s *service) Feedback(ctx context.Context, bidID string, username string, feedback string) (model.Bid, error) {
	//CHECK ACCESS
	tenderID, err := s.bidsRepository.BidTenderID(ctx, bidID)
	if err != nil {
		return model.Bid{}, service_bids.ErrInternal
	}

	userID, err := s.employeeService.IDByUsername(ctx, username)
	if err != nil {
		return model.Bid{}, service_bids.ErrInternal
	}

	organizationID, err := s.organizationRespService.GetOrganizationIDByRepresentative(ctx, userID)
	if err != nil {
		return model.Bid{}, err
	}

	exists, err := s.tenderService.ConfirmTenderCreator(ctx, tenderID, organizationID)
	if err != nil {
		return model.Bid{}, err
	}

	if !exists {
		return model.Bid{}, service_bids.ErrForbidden
	}

	bid, err := s.bidsRepository.BidByID(ctx, bidID)
	if err != nil {
		switch {
		case errors.Is(err, repository_bid.ErrNoBids):
			return model.Bid{}, service_bids.ErrNoBids
		default:
			return model.Bid{}, service_bids.ErrInternal
		}
	}

	err = s.feedbackService.Feedback(ctx, userID, feedback)
	if err != nil {
		return model.Bid{}, service_bids.ErrInternal
	}

	return bid, nil
}

func (s *service) RollbackVersion(ctx context.Context, bidID string, username string, version int) (model.Bid, error) {
	//CHECK ACCESS
	userID, organizationID, err := s.organizationIDAndUserIDByUsername(ctx, username)
	if err != nil {
		return model.Bid{}, err
	}

	authorID, err := s.bidsRepository.BidAuthorID(ctx, bidID)
	if err != nil {
		return model.Bid{}, service_bids.ErrInternal
	}

	if authorID != userID && authorID != organizationID {
		return model.Bid{}, service_bids.ErrForbidden
	}

	bid, err := s.bidsRepository.RollbackVersion(ctx, bidID, version)
	if err != nil {
		switch {
		case errors.Is(err, repository_bid.ErrNoBids):
			return model.Bid{}, service_bids.ErrNoBids
		default:
			return model.Bid{}, service_bids.ErrInternal
		}
	}

	return bid, nil
}

func (s *service) GetReviews(ctx context.Context, tenderID string, authorUsername, requesterUsername string, limit int, offset int) ([]model.Feedback, error) {
	//CHECK ACCESS
	_, organizationID, err := s.organizationIDAndUserIDByUsername(ctx, requesterUsername)
	if err != nil {
		return nil, err
	}

	tenderOrganizationID, err := s.tenderService.TenderOrganizationID(ctx, tenderID)
	if err != nil {
		return nil, err
	}

	if organizationID != tenderOrganizationID {
		return nil, service_bids.ErrForbidden
	}

	reviews, err := s.feedbackService.GetReviews(ctx, authorUsername, limit, offset)
	if err != nil {
		return nil, err
	}
	return reviews, nil
}

func New(bidsRepository repository_bid.Repository, employeeService service_employee.Service, organizationRespService service_organization_resp.Service, tenderService service_tenders.Service, decisionService service_decision.Service, feedbackService service_feedback.Service, logger *slog.Logger) service_bids.Service {
	s := &service{
		bidsRepository:          bidsRepository,
		employeeService:         employeeService,
		tenderService:           tenderService,
		decisionService:         decisionService,
		feedbackService:         feedbackService,
		organizationRespService: organizationRespService,
		logger:                  logger,
	}

	return s
}

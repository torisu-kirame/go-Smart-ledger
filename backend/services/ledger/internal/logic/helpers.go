package logic

import (
	"errors"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/accounting"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	"github.com/smart-ledger/go-smart-ledger/backend/pkg/ledgersvc"
	xerrors "github.com/zeromicro/x/errors"
)

func toCodeErr(err error) error {
	return ToCodeErr(err)
}

// ToCodeErr maps domain errors to HTTP code errors.
func ToCodeErr(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, domain.ErrLedgerNotFound):
		return xerrors.New(404, err.Error())
	case errors.Is(err, domain.ErrMultiNeedsTwo),
		errors.Is(err, domain.ErrPrivateOne),
		errors.Is(err, domain.ErrInvalidMember),
		errors.Is(err, domain.ErrUnauthorized),
		errors.Is(err, domain.ErrCannotInviteSelf),
		errors.Is(err, domain.ErrEntryValidation),
		errors.Is(err, domain.ErrInvalidSchema),
		errors.Is(err, domain.ErrBookkeepingModeMismatch),
		errors.Is(err, domain.ErrMultiTableDisabled),
		errors.Is(err, domain.ErrTableNotFound),
		errors.Is(err, domain.ErrInvalidTable),
		errors.Is(err, domain.ErrTableHasEntries),
		errors.Is(err, ledgersvc.ErrImportHasErrors),
		errors.Is(err, ledgersvc.ErrRestoreConflict),
		errors.Is(err, accounting.ErrInvalidJournal),
		errors.Is(err, accounting.ErrUnbalanced),
		errors.Is(err, accounting.ErrInvalidAccount),
		errors.Is(err, accounting.ErrAccountNotFound),
		errors.Is(err, accounting.ErrPeriodClosed),
		errors.Is(err, accounting.ErrInvalidPeriod),
		errors.Is(err, accounting.ErrInvalidBudget),
		errors.Is(err, accounting.ErrInvalidCurrency),
		errors.Is(err, accounting.ErrInvalidFxRates),
		errors.Is(err, accounting.ErrInvalidTax):
		return xerrors.New(400, err.Error())
	case errors.Is(err, domain.ErrAlreadyMember),
		errors.Is(err, domain.ErrInviteAlreadyPending):
		return xerrors.New(409, err.Error())
	case errors.Is(err, domain.ErrInviteNotFound),
		errors.Is(err, ledgersvc.ErrAttachmentNotFound):
		return xerrors.New(404, err.Error())
	default:
		return xerrors.New(500, err.Error())
	}
}

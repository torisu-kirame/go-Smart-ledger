package logic

import (
	"errors"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/domain"
	xerrors "github.com/zeromicro/x/errors"
)

func toCodeErr(err error) error {
	if err == nil {
		return nil
	}
	switch {
	case errors.Is(err, domain.ErrLedgerNotFound):
		return xerrors.New(404, err.Error())
	case errors.Is(err, domain.ErrMultiNeedsTwo),
		errors.Is(err, domain.ErrPrivateOne),
		errors.Is(err, domain.ErrInvalidMember),
		errors.Is(err, domain.ErrUnauthorized):
		return xerrors.New(400, err.Error())
	default:
		return xerrors.New(500, err.Error())
	}
}

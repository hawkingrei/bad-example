package status

import (
	"go-common/library/ecode"
	"go-common/library/ecode/pb"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ToStatus convert ecode to grpc unknown with ecode code message.
func ToStatus(ec ecode.Error) error {
	if ecode.Equal(ec, ecode.OK) {
		return nil
	}
	if pe, ok := ec.(*pb.Error); ok {
		if ds, err := status.New(codes.Unknown, ec.Error()).WithDetails(pe); err == nil {
			return ds.Err()
		}
	}
	return status.Errorf(codes.Unknown, ec.Error())
}

// ToEcode convert err to ecode.
func ToEcode(err error) ecode.Error {
	var (
		st *status.Status
		ok bool
	)
	//if err, convert grpc error to ecode error
	if st, ok = status.FromError(err); !ok {
		return ecode.Cause(err)
	}
	if st.Code() != codes.Unknown {
		return statusToecode(st)
	}
	// get error detail from grpc status if exists
	if ds := st.Details(); len(ds) > 0 {
		if perr, ok := ds[0].(*pb.Error); ok && perr.GetErrDetail() != nil {
			return perr
		}
	}
	return ecode.String(st.Message())
}

func statusToecode(status *status.Status) (err ecode.Error) {
	switch status.Code() {
	case codes.OK:
		return ecode.OK
	case codes.InvalidArgument:
		return ecode.RequestErr
	case codes.NotFound:
		return ecode.NothingFound
	case codes.PermissionDenied:
		return ecode.AccessDenied
	case codes.Unauthenticated:
		return ecode.Unauthorized
	case codes.ResourceExhausted:
		return ecode.LimitExceed
	case codes.OutOfRange:
		return ecode.RequestErr
	case codes.Unimplemented:
		return ecode.MethodNotAllowed
	case codes.DeadlineExceeded:
		return ecode.Deadline
	case codes.Unavailable:
		return ecode.ServiceUnavailable
	default:
		return ecode.ServerErr
	}
}

package change

type RequestDetector struct {
}

type Detector interface {
	Changed(
		existingMapLocalAddress bool,
		existingIp *string,
		existingIpv6 *string,
		existingDkimKey *string,
		existingLocalIp *string,
		existingRelay bool,
		newMapLocalAddress bool,
		newIp *string,
		newIpv6 *string,
		newDkimKey *string,
		newLocalIp *string,
		newRelay bool) bool
}

func New() *RequestDetector {
	return &RequestDetector{}
}

func (d *RequestDetector) Changed(
	existingMapLocalAddress bool, existingIp *string, existingIpv6 *string, existingDkimKey *string, existingLocalIp *string, existingRelay bool,
	newMapLocalAddress bool, newIp *string, newIpv6 *string, newDkimKey *string, newLocalIp *string, newRelay bool) bool {

	changed := (existingMapLocalAddress != newMapLocalAddress) ||
		!Equals(existingIp, newIp) ||
		!Equals(existingLocalIp, newLocalIp) ||
		!Equals(existingIpv6, newIpv6) ||
		!Equals(existingDkimKey, newDkimKey) ||
		(existingRelay != newRelay)

	return changed
}

func Equals(left *string, right *string) bool {
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}
	return *left == *right
}
